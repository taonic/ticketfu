package ticket

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
			result, truncated := truncateStringSlice(tt.input, tt.limit)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.truncated, truncated)
		})
	}
}
