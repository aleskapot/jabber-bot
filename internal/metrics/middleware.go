package metrics

import (
	"strconv"
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

		// Get method and path
		method := c.Method()
		path := c.Path()

		// Record metrics
		RecordAPIDuration(method, path, statusCode, duration)

		return err
	}
}
