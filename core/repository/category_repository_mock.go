package repository

import (
	"context"
	"sync"

	"mindex-api/core/domain"
)

type CategoryRepositoryMock struct {
	mu         sync.RWMutex
	categories []domain.Category
	nextID     int64
	entryRepo  *EntryRepositoryMock
}

func NewCategoryRepositoryMock(names []string, entryRepo *EntryRepositoryMock) *CategoryRepositoryMock {
	categories := make([]domain.Category, 0, len(names))
	for i, name := range names {
		categories = append(categories, domain.Category{
			ID:   int64(i + 1),
			Name: name,
		})
	}
	return &CategoryRepositoryMock{
		categories: categories,
		nextID:     int64(len(names) + 1),
		entryRepo:  entryRepo,
	}
}

func (m *CategoryRepositoryMock) List(ctx context.Context) ([]domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]domain.Category, 0, len(m.categories))
	for _, item := range m.categories {
		item.EntryCount = m.countEntries(item.Name)
		out = append(out, item)
	}
	return out, nil
}

func (m *CategoryRepositoryMock) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.categories {
		if item.ID == id {
			item.EntryCount = m.countEntries(item.Name)
			return &item, nil
		}
	}
	return nil, ErrCategoryNotFound
}

func (m *CategoryRepositoryMock) ExistsByName(ctx context.Context, name string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.categories {
		if item.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *CategoryRepositoryMock) Create(ctx context.Context, name string) (*domain.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range m.categories {
		if item.Name == name {
			return nil, ErrCategoryExists
		}
	}

	item := domain.Category{ID: m.nextID, Name: name, EntryCount: 0}
	m.nextID++
	m.categories = append(m.categories, item)
	return &item, nil
}

func (m *CategoryRepositoryMock) Update(ctx context.Context, id int64, name string) (*domain.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range m.categories {
		if item.ID != id && item.Name == name {
			return nil, ErrCategoryExists
		}
	}

	for i, item := range m.categories {
		if item.ID != id {
			continue
		}
		oldName := item.Name
		item.Name = name
		m.categories[i] = item
		if m.entryRepo != nil && oldName != name {
			m.entryRepo.renameCategory(oldName, name)
		}
		item.EntryCount = m.countEntries(name)
		return &item, nil
	}
	return nil, ErrCategoryNotFound
}

func (m *CategoryRepositoryMock) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.categories {
		if item.ID != id {
			continue
		}
		if m.countEntries(item.Name) > 0 {
			return ErrCategoryInUse
		}
		m.categories = append(m.categories[:i], m.categories[i+1:]...)
		return nil
	}
	return ErrCategoryNotFound
}

func (m *CategoryRepositoryMock) countEntries(name string) int64 {
	if m.entryRepo == nil {
		return 0
	}
	m.entryRepo.mu.RLock()
	defer m.entryRepo.mu.RUnlock()

	var count int64
	for _, entry := range m.entryRepo.entries {
		if entry.Category == name {
			count++
		}
	}
	return count
}

func (m *EntryRepositoryMock) renameCategory(oldName, newName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, entry := range m.entries {
		if entry.Category == oldName {
			entry.Category = newName
			m.entries[i] = entry
		}
	}
}
