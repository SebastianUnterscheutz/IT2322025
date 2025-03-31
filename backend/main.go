package main

import (
	"log"
	"net/http"
	"time"
)

// LoggerMiddleware loggt jede Anfrage mit ihrem Methodentyp, URL und Antwortzeit
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Logge die Anfrage
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)

		// Rufe den nächsten Handler auf
		next.ServeHTTP(w, r)

		// Logge die Antwortzeit
		log.Printf("Completed %s request for %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	// Einfache HTTP-Handler-Funktion
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	// Verwende die LoggerMiddleware
	http.Handle("/", LoggerMiddleware(handler))

	// Starte den Server
	log.Println("Starting server on :3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
