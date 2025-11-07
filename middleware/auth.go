package middleware

import (
	"alumni-crud-api/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired middleware for authentication
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Token akses diperlukan",
			})
		}

		// Extract token from "Bearer TOKEN"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Format token tidak valid",
			})
		}

		// Validate token
		claims, err := helper.ValidateToken(tokenParts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid atau expired",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID) // Ini sekarang string
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

// Middleware for admin-only access
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"message": "Akses ditolak. Hanya admin yang diizinkan",
			})
		}
		return c.Next()
	}
}
