package main

import (
	"log"
	"os"

	"backend_soundcave/config"
	"backend_soundcave/database"
	"backend_soundcave/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variables dari sistem")
	}

	// Initialize database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Gagal koneksi ke database:", err)
	}

	// Initialize Firebase
	firebaseApp, err := config.InitFirebase()
	if err != nil {
		log.Fatal("Gagal inisialisasi Firebase:", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "SoundCave Backend",
	})

	// Middleware
	app.Use(logger.New())

	// CORS configuration
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "*"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: corsOrigins != "*", // Hanya true jika origin bukan wildcard
	}))

	// Routes
	routes.SetupRoutes(app, db, firebaseApp)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "6002"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0" // Default untuk Docker
	}

	log.Printf("Server berjalan di %s:%s", host, port)
	if err := app.Listen(host + ":" + port); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
