package handlers

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/logbull/logbull-go/logbull/core"
	"github.com/logbull/logbull-go/logbull/internal/formatting"
	"github.com/logbull/logbull-go/logbull/internal/validation"
)

type LogrusHook struct {
	config *core.Config
	sender *core.Sender
	levels []logrus.Level
}

func NewLogrusHook(config core.Config) (*LogrusHook, error) {
	config.ProjectID = strings.TrimSpace(config.ProjectID)
	config.Host = strings.TrimSpace(config.Host)
	config.APIKey = strings.TrimSpace(config.APIKey)

	if config.LogLevel == "" {
		config.LogLevel = core.INFO
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

	sender, err := core.NewSender(&config)
	if err != nil {
		return nil, err
	}

	levels := levelsFromConfig(config.LogLevel)

	return &LogrusHook{
		config: &config,
		sender: sender,
		levels: levels,
	}, nil
}

func (h *LogrusHook) Levels() []logrus.Level {
	return h.levels
}

func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	level := convertLogrusLevel(entry.Level)
	message := entry.Message

	fields := make(map[string]any)
	for key, value := range entry.Data {
		fields[key] = value
	}

	logEntry := core.LogEntry{
		Level:     level.String(),
		Message:   formatting.FormatMessage(message),
		Timestamp: core.GenerateUniqueTimestamp(),
		Fields:    formatting.EnsureFields(fields),
	}

	h.sender.AddLog(logEntry)
	return nil
}

func (h *LogrusHook) Flush() {
	h.sender.Flush()
}

func (h *LogrusHook) Shutdown() {
	h.sender.Shutdown()
}

func convertLogrusLevel(level logrus.Level) core.LogLevel {
	switch level {
	case logrus.DebugLevel, logrus.TraceLevel:
		return core.DEBUG
	case logrus.InfoLevel:
		return core.INFO
	case logrus.WarnLevel:
		return core.WARNING
	case logrus.ErrorLevel:
		return core.ERROR
	case logrus.FatalLevel, logrus.PanicLevel:
		return core.CRITICAL
	default:
		return core.INFO
	}
}

func levelsFromConfig(minLevel core.LogLevel) []logrus.Level {
	allLevels := []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	minPriority := minLevel.Priority()
	var result []logrus.Level

	for _, level := range allLevels {
		logbullLevel := convertLogrusLevel(level)
		if logbullLevel.Priority() >= minPriority {
			result = append(result, level)
		}
	}

	return result
}
