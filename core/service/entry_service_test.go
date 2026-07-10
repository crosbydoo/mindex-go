package service

import (
	"context"
	"errors"
	"testing"

	"mindex-api/core/domain"
	"mindex-api/core/repository"
)

func TestEntryService_List(t *testing.T) {
	repo := repository.NewEntryRepositoryMock([]domain.Entry{
		{ID: 1, Title: "Entry 1", Abstract: "A", Category: "Clinical Psychology", Year: 2024, Author: "A", Source: "S", Type: "Journal", URL: "#"},
	})
	svc := NewEntryService(repo)

	entries, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestEntryService_Create_Valid(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	entry, err := svc.Create(context.Background(), domain.EntryInput{
		Title:    "New Entry",
		Abstract: "Abstract",
		Category: "Clinical Psychology",
		Year:     2024,
		Author:   "Author",
		Source:   "Source",
		Type:     "Journal",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != 1 {
		t.Fatalf("expected id 1, got %d", entry.ID)
	}
}

func TestEntryService_Create_Invalid(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	_, err := svc.Create(context.Background(), domain.EntryInput{
		Title: "Missing fields",
	})
	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf("expected ErrInvalidPayload, got %v", err)
	}
}

func TestEntryService_Update_NotFound(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	_, err := svc.Update(context.Background(), 99, domain.EntryInput{
		Title:    "Updated",
		Abstract: "Abstract",
		Category: "Clinical Psychology",
		Year:     2024,
		Author:   "Author",
		Source:   "Source",
		Type:     "Journal",
	})
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected ErrEntryNotFound, got %v", err)
	}
}

func TestEntryService_Delete_InvalidID(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	err := svc.Delete(context.Background(), 0)
	if !errors.Is(err, ErrInvalidEntryID) {
		t.Fatalf("expected ErrInvalidEntryID, got %v", err)
	}
}

func TestLoginService_Login(t *testing.T) {
	svc := NewLoginService("admin-pass")

	token, err := svc.Login("admin-pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLoginService_InvalidPassword(t *testing.T) {
	svc := NewLoginService("admin-pass")

	_, err := svc.Login("wrong")
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLoginService_AdminNotConfigured(t *testing.T) {
	svc := NewLoginService("")

	_, err := svc.Login("anything")
	if !errors.Is(err, ErrAdminNotConfigured) {
		t.Fatalf("expected ErrAdminNotConfigured, got %v", err)
	}
}

func TestLoginService_Logout(t *testing.T) {
	svc := NewLoginService("admin-pass")

	if err := svc.Logout(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
