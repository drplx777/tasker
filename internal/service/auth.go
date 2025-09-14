package service

import (
	"context"
	"fmt"
	"tasker/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	dbPool    *pgxpool.Pool
	jwtSecret string
	spaceSvc  *SpaceService
}

func NewAuthService(dbPool *pgxpool.Pool, jwtSecret string, spaceSvc *SpaceService) *AuthService {
	return &AuthService{dbPool: dbPool, jwtSecret: jwtSecret, spaceSvc: spaceSvc}
}

func (s *AuthService) getOrCreateRoleID(ctx context.Context, roleName string) (int, error) {
	var id int
	q := `SELECT id FROM roles WHERE name = $1`
	if err := s.dbPool.QueryRow(ctx, q, roleName).Scan(&id); err == nil {
		return id, nil
	}
	// insert
	iq := `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	if err := s.dbPool.QueryRow(ctx, iq, roleName).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	roleName := "user"
	if req.Role == "spaceOwner" {
		roleName = "spaceOwner"
	}

	roleID, err := s.getOrCreateRoleID(ctx, roleName)
	if err != nil {
		return nil, err
	}

	const query = `
        INSERT INTO users (name, surname, middlename, login, roleid, password) 
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, name, surname, middlename, login, roleid
    `

	var user model.User
	err = s.dbPool.QueryRow(ctx, query,
		req.Name,
		req.Surname,
		req.Middlename,
		req.Login,
		roleID,
		string(hashedPassword),
	).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Middlename,
		&user.Login,
		&user.RoleID,
	)
	if err != nil {
		return nil, err
	}

	if req.Role == "spaceOwner" && s.spaceSvc != nil {
		spaceName := req.SpaceName
		if spaceName == "" {
			spaceName = req.Login + "'s space"
		}
		_, err := s.spaceSvc.CreateSpace(ctx, spaceName, user.ID)
		if err != nil {
			return &user, fmt.Errorf("user created but failed to create space: %w", err)
		}
	}

	user.Password = ""
	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, *model.User, time.Time, error) {
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
	exp := time.Now().Add(72 * time.Hour)
	if err != nil {
		return "", nil, exp, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, exp, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.RoleID,
		"exp":  exp.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, exp, err
	}

	user.Password = ""
	return tokenString, &user, exp, nil
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	const query = `
        SELECT id, name, surname, middlename, login, roleID
        FROM users WHERE id = $1
    `

	var user model.User
	err := s.dbPool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
		&user.Middlename,
		&user.Login,
		&user.RoleID,
	)

	return &user, err
}
