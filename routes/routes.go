package routes

import (
	"backend_soundcave/handlers"
	"backend_soundcave/middleware"

	firebase "firebase.google.com/go/v4"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRoutes mengatur semua routes aplikasi
func SetupRoutes(app *fiber.App, db *gorm.DB, firebaseApp *firebase.App) {
	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "SoundCave Backend API",
			"version": "1.0.0",
		})
	})

	// API routes
	api := app.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", func(c *fiber.Ctx) error {
		return handlers.RegisterHandler(c, db)
	})
	auth.Post("/login", func(c *fiber.Ctx) error {
		return handlers.LoginHandler(c, db)
	})
	auth.Post("/google", func(c *fiber.Ctx) error {
		return handlers.GoogleAuthHandler(c, db)
	})

	// Protected routes (require authentication)
	protected := api.Group("", middleware.AuthMiddleware)
	protected.Get("/profile", func(c *fiber.Ctx) error {
		return handlers.GetProfileHandler(c, db)
	})
	protected.Get("/dashboard/stats", func(c *fiber.Ctx) error {
		return handlers.GetDashboardStatsHandler(c, db)
	})
	protected.Get("/dashboard/customer-report", func(c *fiber.Ctx) error {
		return handlers.GetCustomerReportHandler(c, db)
	})

	// Image upload routes
	images := api.Group("/images")
	images.Post("/upload", func(c *fiber.Ctx) error {
		return handlers.UploadImageHandler(c, db)
	})
	images.Post("/upload-multiple", func(c *fiber.Ctx) error {
		return handlers.UploadMultipleImagesHandler(c, db)
	})
	images.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetImagesHandler(c, db)
	})
	images.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteImageHandler(c, db)
	})

	// User CRUD routes
	users := api.Group("/users")
	users.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateUserHandler(c, db)
	})
	users.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetUsersHandler(c, db)
	})
	users.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetUserHandler(c, db)
	})
	users.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateUserHandler(c, db)
	})
	users.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteUserHandler(c, db)
	})

	// Album CRUD routes
	albums := api.Group("/albums")
	albums.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateAlbumHandler(c, db)
	})
	albums.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetAlbumsHandler(c, db)
	})
	albums.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetAlbumHandler(c, db)
	})
	albums.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateAlbumHandler(c, db)
	})
	albums.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteAlbumHandler(c, db)
	})

	// App Info CRUD routes
	appInfo := api.Group("/app-info")
	appInfo.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateAppInfoHandler(c, db)
	})
	appInfo.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetAppInfosHandler(c, db)
	})
	appInfo.Get("/latest", func(c *fiber.Ctx) error {
		return handlers.GetLatestAppInfoHandler(c, db)
	})
	appInfo.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetAppInfoHandler(c, db)
	})
	appInfo.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateAppInfoHandler(c, db)
	})
	appInfo.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteAppInfoHandler(c, db)
	})

	// Artist CRUD routes
	artists := api.Group("/artists")
	artists.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateArtistHandler(c, db)
	})
	artists.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetArtistsHandler(c, db)
	})
	artists.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetArtistHandler(c, db)
	})
	artists.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateArtistHandler(c, db)
	})
	artists.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteArtistHandler(c, db)
	})

	// Genre CRUD routes
	genres := api.Group("/genres")
	genres.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateGenreHandler(c, db)
	})
	genres.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetGenresHandler(c, db)
	})
	genres.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetGenreHandler(c, db)
	})
	genres.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateGenreHandler(c, db)
	})
	genres.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteGenreHandler(c, db)
	})

	// Music CRUD routes
	musics := api.Group("/musics")
	musics.Post("/upload", func(c *fiber.Ctx) error {
		return handlers.UploadMusicHandler(c, db)
	})
	musics.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateMusicHandler(c, db)
	})
	musics.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetMusicsHandler(c, db)
	})
	musics.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetMusicHandler(c, db)
	})
	musics.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateMusicHandler(c, db)
	})
	musics.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteMusicHandler(c, db)
	})
	musics.Post("/:id/play", func(c *fiber.Ctx) error {
		return handlers.IncrementPlayCountHandler(c, db)
	})
	musics.Post("/:id/like", func(c *fiber.Ctx) error {
		return handlers.IncrementLikeCountHandler(c, db)
	})

	// Music Video CRUD routes
	musicVideos := api.Group("/music-videos")
	musicVideos.Post("/upload", func(c *fiber.Ctx) error {
		return handlers.UploadMusicVideoHandler(c, db)
	})
	musicVideos.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateMusicVideoHandler(c, db)
	})
	musicVideos.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetMusicVideosHandler(c, db)
	})
	musicVideos.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetMusicVideoHandler(c, db)
	})
	musicVideos.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateMusicVideoHandler(c, db)
	})
	musicVideos.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteMusicVideoHandler(c, db)
	})

	// Notification CRUD routes
	notifications := api.Group("/notifications")
	notifications.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateNotificationHandler(c, db)
	})
	notifications.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetNotificationsHandler(c, db)
	})
	notifications.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetNotificationHandler(c, db)
	})
	notifications.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateNotificationHandler(c, db)
	})
	notifications.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteNotificationHandler(c, db)
	})
	notifications.Post("/:id/read", func(c *fiber.Ctx) error {
		return handlers.MarkAsReadHandler(c, db)
	})
	notifications.Get("/user/:user_id", func(c *fiber.Ctx) error {
		return handlers.GetUserNotificationsHandler(c, db)
	})
	notifications.Post("/user/:user_id/read-all", func(c *fiber.Ctx) error {
		return handlers.MarkAllAsReadHandler(c, db)
	})

	// Playlist CRUD routes
	playlists := api.Group("/playlists")
	playlists.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreatePlaylistHandler(c, db)
	})
	playlists.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetPlaylistsHandler(c, db)
	})
	playlists.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetPlaylistHandler(c, db)
	})
	playlists.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdatePlaylistHandler(c, db)
	})
	playlists.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeletePlaylistHandler(c, db)
	})

	// Playlist Songs CRUD routes
	playlistSongs := api.Group("/playlist-songs")
	playlistSongs.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreatePlaylistSongHandler(c, db)
	})
	playlistSongs.Get("/playlist/:playlist_id", func(c *fiber.Ctx) error {
		return handlers.GetPlaylistSongsHandler(c, db)
	})
	playlistSongs.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetPlaylistSongHandler(c, db)
	})
	playlistSongs.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdatePlaylistSongHandler(c, db)
	})
	playlistSongs.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeletePlaylistSongHandler(c, db)
	})
	playlistSongs.Delete("/playlist/:playlist_id/music/:music_id", func(c *fiber.Ctx) error {
		return handlers.DeletePlaylistSongByMusicHandler(c, db)
	})

	// Podcast CRUD routes
	podcasts := api.Group("/podcasts")
	podcasts.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreatePodcastHandler(c, db)
	})
	podcasts.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetPodcastsHandler(c, db)
	})
	podcasts.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetPodcastHandler(c, db)
	})
	podcasts.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdatePodcastHandler(c, db)
	})
	podcasts.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeletePodcastHandler(c, db)
	})

	// Subscription Plan CRUD routes
	subscriptionPlans := api.Group("/subscription-plans")
	subscriptionPlans.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateSubscriptionPlanHandler(c, db)
	})
	subscriptionPlans.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetSubscriptionPlansHandler(c, db)
	})
	subscriptionPlans.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetSubscriptionPlanHandler(c, db)
	})
	subscriptionPlans.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateSubscriptionPlanHandler(c, db)
	})
	subscriptionPlans.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteSubscriptionPlanHandler(c, db)
	})

	// News CRUD routes
	news := api.Group("/news")
	news.Post("/", func(c *fiber.Ctx) error {
		return handlers.CreateNewsHandler(c, db)
	})
	news.Get("/", func(c *fiber.Ctx) error {
		return handlers.GetNewsHandler(c, db)
	})
	news.Get("/:id", func(c *fiber.Ctx) error {
		return handlers.GetNewsByIDHandler(c, db)
	})
	news.Put("/:id", func(c *fiber.Ctx) error {
		return handlers.UpdateNewsHandler(c, db)
	})
	news.Delete("/:id", func(c *fiber.Ctx) error {
		return handlers.DeleteNewsHandler(c, db)
	})
}
