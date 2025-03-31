package main

import (
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

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
