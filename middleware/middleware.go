package middleware

import (
	"database/sql"

)

var connection *sql.DB

func Init(db *sql.DB) {
	// add the database connection to the context
	connection = db
}

func GetDB() (*sql.DB) {
	// get the database connection from the context
	return connection
}