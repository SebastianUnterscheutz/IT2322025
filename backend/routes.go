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
		http.Error(w, `{"status":"error","message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var offer Offer

	// JSON-Daten aus dem Request-Body einlesen
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid input data",
		})
		return
	}

	offer.Token = fmt.Sprintf("%s-%s-%s-%s", randomString(4), randomString(4), randomString(4), randomString(3))
	// Validate required fields are not empty
	if offer.Name == "" || offer.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Missing required fields",
		})
		return
	}

	// Validate email format
	if !isValidEmail(offer.Email) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid email format",
		})
		return
	}

	// Validate phone number length
	if offer.PhoneNumber != "" && (len(offer.PhoneNumber) < 10 || len(offer.PhoneNumber) > 15) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid phone number",
		})
		return
	}

	validFrom, err := time.Parse("2006-01-02", offer.ValidFrom)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid date format for valid_from",
		})
		return
	}

	validUntil, err := time.Parse("2006-01-02", offer.ValidUntil)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid date format for valid_until",
		})
		return
	}

	// Validate 'ValidFrom' and 'ValidUntil' fields
	if validFrom.IsZero() || validUntil.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Both valid_from and valid_until must be provided",
		})
		return
	}

	// Ensure 'ValidFrom' is before 'ValidUntil'
	if !validFrom.Before(validUntil) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "valid_from must be before valid_until",
		})
		return
	}

	if offer.OfferLocations == nil {
		log.Println("No locations provided")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "No locations provided",
		})
		return
	}

	if len(offer.OfferLocations) >= 20 {
		log.Println("Too many locations provided")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Too many locations provided",
		})
		return
	}

	for lid, location := range offer.OfferLocations {
		if location.PLZ != "" && location.City != "" {
			address := location.Street + " " + location.HouseNumber + ", " + location.City + ", " + location.PLZ
			lat, lng, err := getCoordinates(address)
			if err != nil {
				log.Printf("Error: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": "Could not get coordinates",
				})
				return
			}
			offer.OfferLocations[lid].Latitude = lat
			offer.OfferLocations[lid].Longitude = lng
		} else {
			plz, city, err := getAdressFromCoordinates(offer.OfferLocations[lid].Latitude, offer.OfferLocations[lid].Longitude)
			if err != nil {
				log.Printf("Error: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"status":  "error",
					"message": "Could not get address",
				})
				return
			}
			offer.OfferLocations[lid].PLZ = plz
			offer.OfferLocations[lid].City = city
		}

		if offer.OfferLocations[lid].Latitude == 0 && offer.OfferLocations[lid].Longitude == 0 && offer.OfferLocations[lid].PLZ == "" && offer.OfferLocations[lid].City == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Invalid coordinates OR PLZ and CITY",
			})
			return
		}
	}

	query := `
		INSERT INTO rides (
			name, first_name, email, class, phone_number, valid_from, valid_until, 
			additional_information, other, token, activated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := dbCon.Prepare(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Could not prepare SQL statement",
		})
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
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Could not execute SQL statement",
		})
		log.Printf("Error executing query: %v", err)
		return
	}

	rideID, err := res.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Could not retrieve last insert ID for rides",
		})
		log.Printf("Error retrieving last insert ID: %v", err)
		return
	}

	for _, location := range offer.OfferLocations {
		insertLocationSQL := `
			INSERT INTO locations_on_the_way (
			rides_id, plz, city, street, house_number, latitude, longitude
			) VALUES (?, ?, ?, ?, ?, ?, ?)`
		stmtLocation, err := dbCon.Prepare(insertLocationSQL)

		_, err = stmtLocation.Exec(
			rideID,
			location.PLZ,
			location.City,
			location.Street,
			location.HouseNumber,
			location.Latitude,
			location.Longitude,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "Could not prepare insert statement for locations",
			})
			log.Printf("Error preparing insert statement for locations: %v", err)
			return
		}
		defer stmtLocation.Close()
	}

	if err := sendActivationEmail(offer.Email, offer.Token); err != nil {
		log.Println("E-Mail konnte nicht gesendet werden:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Offer created but email could not be sent",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Offer created successfully",
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
	// Content-Type setzen
	w.Header().Set("Content-Type", "application/json")

	// Suchparameter aus der URL parsen
	plz := r.URL.Query().Get("plz")
	city := r.URL.Query().Get("city")

	// SQL-Bedingungen für unscharfe Suche
	query := `
		SELECT l.id, l.rides_id, l.plz, l.city, l.street, l.house_number, l.latitude, l.longitude, 
			   r.name, r.first_name, r.email, r.class, r.phone_number, r.valid_from, r.valid_until,
			   r.additional_information, r.other, r.token, r.activated
		FROM locations_on_the_way l
		JOIN rides r ON l.rides_id = r.id
		WHERE l.plz LIKE ? AND l.city LIKE ?`

	// Suchmuster für unscharfe Suche vorbereiten
	plzPattern := "%" + plz + "%"
	cityPattern := "%" + city + "%"

	// Daten abfragen
	rows, err := dbCon.Query(query, plzPattern, cityPattern)
	if err != nil {
		log.Printf("Datenbankfehler: %v", err)
		http.Error(w, "Fehler bei der Suche", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Ergebnisse speichern
	var results []OfferLocations
	for rows.Next() {
		var location OfferLocations
		var ride Offer

		err := rows.Scan(
			&location.ID, &location.RidesID, &location.PLZ, &location.City, &location.Street, &location.HouseNumber,
			&location.Latitude, &location.Longitude, &ride.Name, &ride.FirstName, &ride.Email, &ride.Class,
			&ride.PhoneNumber, &ride.ValidFrom, &ride.ValidUntil, &ride.AdditionalInformation,
			&ride.Other, &ride.Token, &ride.Activated,
		)
		if err != nil {
			log.Printf("Fehler beim Lesen der Ergebnisse: %v", err)
			http.Error(w, "Fehler beim Verarbeiten der Ergebnisse", http.StatusInternalServerError)
			return
		}

		location.Ride = &ride
		results = append(results, location)
	}

	// Wenn keine Ergebnisse vorliegen, leeres Array zurückgeben
	if len(results) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	// Ergebnisse als JSON kodieren und zurückgeben
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Fehler bei der Antwortkodierung: %v", err)
		http.Error(w, "Fehler bei der Ausgabe der Ergebnisse", http.StatusInternalServerError)
	}
}

func editOffer(w http.ResponseWriter, r *http.Request) {
	// Unterstützt nur GET und POST
	if r.Method == http.MethodGet {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token is missing", http.StatusBadRequest)
			return
		}

		// Eintrag abrufen
		var offer Offer
		err := dbCon.QueryRow("SELECT name, first_name, email, class, phone_number, valid_from, valid_until, additional_information FROM rides WHERE token = ?", token).Scan(
			&offer.Name, &offer.FirstName, &offer.Email, &offer.Class, &offer.PhoneNumber, &offer.ValidFrom, &offer.ValidUntil, &offer.AdditionalInformation,
		)

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Offer not found", http.StatusNotFound)
			return
		}

		// JSON-Antwort mit den Angebotsdaten
		json.NewEncoder(w).Encode(offer)
		return
	} else if r.Method == http.MethodPost {
		var offer Offer
		if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
			http.Error(w, "Invalid input data", http.StatusBadRequest)
			return
		}

		// Eingabe validieren
		if offer.Name == "" || offer.Email == "" || !isValidEmail(offer.Email) {
			http.Error(w, "Invalid input data", http.StatusBadRequest)
			return
		}

		// Angebot aktualisieren
		_, err := dbCon.Exec("UPDATE rides SET name = ?, first_name = ?, email = ?, class = ?, phone_number = ?, valid_from = ?, valid_until = ?, additional_information = ? WHERE token = ?",
			offer.Name, offer.FirstName, offer.Email, offer.Class, offer.PhoneNumber, offer.ValidFrom, offer.ValidUntil, offer.AdditionalInformation, offer.Token)
		if err != nil {
			http.Error(w, "Failed to update offer", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Offer updated successfully"))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func activateOffer(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	// Eintrag aktivieren
	_, err := dbCon.Exec("UPDATE rides SET activated = TRUE WHERE token = ?", token)
	if err != nil {
		http.Error(w, "Failed to activate offer", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Offer activated successfully"))
}
