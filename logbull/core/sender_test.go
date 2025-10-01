package core

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewSender(t *testing.T) {
	t.Run("creates sender successfully", func(t *testing.T) {
		config := &Config{
			ProjectID: "12345678-1234-1234-1234-123456789012",
			Host:      "http://localhost:4005",
		}

		sender, err := NewSender(config)
		if err != nil {
			t.Errorf("NewSender() error = %v", err)
		}
		if sender == nil {
			t.Error("NewSender() returned nil")
		}
		if sender != nil {
			defer sender.Shutdown()
		}
	})
}

func TestSender_AddLog(t *testing.T) {
	var requestCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{
			Accepted: 1,
			Rejected: 0,
		})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "test message",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{"key": "value"},
	})

	sender.Flush()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if requestCount == 0 {
		t.Error("Expected at least one HTTP request")
	}
}

func TestSender_Batching(t *testing.T) {
	var receivedBatches []LogBatch
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var batch LogBatch
		json.Unmarshal(body, &batch)

		mu.Lock()
		receivedBatches = append(receivedBatches, batch)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{
			Accepted: len(batch.Logs),
			Rejected: 0,
		})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	for i := 0; i < 10; i++ {
		sender.AddLog(LogEntry{
			Level:     "INFO",
			Message:   "test message",
			Timestamp: GenerateUniqueTimestamp(),
			Fields:    map[string]any{"index": i},
		})
	}

	sender.Flush()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(receivedBatches) == 0 {
		t.Error("No batches received")
	}

	totalLogs := 0
	for _, batch := range receivedBatches {
		totalLogs += len(batch.Logs)
	}

	if totalLogs != 10 {
		t.Errorf("Expected 10 logs total, got %d", totalLogs)
	}
}

func TestSender_LargeBatch(t *testing.T) {
	var receivedBatches []LogBatch
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var batch LogBatch
		json.Unmarshal(body, &batch)

		mu.Lock()
		receivedBatches = append(receivedBatches, batch)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{
			Accepted: len(batch.Logs),
			Rejected: 0,
		})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	for i := 0; i < 1500; i++ {
		sender.AddLog(LogEntry{
			Level:     "INFO",
			Message:   "test message",
			Timestamp: GenerateUniqueTimestamp(),
			Fields:    map[string]any{"index": i},
		})
	}

	sender.Flush()
	time.Sleep(200 * time.Millisecond)
	sender.Flush() // Flush again to send remaining logs
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	totalLogs := 0
	for _, batch := range receivedBatches {
		totalLogs += len(batch.Logs)
		if len(batch.Logs) > batchSize {
			t.Errorf("Batch size %d exceeds maximum %d", len(batch.Logs), batchSize)
		}
	}

	if totalLogs != 1500 {
		t.Errorf("Expected 1500 logs total, got %d", totalLogs)
	}
}

func TestSender_HTTPHeaders(t *testing.T) {
	var contentType, apiKey, userAgent string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		contentType = r.Header.Get("Content-Type")
		apiKey = r.Header.Get("X-API-Key")
		userAgent = r.Header.Get("User-Agent")
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
		APIKey:    "test-api-key",
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "test",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{},
	})

	sender.Flush()
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if contentType != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", contentType)
	}

	if apiKey != "test-api-key" {
		t.Errorf("Expected X-API-Key: test-api-key, got %s", apiKey)
	}

	if userAgent != "LogBull-Go-Client/1.0" {
		t.Errorf("Expected User-Agent: LogBull-Go-Client/1.0, got %s", userAgent)
	}
}

func TestSender_RejectedLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{
			Accepted: 1,
			Rejected: 1,
			Errors: []RejectedLog{
				{
					Index:   1,
					Message: "Invalid log format",
				},
			},
		})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "test message",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{},
	})

	sender.Flush()
	time.Sleep(200 * time.Millisecond)
}

func TestSender_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "test message",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{},
	})

	sender.Flush()
	time.Sleep(200 * time.Millisecond)
}

func TestSender_ConcurrentSending(t *testing.T) {
	var requestCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}
	defer sender.Shutdown()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sender.AddLog(LogEntry{
				Level:     "INFO",
				Message:   "concurrent test",
				Timestamp: GenerateUniqueTimestamp(),
				Fields:    map[string]any{"id": id},
			})
		}(i)
	}

	wg.Wait()
	sender.Flush()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if requestCount == 0 {
		t.Error("Expected at least one request")
	}
}

func TestSender_Shutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "before shutdown",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{},
	})

	sender.Shutdown()

	sender.AddLog(LogEntry{
		Level:     "INFO",
		Message:   "after shutdown",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{},
	})
}

func TestSender_MultipleShutdowns(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, err := NewSender(config)
	if err != nil {
		t.Fatalf("NewSender() error = %v", err)
	}

	sender.Shutdown()
	sender.Shutdown()
	sender.Shutdown()
}

func BenchmarkSender_AddLog(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LogBullResponse{Accepted: 1})
	}))
	defer server.Close()

	config := &Config{
		ProjectID: "12345678-1234-1234-1234-123456789012",
		Host:      server.URL,
	}

	sender, _ := NewSender(config)
	defer sender.Shutdown()

	entry := LogEntry{
		Level:     "INFO",
		Message:   "benchmark message",
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    map[string]any{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sender.AddLog(entry)
	}
}
