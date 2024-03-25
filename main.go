package main

import (
	"database/sql"
	"fmt"
	"os"

	"dcardapp/routing"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"dcardapp/middleware"
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
	db,err := DBconnect()
	if err != nil {
		log.Error(err)
	}
	AddDBToMiddleware(db)
}


func DBconnect() (*sql.DB, error){
	// connect to the database
	// read the connection parameters
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)

	log.Info(psqlInfo)
	// open a database connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error: Could not establish a connection with the database")
	}
	log.Info("Connected to the database")
	return db, nil
}

func AddDBToMiddleware(db *sql.DB) {
	middleware.Init(db)
}

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	api := r.Group("/api/v1")

	routing.AddUserRouter(api)
	return r
}

func main() {
	// create an entry for the advertisement in the database
	// create router
	router := NewRouter()
	router.Run(":8080")

}
