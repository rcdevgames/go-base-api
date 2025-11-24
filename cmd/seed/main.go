package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
	"github.com/rcdevgames/modular-monolith-clean/internal/infrastructure/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("database connection: %v", err)
	}
	defer db.Close()

	if err := runSeeds(context.Background(), db, "db/seeds"); err != nil {
		log.Fatalf("seed data: %v", err)
	}

	log.Println("seeding completed successfully")
}

func runSeeds(ctx context.Context, db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		log.Printf("applying seed %s", entry.Name())
		if err := execSQLBlock(ctx, db, string(content)); err != nil {
			return err
		}
	}
	return nil
}

func execSQLBlock(ctx context.Context, db *sql.DB, sqlBlock string) error {
	statements := strings.Split(sqlBlock, ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, trimmed); err != nil {
			return err
		}
	}
	return nil
}
