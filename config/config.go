package config

import (
	"os"
	_ "github.com/joho/godotenv/autoload"
)

// Path: config/config.go

var (
	Host     = os.Getenv("DB_HOST")
	Port     = os.Getenv("DB_PORT")
	User     = os.Getenv("DB_USERNAME")
	DbName   = os.Getenv("DB_NAME")
	Password = os.Getenv("DB_PASSWORD")
	Sslmode  = os.Getenv("DB_SSLMODE")
)
