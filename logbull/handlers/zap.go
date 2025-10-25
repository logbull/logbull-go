package handlers

import (
	"strings"

	"go.uber.org/zap/zapcore"

	"github.com/logbull/logbull-go/logbull/core"
	"github.com/logbull/logbull-go/logbull/internal/formatting"
	"github.com/logbull/logbull-go/logbull/internal/validation"
)

type ZapCore struct {
	config   *core.Config
	sender   *core.Sender
	fields   []zapcore.Field
	minLevel zapcore.Level
}

func NewZapCore(config core.Config) (*ZapCore, error) {
	config.ProjectID = strings.TrimSpace(config.ProjectID)
	config.Host = strings.TrimSpace(config.Host)
	config.APIKey = strings.TrimSpace(config.APIKey)

	if config.LogLevel == "" {
		config.LogLevel = core.INFO
	}

	// Check if credentials are provided
	if config.ProjectID == "" || config.Host == "" {
		// No credentials: do nothing (Zap will print)
		println(
			"LogBull: No credentials provided for ZapCore. Handler is disabled. Logs will not be sent to LogBull server.",
		)
		return &ZapCore{
			config:   &config,
			sender:   nil,
			fields:   []zapcore.Field{},
			minLevel: convertLogLevelToZap(config.LogLevel),
		}, nil
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

	return &ZapCore{
		config:   &config,
		sender:   sender,
		fields:   []zapcore.Field{},
		minLevel: convertLogLevelToZap(config.LogLevel),
	}, nil
}

func (z *ZapCore) Enabled(level zapcore.Level) bool {
	return level >= z.minLevel
}

func (z *ZapCore) With(fields []zapcore.Field) zapcore.Core {
	newFields := make([]zapcore.Field, len(z.fields)+len(fields))
	copy(newFields, z.fields)
	copy(newFields[len(z.fields):], fields)

	return &ZapCore{
		config:   z.config,
		sender:   z.sender,
		fields:   newFields,
		minLevel: z.minLevel,
	}
}

func (z *ZapCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if z.Enabled(entry.Level) {
		return ce.AddCore(entry, z)
	}
	return ce
}

func (z *ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// If handler is disabled, do nothing
	if z.sender == nil {
		return nil
	}

	allFields := make([]zapcore.Field, len(z.fields)+len(fields))
	copy(allFields, z.fields)
	copy(allFields[len(z.fields):], fields)

	extractedFields := z.extractFields(allFields)

	logEntry := core.LogEntry{
		Level:     convertZapLevel(entry.Level).String(),
		Message:   formatting.FormatMessage(entry.Message),
		Timestamp: core.GenerateUniqueTimestamp(),
		Fields:    formatting.EnsureFields(extractedFields),
	}

	z.sender.AddLog(logEntry)
	return nil
}

func (z *ZapCore) Sync() error {
	if z.sender != nil {
		z.sender.Flush()
	}
	return nil
}

func (z *ZapCore) Shutdown() {
	if z.sender != nil {
		z.sender.Shutdown()
	}
}

func (z *ZapCore) extractFields(fields []zapcore.Field) map[string]any {
	result := make(map[string]any)
	enc := zapcore.NewMapObjectEncoder()

	for _, field := range fields {
		field.AddTo(enc)
	}

	for key, value := range enc.Fields {
		result[key] = value
	}

	return result
}

func convertZapLevel(level zapcore.Level) core.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return core.DEBUG
	case zapcore.InfoLevel:
		return core.INFO
	case zapcore.WarnLevel:
		return core.WARNING
	case zapcore.ErrorLevel:
		return core.ERROR
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return core.CRITICAL
	default:
		return core.INFO
	}
}

func convertLogLevelToZap(level core.LogLevel) zapcore.Level {
	switch level {
	case core.DEBUG:
		return zapcore.DebugLevel
	case core.INFO:
		return zapcore.InfoLevel
	case core.WARNING:
		return zapcore.WarnLevel
	case core.ERROR:
		return zapcore.ErrorLevel
	case core.CRITICAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
