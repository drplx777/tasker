package handler

import (
	"tasker/internal/model"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

type DashboardsHandler struct {
	service *service.DashboardService
}

func NewDashboardsHandler(service *service.DashboardService) *DashboardsHandler {
	return &DashboardsHandler{service: service}
}

func (h *DashboardsHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/ShowDB", h.listDashboards)
	app.Get("/GetDBbyId/:id", h.GetDashboardById)
	app.Post("/CreateDB", h.CreateDB)
}
func (h *DashboardsHandler) listDashboards(c fiber.Ctx) error {
	dashboards, err := h.service.ListDashboards(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list dashboards"})
	}
	return c.JSON(dashboards)
}

func (h *DashboardsHandler) GetDashboardById(c fiber.Ctx) error {
	id := c.Params("id")
	dashboard, err := h.service.GetDashboardById(c, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Dashboard not found"})
	}
	return c.JSON(dashboard)
}

func (h *DashboardsHandler) CreateDB(c fiber.Ctx) error {
	var dashboard model.DashBoards
	if err := c.Bind().JSON(&dashboard); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	createdDashboard, err := h.service.CreateDashboard(c, dashboard)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create dashboard"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdDashboard)
}
