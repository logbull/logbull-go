package formatting

import (
	"strings"
	"testing"
)

func TestFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "normal message",
			message:  "This is a test message",
			expected: "This is a test message",
		},
		{
			name:     "message with leading/trailing whitespace",
			message:  "  test message  ",
			expected: "test message",
		},
		{
			name:     "message at max length",
			message:  strings.Repeat("a", defaultMaxMessageLength),
			expected: strings.Repeat("a", defaultMaxMessageLength),
		},
		{
			name:     "message exceeding max length",
			message:  strings.Repeat("a", defaultMaxMessageLength+100),
			expected: strings.Repeat("a", defaultMaxMessageLength-3) + "...",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMessage(tt.message)
			if result != tt.expected {
				t.Errorf("FormatMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnsureFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]any
		expected map[string]any
	}{
		{
			name: "valid fields",
			fields: map[string]any{
				"user_id": "12345",
				"count":   42,
			},
			expected: map[string]any{
				"user_id": "12345",
				"count":   42,
			},
		},
		{
			name:     "nil fields",
			fields:   nil,
			expected: map[string]any{},
		},
		{
			name:     "empty map",
			fields:   map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "fields with whitespace keys",
			fields: map[string]any{
				"  user_id  ": "12345",
				"count":       42,
			},
			expected: map[string]any{
				"user_id": "12345",
				"count":   42,
			},
		},
		{
			name: "fields with empty key",
			fields: map[string]any{
				"":        "should be skipped",
				"user_id": "12345",
			},
			expected: map[string]any{
				"user_id": "12345",
			},
		},
		{
			name: "fields with whitespace-only key",
			fields: map[string]any{
				"   ":     "should be skipped",
				"user_id": "12345",
			},
			expected: map[string]any{
				"user_id": "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureFields(tt.fields)
			if len(result) != len(tt.expected) {
				t.Errorf("EnsureFields() length = %v, want %v", len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("EnsureFields()[%q] = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}

func TestMergeFields(t *testing.T) {
	tests := []struct {
		name       string
		base       map[string]any
		additional map[string]any
		expected   map[string]any
	}{
		{
			name: "merge two maps",
			base: map[string]any{
				"user_id": "12345",
				"role":    "admin",
			},
			additional: map[string]any{
				"action": "login",
			},
			expected: map[string]any{
				"user_id": "12345",
				"role":    "admin",
				"action":  "login",
			},
		},
		{
			name: "override values",
			base: map[string]any{
				"user_id": "12345",
				"count":   10,
			},
			additional: map[string]any{
				"count": 20,
			},
			expected: map[string]any{
				"user_id": "12345",
				"count":   20,
			},
		},
		{
			name: "nil base",
			base: nil,
			additional: map[string]any{
				"action": "login",
			},
			expected: map[string]any{
				"action": "login",
			},
		},
		{
			name: "nil additional",
			base: map[string]any{
				"user_id": "12345",
			},
			additional: nil,
			expected: map[string]any{
				"user_id": "12345",
			},
		},
		{
			name:       "both nil",
			base:       nil,
			additional: nil,
			expected:   map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeFields(tt.base, tt.additional)
			if len(result) != len(tt.expected) {
				t.Errorf("MergeFields() length = %v, want %v", len(result), len(tt.expected))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("MergeFields()[%q] = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}

func TestIsJSONSerializable(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{
			name:     "string",
			value:    "test",
			expected: true,
		},
		{
			name:     "int",
			value:    42,
			expected: true,
		},
		{
			name:     "float",
			value:    3.14,
			expected: true,
		},
		{
			name:     "bool",
			value:    true,
			expected: true,
		},
		{
			name:     "nil",
			value:    nil,
			expected: true,
		},
		{
			name:     "slice",
			value:    []string{"a", "b"},
			expected: true,
		},
		{
			name:     "map",
			value:    map[string]any{"key": "value"},
			expected: true,
		},
		{
			name:     "channel (not serializable)",
			value:    make(chan int),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONSerializable(tt.value)
			if result != tt.expected {
				t.Errorf("isJSONSerializable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertToString(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "nil",
			value: nil,
		},
		{
			name:  "string",
			value: "test",
		},
		{
			name:  "int",
			value: 42,
		},
		{
			name:  "slice",
			value: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToString(tt.value)
			if result == "" && tt.value != nil && tt.value != "" {
				t.Errorf("convertToString() returned empty string for %v", tt.value)
			}
		})
	}
}
