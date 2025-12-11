package handlers

import (
	"backend_soundcave/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateNewsRequest struct untuk request create news
type CreateNewsRequest struct {
	Title       string  `json:"title" validate:"required"`
	Content     string  `json:"content" validate:"required"`
	Summary     *string `json:"summary"`
	Author      string  `json:"author" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	ImageURL    *string `json:"image_url"`
	PublishedAt *string `json:"published_at"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
	IsPublished *bool   `json:"is_published"`
	IsHeadline  *bool   `json:"is_headline"`
	Tags        *string `json:"tags"`
}

// UpdateNewsRequest struct untuk request update news
type UpdateNewsRequest struct {
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	Summary     *string `json:"summary"`
	Author      *string `json:"author"`
	Category    *string `json:"category"`
	ImageURL    *string `json:"image_url"`
	PublishedAt *string `json:"published_at"` // Format: "2006-01-02 15:04:05" atau "2006-01-02T15:04:05Z"
	IsPublished *bool   `json:"is_published"`
	IsHeadline  *bool   `json:"is_headline"`
	Tags        *string `json:"tags"`
}

// CreateNewsHandler membuat news baru
// @Summary      Create new news
// @Description  Create a new news article
// @Tags         News
// @Accept       json
// @Produce      json
// @Param        request  body      CreateNewsRequest  true  "News Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /news [post]
func CreateNewsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateNewsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse published_at jika ada
	var publishedAt *time.Time
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		// Try multiple date formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		var err error
		for _, format := range formats {
			parsedDate, parseErr := time.Parse(format, *req.PublishedAt)
			if parseErr == nil {
				publishedAt = &parsedDate
				err = nil
				break
			}
			err = parseErr
		}
		if err != nil && publishedAt == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD HH:MM:SS atau YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
	}

	// Set default values
	isPublished := false
	isHeadline := false
	views := 0
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}
	if req.IsHeadline != nil {
		isHeadline = *req.IsHeadline
	}

	// Buat news baru
	news := models.News{
		Title:       req.Title,
		Content:     req.Content,
		Summary:     req.Summary,
		Author:      req.Author,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		PublishedAt: publishedAt,
		IsPublished: &isPublished,
		IsHeadline:  &isHeadline,
		Views:       &views,
		Tags:        req.Tags,
	}

	if err := db.Create(&news).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat news",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "News berhasil dibuat",
		"data":    news,
	})
}

// GetNewsHandler mendapatkan semua news dengan pagination
// @Summary      Get all news
// @Description  Get paginated list of news articles with filtering and search
// @Tags         News
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "Page number" default(1)
// @Param        limit        query     int     false  "Items per page" default(10)
// @Param        category     query     string  false  "Filter by category"
// @Param        author       query     string  false  "Filter by author"
// @Param        published    query     bool    false  "Filter by published status"
// @Param        search       query     string  false  "Search by title or content"
// @Param        sort_by      query     string  false  "Sort field" default(created_at)
// @Param        order        query     string  false  "Sort order" default(desc)
// @Success      200          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /news [get]
func GetNewsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var news []models.News

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.News{})

	// Filter by category jika ada
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Filter by author jika ada
	if author := c.Query("author"); author != "" {
		query = query.Where("author LIKE ?", "%"+author+"%")
	}

	// Filter by is_published jika ada
	if isPublished := c.Query("is_published"); isPublished != "" {
		isPublishedBool, err := strconv.ParseBool(isPublished)
		if err == nil {
			query = query.Where("is_published = ?", isPublishedBool)
		}
	}

	// Search by title, content, atau summary
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR content LIKE ? OR summary LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
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

	// Get news
	if err := query.Offset(offset).Limit(limit).Find(&news).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data news",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    news,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetNewsByIDHandler mendapatkan news by ID
func GetNewsByIDHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var news models.News
	if err := db.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "News tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data news",
			"error":   err.Error(),
		})
	}

	// Increment views
	if news.Views != nil {
		*news.Views++
	} else {
		views := 1
		news.Views = &views
	}
	db.Save(&news)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    news,
	})
}

// UpdateNewsHandler mengupdate news
// @Summary      Update news
// @Description  Update news article information
// @Tags         News
// @Accept       json
// @Produce      json
// @Param        id       path      int              true  "News ID"
// @Param        request  body      UpdateNewsRequest  true  "Update News Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /news/{id} [put]
func UpdateNewsHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var news models.News
	if err := db.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "News tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data news",
			"error":   err.Error(),
		})
	}

	var req UpdateNewsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Title != nil {
		news.Title = *req.Title
	}

	if req.Content != nil {
		news.Content = *req.Content
	}

	if req.Summary != nil {
		news.Summary = req.Summary
	}

	if req.Author != nil {
		news.Author = *req.Author
	}

	if req.Category != nil {
		news.Category = *req.Category
	}

	if req.ImageURL != nil {
		news.ImageURL = req.ImageURL
	}

	if req.PublishedAt != nil && *req.PublishedAt != "" {
		// Try multiple date formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		var parsedDate *time.Time
		for _, format := range formats {
			if date, err := time.Parse(format, *req.PublishedAt); err == nil {
				parsedDate = &date
				break
			}
		}
		if parsedDate != nil {
			news.PublishedAt = parsedDate
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD HH:MM:SS atau YYYY-MM-DD",
			})
		}
	}

	if req.IsPublished != nil {
		news.IsPublished = req.IsPublished
	}

	if req.IsHeadline != nil {
		news.IsHeadline = req.IsHeadline
	}

	if req.Tags != nil {
		news.Tags = req.Tags
	}

	if err := db.Save(&news).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate news",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "News berhasil diupdate",
		"data":    news,
	})
}

// DeleteNewsHandler menghapus news (soft delete)
// @Summary      Delete news
// @Description  Soft delete a news article by ID
// @Tags         News
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "News ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /news/{id} [delete]
func DeleteNewsHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var news models.News
	if err := db.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "News tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data news",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&news).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus news",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "News berhasil dihapus",
	})
}
