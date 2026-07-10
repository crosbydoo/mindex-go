package repository

import (
	"context"
	"sort"
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

func (m *EntryRepositoryMock) List(ctx context.Context, filter domain.ListFilter) ([]domain.Entry, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	filtered := make([]domain.Entry, 0)
	for _, entry := range m.entries {
		if filter.Category != "" && entry.Category != filter.Category {
			continue
		}
		filtered = append(filtered, entry)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Year == filtered[j].Year {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].Year > filtered[j].Year
	})

	total := int64(len(filtered))
	page, limit := domain.NormalizePagination(filter.Page, filter.Limit)
	offset := domain.Offset(page, limit)

	if offset >= len(filtered) {
		return []domain.Entry{}, total, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return append([]domain.Entry(nil), filtered[offset:end]...), total, nil
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
