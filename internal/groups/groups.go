package groups

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"smsotp/internal/database"

	"github.com/go-chi/chi/v5"
)

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type GroupDetail struct {
	Group
	Members []GroupMember `json:"members"`
}

type GroupMember struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type MembersRequest struct {
	ContactIDs []int `json:"contact_ids"`
}

type ListResponse struct {
	Data    []Group `json:"data"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
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
	database.DB.QueryRow("SELECT COUNT(*) FROM groups_").Scan(&total)

	rows, err := database.DB.Query(`
		SELECT g.id, g.name, COALESCE(g.description,''), g.created_at, g.updated_at,
			(SELECT COUNT(*) FROM contact_groups cg WHERE cg.group_id = g.id) as member_count
		FROM groups_ g ORDER BY g.created_at DESC LIMIT ? OFFSET ?`, perPage, offset)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var g Group
		rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt, &g.MemberCount)
		groups = append(groups, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Data: groups, Total: total, Page: page, PerPage: perPage})
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var g GroupDetail
	err := database.DB.QueryRow(
		"SELECT id, name, COALESCE(description,''), created_at, updated_at FROM groups_ WHERE id=?", id,
	).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		http.Error(w, `{"error":"Group not found"}`, http.StatusNotFound)
		return
	}

	rows, _ := database.DB.Query(`
		SELECT c.id, c.name, c.phone FROM contacts c
		JOIN contact_groups cg ON c.id = cg.contact_id
		WHERE cg.group_id = ?`, id)
	defer rows.Close()

	g.Members = []GroupMember{}
	for rows.Next() {
		var m GroupMember
		rows.Scan(&m.ID, &m.Name, &m.Phone)
		g.Members = append(g.Members, m)
	}
	g.MemberCount = len(g.Members)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func HandleCreate(w http.ResponseWriter, r *http.Request) {
	var g Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if g.Name == "" {
		http.Error(w, `{"error":"Name is required"}`, http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec("INSERT INTO groups_ (name, description) VALUES (?, ?)", g.Name, g.Description)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			http.Error(w, `{"error":"Group name already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	g.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var g Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(
		"UPDATE groups_ SET name=?, description=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		g.Name, g.Description, id,
	)
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Group updated"})
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	database.DB.Exec("DELETE FROM groups_ WHERE id=?", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Group deleted"})
}

func HandleAddMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req MembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	for _, cid := range req.ContactIDs {
		database.DB.Exec("INSERT OR IGNORE INTO contact_groups (group_id, contact_id) VALUES (?, ?)", id, cid)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Members added"})
}

func HandleRemoveMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req MembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	for _, cid := range req.ContactIDs {
		database.DB.Exec("DELETE FROM contact_groups WHERE group_id=? AND contact_id=?", id, cid)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Members removed"})
}
