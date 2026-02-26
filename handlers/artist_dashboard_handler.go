package handlers

import (
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetArtistDashboardStatsHandler mendapatkan statistik dashboard untuk user independent/label
// @Summary      Get artist dashboard statistics
// @Description  Get dashboard statistics for independent artist or label users including album, song, and music video counts
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /dashboard/artist-stats [get]
func GetArtistDashboardStatsHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Get user info from JWT context (set by AuthMiddleware)
	userID := c.Locals("user_id").(uint)
	role := c.Locals("role").(string)

	// Validate role - only independent or label allowed
	if role != string(models.RoleIndependent) && role != string(models.RoleLabel) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akses ditolak. Hanya user dengan role independent atau label yang dapat mengakses",
		})
	}

	artistID := int(userID)

	// --- Album stats ---
	var totalAlbums int64
	var totalSingles int64
	var totalEPs int64
	var totalFullAlbums int64
	var totalCompilations int64

	db.Model(&models.Album{}).Where("artist_id = ?", artistID).Count(&totalAlbums)
	db.Model(&models.Album{}).Where("artist_id = ? AND album_type = ?", artistID, "single").Count(&totalSingles)
	db.Model(&models.Album{}).Where("artist_id = ? AND album_type = ?", artistID, "EP").Count(&totalEPs)
	db.Model(&models.Album{}).Where("artist_id = ? AND album_type = ?", artistID, "album").Count(&totalFullAlbums)
	db.Model(&models.Album{}).Where("artist_id = ? AND album_type = ?", artistID, "compilation").Count(&totalCompilations)

	// --- Song (Music) stats ---
	var totalSongs int64
	var totalPlayCount int64
	var totalLikeCount int64

	db.Model(&models.Music{}).Where("artist_id = ?", artistID).Count(&totalSongs)
	db.Raw("SELECT COALESCE(SUM(play_count), 0) FROM musics WHERE artist_id = ? AND deleted_at IS NULL", artistID).Scan(&totalPlayCount)
	db.Raw("SELECT COALESCE(SUM(like_count), 0) FROM musics WHERE artist_id = ? AND deleted_at IS NULL", artistID).Scan(&totalLikeCount)

	// --- Music Video stats ---
	var totalMusicVideos int64
	db.Model(&models.MusicVideo{}).Where("artist_id = ?", artistID).Count(&totalMusicVideos)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id": userID,
			"role":    role,
			"albums": fiber.Map{
				"total":        totalAlbums,
				"singles":      totalSingles,
				"eps":          totalEPs,
				"albums":       totalFullAlbums,
				"compilations": totalCompilations,
			},
			"songs": fiber.Map{
				"total":            totalSongs,
				"total_play_count": totalPlayCount,
				"total_like_count": totalLikeCount,
			},
			"music_videos": fiber.Map{
				"total": totalMusicVideos,
			},
		},
	})
}
