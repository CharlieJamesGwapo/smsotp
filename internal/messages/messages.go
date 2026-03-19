package messages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smsotp/internal/database"
	"smsotp/internal/gateway"

	"github.com/go-chi/chi/v5"
)

type SendRequest struct {
	ContactIDs []int             `json:"contact_ids"`
	GroupIDs   []int             `json:"group_ids"`
	TemplateID *int              `json:"template_id"`
	Body       string            `json:"body"`
	Variables  map[string]string `json:"variables"`
}

type ScheduleRequest struct {
	SendRequest
	ScheduledAt string `json:"scheduled_at"`
}

type Message struct {
	ID           int    `json:"id"`
	TemplateID   *int   `json:"template_id,omitempty"`
	ContactID    *int   `json:"contact_id,omitempty"`
	Phone        string `json:"phone"`
	Body         string `json:"body"`
	Status       string `json:"status"`
	GatewayUsed  string `json:"gateway_used,omitempty"`
	ScheduledAt  string `json:"scheduled_at,omitempty"`
	SentAt       string `json:"sent_at,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    string `json:"created_at"`
	ContactName  string `json:"contact_name,omitempty"`
}

type ListResponse struct {
	Data    []Message `json:"data"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	PerPage int       `json:"per_page"`
}

type recipient struct {
	ID    int
	Name  string
	Phone string
	Email string
}

func resolveRecipients(contactIDs []int, groupIDs []int) []recipient {
	seen := map[string]bool{}
	var results []recipient

	for _, cid := range contactIDs {
		var r recipient
		err := database.DB.QueryRow("SELECT id, name, phone, COALESCE(email,'') FROM contacts WHERE id=?", cid).
			Scan(&r.ID, &r.Name, &r.Phone, &r.Email)
		if err == nil && !seen[r.Phone] {
			seen[r.Phone] = true
			results = append(results, r)
		}
	}

	for _, gid := range groupIDs {
		rows, err := database.DB.Query(`
			SELECT c.id, c.name, c.phone, COALESCE(c.email,'')
			FROM contacts c JOIN contact_groups cg ON c.id = cg.contact_id
			WHERE cg.group_id = ?`, gid)
		if err != nil {
			continue
		}
		for rows.Next() {
			var r recipient
			rows.Scan(&r.ID, &r.Name, &r.Phone, &r.Email)
			if !seen[r.Phone] {
				seen[r.Phone] = true
				results = append(results, r)
			}
		}
		rows.Close()
	}

	return results
}

func resolveBody(templateID *int, body string, variables map[string]string, r recipient) string {
	msg := body

	if templateID != nil && *templateID > 0 {
		var tmplBody string
		err := database.DB.QueryRow("SELECT body FROM templates WHERE id=?", *templateID).Scan(&tmplBody)
		if err == nil {
			msg = tmplBody
		}
	}

	// Replace contact field variables
	msg = strings.ReplaceAll(msg, "{name}", r.Name)
	msg = strings.ReplaceAll(msg, "{phone}", r.Phone)
	msg = strings.ReplaceAll(msg, "{email}", r.Email)

	// Replace custom variables
	for k, v := range variables {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}

	return msg
}

func getDelay() time.Duration {
	var val string
	database.DB.QueryRow("SELECT value FROM settings WHERE key='sms_delay_ms'").Scan(&val)
	ms, err := strconv.Atoi(val)
	if err != nil || ms < 0 {
		ms = 1000
	}
	return time.Duration(ms) * time.Millisecond
}

func HandleSend(w http.ResponseWriter, r *http.Request) {
	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.TemplateID == nil && req.Body == "" {
		http.Error(w, `{"error":"Either template_id or body is required"}`, http.StatusBadRequest)
		return
	}

	recipients := resolveRecipients(req.ContactIDs, req.GroupIDs)
	if len(recipients) == 0 {
		http.Error(w, `{"error":"No recipients found"}`, http.StatusBadRequest)
		return
	}

	// Health check
	if err := gateway.HealthCheck(); err != nil {
		var semEnabled string
		database.DB.QueryRow("SELECT value FROM settings WHERE key='semaphore_enabled'").Scan(&semEnabled)
		if semEnabled != "true" {
			http.Error(w, `{"error":"Phone gateway is unreachable and Semaphore is disabled"}`, http.StatusServiceUnavailable)
			return
		}
	}

	delay := getDelay()
	sent := 0
	failed := 0

	for i, rec := range recipients {
		msg := resolveBody(req.TemplateID, req.Body, req.Variables, rec)
		result := gateway.Send(rec.Phone, msg)

		status := "sent"
		var errMsg *string
		if !result.Success {
			status = "failed"
			errMsg = &result.Error
			failed++
		} else {
			sent++
		}

		var sentAt *string
		if result.Success {
			now := time.Now().Format("2006-01-02 15:04:05")
			sentAt = &now
		}

		database.DB.Exec(`
			INSERT INTO messages (template_id, contact_id, phone, body, status, gateway_used, sent_at, error_message)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			req.TemplateID, rec.ID, rec.Phone, msg, status, result.GatewayUsed, sentAt, errMsg,
		)

		if i < len(recipients)-1 {
			time.Sleep(delay)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"sent": sent, "failed": failed, "total": len(recipients)})
}

func HandleSchedule(w http.ResponseWriter, r *http.Request) {
	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.ScheduledAt == "" {
		http.Error(w, `{"error":"scheduled_at is required"}`, http.StatusBadRequest)
		return
	}

	if req.TemplateID == nil && req.Body == "" {
		http.Error(w, `{"error":"Either template_id or body is required"}`, http.StatusBadRequest)
		return
	}

	recipients := resolveRecipients(req.ContactIDs, req.GroupIDs)
	if len(recipients) == 0 {
		http.Error(w, `{"error":"No recipients found"}`, http.StatusBadRequest)
		return
	}

	// Fan-out at schedule time
	for _, rec := range recipients {
		msg := resolveBody(req.TemplateID, req.Body, req.Variables, rec)
		database.DB.Exec(`
			INSERT INTO messages (template_id, contact_id, phone, body, status, scheduled_at)
			VALUES (?, ?, ?, ?, 'pending', ?)`,
			req.TemplateID, rec.ID, rec.Phone, msg, req.ScheduledAt,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"scheduled": len(recipients)})
}

func HandleList_(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	status := r.URL.Query().Get("status")
	gatewayFilter := r.URL.Query().Get("gateway")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	query := `SELECT m.id, m.template_id, m.contact_id, m.phone, m.body, m.status,
		COALESCE(m.gateway_used,''), COALESCE(m.scheduled_at,''), COALESCE(m.sent_at,''),
		COALESCE(m.error_message,''), m.created_at, COALESCE(c.name,'Unknown')
		FROM messages m LEFT JOIN contacts c ON m.contact_id = c.id`
	countQuery := "SELECT COUNT(*) FROM messages m"

	var conditions []string
	var args []interface{}

	if status != "" {
		conditions = append(conditions, "m.status = ?")
		args = append(args, status)
	}
	if gatewayFilter != "" {
		conditions = append(conditions, "m.gateway_used = ?")
		args = append(args, gatewayFilter)
	}

	if len(conditions) > 0 {
		where := " WHERE " + strings.Join(conditions, " AND ")
		query += where
		countQuery += where
	}

	var total int
	database.DB.QueryRow(countQuery, args...).Scan(&total)

	query += " ORDER BY m.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Database error: %v"}`, err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	msgs := []Message{}
	for rows.Next() {
		var m Message
		rows.Scan(&m.ID, &m.TemplateID, &m.ContactID, &m.Phone, &m.Body, &m.Status,
			&m.GatewayUsed, &m.ScheduledAt, &m.SentAt, &m.ErrorMessage, &m.CreatedAt, &m.ContactName)
		msgs = append(msgs, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Data: msgs, Total: total, Page: page, PerPage: perPage})
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m Message
	err := database.DB.QueryRow(`
		SELECT m.id, m.template_id, m.contact_id, m.phone, m.body, m.status,
		COALESCE(m.gateway_used,''), COALESCE(m.scheduled_at,''), COALESCE(m.sent_at,''),
		COALESCE(m.error_message,''), m.created_at, COALESCE(c.name,'Unknown')
		FROM messages m LEFT JOIN contacts c ON m.contact_id = c.id WHERE m.id=?`, id).
		Scan(&m.ID, &m.TemplateID, &m.ContactID, &m.Phone, &m.Body, &m.Status,
			&m.GatewayUsed, &m.ScheduledAt, &m.SentAt, &m.ErrorMessage, &m.CreatedAt, &m.ContactName)
	if err != nil {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func HandleCancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var status string
	database.DB.QueryRow("SELECT status FROM messages WHERE id=?", id).Scan(&status)
	if status != "pending" {
		http.Error(w, `{"error":"Can only cancel pending messages"}`, http.StatusBadRequest)
		return
	}

	database.DB.Exec("DELETE FROM messages WHERE id=? AND status='pending'", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Scheduled message cancelled"})
}

func HandleRetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m Message
	err := database.DB.QueryRow("SELECT id, phone, body, status FROM messages WHERE id=?", id).
		Scan(&m.ID, &m.Phone, &m.Body, &m.Status)
	if err != nil {
		http.Error(w, `{"error":"Message not found"}`, http.StatusNotFound)
		return
	}
	if m.Status != "failed" {
		http.Error(w, `{"error":"Can only retry failed messages"}`, http.StatusBadRequest)
		return
	}

	result := gateway.Send(m.Phone, m.Body)
	if result.Success {
		now := time.Now().Format("2006-01-02 15:04:05")
		database.DB.Exec("UPDATE messages SET status='sent', gateway_used=?, sent_at=?, error_message=NULL WHERE id=?",
			result.GatewayUsed, now, id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully"})
	} else {
		database.DB.Exec("UPDATE messages SET error_message=? WHERE id=?", result.Error, id)
		http.Error(w, fmt.Sprintf(`{"error":"Retry failed: %s"}`, result.Error), http.StatusServiceUnavailable)
	}
}
