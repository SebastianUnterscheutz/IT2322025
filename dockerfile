# Basierend auf dem offiziellem Golang Docker-Image
FROM golang:1.23

# Erstellen Sie ein Arbeitsverzeichnis
WORKDIR /go/src/app

# Kopieren Sie alle .go-Dateien
COPY /backend/. .

COPY /frontend/. /frontend/.

# Kompilieren Sie das Go-Programm
RUN go build -o main .

# Starten Sie die Anwendung beim Start des Containers
CMD ["./main"]