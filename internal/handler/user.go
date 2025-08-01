package handler

import (
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterPublicRoutes(app *fiber.App) {
	app.Post("/user/login", h.loginHandler)
}

func (h *UserHandler) loginHandler(c fiber.Ctx) error {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	user, err := h.service.Login(c, req.Login, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	return c.JSON(user)
}
