package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
			LogLevel:  INFO,
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v, want nil", err)
		}
		if logger == nil {
			t.Error("NewLogger() returned nil logger")
		}
		if logger != nil {
			defer logger.Shutdown()
		}
	})

	t.Run("invalid project ID", func(t *testing.T) {
		_, err := NewLogger(Config{
			ProjectID: "invalid",
			Host:      "http://localhost:4005",
		})
		if err == nil {
			t.Error("NewLogger() expected error for invalid project ID")
		}
	})

	t.Run("invalid host URL", func(t *testing.T) {
		_, err := NewLogger(Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "invalid",
		})
		if err == nil {
			t.Error("NewLogger() expected error for invalid host")
		}
	})

	t.Run("default log level", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v", err)
		}
		if logger.minLevel != INFO {
			t.Errorf("NewLogger() default log level = %v, want INFO", logger.minLevel)
		}
		defer logger.Shutdown()
	})

	t.Run("trims whitespace", func(t *testing.T) {
		logger, err := NewLogger(Config{
			ProjectID: "  12345678-1234-1234-1234-123456789012  ",
			Host:      "  http://localhost:4005  ",
			APIKey:    "  test-api-key  ",
		})
		if err != nil {
			t.Errorf("NewLogger() error = %v", err)
		}
		defer logger.Shutdown()
	})
}

func TestLogBullLogger_LogMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{
			Accepted: 1,
			Rejected: 0,
		})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  DEBUG,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	tests := []struct {
		name  string
		logFn func(string, map[string]any)
		level string
	}{
		{
			name:  "Debug",
			logFn: logger.Debug,
			level: "DEBUG",
		},
		{
			name:  "Info",
			logFn: logger.Info,
			level: "INFO",
		},
		{
			name:  "Warning",
			logFn: logger.Warning,
			level: "WARNING",
		},
		{
			name:  "Error",
			logFn: logger.Error,
			level: "ERROR",
		},
		{
			name:  "Critical",
			logFn: logger.Critical,
			level: "CRITICAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.logFn("test message", map[string]any{
				"key": "value",
			})
		})
	}
}

func TestLogBullLogger_LevelFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  WARNING,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	logger.Debug("should be filtered", nil)
	logger.Info("should be filtered", nil)
	logger.Warning("should pass", nil)
	logger.Error("should pass", nil)
	logger.Critical("should pass", nil)

	time.Sleep(100 * time.Millisecond)
}

func TestLogBullLogger_WithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  INFO,
	})
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

	if contextLogger.sender != logger.sender {
		t.Error("WithContext() should share the same sender")
	}

	if len(contextLogger.context) != 2 {
		t.Errorf("WithContext() context length = %d, want 2", len(contextLogger.context))
	}

	contextLogger.Info("test message", map[string]any{
		"action": "test",
	})
}

func TestLogBullLogger_ContextMerging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	logger1 := logger.WithContext(map[string]any{
		"base": "value1",
	})

	logger2 := logger1.WithContext(map[string]any{
		"additional": "value2",
		"base":       "overridden",
	})

	if len(logger2.context) != 2 {
		t.Errorf("Context merging failed: got %d fields, want 2", len(logger2.context))
	}

	if logger2.context["base"] != "overridden" {
		t.Error("Context override failed")
	}
}

func TestLogBullLogger_ConcurrentLogging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  DEBUG,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	var wg sync.WaitGroup
	numGoroutines := 10
	numLogs := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				logger.Info("concurrent log", map[string]any{
					"goroutine": id,
					"log_num":   j,
				})
			}
		}(i)
	}

	wg.Wait()
	logger.Flush()
	time.Sleep(500 * time.Millisecond)
}

func TestLogBullLogger_InvalidMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	logger.Info("", nil)
	logger.Info("   ", nil)

	time.Sleep(100 * time.Millisecond)
}

func TestLogBullLogger_InvalidFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	tooManyFields := make(map[string]any)
	for i := 0; i < 101; i++ {
		key := fmt.Sprintf("field_%d", i)
		tooManyFields[key] = i
	}

	logger.Info("test", tooManyFields)

	time.Sleep(100 * time.Millisecond)
}

func TestLogBullLogger_MessageTruncation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, err := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer logger.Shutdown()

	longMessage := strings.Repeat("a", 15000)
	logger.Info(longMessage, nil)

	time.Sleep(100 * time.Millisecond)
}

func BenchmarkLogBullLogger_Info(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, _ := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	defer logger.Shutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", map[string]any{
			"iteration": i,
		})
	}
}

func BenchmarkLogBullLogger_InfoParallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	logger, _ := NewLogger(Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	defer logger.Shutdown()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("benchmark message", map[string]any{
				"iteration": i,
			})
			i++
		}
	})
}
