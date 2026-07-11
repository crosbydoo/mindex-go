package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"mindex-api/core/auth"
	"mindex-api/core/domain"
	"mindex-api/core/repository"
)

var (
	ErrInvalidPayload     = errors.New("invalid entry payload")
	ErrInvalidEntryID     = errors.New("invalid entry id")
	ErrEntryNotFound      = errors.New("entry not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrAdminNotConfigured = errors.New("admin password not configured")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrInvalidPagination  = errors.New("invalid pagination")
	ErrInvalidArchived    = errors.New("invalid archived filter")
)

type EntryService interface {
	List(ctx context.Context, filter domain.ListFilter) (*domain.PaginatedEntries, error)
	ListByCategories(ctx context.Context, page, limit int, archived domain.ArchiveScope) (*domain.CategoriesResult, error)
	Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error)
	Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error)
	Delete(ctx context.Context, id int64) error
	Archive(ctx context.Context, id int64) (*domain.Entry, error)
	Unarchive(ctx context.Context, id int64) (*domain.Entry, error)
	ArchiveMany(ctx context.Context, ids []int64) (*domain.BulkResult, error)
	UnarchiveMany(ctx context.Context, ids []int64) (*domain.BulkResult, error)
	DeleteMany(ctx context.Context, ids []int64) (*domain.BulkResult, error)
}

type LoginService interface {
	Login(password string) (string, error)
	Logout() error
}

type entryService struct {
	repo repository.EntryRepository
}

func NewEntryService(repo repository.EntryRepository) EntryService {
	return &entryService{repo: repo}
}

func (s *entryService) List(ctx context.Context, filter domain.ListFilter) (*domain.PaginatedEntries, error) {
	filter.Category = strings.TrimSpace(filter.Category)
	if filter.Category != "" && !domain.IsValidCategory(filter.Category) {
		return nil, ErrInvalidCategory
	}
	if filter.Archived == "" {
		filter.Archived = domain.ArchiveActive
	}

	filter.Page, filter.Limit = domain.NormalizePagination(filter.Page, filter.Limit)

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	if items == nil {
		items = []domain.Entry{}
	}

	return &domain.PaginatedEntries{
		Items:      items,
		Pagination: domain.BuildPagination(filter.Page, filter.Limit, total),
	}, nil
}

func (s *entryService) ListByCategories(ctx context.Context, page, limit int, archived domain.ArchiveScope) (*domain.CategoriesResult, error) {
	page, limit = domain.NormalizePagination(page, limit)
	if archived == "" {
		archived = domain.ArchiveActive
	}

	categories := make([]domain.CategoryEntries, 0, len(domain.CategoryList))
	for _, category := range domain.CategoryList {
		items, total, err := s.repo.List(ctx, domain.ListFilter{
			Page:     page,
			Limit:    limit,
			Category: category,
			Archived: archived,
		})
		if err != nil {
			return nil, fmt.Errorf("list entries for category %q: %w", category, err)
		}
		if items == nil {
			items = []domain.Entry{}
		}

		categories = append(categories, domain.CategoryEntries{
			Category:   category,
			Items:      items,
			Pagination: domain.BuildPagination(page, limit, total),
		})
	}

	return &domain.CategoriesResult{Categories: categories}, nil
}

func (s *entryService) Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error) {
	if err := domain.ValidateEntryInput(input); err != nil {
		return nil, ErrInvalidPayload
	}
	return s.repo.Create(ctx, input)
}

func (s *entryService) Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error) {
	if id <= 0 {
		return nil, ErrInvalidEntryID
	}
	if err := domain.ValidateEntryInput(input); err != nil {
		return nil, ErrInvalidPayload
	}

	entry, err := s.repo.Update(ctx, id, input)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update entry: %w", err)
	}
	return entry, nil
}

func (s *entryService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidEntryID
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrEntryNotFound
	}
	if err != nil {
		return fmt.Errorf("delete entry: %w", err)
	}
	return nil
}

func (s *entryService) Archive(ctx context.Context, id int64) (*domain.Entry, error) {
	return s.setArchived(ctx, id, true)
}

func (s *entryService) Unarchive(ctx context.Context, id int64) (*domain.Entry, error) {
	return s.setArchived(ctx, id, false)
}

func (s *entryService) ArchiveMany(ctx context.Context, ids []int64) (*domain.BulkResult, error) {
	return s.setArchivedMany(ctx, ids, true)
}

func (s *entryService) UnarchiveMany(ctx context.Context, ids []int64) (*domain.BulkResult, error) {
	return s.setArchivedMany(ctx, ids, false)
}

func (s *entryService) DeleteMany(ctx context.Context, ids []int64) (*domain.BulkResult, error) {
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return nil, ErrInvalidEntryID
	}

	affected, err := s.repo.DeleteMany(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("delete many: %w", err)
	}
	return &domain.BulkResult{Affected: int(affected), IDs: ids}, nil
}

func (s *entryService) setArchived(ctx context.Context, id int64, archived bool) (*domain.Entry, error) {
	if id <= 0 {
		return nil, ErrInvalidEntryID
	}

	entry, err := s.repo.SetArchived(ctx, id, archived)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("set archived: %w", err)
	}
	return entry, nil
}

func (s *entryService) setArchivedMany(ctx context.Context, ids []int64, archived bool) (*domain.BulkResult, error) {
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return nil, ErrInvalidEntryID
	}

	affected, err := s.repo.SetArchivedMany(ctx, ids, archived)
	if err != nil {
		return nil, fmt.Errorf("set archived many: %w", err)
	}
	return &domain.BulkResult{Affected: int(affected), IDs: ids}, nil
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

type loginService struct {
	adminPassword string
}

func NewLoginService(adminPassword string) LoginService {
	return &loginService{adminPassword: adminPassword}
}

func (s *loginService) Login(password string) (string, error) {
	if s.adminPassword == "" {
		return "", ErrAdminNotConfigured
	}
	if !auth.VerifyPassword(password, s.adminPassword) {
		return "", ErrInvalidPassword
	}
	return auth.GenerateToken(s.adminPassword), nil
}

func (s *loginService) Logout() error {
	return nil
}
