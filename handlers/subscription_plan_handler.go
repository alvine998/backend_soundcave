package handlers

import (
	"backend_soundcave/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateSubscriptionPlanRequest struct untuk request create subscription_plan
type CreateSubscriptionPlanRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Price        string                 `json:"price" validate:"required"`
	Duration     string                 `json:"duration" validate:"required"`
	Features     map[string]interface{} `json:"features" validate:"required"`
	MaxDownloads int                    `json:"max_downloads"` // -1 means unlimited
	MaxPlaylists int                    `json:"max_playlists"` // -1 means unlimited
	AudioQuality string                 `json:"audio_quality" validate:"required"`
	AdsEnabled   *bool                  `json:"ads_enabled"`
	OfflineMode  *bool                  `json:"offline_mode"`
	IsPopular    *bool                  `json:"is_popular"`
	Description  string                 `json:"description" validate:"required"`
}

// UpdateSubscriptionPlanRequest struct untuk request update subscription_plan
type UpdateSubscriptionPlanRequest struct {
	Name         *string                `json:"name"`
	Price        *string                `json:"price"`
	Duration     *string                `json:"duration"`
	Features     map[string]interface{} `json:"features"`
	MaxDownloads *int                   `json:"max_downloads"`
	MaxPlaylists *int                   `json:"max_playlists"`
	AudioQuality *string                `json:"audio_quality"`
	AdsEnabled   *bool                  `json:"ads_enabled"`
	OfflineMode  *bool                  `json:"offline_mode"`
	IsPopular    *bool                  `json:"is_popular"`
	Description  *string                `json:"description"`
}

// CreateSubscriptionPlanHandler membuat subscription_plan baru
func CreateSubscriptionPlanHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateSubscriptionPlanRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Set default values
	adsEnabled := true
	offlineMode := false
	isPopular := false

	if req.AdsEnabled != nil {
		adsEnabled = *req.AdsEnabled
	}
	if req.OfflineMode != nil {
		offlineMode = *req.OfflineMode
	}
	if req.IsPopular != nil {
		isPopular = *req.IsPopular
	}

	// Set default untuk max downloads dan playlists jika tidak diisi
	maxDownloads := req.MaxDownloads
	maxPlaylists := req.MaxPlaylists
	if maxDownloads == 0 {
		maxDownloads = -1
	}
	if maxPlaylists == 0 {
		maxPlaylists = -1
	}

	// Convert features map to JSONB
	var features models.JSONB
	if req.Features != nil {
		features = models.JSONB(req.Features)
	}

	// Buat subscription_plan baru
	subscriptionPlan := models.SubscriptionPlan{
		Name:         req.Name,
		Price:        req.Price,
		Duration:     req.Duration,
		Features:     features,
		MaxDownloads: maxDownloads,
		MaxPlaylists: maxPlaylists,
		AudioQuality: req.AudioQuality,
		AdsEnabled:   adsEnabled,
		OfflineMode:  offlineMode,
		IsPopular:    &isPopular,
		Description:  req.Description,
	}

	if err := db.Create(&subscriptionPlan).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat subscription plan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Subscription plan berhasil dibuat",
		"data":    subscriptionPlan,
	})
}

// GetSubscriptionPlansHandler mendapatkan semua subscription_plans dengan pagination
func GetSubscriptionPlansHandler(c *fiber.Ctx, db *gorm.DB) error {
	var subscriptionPlans []models.SubscriptionPlan

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.SubscriptionPlan{})

	// Filter by audio_quality jika ada
	if audioQuality := c.Query("audio_quality"); audioQuality != "" {
		query = query.Where("audio_quality = ?", audioQuality)
	}

	// Filter by ads_enabled jika ada
	if adsEnabled := c.Query("ads_enabled"); adsEnabled != "" {
		adsEnabledBool, err := strconv.ParseBool(adsEnabled)
		if err == nil {
			query = query.Where("ads_enabled = ?", adsEnabledBool)
		}
	}

	// Filter by offline_mode jika ada
	if offlineMode := c.Query("offline_mode"); offlineMode != "" {
		offlineModeBool, err := strconv.ParseBool(offlineMode)
		if err == nil {
			query = query.Where("offline_mode = ?", offlineModeBool)
		}
	}

	// Filter by is_popular jika ada
	if isPopular := c.Query("is_popular"); isPopular != "" {
		isPopularBool, err := strconv.ParseBool(isPopular)
		if err == nil {
			query = query.Where("is_popular = ?", isPopularBool)
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

	// Get subscription_plans
	if err := query.Offset(offset).Limit(limit).Find(&subscriptionPlans).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data subscription plans",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    subscriptionPlans,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetSubscriptionPlanHandler mendapatkan subscription_plan by ID
func GetSubscriptionPlanHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var subscriptionPlan models.SubscriptionPlan
	if err := db.First(&subscriptionPlan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Subscription plan tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data subscription plan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    subscriptionPlan,
	})
}

// UpdateSubscriptionPlanHandler mengupdate subscription_plan
func UpdateSubscriptionPlanHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var subscriptionPlan models.SubscriptionPlan
	if err := db.First(&subscriptionPlan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Subscription plan tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data subscription plan",
			"error":   err.Error(),
		})
	}

	var req UpdateSubscriptionPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Name != nil {
		subscriptionPlan.Name = *req.Name
	}

	if req.Price != nil {
		subscriptionPlan.Price = *req.Price
	}

	if req.Duration != nil {
		subscriptionPlan.Duration = *req.Duration
	}

	if req.Features != nil {
		subscriptionPlan.Features = models.JSONB(req.Features)
	}

	if req.MaxDownloads != nil {
		subscriptionPlan.MaxDownloads = *req.MaxDownloads
	}

	if req.MaxPlaylists != nil {
		subscriptionPlan.MaxPlaylists = *req.MaxPlaylists
	}

	if req.AudioQuality != nil {
		subscriptionPlan.AudioQuality = *req.AudioQuality
	}

	if req.AdsEnabled != nil {
		subscriptionPlan.AdsEnabled = *req.AdsEnabled
	}

	if req.OfflineMode != nil {
		subscriptionPlan.OfflineMode = *req.OfflineMode
	}

	if req.IsPopular != nil {
		subscriptionPlan.IsPopular = req.IsPopular
	}

	if req.Description != nil {
		subscriptionPlan.Description = *req.Description
	}

	if err := db.Save(&subscriptionPlan).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate subscription plan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subscription plan berhasil diupdate",
		"data":    subscriptionPlan,
	})
}

// DeleteSubscriptionPlanHandler menghapus subscription_plan (soft delete)
func DeleteSubscriptionPlanHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var subscriptionPlan models.SubscriptionPlan
	if err := db.First(&subscriptionPlan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Subscription plan tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data subscription plan",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&subscriptionPlan).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus subscription plan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subscription plan berhasil dihapus",
	})
}

