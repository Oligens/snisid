package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Engineer represents an IT staff member of the elite unit
type Engineer struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Role            string    `json:"role"`
	Status          string    `json:"status"`    // "HORS-GRILLE"
	Clearance       string    `json:"clearance"` // "SECRET", "CONFIDENTIAL"
	Certifications  []string  `json:"certifications"`
	ContractEndDate time.Time `json:"contract_end_date"` // 3 years engagement
}

var engineersDB = make(map[string]Engineer)

func enrollHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name           string   `json:"name"`
		Role           string   `json:"role"`
		Certifications []string `json:"certifications"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	engID := fmt.Sprintf("ENG-%d", time.Now().Unix())

	// Phase 16 Policy: Engagement de 3 ans
	contractEnd := time.Now().AddDate(3, 0, 0)

	eng := Engineer{
		ID:              engID,
		Name:            req.Name,
		Role:            req.Role,
		Status:          "HORS-GRILLE", // Statut d'élite
		Clearance:       "SECRET",      // Clearance requise
		Certifications:  req.Certifications,
		ContractEndDate: contractEnd,
	}

	engineersDB[engID] = eng
	log.Printf("[HR] Ingénieur %s enrôlé. Statut: %s, Clearance: %s", eng.Name, eng.Status, eng.Clearance)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eng)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var list []Engineer
	for _, e := range engineersDB {
		list = append(list, e)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func main() {
	http.HandleFunc("/api/hr/enroll", enrollHandler)
	http.HandleFunc("/api/hr/engineers", listHandler)

	fmt.Println("SNISID Human Capacity API (Phase 16) démarré sur le port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
