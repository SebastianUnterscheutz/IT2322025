package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	rand2 "math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	OfferLocations        []OfferLocations
}

// Offer Locations
type OfferLocations struct {
	ID          int     `json:"id"`
	RidesID     int     `json:"rides_id"`
	PLZ         string  `json:"plz"`
	City        string  `json:"city"`
	Street      string  `json:"street"`
	HouseNumber string  `json:"house_number"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
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

	// Überprüfen, ob alle Pflichtfelder ausgefüllt sind
	if offer.Name == "" || offer.Email == "" || offer.Class == "" {
		http.Error(w, "Missing required fields (PLZ, Ort, Name, E-Mail, Klasse)", http.StatusBadRequest)
		return
	}

	offer.Token = fmt.Sprintf("%s-%s-%s-%s", randomString(4), randomString(4), randomString(4), randomString(3))
	offer.Activated = false
	// Validate required fields are not empty
	if offer.Name == "" || offer.FirstName == "" || offer.Email == "" || offer.Class == "" ||
		offer.PhoneNumber == "" || offer.Token == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate email format
	if !isValidEmail(offer.Email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate phone number length (e.g., you can adjust as per your system's requirement)
	if len(offer.PhoneNumber) < 10 || len(offer.PhoneNumber) > 15 {
		http.Error(w, "Invalid phone number", http.StatusBadRequest)
		return
	}

	// Validate 'ValidFrom' and 'ValidUntil' fields
	if offer.ValidFrom.IsZero() || offer.ValidUntil.IsZero() {
		http.Error(w, "Both valid_from and valid_until must be provided", http.StatusBadRequest)
		return
	}

	// Ensure 'ValidFrom' is before 'ValidUntil'
	if !offer.ValidFrom.Before(offer.ValidUntil) {
		http.Error(w, "valid_from must be before valid_until", http.StatusBadRequest)
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

	res, err := stmt.Exec(
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

	rideID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Could not retrieve last insert ID for rides", http.StatusInternalServerError)
		log.Printf("Error retrieving last insert ID: %v", err)
		return
	}

	defer stmt.Close()

	for _, location := range offer.OfferLocations {

		// Example usage
		address := location.Street + " " + location.HouseNumber + ", " + location.City + ", " + location.PLZ
		lat, lng, err := getCoordinates(address)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Latitude: %f, Longitude: %f\n", lat, lng)
		}

		insertLocationSQL := `
			INSERT INTO locations_on_the_way (
			rides_id, plz, city, street, house_number, latitude, longitude
			) VALUES (?, ?, ?, ?, ?)`

		location.Latitude = lat
		location.Longitude = lng

		stmtLocation, err := dbCon.Prepare(insertLocationSQL)

		// Füge Eintrag basierend auf Informationen im `offer` ein
		_, err = stmtLocation.Exec(
			rideID,               // Beispiel rides_id, anpassen, um relevante ID zu übernehmen
			location.PLZ,         // PLZ aus dem offer Struct
			location.City,        // Ort aus dem offer Struct
			location.Street,      // Straße aus dem offer Struct
			location.HouseNumber, // Hausnummer (optional, falls nicht im Struct)
			location.Latitude,
			location.Longitude,
		)

		if err != nil {
			http.Error(w, "Could not prepare insert statement for locations", http.StatusInternalServerError)
			log.Printf("Error preparing insert statement for locations: %v", err)
			return
		}
		defer stmtLocation.Close()
	}

	// Erfolgsmeldung zurückgeben
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Offer created successfully",
		"status":  "success",
	})
}

func isValidEmail(email string) bool {

	validEmailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Validiert eine E-Mail-Adresse anhand von Regex
	return validEmailRegex.MatchString(email)
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

func randomString(n int) string {
	if n <= 0 {
		panic("randomString: length must be greater than 0")
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func getCoordinates(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := fmt.Sprintf("?q=%s&format=json&limit=1", url.QueryEscape(address))

	// Build the full URL
	fullURL := baseURL + params

	// Create an HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse the JSON response
	var data []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// If no data is found, return nil coordinates
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("no results found for the given address")
	}

	// Convert latitude and longitude strings to float64
	lat, err := strconv.ParseFloat(data[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(data[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %w", err)
	}

	return lat, lon, nil
}
