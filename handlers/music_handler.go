package handlers

import (
	"backend_soundcave/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateMusicRequest struct untuk request create music
type CreateMusicRequest struct {
	Title         string  `json:"title" validate:"required"`
	Artist        string  `json:"artist" validate:"required"`
	ArtistID      int     `json:"artist_id" validate:"required"`
	Album         *string `json:"album"`
	AlbumID       *int    `json:"album_id"`
	Genre         string  `json:"genre" validate:"required"`
	ReleaseDate   string  `json:"release_date" validate:"required"` // Format: "2006-01-02"
	Duration      string  `json:"duration" validate:"required"`     // Format: MM:SS atau HH:MM:SS
	Language      string  `json:"language" validate:"required"`
	Explicit      *bool   `json:"explicit"`
	Lyrics        *string `json:"lyrics"`
	Description   *string `json:"description"`
	Tags          *string `json:"tags"`
	AudioFileURL  string  `json:"audio_file_url" validate:"required"`
	CoverImageURL *string `json:"cover_image_url"`
	PlayCount     *int    `json:"play_count"`
	LikeCount     *int    `json:"like_count"`
}

// UpdateMusicRequest struct untuk request update music
type UpdateMusicRequest struct {
	Title         *string `json:"title"`
	Artist        *string `json:"artist"`
	ArtistID      *int    `json:"artist_id"`
	Album         *string `json:"album"`
	AlbumID       *int    `json:"album_id"`
	Genre         *string `json:"genre"`
	ReleaseDate   *string `json:"release_date"` // Format: "2006-01-02"
	Duration      *string `json:"duration"`     // Format: MM:SS atau HH:MM:SS
	Language      *string `json:"language"`
	Explicit      *bool   `json:"explicit"`
	Lyrics        *string `json:"lyrics"`
	Description   *string `json:"description"`
	Tags          *string `json:"tags"`
	AudioFileURL  *string `json:"audio_file_url"`
	CoverImageURL *string `json:"cover_image_url"`
	PlayCount     *int    `json:"play_count"`
	LikeCount     *int    `json:"like_count"`
}

// CreateMusicHandler membuat music baru
func CreateMusicHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateMusicRequest

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

	// Set default values
	playCount := 0
	likeCount := 0
	explicit := false

	if req.PlayCount != nil {
		playCount = *req.PlayCount
	}
	if req.LikeCount != nil {
		likeCount = *req.LikeCount
	}
	if req.Explicit != nil {
		explicit = *req.Explicit
	}

	// Buat music baru
	music := models.Music{
		Title:         req.Title,
		Artist:        req.Artist,
		ArtistID:      req.ArtistID,
		Album:         req.Album,
		AlbumID:       req.AlbumID,
		Genre:         req.Genre,
		ReleaseDate:   &releaseDate,
		Duration:      req.Duration,
		Language:      req.Language,
		Explicit:      &explicit,
		Lyrics:        req.Lyrics,
		Description:   req.Description,
		Tags:          req.Tags,
		AudioFileURL:  req.AudioFileURL,
		CoverImageURL: req.CoverImageURL,
		PlayCount:     &playCount,
		LikeCount:     &likeCount,
	}

	if err := db.Create(&music).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat music",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Music berhasil dibuat",
		"data":    music,
	})
}

// GetMusicsHandler mendapatkan semua musics dengan pagination
func GetMusicsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var musics []models.Music

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Music{})

	// Filter by artist_id jika ada
	if artistID := c.QueryInt("artist_id", 0); artistID > 0 {
		query = query.Where("artist_id = ?", artistID)
	}

	// Filter by album_id jika ada
	if albumID := c.QueryInt("album_id", 0); albumID > 0 {
		query = query.Where("album_id = ?", albumID)
	}

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Filter by language jika ada
	if language := c.Query("language"); language != "" {
		query = query.Where("language = ?", language)
	}

	// Filter by explicit jika ada
	if explicit := c.Query("explicit"); explicit != "" {
		explicitBool, err := strconv.ParseBool(explicit)
		if err == nil {
			query = query.Where("explicit = ?", explicitBool)
		}
	}

	// Search by title, artist, atau album
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR artist LIKE ? OR album LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
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

	// Get musics
	if err := query.Offset(offset).Limit(limit).Find(&musics).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data musics",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    musics,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetMusicHandler mendapatkan music by ID
func GetMusicHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var music models.Music
	if err := db.First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    music,
	})
}

// UpdateMusicHandler mengupdate music
func UpdateMusicHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var music models.Music
	if err := db.First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	var req UpdateMusicRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		music.Title = *req.Title
	}

	if req.Artist != nil {
		music.Artist = *req.Artist
	}

	if req.ArtistID != nil {
		music.ArtistID = *req.ArtistID
	}

	if req.Album != nil {
		music.Album = req.Album
	}

	if req.AlbumID != nil {
		music.AlbumID = req.AlbumID
	}

	if req.Genre != nil {
		music.Genre = *req.Genre
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
		music.ReleaseDate = &releaseDate
	}

	if req.Duration != nil {
		music.Duration = *req.Duration
	}

	if req.Language != nil {
		music.Language = *req.Language
	}

	if req.Explicit != nil {
		music.Explicit = req.Explicit
	}

	if req.Lyrics != nil {
		music.Lyrics = req.Lyrics
	}

	if req.Description != nil {
		music.Description = req.Description
	}

	if req.Tags != nil {
		music.Tags = req.Tags
	}

	if req.AudioFileURL != nil {
		music.AudioFileURL = *req.AudioFileURL
	}

	if req.CoverImageURL != nil {
		music.CoverImageURL = req.CoverImageURL
	}

	if req.PlayCount != nil {
		music.PlayCount = req.PlayCount
	}

	if req.LikeCount != nil {
		music.LikeCount = req.LikeCount
	}

	if err := db.Save(&music).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate music",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Music berhasil diupdate",
		"data":    music,
	})
}

// DeleteMusicHandler menghapus music (soft delete)
func DeleteMusicHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var music models.Music
	if err := db.First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&music).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus music",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Music berhasil dihapus",
	})
}

// IncrementPlayCountHandler menambah play count
func IncrementPlayCountHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var music models.Music
	if err := db.First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	// Increment play count
	if music.PlayCount == nil {
		count := 1
		music.PlayCount = &count
	} else {
		*music.PlayCount++
	}

	if err := db.Save(&music).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate play count",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Play count berhasil diupdate",
		"data":    music,
	})
}

// IncrementLikeCountHandler menambah like count
func IncrementLikeCountHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var music models.Music
	if err := db.First(&music, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	// Increment like count
	if music.LikeCount == nil {
		count := 1
		music.LikeCount = &count
	} else {
		*music.LikeCount++
	}

	if err := db.Save(&music).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate like count",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Like count berhasil diupdate",
		"data":    music,
	})
}
