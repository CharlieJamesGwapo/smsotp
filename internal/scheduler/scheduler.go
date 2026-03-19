package scheduler

import (
	"log"
	"strconv"
	"time"

	"smsotp/internal/database"
	"smsotp/internal/gateway"
)

func Start() {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			processPending()
		}
	}()
	log.Println("Scheduler started — checking every 60 seconds")
}

func processPending() {
	rows, err := database.DB.Query(`
		SELECT id, phone, body FROM messages
		WHERE status='pending' AND scheduled_at IS NOT NULL AND scheduled_at <= datetime('now')
	`)
	if err != nil {
		log.Println("Scheduler query error:", err)
		return
	}
	defer rows.Close()

	type msg struct {
		ID    int
		Phone string
		Body  string
	}

	var pending []msg
	for rows.Next() {
		var m msg
		rows.Scan(&m.ID, &m.Phone, &m.Body)
		pending = append(pending, m)
	}

	if len(pending) == 0 {
		return
	}

	var delayStr string
	database.DB.QueryRow("SELECT value FROM settings WHERE key='sms_delay_ms'").Scan(&delayStr)
	delayMs, err := strconv.Atoi(delayStr)
	if err != nil || delayMs < 0 {
		delayMs = 1000
	}
	delay := time.Duration(delayMs) * time.Millisecond

	log.Printf("Scheduler: sending %d pending messages", len(pending))

	for i, m := range pending {
		result := gateway.Send(m.Phone, m.Body)
		if result.Success {
			now := time.Now().Format("2006-01-02 15:04:05")
			database.DB.Exec("UPDATE messages SET status='sent', gateway_used=?, sent_at=?, error_message=NULL WHERE id=?",
				result.GatewayUsed, now, m.ID)
			log.Printf("Scheduler: sent message %d to %s via %s", m.ID, m.Phone, result.GatewayUsed)
		} else {
			database.DB.Exec("UPDATE messages SET status='failed', error_message=? WHERE id=?",
				result.Error, m.ID)
			log.Printf("Scheduler: failed message %d to %s: %s", m.ID, m.Phone, result.Error)
		}

		if i < len(pending)-1 {
			time.Sleep(delay)
		}
	}
}
