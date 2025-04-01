package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Offer beschreibt die Datenstruktur eines Angebots
type Offer struct {
	Name                  string    `json:"name"`
	FirstName             string    `json:"first_name"`
	Email                 string    `json:"email"`
	Class                 string    `json:"class"`
	PhoneNumber           string    `json:"phone_number"`
	ValidFrom             time.Time `json:"valid_from"`
	ValidUntil            time.Time `json:"valid_until"`
	AdditionalInformation string    `json:"additional_information"`
	Other                 string    `json:"other"`
	Token                 string    `json:"token"`
	Activated             bool      `json:"activated"`
}

// Globale Datenbankverbindung
var dbCon *sql.DB

// createOffer verarbeitet eine POST-Anfrage, um ein Angebot zu erstellen
func createOffer(w http.ResponseWriter, r *http.Request) {
	// Nur POST-Anfragen zulassen
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var offer Offer

	// JSON-Daten aus dem Request-Body einlesen
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	// Prepared Statement für die Datenbankeintragung
	query := `
		INSERT INTO rides (
			name, first_name, email, class, phone_number, valid_from, valid_until, 
			additional_information, other, token, activated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Query ausführen
	stmt, err := dbCon.Prepare(query)
	if err != nil {
		http.Error(w, "Could not prepare SQL statement", http.StatusInternalServerError)
		log.Printf("Error preparing query: %v", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		offer.Name,
		offer.FirstName,
		offer.Email,
		offer.Class,
		offer.PhoneNumber,
		offer.ValidFrom,
		offer.ValidUntil,
		offer.AdditionalInformation,
		offer.Other,
		offer.Token,
		offer.Activated,
	)
	if err != nil {
		http.Error(w, "Could not execute SQL statement", http.StatusInternalServerError)
		log.Printf("Error executing query: %v", err)
		return
	}

	// Erfolgsmeldung zurückgeben
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Offer created successfully",
		"status":  "success",
	})
}

func getOffer(w http.ResponseWriter, r *http.Request) {
	// Überprüfe, ob die Methode GET ist
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Beispiel-Datenstruktur
	exampleData := map[string]string{
		"message": "Hello, world!",
		"status":  "success",
	}

	// Setze den Content-Typ auf JSON
	w.Header().Set("Content-Type", "application/json")

	// Wandle die Datenstruktur in JSON um und sende sie
	if err := json.NewEncoder(w).Encode(exampleData); err != nil {
		http.Error(w, "Could not encode JSON", http.StatusInternalServerError)
		return
	}
}
