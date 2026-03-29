package database

import (
	"fmt"
	"log"

	"choice-matrix-backend/internal/config"
	"choice-matrix-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB(cfg config.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection successfully opened")

	if cfg.SkipAutoMigrate {
		log.Println("SKIP_AUTO_MIGRATE is enabled, auto migration is skipped")
		return nil
	}

	// Auto-migrate models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Folder{},
		&models.Project{},
		&models.Row{},
		&models.Column{},
		&models.Cell{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed")
	return nil
}
