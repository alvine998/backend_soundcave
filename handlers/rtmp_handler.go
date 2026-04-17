package handlers

import (
	"backend_soundcave/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// OnPublishHandler handles Nginx RTMP on_publish callback
func OnPublishHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Nginx sends data as form-data
	streamKey := c.FormValue("name")
	if streamKey == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request: stream key missing")
	}

	var stream models.ArtistStream
	if err := db.Where("stream_key = ?", streamKey).First(&stream).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).SendString("Not Found: invalid stream key")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Error")
	}

	// Update stream status to live and set start time
	now := time.Now()
	if err := db.Model(&stream).Updates(map[string]interface{}{
		"status":     models.StreamStatusLive,
		"started_at": now,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Update Failed")
	}

	return c.Status(fiber.StatusOK).SendString("OK")
}

// OnPublishDoneHandler handles Nginx RTMP on_publish_done callback
func OnPublishDoneHandler(c *fiber.Ctx, db *gorm.DB) error {
	streamKey := c.FormValue("name")
	if streamKey == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
	}

	var stream models.ArtistStream
	if err := db.Where("stream_key = ?", streamKey).First(&stream).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Not Found")
	}

	// Update status to ended
	now := time.Now()
	if err := db.Model(&stream).Updates(map[string]interface{}{
		"status":   models.StreamStatusEnded,
		"ended_at": now,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Update Failed")
	}

	return c.Status(fiber.StatusOK).SendString("OK")
}
