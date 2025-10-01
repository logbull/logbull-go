package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/logbull/logbull-go/logbull/core"
	"github.com/logbull/logbull-go/logbull/internal/formatting"
	"github.com/logbull/logbull-go/logbull/internal/validation"
)

type SlogHandler struct {
	config *core.Config
	sender *core.Sender
	attrs  []slog.Attr
	group  string
}

func NewSlogHandler(config core.Config) (*SlogHandler, error) {
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

	return &SlogHandler{
		config: &config,
		sender: sender,
		attrs:  []slog.Attr{},
		group:  "",
	}, nil
}

func (h *SlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	logbullLevel := convertSlogLevel(level)
	return logbullLevel.Priority() >= h.config.LogLevel.Priority()
}

func (h *SlogHandler) Handle(_ context.Context, record slog.Record) error {
	level := convertSlogLevel(record.Level)
	message := record.Message

	fields := make(map[string]any)

	for _, attr := range h.attrs {
		h.addAttrToFields(fields, attr, h.group)
	}

	record.Attrs(func(attr slog.Attr) bool {
		h.addAttrToFields(fields, attr, h.group)
		return true
	})

	entry := core.LogEntry{
		Level:     level.String(),
		Message:   formatting.FormatMessage(message),
		Timestamp: core.GenerateUniqueTimestamp(),
		Fields:    formatting.EnsureFields(fields),
	}

	h.sender.AddLog(entry)
	return nil
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &SlogHandler{
		config: h.config,
		sender: h.sender,
		attrs:  newAttrs,
		group:  h.group,
	}
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return &SlogHandler{
		config: h.config,
		sender: h.sender,
		attrs:  h.attrs,
		group:  name,
	}
}

func (h *SlogHandler) Flush() {
	h.sender.Flush()
}

func (h *SlogHandler) Shutdown() {
	h.sender.Shutdown()
}

func (h *SlogHandler) addAttrToFields(fields map[string]any, attr slog.Attr, group string) {
	key := attr.Key
	if group != "" {
		key = group + "." + key
	}

	value := attr.Value.Any()
	fields[key] = value
}

func convertSlogLevel(level slog.Level) core.LogLevel {
	switch {
	case level < slog.LevelInfo:
		return core.DEBUG
	case level < slog.LevelWarn:
		return core.INFO
	case level < slog.LevelError:
		return core.WARNING
	default:
		return core.ERROR
	}
}
