package service

import (
	"context"
	"errors"
	"tasker/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	dbPool *pgxpool.Pool
}

func NewUserService(dbPool *pgxpool.Pool) *UserService {
	return &UserService{dbPool: dbPool}
}

func (s *UserService) getOrCreateRoleID(ctx context.Context, roleName string) (int, error) {
	var id int
	if roleName == "" {
		roleName = "user"
	}
	// попробуем найти
	if err := s.dbPool.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1`, roleName).Scan(&id); err == nil {
		return id, nil
	} else {
		// если ошибка не pgx.ErrNoRows — вернуть ошибку
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}
	}
	// создаём роль
	if err := s.dbPool.QueryRow(ctx, `INSERT INTO roles (name) VALUES ($1) RETURNING id`, roleName).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *UserService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// определяем системную роль по строке req.Role
	roleName := req.Role
	if roleName == "" {
		roleName = "user"
	}
	roleID, err := s.getOrCreateRoleID(ctx, roleName)
	if err != nil {
		return nil, err
	}

	const query = `
        INSERT INTO users (name, surname, middlename, login, roleid, password) 
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `

	var userID int
	var middlenamePtr *string
	if req.Middlename != "" {
		middlenamePtr = &req.Middlename
	}

	if err := s.dbPool.QueryRow(ctx, query,
		req.Name,
		req.Surname,
		middlenamePtr,
		req.Login,
		roleID,
		string(hashedPassword),
	).Scan(&userID); err != nil {
		return nil, err
	}

	return &model.User{
		ID:         userID,
		Name:       req.Name,
		Surname:    req.Surname,
		Middlename: middlenamePtr,
		Login:      req.Login,
		RoleID:     roleID,
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

func (s *UserService) GetAllUsers(ctx context.Context) ([]model.User, error) {
	const query = `
		SELECT id, name, surname, middlename
		FROM users;
	`

	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Middlename,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
