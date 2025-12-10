package db

import (
	"log"

	models "github.com/SEDocotor/hellgate-watcher-go/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB opens a GORM connection using the provided Postgres DSN and
// auto-migrates the local models used by this project.
func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(&models.Channel{}, &models.ReportedBattle{}); err != nil {
		// Log migration errors but don't exit — allow the server to start for testing.
		log.Printf("Warning: AutoMigrate failed: %v", err)
	}
}

// CloseDB attempts to close the underlying sql.DB connection if present.
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

// WithDB returns the global *gorm.DB instance; callers should ensure InitDB was called.
func WithDB() *gorm.DB {
	return DB
}
