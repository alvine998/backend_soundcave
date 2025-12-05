package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"backend_soundcave/config"
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// UploadImageHandler menangani upload gambar ke Firebase Storage
func UploadImageHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Parse multipart form
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File gambar tidak ditemukan",
			"error":   err.Error(),
		})
	}

	// Validasi file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file terlalu besar. Maksimal 10MB",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tipe file tidak diizinkan. Hanya gambar (JPEG, PNG, GIF, WEBP)",
		})
	}

	// Buka file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuka file",
			"error":   err.Error(),
		})
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	bucketPath := fmt.Sprintf("images/%s", filename)

	// Upload ke Firebase Storage
	ctx := context.Background()
	bucket, err := config.GetStorageBucket()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengakses Firebase Storage",
			"error":   err.Error(),
		})
	}

	// Buat object writer
	obj := bucket.Object(bucketPath)
	writer := obj.NewWriter(ctx)
	writer.ContentType = file.Header.Get("Content-Type")
	writer.CacheControl = "public, max-age=31536000"

	// Copy file ke Firebase Storage
	if _, err := io.Copy(writer, src); err != nil {
		writer.Close()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal upload file ke Firebase Storage",
			"error":   err.Error(),
		})
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menutup writer",
			"error":   err.Error(),
		})
	}

	// Set public access
	if err := obj.ACL().Set(ctx, "allUsers", "READER"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal set public access",
			"error":   err.Error(),
		})
	}

	// Get public URL
	// Format: https://firebasestorage.googleapis.com/v0/b/{bucket}/o/{encodedPath}?alt=media
	bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET")
	if bucketName == "" {
		attrs, err := obj.Attrs(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mendapatkan URL file",
				"error":   err.Error(),
			})
		}
		bucketName = attrs.Bucket
	}
	
	// Generate public URL
	fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media", 
		bucketName, 
		fmt.Sprintf("images%%2F%s", filename))

	// Simpan informasi ke database
	image := models.Image{
		FileName:    file.Filename,
		FileURL:     fileURL,
		FileSize:    file.Size,
		ContentType: file.Header.Get("Content-Type"),
		BucketPath:  bucketPath,
	}

	if err := db.Create(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan data ke database",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Gambar berhasil diupload",
		"data":    image,
	})
}

// UploadMultipleImagesHandler menangani upload multiple gambar
func UploadMultipleImagesHandler(c *fiber.Ctx, db *gorm.DB) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse form",
			"error":   err.Error(),
		})
	}

	files := form.File["images"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tidak ada file yang diupload",
		})
	}

	var uploadedImages []models.Image
	var errors []string

	ctx := context.Background()
	bucket, err := config.GetStorageBucket()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengakses Firebase Storage",
			"error":   err.Error(),
		})
	}

	// Validasi file types
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	maxSize := int64(10 * 1024 * 1024) // 10MB

	for _, file := range files {
		// Validasi size
		if file.Size > maxSize {
			errors = append(errors, fmt.Sprintf("%s: ukuran terlalu besar", file.Filename))
			continue
		}

		// Validasi type
		if !allowedTypes[file.Header.Get("Content-Type")] {
			errors = append(errors, fmt.Sprintf("%s: tipe file tidak diizinkan", file.Filename))
			continue
		}

		// Buka file
		src, err := file.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// Generate filename
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		bucketPath := fmt.Sprintf("images/%s", filename)

		// Upload ke Firebase
		obj := bucket.Object(bucketPath)
		writer := obj.NewWriter(ctx)
		writer.ContentType = file.Header.Get("Content-Type")
		writer.CacheControl = "public, max-age=31536000"

		if _, err := io.Copy(writer, src); err != nil {
			src.Close()
			writer.Close()
			errors = append(errors, fmt.Sprintf("%s: gagal upload", file.Filename))
			continue
		}

		src.Close()
		if err := writer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: gagal menutup writer", file.Filename))
			continue
		}

		// Set public access
		if err := obj.ACL().Set(ctx, "allUsers", "READER"); err != nil {
			errors = append(errors, fmt.Sprintf("%s: gagal set public access", file.Filename))
			continue
		}

		// Get bucket name untuk URL
		bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET")
		if bucketName == "" {
			attrs, err := obj.Attrs(ctx)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: gagal mendapatkan URL", file.Filename))
				continue
			}
			bucketName = attrs.Bucket
		}

		// Generate public URL
		fileURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media",
			bucketName,
			fmt.Sprintf("images%%2F%s", filename))

		// Simpan ke database
		image := models.Image{
			FileName:    file.Filename,
			FileURL:     fileURL,
			FileSize:    file.Size,
			ContentType: file.Header.Get("Content-Type"),
			BucketPath:  bucketPath,
		}

		if err := db.Create(&image).Error; err != nil {
			errors = append(errors, fmt.Sprintf("%s: gagal menyimpan ke database", file.Filename))
			continue
		}

		uploadedImages = append(uploadedImages, image)
	}

	response := fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%d dari %d gambar berhasil diupload", len(uploadedImages), len(files)),
		"data":    uploadedImages,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetImagesHandler mendapatkan semua gambar
func GetImagesHandler(c *fiber.Ctx, db *gorm.DB) error {
	var images []models.Image

	if err := db.Order("created_at DESC").Find(&images).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data gambar",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    images,
	})
}

// DeleteImageHandler menghapus gambar
func DeleteImageHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var image models.Image
	if err := db.First(&image, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Gambar tidak ditemukan",
		})
	}

	// Hapus dari Firebase Storage
	ctx := context.Background()
	bucket, err := config.GetStorageBucket()
	if err == nil {
		obj := bucket.Object(image.BucketPath)
		if err := obj.Delete(ctx); err != nil {
			// Log error tapi lanjutkan hapus dari database
			fmt.Printf("Gagal menghapus dari Firebase Storage: %v\n", err)
		}
	}

	// Hapus dari database
	if err := db.Delete(&image).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus dari database",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Gambar berhasil dihapus",
	})
}

