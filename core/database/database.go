package database

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	"mindex-api/core/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

//go:embed data/seed-entries.json
var seedData []byte

func NewPool(ctx context.Context, postgresURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(postgresURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres url: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		sqlBytes, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("run migration %s: %w", name, err)
		}
		slog.Info("migration applied", "file", name)
	}

	return nil
}

func SeedIfEmpty(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM entries").Scan(&count); err != nil {
		return fmt.Errorf("count entries: %w", err)
	}

	if count > 0 {
		slog.Info("entries table already seeded", "count", count)
		return nil
	}

	var seeds []domain.EntryInput
	if err := json.Unmarshal(seedData, &seeds); err != nil {
		return fmt.Errorf("unmarshal seed data: %w", err)
	}

	for _, seed := range seeds {
		normalized := domain.NormalizeEntryInput(seed)
		if err := domain.ValidateEntryInput(normalized); err != nil {
			return fmt.Errorf("validate seed entry %q: %w", normalized.Title, err)
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO entries (title, abstract, category, year, author, source, type, url)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
			normalized.Title,
			normalized.Abstract,
			normalized.Category,
			normalized.Year,
			normalized.Author,
			normalized.Source,
			normalized.Type,
			normalized.URL,
		)
		if err != nil {
			return fmt.Errorf("insert seed entry %q: %w", normalized.Title, err)
		}
	}

	slog.Info("seeded entries", "count", len(seeds))
	return nil
}
