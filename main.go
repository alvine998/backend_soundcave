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

	// Log environment check (tanpa expose sensitive data)
	log.Println("Checking environment variables...")
	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Printf("WARNING: %s tidak di-set", envVar)
		} else {
			log.Printf("✓ %s di-set", envVar)
		}
	}

	// Check Firebase service account file
	firebasePath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if firebasePath == "" {
		firebasePath = "/root/firebase-service-account.json"
	}
	if _, err := os.Stat(firebasePath); os.IsNotExist(err) {
		log.Printf("WARNING: Firebase service account file tidak ditemukan di %s", firebasePath)
	} else {
		log.Printf("✓ Firebase service account file ditemukan di %s", firebasePath)
	}

	// Initialize database
	log.Println("Mencoba koneksi ke database...")
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Gagal koneksi ke database: %v", err)
	}
	log.Println("✓ Koneksi database berhasil")

	// Initialize Firebase
	log.Println("Mencoba inisialisasi Firebase...")
	firebaseApp, err := config.InitFirebase()
	if err != nil {
		log.Fatalf("Gagal inisialisasi Firebase: %v", err)
	}
	log.Println("✓ Firebase berhasil diinisialisasi")

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
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
