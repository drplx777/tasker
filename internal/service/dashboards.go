// файл: service/dashboards.go
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
	const query = `SELECT id, name FROM dashboards;`
	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dashboards []model.DashBoards
	for rows.Next() {
		var d model.DashBoards
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		dashboards = append(dashboards, d)
	}
	return dashboards, nil
}
func (s *DashboardService) CreateDashboard(ctx context.Context, dashboard model.DashBoards) (*model.DashBoards, error) {
	const query = `INSERT INTO dashboards (name) VALUES ($1) RETURNING id, name;`
	var d model.DashBoards
	if err := s.dbPool.QueryRow(ctx, query, dashboard.Name).Scan(&d.ID, &d.Name); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *DashboardService) GetDashboardById(ctx context.Context, id string) (*model.DashBoards, error) {
	const query = `SELECT id, name FROM dashboards WHERE id = $1;`
	var d model.DashBoards
	if err := s.dbPool.QueryRow(ctx, query, id).Scan(&d.ID, &d.Name); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *DashboardService) CreateDashboardForSpace(ctx context.Context, spaceID string, name string) (*model.DashBoards, error) {
	const q = `INSERT INTO dashboards (name, space_id) VALUES ($1,$2) RETURNING id, name;`
	var d model.DashBoards
	if err := s.dbPool.QueryRow(ctx, q, name, spaceID).Scan(&d.ID, &d.Name); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *DashboardService) DeleteDashboardFromSpace(ctx context.Context, spaceID string, dashboardID string) error {
	// Удаляем только если dashboard принадлежит пространству
	const q = `DELETE FROM dashboards WHERE id = $1 AND space_id = $2`
	_, err := s.dbPool.Exec(ctx, q, dashboardID, spaceID)
	return err
}

func (s *DashboardService) ListDashboardsBySpace(ctx context.Context, spaceID string) ([]model.DashBoards, error) {
	const q = `SELECT id, name FROM dashboards WHERE space_id = $1`
	rows, err := s.dbPool.Query(ctx, q, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.DashBoards
	for rows.Next() {
		var d model.DashBoards
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}
