package org

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			result, truncated := truncateStringMap(tt.input, tt.limit)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}
