package middleware

import (
	"log/slog"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// пропускаем публичные роуты
		if c.Path() == "/api/login" || c.Path() == "/api/register" {
			return c.Next()
		}

		token := c.Cookies("api_token")
		if token == "" {
			slog.Warn("Token not found in cookies")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token not found"})
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			slog.Warn("Invalid token", "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// sub может прийти как float64 — приведём к int
		userID := 0
		switch v := claims["sub"].(type) {
		case float64:
			userID = int(v)
		case int:
			userID = v
		case int64:
			userID = int(v)
		}

		c.Locals("userID", userID)
		c.Locals("userRole", claims["role"])

		return c.Next()
	}
}
