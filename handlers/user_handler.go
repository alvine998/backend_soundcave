package handlers

import (
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUserRequest struct untuk request create user
type CreateUserRequest struct {
	FullName     string  `json:"full_name" validate:"required"`
	Email        string  `json:"email" validate:"required,email"`
	Password     string  `json:"password" validate:"required,min=6"`
	Phone        *string `json:"phone"`
	Location     *string `json:"location"`
	Bio          *string `json:"bio"`
	ProfileImage *string `json:"profile_image"`
	Role         string  `json:"role"`
}

// UpdateUserRequest struct untuk request update user
type UpdateUserRequest struct {
	FullName     *string `json:"full_name"`
	Email        *string `json:"email" validate:"omitempty,email"`
	Password     *string `json:"password" validate:"omitempty,min=6"`
	Phone        *string `json:"phone"`
	Location     *string `json:"location"`
	Bio          *string `json:"bio"`
	ProfileImage *string `json:"profile_image"`
	Role         *string `json:"role"`
}

// CreateUserHandler membuat user baru
// @Summary      Create new user
// @Description  Create a new user account (admin only)
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "User Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /users [post]
func CreateUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
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

	// Set role default
	role := models.RoleUser
	if req.Role != "" {
		switch req.Role {
		case "admin":
			role = models.RoleAdmin
		case "premium":
			role = models.RolePremium
		case "user":
			role = models.RoleUser
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Role tidak valid. Pilih: user, admin, atau premium",
			})
		}
	}

	// Buat user baru
	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		FullName:     req.FullName,
		Email:        req.Email,
		Password:     &hashedPasswordStr,
		Phone:        req.Phone,
		Location:     req.Location,
		Bio:          req.Bio,
		ProfileImage: req.ProfileImage,
		Role:         role,
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User berhasil dibuat",
		"data":    user,
	})
}

// GetUsersHandler mendapatkan semua users dengan pagination
// @Summary      Get all users
// @Description  Get paginated list of users with filtering and search
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number" default(1)
// @Param        limit   query     int     false  "Items per page" default(10)
// @Param        role    query     string  false  "Filter by role"
// @Param        search  query     string  false  "Search by name or email"
// @Success      200     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /users [get]
func GetUsersHandler(c *fiber.Ctx, db *gorm.DB) error {
	var users []models.User

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.User{})

	// Filter by role jika ada
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// Search by name atau email
	if search := c.Query("search"); search != "" {
		query = query.Where("full_name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get users
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data users",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetUserHandler mendapatkan user by ID
// @Summary      Get user by ID
// @Description  Get user details by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{id} [get]
func GetUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
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

// UpdateUserHandler mengupdate user
// @Summary      Update user
// @Description  Update user information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "User ID"
// @Param        request  body      UpdateUserRequest  true  "Update User Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /users/{id} [put]
func UpdateUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
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

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.FullName != nil {
		user.FullName = *req.FullName
	}

	if req.Email != nil {
		// Cek email sudah digunakan oleh user lain
		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", *req.Email, id).First(&existingUser).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Email sudah digunakan",
			})
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal hash password",
				"error":   err.Error(),
			})
		}
		hashedPasswordStr := string(hashedPassword)
		user.Password = &hashedPasswordStr
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Location != nil {
		user.Location = req.Location
	}

	if req.Bio != nil {
		user.Bio = req.Bio
	}

	if req.ProfileImage != nil {
		user.ProfileImage = req.ProfileImage
	}

	if req.Role != nil {
		switch *req.Role {
		case "admin":
			user.Role = models.RoleAdmin
		case "premium":
			user.Role = models.RolePremium
		case "user":
			user.Role = models.RoleUser
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Role tidak valid. Pilih: user, admin, atau premium",
			})
		}
	}

	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User berhasil diupdate",
		"data":    user,
	})
}

// DeleteUserHandler menghapus user (soft delete)
// @Summary      Delete user
// @Description  Soft delete a user by ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{id} [delete]
func DeleteUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
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

	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User berhasil dihapus",
	})
}
