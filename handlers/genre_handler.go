package handlers

import (
	"backend_soundcave/models"
	"regexp"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateGenreRequest struct untuk request create genre
type CreateGenreRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Color       *string `json:"color"` // Hex color code (e.g., #FF5733)
}

// UpdateGenreRequest struct untuk request update genre
type UpdateGenreRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"` // Hex color code (e.g., #FF5733)
}

// validateHexColor memvalidasi format hex color
func validateHexColor(color string) bool {
	if color == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, color)
	return matched
}

// CreateGenreHandler membuat genre baru
// @Summary      Create new genre
// @Description  Create a new music genre
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        request  body      CreateGenreRequest  true  "Genre Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /genres [post]
func CreateGenreHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateGenreRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi name sudah ada
	var existingGenre models.Genre
	if err := db.Where("name = ?", req.Name).First(&existingGenre).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Genre dengan nama tersebut sudah ada",
		})
	}

	// Validasi color format jika ada
	if req.Color != nil && *req.Color != "" {
		if !validateHexColor(*req.Color) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format color tidak valid. Gunakan format hex color (e.g., #FF5733 atau #F57)",
			})
		}
	}

	// Buat genre baru
	genre := models.Genre{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}

	if err := db.Create(&genre).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat genre",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Genre berhasil dibuat",
		"data":    genre,
	})
}

// GetGenresHandler mendapatkan semua genres dengan pagination
// @Summary      Get all genres
// @Description  Get paginated list of genres with filtering and search
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        search   query     string  false  "Search by name"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /genres [get]
func GetGenresHandler(c *fiber.Ctx, db *gorm.DB) error {
	var genres []models.Genre

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Genre{})

	// Search by name
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// Sort by name atau created_at
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	query = query.Order(sortBy + " " + order)

	// Get total count
	var total int64
	query.Count(&total)

	// Get genres
	if err := query.Offset(offset).Limit(limit).Find(&genres).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data genres",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    genres,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetGenreHandler mendapatkan genre by ID
// @Summary      Get genre by ID
// @Description  Get genre details by ID
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Genre ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /genres/{id} [get]
func GetGenreHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var genre models.Genre
	if err := db.First(&genre, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Genre tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data genre",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    genre,
	})
}

// UpdateGenreHandler mengupdate genre
// @Summary      Update genre
// @Description  Update genre information
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        id       path      int                true  "Genre ID"
// @Param        request  body      UpdateGenreRequest  true  "Update Genre Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /genres/{id} [put]
func UpdateGenreHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var genre models.Genre
	if err := db.First(&genre, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Genre tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data genre",
			"error":   err.Error(),
		})
	}

	var req UpdateGenreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Name != nil {
		// Cek name sudah digunakan oleh genre lain
		var existingGenre models.Genre
		if err := db.Where("name = ? AND id != ?", *req.Name, id).First(&existingGenre).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Genre dengan nama tersebut sudah ada",
			})
		}
		genre.Name = *req.Name
	}

	if req.Description != nil {
		genre.Description = *req.Description
	}

	if req.Color != nil {
		// Validasi color format jika ada
		if *req.Color != "" && !validateHexColor(*req.Color) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format color tidak valid. Gunakan format hex color (e.g., #FF5733 atau #F57)",
			})
		}
		genre.Color = req.Color
	}

	if err := db.Save(&genre).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate genre",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Genre berhasil diupdate",
		"data":    genre,
	})
}

// DeleteGenreHandler menghapus genre (soft delete)
// @Summary      Delete genre
// @Description  Soft delete a genre by ID
// @Tags         Genres
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Genre ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /genres/{id} [delete]
func DeleteGenreHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var genre models.Genre
	if err := db.First(&genre, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Genre tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data genre",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&genre).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus genre",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Genre berhasil dihapus",
	})
}
