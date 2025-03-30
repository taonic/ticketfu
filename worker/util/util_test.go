package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		limit     int
		expected  []string
		truncated bool
	}{
		{
			name:      "No truncation needed",
			input:     []string{"a", "b", "c"},
			limit:     10,
			expected:  []string{"a", "b", "c"},
			truncated: false,
		},
		{
			name:      "Exact limit",
			input:     []string{"a", "b", "c"},
			limit:     3,
			expected:  []string{"a", "b", "c"},
			truncated: false,
		},
		{
			name:      "Truncation required",
			input:     []string{"a", "bb", "ccc", "dddd"},
			limit:     5,
			expected:  []string{"ccc", "dddd"},
			truncated: true,
		},
		{
			name:      "Empty input",
			input:     []string{},
			limit:     10,
			expected:  []string{},
			truncated: false,
		},
		{
			name:      "Negative limit",
			input:     []string{"a", "b"},
			limit:     -1,
			expected:  []string{},
			truncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, truncated := TruncateStringSlice(tt.input, tt.limit)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}

func TestTruncateStringMap(t *testing.T) {
	tests := []struct {
		name      string
		input     map[int64]string
		limit     int
		expected  map[int64]string
		truncated bool
	}{
		{
			name:      "No truncation needed",
			input:     map[int64]string{1: "a", 2: "b", 3: "c"},
			limit:     5,
			expected:  map[int64]string{1: "a", 2: "b", 3: "c"},
			truncated: false,
		},
		{
			name:      "Exact limit",
			input:     map[int64]string{1: "a", 2: "b"},
			limit:     2,
			expected:  map[int64]string{1: "a", 2: "b"},
			truncated: false,
		},
		{
			name:      "Truncation required",
			input:     map[int64]string{1: "a", 2: "bb", 3: "ccc", 4: "dddd"},
			limit:     2,
			expected:  map[int64]string{3: "ccc", 4: "dddd"},
			truncated: true,
		},
		{
			name:      "Empty input",
			input:     map[int64]string{},
			limit:     10,
			expected:  map[int64]string{},
			truncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, truncated := TruncateStringMap(tt.input, tt.limit)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}
