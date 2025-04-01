package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func Init() *sql.DB {

	// Hole die Umgebungsvariablen
	user := os.Getenv("MARIADB_USER")
	password := os.Getenv("MARIADB_PASSWORD")
	database := os.Getenv("MARIADB_DATABASE")
	hostname := os.Getenv("MARIADB_HOSTNAME")

	// Erstelle den DSN (Data Source Name) für die MySQL-Verbindung
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, hostname, database)

	// Verbinde dich mit der MariaDB-Datenbank
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v\n", err)
	}

	// Teste die Verbindung
	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping the database: %v\n", err)
	}

	// Erstelle eine Tabelle, falls sie nicht existiert
	// Erstelle die Tabelle `rides`
	createRidesTableSQL := `
	CREATE TABLE IF NOT EXISTS rides (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		first_name VARCHAR(100),
		email VARCHAR(100) NOT NULL,
		class VARCHAR(10),
		phone_number VARCHAR(20),
		valid_from DATE,
		valid_until DATE,
		additional_information TEXT,
		other TEXT,
		token VARCHAR(100),
		activated BOOLEAN
	);`

	_, err = db.Exec(createRidesTableSQL)
	if err != nil {
		log.Fatalf("Could not create rides table: %v\n", err)
	}

	// Remove
	remove := "DROP TABLE IF EXISTS locations_on_the_way;"
	_, err = db.Exec(remove)

	// Erstelle die Tabelle `locations_on_the_way`
	createLocationsTableSQL := `
	CREATE TABLE IF NOT EXISTS locations_on_the_way (
		id INT AUTO_INCREMENT PRIMARY KEY,
		rides_id INT NOT NULL,
		plz VARCHAR(10) NOT NULL,
		city VARCHAR(100) NOT NULL,
		street VARCHAR(100),
		house_number VARCHAR(10),
	    latitude FLOAT,
	    longitude FLOAt,
		FOREIGN KEY (rides_id) REFERENCES rides(id)
	);`

	_, err = db.Exec(createLocationsTableSQL)
	if err != nil {
		log.Fatalf("Could not create locations_on_the_way table: %v\n", err)
	}

	log.Println("Database initialized successfully.")

	return db
}
