package main

import (
	"log"
	"os"
	"shiftplanner/backend/internal/api"
	"shiftplanner/backend/internal/database"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.CloseDB()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())

	// Get allowed origins from environment variable or use default (allow all)
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "*" // Allow all origins by default
	}

	// Split comma-separated origins
	allowedOrigins := []string{}
	hasWildcard := false
	for _, origin := range strings.Split(allowedOriginsStr, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "*" {
			hasWildcard = true
		} else {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}

	corsConfig := cors.Config{
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: !hasWildcard, // Wildcard ile credentials çalışmaz
	}

	if hasWildcard {
		corsConfig.AllowOrigins = "*"
	} else {
		corsConfig.AllowOrigins = strings.Join(allowedOrigins, ",")
	}

	app.Use(cors.New(corsConfig))

	// Auth routes (unprotected)
	app.Post("/api/auth/register", api.Register)
	app.Post("/api/auth/login", api.Login)
	app.Post("/api/auth/logout", api.Logout)

	// Holidays route (unprotected)
	app.Get("/api/holidays", api.GetHolidays)

	// API routes (protected)
	apiGroup := app.Group("/api", api.AuthMiddleware)
	apiGroup.Get("/members", api.GetMembers)
	apiGroup.Post("/members", api.CreateMember)
	apiGroup.Delete("/members/:id", api.DeleteMember)
	apiGroup.Get("/shifts", api.GetShifts)
	apiGroup.Post("/shifts/generate", api.GenerateShifts)
	apiGroup.Delete("/shifts", api.ClearAllShifts)
	apiGroup.Get("/stats", api.GetStats)
	apiGroup.Get("/leave-days", api.GetLeaveDays)
	apiGroup.Post("/leave-days", api.CreateLeaveDay)
	apiGroup.Delete("/leave-days/:id", api.DeleteLeaveDay)
	apiGroup.Put("/shifts/date", api.UpdateShiftForDate)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
