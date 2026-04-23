package handlers

import (
	"backend_soundcave/models"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// srsCallbackBody is the JSON body sent by SRS for all HTTP callbacks.
type srsCallbackBody struct {
	Action   string `json:"action"`
	Stream   string `json:"stream"` // stream key
	App      string `json:"app"`
	ClientID string `json:"client_id"`
	Vhost    string `json:"vhost"`
	IP       string `json:"ip"`
	Param    string `json:"param"` // query params, e.g. "?secret=xxx"
}

// srsResponse returns the standard SRS success/failure response.
// SRS treats code != 0 as a rejection and will disconnect the publisher.
func srsResponse(c *fiber.Ctx, code int, msg string) error {
	return c.JSON(fiber.Map{"code": code, "msg": msg})
}

// validateSRSSecret checks the shared secret passed by SRS in the stream param.
// Set SRS_CALLBACK_SECRET env var on both the backend and the SRS config.
// If the env var is empty, secret validation is skipped.
func validateSRSSecret(param string) bool {
	secret := os.Getenv("SRS_CALLBACK_SECRET")
	if secret == "" {
		return true
	}
	// param looks like "?secret=xxx" or "&secret=xxx"
	expected := "secret=" + secret
	return len(param) > 0 && containsParam(param, expected)
}

func containsParam(param, kv string) bool {
	// Simple substring match is fine for a fixed key=value pair.
	for i := 0; i <= len(param)-len(kv); i++ {
		if param[i:i+len(kv)] == kv {
			return true
		}
	}
	return false
}

// OnPublishHandler handles SRS on_publish callback.
// SRS calls this when a publisher connects and starts streaming.
// Marks the stream as LIVE and records started_at.
//
// @Summary      SRS on_publish callback
// @Description  Called by SRS when a stream publisher connects. Marks stream as live.
// @Tags         SRS Webhooks
// @Accept       json
// @Produce      json
// @Param        body  body  srsCallbackBody  true  "SRS callback payload"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      403   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Router       /srs/on_publish [post]
func OnPublishHandler(c *fiber.Ctx, db *gorm.DB) error {
	var body srsCallbackBody
	if err := c.BodyParser(&body); err != nil {
		return srsResponse(c, 400, "bad request: cannot parse body")
	}

	if body.Stream == "" {
		return srsResponse(c, 400, "bad request: stream key missing")
	}

	if !validateSRSSecret(body.Param) {
		return srsResponse(c, 403, "forbidden: invalid secret")
	}

	var stream models.ArtistStream
	if err := db.Where("stream_key = ?", body.Stream).First(&stream).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return srsResponse(c, 404, "not found: invalid stream key")
		}
		return srsResponse(c, 500, "internal error")
	}

	now := time.Now()
	if err := db.Model(&stream).Updates(map[string]interface{}{
		"status":     models.StreamStatusLive,
		"started_at": now,
	}).Error; err != nil {
		return srsResponse(c, 500, "failed to update stream status")
	}

	return srsResponse(c, 0, "ok")
}

// OnUnpublishHandler handles SRS on_unpublish callback.
// SRS calls this when a publisher disconnects or the stream ends.
// Marks the stream as ENDED and records ended_at.
//
// @Summary      SRS on_unpublish callback
// @Description  Called by SRS when a stream publisher disconnects. Marks stream as ended.
// @Tags         SRS Webhooks
// @Accept       json
// @Produce      json
// @Param        body  body  srsCallbackBody  true  "SRS callback payload"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      404   {object}  map[string]interface{}
// @Router       /srs/on_unpublish [post]
func OnUnpublishHandler(c *fiber.Ctx, db *gorm.DB) error {
	var body srsCallbackBody
	if err := c.BodyParser(&body); err != nil {
		return srsResponse(c, 400, "bad request: cannot parse body")
	}

	if body.Stream == "" {
		return srsResponse(c, 400, "bad request: stream key missing")
	}

	var stream models.ArtistStream
	if err := db.Where("stream_key = ?", body.Stream).First(&stream).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return srsResponse(c, 404, "not found: invalid stream key")
		}
		return srsResponse(c, 500, "internal error")
	}

	now := time.Now()
	if err := db.Model(&stream).Updates(map[string]interface{}{
		"status":   models.StreamStatusEnded,
		"ended_at": now,
	}).Error; err != nil {
		return srsResponse(c, 500, "failed to update stream status")
	}

	return srsResponse(c, 0, "ok")
}
