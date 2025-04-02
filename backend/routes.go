package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Offer beschreibt die Datenstruktur eines Angebots
type Offer struct {
	Name                  string           `json:"name"`
	FirstName             string           `json:"first_name"`
	Email                 string           `json:"email"`
	Class                 string           `json:"class"`
	PhoneNumber           string           `json:"phone_number"`
	ValidFrom             string           `json:"valid_from"`
	ValidUntil            string           `json:"valid_until"`
	AdditionalInformation string           `json:"additional_information"`
	Other                 string           `json:"other"`
	Token                 string           `json:"token"`
	Activated             bool             `json:"activated"`
	OfferLocations        []OfferLocations `json:"offer_locations"`
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
	Ride        *Offer  `json:"ride"`
}

// Globale Datenbankverbindung
var dbCon *sql.DB

// createOffer handles requests to create a new offer and insert data into the database.
// It validates the input, including required fields, email format, and date constraints.
// The function also processes and inserts related location data and returns a success message upon completion.
func createOffer(w http.ResponseWriter, r *http.Request) {
	// Nur POST-Anfragen zulassen
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var offer Offer

	// JSON-Daten aus dem Request-Body einlesen
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	offer.Token = fmt.Sprintf("%s-%s-%s-%s", randomString(4), randomString(4), randomString(4), randomString(3))
	offer.Activated = true
	// Validate required fields are not empty
	if offer.Name == "" || offer.Email == "" {
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

	validFrom, err := time.Parse("2006-01-02", offer.ValidFrom)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	validUntil, err := time.Parse("2006-01-02", offer.ValidUntil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Validate 'ValidFrom' and 'ValidUntil' fields
	if validFrom.IsZero() || validFrom.IsZero() {
		http.Error(w, "Both valid_from and valid_until must be provided", http.StatusBadRequest)
		return
	}

	// Ensure 'ValidFrom' is before 'ValidUntil'
	if !validFrom.Before(validUntil) {
		http.Error(w, "valid_from must be before valid_until", http.StatusBadRequest)
		return
	}

	if offer.OfferLocations == nil {
		fmt.Println("No locations provided")
		http.Error(w, "No locations provided", http.StatusBadRequest)
		return
	}

	if len(offer.OfferLocations) >= 20 {
		fmt.Println("Too many locations provided")
		http.Error(w, "Too many locations provided", http.StatusBadRequest)
		return
	}

	for lid, location := range offer.OfferLocations {

		if location.PLZ != "" && location.City != "" {
			address := location.Street + " " + location.HouseNumber + ", " + location.City + ", " + location.PLZ
			lat, lng, err := getCoordinates(address)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				http.Error(w, "Could not get coordinates", http.StatusInternalServerError)
				return
			} else {
				fmt.Printf("Latitude: %f, Longitude: %f\n", lat, lng)
			}

			offer.OfferLocations[lid].Latitude = lat
			offer.OfferLocations[lid].Longitude = lng
		}

		if offer.OfferLocations[lid].Latitude == 0 && offer.OfferLocations[lid].Longitude == 0 {
			http.Error(w, "Invalid coordinates", http.StatusBadRequest)
			return
		}

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
		validFrom,
		validUntil,
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

		insertLocationSQL := `
			INSERT INTO locations_on_the_way (
			rides_id, plz, city, street, house_number, latitude, longitude
			) VALUES (?, ?, ?, ?, ?, ?, ?)`

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

func getOffer(w http.ResponseWriter, r *http.Request) {
	// Überprüfe, ob die Methode GET ist
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Abfrage der Daten aus den Tabellen locations_on_the_way und rides
	query := `
	SELECT 
		l.id, l.rides_id, l.plz, l.city, l.street, l.house_number, l.latitude, l.longitude,
		r.name, r.first_name, r.email, r.class, r.phone_number, r.valid_from, r.valid_until,
		r.additional_information, r.other, r.token, r.activated
	FROM locations_on_the_way l
	JOIN rides r ON l.rides_id = r.id
	`
	rows, err := dbCon.Query(query)
	if err != nil {
		http.Error(w, "Could not query locations and rides", http.StatusInternalServerError)
		log.Printf("Error querying locations and rides: %v", err)
		return
	}
	defer rows.Close()

	// Struktur, die sowohl Locations als auch die zugehörigen Ride-Informationen enthält
	type LocationWithRide struct {
		OfferLocations
		Ride *Offer `json:"ride"`
	}

	locationsWithRides := []LocationWithRide{}
	for rows.Next() {
		var location LocationWithRide
		location.Ride = &Offer{} // Ensure the Ride field is initialized (not nil)
		if err := rows.Scan(
			&location.ID, &location.RidesID, &location.PLZ, &location.City, &location.Street, &location.HouseNumber,
			&location.Latitude, &location.Longitude, &location.Ride.Name, &location.Ride.FirstName, &location.Ride.Email,
			&location.Ride.Class, &location.Ride.PhoneNumber, &location.Ride.ValidFrom, &location.Ride.ValidUntil,
			&location.Ride.AdditionalInformation, &location.Ride.Other, &location.Ride.Token, &location.Ride.Activated,
		); err != nil {
			http.Error(w, "Could not scan location and ride data", http.StatusInternalServerError)
			log.Printf("Error scanning row: %v", err)
			return
		}

		locationsWithRides = append(locationsWithRides, location)

	}

	// Setze den Content-Typ auf JSON
	w.Header().Set("Content-Type", "application/json")

	// Wandle die gesammelten Daten in JSON um und sende sie
	if err := json.NewEncoder(w).Encode(locationsWithRides); err != nil {
		http.Error(w, "Could not encode JSON", http.StatusInternalServerError)
		return
	}
}

func searchOffers(w http.ResponseWriter, r *http.Request) {
	// Überprüfe, ob die Methode GET ist
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Lese die Query-Parameter
	plz := r.URL.Query().Get("plz")
	city := r.URL.Query().Get("city")

	query := `
		SELECT o.id, o.rides_id, o.plz, o.city, o.street, o.house_number, o.latitude, o.longitude, 
		       r.name, r.first_name, r.email, r.class, r.phone_number, r.valid_from, r.valid_until,
		       r.additional_information, r.other, r.token, r.activated
		FROM locations_on_the_way AS o
		INNER JOIN rides AS r ON o.rides_id = r.id
		WHERE o.plz LIKE ? AND o.city LIKE ?
	` // Die Abfrage sucht nach PLZ und Ort in der "locations_on_the_way"-Tabelle

	// Füge Platzhalter `%` hinzu, wenn keine Werte angegeben werden
	if plz == "" {
		plz = "%"
	}
	if city == "" {
		city = "%"
	}

	rows, err := dbCon.Query(query, plz, city)
	if err != nil {
		http.Error(w, "Could not query offers", http.StatusInternalServerError)
		log.Printf("Query error: %v", err)
		return
	}
	defer rows.Close()

	// Ergebnisse sammeln
	var results []OfferLocations
	for rows.Next() {
		var location OfferLocations
		var ride Offer

		// Lese die Ergebnisse
		if err := rows.Scan(
			&location.ID,
			&location.RidesID,
			&location.PLZ,
			&location.City,
			&location.Street,
			&location.HouseNumber,
			&location.Latitude,
			&location.Longitude,
			&ride.Name,
			&ride.FirstName,
			&ride.Email,
			&ride.Class,
			&ride.PhoneNumber,
			&ride.ValidFrom,
			&ride.ValidUntil,
			&ride.AdditionalInformation,
			&ride.Other,
			&ride.Token,
			&ride.Activated,
		); err != nil {
			http.Error(w, "Could not scan results", http.StatusInternalServerError)
			log.Printf("Error scanning results: %v", err)
			return
		}

		// Füge das Angebot zur Location hinzu
		location.Ride = &ride

		results = append(results, location)

	}

	// Ergebnisse als JSON ausgeben
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Could not encode results to JSON", http.StatusInternalServerError)
		log.Printf("Encoding error: %v", err)
	}
}
