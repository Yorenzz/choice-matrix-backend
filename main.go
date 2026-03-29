package main

import (
	"log"

	"choice-matrix-backend/internal/api/routes"
	"choice-matrix-backend/internal/cache"
	"choice-matrix-backend/internal/config"
	"choice-matrix-backend/internal/database"
	"choice-matrix-backend/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Configuration
	cfg := config.LoadConfig()
	utils.SetJWTSecret(cfg.JWTSecret)

	if err := database.ConnectDB(cfg); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	if err := cache.ConnectRedis(cfg); err != nil {
		log.Fatalf("Redis initialization failed: %v", err)
	}

	// Set up Gin Server
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Register Routes
	routes.RegisterRoutes(r, database.DB, cache.Client, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// Run Server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
