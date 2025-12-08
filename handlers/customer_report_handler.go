package handlers

import (
	"backend_soundcave/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GetCustomerReportHandler mendapatkan laporan customer dengan revenue, growth, dll
func GetCustomerReportHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Get period from query (default: 30 days)
	periodDays := c.QueryInt("period", 30)
	if periodDays <= 0 {
		periodDays = 30
	}

	// Calculate date ranges
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousPeriodStart := periodStart.AddDate(0, 0, -periodDays)

	report := make(map[string]interface{})

	// ========== USER STATISTICS ==========
	var totalUsers int64
	var totalPremiumUsers int64
	var totalRegularUsers int64
	var totalAdminUsers int64

	db.Model(&models.User{}).Count(&totalUsers)
	db.Model(&models.User{}).Where("role = ?", "premium").Count(&totalPremiumUsers)
	db.Model(&models.User{}).Where("role = ?", "user").Count(&totalRegularUsers)
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&totalAdminUsers)

	// Current period user growth
	var currentPeriodNewUsers int64
	var previousPeriodNewUsers int64
	db.Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", periodStart, now).
		Count(&currentPeriodNewUsers)
	db.Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", previousPeriodStart, periodStart).
		Count(&previousPeriodNewUsers)

	// Calculate user growth percentage
	userGrowthPercent := float64(0)
	if previousPeriodNewUsers > 0 {
		userGrowthPercent = ((float64(currentPeriodNewUsers) - float64(previousPeriodNewUsers)) / float64(previousPeriodNewUsers)) * 100
	} else if currentPeriodNewUsers > 0 {
		userGrowthPercent = 100
	}

	// Premium user growth
	var currentPeriodNewPremium int64
	var previousPeriodNewPremium int64
	db.Model(&models.User{}).
		Where("role = ? AND created_at >= ? AND created_at < ?", "premium", periodStart, now).
		Count(&currentPeriodNewPremium)
	db.Model(&models.User{}).
		Where("role = ? AND created_at >= ? AND created_at < ?", "premium", previousPeriodStart, periodStart).
		Count(&previousPeriodNewPremium)

	premiumGrowthPercent := float64(0)
	if previousPeriodNewPremium > 0 {
		premiumGrowthPercent = ((float64(currentPeriodNewPremium) - float64(previousPeriodNewPremium)) / float64(previousPeriodNewPremium)) * 100
	} else if currentPeriodNewPremium > 0 {
		premiumGrowthPercent = 100
	}

	// Daily user growth (last 7 days)
	dailyUserGrowth := make([]map[string]interface{}, 7)
	for i := 6; i >= 0; i-- {
		dayStart := now.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)
		var dayCount int64
		db.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&dayCount)
		dailyUserGrowth[6-i] = map[string]interface{}{
			"date":  dayStart.Format("2006-01-02"),
			"count": dayCount,
		}
	}

	// Monthly user growth (last 6 months)
	monthlyUserGrowth := make([]map[string]interface{}, 6)
	for i := 5; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(i+1), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)
		var monthCount int64
		db.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", monthStart, monthEnd).
			Count(&monthCount)
		monthlyUserGrowth[5-i] = map[string]interface{}{
			"month": monthStart.Format("2006-01"),
			"count": monthCount,
		}
	}

	// ========== REVENUE ESTIMATION ==========
	// Get subscription plans and calculate estimated revenue
	var subscriptionPlans []models.SubscriptionPlan
	db.Find(&subscriptionPlans)

	// Calculate estimated monthly revenue based on premium users
	// Assuming average subscription price (we'll use the first plan or calculate average)
	estimatedMonthlyRevenue := float64(0)
	estimatedAnnualRevenue := float64(0)

	if len(subscriptionPlans) > 0 {
		// Try to parse prices and calculate average
		totalPrice := float64(0)
		validPrices := 0
		for _, plan := range subscriptionPlans {
			// Try to extract numeric value from price string (e.g., "$9.99" -> 9.99)
			priceStr := plan.Price
			// Remove common currency symbols and parse
			priceStr = removeCurrencySymbols(priceStr)
			if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
				totalPrice += price
				validPrices++
			}
		}

		if validPrices > 0 {
			avgPrice := totalPrice / float64(validPrices)
			estimatedMonthlyRevenue = avgPrice * float64(totalPremiumUsers)
			estimatedAnnualRevenue = estimatedMonthlyRevenue * 12
		}
	}

	// Revenue growth (based on premium user growth)
	revenueGrowthPercent := premiumGrowthPercent

	// ========== CONVERSION METRICS ==========
	conversionRate := float64(0)
	if totalUsers > 0 {
		conversionRate = (float64(totalPremiumUsers) / float64(totalUsers)) * 100
	}

	// ========== ENGAGEMENT METRICS ==========
	var totalPlaylists int64
	var totalPlaylistSongs int64
	var avgPlaylistsPerUser float64
	var avgSongsPerPlaylist float64

	db.Model(&models.Playlist{}).Count(&totalPlaylists)
	db.Model(&models.PlaylistSong{}).Count(&totalPlaylistSongs)

	if totalUsers > 0 {
		avgPlaylistsPerUser = float64(totalPlaylists) / float64(totalUsers)
	}
	if totalPlaylists > 0 {
		avgSongsPerPlaylist = float64(totalPlaylistSongs) / float64(totalPlaylists)
	}

	// ========== RETENTION METRICS ==========
	// Active users (users who registered in last 30 days or have recent activity)
	var activeUsers int64
	db.Model(&models.User{}).
		Where("created_at >= ?", periodStart).
		Count(&activeUsers)

	retentionRate := float64(0)
	if totalUsers > 0 {
		retentionRate = (float64(activeUsers) / float64(totalUsers)) * 100
	}

	// ========== USER DISTRIBUTION ==========
	userDistribution := map[string]interface{}{
		"premium":            totalPremiumUsers,
		"regular":            totalRegularUsers,
		"admin":              totalAdminUsers,
		"premium_percentage": conversionRate,
	}

	// ========== BUILD REPORT ==========
	report["period"] = map[string]interface{}{
		"days":           periodDays,
		"start_date":     periodStart.Format("2006-01-02"),
		"end_date":       now.Format("2006-01-02"),
		"previous_start": previousPeriodStart.Format("2006-01-02"),
		"previous_end":   periodStart.Format("2006-01-02"),
	}

	report["user_statistics"] = map[string]interface{}{
		"total_users":               totalUsers,
		"premium_users":             totalPremiumUsers,
		"regular_users":             totalRegularUsers,
		"admin_users":               totalAdminUsers,
		"current_period_new_users":  currentPeriodNewUsers,
		"previous_period_new_users": previousPeriodNewUsers,
		"user_growth_percent":       roundFloat(userGrowthPercent, 2),
		"premium_growth_percent":    roundFloat(premiumGrowthPercent, 2),
		"daily_growth":              dailyUserGrowth,
		"monthly_growth":            monthlyUserGrowth,
		"user_distribution":         userDistribution,
	}

	report["revenue"] = map[string]interface{}{
		"estimated_monthly_revenue": estimatedMonthlyRevenue,
		"estimated_annual_revenue":  estimatedAnnualRevenue,
		"revenue_growth_percent":    roundFloat(revenueGrowthPercent, 2),
		"premium_subscribers":       totalPremiumUsers,
		"currency":                  "USD", // Default, bisa diambil dari config
	}

	report["conversion_metrics"] = map[string]interface{}{
		"conversion_rate":         roundFloat(conversionRate, 2),
		"total_users":             totalUsers,
		"premium_users":           totalPremiumUsers,
		"regular_users":           totalRegularUsers,
		"current_period_premium":  currentPeriodNewPremium,
		"previous_period_premium": previousPeriodNewPremium,
	}

	report["engagement_metrics"] = map[string]interface{}{
		"total_playlists":        totalPlaylists,
		"total_playlist_songs":   totalPlaylistSongs,
		"avg_playlists_per_user": roundFloat(avgPlaylistsPerUser, 2),
		"avg_songs_per_playlist": roundFloat(avgSongsPerPlaylist, 2),
		"active_users":           activeUsers,
		"retention_rate":         roundFloat(retentionRate, 2),
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    report,
	})
}

// Helper function to remove currency symbols
func removeCurrencySymbols(s string) string {
	// Remove common currency symbols and spaces
	result := s
	result = strings.ReplaceAll(result, "$", "")
	result = strings.ReplaceAll(result, "€", "")
	result = strings.ReplaceAll(result, "£", "")
	result = strings.ReplaceAll(result, "¥", "")
	result = strings.ReplaceAll(result, "₹", "")
	result = strings.ReplaceAll(result, "Rp", "")
	result = strings.ReplaceAll(result, " ", "")
	result = strings.ReplaceAll(result, ",", "")
	return result
}

// Helper function to round float to specified decimal places
func roundFloat(val float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(val*multiplier+0.5)) / multiplier
}
