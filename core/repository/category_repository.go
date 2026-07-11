package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mindex-api/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryInUse = errors.New("category in use")
var ErrCategoryExists = errors.New("category already exists")

type CategoryRepository interface {
	List(ctx context.Context) ([]domain.Category, error)
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Create(ctx context.Context, name string) (*domain.Category, error)
	Update(ctx context.Context, id int64, name string) (*domain.Category, error)
	Delete(ctx context.Context, id int64) error
}

type PgxCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPgxCategoryRepository(pool *pgxpool.Pool) *PgxCategoryRepository {
	return &PgxCategoryRepository{pool: pool}
}

func (r *PgxCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.name,
		       (SELECT COUNT(*) FROM entries e WHERE e.category = c.name) AS entry_count
		FROM categories c
		ORDER BY c.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Category, 0)
	for rows.Next() {
		var item domain.Category
		if err := rows.Scan(&item.ID, &item.Name, &item.EntryCount); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return items, nil
}

func (r *PgxCategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	var item domain.Category
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.name,
		       (SELECT COUNT(*) FROM entries e WHERE e.category = c.name) AS entry_count
		FROM categories c
		WHERE c.id = $1
	`, id).Scan(&item.ID, &item.Name, &item.EntryCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	return &item, nil
}

func (r *PgxCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)
	`, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists category: %w", err)
	}
	return exists, nil
}

func (r *PgxCategoryRepository) Create(ctx context.Context, name string) (*domain.Category, error) {
	var item domain.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name
	`, name).Scan(&item.ID, &item.Name)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCategoryExists
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	item.EntryCount = 0
	return &item, nil
}

func (r *PgxCategoryRepository) Update(ctx context.Context, id int64, name string) (*domain.Category, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update category: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldName string
	err = tx.QueryRow(ctx, `SELECT name FROM categories WHERE id = $1`, id).Scan(&oldName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCategoryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get category name: %w", err)
	}

	var item domain.Category
	err = tx.QueryRow(ctx, `
		UPDATE categories
		SET name = $1
		WHERE id = $2
		RETURNING id, name
	`, name, id).Scan(&item.ID, &item.Name)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCategoryExists
		}
		return nil, fmt.Errorf("update category: %w", err)
	}

	if oldName != name {
		if _, err := tx.Exec(ctx, `
			UPDATE entries SET category = $1 WHERE category = $2
		`, name, oldName); err != nil {
			return nil, fmt.Errorf("rename entry categories: %w", err)
		}
	}

	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM entries WHERE category = $1
	`, name).Scan(&item.EntryCount); err != nil {
		return nil, fmt.Errorf("count category entries: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update category: %w", err)
	}
	return &item, nil
}

func (r *PgxCategoryRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete category: %w", err)
	}
	defer tx.Rollback(ctx)

	var name string
	err = tx.QueryRow(ctx, `SELECT name FROM categories WHERE id = $1`, id).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrCategoryNotFound
	}
	if err != nil {
		return fmt.Errorf("get category: %w", err)
	}

	var count int64
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM entries WHERE category = $1
	`, name).Scan(&count); err != nil {
		return fmt.Errorf("count category entries: %w", err)
	}
	if count > 0 {
		return ErrCategoryInUse
	}

	tag, err := tx.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete category: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "duplicate key")
}
