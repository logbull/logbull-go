package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/logbull/logbull-go/logbull/core"
)

func TestNewZapCore(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		zapCore, err := NewZapCore(core.Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewZapCore() error = %v", err)
		}
		if zapCore == nil {
			t.Error("NewZapCore() returned nil")
		}
		if zapCore != nil {
			defer zapCore.Shutdown()
		}
	})

	t.Run("invalid project ID", func(t *testing.T) {
		_, err := NewZapCore(core.Config{
			ProjectID: "invalid",
			Host:      "http://localhost:4005",
		})
		if err == nil {
			t.Error("NewZapCore() expected error for invalid project ID")
		}
	})
}

func TestZapCore_Enabled(t *testing.T) {
	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      "http://localhost:4005",
		LogLevel:  core.WARNING,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	tests := []struct {
		level   zapcore.Level
		enabled bool
	}{
		{zapcore.DebugLevel, false},
		{zapcore.InfoLevel, false},
		{zapcore.WarnLevel, true},
		{zapcore.ErrorLevel, true},
		{zapcore.FatalLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := zapCore.Enabled(tt.level); got != tt.enabled {
				t.Errorf("Enabled() = %v, want %v", got, tt.enabled)
			}
		})
	}
}

func TestZapCore_Write(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	logger := zap.New(zapCore)

	logger.Info("test message",
		zap.String("user_id", "12345"),
		zap.Int("count", 42),
	)

	zapCore.Sync()
	time.Sleep(100 * time.Millisecond)
}

func TestZapCore_With(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	coreWithFields := zapCore.With([]zapcore.Field{
		zap.String("request_id", "req_123"),
	})

	logger := zap.New(coreWithFields)

	logger.Info("test with fields",
		zap.String("action", "test"),
	)

	zapCore.Sync()
	time.Sleep(100 * time.Millisecond)
}

func TestZapCore_Check(t *testing.T) {
	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      "http://localhost:4005",
		LogLevel:  core.WARNING,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	entry := zapcore.Entry{
		Level: zapcore.InfoLevel,
	}

	ce := zapCore.Check(entry, nil)
	if ce != nil {
		t.Error("Check() should return nil for filtered levels")
	}

	entry.Level = zapcore.ErrorLevel
	ce = zapCore.Check(entry, &zapcore.CheckedEntry{})
	if ce == nil {
		t.Error("Check() should return non-nil for enabled levels")
	}
}

func TestZapCore_Sync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	logger := zap.New(zapCore)

	logger.Info("test sync")

	err = logger.Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestZapCore_AllLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  core.DEBUG,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	logger := zap.New(zapCore)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	zapCore.Sync()
	time.Sleep(100 * time.Millisecond)
}

func TestConvertZapLevel(t *testing.T) {
	tests := []struct {
		zapLevel      zapcore.Level
		expectedLevel core.LogLevel
	}{
		{zapcore.DebugLevel, core.DEBUG},
		{zapcore.InfoLevel, core.INFO},
		{zapcore.WarnLevel, core.WARNING},
		{zapcore.ErrorLevel, core.ERROR},
		{zapcore.DPanicLevel, core.CRITICAL},
		{zapcore.PanicLevel, core.CRITICAL},
		{zapcore.FatalLevel, core.CRITICAL},
	}

	for _, tt := range tests {
		t.Run(tt.zapLevel.String(), func(t *testing.T) {
			result := convertZapLevel(tt.zapLevel)
			if result != tt.expectedLevel {
				t.Errorf("convertZapLevel() = %v, want %v", result, tt.expectedLevel)
			}
		})
	}
}

func TestConvertLogLevelToZap(t *testing.T) {
	tests := []struct {
		logbullLevel core.LogLevel
		expectedZap  zapcore.Level
	}{
		{core.DEBUG, zapcore.DebugLevel},
		{core.INFO, zapcore.InfoLevel},
		{core.WARNING, zapcore.WarnLevel},
		{core.ERROR, zapcore.ErrorLevel},
		{core.CRITICAL, zapcore.FatalLevel},
	}

	for _, tt := range tests {
		t.Run(string(tt.logbullLevel), func(t *testing.T) {
			result := convertLogLevelToZap(tt.logbullLevel)
			if result != tt.expectedZap {
				t.Errorf("convertLogLevelToZap() = %v, want %v", result, tt.expectedZap)
			}
		})
	}
}

func TestZapCore_ComplexFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	zapCore, err := NewZapCore(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewZapCore() error = %v", err)
	}
	defer zapCore.Shutdown()

	logger := zap.New(zapCore)

	logger.Info("complex fields",
		zap.String("string", "value"),
		zap.Int("int", 42),
		zap.Float64("float", 3.14),
		zap.Bool("bool", true),
		zap.Strings("array", []string{"a", "b", "c"}),
	)

	zapCore.Sync()
	time.Sleep(100 * time.Millisecond)
}
