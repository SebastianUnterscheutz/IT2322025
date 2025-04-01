package main

import (
	"log"
	"main/db"
	"net/http"
)

func main() {
	// Setze das Verzeichnis für die statischen Dateien
	staticDir := "./frontend"
	dbCon := db.Init()

	println(dbCon)
	// Erzeuge einen FileServer für das Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))

	// Routen Sie alle Anfragen an den FileServer
	http.Handle("/", fileServer)

	// Registriere die Route /api/rides mit dem Handler handleRoute
	http.HandleFunc("/api/create/offer", createOffer)
	http.HandleFunc("/api/get/offers", getOffer)

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
