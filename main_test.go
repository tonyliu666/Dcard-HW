// test database connection
package main 
import (
	"database/sql"
	"fmt"
	"testing"
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
		return
	}

	err = db.Ping()
	if err != nil {
		t.Errorf("Error: Could not establish a connection with the database")
		return
	}
}