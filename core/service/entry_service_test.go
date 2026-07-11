package service

import (
	"context"
	"errors"
	"testing"

	"mindex-api/core/domain"
	"mindex-api/core/repository"
)

func sampleEntries() []domain.Entry {
	return []domain.Entry{
		{ID: 1, Title: "Entry 1", Abstract: "A", Category: "Clinical Psychology", Year: 2024, Author: "A", Source: "S", Type: "Journal", URL: "#"},
		{ID: 2, Title: "Entry 2", Abstract: "B", Category: "Clinical Psychology", Year: 2023, Author: "B", Source: "S", Type: "Article", URL: "#"},
		{ID: 3, Title: "Entry 3", Abstract: "C", Category: "Mental Health", Year: 2022, Author: "C", Source: "S", Type: "Journal", URL: "#"},
	}
}

func TestEntryService_List(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(sampleEntries())
	svc := NewEntryService(repo)

	result, err := svc.List(context.Background(), domain.ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result.Items))
	}
	if result.Pagination.Total != 3 {
		t.Fatalf("expected total 3, got %d", result.Pagination.Total)
	}
}

func TestEntryService_List_WithCategory(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(sampleEntries())
	svc := NewEntryService(repo)

	result, err := svc.List(context.Background(), domain.ListFilter{
		Page:     1,
		Limit:    10,
		Category: "Clinical Psychology",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Items))
	}
}

func TestEntryService_List_InvalidCategory(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	_, err := svc.List(context.Background(), domain.ListFilter{
		Category: "Unknown",
	})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("expected ErrInvalidCategory, got %v", err)
	}
}

func TestEntryService_ListByCategories(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(sampleEntries())
	svc := NewEntryService(repo)

	result, err := svc.ListByCategories(context.Background(), 1, 10, domain.ArchiveActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Categories) != len(domain.CategoryList) {
		t.Fatalf("expected %d categories, got %d", len(domain.CategoryList), len(result.Categories))
	}

	var clinical *domain.CategoryEntries
	for i := range result.Categories {
		if result.Categories[i].Category == "Clinical Psychology" {
			clinical = &result.Categories[i]
			break
		}
	}
	if clinical == nil {
		t.Fatal("expected Clinical Psychology category")
	}
	if clinical.Pagination.Total != 2 {
		t.Fatalf("expected clinical total 2, got %d", clinical.Pagination.Total)
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

func TestEntryService_ArchiveAndUnarchive(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(sampleEntries())
	svc := NewEntryService(repo)

	archived, err := svc.Archive(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !archived.IsArchived {
		t.Fatal("expected entry to be archived")
	}

	active, err := svc.List(context.Background(), domain.ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(active.Items) != 2 {
		t.Fatalf("expected 2 active entries, got %d", len(active.Items))
	}

	onlyArchived, err := svc.List(context.Background(), domain.ListFilter{
		Page:     1,
		Limit:    10,
		Archived: domain.ArchiveOnly,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(onlyArchived.Items) != 1 {
		t.Fatalf("expected 1 archived entry, got %d", len(onlyArchived.Items))
	}

	restored, err := svc.Unarchive(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if restored.IsArchived {
		t.Fatal("expected entry to be unarchived")
	}
}

func TestEntryService_Archive_NotFound(t *testing.T) {
	repo := repository.NewEntryRepositoryMock(nil)
	svc := NewEntryService(repo)

	_, err := svc.Archive(context.Background(), 99)
	if !errors.Is(err, ErrEntryNotFound) {
		t.Fatalf("expected ErrEntryNotFound, got %v", err)
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
