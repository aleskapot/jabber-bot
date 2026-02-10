package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// AuthMiddleware provides API key authentication middleware
func (s *Server) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication if disabled
		if !s.config.API.Enabled {
			return c.Next()
		}

		logger := c.Locals("logger").(*zap.Logger)

		// Get API key from header
		apiKey := c.Get("API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer token
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey != s.config.API.APIKey {
			logger.Warn("Unauthorized access attempt",
				zap.String("remote_addr", c.IP()),
				zap.String("user_agent", c.Get("User-Agent")),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)

			response := map[string]interface{}{
				"success": false,
				"error":   "Unauthorized - valid API key required",
				"code":    401,
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response)
		}

		logger.Debug("API key authenticated successfully",
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		// Continue to next handler
		return c.Next()
	}
}

// OptionalAuthMiddleware provides optional API key authentication
// Returns user info in context if key is provided, but doesn't reject if missing
func (s *Server) OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication if disabled
		if !s.config.API.Enabled {
			return c.Next()
		}

		// Get API key from header
		apiKey := c.Get("API-Key")
		if apiKey == "" {
			// Also check Authorization header with Bearer token
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Set authentication status in context
		if apiKey == s.config.API.APIKey {
			c.Locals("authenticated", true)
		} else {
			c.Locals("authenticated", false)
		}

		// Continue to next handler
		return c.Next()
	}
}
