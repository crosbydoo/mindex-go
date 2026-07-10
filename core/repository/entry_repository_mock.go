package repository

import (
	"context"
	"sync"

	"mindex-api/core/domain"
)

type EntryRepositoryMock struct {
	mu      sync.RWMutex
	entries []domain.Entry
	nextID  int64
}

func NewEntryRepositoryMock(entries []domain.Entry) *EntryRepositoryMock {
	nextID := int64(1)
	for _, entry := range entries {
		if entry.ID >= nextID {
			nextID = entry.ID + 1
		}
	}
	return &EntryRepositoryMock{
		entries: append([]domain.Entry(nil), entries...),
		nextID:  nextID,
	}
}

func (m *EntryRepositoryMock) List(ctx context.Context) ([]domain.Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := append([]domain.Entry(nil), m.entries...)
	return result, nil
}

func (m *EntryRepositoryMock) Create(ctx context.Context, input domain.EntryInput) (*domain.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	normalized := domain.NormalizeEntryInput(input)
	entry := domain.Entry{
		ID:       m.nextID,
		Title:    normalized.Title,
		Abstract: normalized.Abstract,
		Category: normalized.Category,
		Year:     normalized.Year,
		Author:   normalized.Author,
		Source:   normalized.Source,
		Type:     normalized.Type,
		URL:      normalized.URL,
	}
	m.nextID++
	m.entries = append(m.entries, entry)
	return &entry, nil
}

func (m *EntryRepositoryMock) Update(ctx context.Context, id int64, input domain.EntryInput) (*domain.Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	normalized := domain.NormalizeEntryInput(input)
	for i, entry := range m.entries {
		if entry.ID != id {
			continue
		}
		updated := domain.Entry{
			ID:       id,
			Title:    normalized.Title,
			Abstract: normalized.Abstract,
			Category: normalized.Category,
			Year:     normalized.Year,
			Author:   normalized.Author,
			Source:   normalized.Source,
			Type:     normalized.Type,
			URL:      normalized.URL,
		}
		m.entries[i] = updated
		return &updated, nil
	}
	return nil, ErrNotFound
}

func (m *EntryRepositoryMock) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, entry := range m.entries {
		if entry.ID != id {
			continue
		}
		m.entries = append(m.entries[:i], m.entries[i+1:]...)
		return nil
	}
	return ErrNotFound
}
