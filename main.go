package main

import (
	"database/sql"
	"dcardapp/config"
	"dcardapp/routing"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"dcardapp/middleware"

	"github.com/aviddiviner/gin-limit"
)

// read from .emv file

func init() {
	db, err := DBconnect()
	if err != nil {
		log.Error(err)
	}
	AddDBToMiddleware(db)
}

func DBconnect() (*sql.DB, error) {
	// connect to the database
	// read the connection parameters
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DbName, config.Sslmode)

	// open a database connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the database: %v", err)
	}
	log.Info("Connected to the database")
	return db, nil
}

func AddDBToMiddleware(db *sql.DB) {
	middleware.Init(db)
}

func NewRouter() *gin.Engine {
	// r := gin.New()
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
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
	router.Use(limit.MaxAllowed(3))
	router.Run(":8080")
}
