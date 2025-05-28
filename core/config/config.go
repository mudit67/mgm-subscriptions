package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	MongoURI     string
	DatabaseName string
	JWTSecret    string
	JWTExpiry    string
}

func LoadConfig() *Config {
	// Load .env file - ignores error if file doesn't exist
	_ = godotenv.Load()

	return &Config{
		Port:         getEnv("PORT"),
		MongoURI:     getEnv("MONGO_URI"),
		DatabaseName: getEnv("DATABASE_NAME"),
		JWTSecret:    getEnv("JWT_SECRET"),
		JWTExpiry:    getEnv("JWT_EXPIRY"),
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else {
		panic("Error: Loading Env File")
	}
}
