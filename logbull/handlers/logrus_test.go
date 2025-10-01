package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/logbull/logbull-go/logbull/core"
)

func TestNewLogrusHook(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		hook, err := NewLogrusHook(core.Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		})
		if err != nil {
			t.Errorf("NewLogrusHook() error = %v", err)
		}
		if hook == nil {
			t.Error("NewLogrusHook() returned nil")
		}
		if hook != nil {
			defer hook.Shutdown()
		}
	})

	t.Run("invalid project ID", func(t *testing.T) {
		_, err := NewLogrusHook(core.Config{
			ProjectID: "invalid",
			Host:      "http://localhost:4005",
		})
		if err == nil {
			t.Error("NewLogrusHook() expected error for invalid project ID")
		}
	})
}

func TestLogrusHook_Levels(t *testing.T) {
	hook, err := NewLogrusHook(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      "http://localhost:4005",
		LogLevel:  core.WARNING,
	})
	if err != nil {
		t.Fatalf("NewLogrusHook() error = %v", err)
	}
	defer hook.Shutdown()

	levels := hook.Levels()

	expectedLevels := map[logrus.Level]bool{
		logrus.WarnLevel:  true,
		logrus.ErrorLevel: true,
		logrus.FatalLevel: true,
		logrus.PanicLevel: true,
	}

	for _, level := range levels {
		if !expectedLevels[level] {
			t.Errorf("Unexpected level in Levels(): %v", level)
		}
	}
}

func TestLogrusHook_Fire(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	hook, err := NewLogrusHook(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogrusHook() error = %v", err)
	}
	defer hook.Shutdown()

	logger := logrus.New()
	logger.AddHook(hook)

	logger.WithFields(logrus.Fields{
		"user_id": "12345",
		"action":  "login",
	}).Info("User logged in")

	hook.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestLogrusHook_AllLevels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	hook, err := NewLogrusHook(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  core.DEBUG,
	})
	if err != nil {
		t.Fatalf("NewLogrusHook() error = %v", err)
	}
	defer hook.Shutdown()

	logger := logrus.New()
	logger.AddHook(hook)
	logger.SetLevel(logrus.TraceLevel)

	logger.Trace("trace message")
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	hook.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestLogrusHook_WithFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	hook, err := NewLogrusHook(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	})
	if err != nil {
		t.Fatalf("NewLogrusHook() error = %v", err)
	}
	defer hook.Shutdown()

	logger := logrus.New()
	logger.AddHook(hook)

	logger.WithFields(logrus.Fields{
		"string":  "value",
		"int":     42,
		"float":   3.14,
		"bool":    true,
		"complex": map[string]string{"nested": "value"},
	}).Info("Complex fields test")

	hook.Flush()
	time.Sleep(100 * time.Millisecond)
}

func TestConvertLogrusLevel(t *testing.T) {
	tests := []struct {
		logrusLevel   logrus.Level
		expectedLevel core.LogLevel
	}{
		{logrus.TraceLevel, core.DEBUG},
		{logrus.DebugLevel, core.DEBUG},
		{logrus.InfoLevel, core.INFO},
		{logrus.WarnLevel, core.WARNING},
		{logrus.ErrorLevel, core.ERROR},
		{logrus.FatalLevel, core.CRITICAL},
		{logrus.PanicLevel, core.CRITICAL},
	}

	for _, tt := range tests {
		t.Run(tt.logrusLevel.String(), func(t *testing.T) {
			result := convertLogrusLevel(tt.logrusLevel)
			if result != tt.expectedLevel {
				t.Errorf("convertLogrusLevel() = %v, want %v", result, tt.expectedLevel)
			}
		})
	}
}

func TestLevelsFromConfig(t *testing.T) {
	tests := []struct {
		minLevel         core.LogLevel
		shouldContain    []logrus.Level
		shouldNotContain []logrus.Level
	}{
		{
			minLevel: core.DEBUG,
			shouldContain: []logrus.Level{
				logrus.TraceLevel,
				logrus.DebugLevel,
				logrus.InfoLevel,
				logrus.WarnLevel,
				logrus.ErrorLevel,
				logrus.FatalLevel,
				logrus.PanicLevel,
			},
			shouldNotContain: []logrus.Level{},
		},
		{
			minLevel: core.WARNING,
			shouldContain: []logrus.Level{
				logrus.WarnLevel,
				logrus.ErrorLevel,
				logrus.FatalLevel,
				logrus.PanicLevel,
			},
			shouldNotContain: []logrus.Level{
				logrus.TraceLevel,
				logrus.DebugLevel,
				logrus.InfoLevel,
			},
		},
		{
			minLevel: core.ERROR,
			shouldContain: []logrus.Level{
				logrus.ErrorLevel,
				logrus.FatalLevel,
				logrus.PanicLevel,
			},
			shouldNotContain: []logrus.Level{
				logrus.TraceLevel,
				logrus.DebugLevel,
				logrus.InfoLevel,
				logrus.WarnLevel,
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.minLevel), func(t *testing.T) {
			levels := levelsFromConfig(tt.minLevel)

			levelMap := make(map[logrus.Level]bool)
			for _, level := range levels {
				levelMap[level] = true
			}

			for _, level := range tt.shouldContain {
				if !levelMap[level] {
					t.Errorf("Expected level %v to be included", level)
				}
			}

			for _, level := range tt.shouldNotContain {
				if levelMap[level] {
					t.Errorf("Expected level %v to be excluded", level)
				}
			}
		})
	}
}

func TestLogrusHook_LevelFiltering(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(core.LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	hook, err := NewLogrusHook(core.Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		LogLevel:  core.ERROR,
	})
	if err != nil {
		t.Fatalf("NewLogrusHook() error = %v", err)
	}
	defer hook.Shutdown()

	logger := logrus.New()
	logger.AddHook(hook)
	logger.SetLevel(logrus.TraceLevel)

	logger.Debug("should be filtered")
	logger.Info("should be filtered")
	logger.Warn("should be filtered")
	logger.Error("should pass")

	hook.Flush()
	time.Sleep(100 * time.Millisecond)
}
