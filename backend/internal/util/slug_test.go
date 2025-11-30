package util

import (
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "simple title",
			title:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "title with multiple spaces",
			title:    "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "title with special characters",
			title:    "Hello, World! How are you?",
			expected: "hello-world-how-are-you",
		},
		{
			name:     "title with numbers",
			title:    "Top 10 Tips for 2024",
			expected: "top-10-tips-for-2024",
		},
		{
			name:     "title with leading/trailing spaces",
			title:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "title with mixed case",
			title:    "HELLO World",
			expected: "hello-world",
		},
		{
			name:     "title with dashes",
			title:    "Hello - World",
			expected: "hello-world",
		},
		{
			name:     "title with underscores",
			title:    "Hello_World",
			expected: "hello-world",
		},
		{
			name:     "title with unicode characters",
			title:    "Café au Lait",
			expected: "cafe-au-lait",
		},
		{
			name:     "empty title",
			title:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			title:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "title with consecutive special chars",
			title:    "Hello...World!!!",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.title)
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}

func TestGenerateUniqueSlug(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		existingSlugs []string
		wantPrefix    string
	}{
		{
			name:          "no conflict",
			title:         "Hello World",
			existingSlugs: []string{},
			wantPrefix:    "hello-world",
		},
		{
			name:          "with conflict - adds suffix",
			title:         "Hello World",
			existingSlugs: []string{"hello-world"},
			wantPrefix:    "hello-world-",
		},
		{
			name:          "multiple conflicts",
			title:         "Hello World",
			existingSlugs: []string{"hello-world", "hello-world-1", "hello-world-2"},
			wantPrefix:    "hello-world-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkExists := func(slug string) bool {
				for _, existing := range tt.existingSlugs {
					if existing == slug {
						return true
					}
				}
				return false
			}

			result := GenerateUniqueSlug(tt.title, checkExists)

			// Check that result starts with expected prefix
			if len(result) < len(tt.wantPrefix) {
				t.Errorf("GenerateUniqueSlug(%q) = %q, want prefix %q", tt.title, result, tt.wantPrefix)
				return
			}

			// For no conflict case, should be exact match
			if len(tt.existingSlugs) == 0 {
				if result != tt.wantPrefix {
					t.Errorf("GenerateUniqueSlug(%q) = %q, want %q", tt.title, result, tt.wantPrefix)
				}
				return
			}

			// For conflict cases, should have the prefix and not exist in existing slugs
			if result[:len(tt.wantPrefix)] != tt.wantPrefix {
				t.Errorf("GenerateUniqueSlug(%q) = %q, want prefix %q", tt.title, result, tt.wantPrefix)
			}

			if checkExists(result) {
				t.Errorf("GenerateUniqueSlug(%q) = %q, but slug already exists", tt.title, result)
			}
		})
	}
}
