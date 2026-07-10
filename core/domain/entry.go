package domain

import (
	"fmt"
	"strings"
)

const DefaultURL = "#"

var ValidCategories = map[string]struct{}{
	"Clinical Psychology":       {},
	"Developmental Psychology":  {},
	"Cognitive Psychology":      {},
	"Social Psychology":         {},
	"Educational Psychology":    {},
	"Mental Health":             {},
	"Research Methods":          {},
}

var ValidEntryTypes = map[string]struct{}{
	"Journal":           {},
	"Article":           {},
	"Thesis":            {},
	"Literature Review": {},
}

type Entry struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	Category string `json:"category"`
	Year     int    `json:"year"`
	Author   string `json:"author"`
	Source   string `json:"source"`
	Type     string `json:"type"`
	URL      string `json:"url"`
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

	if _, ok := ValidCategories[normalized.Category]; !ok {
		return fmt.Errorf("invalid category")
	}

	if _, ok := ValidEntryTypes[normalized.Type]; !ok {
		return fmt.Errorf("invalid type")
	}

	return nil
}
