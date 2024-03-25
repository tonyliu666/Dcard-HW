package service

import (
	"database/sql"
	"fmt"
	"os"
	// load the .env file
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"testing"
)

var (
	host     = os.Getenv("DB_HOST")
	port     = os.Getenv("DB_PORT")
	user     = os.Getenv("DB_USERNAME")

	dbName   = os.Getenv("DB_NAME")
	password = os.Getenv("DB_PASSWORD")
	sslmode  = os.Getenv("DB_SSLMODE")
)

func TestDBconnect(t *testing.T) {
	// connect to the database
	// read the connection parameters

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)

	// open a database connection
	var err error
	db, err := sql.Open("postgres", psqlInfo)
	
	if err != nil {
		t.Errorf("Error: Could not establish a connection with the database")
	}

	insertStmt := `INSERT INTO advertisement (title, start_at, end_at, conditions) VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(insertStmt, "Ad1", "2021-01-01", "2021-12-31", `{"ageStart": 25, "ageEnd": 35, "country": ["TW", "JP","US"], "platform": ["android", "ios"]}`)
	if err != nil {
		t.Errorf("insert failed: %v", err)
	}

}
