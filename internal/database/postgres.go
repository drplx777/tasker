package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	if err := createTables(ctx, pool); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	log.Println("Database connected and initialized")
	return pool, nil
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		
		-- Создание таблицы Users
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			middlename TEXT,
			login TEXT NOT NULL UNIQUE,
			roleID INTEGER NOT NULL,
			password TEXT NOT NULL
		);
		
		-- Создание таблицы Tasks
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			"reporterD" TEXT NOT NULL,
			"assignerID" TEXT,
			"reviewerID" TEXT,
			"approverID" TEXT NOT NULL,
			"approveStatus" TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			"started_At" TIMESTAMPTZ,
			done_at TIMESTAMPTZ,
			deadline TEXT NOT NULL,
			"dashboardID" TEXT NOT NULL,
			"blockedBy" TEXT[]
		);
	`

	_, err := pool.Exec(ctx, query)
	return err
}
