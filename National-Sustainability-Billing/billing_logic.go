package main

import (
	"database/sql"
	"errors"
	"log"

	_ "modernc.org/sqlite"
)

const HardwareRefreshRatio = 0.30

var DB *sql.DB

// InitDB initializes the SQLite database and creates the Ledger table
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS ledger (
		id INTEGER PRIMARY KEY,
		treasury_account REAL NOT NULL,
		hardware_refresh_fund REAL NOT NULL,
		total_transactions INTEGER NOT NULL
	);`

	_, err = DB.Exec(createTableQuery)
	if err != nil {
		return err
	}

	// Initialize the row if empty
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM ledger").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = DB.Exec("INSERT INTO ledger (id, treasury_account, hardware_refresh_fund, total_transactions) VALUES (1, 0.0, 0.0, 0)")
		if err != nil {
			return err
		}
	}

	return nil
}

// TransactionRequest payload
type TransactionRequest struct {
	ServiceID string  `json:"service_id"`
	Amount    float64 `json:"amount"`
	Client    string  `json:"client"`
}

// ProcessTransaction handles the payment using ACID SQL transactions
func ProcessTransaction(amount float64) error {
	if amount <= 0 {
		return errors.New("le montant de la transaction doit être positif")
	}

	refreshCut := amount * HardwareRefreshRatio
	treasuryCut := amount - refreshCut

	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Update the single ledger row securely (Row-level lock in a real DB, file lock in SQLite)
	updateQuery := `
		UPDATE ledger 
		SET treasury_account = treasury_account + ?, 
		    hardware_refresh_fund = hardware_refresh_fund + ?, 
		    total_transactions = total_transactions + 1 
		WHERE id = 1`

	_, err = tx.Exec(updateQuery, treasuryCut, refreshCut)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Printf("[BILLING] Transaction traitée: +%.2f BRH, +%.2f Trésor", refreshCut, treasuryCut)
	return nil
}

// GetLedgerBalance returns the current state
func GetLedgerBalance() (float64, float64, int) {
	var treasury, refresh float64
	var count int
	err := DB.QueryRow("SELECT treasury_account, hardware_refresh_fund, total_transactions FROM ledger WHERE id = 1").Scan(&treasury, &refresh, &count)
	if err != nil {
		log.Printf("Erreur lecture ledger: %v", err)
		return 0, 0, 0
	}
	return treasury, refresh, count
}
