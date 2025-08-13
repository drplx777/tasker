package service

import (
	"context"
	"errors"
	"time"

	"tasker/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpaceService struct {
	dbPool *pgxpool.Pool
}

func NewSpaceService(dbPool *pgxpool.Pool) *SpaceService {
	return &SpaceService{dbPool: dbPool}
}

// CreateSpace создаёт запись в spaces и добавляет создателя в space_memberships (в транзакции).
func (s *SpaceService) CreateSpace(ctx context.Context, name string, creatorID int) (model.Space, error) {
	id := uuid.New().String()
	// Acquire соединение из пула и начинаем транзакцию
	conn, err := s.dbPool.Acquire(ctx)
	if err != nil {
		return model.Space{}, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return model.Space{}, err
	}
	// откат в случае паники/ошибки
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q1 := `INSERT INTO spaces (id, name, creator_id) VALUES ($1, $2, $3)`
	if _, err := tx.Exec(ctx, q1, id, name, creatorID); err != nil {
		return model.Space{}, err
	}

	q2 := `INSERT INTO space_memberships (space_id, user_id, role) VALUES ($1,$2,'admin')`
	if _, err := tx.Exec(ctx, q2, id, creatorID); err != nil {
		return model.Space{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Space{}, err
	}

	return model.Space{ID: id, Name: name, CreatorID: creatorID, CreatedAt: time.Now()}, nil
}

// AddMember добавляет/обновляет членство (invite / set role)
func (s *SpaceService) AddMember(ctx context.Context, spaceID string, userID int, role string) error {
	if role == "" {
		role = "member"
	}
	q := `
		INSERT INTO space_memberships (space_id, user_id, role)
		VALUES ($1,$2,$3)
		ON CONFLICT (space_id,user_id) DO UPDATE SET role = EXCLUDED.role
	`
	_, err := s.dbPool.Exec(ctx, q, spaceID, userID, role)
	return err
}

// IsMember проверяет есть ли пользователь в пространстве и возвращает роль.
func (s *SpaceService) IsMember(ctx context.Context, spaceID string, userID int) (bool, string, error) {
	var role string
	q := `SELECT role FROM space_memberships WHERE space_id=$1 AND user_id=$2`
	err := s.dbPool.QueryRow(ctx, q, spaceID, userID).Scan(&role)
	if err != nil {
		// pgx v5 возвращает pgx.ErrNoRows, сравниваем корректно
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", err
	}
	return true, role, nil
}
