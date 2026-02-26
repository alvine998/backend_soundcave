package handlers

import (
	"backend_soundcave/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateArtistRequest struct untuk request create artist
type CreateArtistRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Bio          string                 `json:"bio" validate:"required"`
	Genre        string                 `json:"genre"`
	Country      string                 `json:"country"`
	DebutYear    string                 `json:"debut_year" validate:"required,len=4"`
	Website      *string                `json:"website"`
	Email        string                 `json:"email" validate:"required,email"`
	Phone        *string                `json:"phone"`
	SocialMedia  map[string]interface{} `json:"social_media"`
	ProfileImage *string                `json:"profile_image"`
}

// UpdateArtistRequest struct untuk request update artist
type UpdateArtistRequest struct {
	Name         *string                `json:"name"`
	Bio          *string                `json:"bio"`
	Genre        *string                `json:"genre"`
	Country      *string                `json:"country"`
	DebutYear    *string                `json:"debut_year" validate:"omitempty,len=4"`
	Website      *string                `json:"website"`
	Email        *string                `json:"email" validate:"omitempty,email"`
	Phone        *string                `json:"phone"`
	SocialMedia  map[string]interface{} `json:"social_media"`
	ProfileImage *string                `json:"profile_image"`
}

// CreateArtistHandler membuat artist baru
// @Summary      Create new artist
// @Description  Create a new artist
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        request  body      CreateArtistRequest  true  "Artist Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists [post]
func CreateArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateArtistRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi email sudah ada
	var existingArtist models.Artist
	if err := db.Where("email = ?", req.Email).First(&existingArtist).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Email sudah terdaftar",
		})
	}

	// Convert social media map to JSONB
	var socialMedia models.JSONB
	if req.SocialMedia != nil {
		socialMedia = models.JSONB(req.SocialMedia)
	}

	// Buat artist baru
	artist := models.Artist{
		Name:         req.Name,
		Bio:          req.Bio,
		Genre:        req.Genre,
		Country:      req.Country,
		DebutYear:    req.DebutYear,
		Website:      req.Website,
		Email:        req.Email,
		Phone:        req.Phone,
		SocialMedia:  socialMedia,
		ProfileImage: req.ProfileImage,
	}

	if err := db.Create(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil dibuat",
		"data":    artist,
	})
}

// GetArtistsHandler mendapatkan semua artists dengan pagination
// @Summary      Get all artists
// @Description  Get paginated list of artists with filtering and search
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        search   query     string  false  "Search by name"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists [get]
func GetArtistsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var artists []models.Artist

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Artist{})

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Filter by country jika ada
	if country := c.Query("country"); country != "" {
		query = query.Where("country = ?", country)
	}

	// Filter by debut year jika ada
	if debutYear := c.Query("debut_year"); debutYear != "" {
		query = query.Where("debut_year = ?", debutYear)
	}

	// Search by name atau email
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get artists
	if err := query.Offset(offset).Limit(limit).Find(&artists).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artists",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    artists,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetArtistHandler mendapatkan artist by ID
// @Summary      Get artist by ID
// @Description  Get artist details by ID
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Artist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/{id} [get]
func GetArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    artist,
	})
}

// UpdateArtistHandler mengupdate artist
// @Summary      Update artist
// @Description  Update artist information
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "Artist ID"
// @Param        request  body      UpdateArtistRequest  true  "Update Artist Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/{id} [put]
func UpdateArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	var req UpdateArtistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update fields jika ada
	if req.Name != nil {
		artist.Name = *req.Name
	}

	if req.Bio != nil {
		artist.Bio = *req.Bio
	}

	if req.Genre != nil {
		artist.Genre = *req.Genre
	}

	if req.Country != nil {
		artist.Country = *req.Country
	}

	if req.DebutYear != nil {
		artist.DebutYear = *req.DebutYear
	}

	if req.Website != nil {
		artist.Website = req.Website
	}

	if req.Email != nil {
		// Cek email sudah digunakan oleh artist lain
		var existingArtist models.Artist
		if err := db.Where("email = ? AND id != ?", *req.Email, id).First(&existingArtist).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Email sudah digunakan",
			})
		}
		artist.Email = *req.Email
	}

	if req.Phone != nil {
		artist.Phone = req.Phone
	}

	if req.SocialMedia != nil {
		artist.SocialMedia = models.JSONB(req.SocialMedia)
	}

	if req.ProfileImage != nil {
		artist.ProfileImage = req.ProfileImage
	}

	if err := db.Save(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil diupdate",
		"data":    artist,
	})
}

// DeleteArtistHandler menghapus artist (soft delete)
// @Summary      Delete artist
// @Description  Soft delete an artist by ID
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Artist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/{id} [delete]
func DeleteArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var artist models.Artist
	if err := db.First(&artist, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Artist berhasil dihapus",
	})
}

// GetRandomArtistsHandler mendapatkan list random artists
// @Summary      Get random artists
// @Description  Get a list of random artists
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        limit    query     int     false  "Number of artists to return" default(10)
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/random [get]
func GetRandomArtistsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var artists []models.Artist

	// Get limit dari query parameter, default 10
	limit := c.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit untuk menghindari query yang terlalu besar
	}

	// Query random artists menggunakan ORDER BY RAND()
	// Untuk MySQL/MariaDB, RAND() akan menghasilkan random order
	if err := db.Order("RAND()").Limit(limit).Find(&artists).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data random artists",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    artists,
		"count":   len(artists),
	})
}

// FollowArtistHandler menangani follow artist oleh user
// @Summary      Follow artist
// @Description  Follow an artist and increment their follower count
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Artist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      409  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/{id}/follow [post]
func FollowArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	artistID := c.Params("id")
	currentUserID := c.Locals("user_id").(uint)

	var artist models.Artist
	if err := db.First(&artist, artistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	// Check if already following
	currentUserIDStr := fmt.Sprintf("%d", currentUserID)
	if artist.Followers != nil {
		for _, followerID := range artist.Followers {
			if followerID == currentUserIDStr {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Anda sudah follow artist ini",
				})
			}
		}
	}

	// Add follower
	if artist.Followers == nil {
		artist.Followers = models.JSONStringArray{}
	}
	artist.Followers = append(artist.Followers, currentUserIDStr)
	artist.TotalFollower = len(artist.Followers)

	if err := db.Save(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal follow artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil follow artist",
		"data": fiber.Map{
			"artist_id":      artist.ID,
			"total_follower": artist.TotalFollower,
		},
	})
}

// UnfollowArtistHandler menangani unfollow artist oleh user
// @Summary      Unfollow artist
// @Description  Unfollow an artist and decrement their follower count
// @Tags         Artists
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Artist ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /artists/{id}/unfollow [post]
func UnfollowArtistHandler(c *fiber.Ctx, db *gorm.DB) error {
	artistID := c.Params("id")
	currentUserID := c.Locals("user_id").(uint)

	var artist models.Artist
	if err := db.First(&artist, artistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Artist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data artist",
			"error":   err.Error(),
		})
	}

	// Remove follower
	currentUserIDStr := fmt.Sprintf("%d", currentUserID)
	found := false
	var newFollowers models.JSONStringArray
	for _, followerID := range artist.Followers {
		if followerID == currentUserIDStr {
			found = true
			continue
		}
		newFollowers = append(newFollowers, followerID)
	}

	if !found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Anda belum follow artist ini",
		})
	}

	artist.Followers = newFollowers
	artist.TotalFollower = len(newFollowers)

	if err := db.Save(&artist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal unfollow artist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil unfollow artist",
		"data": fiber.Map{
			"artist_id":      artist.ID,
			"total_follower": artist.TotalFollower,
		},
	})
}
