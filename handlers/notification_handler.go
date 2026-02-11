package handlers

import (
	"backend_soundcave/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateNotificationRequest struct untuk request create notification
type CreateNotificationRequest struct {
	UserID  int    `json:"user_id" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
	Date    string `json:"date" validate:"required"` // Format: "2006-01-02"
	IsRead  *bool  `json:"is_read"`
	Type    string `json:"type" validate:"omitempty,oneof=info success warning error"`
}

// UpdateNotificationRequest struct untuk request update notification
type UpdateNotificationRequest struct {
	UserID  *int    `json:"user_id"`
	Title   *string `json:"title"`
	Message *string `json:"message"`
	Date    *string `json:"date"` // Format: "2006-01-02"
	IsRead  *bool   `json:"is_read"`
	Type    *string `json:"type" validate:"omitempty,oneof=info success warning error"`
}

// CreateNotificationHandler membuat notification baru
// @Summary      Create new notification
// @Description  Create a new notification
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        request  body      CreateNotificationRequest  true  "Notification Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /notifications [post]
func CreateNotificationHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateNotificationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
			"error":   err.Error(),
		})
	}

	// Set default values
	isRead := false
	notificationType := models.NotificationTypeInfo

	if req.IsRead != nil {
		isRead = *req.IsRead
	}

	if req.Type != "" {
		switch req.Type {
		case "info":
			notificationType = models.NotificationTypeInfo
		case "success":
			notificationType = models.NotificationTypeSuccess
		case "warning":
			notificationType = models.NotificationTypeWarning
		case "error":
			notificationType = models.NotificationTypeError
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Type tidak valid. Pilih: info, success, warning, atau error",
			})
		}
	}

	// Buat notification baru
	notification := models.Notification{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Date:    date,
		IsRead:  &isRead,
		Type:    notificationType,
	}

	if err := db.Create(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat notification",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Notification berhasil dibuat",
		"data":    notification,
	})
}

// GetNotificationsHandler mendapatkan semua notifications dengan pagination
// @Summary      Get all notifications
// @Description  Get paginated list of notifications with filtering
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        type     query     string  false  "Filter by type"
// @Param        search   query     string  false  "Search by title or message"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /notifications [get]
func GetNotificationsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var notifications []models.Notification

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Notification{})

	// Filter by user_id jika ada
	if userID := c.QueryInt("user_id", 0); userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by is_read jika ada
	if isRead := c.Query("is_read"); isRead != "" {
		isReadBool, err := strconv.ParseBool(isRead)
		if err == nil {
			query = query.Where("is_read = ?", isReadBool)
		}
	}

	// Filter by type jika ada
	if notificationType := c.Query("type"); notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	// Search by title atau message
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR message LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get notifications
	if err := query.Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notifications",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    notifications,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetNotificationHandler mendapatkan notification by ID
// @Summary      Get notification by ID
// @Description  Get notification details by ID
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /notifications/{id} [get]
func GetNotificationHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var notification models.Notification
	if err := db.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Notification tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notification",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    notification,
	})
}

// GetUserNotificationsHandler mendapatkan notifications by user_id
// @Summary      Get user notifications
// @Description  Get paginated list of notifications for a specific user
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        user_id  path      int     true  "User ID"
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        is_read  query     bool    false  "Filter by read status"
// @Param        type     query     string  false  "Filter by type"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /notifications/user/{user_id} [get]
func GetUserNotificationsHandler(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Params("user_id")

	var notifications []models.Notification

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Notification{}).Where("user_id = ?", userID)

	// Filter by is_read jika ada
	if isRead := c.Query("is_read"); isRead != "" {
		isReadBool, err := strconv.ParseBool(isRead)
		if err == nil {
			query = query.Where("is_read = ?", isReadBool)
		}
	}

	// Filter by type jika ada
	if notificationType := c.Query("type"); notificationType != "" {
		query = query.Where("type = ?", notificationType)
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

	// Get notifications
	if err := query.Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notifications",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    notifications,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// UpdateNotificationHandler mengupdate notification
// @Summary      Update notification
// @Description  Update notification information
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Notification ID"
// @Param        request  body      UpdateNotificationRequest  true  "Update Notification Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /notifications/{id} [put]
func UpdateNotificationHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var notification models.Notification
	if err := db.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Notification tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notification",
			"error":   err.Error(),
		})
	}

	var req UpdateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.UserID != nil {
		notification.UserID = *req.UserID
	}

	if req.Title != nil {
		notification.Title = *req.Title
	}

	if req.Message != nil {
		notification.Message = *req.Message
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		notification.Date = date
	}

	if req.IsRead != nil {
		notification.IsRead = req.IsRead
	}

	if req.Type != nil {
		switch *req.Type {
		case "info":
			notification.Type = models.NotificationTypeInfo
		case "success":
			notification.Type = models.NotificationTypeSuccess
		case "warning":
			notification.Type = models.NotificationTypeWarning
		case "error":
			notification.Type = models.NotificationTypeError
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Type tidak valid. Pilih: info, success, warning, atau error",
			})
		}
	}

	if err := db.Save(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate notification",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Notification berhasil diupdate",
		"data":    notification,
	})
}

// MarkAsReadHandler menandai notification sebagai sudah dibaca
// @Summary      Mark notification as read
// @Description  Mark a notification as read by ID
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /notifications/{id}/read [post]
func MarkAsReadHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var notification models.Notification
	if err := db.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Notification tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notification",
			"error":   err.Error(),
		})
	}

	isRead := true
	notification.IsRead = &isRead

	if err := db.Save(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate notification",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Notification berhasil ditandai sebagai sudah dibaca",
		"data":    notification,
	})
}

// MarkAllAsReadHandler menandai semua notifications user sebagai sudah dibaca
// @Summary      Mark all notifications as read
// @Description  Mark all notifications for a specific user as read
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        user_id  path      int  true  "User ID"
// @Success      200      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /notifications/user/{user_id}/read-all [post]
func MarkAllAsReadHandler(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Params("user_id")

	result := db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate notifications",
			"error":   result.Error.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Semua notifications berhasil ditandai sebagai sudah dibaca",
		"count":   result.RowsAffected,
	})
}

// DeleteNotificationHandler menghapus notification (soft delete)
// @Summary      Delete notification
// @Description  Soft delete a notification by ID
// @Tags         Notifications
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Notification ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /notifications/{id} [delete]
func DeleteNotificationHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var notification models.Notification
	if err := db.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Notification tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data notification",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&notification).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus notification",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Notification berhasil dihapus",
	})
}
