package lib

import (
	"reflect"
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

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []interface{}
	}{
		{
			name:     "Unique empty slice",
			input:    []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "Unique integers",
			input:    []interface{}{1, 2, 2, 3, 4, 4, 5},
			expected: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "Unique strings",
			input:    []interface{}{"apple", "banana", "apple", "cherry", "banana"},
			expected: []interface{}{"apple", "banana", "cherry"},
		},
		{
			name:     "Unique mixed types",
			input:    []interface{}{"apple", 1, "banana", 2, "apple", 1},
			expected: []interface{}{"apple", 1, "banana", 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unique(tt.input)
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
