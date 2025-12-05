package handlers

import (
	"backend_soundcave/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreatePlaylistRequest struct untuk request create playlist
type CreatePlaylistRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
	CoverImage  *string `json:"cover_image"`
}

// UpdatePlaylistRequest struct untuk request update playlist
type UpdatePlaylistRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
	CoverImage  *string `json:"cover_image"`
}

// CreatePlaylistHandler membuat playlist baru
func CreatePlaylistHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreatePlaylistRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Set default values
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	// Buat playlist baru
	playlist := models.Playlist{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    &isPublic,
		CoverImage:  req.CoverImage,
	}

	if err := db.Create(&playlist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Playlist berhasil dibuat",
		"data":    playlist,
	})
}

// GetPlaylistsHandler mendapatkan semua playlists dengan pagination
func GetPlaylistsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var playlists []models.Playlist

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Playlist{})

	// Filter by is_public jika ada
	if isPublic := c.Query("is_public"); isPublic != "" {
		isPublicBool, err := strconv.ParseBool(isPublic)
		if err == nil {
			query = query.Where("is_public = ?", isPublicBool)
		}
	}

	// Search by name atau description
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get playlists
	if err := query.Offset(offset).Limit(limit).Find(&playlists).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlists",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    playlists,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetPlaylistHandler mendapatkan playlist by ID
func GetPlaylistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlist models.Playlist
	if err := db.First(&playlist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    playlist,
	})
}

// UpdatePlaylistHandler mengupdate playlist
func UpdatePlaylistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlist models.Playlist
	if err := db.First(&playlist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist",
			"error":   err.Error(),
		})
	}

	var req UpdatePlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Name != nil {
		playlist.Name = *req.Name
	}

	if req.Description != nil {
		playlist.Description = req.Description
	}

	if req.IsPublic != nil {
		playlist.IsPublic = req.IsPublic
	}

	if req.CoverImage != nil {
		playlist.CoverImage = req.CoverImage
	}

	if err := db.Save(&playlist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Playlist berhasil diupdate",
		"data":    playlist,
	})
}

// DeletePlaylistHandler menghapus playlist (soft delete)
func DeletePlaylistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlist models.Playlist
	if err := db.First(&playlist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&playlist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Playlist berhasil dihapus",
	})
}

