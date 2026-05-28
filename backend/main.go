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

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" || dbURL == "postgresql://postgres:your_password@db.supabase.co:5432/postgres?sslmode=require" {
		log.Println("WARNING: DATABASE_URL environment variable is empty or is set to default placeholders.")
		log.Println("Please update backend/.env with your actual Supabase PostgreSQL connection string!")
		log.Println("Example format: postgresql://postgres:[your-password]@[your-supabase-host]:5432/postgres?sslmode=require")
	}

	// Initialize database connection
	var err error
	if dbURL != "" && dbURL != "postgresql://postgres:your_password@db.supabase.co:5432/postgres?sslmode=require" {
		_, err = InitDB(dbURL)
		if err != nil {
			log.Printf("ERROR: Failed to connect to database: %v", err)
			log.Println("FALLBACK: Starting in LOCAL FALLBACK MODE (saving data locally in backend/data/invitations.json)!")
			useLocalFallback = true
		}
	} else {
		log.Println("WARNING: No DATABASE_URL provided or using placeholders.")
		log.Println("FALLBACK: Starting in LOCAL FALLBACK MODE (saving data locally in backend/data/invitations.json)!")
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
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Critical: Server failed to start: %v", err)
	}
}
