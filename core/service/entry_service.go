package service

import (
	"context"
	"errors"
	"fmt"

	"mindex-api/core/auth"
	"mindex-api/core/domain"
	"mindex-api/core/repository"
)

var (
	ErrInvalidPayload   = errors.New("invalid entry payload")
	ErrInvalidEntryID   = errors.New("invalid entry id")
	ErrEntryNotFound    = errors.New("entry not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrAdminNotConfigured = errors.New("admin password not configured")
)

type EntryService interface {
	List(ctx context.Context) ([]domain.Entry, error)
	Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error)
	Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error)
	Delete(ctx context.Context, id int64) error
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

func (s *entryService) List(ctx context.Context) ([]domain.Entry, error) {
	entries, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	if entries == nil {
		entries = []domain.Entry{}
	}
	return entries, nil
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
