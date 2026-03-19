package settings

import (
	"encoding/json"
	"net/http"

	"smsotp/internal/database"
)

type Settings struct {
	PhoneGatewayURL    string `json:"phone_gateway_url"`
	SemaphoreAPIKey    string `json:"semaphore_api_key"`
	SemaphoreEnabled   string `json:"semaphore_enabled"`
	SemaphoreSenderName string `json:"semaphore_sender_name"`
	SMSDelayMs         string `json:"sms_delay_ms"`
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	s := Settings{}
	database.DB.QueryRow("SELECT value FROM settings WHERE key='phone_gateway_url'").Scan(&s.PhoneGatewayURL)
	database.DB.QueryRow("SELECT value FROM settings WHERE key='semaphore_api_key'").Scan(&s.SemaphoreAPIKey)
	database.DB.QueryRow("SELECT value FROM settings WHERE key='semaphore_enabled'").Scan(&s.SemaphoreEnabled)
	database.DB.QueryRow("SELECT value FROM settings WHERE key='semaphore_sender_name'").Scan(&s.SemaphoreSenderName)
	database.DB.QueryRow("SELECT value FROM settings WHERE key='sms_delay_ms'").Scan(&s.SMSDelayMs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	updates := map[string]string{
		"phone_gateway_url":     s.PhoneGatewayURL,
		"semaphore_api_key":     s.SemaphoreAPIKey,
		"semaphore_enabled":     s.SemaphoreEnabled,
		"semaphore_sender_name": s.SemaphoreSenderName,
		"sms_delay_ms":          s.SMSDelayMs,
	}

	for k, v := range updates {
		if v != "" {
			database.DB.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", k, v)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated"})
}
