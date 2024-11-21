package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser       string
	DBPassword   string
	DBName       string
	DBHost       string
	DBPort       string
	ORDERCARTGRPCPORT string
	JWTSecretKey string
}

func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("No .env file found")
	}

	return Config{
		DBUser:       os.Getenv("DBUSER"),
		DBPassword:   os.Getenv("DBPASSWORD"),
		DBName:       os.Getenv("DBNAME"),
		DBHost:       os.Getenv("DBHOST"),
		DBPort:       os.Getenv("DBPORT"),
		ORDERCARTGRPCPORT: os.Getenv("ORDERCARTGRPCPORT"),
		JWTSecretKey: os.Getenv("JWTSECRET"),
	}
}