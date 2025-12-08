package handlers

import (
	"backend_soundcave/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreatePodcastRequest struct untuk request create podcast
type CreatePodcastRequest struct {
	Title         string  `json:"title" validate:"required"`
	Host          string  `json:"host" validate:"required"`
	ReleaseDate   string  `json:"release_date" validate:"required"` // Format: "2006-01-02"
	Duration      string  `json:"duration" validate:"required"`     // Format: MM:SS atau HH:MM:SS
	Category      string  `json:"category" validate:"required"`
	Description   string  `json:"description" validate:"required"`
	EpisodeNumber *int    `json:"episode_number"`
	Season        *int    `json:"season"`
	AudioURL      string  `json:"audio_url" validate:"required"`
	Thumbnail     *string `json:"thumbnail"`
}

// UpdatePodcastRequest struct untuk request update podcast
type UpdatePodcastRequest struct {
	Title         *string `json:"title"`
	Host          *string `json:"host"`
	ReleaseDate   *string `json:"release_date"` // Format: "2006-01-02"
	Duration      *string `json:"duration"`     // Format: MM:SS atau HH:MM:SS
	Category      *string `json:"category"`
	Description   *string `json:"description"`
	EpisodeNumber *int    `json:"episode_number"`
	Season        *int    `json:"season"`
	AudioURL      *string `json:"audio_url"`
	Thumbnail     *string `json:"thumbnail"`
}

// CreatePodcastHandler membuat podcast baru
func CreatePodcastHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreatePodcastRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse release date
	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
			"error":   err.Error(),
		})
	}

	// Buat podcast baru
	podcast := models.Podcast{
		Title:         req.Title,
		Host:          req.Host,
		ReleaseDate:   &releaseDate,
		Duration:      req.Duration,
		Category:      req.Category,
		Description:   req.Description,
		EpisodeNumber: req.EpisodeNumber,
		Season:        req.Season,
		AudioURL:      req.AudioURL,
		Thumbnail:     req.Thumbnail,
	}

	if err := db.Create(&podcast).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat podcast",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Podcast berhasil dibuat",
		"data":    podcast,
	})
}

// GetPodcastsHandler mendapatkan semua podcasts dengan pagination
func GetPodcastsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var podcasts []models.Podcast

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Podcast{})

	// Filter by host jika ada
	if host := c.Query("host"); host != "" {
		query = query.Where("host LIKE ?", "%"+host+"%")
	}

	// Filter by category jika ada
	if category := c.Query("category"); category != "" {
		query = query.Where("category LIKE ?", "%"+category+"%")
	}

	// Filter by season jika ada
	if season := c.QueryInt("season", 0); season > 0 {
		query = query.Where("season = ?", season)
	}

	// Search by title, host, atau description
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR host LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
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

	// Get podcasts
	if err := query.Offset(offset).Limit(limit).Find(&podcasts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data podcasts",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    podcasts,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetPodcastHandler mendapatkan podcast by ID
func GetPodcastHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var podcast models.Podcast
	if err := db.First(&podcast, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Podcast tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data podcast",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    podcast,
	})
}

// UpdatePodcastHandler mengupdate podcast
func UpdatePodcastHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var podcast models.Podcast
	if err := db.First(&podcast, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Podcast tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data podcast",
			"error":   err.Error(),
		})
	}

	var req UpdatePodcastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		podcast.Title = *req.Title
	}

	if req.Host != nil {
		podcast.Host = *req.Host
	}

	if req.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *req.ReleaseDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		podcast.ReleaseDate = &releaseDate
	}

	if req.Duration != nil {
		podcast.Duration = *req.Duration
	}

	if req.Category != nil {
		podcast.Category = *req.Category
	}

	if req.Description != nil {
		podcast.Description = *req.Description
	}

	if req.EpisodeNumber != nil {
		podcast.EpisodeNumber = req.EpisodeNumber
	}

	if req.Season != nil {
		podcast.Season = req.Season
	}

	if req.AudioURL != nil {
		podcast.AudioURL = *req.AudioURL
	}

	if req.Thumbnail != nil {
		podcast.Thumbnail = req.Thumbnail
	}

	if err := db.Save(&podcast).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate podcast",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Podcast berhasil diupdate",
		"data":    podcast,
	})
}

// DeletePodcastHandler menghapus podcast (soft delete)
func DeletePodcastHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var podcast models.Podcast
	if err := db.First(&podcast, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Podcast tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data podcast",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&podcast).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus podcast",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Podcast berhasil dihapus",
	})
}
