package domain

import (
	"fmt"
	"strings"
)

// DefaultCategoryNames are seeded into the categories table on first migration.
var DefaultCategoryNames = []string{
	"Clinical Psychology",
	"Developmental Psychology",
	"Cognitive Psychology",
	"Social Psychology",
	"Educational Psychology",
	"Mental Health",
	"Research Methods",
}

// CategoryList is kept for backward-compatible tests/docs defaults.
// Runtime category lists come from the database.
var CategoryList = DefaultCategoryNames

type Category struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	EntryCount int64  `json:"entry_count"`
}

type CategoryInput struct {
	Name string `json:"name"`
}

type CategoryListResult struct {
	Items []Category `json:"items"`
}

func NormalizeCategoryInput(input CategoryInput) CategoryInput {
	return CategoryInput{Name: strings.TrimSpace(input.Name)}
}

func ValidateCategoryInput(input CategoryInput) error {
	normalized := NormalizeCategoryInput(input)
	if normalized.Name == "" {
		return fmt.Errorf("category name is required")
	}
	if len(normalized.Name) > 100 {
		return fmt.Errorf("category name too long")
	}
	return nil
}
