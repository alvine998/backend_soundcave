package middleware

import (
	"backend_soundcave/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware middleware untuk memverifikasi JWT token
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token tidak ditemukan",
		})
	}

	// Extract token dari "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Format token tidak valid",
		})
	}

	token := parts[1]

	// Validate token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Token tidak valid atau sudah expired",
			"error":   err.Error(),
		})
	}

	// Simpan claims ke context
	c.Locals("user_id", uint(claims.UserID))
	c.Locals("email", claims.Email)
	c.Locals("role", claims.Role)

	return c.Next()
}

// AdminMiddleware middleware untuk memverifikasi admin role
func AdminMiddleware(c *fiber.Ctx) error {
	role := c.Locals("role")
	if role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akses ditolak. Hanya admin yang dapat mengakses",
		})
	}

	return c.Next()
}

