package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	_ "github.com/joho/godotenv/autoload"
)

// read from .emv file
var (
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	user     = os.Getenv("DB_USERNAME")
	dbName   = os.Getenv("DB_NAME")
	password = os.Getenv("DB_PASSWORD")
	sslmode  = os.Getenv("DB_SSLMODE")

)

func init() {
	DBconnect()
}

var db *sql.DB

func DBconnect() {
	// connect to the database
	// read the connection parameters
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)

	log.Info(psqlInfo)
	// open a database connection
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error(err)
		return
	}
	err = db.Ping()
	if err != nil {
		log.Error("Error: Could not establish a connection with the database")
		return
	}
	log.Info("Connected to the database")

}

func main() {
	// create an entry for the advertisement in the database
	if db == nil {
		log.Error("Database connection is nil")
		return
	}

	// Insert a new advertisement into the database
	// conditions is a JSONB type, contain the field like age,gender,country and platform

	// Query the database for the advertisement
	rows, err := db.Query("SELECT id, title, start_at, end_at, conditions FROM advertisement")
	if err != nil {
		log.Error(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var startAt string
		var endAt string
		var conditions string
		err = rows.Scan(&id, &title, &startAt, &endAt, &conditions)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info(title, startAt, endAt, conditions)
	}
}
