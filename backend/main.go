package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// Setze das Verzeichnis für die statischen Dateien
	staticDir := "./frontend"

	// Erzeuge einen FileServer für das Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))

	// Routen Sie alle Anfragen an den FileServer
	http.Handle("/", fileServer)

	// Registriere die Route /api/rides mit dem Handler handleRoute
	http.HandleFunc("/api/rides", handleRoute)

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}

func handleRoute(w http.ResponseWriter, r *http.Request) {
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
