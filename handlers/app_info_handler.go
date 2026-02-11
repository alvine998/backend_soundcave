package handlers

import (
	"backend_soundcave/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateAppInfoRequest struct untuk request create app_info
type CreateAppInfoRequest struct {
	AppName     string                 `json:"app_name" validate:"required"`
	Tagline     string                 `json:"tagline" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	Version     *string                `json:"version"`
	LaunchDate  *string                `json:"launch_date"` // Format: "2006-01-02"
	Email       string                 `json:"email" validate:"required,email"`
	Phone       string                 `json:"phone" validate:"required"`
	Address     string                 `json:"address" validate:"required"`
	SocialMedia map[string]interface{} `json:"social_media"`
	AppLinks    map[string]interface{} `json:"app_links"`
	Legal       map[string]interface{} `json:"legal"`
	Features    map[string]interface{} `json:"features"`
	Stats       map[string]interface{} `json:"stats"`
}

// UpdateAppInfoRequest struct untuk request update app_info
type UpdateAppInfoRequest struct {
	AppName     *string                `json:"app_name"`
	Tagline     *string                `json:"tagline"`
	Description *string                `json:"description"`
	Version     *string                `json:"version"`
	LaunchDate  *string                `json:"launch_date"` // Format: "2006-01-02"
	Email       *string                `json:"email" validate:"omitempty,email"`
	Phone       *string                `json:"phone"`
	Address     *string                `json:"address"`
	SocialMedia map[string]interface{} `json:"social_media"`
	AppLinks    map[string]interface{} `json:"app_links"`
	Legal       map[string]interface{} `json:"legal"`
	Features    map[string]interface{} `json:"features"`
	Stats       map[string]interface{} `json:"stats"`
}

// CreateAppInfoHandler membuat app_info baru
// @Summary      Create new app info
// @Description  Create a new app information entry
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Param        request  body      CreateAppInfoRequest  true  "App Info Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /app-info [post]
func CreateAppInfoHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateAppInfoRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse launch date jika ada
	var launchDate *time.Time
	if req.LaunchDate != nil && *req.LaunchDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.LaunchDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		launchDate = &parsedDate
	}

	// Convert maps to JSONB
	var socialMedia models.JSONB
	if req.SocialMedia != nil {
		socialMedia = models.JSONB(req.SocialMedia)
	}

	var appLinks models.JSONB
	if req.AppLinks != nil {
		appLinks = models.JSONB(req.AppLinks)
	}

	var legal models.JSONB
	if req.Legal != nil {
		legal = models.JSONB(req.Legal)
	}

	var features models.JSONB
	if req.Features != nil {
		features = models.JSONB(req.Features)
	}

	var stats models.JSONB
	if req.Stats != nil {
		stats = models.JSONB(req.Stats)
	}

	// Buat app_info baru
	appInfo := models.AppInfo{
		AppName:     req.AppName,
		Tagline:     req.Tagline,
		Description: req.Description,
		Version:     req.Version,
		LaunchDate:  launchDate,
		Email:       req.Email,
		Phone:       req.Phone,
		Address:     req.Address,
		SocialMedia: socialMedia,
		AppLinks:    appLinks,
		Legal:       legal,
		Features:    features,
		Stats:       stats,
	}

	if err := db.Create(&appInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "App info berhasil dibuat",
		"data":    appInfo,
	})
}

// GetAppInfosHandler mendapatkan semua app_info dengan pagination
// @Summary      Get all app info
// @Description  Get paginated list of app information entries
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        search   query     string  false  "Search by app name or tagline"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /app-info [get]
func GetAppInfosHandler(c *fiber.Ctx, db *gorm.DB) error {
	var appInfos []models.AppInfo

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.AppInfo{})

	// Search by app_name atau tagline
	if search := c.Query("search"); search != "" {
		query = query.Where("app_name LIKE ? OR tagline LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get app_infos
	if err := query.Offset(offset).Limit(limit).Find(&appInfos).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    appInfos,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetAppInfoHandler mendapatkan app_info by ID
// @Summary      Get app info by ID
// @Description  Get app information details by ID
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "App Info ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /app-info/{id} [get]
func GetAppInfoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var appInfo models.AppInfo
	if err := db.First(&appInfo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "App info tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    appInfo,
	})
}

// GetLatestAppInfoHandler mendapatkan app_info terbaru (biasanya hanya satu)
// @Summary      Get latest app info
// @Description  Get the latest app information entry
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /app-info/latest [get]
func GetLatestAppInfoHandler(c *fiber.Ctx, db *gorm.DB) error {
	var appInfo models.AppInfo
	if err := db.Order("created_at DESC").First(&appInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "App info tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    appInfo,
	})
}

// UpdateAppInfoHandler mengupdate app_info
// @Summary      Update app info
// @Description  Update app information
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "App Info ID"
// @Param        request  body      UpdateAppInfoRequest  true  "Update App Info Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /app-info/{id} [put]
func UpdateAppInfoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var appInfo models.AppInfo
	if err := db.First(&appInfo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "App info tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data app_info",
			"error":   err.Error(),
		})
	}

	var req UpdateAppInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.AppName != nil {
		appInfo.AppName = *req.AppName
	}

	if req.Tagline != nil {
		appInfo.Tagline = *req.Tagline
	}

	if req.Description != nil {
		appInfo.Description = *req.Description
	}

	if req.Version != nil {
		appInfo.Version = req.Version
	}

	if req.LaunchDate != nil && *req.LaunchDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.LaunchDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		appInfo.LaunchDate = &parsedDate
	}

	if req.Email != nil {
		appInfo.Email = *req.Email
	}

	if req.Phone != nil {
		appInfo.Phone = *req.Phone
	}

	if req.Address != nil {
		appInfo.Address = *req.Address
	}

	if req.SocialMedia != nil {
		appInfo.SocialMedia = models.JSONB(req.SocialMedia)
	}

	if req.AppLinks != nil {
		appInfo.AppLinks = models.JSONB(req.AppLinks)
	}

	if req.Legal != nil {
		appInfo.Legal = models.JSONB(req.Legal)
	}

	if req.Features != nil {
		appInfo.Features = models.JSONB(req.Features)
	}

	if req.Stats != nil {
		appInfo.Stats = models.JSONB(req.Stats)
	}

	if err := db.Save(&appInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "App info berhasil diupdate",
		"data":    appInfo,
	})
}

// DeleteAppInfoHandler menghapus app_info (soft delete)
// @Summary      Delete app info
// @Description  Soft delete an app information entry by ID
// @Tags         AppInfo
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "App Info ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /app-info/{id} [delete]
func DeleteAppInfoHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var appInfo models.AppInfo
	if err := db.First(&appInfo, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "App info tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data app_info",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&appInfo).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus app_info",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "App info berhasil dihapus",
	})
}
