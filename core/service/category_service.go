package service

import (
	"context"
	"errors"
	"fmt"

	"mindex-api/core/domain"
	"mindex-api/core/repository"
)

var (
	ErrInvalidCategoryPayload = errors.New("invalid category payload")
	ErrInvalidCategoryID      = errors.New("invalid category id")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryInUse          = errors.New("category in use")
	ErrCategoryExists         = errors.New("category already exists")
)

type CategoryService interface {
	List(ctx context.Context) (*domain.CategoryListResult, error)
	Create(ctx context.Context, input domain.CategoryInput) (*domain.Category, error)
	Update(ctx context.Context, id int64, input domain.CategoryInput) (*domain.Category, error)
	Delete(ctx context.Context, id int64) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) List(ctx context.Context) (*domain.CategoryListResult, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	if items == nil {
		items = []domain.Category{}
	}
	return &domain.CategoryListResult{Items: items}, nil
}

func (s *categoryService) Create(ctx context.Context, input domain.CategoryInput) (*domain.Category, error) {
	if err := domain.ValidateCategoryInput(input); err != nil {
		return nil, ErrInvalidCategoryPayload
	}
	normalized := domain.NormalizeCategoryInput(input)

	item, err := s.repo.Create(ctx, normalized.Name)
	if errors.Is(err, repository.ErrCategoryExists) {
		return nil, ErrCategoryExists
	}
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return item, nil
}

func (s *categoryService) Update(ctx context.Context, id int64, input domain.CategoryInput) (*domain.Category, error) {
	if id <= 0 {
		return nil, ErrInvalidCategoryID
	}
	if err := domain.ValidateCategoryInput(input); err != nil {
		return nil, ErrInvalidCategoryPayload
	}
	normalized := domain.NormalizeCategoryInput(input)

	item, err := s.repo.Update(ctx, id, normalized.Name)
	if errors.Is(err, repository.ErrCategoryNotFound) {
		return nil, ErrCategoryNotFound
	}
	if errors.Is(err, repository.ErrCategoryExists) {
		return nil, ErrCategoryExists
	}
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return item, nil
}

func (s *categoryService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidCategoryID
	}

	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrCategoryNotFound) {
		return ErrCategoryNotFound
	}
	if errors.Is(err, repository.ErrCategoryInUse) {
		return ErrCategoryInUse
	}
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}
