package service

import (
	"context"
	"tasker/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	dbPool *pgxpool.Pool
}

func NewUserService(dbPool *pgxpool.Pool) *UserService {
	return &UserService{dbPool: dbPool}
}

func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	const query = `
        INSERT INTO users (name, surname, middlename, login, roleID, password) 
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `

	var userID int
	err = s.dbPool.QueryRow(ctx, query,
		req.Name,
		req.Surname,
		req.Middlename,
		req.Login,
		req.RoleID,
		string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:         userID,
		Name:       req.Name,
		Surname:    req.Surname,
		Middlename: &req.Middlename,
		Login:      req.Login,
		RoleID:     req.RoleID,
	}, nil
}

func (s *UserService) Login(ctx context.Context, login, password string) (*model.User, error) {
	const query = `
        SELECT id, name, surname, middlename, login, roleID, password 
        FROM users WHERE login = $1
    `

	var user model.User
	err := s.dbPool.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Middlename,
		&user.Login,
		&user.RoleID,
		&user.Password,
	)

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	user.Password = ""
	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*model.User, error) {
	const query = `
        SELECT id, name, surname, middlename, login, roleID
        FROM users WHERE id = $1
    `

	var user model.User
	err := s.dbPool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Middlename,
		&user.Login,
		&user.RoleID,
	)

	return &user, err
}
