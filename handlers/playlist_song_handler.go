package handlers

import (
	"backend_soundcave/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreatePlaylistSongRequest struct untuk request create playlist song
type CreatePlaylistSongRequest struct {
	PlaylistID uint `json:"playlist_id" validate:"required"`
	MusicID    uint `json:"music_id" validate:"required"`
	Position   *int `json:"position"` // Optional, jika tidak di-set akan ditambahkan di akhir
}

// UpdatePlaylistSongRequest struct untuk request update playlist song
type UpdatePlaylistSongRequest struct {
	Position *int `json:"position"`
}

// CreatePlaylistSongHandler menambahkan lagu ke playlist
// @Summary      Add song to playlist
// @Description  Add a music track to a playlist
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        request  body      CreatePlaylistSongRequest  true  "Playlist Song Request"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs [post]
func CreatePlaylistSongHandler(c *fiber.Ctx, db *gorm.DB) error {
	var req CreatePlaylistSongRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Validasi playlist exists
	var playlist models.Playlist
	if err := db.First(&playlist, req.PlaylistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist",
			"error":   err.Error(),
		})
	}

	// Validasi music exists
	var music models.Music
	if err := db.First(&music, req.MusicID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Music tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data music",
			"error":   err.Error(),
		})
	}

	// Cek apakah lagu sudah ada di playlist
	var existingPlaylistSong models.PlaylistSong
	if err := db.Where("playlist_id = ? AND music_id = ?", req.PlaylistID, req.MusicID).First(&existingPlaylistSong).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Lagu sudah ada di playlist ini",
		})
	}

	// Tentukan position
	position := 0
	if req.Position != nil {
		position = *req.Position
	} else {
		// Jika tidak di-set, ambil position terakhir + 1
		var count int64
		db.Model(&models.PlaylistSong{}).
			Where("playlist_id = ?", req.PlaylistID).
			Count(&count)

		if count > 0 {
			var lastPlaylistSong models.PlaylistSong
			db.Where("playlist_id = ?", req.PlaylistID).
				Order("position DESC").
				First(&lastPlaylistSong)
			position = lastPlaylistSong.Position + 1
		} else {
			position = 0
		}
	}

	// Buat playlist song baru
	playlistSong := models.PlaylistSong{
		PlaylistID: req.PlaylistID,
		MusicID:    req.MusicID,
		Position:   position,
	}

	if err := db.Create(&playlistSong).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menambahkan lagu ke playlist",
			"error":   err.Error(),
		})
	}

	// Load relations untuk response
	db.Preload("Playlist").Preload("Music").First(&playlistSong, playlistSong.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Lagu berhasil ditambahkan ke playlist",
		"data":    playlistSong,
	})
}

// GetPlaylistSongsHandler mendapatkan semua lagu dalam playlist
// @Summary      Get playlist songs
// @Description  Get all songs in a playlist
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        playlist_id  path      int  true  "Playlist ID"
// @Success      200          {object}  map[string]interface{}
// @Failure      401          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs/playlist/{playlist_id} [get]
func GetPlaylistSongsHandler(c *fiber.Ctx, db *gorm.DB) error {
	playlistID := c.Params("playlist_id")

	// Validasi playlist exists
	var playlist models.Playlist
	if err := db.First(&playlist, playlistID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist",
			"error":   err.Error(),
		})
	}

	var playlistSongs []models.PlaylistSong
	if err := db.Where("playlist_id = ?", playlistID).
		Preload("Music").
		Order("position ASC").
		Find(&playlistSongs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist songs",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    playlistSongs,
	})
}

// GetPlaylistSongHandler mendapatkan playlist song by ID
// @Summary      Get playlist song by ID
// @Description  Get playlist song details by ID
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Playlist Song ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs/{id} [get]
func GetPlaylistSongHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlistSong models.PlaylistSong
	if err := db.Preload("Playlist").Preload("Music").First(&playlistSong, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist song tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist song",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    playlistSong,
	})
}

// UpdatePlaylistSongHandler mengupdate playlist song (biasanya untuk mengubah position)
// @Summary      Update playlist song
// @Description  Update playlist song position or information
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        id       path      int                    true  "Playlist Song ID"
// @Param        request  body      UpdatePlaylistSongRequest  true  "Update Playlist Song Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      401      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs/{id} [put]
func UpdatePlaylistSongHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlistSong models.PlaylistSong
	if err := db.First(&playlistSong, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist song tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist song",
			"error":   err.Error(),
		})
	}

	var req UpdatePlaylistSongRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	// Update position jika ada
	if req.Position != nil {
		playlistSong.Position = *req.Position
	}

	if err := db.Save(&playlistSong).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate playlist song",
			"error":   err.Error(),
		})
	}

	// Load relations untuk response
	db.Preload("Playlist").Preload("Music").First(&playlistSong, playlistSong.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Playlist song berhasil diupdate",
		"data":    playlistSong,
	})
}

// DeletePlaylistSongHandler menghapus lagu dari playlist
// DeletePlaylistSongHandler menghapus playlist song (soft delete)
// @Summary      Delete playlist song
// @Description  Remove a song from playlist by playlist song ID
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Playlist Song ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs/{id} [delete]
func DeletePlaylistSongHandler(c *fiber.Ctx, db *gorm.DB) error {
	id := c.Params("id")

	var playlistSong models.PlaylistSong
	if err := db.First(&playlistSong, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Playlist song tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist song",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&playlistSong).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus lagu dari playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Lagu berhasil dihapus dari playlist",
	})
}

// DeletePlaylistSongByMusicHandler menghapus lagu dari playlist berdasarkan playlist_id dan music_id
// @Summary      Delete playlist song by playlist and music
// @Description  Remove a song from playlist by playlist ID and music ID
// @Tags         PlaylistSongs
// @Accept       json
// @Produce      json
// @Param        playlist_id  path      int  true  "Playlist ID"
// @Param        music_id     path      int  true  "Music ID"
// @Success      200          {object}  map[string]interface{}
// @Failure      401          {object}  map[string]interface{}
// @Failure      404          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /playlist-songs/playlist/{playlist_id}/music/{music_id} [delete]
func DeletePlaylistSongByMusicHandler(c *fiber.Ctx, db *gorm.DB) error {
	playlistID := c.Params("playlist_id")
	musicID := c.Params("music_id")

	var playlistSong models.PlaylistSong
	if err := db.Where("playlist_id = ? AND music_id = ?", playlistID, musicID).First(&playlistSong).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Lagu tidak ditemukan di playlist ini",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data playlist song",
			"error":   err.Error(),
		})
	}

	if err := db.Delete(&playlistSong).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus lagu dari playlist",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Lagu berhasil dihapus dari playlist",
	})
}
