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
	// Создаём структуры таблиц в правильном порядке
	baseSQL := `
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    CREATE TABLE IF NOT EXISTS roles (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        maindashboard INTEGER
    );

    CREATE TABLE IF NOT EXISTS dashboards (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name TEXT NOT NULL,
        surname TEXT NOT NULL,
        middlename TEXT,
        login TEXT NOT NULL UNIQUE,
        roleid INTEGER NOT NULL,
        password TEXT NOT NULL,
        token TEXT,
        spaces TEXT[]    -- временно храним, если есть старые данные
    );

    -- spaces нужно создать ДО space_memberships, т.к. у latter есть FK на spaces
    CREATE TABLE IF NOT EXISTS spaces (
        id TEXT PRIMARY KEY DEFAULT (uuid_generate_v4()::text),
        name TEXT NOT NULL,
        creator_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );

    CREATE TABLE IF NOT EXISTS space_memberships (
        space_id TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
        user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        role TEXT NOT NULL DEFAULT 'member',
        joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
        PRIMARY KEY (space_id, user_id)
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
        "blockedBy" TEXT[],
        space TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_users_roleid ON users(roleid);
    CREATE INDEX IF NOT EXISTS idx_tasks_dashboardid ON tasks("dashboardID");
    CREATE INDEX IF NOT EXISTS idx_tasks_space ON tasks(space);
    CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at DESC);
    `

	if _, err := pool.Exec(ctx, baseSQL); err != nil {
		return fmt.Errorf("create base tables: %w", err)
	}

	// Если есть users.spaces (старые данные), мигрируем их в spaces и space_memberships
	var hasSpacesColumn bool
	err := pool.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'spaces'
        )
    `).Scan(&hasSpacesColumn)
	if err != nil {
		return fmt.Errorf("check spaces column: %w", err)
	}

	if hasSpacesColumn {
		// 1) Вставляем отсутствующие записи в spaces (id = s_id, name = s_id, creator = user)
		//    чтобы FK на space_memberships не падал, если ранее у пользователей были ссылки на пространства, но записей в spaces не было.
		insertMissingSpaces := `
        INSERT INTO spaces (id, name, creator_id)
        SELECT DISTINCT s_id, s_id, u.id
        FROM users u, unnest(u.spaces) AS s_id
        WHERE NOT EXISTS (SELECT 1 FROM spaces sp WHERE sp.id = s_id)
        `
		if _, err := pool.Exec(ctx, insertMissingSpaces); err != nil {
			return fmt.Errorf("insert missing spaces: %w", err)
		}

		// 2) Мигрируем memberships
		insertMemberships := `
        INSERT INTO space_memberships (space_id, user_id, role)
        SELECT s_id, u.id, 'member'
        FROM users u, unnest(u.spaces) AS s_id
        ON CONFLICT (space_id, user_id) DO NOTHING
        `
		if _, err := pool.Exec(ctx, insertMemberships); err != nil {
			return fmt.Errorf("migrate memberships: %w", err)
		}

		// 3) По желанию — удалить колонку users.spaces (закомментировано; раскомментируй когда будешь уверен)
		// if _, err := pool.Exec(ctx, `ALTER TABLE users DROP COLUMN IF EXISTS spaces;`); err != nil {
		//     return fmt.Errorf("drop users.spaces: %w", err)
		// }
	}

	return nil
}
