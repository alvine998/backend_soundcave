package database

import (
	"fmt"
	"os"

	"backend_soundcave/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect membuat koneksi ke database MySQL
func Connect() (*gorm.DB, error) {
	// Ambil konfigurasi dari environment variables
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "soundcave"
	}

	// Format DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	// Koneksi ke database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke database: %w", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&models.Image{},
		&models.User{},
		&models.Album{},
		&models.AppInfo{},
		&models.Artist{},
		&models.Genre{},
		&models.Music{},
		&models.MusicVideo{},
		&models.Notification{},
		&models.Playlist{},
		&models.PlaylistSong{},
		&models.Podcast{},
		&models.SubscriptionPlan{},
		&models.News{},
	)
	if err != nil {
		return nil, fmt.Errorf("gagal migrate database: %w", err)
	}

	DB = db
	return db, nil
}
