package db

import (
	"log"

	models "github.com/SEDocotor/hellgate-watcher-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(&models.Channel{}, &models.ReportedBattle{}); err != nil {
		log.Printf("Warning: AutoMigrate failed: %v", err)
	}
}

func CloseDB() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func WithDB() *gorm.DB {
	return DB
}
