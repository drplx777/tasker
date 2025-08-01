package middleware

import (
	"log/slog"
	"strings"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Path() == "/api/login" || c.Path() == "/api/register" {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			slog.Warn("Authorization header missing")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header missing"})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			slog.Warn("Invalid token format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token format"})
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
