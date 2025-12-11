package handlers

import (
	"backend_soundcave/models"
	"backend_soundcave/utils"
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

// RegisterRequest struct untuk request register
type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone"`
}

// LoginRequest struct untuk request login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// GoogleAuthRequest struct untuk request Google Auth
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

// RegisterHandler menangani registrasi user baru
// @Summary      Register new user
// @Description  Register a new user account with role "user"
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "Register Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /auth/register [post]
func RegisterHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi field required
	if req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Field full_name wajib diisi",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Field email wajib diisi",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Field password wajib diisi",
		})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Password minimal 6 karakter",
		})
	}

	// Validasi email sudah ada
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Email sudah terdaftar",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal hash password",
			"error":   err.Error(),
		})
	}

	// Buat user baru
	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: &hashedPasswordStr,
		Phone:    phone,
		Role:     models.RoleUser, // Default role adalah "user"
	}

	if err := db.Create(&user).Error; err != nil {
		// Log error untuk debugging
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat user",
			"error":   err.Error(),
		})
	}

	// Return user data tanpa token (user perlu login terpisah untuk mendapatkan token)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Registrasi berhasil. Silakan login untuk mendapatkan token.",
		"data":    user,
	})
}

// LoginHandler menangani login user
// @Summary      User login
// @Description  Login with email and password to get JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "Login Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /auth/login [post]
func LoginHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Cari user by email
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Email atau password salah",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	// Verifikasi password
	if user.Password == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User ini menggunakan Google Auth. Silakan login dengan Google",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Email atau password salah",
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal generate token",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data": fiber.Map{
			"user":  user,
			"token": token,
		},
	})
}

// GoogleAuthHandler menangani login dengan Firebase Google Auth
// @Summary      Google authentication
// @Description  Login or register using Google ID token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      GoogleAuthRequest  true  "Google Auth Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /auth/google [post]
func GoogleAuthHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req GoogleAuthRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi ID token dengan Firebase
	ctx := context.Background()
	audience := os.Getenv("FIREBASE_AUDIENCE")
	if audience == "" {
		// Default audience untuk Firebase
		audience = "soundcave-app" // Ganti dengan client ID Firebase Anda
	}

	payload, err := idtoken.Validate(ctx, req.IDToken, audience)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token Google tidak valid",
			"error":   err.Error(),
		})
	}

	// Extract user info dari token
	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Email tidak ditemukan di token",
		})
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	// Cek apakah user sudah ada
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// User belum ada, buat user baru
		var profileImage *string
		if picture != "" {
			profileImage = &picture
		}
		user = models.User{
			FullName:     name,
			Email:        email,
			Password:     nil, // Tidak ada password untuk Google Auth
			ProfileImage: profileImage,
			Role:         models.RoleUser,
		}

		if err := db.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membuat user",
				"error":   err.Error(),
			})
		}
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	} else {
		// User sudah ada, update profile image jika ada
		if picture != "" {
			user.ProfileImage = &picture
			if name != "" {
				user.FullName = name
			}
			db.Save(&user)
		}
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal generate token",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login Google berhasil",
		"data": fiber.Map{
			"user":  user,
			"token": token,
		},
	})
}

// GetProfileHandler mendapatkan profile user yang sedang login
// @Summary      Get user profile
// @Description  Get current authenticated user profile
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /profile [get]
func GetProfileHandler(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "User ID tidak valid",
		})
	}

	var user models.User
	if err := db.Select("*").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}
