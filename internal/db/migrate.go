package db

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	// Create migrations tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Read available migrations
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()

		// Check if already applied
		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
			name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		// Read and execute migration
		sql, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		log.Printf("applying migration: %s", name)
		_, err = pool.Exec(ctx, string(sql))
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		_, err = pool.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			name,
		)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}

	return nil
}
