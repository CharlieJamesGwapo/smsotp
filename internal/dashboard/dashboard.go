package dashboard

import (
	"encoding/json"
	"net/http"

	"smsotp/internal/database"
)

type Stats struct {
	TotalContacts    int     `json:"total_contacts"`
	TotalGroups      int     `json:"total_groups"`
	TotalTemplates   int     `json:"total_templates"`
	MessagesSentToday int    `json:"messages_sent_today"`
	MessagesSentWeek  int    `json:"messages_sent_week"`
	MessagesSentMonth int    `json:"messages_sent_month"`
	MessagesFailed    int    `json:"messages_failed"`
	MessagesScheduled int    `json:"messages_scheduled"`
	SuccessRate       float64 `json:"success_rate"`
}

func HandleStats(w http.ResponseWriter, r *http.Request) {
	var s Stats

	database.DB.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&s.TotalContacts)
	database.DB.QueryRow("SELECT COUNT(*) FROM groups_").Scan(&s.TotalGroups)
	database.DB.QueryRow("SELECT COUNT(*) FROM templates").Scan(&s.TotalTemplates)

	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE status='sent' AND DATE(sent_at)=DATE('now')").Scan(&s.MessagesSentToday)
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE status='sent' AND sent_at >= datetime('now', '-7 days')").Scan(&s.MessagesSentWeek)
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE status='sent' AND sent_at >= datetime('now', '-30 days')").Scan(&s.MessagesSentMonth)
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE status='failed'").Scan(&s.MessagesFailed)
	database.DB.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE status='pending' AND scheduled_at IS NOT NULL").Scan(&s.MessagesScheduled)

	var totalSent, totalAll int
	database.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE status='sent'").Scan(&totalSent)
	database.DB.QueryRow("SELECT COUNT(*) FROM messages WHERE status IN ('sent','failed')").Scan(&totalAll)
	if totalAll > 0 {
		s.SuccessRate = float64(totalSent) / float64(totalAll) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}
