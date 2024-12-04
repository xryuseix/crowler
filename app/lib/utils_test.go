package lib

import (
	"net/url"
	"reflect"
	"sort"
	"testing"
)

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		filterFn func(string) bool
		expected []string
	}{
		{
			name:  "Filter empty slice",
			input: []string{},
			filterFn: func(s string) bool {
				return len(s) > 0
			},
			expected: []string{},
		},
		{
			name:  "Filter strings with length greater than 3",
			input: []string{"go", "golang", "test", "filter"},
			filterFn: func(s string) bool {
				return len(s) > 3
			},
			expected: []string{"golang", "test", "filter"},
		},
		{
			name:  "Filter strings containing 'a'",
			input: []string{"apple", "banana", "cherry", "date"},
			filterFn: func(s string) bool {
				return contains(s, 'a')
			},
			expected: []string{"apple", "banana", "date"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.filterFn)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSplitBySpace(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Split empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Split single word strings",
			input:    []string{"go", "lang"},
			expected: []string{"go", "lang"},
		},
		{
			name:     "Split strings with spaces",
			input:    []string{"go lang", "test filter space"},
			expected: []string{"go", "lang", "test", "filter", "space"},
		},
		{
			name:     "Split strings with multiple spaces",
			input:    []string{"go  lang", "test  filter"},
			expected: []string{"go", "", "lang", "test", "", "filter"},
		},
		{
			name:     "Split strings with leading and trailing spaces",
			input:    []string{" go ", " lang "},
			expected: []string{"", "go", "", "", "lang", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitBySpace(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToAbsoluteLink(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")

	tests := []struct {
		name     string
		base     *url.URL
		links    []string
		expected []string
	}{
		{
			name:     "Empty links",
			base:     baseURL,
			links:    []string{},
			expected: []string{},
		},
		{
			name:     "Absolute links",
			base:     baseURL,
			links:    []string{"https://example.com/page1", "https://example.com/page2"},
			expected: []string{"https://example.com/page1", "https://example.com/page2"},
		},
		{
			name:     "Protocol-relative links",
			base:     baseURL,
			links:    []string{"//example.com/page1", "//example.com/page2"},
			expected: []string{"https://example.com/page1", "https://example.com/page2"},
		},
		{
			name:     "Path-relative links",
			base:     baseURL,
			links:    []string{"/page1", "/page2"},
			expected: []string{"https://example.com/page1", "https://example.com/page2"},
		},
		{
			name:     "Mixed links",
			base:     baseURL,
			links:    []string{"https://example2.com/page1", "//example.com/page2", "/page3"},
			expected: []string{"https://example2.com/page1", "https://example.com/page2", "https://example.com/page3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToAbsoluteLinks(tt.base, tt.links)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Hash empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Unique empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Unique with no duplicates",
			input:    []string{"go", "lang", "test"},
			expected: []string{"go", "lang", "test"},
		},
		{
			name:     "Unique with duplicates",
			input:    []string{"go", "lang", "test", "go", "lang"},
			expected: []string{"go", "lang", "test"},
		},
		{
			name:     "Unique with all duplicates",
			input:    []string{"go", "go", "go"},
			expected: []string{"go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unique(tt.input)
			sort.Strings(result)
			sort.Strings(tt.expected)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func contains(s string, char rune) bool {
	for _, c := range s {
		if c == char {
			return true
		}
	}
	return false
}
