package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func SlogLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		duration := time.Since(start)
		slog.Info("Request from api",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
			"ip", c.IP(),
			"request_id", c.Get(fiber.HeaderXRequestID),
			"user_agent", c.Get(fiber.HeaderUserAgent),
		)

		return err
	}
}
