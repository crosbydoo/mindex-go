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

func TestBuildPagination(t *testing.T) {
	p := BuildPagination(2, 20, 342)
	if p.Page != 2 || p.Limit != 20 || p.Total != 342 || p.TotalPages != 18 {
		t.Fatalf("unexpected pagination values: %+v", p)
	}
	if !p.HasNext {
		t.Fatal("expected has_next true")
	}
	if !p.HasPrev {
		t.Fatal("expected has_prev true")
	}

	first := BuildPagination(1, 20, 342)
	if first.HasPrev {
		t.Fatal("expected has_prev false on first page")
	}
	if !first.HasNext {
		t.Fatal("expected has_next true on first page")
	}

	last := BuildPagination(18, 20, 342)
	if last.HasNext {
		t.Fatal("expected has_next false on last page")
	}
	if !last.HasPrev {
		t.Fatal("expected has_prev true on last page")
	}

	empty := BuildPagination(1, 10, 0)
	if empty.HasNext || empty.HasPrev || empty.TotalPages != 0 {
		t.Fatalf("unexpected empty pagination: %+v", empty)
	}
}

func TestParseArchiveScope(t *testing.T) {
	cases := []struct {
		in   string
		want ArchiveScope
		err  bool
	}{
		{"", ArchiveActive, false},
		{"false", ArchiveActive, false},
		{"active", ArchiveActive, false},
		{"true", ArchiveOnly, false},
		{"archived", ArchiveOnly, false},
		{"all", ArchiveAll, false},
		{"nope", "", true},
	}
	for _, tc := range cases {
		got, err := ParseArchiveScope(tc.in)
		if tc.err {
			if err == nil {
				t.Fatalf("input %q: expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("input %q: unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("input %q: expected %q, got %q", tc.in, tc.want, got)
		}
	}
}
