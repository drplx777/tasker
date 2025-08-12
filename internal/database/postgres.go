package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Настройка пула соединений
	config.MaxConns = 10
	config.MinConns = 2
	config.HealthCheckPeriod = 1 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute

	// Экспоненциальная задержка для повторных попыток
	var pool *pgxpool.Pool
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				break
			}
		}

		if i < maxAttempts-1 {
			delay := time.Second * time.Duration(i*2)
			log.Printf("Database connection failed (attempt %d/%d), retrying in %v: %v",
				i+1, maxAttempts, delay, err)
			time.Sleep(delay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxAttempts, err)
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

        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            surname TEXT NOT NULL,
            middlename TEXT,
            login TEXT NOT NULL UNIQUE,
            roleID INTEGER NOT NULL,
            password TEXT NOT NULL
        );

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

	if _, err := pool.Exec(ctx, query); err != nil {
		return err
	}

	// Добавляем новое поле, если его ещё нет
	if _, err := pool.Exec(ctx, `ALTER TABLE tasks ADD COLUMN IF NOT EXISTS space TEXT;`); err != nil {
		return err
	}

	return nil
}
