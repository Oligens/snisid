package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Document struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"` // DRAFT, PENDING_SIG, SIGNED, REJECTED
	Signature string    `json:"signature"`
	SignerID  string    `json:"signer_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SignRequest struct {
	DocumentID string `json:"document_id"`
	SignerID   string `json:"signer_id"`
	PinCode    string `json:"pin_code"`
}

var db *gorm.DB

func initDB() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=snisid password=snisid dbname=executive_ops port=5432 sslmode=disable"
	}
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v, continuing with mock DB if needed", err)
		return
	}
	db.AutoMigrate(&Document{})
	log.Println("Database connected and migrated.")
}

func SignDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve document (simulated if no DB, else real)
	var doc Document
	if db != nil {
		if err := db.First(&doc, "id = ?", req.DocumentID).Error; err != nil {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
	} else {
		// Mock document if DB fails
		doc = Document{
			ID:      req.DocumentID,
			Title:   "Arrete Presidentiel Mock",
			Content: "Le Président de la République arrête...",
			Status:  "PENDING_SIG",
		}
	}

	if doc.Status == "SIGNED" {
		http.Error(w, "Document is already signed", http.StatusConflict)
		return
	}

	// 1. PKI Smartcard Emulation (Hardware Simulation)
	// In a real scenario, the document hash is signed by the Smartcard's private key.
	// Here, we simulate the validation.
	if req.PinCode != "1234" { // Mock PIN validation
		http.Error(w, "Invalid Smartcard PIN", http.StatusUnauthorized)
		return
	}

	// Compute document hash (SHA-256)
	hash := sha256.New()
	hash.Write([]byte(doc.Content))
	docHash := hex.EncodeToString(hash.Sum(nil))

	// Emulate QES (Qualified Electronic Signature) sealing
	mockSignature := "QES_SEALED_" + req.SignerID + "_" + docHash

	doc.Status = "SIGNED"
	doc.Signature = mockSignature
	doc.SignerID = req.SignerID
	doc.UpdatedAt = time.Now()

	if db != nil {
		db.Save(&doc)
	}

	// 2. Integration Phase 11 : Notifier le moteur BPMN
	bpmn := NewBPMNClient()
	_ = bpmn.NotifySignature(doc.ID, req.SignerID, mockSignature)

	// 3. Integration Phase 5 : Stockage Inaltérable (WORM)
	worm := NewWORMClient()
	// Dans un cas réel, on passe le PDF/A généré. Ici on passe le contenu en bytes.
	_ = worm.ArchiveDocument(doc.ID, []byte(doc.Content), docHash)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var doc Document
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	doc.Status = "DRAFT"
	doc.CreatedAt = time.Now()

	if db != nil {
		db.Create(&doc)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	if db != nil {
		db.Find(&docs)
	} else {
		docs = append(docs, Document{ID: "DOC-1", Title: "Mock Doc", Status: "PENDING_SIG"})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func main() {
	initDB()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Post("/documents", CreateDocumentHandler)
		r.Get("/documents", ListDocumentsHandler)
		r.Post("/sign", SignDocumentHandler)
	})

	log.Println("Executive Operations API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
