package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type CompanyInfo struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Cuit    string `json:"cuit"`
	Address string `json:"address"`
	Email   string `json:"email"`
	Logo    string `json:"logo"`
}

type Item struct {
	Qty         float64 `json:"qty"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unitPrice"`
	Total       float64 `json:"total"`
}

type LayoutConfig struct {
	PageSize      string  `json:"pageSize"`
	WidthMm       float64 `json:"widthMm"`
	HeightMm      float64 `json:"heightMm"`
	MarginTop     float64 `json:"marginTop"`
	MarginBottom  float64 `json:"marginBottom"`
	MarginLeft    float64 `json:"marginLeft"`
	MarginRight   float64 `json:"marginRight"`
	ColQtyWidth   float64 `json:"colQtyWidth"`
	ColDescWidth  float64 `json:"colDescWidth"`
	ColPriceWidth float64 `json:"colPriceWidth"`
	ColTotalWidth float64 `json:"colTotalWidth"`
	FontSizeScale float64 `json:"fontSizeScale"`
}

type Invitation struct {
	ID              string       `json:"id"`
	Type            string       `json:"type"` // "remito" | "orden"
	DocNumber       string       `json:"docNumber"`
	Emisor          CompanyInfo  `json:"emisor"`
	Receptor        CompanyInfo  `json:"receptor"`
	EventDate       time.Time    `json:"eventDate"`
	Items           []Item       `json:"items"`
	Comments        string       `json:"comments"`
	Theme           string       `json:"theme"`           // "classic" | "coffee" | "cyberpunk" | "retro"
	Status          string       `json:"status"`          // "pending" | "accepted" | "declined"
	Signature       string       `json:"signature"`       // base64 PNG of signature
	RejectionReason string       `json:"rejectionReason"` // reason if rejected
	Layout          LayoutConfig `json:"layout"`
	CreatedAt       time.Time    `json:"createdAt"`
}

var DB *sql.DB

var (
	useLocalFallback = false
	localMutex       sync.Mutex
	localFilePath    = "data/invitations.json"
)

// Helper to save to local file
func saveLocalData(invitations []Invitation) error {
	localMutex.Lock()
	defer localMutex.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(invitations, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(localFilePath, data, 0644)
}

// Helper to read from local file
func readLocalData() ([]Invitation, error) {
	localMutex.Lock()
	defer localMutex.Unlock()

	if _, err := os.Stat(localFilePath); os.IsNotExist(err) {
		return []Invitation{}, nil
	}

	data, err := os.ReadFile(localFilePath)
	if err != nil {
		return nil, err
	}

	var invitations []Invitation
	if err := json.Unmarshal(data, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// Simple random UUID generator in pure Go to avoid external deps
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping database to verify connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("Successfully connected to Supabase PostgreSQL database!")
	DB = db

	// Auto-create table if not exists
	err = AutoMigrate(db)
	if err != nil {
		log.Printf("Warning: auto-migration failed: %v. Please make sure the table exists.", err)
	}

	return db, nil
}

func AutoMigrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS invitations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		type VARCHAR(20) NOT NULL,
		doc_number VARCHAR(50) NOT NULL,
		emisor JSONB NOT NULL,
		receptor JSONB NOT NULL,
		event_date TIMESTAMP WITH TIME ZONE NOT NULL,
		items JSONB NOT NULL,
		comments TEXT,
		theme VARCHAR(30) DEFAULT 'classic',
		status VARCHAR(20) DEFAULT 'pending',
		signature TEXT,
		rejection_reason TEXT,
		layout JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	log.Println("Database migration completed (table invitations checked/created)")
	return nil
}

func GetInvitations(db *sql.DB) ([]Invitation, error) {
	if useLocalFallback || db == nil {
		return readLocalData()
	}

	rows, err := db.Query("SELECT id, type, doc_number, emisor, receptor, event_date, items, comments, theme, status, signature, rejection_reason, layout, created_at FROM invitations ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []Invitation
	for rows.Next() {
		var inv Invitation
		var emisorJSON, receptorJSON, itemsJSON, layoutJSON []byte
		var comments, signature, rejectionReason sql.NullString

		err := rows.Scan(
			&inv.ID,
			&inv.Type,
			&inv.DocNumber,
			&emisorJSON,
			&receptorJSON,
			&inv.EventDate,
			&itemsJSON,
			&comments,
			&inv.Theme,
			&inv.Status,
			&signature,
			&rejectionReason,
			&layoutJSON,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(emisorJSON, &inv.Emisor); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(receptorJSON, &inv.Receptor); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &inv.Items); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(layoutJSON, &inv.Layout); err != nil {
			return nil, err
		}

		inv.Comments = comments.String
		inv.Signature = signature.String
		inv.RejectionReason = rejectionReason.String

		invitations = append(invitations, inv)
	}

	return invitations, nil
}

func GetInvitationByID(db *sql.DB, id string) (*Invitation, error) {
	if useLocalFallback || db == nil {
		list, err := readLocalData()
		if err != nil {
			return nil, err
		}
		for _, inv := range list {
			if inv.ID == id {
				return &inv, nil
			}
		}
		return nil, fmt.Errorf("invitation not found")
	}

	var inv Invitation
	var emisorJSON, receptorJSON, itemsJSON, layoutJSON []byte
	var comments, signature, rejectionReason sql.NullString

	err := db.QueryRow("SELECT id, type, doc_number, emisor, receptor, event_date, items, comments, theme, status, signature, rejection_reason, layout, created_at FROM invitations WHERE id = $1", id).Scan(
		&inv.ID,
		&inv.Type,
		&inv.DocNumber,
		&emisorJSON,
		&receptorJSON,
		&inv.EventDate,
		&itemsJSON,
		&comments,
		&inv.Theme,
		&inv.Status,
		&signature,
		&rejectionReason,
		&layoutJSON,
		&inv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(emisorJSON, &inv.Emisor); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(receptorJSON, &inv.Receptor); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(itemsJSON, &inv.Items); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(layoutJSON, &inv.Layout); err != nil {
		return nil, err
	}

	inv.Comments = comments.String
	inv.Signature = signature.String
	inv.RejectionReason = rejectionReason.String

	return &inv, nil
}

func CreateInvitation(db *sql.DB, inv *Invitation) error {
	if useLocalFallback || db == nil {
		list, err := readLocalData()
		if err != nil {
			return err
		}
		inv.ID = generateUUID()
		inv.CreatedAt = time.Now()
		inv.Status = "pending"
		
		// Prepend new invitation
		list = append([]Invitation{*inv}, list...)
		err = saveLocalData(list)
		if err != nil {
			return err
		}
		return nil
	}

	emisorJSON, err := json.Marshal(inv.Emisor)
	if err != nil {
		return err
	}
	receptorJSON, err := json.Marshal(inv.Receptor)
	if err != nil {
		return err
	}
	itemsJSON, err := json.Marshal(inv.Items)
	if err != nil {
		return err
	}
	layoutJSON, err := json.Marshal(inv.Layout)
	if err != nil {
		return err
	}

	var newID string
	query := `
		INSERT INTO invitations (type, doc_number, emisor, receptor, event_date, items, comments, theme, status, layout)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err = db.QueryRow(
		query,
		inv.Type,
		inv.DocNumber,
		emisorJSON,
		receptorJSON,
		inv.EventDate,
		itemsJSON,
		inv.Comments,
		inv.Theme,
		"pending",
		layoutJSON,
	).Scan(&newID, &inv.CreatedAt)

	if err != nil {
		return err
	}

	inv.ID = newID
	inv.Status = "pending"
	return nil
}

func UpdateInvitationRSVP(db *sql.DB, id string, status string, signature string, rejectionReason string) error {
	if useLocalFallback || db == nil {
		list, err := readLocalData()
		if err != nil {
			return err
		}
		found := false
		for i, inv := range list {
			if inv.ID == id {
				list[i].Status = status
				list[i].Signature = signature
				list[i].RejectionReason = rejectionReason
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invitation with id %s not found", id)
		}
		return saveLocalData(list)
	}

	var sigVal, rejVal interface{}
	if signature != "" {
		sigVal = signature
	} else {
		sigVal = nil
	}

	if rejectionReason != "" {
		rejVal = rejectionReason
	} else {
		rejVal = nil
	}

	query := `
		UPDATE invitations
		SET status = $1, signature = $2, rejection_reason = $3
		WHERE id = $4`

	result, err := db.Exec(query, status, sigVal, rejVal, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation with id %s not found", id)
	}

	return nil
}

func UpdateInvitation(db *sql.DB, id string, inv *Invitation) error {
	if useLocalFallback || db == nil {
		list, err := readLocalData()
		if err != nil {
			return err
		}
		found := false
		for i, item := range list {
			if item.ID == id {
				list[i].Type = inv.Type
				list[i].DocNumber = inv.DocNumber
				list[i].Emisor = inv.Emisor
				list[i].Receptor = inv.Receptor
				list[i].EventDate = inv.EventDate
				list[i].Items = inv.Items
				list[i].Comments = inv.Comments
				list[i].Theme = inv.Theme
				list[i].Layout = inv.Layout
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invitation with id %s not found", id)
		}
		return saveLocalData(list)
	}

	emisorJSON, err := json.Marshal(inv.Emisor)
	if err != nil {
		return err
	}
	receptorJSON, err := json.Marshal(inv.Receptor)
	if err != nil {
		return err
	}
	itemsJSON, err := json.Marshal(inv.Items)
	if err != nil {
		return err
	}
	layoutJSON, err := json.Marshal(inv.Layout)
	if err != nil {
		return err
	}

	query := `
		UPDATE invitations
		SET type = $1, doc_number = $2, emisor = $3, receptor = $4, event_date = $5, items = $6, comments = $7, theme = $8, layout = $9
		WHERE id = $10`

	result, err := db.Exec(query, inv.Type, inv.DocNumber, emisorJSON, receptorJSON, inv.EventDate, itemsJSON, inv.Comments, inv.Theme, layoutJSON, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation with id %s not found", id)
	}

	return nil
}

func DeleteInvitation(db *sql.DB, id string) error {
	if useLocalFallback || db == nil {
		list, err := readLocalData()
		if err != nil {
			return err
		}
		found := false
		var newList []Invitation
		for _, item := range list {
			if item.ID == id {
				found = true
				continue
			}
			newList = append(newList, item)
		}
		if !found {
			return fmt.Errorf("invitation with id %s not found", id)
		}
		return saveLocalData(newList)
	}

	query := `DELETE FROM invitations WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation with id %s not found", id)
	}

	return nil
}
