package domain

import "testing"

func TestValidateEntryInput_Valid(t *testing.T) {
	input := EntryInput{
		Title:    "  Test Title  ",
		Abstract: "Abstract text",
		Category: "Clinical Psychology",
		Year:     2024,
		Author:   "Jane Doe",
		Source:   "Test Journal",
		Type:     "Journal",
		URL:      "",
	}

	if err := ValidateEntryInput(input); err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	normalized := NormalizeEntryInput(input)
	if normalized.Title != "Test Title" {
		t.Fatalf("expected trimmed title, got %q", normalized.Title)
	}
	if normalized.URL != DefaultURL {
		t.Fatalf("expected default url, got %q", normalized.URL)
	}
}

func TestValidateEntryInput_MissingRequired(t *testing.T) {
	input := EntryInput{
		Abstract: "Abstract text",
		Category: "Clinical Psychology",
		Year:     2024,
		Author:   "Jane Doe",
		Source:   "Test Journal",
		Type:     "Journal",
	}

	if err := ValidateEntryInput(input); err == nil {
		t.Fatal("expected validation error for missing title")
	}
}

func TestValidateEntryInput_InvalidCategory(t *testing.T) {
	input := EntryInput{
		Title:    "Test",
		Abstract: "Abstract",
		Category: "Invalid Category",
		Year:     2024,
		Author:   "Author",
		Source:   "Source",
		Type:     "Journal",
	}

	if err := ValidateEntryInput(input); err == nil {
		t.Fatal("expected validation error for invalid category")
	}
}

func TestValidateEntryInput_InvalidType(t *testing.T) {
	input := EntryInput{
		Title:    "Test",
		Abstract: "Abstract",
		Category: "Clinical Psychology",
		Year:     2024,
		Author:   "Author",
		Source:   "Source",
		Type:     "Book",
	}

	if err := ValidateEntryInput(input); err == nil {
		t.Fatal("expected validation error for invalid type")
	}
}

func TestValidateEntryInput_InvalidYear(t *testing.T) {
	input := EntryInput{
		Title:    "Test",
		Abstract: "Abstract",
		Category: "Clinical Psychology",
		Year:     0,
		Author:   "Author",
		Source:   "Source",
		Type:     "Journal",
	}

	if err := ValidateEntryInput(input); err == nil {
		t.Fatal("expected validation error for invalid year")
	}
}
