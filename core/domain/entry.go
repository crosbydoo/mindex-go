package domain

import (
	"fmt"
	"strings"
)

const (
	DefaultURL   = "#"
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

var ValidEntryTypes = map[string]struct{}{
	"Journal":           {},
	"Article":           {},
	"Thesis":            {},
	"Literature Review": {},
}

type ArchiveScope string

const (
	ArchiveActive ArchiveScope = "active" // default: is_archived = false
	ArchiveOnly   ArchiveScope = "archived"
	ArchiveAll    ArchiveScope = "all"
)

type Entry struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Abstract   string `json:"abstract"`
	Category   string `json:"category"`
	Year       int    `json:"year"`
	Author     string `json:"author"`
	Source     string `json:"source"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	IsArchived bool   `json:"is_archived"`
}

type EntryIDsInput struct {
	IDs []int64 `json:"ids"`
}

type BulkResult struct {
	Affected int     `json:"affected"`
	IDs      []int64 `json:"ids"`
}

func ParseArchiveScope(raw string) (ArchiveScope, error) {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "", "false", "active":
		return ArchiveActive, nil
	case "true", "archived":
		return ArchiveOnly, nil
	case "all":
		return ArchiveAll, nil
	default:
		return "", fmt.Errorf("invalid archived filter")
	}
}

type EntryInput struct {
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Category string `json:"category"`
	Year     int    `json:"year"`
	Author   string `json:"author"`
	Source   string `json:"source"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type PaginatedEntries struct {
	Items      []Entry    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type CategoryEntries struct {
	Category   string     `json:"category"`
	Items      []Entry    `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type CategoriesResult struct {
	Categories []CategoryEntries `json:"categories"`
}

type ListFilter struct {
	Page     int
	Limit    int
	Category string
	Archived ArchiveScope
}

func NormalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return page, limit
}

func BuildPagination(page, limit int, total int64) Pagination {
	page, limit = NormalizePagination(page, limit)
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	if total == 0 {
		totalPages = 0
	}
	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    totalPages > 0 && page < totalPages,
		HasPrev:    totalPages > 0 && page > 1,
	}
}

func Offset(page, limit int) int {
	page, limit = NormalizePagination(page, limit)
	return (page - 1) * limit
}

func IsValidCategory(category string) bool {
	name := strings.TrimSpace(category)
	for _, item := range CategoryList {
		if item == name {
			return true
		}
	}
	return false
}

func NormalizeEntryInput(input EntryInput) EntryInput {
	return EntryInput{
		Title:    strings.TrimSpace(input.Title),
		Abstract: strings.TrimSpace(input.Abstract),
		Category: strings.TrimSpace(input.Category),
		Year:     input.Year,
		Author:   strings.TrimSpace(input.Author),
		Source:   strings.TrimSpace(input.Source),
		Type:     strings.TrimSpace(input.Type),
		URL:      normalizeURL(input.URL),
	}
}

func normalizeURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return DefaultURL
	}
	return trimmed
}

func ValidateEntryInput(input EntryInput) error {
	normalized := NormalizeEntryInput(input)

	if normalized.Title == "" ||
		normalized.Abstract == "" ||
		normalized.Category == "" ||
		normalized.Author == "" ||
		normalized.Source == "" ||
		normalized.Type == "" ||
		normalized.Year <= 0 {
		return fmt.Errorf("required fields must be non-empty")
	}

	if _, ok := ValidEntryTypes[normalized.Type]; !ok {
		return fmt.Errorf("invalid type")
	}

	return nil
}
