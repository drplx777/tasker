package handler

import (
	"strconv"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

type SpaceHandler struct {
	spaceSvc *service.SpaceService
	dashSvc  *service.DashboardService
}

func NewSpaceHandler(spaceSvc *service.SpaceService, dashSvc *service.DashboardService) *SpaceHandler {
	return &SpaceHandler{spaceSvc: spaceSvc, dashSvc: dashSvc}
}

func (h *SpaceHandler) RegisterRoutes(app *fiber.App) {
	grp := app.Group("/spaces")
	grp.Post("/:id/dashboards", h.createDashboard)
	grp.Delete("/:id/dashboards/:dbid", h.deleteDashboard)
	grp.Get("/:id/dashboards", h.listSpaceDashboards)

	grp.Delete("/:id/members/:userId", h.removeMember)
	grp.Put("/:id/members/:userId", h.updateMemberRole)
	grp.Get("/:id/token", h.getSpaceToken)
	grp.Post("/join", h.joinByToken)

	grp.Post("/:id/roles", h.addRole)
	grp.Delete("/:id/roles/:role", h.removeRole)
	grp.Get("/:id/roles", h.listRoles)

	grp.Get("/:id", h.getSpaceByID)
	grp.Get("/", h.listMySpaces)
}

func (h *SpaceHandler) createDashboard(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")

	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can create dashboards"})
	}

	var in struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&in); err != nil || in.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name required"})
	}

	d, err := h.dashSvc.CreateDashboardForSpace(c, spaceID, in.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(d)
}

func (h *SpaceHandler) deleteDashboard(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	dbid := c.Params("dbid")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can delete dashboards"})
	}
	if err := h.dashSvc.DeleteDashboardFromSpace(c, spaceID, dbid); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SpaceHandler) listSpaceDashboards(c fiber.Ctx) error {
	spaceID := c.Params("id")
	ds, err := h.dashSvc.ListDashboardsBySpace(c, spaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ds)
}

func (h *SpaceHandler) removeMember(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can remove members"})
	}

	uidStr := c.Params("userId")
	targetID, _ := strconv.Atoi(uidStr)
	if targetID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	if err := h.spaceSvc.RemoveMember(c, spaceID, targetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SpaceHandler) updateMemberRole(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can update member roles"})
	}

	uidStr := c.Params("userId")
	targetID, _ := strconv.Atoi(uidStr)
	var in struct {
		Role string `json:"role"`
	}
	if err := c.Bind().Body(&in); err != nil || in.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role required"})
	}

	if err := h.spaceSvc.UpdateMemberRole(c, spaceID, targetID, in.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SpaceHandler) getSpaceToken(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can get space token"})
	}

	return c.JSON(fiber.Map{"token": spaceID})
}

func (h *SpaceHandler) joinByToken(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var in struct {
		Token string `json:"token"`
		Role  string `json:"role,omitempty"`
	}
	if err := c.Bind().Body(&in); err != nil || in.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token required"})
	}
	if in.Role == "" {
		in.Role = "member"
	}
	if err := h.spaceSvc.JoinByToken(c, in.Token, uid, in.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SpaceHandler) addRole(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can manage roles"})
	}
	var in struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&in); err != nil || in.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name required"})
	}
	if err := h.spaceSvc.AddSpaceRole(c, spaceID, in.Name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusCreated)
}

func (h *SpaceHandler) removeRole(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can manage roles"})
	}
	roleName := c.Params("role")
	if roleName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role required"})
	}
	if err := h.spaceSvc.RemoveSpaceRole(c, spaceID, roleName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SpaceHandler) listRoles(c fiber.Ctx) error {
	spaceID := c.Params("id")
	roles, err := h.spaceSvc.ListSpaceRoles(c, spaceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(roles)
}

func (h *SpaceHandler) getSpaceByID(c fiber.Ctx) error {
	spaceID := c.Params("id")
	if spaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "space id required"})
	}
	return c.JSON(fiber.Map{"id": spaceID})
}

func (h *SpaceHandler) listMySpaces(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented, add ListSpaces in SpaceService"})
}

// getUserIDFromCtx — helper
func getUserIDFromCtx(c fiber.Ctx) (int, error) {
	if v := c.Locals("userID"); v != nil {
		switch t := v.(type) {
		case int:
			return t, nil
		case int64:
			return int(t), nil
		case string:
			if id, err := strconv.Atoi(t); err == nil {
				return id, nil
			}
		case float64:
			return int(t), nil
		}
	}

	if h := c.Get("X-User-ID"); h != "" {
		if id, err := strconv.Atoi(h); err == nil {
			return id, nil
		}
	}

	return 0, fiber.ErrUnauthorized
}
