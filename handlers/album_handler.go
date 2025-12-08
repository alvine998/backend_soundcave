package handlers

import (
	"backend_soundcave/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateAlbumRequest struct untuk request create album
type CreateAlbumRequest struct {
	Title       string  `json:"title" validate:"required"`
	ArtistID    int     `json:"artist_id" validate:"required"`
	Artist      string  `json:"artist" validate:"required"`
	ReleaseDate string  `json:"release_date" validate:"required"` // Format: "2006-01-02"
	AlbumType   string  `json:"album_type" validate:"required,oneof=single EP album compilation"`
	Genre       string  `json:"genre"`
	TotalTracks int     `json:"total_tracks" validate:"min=0"`
	RecordLabel *string `json:"record_label"`
}

// UpdateAlbumRequest struct untuk request update album
type UpdateAlbumRequest struct {
	Title       *string `json:"title"`
	ArtistID    *int    `json:"artist_id"`
	Artist      *string `json:"artist"`
	ReleaseDate *string `json:"release_date"` // Format: "2006-01-02"
	AlbumType   *string `json:"album_type" validate:"omitempty,oneof=single EP album compilation"`
	Genre       *string `json:"genre"`
	TotalTracks *int    `json:"total_tracks" validate:"omitempty,min=0"`
	RecordLabel *string `json:"record_label"`
}

// CreateAlbumHandler membuat album baru
func CreateAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateAlbumRequest

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

	// Validasi album type
	var albumType models.AlbumType
	switch req.AlbumType {
	case "single":
		albumType = models.AlbumTypeSingle
	case "EP":
		albumType = models.AlbumTypeEP
	case "album":
		albumType = models.AlbumTypeAlbum
	case "compilation":
		albumType = models.AlbumTypeCompilation
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Album type tidak valid. Pilih: single, EP, album, atau compilation",
		})
	}

	// Buat album baru
	album := models.Album{
		Title:       req.Title,
		ArtistID:    req.ArtistID,
		Artist:      req.Artist,
		ReleaseDate: &releaseDate,
		AlbumType:   albumType,
		Genre:       req.Genre,
		TotalTracks: req.TotalTracks,
		RecordLabel: req.RecordLabel,
	}

	if err := db.Create(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil dibuat",
		"data":    album,
	})
}

// GetAlbumsHandler mendapatkan semua albums dengan pagination
func GetAlbumsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var albums []models.Album

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Album{})

	// Filter by artist_id jika ada
	if artistID := c.QueryInt("artist_id", 0); artistID > 0 {
		query = query.Where("artist_id = ?", artistID)
	}

	// Filter by album_type jika ada
	if albumType := c.Query("album_type"); albumType != "" {
		query = query.Where("album_type = ?", albumType)
	}

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Search by title atau artist
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR artist LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get albums
	if err := query.Offset(offset).Limit(limit).Find(&albums).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data albums",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    albums,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetAlbumHandler mendapatkan album by ID
func GetAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    album,
	})
}

// UpdateAlbumHandler mengupdate album
func UpdateAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	var req UpdateAlbumRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		album.Title = *req.Title
	}

	if req.ArtistID != nil {
		album.ArtistID = *req.ArtistID
	}

	if req.Artist != nil {
		album.Artist = *req.Artist
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
		album.ReleaseDate = &releaseDate
	}

	if req.AlbumType != nil {
		switch *req.AlbumType {
		case "single":
			album.AlbumType = models.AlbumTypeSingle
		case "EP":
			album.AlbumType = models.AlbumTypeEP
		case "album":
			album.AlbumType = models.AlbumTypeAlbum
		case "compilation":
			album.AlbumType = models.AlbumTypeCompilation
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Album type tidak valid. Pilih: single, EP, album, atau compilation",
			})
		}
	}

	if req.Genre != nil {
		album.Genre = *req.Genre
	}

	if req.TotalTracks != nil {
		album.TotalTracks = *req.TotalTracks
	}

	if req.RecordLabel != nil {
		album.RecordLabel = req.RecordLabel
	}

	if err := db.Save(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil diupdate",
		"data":    album,
	})
}

// DeleteAlbumHandler menghapus album (soft delete)
func DeleteAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil dihapus",
	})
}
