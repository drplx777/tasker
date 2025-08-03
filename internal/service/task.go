package service

import (
	"context"
	"errors"
	"fmt"
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
	var dashboardID string
	err := s.dbPool.QueryRow(ctx, "SELECT id FROM dashboards WHERE name = $1", task.DashboardName).Scan(&dashboardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("dashboard not found: %s", task.DashboardName)
		}
		return nil, err
	}
	task.DashboardID = dashboardID
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
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING 
            id,
            created_at,
            updated_at,
            "started_At",
            done_at
    `

	newTask := task

	err = s.dbPool.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.ReporterID,
		task.AssignerID,
		task.ReviewerID,
		task.ApproverID,
		task.ApproveStatus,
		task.DeadLine,
		task.DashboardID,
		task.BlockedBy,
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

func (s *TaskService) UpdateTask(ctx context.Context, id string, task model.Task) (*model.Task, error) {
	const query = `
        UPDATE tasks SET
            title = COALESCE($2, title),
            description = COALESCE($3, description),
            status = COALESCE($4, status),
            "assignerID" = COALESCE($5, "assignerID"),
            "reviewerID" = COALESCE($6, "reviewerID"),
            "approveStatus" = COALESCE($7, "approveStatus"),
            "started_At" = COALESCE($8, "started_At"),
            done_at = COALESCE($9, done_at),
            deadline = COALESCE($10, deadline),
            "dashboardID" = COALESCE($11, "dashboardID"),
            "blockedBy" = COALESCE($12, "blockedBy"),
            updated_at = NOW()
        WHERE id = $1
        RETURNING 
            id, title, description, status, "reporterD", "assignerID", "reviewerID", 
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy"
    `

	var updatedTask model.Task
	err := s.dbPool.QueryRow(ctx, query, id,
		task.Title,
		task.Description,
		task.Status,
		task.AssignerID,
		task.ReviewerID,
		task.ApproveStatus,
		task.StartedAt,
		task.CompletedAt,
		task.DeadLine,
		task.DashboardID,
		task.BlockedBy,
	).Scan(
		&updatedTask.ID,
		&updatedTask.Title,
		&updatedTask.Description,
		&updatedTask.Status,
		&updatedTask.ReporterID,
		&updatedTask.AssignerID,
		&updatedTask.ReviewerID,
		&updatedTask.ApproverID,
		&updatedTask.ApproveStatus,
		&updatedTask.CreatedAt,
		&updatedTask.UpdatedAt,
		&updatedTask.StartedAt,
		&updatedTask.CompletedAt,
		&updatedTask.DeadLine,
		&updatedTask.DashboardID,
		&updatedTask.BlockedBy,
	)

	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	_, err := s.dbPool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

func (s *TaskService) MarkTaskDone(ctx context.Context, id string) (*model.Task, error) {
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
