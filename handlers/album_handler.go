package handlers

import (
	"backend_soundcave/config"
	"backend_soundcave/models"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateAlbumRequest struct untuk request create album
type CreateAlbumRequest struct {
	Title       string  `json:"title" form:"title" validate:"required"`
	ArtistID    int     `json:"artist_id" form:"artist_id" validate:"required"`
	Artist      string  `json:"artist" form:"artist" validate:"required"`
	ReleaseDate string  `json:"release_date" form:"release_date" validate:"required"` // Format: "2006-01-02"
	AlbumType   string  `json:"album_type" form:"album_type" validate:"required,oneof=single EP album compilation"`
	Genre       string  `json:"genre" form:"genre"`
	TotalTracks int     `json:"total_tracks" form:"total_tracks" validate:"min=0"`
	RecordLabel *string `json:"record_label" form:"record_label"`
	Image       *string `json:"image" form:"image"`
}

// UpdateAlbumRequest struct untuk request update album
type UpdateAlbumRequest struct {
	Title       *string `json:"title" form:"title"`
	ArtistID    *int    `json:"artist_id" form:"artist_id"`
	Artist      *string `json:"artist" form:"artist"`
	ReleaseDate *string `json:"release_date" form:"release_date"` // Format: "2006-01-02"
	AlbumType   *string `json:"album_type" form:"album_type" validate:"omitempty,oneof=single EP album compilation"`
	Genre       *string `json:"genre" form:"genre"`
	TotalTracks *int    `json:"total_tracks" form:"total_tracks" validate:"omitempty,min=0"`
	RecordLabel *string `json:"record_label" form:"record_label"`
	Image       *string `json:"image" form:"image"`
}

// CreateAlbumHandler membuat album baru
// @Summary      Create new album
// @Description  Create a new album
// @Tags         Albums
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        title         formData  string  true   "Album Title"
// @Param        artist_id     formData  int     true   "Artist ID"
// @Param        artist        formData  string  true   "Artist Name"
// @Param        release_date  formData  string  true   "Release Date (YYYY-MM-DD)"
// @Param        album_type    formData  string  true   "Album Type (single, EP, album, compilation)"
// @Param        genre         formData  string  false  "Genre"
// @Param        total_tracks  formData  int     false  "Total Tracks"
// @Param        record_label  formData  string  false  "Record Label"
// @Param        image         formData  file    false  "Album cover image file"
// @Param        image_url     formData  string  false  "Album cover image URL (if not uploading file)"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /albums [post]
func CreateAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreateAlbumRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Handle image upload from multipart form
	file, err := c.FormFile("image")
	if err == nil {
		// Upload image to Firebase Storage
		imageURL, err := uploadAlbumImageToFirebase(c, file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal upload album cover",
				"error":   err.Error(),
			})
		}
		req.Image = &imageURL
	} else if req.Image == nil {
		// Try to get image_url from form if image file is not provided
		imageURL := c.FormValue("image_url")
		if imageURL != "" {
			req.Image = &imageURL
		}
	}

	// Parse release date
	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
			"error":   err.Error(),
		})
	}

	// Validasi album type
	var albumType models.AlbumType
	switch req.AlbumType {
	case "single":
		albumType = models.AlbumTypeSingle
	case "EP":
		albumType = models.AlbumTypeEP
	case "album":
		albumType = models.AlbumTypeAlbum
	case "compilation":
		albumType = models.AlbumTypeCompilation
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Album type tidak valid. Pilih: single, EP, album, atau compilation",
		})
	}

	// Buat album baru
	album := models.Album{
		Title:       req.Title,
		ArtistID:    req.ArtistID,
		Artist:      req.Artist,
		ReleaseDate: &releaseDate,
		AlbumType:   albumType,
		Genre:       req.Genre,
		TotalTracks: req.TotalTracks,
		RecordLabel: req.RecordLabel,
		Image:       req.Image,
	}

	if err := db.Create(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil dibuat",
		"data":    album,
	})
}

// GetAlbumsHandler mendapatkan semua albums dengan pagination
// @Summary      Get all albums
// @Description  Get paginated list of albums with filtering and search
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number" default(1)
// @Param        limit    query     int     false  "Items per page" default(10)
// @Param        artist   query     string  false  "Filter by artist"
// @Param        search   query     string  false  "Search by title or artist"
// @Param        sort_by  query     string  false  "Sort field" default(created_at)
// @Param        order    query     string  false  "Sort order" default(desc)
// @Success      200      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /albums [get]
func GetAlbumsHandler(c *fiber.Ctx, db *gorm.DB) error {
	var albums []models.Album

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Query dengan pagination
	query := db.Model(&models.Album{})

	// Filter by artist_id jika ada
	if artistID := c.QueryInt("artist_id", 0); artistID > 0 {
		query = query.Where("artist_id = ?", artistID)
	}

	// Filter by album_type jika ada
	if albumType := c.Query("album_type"); albumType != "" {
		query = query.Where("album_type = ?", albumType)
	}

	// Filter by genre jika ada
	if genre := c.Query("genre"); genre != "" {
		query = query.Where("genre LIKE ?", "%"+genre+"%")
	}

	// Search by title atau artist
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR artist LIKE ?", "%"+search+"%", "%"+search+"%")
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

	// Get albums
	if err := query.Offset(offset).Limit(limit).Find(&albums).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data albums",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    albums,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetAlbumHandler mendapatkan album by ID
// @Summary      Get album by ID
// @Description  Get album details by ID
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Album ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /albums/{id} [get]
func GetAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    album,
	})
}

// UpdateAlbumHandler mengupdate album
// @Summary      Update album
// @Description  Update album information
// @Tags         Albums
// @Accept       json
// @Accept       multipart/form-data
// @Produce      json
// @Param        id            path      int     true   "Album ID"
// @Param        title         formData  string  false  "Album Title"
// @Param        artist_id     formData  int     false  "Artist ID"
// @Param        artist        formData  string  false  "Artist Name"
// @Param        release_date  formData  string  false  "Release Date (YYYY-MM-DD)"
// @Param        album_type    formData  string  false  "Album Type (single, EP, album, compilation)"
// @Param        genre         formData  string  false  "Genre"
// @Param        total_tracks  formData  int     false  "Total Tracks"
// @Param        record_label  formData  string  false  "Record Label"
// @Param        image         formData  file    false  "Album cover image file"
// @Param        image_url     formData  string  false  "Album cover image URL (if not uploading file)"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /albums/{id} [put]
func UpdateAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	var req UpdateAlbumRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Handle image upload from multipart form
	file, err := c.FormFile("image")
	if err == nil {
		// Upload image to Firebase Storage
		imageURL, err := uploadAlbumImageToFirebase(c, file)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal upload album cover",
				"error":   err.Error(),
			})
		}
		req.Image = &imageURL
	} else {
		// Try to get image_url from form if image file is not provided
		imageURL := c.FormValue("image_url")
		if imageURL != "" {
			req.Image = &imageURL
		}
	}

	// Update fields jika ada
	if req.Title != nil {
		album.Title = *req.Title
	}

	if req.ArtistID != nil {
		album.ArtistID = *req.ArtistID
	}

	if req.Artist != nil {
		album.Artist = *req.Artist
	}

	if req.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *req.ReleaseDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Format tanggal tidak valid. Gunakan format: YYYY-MM-DD",
				"error":   err.Error(),
			})
		}
		album.ReleaseDate = &releaseDate
	}

	if req.AlbumType != nil {
		switch *req.AlbumType {
		case "single":
			album.AlbumType = models.AlbumTypeSingle
		case "EP":
			album.AlbumType = models.AlbumTypeEP
		case "album":
			album.AlbumType = models.AlbumTypeAlbum
		case "compilation":
			album.AlbumType = models.AlbumTypeCompilation
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Album type tidak valid. Pilih: single, EP, album, atau compilation",
			})
		}
	}

	if req.Genre != nil {
		album.Genre = *req.Genre
	}

	if req.TotalTracks != nil {
		album.TotalTracks = *req.TotalTracks
	}

	if req.RecordLabel != nil {
		album.RecordLabel = req.RecordLabel
	}

	if req.Image != nil {
		album.Image = req.Image
	}

	if err := db.Save(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil diupdate",
		"data":    album,
	})
}

// DeleteAlbumHandler menghapus album (soft delete)
// @Summary      Delete album
// @Description  Soft delete an album by ID
// @Tags         Albums
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Album ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /albums/{id} [delete]
func DeleteAlbumHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var album models.Album
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Album tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data album",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&album).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus album",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Album berhasil dihapus",
	})
}

// uploadAlbumImageToFirebase helper function to upload album cover to Firebase
func uploadAlbumImageToFirebase(c *fiber.Ctx, file *multipart.FileHeader) (string, error) {
	// Validasi file size (max 5GB, but here 10MB is enough for cover)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return "", fmt.Errorf("ukuran file terlalu besar (maksimal 10MB)")
	}

	// Validasi file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return "", fmt.Errorf("tipe file tidak diizinkan. Hanya gambar (JPEG, PNG, GIF, WEBP)")
	}

	// Buka file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	bucketPath := fmt.Sprintf("albums/%s", filename)

	// Upload ke Firebase Storage
	ctx := context.Background()
	bucket, err := config.GetStorageBucket()
	if err != nil {
		return "", err
	}
	if bucket == nil {
		return "", fmt.Errorf("firebase storage bucket tidak tersedia")
	}

	// Buat object writer
	obj := bucket.Object(bucketPath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = file.Header.Get("Content-Type")
	writer.CacheControl = "public, max-age=31536000"

	// Copy file ke Firebase Storage
	if _, err := io.Copy(writer, src); err != nil {
		writer.Close()
		return "", err
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return "", err
	}

	// Set public access
	if err := obj.ACL().Set(ctx, "allUsers", "READER"); err != nil {
		return "", err
	}

	// Get public URL
	bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET")
	if bucketName == "" {
		attrs, err := obj.Attrs(ctx)
		if err != nil {
			return "", err
		}
		bucketName = attrs.Bucket
	}

	imageURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media",
		bucketName, url.QueryEscape(bucketPath))

	return imageURL, nil
}
