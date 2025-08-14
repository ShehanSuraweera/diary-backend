package database

import (
	"context"
	"diary-backend/internal/config"
	"diary-backend/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateResource inserts a new resource into the database
func CreateResource(ctx context.Context, resource *models.Resource) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.WithContext(ctx).Create(resource).Error
}

var DB *gorm.DB

func Connect(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
