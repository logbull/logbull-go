package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logbull/logbull-go/logbull/core"
)

func TestNewSlogHandler(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		handler, err := NewSlogHandler(core.Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewSlogHandler() error = %v", err)
		}
		if handler == nil {
			t.Error("NewSlogHandler() returned nil")
		}
		if handler != nil {
			defer handler.Shutdown()
		}
	})

	t.Run("invalid project ID", func(t *testing.T) {
		_, err := NewSlogHandler(core.Config{
			ProjectID: "invalid",
			Host:      "http://localhost:4005",
		})
		if err == nil {
			t.Error("NewSlogHandler() expected error for invalid project ID")
		}
	})
}

func TestSlogHandler_Enabled(t *testing.T) {
	handler, err := NewSlogHandler(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      "http://localhost:4005",
		LogLevel:  core.WARNING,
	})
	if err != nil {
		t.Fatalf("NewSlogHandler() error = %v", err)
	}
	defer handler.Shutdown()

	ctx := context.Background()

	tests := []struct {
		level   slog.Level
		enabled bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := handler.Enabled(ctx, tt.level); got != tt.enabled {
				t.Errorf("Enabled() = %v, want %v", got, tt.enabled)
			}
		})
	}
}

func TestSlogHandler_Handle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	handler, err := NewSlogHandler(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewSlogHandler() error = %v", err)
	}
	defer handler.Shutdown()

	logger := slog.New(handler)

	logger.Info("test message",
		slog.String("user_id", "12345"),
		slog.Int("count", 42),
	)

	handler.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestSlogHandler_WithAttrs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	handler, err := NewSlogHandler(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewSlogHandler() error = %v", err)
	}
	defer handler.Shutdown()

	handlerWithAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("request_id", "req_123"),
	})

	logger := slog.New(handlerWithAttrs)

	logger.Info("test with attrs",
		slog.String("action", "test"),
	)

	handler.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestSlogHandler_WithGroup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	handler, err := NewSlogHandler(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewSlogHandler() error = %v", err)
	}
	defer handler.Shutdown()

	handlerWithGroup := handler.WithGroup("request")

	logger := slog.New(handlerWithGroup)

	logger.Info("test with group",
		slog.String("method", "POST"),
		slog.Int("status", 200),
	)

	handler.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestSlogHandler_Groups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	handler, err := NewSlogHandler(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewSlogHandler() error = %v", err)
	}
	defer handler.Shutdown()

	logger := slog.New(handler)

	logger.Info("request processed",
		slog.Group("request",
			slog.String("method", "POST"),
			slog.String("path", "/api/users"),
			slog.Int("status", 201),
		),
	)

	handler.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestConvertSlogLevel(t *testing.T) {
	tests := []struct {
		slogLevel     slog.Level
		expectedLevel core.LogLevel
	}{
		{slog.LevelDebug, core.DEBUG},
		{slog.LevelInfo, core.INFO},
		{slog.LevelWarn, core.WARNING},
		{slog.LevelError, core.ERROR},
		{slog.LevelError + 4, core.ERROR},
	}

	for _, tt := range tests {
		t.Run(tt.slogLevel.String(), func(t *testing.T) {
			result := convertSlogLevel(tt.slogLevel)
			if result != tt.expectedLevel {
				t.Errorf("convertSlogLevel() = %v, want %v", result, tt.expectedLevel)
			}
		})
	}
}
