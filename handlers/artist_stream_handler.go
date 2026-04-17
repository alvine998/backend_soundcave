package handlers

import (
	"backend_soundcave/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// StartStreamRequest struct untuk request start stream
type StartStreamRequest struct {
	Title       string     `json:"title" validate:"required"`
	Description string     `json:"description"`
	Thumbnail   *string    `json:"thumbnail"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// generateRandomKey generates a random string for the stream key
func generateRandomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// StartStreamHandler menangani artist mulai streaming
// @Summary      Start live stream
// @Description  Start a new live stream for an artist
// @Tags         ArtistStreams
// @Accept       json
// @Produce      json
// @Param        request  body      StartStreamRequest  true  "Start Stream Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artist-streams/start [post]
func StartStreamHandler(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	// Validasi role - independent, label, atau admin yang bisa stream
	if role != string(models.RoleIndependent) && role != string(models.RoleLabel) && role != string(models.RoleAdmin) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya artist, label, atau admin yang dapat memulai live stream",
		})
	}

	var req StartStreamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Get or create artist record
	var artist models.Artist
	var artistID uint

	fmt.Printf("DEBUG: userID=%d, role=%s\n", userID, role)

	// For independent and label roles, find artist by ref_user_id
	if role == string(models.RoleIndependent) || role == string(models.RoleLabel) {
		fmt.Printf("DEBUG: Looking for artist with ref_user_id=%d\n", userID)
		if err := db.Where("ref_user_id = ?", userID).First(&artist).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("DEBUG: Artist not found, creating new one\n")
				// Artist belum ada, buat baru
				var user models.User
				if err := db.First(&user, userID).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Gagal mengambil data user",
						"error":   err.Error(),
					})
				}

				artist = models.Artist{
					RefUserID:    &userID,
					Name:         user.FullName,
					Email:        user.Email,
					Phone:        user.Phone,
					ProfileImage: user.ProfileImage,
					Bio:          "Independent Artist",
				}
				if err := db.Create(&artist).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Gagal membuat record artist",
						"error":   err.Error(),
					})
				}
				fmt.Printf("DEBUG: New artist created with ID=%d\n", artist.ID)
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal mengambil data artist",
					"error":   err.Error(),
				})
			}
		}
		artistID = artist.ID
		fmt.Printf("DEBUG: Set artistID=%d for independent/label role\n", artistID)
	} else if role == string(models.RoleAdmin) {
		// For admin, create artist with ref_user_id = userID
		fmt.Printf("DEBUG: Looking for admin artist with ref_user_id=%d\n", userID)
		if err := db.Where("ref_user_id = ?", userID).First(&artist).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("DEBUG: Admin artist not found, creating new one\n")
				var user models.User
				if err := db.First(&user, userID).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Gagal mengambil data user",
						"error":   err.Error(),
					})
				}

				artist = models.Artist{
					RefUserID:    &userID,
					Name:         user.FullName,
					Email:        user.Email,
					Phone:        user.Phone,
					ProfileImage: user.ProfileImage,
					Bio:          "Official Administrator",
				}
				if err := db.Create(&artist).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"message": "Gagal membuat record artist",
						"error":   err.Error(),
					})
				}
				fmt.Printf("DEBUG: New admin artist created with ID=%d\n", artist.ID)
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Gagal mengambil data artist",
					"error":   err.Error(),
				})
			}
		}
		artistID = artist.ID
		fmt.Printf("DEBUG: Set artistID=%d for admin role\n", artistID)
	}

	// Validate artist ID was set
	if artistID == 0 {
		fmt.Printf("DEBUG: artistID is 0! role=%s, checking conditions\n", role)
		fmt.Printf("DEBUG: RoleIndependent=%s, RoleLabel=%s, RoleAdmin=%s\n",
			string(models.RoleIndependent), string(models.RoleLabel), string(models.RoleAdmin))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mendapatkan ID artist. Role tidak dikenali atau tidak memiliki akses untuk streaming.",
			"debug": fiber.Map{
				"role":   role,
				"userID": userID,
			},
		})
	}

	// Cek apakah ada stream yang masih live
	var activeStream models.ArtistStream
	if err := db.Where("artist_id = ? AND status = ?", artistID, models.StreamStatusLive).First(&activeStream).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Anda sudah memiliki stream yang sedang berlangsung",
			"data":    activeStream,
		})
	}

	// Generate unique stream key
	streamKey := fmt.Sprintf("%d_%s", userID, generateRandomKey(8))

	// Get base URLs from environment variables
	rtmpBaseURL := os.Getenv("RTMP_SERVER_URL")
	if rtmpBaseURL == "" {
		rtmpBaseURL = "rtmp://localhost/live"
	}

	hlsBaseURL := os.Getenv("HLS_SERVER_URL")
	if hlsBaseURL == "" {
		hlsBaseURL = "http://localhost:8080/hls"
	}

	// Determine status: if scheduled_at is provided and in the future, set as scheduled
	status := models.StreamStatusLive
	var startedAt time.Time
	if req.ScheduledAt != nil && req.ScheduledAt.After(time.Now()) {
		status = models.StreamStatusScheduled
	} else {
		startedAt = time.Now()
	}

	// Buat stream baru dengan artist_id yang benar
	stream := models.ArtistStream{
		ArtistID:    int32(artistID),
		Title:       req.Title,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		ScheduledAt: req.ScheduledAt,
		StreamKey:   streamKey,
		IngestURL:   fmt.Sprintf("%s/%s", rtmpBaseURL, streamKey),
		PlaybackURL: fmt.Sprintf("%s/%s.m3u8", hlsBaseURL, streamKey),
		Status:      status,
		StartedAt:   startedAt,
	}

	if err := db.Create(&stream).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memulai stream",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Stream berhasil dimulai",
		"data":    stream,
	})
}

// EndStreamHandler menangani artist mengakhiri streaming
// @Summary      End live stream
// @Description  End an active live stream for an artist
// @Tags         ArtistStreams
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Stream ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artist-streams/end/{id} [post]
func EndStreamHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(uint)

	var stream models.ArtistStream
	if err := db.Preload("Artist").First(&stream, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Stream tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data stream",
			"error":   err.Error(),
		})
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	// Pastikan yang mengakhiri adalah pemilik stream atau admin
	isOwner := stream.Artist.RefUserID != nil && *stream.Artist.RefUserID == userID
	isAdmin := user.Role == models.RoleAdmin

	if !isOwner && !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki akses untuk mengakhiri stream ini",
		})
	}

	// Update status
	now := time.Now()

	if err := db.Model(&stream).Updates(map[string]interface{}{
		"status":   models.StreamStatusEnded,
		"ended_at": now,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengakhiri stream",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Stream berhasil diakhiri",
		"data":    stream,
	})
}

// GetActiveStreamsHandler mendapatkan list stream yang sedang live
// @Summary      Get active streams
// @Description  Get list of currently active live streams
// @Tags         ArtistStreams
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artist-streams/active [get]
func GetActiveStreamsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var streams []models.ArtistStream

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	query := db.Preload("Artist").Where("status = ?", models.StreamStatusLive).Order("viewer_count desc")

	var total int64
	db.Model(&models.ArtistStream{}).Where("status = ?", models.StreamStatusLive).Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&streams).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data stream aktif",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    streams,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetStreamDetailsHandler mendapatkan detail stream
// @Summary      Get stream details
// @Description  Get details of a specific live stream
// @Tags         ArtistStreams
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Stream ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artist-streams/{id} [get]
func GetStreamDetailsHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var stream models.ArtistStream
	if err := db.Preload("Artist").First(&stream, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Stream tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data stream",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    stream,
	})
}

// GetArtistStreamHistoryHandler mendapatkan history streaming artist dengan pagination dan filter
// @Summary      Get artist stream history
// @Description  Get history of artist streams with pagination and filters for status and artist
// @Tags         ArtistStreams
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "Page number" default(1)
// @Param        limit      query     int     false  "Items per page" default(10)
// @Param        status     query     string  false  "Stream status filter (scheduled, live, ended)"
// @Param        artist_id  query     int     false  "Artist ID filter"
// @Success      200        {object}  map[string]interface{}
// @Failure      400        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artist-streams/history [get]
func GetArtistStreamHistoryHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Get pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status")
	artistID := c.QueryInt("artist_id", 0)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := db.Preload("Artist").Order("created_at DESC")

	// Filter by status if provided
	if status != "" {
		validStatuses := []string{string(models.StreamStatusScheduled), string(models.StreamStatusLive), string(models.StreamStatusEnded)}
		statusFound := false
		for _, s := range validStatuses {
			if s == status {
				statusFound = true
				break
			}
		}
		if !statusFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid status. Must be one of: scheduled, live, ended",
			})
		}
		query = query.Where("status = ?", status)
	}

	// Filter by artist_id if provided
	if artistID > 0 {
		query = query.Where("artist_id = ?", artistID)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.ArtistStream{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to count streams",
			"error":   err.Error(),
		})
	}

	// Get paginated results
	var streams []models.ArtistStream
	if err := query.Offset(offset).Limit(limit).Find(&streams).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve stream history",
			"error":   err.Error(),
		})
	}

	totalPages := (int(total) + limit - 1) / limit

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    streams,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
