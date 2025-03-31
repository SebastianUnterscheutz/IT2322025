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
	defer db.Close()

	// Teste die Verbindung
	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping the database: %v\n", err)
	}

	// Erstelle eine Tabelle, falls sie nicht existiert
	createTableSQL := `
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
		);


		CREATE TABLE IF NOT EXISTS locations_on_the_way (
    		id INT AUTO_INCREMENT PRIMARY KEY,
    		rides_id INT NOT NULL,
    		plz VARCHAR(10) NOT NULL,
    		city VARCHAR(100) NOT NULL,
    		street VARCHAR(100),
    		house_number VARCHAR(10),
    		FOREIGN KEY (rides_id) REFERENCES rides(id)
		);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Could not create table: %v\n", err)
	}

	log.Println("Database initialized successfully.")

	return db
}
