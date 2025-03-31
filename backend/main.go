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

	// Handler für die Wurzel-URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		files := []string{}

		// Durchlaufe alle Dateien im staticDir
		err := filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			http.Error(w, "Could not read files", http.StatusInternalServerError)
			log.Println("Error walking the path:", err)
			return
		}

		// Template um die Dateien anzuzeigen
		tmpl := `<html><body><h1>Files in /frontend</h1><ul>{{range .}}<li>{{.}}</li>{{end}}</ul></body></html>`
		t, err := template.New("files").Parse(tmpl)
		if err != nil {
			http.Error(w, "Could not create template", http.StatusInternalServerError)
			log.Println("Template Error:", err)
			return
		}

		// Render das Template mit der Liste der Dateien
		if err := t.Execute(w, files); err != nil {
			http.Error(w, "Could not execute template", http.StatusInternalServerError)
			log.Println("Execution Error:", err)
		}
	})

	// Erzeuge einen FileServer für das statische Verzeichnis
	fileServer := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Starte den Server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
