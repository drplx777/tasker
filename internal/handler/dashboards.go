package handler

import (
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
}
func (h *DashboardsHandler) listDashboards(c fiber.Ctx) error {
	dashboards, err := h.service.ListDashboards(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list dashboards"})
	}
	return c.JSON(dashboards)
}
