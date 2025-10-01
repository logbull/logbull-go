package validation

import (
	"strings"
	"testing"
)

func TestValidateProjectID(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		wantErr   bool
	}{
		{
			name:      "valid UUID",
			projectID: "12345678-1234-1234-1234-123456789012",
			wantErr:   false,
		},
		{
			name:      "valid UUID with uppercase",
			projectID: "12345678-1234-1234-1234-12345678ABCD",
			wantErr:   false,
		},
		{
			name:      "empty string",
			projectID: "",
			wantErr:   true,
		},
		{
			name:      "invalid format - no dashes",
			projectID: "12345678123412341234123456789012",
			wantErr:   true,
		},
		{
			name:      "invalid format - wrong length",
			projectID: "12345678-1234-1234-1234-1234567890",
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			projectID: "   ",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectID(tt.projectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostURL(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			host:    "http://localhost:4005",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			host:    "https://logbull.example.com",
			wantErr: false,
		},
		{
			name:    "valid https with path",
			host:    "https://example.com/logbull",
			wantErr: false,
		},
		{
			name:    "empty string",
			host:    "",
			wantErr: true,
		},
		{
			name:    "invalid scheme - ftp",
			host:    "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "no scheme",
			host:    "example.com",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			host:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostURL(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "abc123_xyz-789.test",
			wantErr: false,
		},
		{
			name:    "minimum length",
			apiKey:  "1234567890",
			wantErr: false,
		},
		{
			name:    "too short",
			apiKey:  "short",
			wantErr: true,
		},
		{
			name:    "empty string",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			apiKey:  "invalid@key!here",
			wantErr: true,
		},
		{
			name:    "spaces",
			apiKey:  "has spaces in it",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIKey(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "valid message",
			message: "This is a valid log message",
			wantErr: false,
		},
		{
			name:    "empty string",
			message: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			message: "   ",
			wantErr: true,
		},
		{
			name:    "message at max length",
			message: strings.Repeat("a", maxMessageLength),
			wantErr: false,
		},
		{
			name:    "message too long",
			message: strings.Repeat("a", maxMessageLength+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  map[string]any
		wantErr bool
	}{
		{
			name: "valid fields",
			fields: map[string]any{
				"user_id": "12345",
				"action":  "login",
			},
			wantErr: false,
		},
		{
			name:    "nil fields",
			fields:  nil,
			wantErr: false,
		},
		{
			name:    "empty map",
			fields:  map[string]any{},
			wantErr: false,
		},
		{
			name: "too many fields",
			fields: func() map[string]any {
				m := make(map[string]any)
				for i := 0; i < maxFieldsCount+1; i++ {
					m[string(rune('a'+i))] = i
				}
				return m
			}(),
			wantErr: true,
		},
		{
			name: "field key too long",
			fields: map[string]any{
				strings.Repeat("a", maxFieldKeyLen+1): "value",
			},
			wantErr: true,
		},
		{
			name: "empty field key",
			fields: map[string]any{
				"": "value",
			},
			wantErr: true,
		},
		{
			name: "whitespace field key",
			fields: map[string]any{
				"   ": "value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogFields(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogFields() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
