package handlers

import (
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetDashboardStatsHandler mendapatkan statistik dashboard
// @Summary      Get dashboard statistics
// @Description  Get dashboard statistics including counts and metrics
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /dashboard/stats [get]
func GetDashboardStatsHandler(c *fiber.Ctx, db *gorm.DB) error {
	stats := make(map[string]interface{})

	// Total counts
	var totalUsers int64
	var totalAlbums int64
	var totalMusics int64
	var totalArtists int64
	var totalPlaylists int64
	var totalGenres int64
	var totalPodcasts int64
	var totalMusicVideos int64
	var totalNotifications int64
	var totalSubscriptionPlans int64
	var totalImages int64
	var totalPlaylistSongs int64

	// Count semua model
	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.Album{}).Count(&totalAlbums)
	db.Model(&models.Music{}).Count(&totalMusics)
	db.Model(&models.Artist{}).Count(&totalArtists)
	db.Model(&models.Playlist{}).Count(&totalPlaylists)
	db.Model(&models.Genre{}).Count(&totalGenres)
	db.Model(&models.Podcast{}).Count(&totalPodcasts)
	db.Model(&models.MusicVideo{}).Count(&totalMusicVideos)
	db.Model(&models.Notification{}).Count(&totalNotifications)
	db.Model(&models.SubscriptionPlan{}).Count(&totalSubscriptionPlans)
	db.Model(&models.Image{}).Count(&totalImages)
	db.Model(&models.PlaylistSong{}).Count(&totalPlaylistSongs)

	// Music statistics - get total play count and like count
	var totalPlayCount int64
	var totalLikeCount int64

	// Use Raw SQL for SUM since play_count and like_count are nullable
	db.Raw("SELECT COALESCE(SUM(play_count), 0) FROM musics WHERE deleted_at IS NULL").Scan(&totalPlayCount)
	db.Raw("SELECT COALESCE(SUM(like_count), 0) FROM musics WHERE deleted_at IS NULL").Scan(&totalLikeCount)

	// User statistics by role
	var totalAdminUsers int64
	var totalPremiumUsers int64
	var totalRegularUsers int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&totalAdminUsers)
	db.Model(&models.User{}).Where("role = ?", "premium").Count(&totalPremiumUsers)
	db.Model(&models.User{}).Where("role = ?", "user").Count(&totalRegularUsers)

	// Album statistics by type
	var totalSingles int64
	var totalEPs int64
	var totalFullAlbums int64
	var totalCompilations int64
	db.Model(&models.Album{}).Where("album_type = ?", "single").Count(&totalSingles)
	db.Model(&models.Album{}).Where("album_type = ?", "EP").Count(&totalEPs)
	db.Model(&models.Album{}).Where("album_type = ?", "album").Count(&totalFullAlbums)
	db.Model(&models.Album{}).Where("album_type = ?", "compilation").Count(&totalCompilations)

	// Playlist statistics
	var totalPublicPlaylists int64
	var totalPrivatePlaylists int64
	db.Model(&models.Playlist{}).Where("is_public = ?", true).Count(&totalPublicPlaylists)
	db.Model(&models.Playlist{}).Where("is_public = ?", false).Count(&totalPrivatePlaylists)

	// Notification statistics
	var totalUnreadNotifications int64
	db.Model(&models.Notification{}).Where("is_read = ?", false).Count(&totalUnreadNotifications)

	// Recent activity (last 7 days)
	var recentUsers int64
	var recentMusics int64
	var recentAlbums int64
	db.Model(&models.User{}).Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&recentUsers)
	db.Model(&models.Music{}).Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&recentMusics)
	db.Model(&models.Album{}).Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Count(&recentAlbums)

	// Build stats response
	stats["totals"] = fiber.Map{
		"users":              totalUsers,
		"albums":             totalAlbums,
		"musics":             totalMusics,
		"artists":            totalArtists,
		"playlists":          totalPlaylists,
		"genres":             totalGenres,
		"podcasts":           totalPodcasts,
		"music_videos":       totalMusicVideos,
		"notifications":      totalNotifications,
		"subscription_plans": totalSubscriptionPlans,
		"images":             totalImages,
		"playlist_songs":     totalPlaylistSongs,
	}

	stats["music_stats"] = fiber.Map{
		"total_play_count": totalPlayCount,
		"total_like_count": totalLikeCount,
	}

	stats["user_stats"] = fiber.Map{
		"total":   totalUsers,
		"admin":   totalAdminUsers,
		"premium": totalPremiumUsers,
		"regular": totalRegularUsers,
	}

	stats["album_stats"] = fiber.Map{
		"total":        totalAlbums,
		"singles":      totalSingles,
		"eps":          totalEPs,
		"albums":       totalFullAlbums,
		"compilations": totalCompilations,
	}

	stats["playlist_stats"] = fiber.Map{
		"total":   totalPlaylists,
		"public":  totalPublicPlaylists,
		"private": totalPrivatePlaylists,
	}

	stats["notification_stats"] = fiber.Map{
		"total":  totalNotifications,
		"unread": totalUnreadNotifications,
		"read":   totalNotifications - totalUnreadNotifications,
	}

	stats["recent_activity"] = fiber.Map{
		"last_7_days": fiber.Map{
			"new_users":  recentUsers,
			"new_musics": recentMusics,
			"new_albums": recentAlbums,
		},
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}
