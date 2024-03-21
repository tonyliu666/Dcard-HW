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
	DBconstruct()
}

var db *sql.DB

func DBconstruct() {
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

func main() {
	// create an entry for the advertisement in the database
	if db == nil {
		log.Error("Database connection is nil")
		return
	}

	// Insert a new advertisement into the database

	insertStmt := `
		INSERT INTO advertisement (title, start_at, end_at, conditions)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.Exec(insertStmt, "New Advertisement", "2021-01-01T00:00:00Z", "2021-12-31T23:59:59Z", `{"age": 18, "location": "USA"}`)
	if err != nil {
		log.Error(err)
	}
	
	defer db.Close()

	// Query the database for the advertisement
	rows, err := db.Query("SELECT id, title, start_at, end_at, conditions FROM advertisement")
	if err != nil {
		log.Error(err)
		return
	}
	defer rows.Close()

	log.Info("id | title | start_at | end_at | conditions")
}
