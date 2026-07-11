package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mindex-api/core/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("entry not found")

const entryColumns = `id, title, abstract, category, year, author, source, type, url, is_archived`

type EntryRepository interface {
	List(ctx context.Context, filter domain.ListFilter) ([]domain.Entry, int64, error)
	Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error)
	Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error)
	Delete(ctx context.Context, id int64) error
	SetArchived(ctx context.Context, id int64, archived bool) (*domain.Entry, error)
}

type PgxEntryRepository struct {
	pool *pgxpool.Pool
}

func NewPgxEntryRepository(pool *pgxpool.Pool) *PgxEntryRepository {
	return &PgxEntryRepository{pool: pool}
}

func (r *PgxEntryRepository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Entry, int64, error) {
	page, limit := domain.NormalizePagination(filter.Page, filter.Limit)
	offset := domain.Offset(page, limit)

	args := make([]any, 0, 4)
	clauses := make([]string, 0, 2)
	argPos := 1

	if filter.Category != "" {
		clauses = append(clauses, fmt.Sprintf("category = $%d", argPos))
		args = append(args, filter.Category)
		argPos++
	}

	switch filter.Archived {
	case domain.ArchiveOnly:
		clauses = append(clauses, "is_archived = TRUE")
	case domain.ArchiveAll:
		// no archive filter
	default:
		clauses = append(clauses, "is_archived = FALSE")
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM entries " + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count entries: %w", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s
		FROM entries
		%s
		ORDER BY year DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, entryColumns, where, argPos, argPos+1)

	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	entries := make([]domain.Entry, 0)
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate entries: %w", err)
	}

	return entries, total, nil
}

func (r *PgxEntryRepository) Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error) {
	normalized := domain.NormalizeEntryInput(input)

	var entry domain.Entry
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO entries (title, abstract, category, year, author, source, type, url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING %s
	`, entryColumns),
		normalized.Title,
		normalized.Abstract,
		normalized.Category,
		normalized.Year,
		normalized.Author,
		normalized.Source,
		normalized.Type,
		normalized.URL,
	).Scan(
		&entry.ID,
		&entry.Title,
		&entry.Abstract,
		&entry.Category,
		&entry.Year,
		&entry.Author,
		&entry.Source,
		&entry.Type,
		&entry.URL,
		&entry.IsArchived,
	)
	if err != nil {
		return nil, fmt.Errorf("insert entry: %w", err)
	}

	return &entry, nil
}

func (r *PgxEntryRepository) Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error) {
	normalized := domain.NormalizeEntryInput(input)

	var entry domain.Entry
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		UPDATE entries
		SET title = $1, abstract = $2, category = $3, year = $4,
		    author = $5, source = $6, type = $7, url = $8
		WHERE id = $9
		RETURNING %s
	`, entryColumns),
		normalized.Title,
		normalized.Abstract,
		normalized.Category,
		normalized.Year,
		normalized.Author,
		normalized.Source,
		normalized.Type,
		normalized.URL,
		id,
	).Scan(
		&entry.ID,
		&entry.Title,
		&entry.Abstract,
		&entry.Category,
		&entry.Year,
		&entry.Author,
		&entry.Source,
		&entry.Type,
		&entry.URL,
		&entry.IsArchived,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update entry: %w", err)
	}

	return &entry, nil
}

func (r *PgxEntryRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM entries WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgxEntryRepository) SetArchived(ctx context.Context, id int64, archived bool) (*domain.Entry, error) {
	var entry domain.Entry
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		UPDATE entries
		SET is_archived = $1,
		    archived_at = CASE WHEN $1 THEN NOW() ELSE NULL END
		WHERE id = $2
		RETURNING %s
	`, entryColumns), archived, id).Scan(
		&entry.ID,
		&entry.Title,
		&entry.Abstract,
		&entry.Category,
		&entry.Year,
		&entry.Author,
		&entry.Source,
		&entry.Type,
		&entry.URL,
		&entry.IsArchived,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("set archived: %w", err)
	}
	return &entry, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanEntry(row scannable) (domain.Entry, error) {
	var entry domain.Entry
	if err := row.Scan(
		&entry.ID,
		&entry.Title,
		&entry.Abstract,
		&entry.Category,
		&entry.Year,
		&entry.Author,
		&entry.Source,
		&entry.Type,
		&entry.URL,
		&entry.IsArchived,
	); err != nil {
		return domain.Entry{}, fmt.Errorf("scan entry: %w", err)
	}
	return entry, nil
}
