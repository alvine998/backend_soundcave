package handlers

import (
	"backend_soundcave/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateCavelistRequest struct untuk request create cavelist
type CreateCavelistRequest struct {
	Title           string  `json:"title" validate:"required"`
	Description     *string `json:"description"`
	IsPromotion     *bool   `json:"is_promotion"`
	ExpiryPromotion *string `json:"expiry_promotion"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
	VideoURL        string  `json:"video_url" validate:"required"`
	ArtistID        int     `json:"artist_id" validate:"required"`
	ArtistName      string  `json:"artist_name" validate:"required"`
	Status          *string `json:"status"`       // "draft" atau "publish"
	PublishedAt     *string `json:"published_at"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
}

// UpdateCavelistRequest struct untuk request update cavelist
type UpdateCavelistRequest struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	IsPromotion     *bool   `json:"is_promotion"`
	ExpiryPromotion *string `json:"expiry_promotion"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
	VideoURL        *string `json:"video_url"`
	ArtistID        *int    `json:"artist_id"`
	ArtistName      *string `json:"artist_name"`
	Status          *string `json:"status"`       // "draft" atau "publish"
	PublishedAt     *string `json:"published_at"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
}

// parseDateTime helper function untuk parse datetime dengan multiple formats
func parseDateTime(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if parsedDate, err := time.Parse(format, dateStr); err == nil {
			return &parsedDate, nil
		}
	}
	return nil, fiber.NewError(fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD HH:MM:SS atau YYYY-MM-DD")
}

// CreateCavelistHandler membuat cavelist baru
// @Summary      Create new cavelist
// @Description  Create a new cavelist
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        request  body      CreateCavelistRequest  true  "Cavelist Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /cavelists [post]
func CreateCavelistHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateCavelistRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse expiry_promotion jika ada
	var expiryPromotion *time.Time
	if req.ExpiryPromotion != nil && *req.ExpiryPromotion != "" {
		parsed, err := parseDateTime(*req.ExpiryPromotion)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		expiryPromotion = parsed
	}

	// Parse published_at jika ada
	var publishedAt *time.Time
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		parsed, err := parseDateTime(*req.PublishedAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		publishedAt = parsed
	}

	// Set default values
	isPromotion := false
	viewers := 0
	likes := 0
	shares := 0
	status := models.CavelistStatusDraft

	if req.IsPromotion != nil {
		isPromotion = *req.IsPromotion
	}

	if req.Status != nil {
		if *req.Status == "publish" {
			status = models.CavelistStatusPublish
		} else {
			status = models.CavelistStatusDraft
		}
	}

	// Buat cavelist baru
	cavelist := models.Cavelist{
		Title:           req.Title,
		Description:     req.Description,
		IsPromotion:     &isPromotion,
		ExpiryPromotion: expiryPromotion,
		Viewers:         &viewers,
		Likes:           &likes,
		Shares:          &shares,
		VideoURL:        req.VideoURL,
		ArtistID:        req.ArtistID,
		ArtistName:      req.ArtistName,
		Status:          status,
		PublishedAt:     publishedAt,
	}

	if err := db.Create(&cavelist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat cavelist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Cavelist berhasil dibuat",
		"data":    cavelist,
	})
}

// GetCavelistsHandler mendapatkan semua cavelist dengan pagination
// @Summary      Get all cavelists
// @Description  Get paginated list of cavelists with filtering and search
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "Page number" default(1)
// @Param        limit        query     int     false  "Items per page" default(10)
// @Param        artist_id    query     int     false  "Filter by artist_id"
// @Param        status       query     string  false  "Filter by status (draft/publish)"
// @Param        is_promotion query     bool    false  "Filter by is_promotion"
// @Param        search       query     string  false  "Search by title or description"
// @Param        sort_by      query     string  false  "Sort field" default(created_at)
// @Param        order        query     string  false  "Sort order" default(desc)
// @Success      200          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /cavelists [get]
func GetCavelistsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var cavelists []models.Cavelist

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Cavelist{})

	// Filter by artist_id jika ada
	if artistID := c.Query("artist_id"); artistID != "" {
		if artistIDInt, err := strconv.Atoi(artistID); err == nil {
			query = query.Where("artist_id = ?", artistIDInt)
		}
	}

	// Filter by status jika ada
	if status := c.Query("status"); status != "" {
		if status == "draft" || status == "publish" {
			query = query.Where("status = ?", status)
		}
	}

	// Filter by is_promotion jika ada
	if isPromotion := c.Query("is_promotion"); isPromotion != "" {
		isPromotionBool, err := strconv.ParseBool(isPromotion)
		if err == nil {
			query = query.Where("is_promotion = ?", isPromotionBool)
		}
	}

	// Search by title atau description
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Sort
	sortBy := c.Query("sort_by", "created_at")
	order := c.Query("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	query = query.Order(sortBy + " " + order)

	// Get total count
	var total int64
	query.Count(&total)

	// Get cavelists
	if err := query.Offset(offset).Limit(limit).Find(&cavelists).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    cavelists,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetCavelistByIDHandler mendapatkan cavelist by ID
// @Summary      Get cavelist by ID
// @Description  Get a single cavelist by ID
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Cavelist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cavelists/{id} [get]
func GetCavelistByIDHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var cavelist models.Cavelist
	if err := db.First(&cavelist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Cavelist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	// Increment viewers
	if cavelist.Viewers != nil {
		*cavelist.Viewers++
	} else {
		viewers := 1
		cavelist.Viewers = &viewers
	}
	db.Save(&cavelist)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    cavelist,
	})
}

// UpdateCavelistHandler mengupdate cavelist
// @Summary      Update cavelist
// @Description  Update cavelist information
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "Cavelist ID"
// @Param        request  body      UpdateCavelistRequest  true  "Update Cavelist Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /cavelists/{id} [put]
func UpdateCavelistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var cavelist models.Cavelist
	if err := db.First(&cavelist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Cavelist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	var req UpdateCavelistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		cavelist.Title = *req.Title
	}

	if req.Description != nil {
		cavelist.Description = req.Description
	}

	if req.IsPromotion != nil {
		cavelist.IsPromotion = req.IsPromotion
	}

	if req.ExpiryPromotion != nil && *req.ExpiryPromotion != "" {
		parsed, err := parseDateTime(*req.ExpiryPromotion)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		cavelist.ExpiryPromotion = parsed
	}

	if req.VideoURL != nil {
		cavelist.VideoURL = *req.VideoURL
	}

	if req.ArtistID != nil {
		cavelist.ArtistID = *req.ArtistID
	}

	if req.ArtistName != nil {
		cavelist.ArtistName = *req.ArtistName
	}

	if req.Status != nil {
		if *req.Status == "publish" {
			cavelist.Status = models.CavelistStatusPublish
		} else {
			cavelist.Status = models.CavelistStatusDraft
		}
	}

	if req.PublishedAt != nil && *req.PublishedAt != "" {
		parsed, err := parseDateTime(*req.PublishedAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}
		cavelist.PublishedAt = parsed
	}

	if err := db.Save(&cavelist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate cavelist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Cavelist berhasil diupdate",
		"data":    cavelist,
	})
}

// DeleteCavelistHandler menghapus cavelist (soft delete)
// @Summary      Delete cavelist
// @Description  Soft delete a cavelist by ID
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Cavelist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cavelists/{id} [delete]
func DeleteCavelistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var cavelist models.Cavelist
	if err := db.First(&cavelist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Cavelist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&cavelist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus cavelist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Cavelist berhasil dihapus",
	})
}

// IncrementCavelistLikesHandler menambah likes cavelist
// @Summary      Increment cavelist likes
// @Description  Increment likes count for a cavelist
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Cavelist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cavelists/{id}/like [post]
func IncrementCavelistLikesHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var cavelist models.Cavelist
	if err := db.First(&cavelist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Cavelist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	// Increment likes
	if cavelist.Likes != nil {
		*cavelist.Likes++
	} else {
		likes := 1
		cavelist.Likes = &likes
	}

	if err := db.Save(&cavelist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate likes",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Likes berhasil diupdate",
		"data":    cavelist,
	})
}

// IncrementCavelistSharesHandler menambah shares cavelist
// @Summary      Increment cavelist shares
// @Description  Increment shares count for a cavelist
// @Tags         Cavelists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Cavelist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /cavelists/{id}/share [post]
func IncrementCavelistSharesHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var cavelist models.Cavelist
	if err := db.First(&cavelist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Cavelist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data cavelist",
			"error":   err.Error(),
		})
	}

	// Increment shares
	if cavelist.Shares != nil {
		*cavelist.Shares++
	} else {
		shares := 1
		cavelist.Shares = &shares
	}

	if err := db.Save(&cavelist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate shares",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Shares berhasil diupdate",
		"data":    cavelist,
	})
}
