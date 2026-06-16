package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FiberMiddleware returns Fiber middleware that records API request metrics
func FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		statusCode := strconv.Itoa(c.Response().StatusCode())

		// Force-copy strings from fast-http buffer to avoid label corruption
		// when request context is recycled by the pool
		method := strings.Clone(c.Method())
		path := strings.Clone(c.Path())

		// Record metrics
		RecordAPIDuration(method, path, statusCode, duration)

		return err
	}
}
