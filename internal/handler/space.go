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
	grp.Post("/:id/dashboards", h.createDashboard)         // POST /spaces/:id/dashboards
	grp.Delete("/:id/dashboards/:dbid", h.deleteDashboard) // DELETE /spaces/:id/dashboards/:dbid
	grp.Get("/:id/dashboards", h.listSpaceDashboards)      // GET /spaces/:id/dashboards

	grp.Delete("/:id/members/:userId", h.removeMember)  // DELETE member
	grp.Put("/:id/members/:userId", h.updateMemberRole) // PUT change role
	grp.Get("/:id/token", h.getSpaceToken)              // GET token (space id)
	grp.Post("/join", h.joinByToken)                    // POST /spaces/jjoin body { token, role(opt) }

	// roles
	grp.Post("/:id/roles", h.addRole)
	grp.Delete("/:id/roles/:role", h.removeRole)
	grp.Get("/:id/roles", h.listRoles)

	grp.Get("/:id", h.getSpaceByID) // GET /spaces/:id
	grp.Get("/", h.listMySpaces)    // GET /spaces
}

func (h *SpaceHandler) createDashboard(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	// проверка прав: только admin может
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

// remove member
func (h *SpaceHandler) removeMember(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	spaceID := c.Params("id")
	// только admin
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

// update member role
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

// get token (space id) — доступен только admin/creator
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

// join by token
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

// roles
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

// getSpaceByID — GET /spaces/:id
func (h *SpaceHandler) getSpaceByID(c fiber.Ctx) error {
	spaceID := c.Params("id")
	if spaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "space id required"})
	}

	// Если позже добавим SpaceService.GetSpaceByID, здесь можно вернуть полную запись.
	// Сейчас возвращаем минимальную информацию.
	return c.JSON(fiber.Map{"id": spaceID})
}

// listMySpaces — GET /spaces
// Возвращает список пространств, где текущий пользователь состоит.
// Пока не реализовано в SpaceService, возвращаем 501.
func (h *SpaceHandler) listMySpaces(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "not implemented, add ListSpaces in SpaceService"})
}

// getUserIDFromCtx получает user id (int) из контекста Fiber.
// Сначала смотрит c.Locals("userID"), затем заголовок X-User-ID.
func getUserIDFromCtx(c fiber.Ctx) (int, error) {
	// Попробуем c.Locals (middleware может положить туда user id)
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
		}
	}

	// fallback: заголовок X-User-ID
	if h := c.Get("X-User-ID"); h != "" {
		if id, err := strconv.Atoi(h); err == nil {
			return id, nil
		}
	}

	return 0, fiber.ErrUnauthorized
}
