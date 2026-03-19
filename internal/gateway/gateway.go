package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"smsotp/internal/database"
)

type SendResult struct {
	Success     bool
	GatewayUsed string
	Error       string
}

func getSetting(key string) string {
	var val string
	database.DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	return val
}

func sendViaPhone(phone, message string) error {
	url := getSetting("phone_gateway_url") + "/send-sms"
	payload, _ := json.Marshal(map[string]string{
		"phone":   phone,
		"message": message,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("phone gateway error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("phone gateway returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func sendViaSemaphore(phone, message string) error {
	apiKey := getSetting("semaphore_api_key")
	senderName := getSetting("semaphore_sender_name")

	payload, _ := json.Marshal(map[string]string{
		"apikey":      apiKey,
		"number":      phone,
		"message":     message,
		"sendername":  senderName,
	})

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("https://api.semaphore.co/api/v4/messages", "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("semaphore error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("semaphore returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func HealthCheck() error {
	url := getSetting("phone_gateway_url")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func Send(phone, message string) SendResult {
	// Try phone gateway first
	err := sendViaPhone(phone, message)
	if err == nil {
		return SendResult{Success: true, GatewayUsed: "phone"}
	}

	// Try Semaphore fallback
	if getSetting("semaphore_enabled") == "true" {
		semErr := sendViaSemaphore(phone, message)
		if semErr == nil {
			return SendResult{Success: true, GatewayUsed: "semaphore"}
		}
		return SendResult{Success: false, Error: fmt.Sprintf("phone: %v; semaphore: %v", err, semErr)}
	}

	return SendResult{Success: false, Error: err.Error()}
}
