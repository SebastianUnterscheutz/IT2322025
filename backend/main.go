package main

import (
	"log"
	"main/db"
	"net/http"
)

func main() {
	// Setze das Verzeichnis für die statischen Dateien
	staticDir := "./frontend"
	staticImages := "./frontend/images"
	dbCon = db.Init()

	println(dbCon)
	// Erzeuge einen FileServer für das Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))

	// Routen Sie alle Anfragen an den FileServer
	http.Handle("/", fileServer)
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(staticImages))))

	// Registriere die Route /api/rides mit dem Handler handleRoute
	http.HandleFunc("/api/create/offer", createOffer)
	http.HandleFunc("/api/get/offers", getOffer)
	http.HandleFunc("/api/search/offers", searchOffers)
	http.HandleFunc("/api/edit/offer", editOffer)
	http.HandleFunc("/api/activate/offer", activateOffer)

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
