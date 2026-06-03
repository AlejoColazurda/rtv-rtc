package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

// Helper to write JSON error responses
func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Helper to write JSON success responses
func writeJSONResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// Enable CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Require database connection middleware to prevent nil pointer panics
func requireDB(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DB == nil && !useLocalFallback {
			writeJSONError(w, "La base de datos de Supabase no está conectada. Por favor, configura DATABASE_URL en backend/.env", http.StatusServiceUnavailable)
			return
		}
		next(w, r)
	}
}

// handleInvitations handles GET /api/invitations and POST /api/invitations
func handleInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		invitations, err := GetInvitations(DB)
		if err != nil {
			log.Printf("Error fetching invitations: %v", err)
			writeJSONError(w, "Failed to retrieve invitations", http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, invitations, http.StatusOK)
		return
	}

	if r.Method == http.MethodPost {
		var inv Invitation
		err := json.NewDecoder(r.Body).Decode(&inv)
		if err != nil {
			writeJSONError(w, "Invalid invitation payload", http.StatusBadRequest)
			return
		}

		// Validation
		if inv.Type != "remito" && inv.Type != "orden" {
			writeJSONError(w, "Type must be 'remito' or 'orden'", http.StatusBadRequest)
			return
		}

		err = CreateInvitation(DB, &inv)
		if err != nil {
			log.Printf("Error creating invitation: %v", err)
			writeJSONError(w, "Failed to create invitation", http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, inv, http.StatusCreated)
		return
	}

	writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleInvitationDetail handles GET /api/invitations/{id}
func handleInvitationDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	inv, err := GetInvitationByID(DB, id)
	if err != nil {
		log.Printf("Error fetching invitation %s: %v", id, err)
		writeJSONError(w, "Invitation not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, inv, http.StatusOK)
}

type RSVPRequest struct {
	Status          string `json:"status"` // "accepted" | "declined"
	Signature       string `json:"signature"`
	RejectionReason string `json:"rejectionReason"`
}

// handleInvitationRSVP handles PUT /api/invitations/{id}/rsvp
func handleInvitationRSVP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPut {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rsvp RSVPRequest
	err := json.NewDecoder(r.Body).Decode(&rsvp)
	if err != nil {
		writeJSONError(w, "Invalid RSVP payload", http.StatusBadRequest)
		return
	}

	if rsvp.Status != "accepted" && rsvp.Status != "declined" {
		writeJSONError(w, "Status must be 'accepted' or 'declined'", http.StatusBadRequest)
		return
	}

	err = UpdateInvitationRSVP(DB, id, rsvp.Status, rsvp.Signature, rsvp.RejectionReason)
	if err != nil {
		log.Printf("Error updating RSVP for %s: %v", id, err)
		writeJSONError(w, "Failed to update RSVP status", http.StatusInternalServerError)
		return
	}

	// Fetch updated invitation to return
	updatedInv, err := GetInvitationByID(DB, id)
	if err != nil {
		writeJSONResponse(w, map[string]string{"message": "RSVP updated successfully"}, http.StatusOK)
		return
	}

	writeJSONResponse(w, updatedInv, http.StatusOK)
}

// handleInvitationPDF handles GET /api/invitations/{id}/pdf
func handleInvitationPDF(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodGet {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	inv, err := GetInvitationByID(DB, id)
	if err != nil {
		log.Printf("Error fetching invitation for PDF %s: %v", id, err)
		writeJSONError(w, "Invitation not found", http.StatusNotFound)
		return
	}

	// Render into a buffer first so a generation failure returns a clean error
	// instead of streaming a half-written / 0-byte file to the client.
	var buf bytes.Buffer
	if err := GenerateInvitationPDF(inv, &buf); err != nil {
		log.Printf("Error generating PDF for %s: %v", id, err)
		writeJSONError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	filename := "remito_" + inv.DocNumber + ".pdf"
	if inv.Type == "orden" {
		filename = "orden_compra_" + inv.DocNumber + ".pdf"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(buf.Bytes())
}

// handleUpdateInvitation handles PUT /api/invitations/{id}
func handleUpdateInvitation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPut {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var inv Invitation
	err := json.NewDecoder(r.Body).Decode(&inv)
	if err != nil {
		writeJSONError(w, "Invalid invitation payload", http.StatusBadRequest)
		return
	}

	// Validation
	if inv.Type != "remito" && inv.Type != "orden" {
		writeJSONError(w, "Type must be 'remito' or 'orden'", http.StatusBadRequest)
		return
	}

	err = UpdateInvitation(DB, id, &inv)
	if err != nil {
		log.Printf("Error updating invitation %s: %v", id, err)
		writeJSONError(w, "Failed to update invitation", http.StatusInternalServerError)
		return
	}

	// Fetch updated invitation to return
	updatedInv, err := GetInvitationByID(DB, id)
	if err != nil {
		writeJSONResponse(w, map[string]string{"message": "Invitation updated successfully"}, http.StatusOK)
		return
	}

	writeJSONResponse(w, updatedInv, http.StatusOK)
}

// handleDeleteInvitation handles DELETE /api/invitations/{id}
func handleDeleteInvitation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, "Missing invitation ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodDelete {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := DeleteInvitation(DB, id)
	if err != nil {
		log.Printf("Error deleting invitation %s: %v", id, err)
		writeJSONError(w, "Failed to delete invitation", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{"message": "Invitation deleted successfully"}, http.StatusOK)
}
