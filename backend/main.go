package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (ignore error if it doesn't)
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	const placeholderDBURL = "postgresql://postgres:your_password@db.supabase.co:5432/postgres?sslmode=require"
	dbURL := os.Getenv("DATABASE_URL")

	// Decide storage backend. An empty/placeholder DATABASE_URL is a valid,
	// intentional configuration: the app runs in LOCAL mode and persists data
	// to backend/data/invitations.json. The whole circuit (CRUD, RSVP, PDF)
	// works the same in either mode.
	if dbURL == "" || dbURL == placeholderDBURL {
		log.Println("LOCAL mode: no DATABASE_URL set — data is stored in backend/data/invitations.json.")
		log.Println("To use Supabase instead, set DATABASE_URL in backend/.env and restart.")
		useLocalFallback = true
	} else if _, err := InitDB(dbURL); err != nil {
		log.Printf("WARNING: could not connect to the database: %v", err)
		log.Println("FALLBACK: continuing in LOCAL mode — data is stored in backend/data/invitations.json.")
		useLocalFallback = true
	}

	// Create Router using Go 1.22+ ServeMux routing capabilities
	mux := http.NewServeMux()

	// Register API endpoints
	mux.HandleFunc("GET /api/invitations", enableCORS(requireDB(handleInvitations)))
	mux.HandleFunc("POST /api/invitations", enableCORS(requireDB(handleInvitations)))
	mux.HandleFunc("GET /api/invitations/{id}", enableCORS(requireDB(handleInvitationDetail)))
	mux.HandleFunc("PUT /api/invitations/{id}", enableCORS(requireDB(handleUpdateInvitation)))
	mux.HandleFunc("DELETE /api/invitations/{id}", enableCORS(requireDB(handleDeleteInvitation)))
	mux.HandleFunc("PUT /api/invitations/{id}/rsvp", enableCORS(requireDB(handleInvitationRSVP)))
	mux.HandleFunc("GET /api/invitations/{id}/pdf", enableCORS(requireDB(handleInvitationPDF)))

	// CORS support for preflight OPTIONS requests on all routes
	mux.HandleFunc("OPTIONS /api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Critical: Server failed to start: %v", err)
	}
}
