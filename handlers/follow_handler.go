package handlers

import (
	"backend_soundcave/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// FollowRequest struct untuk request follow user
type FollowRequest struct {
	TargetUserID uint `json:"target_user_id" validate:"required"`
}

// FollowUserHandler menangani follow user (independent/label) oleh user biasa
// @Summary      Follow user
// @Description  Follow an independent artist or label user. Only users with role "user" can follow users with role "independent" or "label"
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      FollowRequest  true  "Follow Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      409      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/follow [post]
func FollowUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Get current user info from JWT
	currentUserID := c.Locals("user_id").(uint)
	currentRole := c.Locals("role").(string)

	// Only role "user" can follow
	if currentRole != string(models.RoleUser) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya user dengan role 'user' yang dapat follow",
		})
	}

	var req FollowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	if req.TargetUserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Field target_user_id wajib diisi",
		})
	}

	// Cannot follow yourself
	if req.TargetUserID == currentUserID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tidak dapat follow diri sendiri",
		})
	}

	// Find target user and validate role
	var targetUser models.User
	if err := db.First(&targetUser, req.TargetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User target tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user target",
			"error":   err.Error(),
		})
	}

	// Target must be independent or label
	if targetUser.Role != models.RoleIndependent && targetUser.Role != models.RoleLabel {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Hanya bisa follow user dengan role 'independent' atau 'label'",
		})
	}

	// Check if already following
	currentUserIDStr := fmt.Sprintf("%d", currentUserID)
	if targetUser.Followers != nil {
		for _, followerID := range targetUser.Followers {
			if followerID == currentUserIDStr {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Anda sudah follow user ini",
				})
			}
		}
	}

	// Add current user to target's followers
	if targetUser.Followers == nil {
		targetUser.Followers = models.JSONStringArray{}
	}
	targetUser.Followers = append(targetUser.Followers, currentUserIDStr)
	targetUser.TotalFollower = len(targetUser.Followers)

	if err := db.Save(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal follow user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil follow user",
		"data": fiber.Map{
			"target_user_id": targetUser.ID,
			"target_name":    targetUser.FullName,
			"total_follower": targetUser.TotalFollower,
		},
	})
}

// UnfollowUserHandler menangani unfollow user
// @Summary      Unfollow user
// @Description  Unfollow an independent artist or label user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        request  body      FollowRequest  true  "Unfollow Request"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Failure      403      {object}  map[string]interface{}
// @Failure      404      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/unfollow [post]
func UnfollowUserHandler(c *fiber.Ctx, db *gorm.DB) error {
	// Get current user info from JWT
	currentUserID := c.Locals("user_id").(uint)
	currentRole := c.Locals("role").(string)

	// Only role "user" can unfollow
	if currentRole != string(models.RoleUser) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Hanya user dengan role 'user' yang dapat unfollow",
		})
	}

	var req FollowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Gagal parse request body",
			"error":   err.Error(),
		})
	}

	if req.TargetUserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Field target_user_id wajib diisi",
		})
	}

	// Find target user
	var targetUser models.User
	if err := db.First(&targetUser, req.TargetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User target tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user target",
			"error":   err.Error(),
		})
	}

	// Check if following and remove
	currentUserIDStr := fmt.Sprintf("%d", currentUserID)
	found := false
	newFollowers := models.JSONStringArray{}
	for _, followerID := range targetUser.Followers {
		if followerID == currentUserIDStr {
			found = true
			continue
		}
		newFollowers = append(newFollowers, followerID)
	}

	if !found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Anda belum follow user ini",
		})
	}

	targetUser.Followers = newFollowers
	targetUser.TotalFollower = len(newFollowers)

	if err := db.Save(&targetUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal unfollow user",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil unfollow user",
		"data": fiber.Map{
			"target_user_id": targetUser.ID,
			"target_name":    targetUser.FullName,
			"total_follower": targetUser.TotalFollower,
		},
	})
}
