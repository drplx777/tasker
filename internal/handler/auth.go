package handler

import (
	"tasker/internal/model"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	app.Post("/api/register", h.registerHandler)
	app.Post("/api/login", h.loginHandler)
	app.Get("/api/validate", h.validateTokenHandler)
}

// Все хендлеры принимают fiber.Ctx (интерфейс), который реализует context.Context
func (h *AuthHandler) registerHandler(c fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Передаём сам Ctx как context.Context
	user, err := h.service.Register(c, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Registration failed"})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *AuthHandler) loginHandler(c fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Передаём Ctx напрямую
	token, _, exp, err := h.service.Login(c, req.Login, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}
	c.Cookie(&fiber.Cookie{
		Expires:  exp,
		Name:     "api_token",
		Value:    token,
		Path:     "/",
		Domain:   "localhost",
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   false,
		HTTPOnly: false,
	})

	return c.SendStatus(200)
}

func (h *AuthHandler) validateTokenHandler(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Token is required"})
	}

	claims, err := h.service.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	return c.JSON(fiber.Map{
		"valid":    true,
		"userID":   claims["sub"],
		"userRole": claims["role"],
	})
}

func (h *AuthHandler) GetUserHandler(c fiber.Ctx) error {
	// Получаем userID из middleware
	userIDVal := c.Locals("userID")
	userID, ok := userIDVal.(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
	}

	// Ctx реализует context.Context
	user, err := h.service.GetUserByID(c, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}
