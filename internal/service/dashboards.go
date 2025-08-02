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
	const query = ``

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
