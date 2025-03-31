package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	staticDir := "./frontend"

	// Erzeuge einen FileServer für das statische Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))
	http.Handle("/" fileServer))

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
