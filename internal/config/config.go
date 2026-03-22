package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	DBHost    string
	DBUser    string
	DBPass    string
	DBName    string
	DBPort    string
	JWTSecret string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	return Config{
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASS", "postgres"),
		DBName:    getEnv("DB_NAME", "choice_matrix"),
		DBPort:    getEnv("DB_PORT", "5432"),
		JWTSecret: getEnv("JWT_SECRET", "my-super-secret-choice-matrix-key"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
