package main

import (
	"log"
	"os"

	"backend_soundcave/config"
	"backend_soundcave/database"
	_ "backend_soundcave/docs" // Swagger docs
	"backend_soundcave/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

// @title           SoundCave Backend API
// @version         1.0
// @description     API documentation for SoundCave Backend - Music streaming platform
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@soundcave.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      api.rezim.site
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

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
		AppName:   "SoundCave Backend",
		BodyLimit: 300 * 1024 * 1024, // 300MB untuk support upload audio (50MB), music video (150MB), dan podcast video (200MB) dengan safety margin
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	// Rate limiting: 30 requests per minute (DISABLED)
	// app.Use(limiter.New(limiter.Config{
	// 	Max:        30,
	// 	Expiration: 1 * time.Minute,
	// 	KeyGenerator: func(c *fiber.Ctx) string {
	// 		// Use IP address as key for rate limiting
	// 		return c.IP()
	// 	},
	// 	LimitReached: func(c *fiber.Ctx) error {
	// 		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
	// 			"success": false,
	// 			"message": "Terlalu banyak request. Maksimal 30 requests per menit.",
	// 		})
	// 	},
	// 	SkipFailedRequests:     false,
	// 	SkipSuccessfulRequests: false,
	// }))

	// CORS configuration - Allow all origins
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: false, // Must be false when AllowOrigins is "*"
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours
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
