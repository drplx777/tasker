package middleware

import (
	"log/slog"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
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

		userID, ok := claims["sub"].(float64)
		if !ok {
			slog.Error("Invalid user ID in token", "userID", claims["sub"])
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
		}

		c.Locals("userID", int(userID))
		c.Locals("userRole", claims["role"])

		return c.Next()
	}
}
