package middleware

import (
	"database/sql"
	"dcrad-background/config"
	"fmt"
	_ "github.com/lib/pq"

	//log "github.com/sirupsen/logrus"
)

var connection *sql.DB

func Init(db *sql.DB) {
	// add the database connection to the context
	connection = db
}

func GetDB() (*sql.DB,error) {
	// get the database connection from the context
	if connection == nil {
		// create a new connection
		// read the environment variables
		// and connect to the database
		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.User, config.Password, config.DbName, config.Sslmode)
		// open a database connection
		var err error
		connection, err = sql.Open("postgres", psqlInfo)

		if err != nil {
			return nil, fmt.Errorf("cannot connect to the database: %v", err)
		}

	}
	return connection,nil
}