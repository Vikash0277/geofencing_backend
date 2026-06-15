package database

import (
	"geofencing_backend/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

var DB *gorm.DB

func ConnectDatabase() {

	// Load env
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}

	dsn := os.Getenv("DATABASE_URL")

	// Connect DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	log.Println("Connected to database")

	migration := os.Getenv("MIGRATION") == "true"

	if migration {
		err = DB.AutoMigrate(
			&models.User{},
			&models.Geofence{},
			&models.Vehicle{},
			&models.VehicleLocation{},
			&models.AlertConfig{},
			&models.AlertEvent{},
			&models.Violation{},
			&models.TrackMe{},
		)

		if err != nil {
			log.Fatal("AutoMigrate failed:", err)
		}

		log.Println("Database migrated successfully")
	}
}
