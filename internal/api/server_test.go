package api

import (
	"net/http/httptest"
	"testing"

	"jabber-bot/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestServerErrorHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		API: config.APIConfig{
			Port: 8080,
			Host: "localhost",
		},
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		return c.Next()
	})

	// Test route that returns error
	app.Get("/test-error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "Test error message")
	})

	// Perform request
	req := httptest.NewRequest("GET", "/test-error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestServerErrorHandler_UnexpectedError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		API: config.APIConfig{
			Port: 8080,
			Host: "localhost",
		},
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("logger", logger)
		c.Locals("config", cfg)
		return c.Next()
	})

	// Test route that returns unexpected error
	app.Get("/unexpected-error", func(c *fiber.Ctx) error {
		return assert.AnError
	})

	// Perform request
	req := httptest.NewRequest("GET", "/unexpected-error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
