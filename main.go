package main

import (
	"log"
	"net/http"
	"os"

	"smsotp/internal/auth"
	"smsotp/internal/contacts"
	"smsotp/internal/dashboard"
	"smsotp/internal/groups"
	"smsotp/internal/messages"
	"smsotp/internal/scheduler"
	"smsotp/internal/settings"
	"smsotp/internal/templates"

	"smsotp/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	database.Init("sms.db")
	auth.EnsureAdmin()
	scheduler.Start()

	// Get allowed frontend origin from env (for Vercel)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL, "http://localhost:5173", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health check for Render
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Public routes
	r.Post("/api/auth/login", auth.HandleLogin)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)

		// Auth
		r.Get("/api/auth/me", auth.HandleMe)
		r.Put("/api/auth/password", auth.HandleChangePassword)

		// Dashboard
		r.Get("/api/dashboard/stats", dashboard.HandleStats)

		// Contacts
		r.Get("/api/contacts", contacts.HandleList)
		r.Post("/api/contacts", contacts.HandleCreate)
		r.Put("/api/contacts/{id}", contacts.HandleUpdate)
		r.Delete("/api/contacts/{id}", contacts.HandleDelete)
		r.Post("/api/contacts/import", contacts.HandleImport)

		// Groups
		r.Get("/api/groups", groups.HandleList)
		r.Get("/api/groups/{id}", groups.HandleGet)
		r.Post("/api/groups", groups.HandleCreate)
		r.Put("/api/groups/{id}", groups.HandleUpdate)
		r.Delete("/api/groups/{id}", groups.HandleDelete)
		r.Post("/api/groups/{id}/members", groups.HandleAddMembers)
		r.Delete("/api/groups/{id}/members", groups.HandleRemoveMembers)

		// Templates
		r.Get("/api/templates", templates.HandleList)
		r.Post("/api/templates", templates.HandleCreate)
		r.Put("/api/templates/{id}", templates.HandleUpdate)
		r.Delete("/api/templates/{id}", templates.HandleDelete)

		// Messages
		r.Post("/api/messages/send", messages.HandleSend)
		r.Post("/api/messages/schedule", messages.HandleSchedule)
		r.Get("/api/messages", messages.HandleList_)
		r.Get("/api/messages/{id}", messages.HandleGet)
		r.Delete("/api/messages/{id}", messages.HandleCancel)
		r.Post("/api/messages/{id}/retry", messages.HandleRetry)

		// Settings
		r.Get("/api/settings", settings.HandleGet)
		r.Put("/api/settings", settings.HandleUpdate)
	})

	// Get port from env (Render sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("SMS Notification System running on port %s", port)
	log.Println("Default login: admin / admin123")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
