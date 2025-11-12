package api

import (
	"shiftplanner/backend/internal/auth"
	"shiftplanner/backend/internal/storage"

	"github.com/gofiber/fiber/v2"
)

// RegisterRequest registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Register registers a new user
func Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	// Username validation
	if len(req.Username) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username must be at least 3 characters",
		})
	}

	if len(req.Password) < 4 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 4 characters",
		})
	}

	// Create user
	user, err := storage.CreateUser(req.Username, req.Password)
	if err != nil {
		// Username might already be in use
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":  user,
		"token": session.Token,
	})
}

// Login handles user login
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	// Find user
	user, passwordHash, err := storage.GetUserByUsername(req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Validate password
	if !storage.ValidatePassword(req.Password, passwordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Create session
	session, err := auth.CreateSession(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"user":  user,
		"token": session.Token,
	})
}

// Logout handles user logout
func Logout(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token != "" {
		if err := auth.DeleteSession(token); err != nil {
			// Return 200 even if error (idempotent)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}
