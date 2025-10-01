package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	batchSize     = 1_000
	batchInterval = 1 * time.Second
	queueCapacity = 10_000
	minWorkers    = 1
	maxWorkers    = 10
	httpTimeout   = 30 * time.Second
)

type Sender struct {
	config       *Config
	logQueue     chan LogEntry
	stopCh       chan struct{}
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	client       *http.Client
	workerSem    chan struct{}
}

func NewSender(config *Config) (*Sender, error) {
	s := &Sender{
		config:    config,
		logQueue:  make(chan LogEntry, queueCapacity),
		stopCh:    make(chan struct{}),
		client:    &http.Client{Timeout: httpTimeout},
		workerSem: make(chan struct{}, maxWorkers),
	}

	for i := 0; i < minWorkers; i++ {
		s.workerSem <- struct{}{}
	}

	registerSender(s)

	s.wg.Add(1)
	go s.batchProcessor()

	return s, nil
}

func (s *Sender) AddLog(entry LogEntry) {
	select {
	case s.logQueue <- entry:
	case <-s.stopCh:
	default:
		fmt.Fprintf(os.Stderr, "LogBull: log queue full, dropping log\n")
	}
}

func (s *Sender) Flush() {
	s.sendBatch()
}

func (s *Sender) Shutdown() {
	s.shutdownOnce.Do(func() {
		close(s.stopCh)
		s.sendBatch()
		s.wg.Wait()
	})
}

func (s *Sender) batchProcessor() {
	defer s.wg.Done()

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendBatch()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Sender) sendBatch() {
	var logs []LogEntry

	for i := 0; i < batchSize; i++ {
		select {
		case log := <-s.logQueue:
			logs = append(logs, log)
		default:
			goto send
		}
	}

send:
	if len(logs) == 0 {
		return
	}

	select {
	case <-s.workerSem:
		s.wg.Add(1)
		go func(batch []LogEntry) {
			defer s.wg.Done()
			defer func() { s.workerSem <- struct{}{} }()

			s.sendHTTPRequest(batch)
		}(logs)
	default:
		s.wg.Add(1)
		go func(batch []LogEntry) {
			defer s.wg.Done()
			s.sendHTTPRequest(batch)
		}(logs)
	}
}

func (s *Sender) sendHTTPRequest(logs []LogEntry) {
	batch := LogBatch{Logs: logs}

	data, err := json.Marshal(batch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: failed to marshal batch: %v\n", err)
		return
	}

	url := fmt.Sprintf("%s/api/v1/logs/receiving/%s", s.config.Host, s.config.ProjectID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "LogBull-Go-Client/1.0")
	if s.config.APIKey != "" {
		req.Header.Set("X-API-Key", s.config.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: HTTP request failed: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "LogBull: failed to close response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: failed to read response: %v\n", err)
		return
	}

	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		fmt.Fprintf(os.Stderr, "LogBull: server returned status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	var response LogBullResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return
	}

	if response.Rejected > 0 {
		s.handleRejectedLogs(response, logs)
	}
}

func (s *Sender) handleRejectedLogs(response LogBullResponse, sentLogs []LogEntry) {
	fmt.Fprintf(os.Stderr, "LogBull: Rejected %d log entries\n", response.Rejected)

	if len(response.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "LogBull: Rejected log details:\n")
		for _, err := range response.Errors {
			if err.Index >= 0 && err.Index < len(sentLogs) {
				log := sentLogs[err.Index]
				fmt.Fprintf(os.Stderr, "  - Log #%d rejected (%s):\n", err.Index, err.Message)
				fmt.Fprintf(os.Stderr, "    Level: %s\n", log.Level)
				fmt.Fprintf(os.Stderr, "    Message: %s\n", log.Message)
				fmt.Fprintf(os.Stderr, "    Timestamp: %s\n", log.Timestamp)
				if len(log.Fields) > 0 {
					fmt.Fprintf(os.Stderr, "    Fields: %v\n", log.Fields)
				}
			}
		}
	}
}
