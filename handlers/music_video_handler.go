package handlers

import (
	"backend_soundcave/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateMusicVideoRequest struct untuk request create music_video
type CreateMusicVideoRequest struct {
	Title       string  `json:"title" validate:"required"`
	ArtistID    int     `json:"artist_id" validate:"required"`
	Artist      string  `json:"artist" validate:"required"`
	ReleaseDate string  `json:"release_date" validate:"required"` // Format: "2006-01-02"
	Duration    string  `json:"duration" validate:"required"`     // Format: MM:SS atau HH:MM:SS
	Genre       string  `json:"genre" validate:"required"`
	Description *string `json:"description"`
	VideoURL    string  `json:"video_url" validate:"required"`
	Thumbnail   *string `json:"thumbnail"`
	SubmittedBy string  `json:"submitted_by"`
}

// UpdateMusicVideoRequest struct untuk request update music_video
type UpdateMusicVideoRequest struct {
	Title       *string `json:"title"`
	ArtistID    *int    `json:"artist_id"`
	Artist      *string `json:"artist"`
	ReleaseDate *string `json:"release_date"` // Format: "2006-01-02"
	Duration    *string `json:"duration"`     // Format: MM:SS atau HH:MM:SS
	Genre       *string `json:"genre"`
	Description *string `json:"description"`
	VideoURL    *string `json:"video_url"`
	Thumbnail   *string `json:"thumbnail"`
	SubmittedBy *string `json:"submitted_by"`
	IsApproved  *int    `json:"is_approved"`
	ApprovedBy  *int    `json:"approved_by"`
	IsHighlight *int    `json:"is_highlight"`
}

// ApproveMusicVideoRequest struct untuk request approve music video
type ApproveMusicVideoRequest struct {
	UserID uint `json:"user_id" validate:"required"`
}

// CreateMusicVideoHandler membuat music_video baru
// @Summary      Create new music video
// @Description  Create a new music video
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        request  body      CreateMusicVideoRequest  true  "Music Video Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos [post]
func CreateMusicVideoHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateMusicVideoRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse release date (optional)
	var releaseDatePtr *time.Time
	if req.ReleaseDate != "" {
		releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		releaseDatePtr = &releaseDate
	}

	// Buat music_video baru
	musicVideo := models.MusicVideo{
		Title:       req.Title,
		ArtistID:    req.ArtistID,
		Artist:      req.Artist,
		ReleaseDate: releaseDatePtr,
		Duration:    req.Duration,
		Genre:       req.Genre,
		Description: req.Description,
		VideoURL:    req.VideoURL,
		Thumbnail:   req.Thumbnail,
		SubmittedBy: "artist",
		IsApproved:  new(int), // Default 0
		IsHighlight: new(int), // Default 0
	}
	*musicVideo.IsApproved = 0
	*musicVideo.IsHighlight = 0

	// Validate submitted_by
	if req.SubmittedBy != "" {
		if req.SubmittedBy != "artist" && req.SubmittedBy != "label" && req.SubmittedBy != "admin" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "submitted_by harus salah satu dari: artist, label, admin",
			})
		}
		musicVideo.SubmittedBy = req.SubmittedBy
	}

	if err := db.Create(&musicVideo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat music video",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Music video berhasil dibuat",
		"data":    musicVideo,
	})
}

// GetMusicVideosHandler mendapatkan semua music_videos dengan pagination
// @Summary      Get all music videos
// @Description  Get paginated list of music videos with filtering and search
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "Page number" default(1)
// @Param        limit      query     int     false  "Items per page" default(10)
// @Param        artist_id  query     int     false  "Filter by artist ID"
// @Param        genre      query     string  false  "Filter by genre"
// @Param        search     query     string  false  "Search by title or artist"
// @Param        sort_by    query     string  false  "Sort field" default(created_at)
// @Param        order      query     string  false  "Sort order" default(desc)
// @Param        is_approved query    int     false  "Filter by approval status (0, 1, or 2)"
// @Param        is_highlight query   int     false  "Filter by highlight status (0 or 1)"
// @Param        submitted_by query   string  false  "Filter by submitter (artist, label, admin)"
// @Success      200        {object}  map[string]interface{}
// @Failure      401        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos [get]
func GetMusicVideosHandler(c *fiber.Ctx, db *gorm.DB) error {
	var musicVideos []models.MusicVideo

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.MusicVideo{})

	// Filter by artist_id jika ada
	if artistID := c.QueryInt("artist_id", 0); artistID > 0 {
		query = query.Where("artist_id = ?", artistID)
	}

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Search by title atau artist
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR artist LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Filter by is_approved jika ada
	if isApproved := c.Query("is_approved"); isApproved != "" {
		query = query.Where("is_approved = ?", isApproved)
	}

	// Filter by is_highlight jika ada
	if isHighlight := c.Query("is_highlight"); isHighlight != "" {
		query = query.Where("is_highlight = ?", isHighlight)
	}

	// Filter by submitted_by jika ada
	if submittedBy := c.Query("submitted_by"); submittedBy != "" {
		query = query.Where("submitted_by = ?", submittedBy)
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

	// Get music_videos
	if err := query.Offset(offset).Limit(limit).Find(&musicVideos).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music videos",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    musicVideos,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetMusicVideoHandler mendapatkan music_video by ID
// @Summary      Get music video by ID
// @Description  Get music video details by ID
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Music Video ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos/{id} [get]
func GetMusicVideoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var musicVideo models.MusicVideo
	if err := db.First(&musicVideo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music video tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music video",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    musicVideo,
	})
}

// UpdateMusicVideoHandler mengupdate music_video
// @Summary      Update music video
// @Description  Update music video information
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Music Video ID"
// @Param        request  body      UpdateMusicVideoRequest  true  "Update Music Video Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos/{id} [put]
func UpdateMusicVideoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var musicVideo models.MusicVideo
	if err := db.First(&musicVideo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music video tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music video",
			"error":   err.Error(),
		})
	}

	var req UpdateMusicVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		musicVideo.Title = *req.Title
	}

	if req.ArtistID != nil {
		musicVideo.ArtistID = *req.ArtistID
	}

	if req.Artist != nil {
		musicVideo.Artist = *req.Artist
	}

	if req.ReleaseDate != nil {
		if *req.ReleaseDate == "" {
			musicVideo.ReleaseDate = nil
		} else {
			releaseDate, err := time.Parse("2006-01-02", *req.ReleaseDate)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
					"error":   err.Error(),
				})
			}
			musicVideo.ReleaseDate = &releaseDate
		}
	}

	if req.Duration != nil {
		musicVideo.Duration = *req.Duration
	}

	if req.Genre != nil {
		musicVideo.Genre = *req.Genre
	}

	if req.Description != nil {
		musicVideo.Description = req.Description
	}

	if req.VideoURL != nil {
		musicVideo.VideoURL = *req.VideoURL
	}

	if req.Thumbnail != nil {
		musicVideo.Thumbnail = req.Thumbnail
	}

	if req.IsApproved != nil {
		musicVideo.IsApproved = req.IsApproved
	}

	if req.ApprovedBy != nil {
		musicVideo.ApprovedBy = req.ApprovedBy
	}

	if req.IsHighlight != nil {
		musicVideo.IsHighlight = req.IsHighlight
	}

	if req.SubmittedBy != nil {
		if *req.SubmittedBy != "artist" && *req.SubmittedBy != "label" && *req.SubmittedBy != "admin" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "submitted_by harus salah satu dari: artist, label, admin",
			})
		}
		musicVideo.SubmittedBy = *req.SubmittedBy
	}

	if err := db.Save(&musicVideo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate music video",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Music video berhasil diupdate",
		"data":    musicVideo,
	})
}

// DeleteMusicVideoHandler menghapus music_video (soft delete)
// @Summary      Delete music video
// @Description  Soft delete a music video by ID
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Music Video ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos/{id} [delete]
func DeleteMusicVideoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var musicVideo models.MusicVideo
	if err := db.First(&musicVideo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music video tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music video",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&musicVideo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus music video",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Music video berhasil dihapus",
	})
}

// IncrementMusicVideoStreamHandler menambah total stream music video
// @Summary      Increment stream count
// @Description  Increment total stream count for a music video
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Music Video ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos/{id}/stream [post]
func IncrementMusicVideoStreamHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var musicVideo models.MusicVideo
	if err := db.First(&musicVideo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music video tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music video",
			"error":   err.Error(),
		})
	}

	// Increment total stream
	if musicVideo.TotalStream == nil {
		count := 1
		musicVideo.TotalStream = &count
	} else {
		*musicVideo.TotalStream++
	}

	if err := db.Save(&musicVideo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate stream count",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Stream count berhasil diupdate",
		"data":    musicVideo,
	})
}

// ApproveMusicVideoHandler menyetujui music video
// @Summary      Approve music video
// @Description  Approve a music video by ID
// @Tags         MusicVideos
// @Accept       json
// @Produce      json
// @Param        id       path      int                     true  "Music Video ID"
// @Param        request  body      ApproveMusicVideoRequest  true  "Approve Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /music-videos/{id}/approve [put]
func ApproveMusicVideoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var musicVideo models.MusicVideo
	if err := db.First(&musicVideo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music video tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music video",
			"error":   err.Error(),
		})
	}

	var req ApproveMusicVideoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Set is_approved = 1 dan approved_by = user_id
	isApproved := 1
	approvedBy := int(req.UserID)
	musicVideo.IsApproved = &isApproved
	musicVideo.ApprovedBy = &approvedBy

	if err := db.Save(&musicVideo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyetujui music video",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Music video berhasil diapprove",
		"data":    musicVideo,
	})
}
