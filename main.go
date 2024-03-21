// package main

// // use golang gin framework
// import (

// 	"github.com/gin-gonic/gin"
// )

// func init() {
// 	DBconnect()
// }

// var DBconnect = func() {
// 	// connect to database

// }

// func main() {

// 	// use gin framework
// 	r := gin.Default()

// 	v1 := r.Group("/api/v1")
// 	AddUserRouter(v1)

// 	// define route
// 	r.GET("/", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "Hello World",
// 		})
// 	})

// 	// run server
// 	r.Run(":8080")

// }
package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	host     = "localhost"
	port     = 5434
	user     = "postgres"
	password = "t870101"
)

func init() {
	DBconnect()
	DBconstruct()
}

var db *sql.DB

func DBconnect() {
	// connect to the database
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error(err)
		return
	}

	// check the connection
	err = db.Ping()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("Successfully connected to the database!")
}

func DBconstruct() {
	// create the advertisement table
	createStmt := `
		CREATE TABLE IF NOT EXISTS advertisement (
			id SERIAL PRIMARY KEY,
			title TEXT,
			start_at TIMESTAMP,
			end_at TIMESTAMP,
			conditions JSONB
		)
	`
	_, err := db.Exec(createStmt)
	if err != nil {
		log.Error(err)
		return
	}
}

func main() {
	// create an entry for the advertisement in the database
	if db == nil {
		log.Error("Database connection is nil")
		return
	}

	// Insert a new advertisement into the database
	// conditions is a JSONB type, contain the field like age,gender,country and platform

	insertStmt := `
		INSERT INTO advertisement (title, start_at, end_at, conditions)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(insertStmt, "Ad1", "2021-01-01", "2021-12-31", `{"age": 18,
		"gender": "M", "country": "US", "platform": "ios"}`)
	
	if err != nil {
		log.Error(err)
		return
	}
	
	defer db.Close()

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
