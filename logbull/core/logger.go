package core

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/logbull/logbull-go/logbull/internal/formatting"
	"github.com/logbull/logbull-go/logbull/internal/validation"
)

type LogBullLogger struct {
	config   *Config
	sender   *Sender
	minLevel LogLevel
	context  map[string]any
	mu       sync.RWMutex
}

func NewLogger(config Config) (*LogBullLogger, error) {
	config.ProjectID = strings.TrimSpace(config.ProjectID)
	config.Host = strings.TrimSpace(config.Host)
	config.APIKey = strings.TrimSpace(config.APIKey)

	if config.LogLevel == "" {
		config.LogLevel = INFO
	}

	if err := validation.ValidateProjectID(config.ProjectID); err != nil {
		return nil, err
	}

	if err := validation.ValidateHostURL(config.Host); err != nil {
		return nil, err
	}

	if config.APIKey != "" {
		if err := validation.ValidateAPIKey(config.APIKey); err != nil {
			return nil, err
		}
	}

	sender, err := NewSender(&config)
	if err != nil {
		return nil, err
	}

	return &LogBullLogger{
		config:   &config,
		sender:   sender,
		minLevel: config.LogLevel,
		context:  make(map[string]any),
	}, nil
}

func (l *LogBullLogger) Debug(message string, fields map[string]any) {
	l.log(DEBUG, message, fields)
}

func (l *LogBullLogger) Info(message string, fields map[string]any) {
	l.log(INFO, message, fields)
}

func (l *LogBullLogger) Warning(message string, fields map[string]any) {
	l.log(WARNING, message, fields)
}

func (l *LogBullLogger) Error(message string, fields map[string]any) {
	l.log(ERROR, message, fields)
}

func (l *LogBullLogger) Critical(message string, fields map[string]any) {
	l.log(CRITICAL, message, fields)
}

func (l *LogBullLogger) WithContext(context map[string]any) *LogBullLogger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	mergedContext := formatting.MergeFields(l.context, context)

	return &LogBullLogger{
		config:   l.config,
		sender:   l.sender,
		minLevel: l.minLevel,
		context:  mergedContext,
	}
}

func (l *LogBullLogger) Flush() {
	l.sender.Flush()
}

func (l *LogBullLogger) Shutdown() {
	l.sender.Shutdown()
}

func (l *LogBullLogger) log(level LogLevel, message string, fields map[string]any) {
	if level.Priority() < l.minLevel.Priority() {
		return
	}

	if err := validation.ValidateLogMessage(message); err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: invalid log message: %v\n", err)
		return
	}

	if err := validation.ValidateLogFields(fields); err != nil {
		fmt.Fprintf(os.Stderr, "LogBull: invalid log fields: %v\n", err)
		return
	}

	l.mu.RLock()
	mergedFields := formatting.MergeFields(l.context, fields)
	l.mu.RUnlock()

	entry := LogEntry{
		Level:     level.String(),
		Message:   formatting.FormatMessage(message),
		Timestamp: GenerateUniqueTimestamp(),
		Fields:    formatting.EnsureFields(mergedFields),
	}

	l.printToConsole(entry)
	l.sender.AddLog(entry)
}

func (l *LogBullLogger) printToConsole(entry LogEntry) {
	output := fmt.Sprintf("[%s] [%s] %s", entry.Timestamp, entry.Level, entry.Message)

	if len(entry.Fields) > 0 {
		var fields []string
		for k, v := range entry.Fields {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		output += fmt.Sprintf(" (%s)", strings.Join(fields, ", "))
	}

	if entry.Level == "ERROR" || entry.Level == "CRITICAL" {
		fmt.Fprintln(os.Stderr, output)
	} else {
		fmt.Println(output)
	}
}
