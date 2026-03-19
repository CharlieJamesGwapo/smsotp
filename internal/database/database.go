package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) {
	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	migrate()
}

func migrate() {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS admin (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY NOT NULL,
			value TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			email TEXT,
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS groups_ (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS contact_groups (
			group_id INTEGER NOT NULL,
			contact_id INTEGER NOT NULL,
			PRIMARY KEY (group_id, contact_id),
			FOREIGN KEY (group_id) REFERENCES groups_(id) ON DELETE CASCADE,
			FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_id INTEGER,
			contact_id INTEGER,
			phone TEXT NOT NULL,
			body TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			gateway_used TEXT,
			scheduled_at DATETIME,
			sent_at DATETIME,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (template_id) REFERENCES templates(id),
			FOREIGN KEY (contact_id) REFERENCES contacts(id)
		)`,
	}

	for _, t := range tables {
		if _, err := DB.Exec(t); err != nil {
			log.Fatal("Migration failed:", err)
		}
	}

	seedDefaults()
}

func seedDefaults() {
	defaults := map[string]string{
		"phone_gateway_url":     "http://192.168.100.165:8080",
		"semaphore_api_key":     "",
		"semaphore_enabled":     "false",
		"semaphore_sender_name": "",
		"sms_delay_ms":          "1000",
	}
	for k, v := range defaults {
		DB.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)", k, v)
	}
}
