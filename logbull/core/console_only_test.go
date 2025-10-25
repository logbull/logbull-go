package core

import (
	"testing"
)

func TestNewLogger_ConsoleOnlyMode(t *testing.T) {
	t.Run("no credentials provided", func(t *testing.T) {
		logger, err := NewLogger(Config{})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger.sender != nil {
			t.Error("NewLogger() sender should be nil in console-only mode")
		}
		defer logger.Shutdown()
	})

	t.Run("only project ID provided", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger.sender != nil {
			t.Error("NewLogger() sender should be nil in console-only mode")
		}
		defer logger.Shutdown()
	})

	t.Run("only host provided", func(t *testing.T) {
		logger, err := NewLogger(Config{
			Host: "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger.sender != nil {
			t.Error("NewLogger() sender should be nil in console-only mode")
		}
		defer logger.Shutdown()
	})

	t.Run("empty strings for credentials", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "",
			Host:      "",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger.sender != nil {
			t.Error("NewLogger() sender should be nil in console-only mode")
		}
		defer logger.Shutdown()
	})

	t.Run("with full credentials not in console-only mode", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger.sender == nil {
			t.Error("NewLogger() sender should not be nil with full credentials")
		}
		defer logger.Shutdown()
	})
}

func TestLogBullLogger_ConsoleOnlyLogging(t *testing.T) {
	logger, err := NewLogger(Config{})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	// These should all work without errors (just print to console)
	logger.Debug("debug message", map[string]any{"key": "value"})
	logger.Info("info message", map[string]any{"key": "value"})
	logger.Warning("warning message", map[string]any{"key": "value"})
	logger.Error("error message", map[string]any{"key": "value"})
	logger.Critical("critical message", map[string]any{"key": "value"})

	// Verify no panics occurred
}

func TestLogBullLogger_ConsoleOnlyWithContext(t *testing.T) {
	logger, err := NewLogger(Config{})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	contextLogger := logger.WithContext(map[string]any{
		"request_id": "req_123",
		"user_id":    "user_456",
	})

	if contextLogger == nil {
		t.Error("WithContext() returned nil")
		return
	}

	if contextLogger.sender != nil {
		t.Error("WithContext() sender should be nil in console-only mode")
	}

	if len(contextLogger.context) != 2 {
		t.Errorf("WithContext() context length = %d, want 2", len(contextLogger.context))
	}

	contextLogger.Info("test message", map[string]any{
		"action": "test",
	})
}

func TestLogBullLogger_ConsoleOnlyFlushAndShutdown(t *testing.T) {
	logger, err := NewLogger(Config{})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	logger.Info("test message", nil)

	// These should not panic even when sender is nil
	logger.Flush()
	logger.Shutdown()
}

func TestLogBullLogger_ConsoleOnlyLevelFiltering(t *testing.T) {
	logger, err := NewLogger(Config{
		LogLevel: WARNING,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	// These should be filtered based on level even in console-only mode
	logger.Debug("should be filtered", nil)
	logger.Info("should be filtered", nil)
	logger.Warning("should pass", nil)
	logger.Error("should pass", nil)
	logger.Critical("should pass", nil)

	// Verify no panics occurred
}
