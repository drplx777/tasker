package handler

import (
	"strconv"
	"tasker/internal/service"

	"github.com/gofiber/fiber/v3"
)

// SpaceHandler обрабатывает запросы, связанные с Spaces.
type SpaceHandler struct {
	spaceSvc *service.SpaceService
}

// NewSpaceHandler создаёт новый SpaceHandler.
func NewSpaceHandler(spaceSvc *service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaceSvc: spaceSvc}
}

// RegisterRoutes регистрирует роуты, связанные с пространствами.
func (h *SpaceHandler) RegisterRoutes(app *fiber.App) {
	grp := app.Group("/spaces")
	grp.Post("/", h.createSpace)             // POST /spaces
	grp.Post("/:id/invite", h.inviteToSpace) // POST /spaces/:id/invite
	grp.Get("/:id", h.getSpaceByID)          // GET /spaces/:id
	grp.Get("/", h.listMySpaces)             // GET /spaces
}

// createSpace — POST /spaces
// Body: { "name": "Frontend Team" }
func (h *SpaceHandler) createSpace(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var in struct {
		Name string `json:"name"`
	}
	if err := c.Bind().JSON(&in); err != nil || in.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body, name is required"})
	}

	sp, err := h.spaceSvc.CreateSpace(c, in.Name, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create space", "detail": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(sp)
}

// inviteToSpace — POST /spaces/:id/invite
// Body: { "userId": 42, "role": "member" }
func (h *SpaceHandler) inviteToSpace(c fiber.Ctx) error {
	uid, err := getUserIDFromCtx(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	spaceID := c.Params("id")
	if spaceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "space id required"})
	}

	// Проверяем, что вызывающий пользователь — админ в этом пространстве
	isMember, role, err := h.spaceSvc.IsMember(c, spaceID, uid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error", "detail": err.Error()})
	}
	if !isMember || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "only admin can invite users to space"})
	}

	var in struct {
		UserID int    `json:"userId"`
		Role   string `json:"role"`
	}
	if err := c.Bind().JSON(&in); err != nil || in.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body, userId required"})
	}
	if in.Role == "" {
		in.Role = "member"
	}

	if err := h.spaceSvc.AddMember(c, spaceID, in.UserID, in.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to add member", "detail": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
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
