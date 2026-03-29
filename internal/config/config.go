package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	SkipAutoMigrate bool
	DBHost          string
	DBUser          string
	DBPass          string
	DBName          string
	DBPort          string
	JWTSecret       string
	RedisHost       string
	RedisPort       string
	RedisPassword   string
	RedisDB         int
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	return Config{
		Port:            getEnv("PORT", "8080"),
		SkipAutoMigrate: getEnvAsBool("SKIP_AUTO_MIGRATE", false),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPass:          getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "choice_matrix"),
		DBPort:          getEnv("DB_PORT", "5432"),
		JWTSecret:       getEnv("JWT_SECRET", "my-super-secret-choice-matrix-key"),
		RedisHost:       getEnv("REDIS_HOST", "localhost"),
		RedisPort:       getEnv("REDIS_PORT", "6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		RedisDB:         getEnvAsInt("REDIS_DB", 0),
		AccessTokenTTL:  getEnvAsDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getEnvAsDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}

	var result int
	_, err := fmt.Sscanf(value, "%d", &result)
	if err != nil {
		return fallback
	}

	return result
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
