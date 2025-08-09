package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"tasker/internal/model"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskService struct {
	dbPool *pgxpool.Pool
}

func NewTaskService(dbPool *pgxpool.Pool) *TaskService {
	return &TaskService{dbPool: dbPool}
}

func (s *TaskService) CreateTask(ctx context.Context, task model.Task) (*model.Task, error) {
	// Подготовка поля blockedBy из массива Blockers
	const defaultStatus = "to-do"

	var blockedBy sql.NullString
	if len(task.BlockedBy) > 0 && task.BlockedBy[0] != "" {
		blockedBy.String = task.BlockedBy[0]
		blockedBy.Valid = true
	}

	const query = `
    INSERT INTO tasks (
        title,
        description,
        status,
        "reporterD",
        "assignerID",
        "reviewerID",
        "approverID",
        "approveStatus",
        deadline,
        "dashboardID",
        "blockedBy"
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    RETURNING
        id,
        created_at,
        updated_at,
        "started_At",
        done_at
    `

	newTask := task                // копируем входную структуру
	newTask.Status = defaultStatus // устанавливаем статус по умолчанию
	err := s.dbPool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		newTask.Status,
		task.ReporterID,
		task.AssignerID,
		task.ReviewerID,
		task.ApproverID,
		task.ApproveStatus,
		task.DeadLine,
		task.DashboardID,
		blockedBy,
	).Scan(
		&newTask.ID,
		&newTask.CreatedAt,
		&newTask.UpdatedAt,
		&newTask.StartedAt,
		&newTask.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &newTask, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*model.Task, error) {
	const query = `
        SELECT 
            id, title, description, status, "reporterD", "assignerID", "reviewerID", 
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy"
        FROM tasks WHERE id = $1
    `
	var task model.Task
	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.ReporterID,
		&task.AssignerID,
		&task.ReviewerID,
		&task.ApproverID,
		&task.ApproveStatus,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.StartedAt,
		&task.CompletedAt,
		&task.DeadLine,
		&task.DashboardID,
		&task.BlockedBy,
	)

	return &task, err
}

func (s *TaskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	const query = `
        SELECT 
            id, title, description, status, "reporterD", "assignerID", "reviewerID", 
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy"
        FROM tasks
    `
	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.ReporterID,
			&task.AssignerID,
			&task.ReviewerID,
			&task.ApproverID,
			&task.ApproveStatus,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.StartedAt,
			&task.CompletedAt,
			&task.DeadLine,
			&task.DashboardID,
			&task.BlockedBy,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
func (s *TaskService) GetTasksByDashboardID(ctx context.Context, dashboardID string) ([]model.Task, error) {
	const query = `
		SELECT 
			id, title, description, status, "reporterD", "assignerID", "reviewerID", 
			"approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
			deadline, "dashboardID", "blockedBy"
		FROM tasks WHERE "dashboardID" = $1
	`

	rows, err := s.dbPool.Query(ctx, query, dashboardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var task model.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.ReporterID,
			&task.AssignerID,
			&task.ReviewerID,
			&task.ApproverID,
			&task.ApproveStatus,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.StartedAt,
			&task.CompletedAt,
			&task.DeadLine,
			&task.DashboardID,
			&task.BlockedBy,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, patch model.TaskPatch) (*model.Task, error) {
	// собираем части SET и аргументы
	set := []string{}
	args := []any{id}
	idx := 2

	push := func(col string, val any) {
		set = append(set, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, val)
		idx++
	}

	if patch.Title != nil {
		push("title", *patch.Title)
	}
	if patch.Description != nil {
		push("description", *patch.Description)
	}
	if patch.Status != nil {
		push("status", *patch.Status)
	}
	if patch.AssignerID != nil {
		push(`"assignerID"`, *patch.AssignerID)
	}
	if patch.ReviewerID != nil {
		push(`"reviewerID"`, *patch.ReviewerID)
	}
	if patch.ApproveStatus != nil {
		push(`"approveStatus"`, *patch.ApproveStatus)
	}
	if patch.StartedAt != nil {
		push(`"started_At"`, *patch.StartedAt)
	}
	if patch.CompletedAt != nil {
		push("done_at", *patch.CompletedAt)
	}
	if patch.DeadLine != nil {
		push("deadline", *patch.DeadLine)
	}
	if patch.DashboardID != nil {
		push(`"dashboardID"`, *patch.DashboardID)
	}
	if patch.BlockedBy != nil {
		// передаём []string как аргумент (pgx поддерживает массивы)
		push(`"blockedBy"`, *patch.BlockedBy)
	}
	if patch.ReporterID != nil {
		push(`"reporterD"`, *patch.ReporterID) // если в БД у тебя действительно колонка "reporterD"
	}
	if patch.ApproverID != nil {
		push(`"approverID"`, *patch.ApproverID)
	}
	if patch.ApproveStatus != nil {
		// уже добавляли выше, но оставил в случае расширения
	}

	// всегда обновляем updated_at
	set = append(set, "updated_at = NOW()")

	if len(set) == 1 { // только updated_at — ничего не передано
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
        UPDATE tasks
        SET %s
        WHERE id = $1
        RETURNING
            id, title, description, status, "reporterD", "assignerID", "reviewerID",
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy"
    `, strings.Join(set, ", "))

	var updated model.Task
	err := s.dbPool.QueryRow(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Description,
		&updated.Status,
		&updated.ReporterID,
		&updated.AssignerID,
		&updated.ReviewerID,
		&updated.ApproverID,
		&updated.ApproveStatus,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.StartedAt,
		&updated.CompletedAt,
		&updated.DeadLine,
		&updated.DashboardID,
		&updated.BlockedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task %s not found", id)
		}
		return nil, err
	}

	return &updated, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	_, err := s.dbPool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (s *TaskService) MarkTaskDone(ctx context.Context, id string) (*model.Task, error) {
	const selectQuery = `
        SELECT "blockedBy", "approveStatus"
        FROM tasks
        WHERE id = $1
    `
	var blockedBy sql.NullString
	var approveStatus string

	err := s.dbPool.QueryRow(ctx, selectQuery, id).Scan(&blockedBy, &approveStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %q not found", id)
		}
		return nil, err
	}

	// 2) Проверяем, заблокирована ли задача
	if blockedBy.Valid && blockedBy.String != "" {
		return nil, fmt.Errorf("cannot mark done: task is blocked by %s", blockedBy.String)
	}

	// 3) Проверяем статус одобрения
	if approveStatus == "need-approve" {
		return nil, errors.New("cannot mark done: task needs approval")
	}

	const query = `
        UPDATE tasks SET
            status = 'done',
            done_at = NOW(),
            updated_at = NOW()
        WHERE id = $1
        RETURNING 
            id, title, description, status, "reporterD", "assignerID", "reviewerID", 
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy"
    `

	var task model.Task
	err = s.dbPool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.ReporterID,
		&task.AssignerID,
		&task.ReviewerID,
		&task.ApproverID,
		&task.ApproveStatus,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.StartedAt,
		&task.CompletedAt,
		&task.DeadLine,
		&task.DashboardID,
		&task.BlockedBy,
	)

	return &task, err
}

//TO-DO: UpdateApproveStatugos
