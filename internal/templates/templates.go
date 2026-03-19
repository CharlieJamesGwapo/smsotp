package templates

import (
	"encoding/json"
	"net/http"
	"strconv"

	"smsotp/internal/database"

	"github.com/go-chi/chi/v5"
)

type Template struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse struct {
	Data    []Template `json:"data"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
}

func HandleList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM templates").Scan(&total)

	rows, err := database.DB.Query(
		"SELECT id, name, body, created_at, updated_at FROM templates ORDER BY created_at DESC LIMIT ? OFFSET ?",
		perPage, offset,
	)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	templates := []Template{}
	for rows.Next() {
		var t Template
		rows.Scan(&t.ID, &t.Name, &t.Body, &t.CreatedAt, &t.UpdatedAt)
		templates = append(templates, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Data: templates, Total: total, Page: page, PerPage: perPage})
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	var t Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if t.Name == "" || t.Body == "" {
		http.Error(w, `{"error":"Name and body are required"}`, http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec("INSERT INTO templates (name, body) VALUES (?, ?)", t.Name, t.Body)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	database.DB.Exec(
		"UPDATE templates SET name=?, body=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		t.Name, t.Body, id,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Template updated"})
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	database.DB.Exec("DELETE FROM templates WHERE id=?", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Template deleted"})
}
