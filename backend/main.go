package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	// Setze den Pfad für die statischen Dateien
	staticDir := "./frontend"

	// Erzeuge einen FileServer für das statische Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))

	// Routen Sie alle Anfragen, die mit /static/ beginnen, an den FileServer
	http.Handle("/", http.StripPrefix("/", fileServer))

	// Starte den Server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
