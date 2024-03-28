// test database connection
package main

import (
	"database/sql"
	"dcardapp/config"
	"fmt"
	"testing"
)

func TestDBconnect(t *testing.T) {
	// connect to the database
	// read the connection parameters
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DbName, config.Sslmode)

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
