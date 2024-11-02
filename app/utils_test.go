package main

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
			result := filter(tt.input, tt.filterFn)
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
