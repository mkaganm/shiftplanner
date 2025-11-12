package api

import (
	"shiftplanner/backend/internal/auth"

	"github.com/gofiber/fiber/v2"
)

const userIDKey = "userID"

// AuthMiddleware authentication middleware
func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	userID, err := auth.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Add UserID to locals
	c.Locals(userIDKey, userID)
	return c.Next()
}

// GetUserID gets user ID from Fiber context
func GetUserID(c *fiber.Ctx) int {
	if userID, ok := c.Locals(userIDKey).(int); ok {
		return userID
	}
	return 0
}
