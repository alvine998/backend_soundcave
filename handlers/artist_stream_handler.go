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

	// Validasi role - hanya independent atau label yang bisa stream
	if role != string(models.RoleIndependent) && role != string(models.RoleLabel) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya artist atau label yang dapat memulai live stream",
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

	// Cek apakah ada stream yang masih live
	var activeStream models.ArtistStream
	if err := db.Where("artist_id = ? AND status = ?", userID, models.StreamStatusLive).First(&activeStream).Error; err == nil {
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

	// Buat stream baru
	stream := models.ArtistStream{
		ArtistID:    int32(userID),
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
	if err := db.First(&stream, id).Error; err != nil {
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

	// Pastikan yang mengakhiri adalah pemilik stream
	if stream.ArtistID != int32(userID) {
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
