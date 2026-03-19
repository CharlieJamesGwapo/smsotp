package contacts

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"smsotp/internal/database"

	"github.com/go-chi/chi/v5"
)

type Contact struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email,omitempty"`
	Notes     string `json:"notes,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse struct {
	Data    []Contact `json:"data"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	PerPage int       `json:"per_page"`
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = regexp.MustCompile(`[^0-9+]`).ReplaceAllString(phone, "")
	if strings.HasPrefix(phone, "+63") {
		phone = "0" + phone[3:]
	} else if strings.HasPrefix(phone, "63") && len(phone) == 12 {
		phone = "0" + phone[2:]
	}
	return phone
}

func HandleList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	search := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	var rows *interface{}
	_ = rows

	query := "SELECT id, name, phone, COALESCE(email,''), COALESCE(notes,''), created_at, updated_at FROM contacts"
	countQuery := "SELECT COUNT(*) FROM contacts"
	var args []interface{}

	if search != "" {
		filter := " WHERE name LIKE ? OR phone LIKE ? OR email LIKE ?"
		query += filter
		countQuery += filter
		s := "%" + search + "%"
		args = append(args, s, s, s)
	}

	database.DB.QueryRow(countQuery, args...).Scan(&total)

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)

	dbRows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer dbRows.Close()

	contacts := []Contact{}
	for dbRows.Next() {
		var c Contact
		dbRows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
		contacts = append(contacts, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{
		Data:    contacts,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	var c Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	c.Phone = normalizePhone(c.Phone)
	if c.Name == "" || c.Phone == "" {
		http.Error(w, `{"error":"Name and phone are required"}`, http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO contacts (name, phone, email, notes) VALUES (?, ?, ?, ?)",
		c.Name, c.Phone, c.Email, c.Notes,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, `{"error":"Phone number already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	c.Phone = normalizePhone(c.Phone)
	_, err := database.DB.Exec(
		"UPDATE contacts SET name=?, phone=?, email=?, notes=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		c.Name, c.Phone, c.Email, c.Notes, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, `{"error":"Phone number already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Contact updated"})
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	database.DB.Exec("DELETE FROM contacts WHERE id=?", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Contact deleted"})
}

func HandleImport(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1 << 20) // 1MB max
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"No file uploaded"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, `{"error":"Invalid CSV file"}`, http.StatusBadRequest)
		return
	}

	imported := 0
	skipped := 0
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 2 {
			continue
		}
		name := strings.TrimSpace(row[0])
		phone := normalizePhone(row[1])
		email := ""
		notes := ""
		if len(row) > 2 {
			email = strings.TrimSpace(row[2])
		}
		if len(row) > 3 {
			notes = strings.TrimSpace(row[3])
		}

		_, err := database.DB.Exec(
			"INSERT OR IGNORE INTO contacts (name, phone, email, notes) VALUES (?, ?, ?, ?)",
			name, phone, email, notes,
		)
		if err != nil {
			skipped++
		} else {
			imported++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"imported": imported, "skipped": skipped})
}
