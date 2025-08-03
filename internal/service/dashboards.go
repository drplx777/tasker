package service

import (
	"context"
	"tasker/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardService struct {
	dbPool *pgxpool.Pool
}

func NewDashboardService(dbPool *pgxpool.Pool) *DashboardService {
	return &DashboardService{dbPool: dbPool}
}

func (s *DashboardService) ListDashboards(ctx context.Context) ([]model.DashBoards, error) {
	const query = `
        SELECT id,
               name
        FROM dashboards;
    `

	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dashboards := []model.DashBoards{}
	for rows.Next() {
		var dashboard model.DashBoards
		err := rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
		)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}

func (s *DashboardService) GetDashboardById(ctx context.Context, id string) (*model.DashBoards, error) {
	const query = `
		SELECT id,
			   name
		FROM dashboards
		WHERE id = $1;
	`

	var dashboard model.DashBoards
	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&dashboard.ID,
		&dashboard.Name,
	)
	if err != nil {
		return nil, err
	}
	return &dashboard, nil
}

func (s *DashboardService) CreateDashboard(ctx context.Context, dashboard model.DashBoards) (*model.DashBoards, error) {
	const query = `
		INSERT INTO dashboards (name)
		VALUES ($1)
		RETURNING id, name;
	`

	var newDashboard model.DashBoards
	err := s.dbPool.QueryRow(ctx, query, dashboard.Name).Scan(
		&newDashboard.ID,
		&newDashboard.Name,
	)
	if err != nil {
		return nil, err
	}
	return &newDashboard, nil
}
