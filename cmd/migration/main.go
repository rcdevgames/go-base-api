package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
	"github.com/rcdevgames/modular-monolith-clean/internal/infrastructure/database"
)

type Migration struct {
	Name    string
	UpSQL   string
	DownSQL string
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: go run ./cmd/migration <up|down>")
	}
	command := strings.ToLower(os.Args[1])
	if command != "up" && command != "down" {
		log.Fatalf("unknown command %q: expected up or down", command)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("database connection: %v", err)
	}
	defer db.Close()

	migrations, err := loadMigrations("db/migrations")
	if err != nil {
		log.Fatalf("load migrations: %v", err)
	}

	ctx := context.Background()
	if err := ensureSchemaMigrationsTable(ctx, db); err != nil {
		log.Fatalf("ensure schema_migrations: %v", err)
	}

	switch command {
	case "up":
		if err := runMigrationsUp(ctx, db, migrations); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
	case "down":
		if err := runMigrationsDown(ctx, db, migrations); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
	}

	log.Println("migration completed successfully")
}

func loadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		upSQL, downSQL, err := parseMigration(string(content))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		migrations = append(migrations, Migration{
			Name:    entry.Name(),
			UpSQL:   upSQL,
			DownSQL: downSQL,
		})
	}
	return migrations, nil
}

func parseMigration(content string) (string, string, error) {
	const upMarker = "-- +migrate Up"
	const downMarker = "-- +migrate Down"

	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return "", "", errors.New("missing Up marker")
	}
	downIdx := strings.Index(content, downMarker)
	if downIdx == -1 {
		return "", "", errors.New("missing Down marker")
	}
	upSQL := strings.TrimSpace(content[upIdx+len(upMarker) : downIdx])
	downSQL := strings.TrimSpace(content[downIdx+len(downMarker):])
	return upSQL, downSQL, nil
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	_, err := db.ExecContext(ctx, query)
	return err
}

func runMigrationsUp(ctx context.Context, db *sql.DB, migrations []Migration) error {
	for _, migration := range migrations {
		applied, err := isMigrationApplied(ctx, db, migration.Name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		log.Printf("applying %s", migration.Name)
		if err := applyMigrationUp(ctx, db, migration); err != nil {
			return err
		}
	}
	return nil
}

func runMigrationsDown(ctx context.Context, db *sql.DB, migrations []Migration) error {
	nameToMigration := make(map[string]Migration, len(migrations))
	for _, m := range migrations {
		nameToMigration[m.Name] = m
	}

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}
	for _, name := range applied {
		migration, ok := nameToMigration[name]
		if !ok {
			log.Printf("skipping unknown applied migration %s", name)
			continue
		}
		log.Printf("reverting %s", name)
		if err := applyMigrationDown(ctx, db, migration); err != nil {
			return err
		}
	}
	return nil
}

func applyMigrationUp(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := execStatements(ctx, tx, migration.UpSQL); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(name, applied_at) VALUES ($1, $2)`, migration.Name, time.Now()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func applyMigrationDown(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := execStatements(ctx, tx, migration.DownSQL); err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE name = $1`, migration.Name); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func execStatements(ctx context.Context, tx *sql.Tx, sqlBlock string) error {
	statements := strings.Split(sqlBlock, ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, trimmed); err != nil {
			return err
		}
	}
	return nil
}

func isMigrationApplied(ctx context.Context, db *sql.DB, name string) (bool, error) {
	const query = `SELECT 1 FROM schema_migrations WHERE name = $1`
	var dummy int
	err := db.QueryRowContext(ctx, query, name).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) ([]string, error) {
	const query = `SELECT name FROM schema_migrations ORDER BY applied_at DESC`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
