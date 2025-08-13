package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tasker/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskService struct {
	dbPool *pgxpool.Pool
	spaces *SpaceService
}

// NewTaskService принимает пул и инстанс SpaceService (для проверки членства).
func NewTaskService(dbPool *pgxpool.Pool, spaces *SpaceService) *TaskService {
	return &TaskService{dbPool: dbPool, spaces: spaces}
}

func (s *TaskService) CreateTask(ctx context.Context, task model.Task) (*model.Task, error) {
	// валидация поля space
	if task.Space == nil || *task.Space == "" {
		return nil, fmt.Errorf("space cannot be empty")
	}

	// проверяем, что репортер — член пространства
	reporterID, err := strconv.Atoi(task.ReporterID)
	if err != nil {
		return nil, fmt.Errorf("invalid reporter ID: %w", err)
	}
	isMember, _, err := s.spaces.IsMember(ctx, *task.Space, reporterID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("reporter is not a member of the space")
	}

	// статус по умолчанию, если не задан
	if task.Status == "" {
		task.Status = "to-do"
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
        "blockedBy",
        space
    )
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
    RETURNING id, created_at, updated_at, "started_At", done_at
    `

	var blockedBy []string
	if len(task.BlockedBy) > 0 {
		blockedBy = task.BlockedBy
	}

	// выполняем вставку и считываем автогенерируемые поля
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
		blockedBy,
		task.Space,
	).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.StartedAt,
		&task.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, id string) (*model.Task, error) {
	const query = `
    SELECT
      t.id, t.title, t.description, t.status, t."reporterD", t."assignerID", t."reviewerID", t."approverID",
      t."approveStatus", t.created_at, t.updated_at, t."started_At", t.done_at,
      t.deadline, t."dashboardID", t."blockedBy", t.space,
      (rep.name || ' ' || rep.surname) AS reporter_name,
      (ass.name || ' ' || ass.surname) AS assigner_name,
      (app.name || ' ' || app.surname) AS approver_name,
      d.name AS dashboard_name
    FROM tasks t
    LEFT JOIN users rep ON rep.id::text = t."reporterD"
    LEFT JOIN users ass ON ass.id::text = t."assignerID"
    LEFT JOIN users app ON app.id::text = t."approverID"
    LEFT JOIN dashboards d ON d.id::text = t."dashboardID"
    WHERE t.id = $1
    `

	var task model.Task
	var reporterName sql.NullString
	var assignerName sql.NullString
	var approverName sql.NullString
	var dashboardName sql.NullString
	var space sql.NullString
	var blockedBy []string

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
		&blockedBy,
		&space,
		&reporterName,
		&assignerName,
		&approverName,
		&dashboardName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task %s not found", id)
		}
		return nil, err
	}

	task.BlockedBy = blockedBy
	if space.Valid {
		sv := space.String
		task.Space = &sv
	}

	if reporterName.Valid {
		v := reporterName.String
		task.ReporterName = &v
	}
	if assignerName.Valid {
		v := assignerName.String
		task.AssignerName = &v
	}
	if approverName.Valid {
		v := approverName.String
		task.ApproverName = &v
	}
	if dashboardName.Valid {
		v := dashboardName.String
		task.DashboardName = &v
	}

	return &task, nil
}

func (s *TaskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	const query = `
        SELECT 
            id, title, description, status, "reporterD", "assignerID", "reviewerID", 
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy", space
        FROM tasks
    `
	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var blockedBy []string
		var space sql.NullString

		if err := rows.Scan(
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
			&blockedBy,
			&space,
		); err != nil {
			return nil, err
		}

		task.BlockedBy = blockedBy
		if space.Valid {
			sv := space.String
			task.Space = &sv
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) GetTasksByDashboardID(ctx context.Context, dashboardID string) ([]model.Task, error) {
	const query = `
		SELECT 
			id, title, description, status, "reporterD", "assignerID", "reviewerID", 
			"approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
			deadline, "dashboardID", "blockedBy", space
		FROM tasks WHERE "dashboardID" = $1
	`

	rows, err := s.dbPool.Query(ctx, query, dashboardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var blockedBy []string
		var space sql.NullString

		if err := rows.Scan(
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
			&blockedBy,
			&space,
		); err != nil {
			return nil, err
		}

		task.BlockedBy = blockedBy
		if space.Valid {
			sv := space.String
			task.Space = &sv
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, patch model.TaskPatch) (*model.Task, error) {
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
		push(`"blockedBy"`, *patch.BlockedBy)
	}
	if patch.ReporterID != nil {
		push(`"reporterD"`, *patch.ReporterID)
	}
	if patch.ApproverID != nil {
		push(`"approverID"`, *patch.ApproverID)
	}

	// всегда обновляем updated_at
	push("updated_at", time.Now())

	query := fmt.Sprintf(`
        UPDATE tasks
        SET %s
        WHERE id = $1
        RETURNING
            id, title, description, status, "reporterD", "assignerID", "reviewerID",
            "approverID", "approveStatus", created_at, updated_at, "started_At", done_at,
            deadline, "dashboardID", "blockedBy", space
    `, strings.Join(set, ", "))

	var updated model.Task
	// QueryRow принимает args... (распакуем slice)
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
		&updated.Space,
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
        SELECT blockedBy, "approveStatus"
        FROM tasks
        WHERE id = $1
    `
	var blockedBy []string
	var approveStatus string

	err := s.dbPool.QueryRow(ctx, selectQuery, id).Scan(&blockedBy, &approveStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("task with id %q not found", id)
		}
		return nil, err
	}

	// 2) Проверяем, заблокирована ли задача
	if len(blockedBy) > 0 {
		return nil, fmt.Errorf("cannot mark done: task is blocked by %v", blockedBy)
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
            deadline, "dashboardID", "blockedBy", space
    `

	var task model.Task
	var blocked []string
	var space sql.NullString

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
		&blocked,
		&space,
	)
	if err != nil {
		return nil, err
	}

	task.BlockedBy = blocked
	if space.Valid {
		sv := space.String
		task.Space = &sv
	}

	return &task, nil
}

// TODO: UpdateApproveStatuses
