package core

import "testing"

func TestLogLevel_Priority(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected int
	}{
		{DEBUG, 10},
		{INFO, 20},
		{WARNING, 30},
		{ERROR, 40},
		{CRITICAL, 50},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.Priority(); got != tt.expected {
				t.Errorf("LogLevel.Priority() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARNING, "WARNING"},
		{ERROR, "ERROR"},
		{CRITICAL, "CRITICAL"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogLevel_Ordering(t *testing.T) {
	if DEBUG.Priority() >= INFO.Priority() {
		t.Error("DEBUG should have lower priority than INFO")
	}
	if INFO.Priority() >= WARNING.Priority() {
		t.Error("INFO should have lower priority than WARNING")
	}
	if WARNING.Priority() >= ERROR.Priority() {
		t.Error("WARNING should have lower priority than ERROR")
	}
	if ERROR.Priority() >= CRITICAL.Priority() {
		t.Error("ERROR should have lower priority than CRITICAL")
	}
}
