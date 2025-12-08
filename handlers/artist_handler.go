package handlers

import (
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateArtistRequest struct untuk request create artist
type CreateArtistRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Bio          string                 `json:"bio" validate:"required"`
	Genre        string                 `json:"genre"`
	Country      string                 `json:"country"`
	DebutYear    string                 `json:"debut_year" validate:"required,len=4"`
	Website      *string                `json:"website"`
	Email        string                 `json:"email" validate:"required,email"`
	Phone        *string                `json:"phone"`
	SocialMedia  map[string]interface{} `json:"social_media"`
	ProfileImage *string                `json:"profile_image"`
}

// UpdateArtistRequest struct untuk request update artist
type UpdateArtistRequest struct {
	Name         *string                `json:"name"`
	Bio          *string                `json:"bio"`
	Genre        *string                `json:"genre"`
	Country      *string                `json:"country"`
	DebutYear    *string                `json:"debut_year" validate:"omitempty,len=4"`
	Website      *string                `json:"website"`
	Email        *string                `json:"email" validate:"omitempty,email"`
	Phone        *string                `json:"phone"`
	SocialMedia  map[string]interface{} `json:"social_media"`
	ProfileImage *string                `json:"profile_image"`
}

// CreateArtistHandler membuat artist baru
func CreateArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateArtistRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi email sudah ada
	var existingArtist models.Artist
	if err := db.Where("email = ?", req.Email).First(&existingArtist).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Email sudah terdaftar",
		})
	}

	// Convert social media map to JSONB
	var socialMedia models.JSONB
	if req.SocialMedia != nil {
		socialMedia = models.JSONB(req.SocialMedia)
	}

	// Buat artist baru
	artist := models.Artist{
		Name:         req.Name,
		Bio:          req.Bio,
		Genre:        req.Genre,
		Country:      req.Country,
		DebutYear:    req.DebutYear,
		Website:      req.Website,
		Email:        req.Email,
		Phone:        req.Phone,
		SocialMedia:  socialMedia,
		ProfileImage: req.ProfileImage,
	}

	if err := db.Create(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil dibuat",
		"data":    artist,
	})
}

// GetArtistsHandler mendapatkan semua artists dengan pagination
func GetArtistsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var artists []models.Artist

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Artist{})

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Filter by country jika ada
	if country := c.Query("country"); country != "" {
		query = query.Where("country = ?", country)
	}

	// Filter by debut year jika ada
	if debutYear := c.Query("debut_year"); debutYear != "" {
		query = query.Where("debut_year = ?", debutYear)
	}

	// Search by name atau email
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Sort by created_at
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	query = query.Order(sortBy + " " + order)

	// Get total count
	var total int64
	query.Count(&total)

	// Get artists
	if err := query.Offset(offset).Limit(limit).Find(&artists).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artists",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    artists,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetArtistHandler mendapatkan artist by ID
func GetArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    artist,
	})
}

// UpdateArtistHandler mengupdate artist
func UpdateArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	var req UpdateArtistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Name != nil {
		artist.Name = *req.Name
	}

	if req.Bio != nil {
		artist.Bio = *req.Bio
	}

	if req.Genre != nil {
		artist.Genre = *req.Genre
	}

	if req.Country != nil {
		artist.Country = *req.Country
	}

	if req.DebutYear != nil {
		artist.DebutYear = *req.DebutYear
	}

	if req.Website != nil {
		artist.Website = req.Website
	}

	if req.Email != nil {
		// Cek email sudah digunakan oleh artist lain
		var existingArtist models.Artist
		if err := db.Where("email = ? AND id != ?", *req.Email, id).First(&existingArtist).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Email sudah digunakan",
			})
		}
		artist.Email = *req.Email
	}

	if req.Phone != nil {
		artist.Phone = req.Phone
	}

	if req.SocialMedia != nil {
		artist.SocialMedia = models.JSONB(req.SocialMedia)
	}

	if req.ProfileImage != nil {
		artist.ProfileImage = req.ProfileImage
	}

	if err := db.Save(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil diupdate",
		"data":    artist,
	})
}

// DeleteArtistHandler menghapus artist (soft delete)
func DeleteArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil dihapus",
	})
}
