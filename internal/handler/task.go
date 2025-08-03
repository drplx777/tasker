package handler

import (
	"tasker/internal/model"
	"tasker/internal/service"
	"time"

	"github.com/gofiber/fiber/v3"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/list", h.listTasks)
	app.Post("/create", h.createTask)
	app.Get("/task/by_id/:id", h.getTaskByID)
	app.Put("/update/:id", h.updateTask)
	app.Delete("/delete/:id", h.deleteTask)
	app.Put("/done/:id", h.doneTask)
	app.Get("/tasklist", h.mockTasks)
	app.Get("/taskByDB/:id", h.GetTasksByDashboardID)
}

func (h *TaskHandler) createTask(c fiber.Ctx) error {
	var task model.Task
	if err := c.Bind().JSON(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	createdTask, err := h.service.CreateTask(c, task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create task"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdTask)
}

func (h *TaskHandler) listTasks(c fiber.Ctx) error {
	tasks, err := h.service.ListTasks(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list tasks"})
	}
	return c.JSON(tasks)
}
func (h *TaskHandler) GetTasksByDashboardID(c fiber.Ctx) error {
	id := c.Params("id")
	tasks, err := h.service.GetTasksByDashboardID(c, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Tasks not found for this dashboard"})
	}
	return c.JSON(tasks)
}

func (h *TaskHandler) getTaskByID(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.GetTaskByID(c, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Task not found"})
	}
	return c.JSON(task)
}

func (h *TaskHandler) updateTask(c fiber.Ctx) error {
	id := c.Params("id")
	var task model.Task
	if err := c.Bind().JSON(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	updatedTask, err := h.service.UpdateTask(c, id, task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task"})
	}

	return c.JSON(updatedTask)
}

func (h *TaskHandler) deleteTask(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteTask(c, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete task"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TaskHandler) doneTask(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.service.MarkTaskDone(c, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to mark task as done"})
	}
	return c.JSON(task)
}

// ВРЕМЕННО
func (h *TaskHandler) mockTasks(c fiber.Ctx) error {
	assigner2 := "user-2"
	assigner3 := "user-3"
	assigner4 := "user-4"

	reviewer1 := "user-1"
	reviewer4 := "user-4"

	t1CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-20T09:00:00Z")
	t1StartedAt, _ := time.Parse(time.RFC3339, "2025-07-21T10:00:00Z")

	t2CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-19T14:00:00Z")

	t3CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-15T08:00:00Z")
	t3StartedAt, _ := time.Parse(time.RFC3339, "2025-07-16T09:30:00Z")
	t3CompletedAt, _ := time.Parse(time.RFC3339, "2025-07-20T17:00:00Z")

	t4CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-21T10:00:00Z")

	t5CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-22T12:00:00Z")
	t5StartedAt, _ := time.Parse(time.RFC3339, "2025-07-23T13:00:00Z")

	t6CreatedAt, _ := time.Parse(time.RFC3339, "2025-07-15T10:00:00Z")
	t6CompletedAt, _ := time.Parse(time.RFC3339, "2025-07-18T18:00:00Z")

	tasks := []model.Task{
		{
			ID:            "1",
			Title:         "Добавить авторизацию",
			Description:   "Регистрация, логин, защита роутов",
			Status:        "in-progress",
			ReporterID:    "user-1",
			AssignerID:    &assigner2,
			ReviewerID:    &reviewer1,
			ApproverID:    "user-1",
			ApproveStatus: "approved",
			CreatedAt:     t1CreatedAt,
			StartedAt:     &t1StartedAt,
			CompletedAt:   nil,
			DeadLine:      "2025-07-30",
			DashboardID:   "dash-1",
			BlockedBy:     []string{},
		},
		{
			ID:            "2",
			Title:         "Создать UI главного дашборда",
			Description:   "Компоненты, таблицы, графики",
			Status:        "to-do",
			ReporterID:    "user-2",
			AssignerID:    nil,
			ReviewerID:    nil,
			ApproverID:    "user-2",
			ApproveStatus: "need-approval",
			CreatedAt:     t2CreatedAt,
			StartedAt:     nil,
			CompletedAt:   nil,
			DeadLine:      "2025-08-01",
			DashboardID:   "dash-1",
			BlockedBy:     []string{"1"},
		},
		{
			ID:            "3",
			Title:         "Настроить CI/CD",
			Description:   "GitHub Actions + Vercel",
			Status:        "done",
			ReporterID:    "user-1",
			AssignerID:    &assigner3,
			ReviewerID:    &reviewer1,
			ApproverID:    "user-4",
			ApproveStatus: "approved",
			CreatedAt:     t3CreatedAt,
			StartedAt:     &t3StartedAt,
			CompletedAt:   &t3CompletedAt,
			DeadLine:      "2025-07-25",
			DashboardID:   "dash-2",
			BlockedBy:     []string{},
		},
		{
			ID:            "4",
			Title:         "Рефакторинг компонентов",
			Description:   "Разделение логики и UI",
			Status:        "blocked",
			ReporterID:    "user-2",
			AssignerID:    &assigner4,
			ReviewerID:    nil,
			ApproverID:    "user-3",
			ApproveStatus: "need-approval",
			CreatedAt:     t4CreatedAt,
			StartedAt:     nil,
			CompletedAt:   nil,
			DeadLine:      "2025-07-29",
			DashboardID:   "dash-2",
			BlockedBy:     []string{"1", "3"},
		},
		{
			ID:            "5",
			Title:         "Интеграция с беком",
			Description:   "REST API, fetch, hook-логика",
			Status:        "review",
			ReporterID:    "user-3",
			AssignerID:    &assigner2,
			ReviewerID:    &reviewer4,
			ApproverID:    "user-3",
			ApproveStatus: "approval",
			CreatedAt:     t5CreatedAt,
			StartedAt:     &t5StartedAt,
			CompletedAt:   nil,
			DeadLine:      "2025-08-02",
			DashboardID:   "dash-1",
			BlockedBy:     []string{},
		},
		{
			ID:            "6",
			Title:         "Создать страницу профиля",
			Description:   "Имя, логин, роль, смена пароля",
			Status:        "canceled",
			ReporterID:    "user-1",
			AssignerID:    nil,
			ReviewerID:    nil,
			ApproverID:    "user-1",
			ApproveStatus: "rejected",
			CreatedAt:     t6CreatedAt,
			StartedAt:     nil,
			CompletedAt:   &t6CompletedAt,
			DeadLine:      "2025-08-05",
			DashboardID:   "dash-5",
			BlockedBy:     []string{},
		},
	}

	return c.JSON(tasks)
}
