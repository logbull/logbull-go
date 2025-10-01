// Package logbull provides a Go client library for sending logs to LogBull server.
//
// LogBull supports multiple integration patterns:
//   - Standalone logger with LogBullLogger
//   - Standard library slog integration with SlogHandler
//   - Uber-go zap integration with ZapCore
//   - Sirupsen logrus integration with LogrusHook
//
// All components support asynchronous log sending with automatic batching,
// context management, and thread-safe operations.
package logbull

import (
	"github.com/logbull/logbull-go/logbull/core"
	"github.com/logbull/logbull-go/logbull/handlers"
)

type (
	Config        = core.Config
	LogLevel      = core.LogLevel
	LogEntry      = core.LogEntry
	LogBullLogger = core.LogBullLogger
	SlogHandler   = handlers.SlogHandler
	ZapCore       = handlers.ZapCore
	LogrusHook    = handlers.LogrusHook
)

const (
	DEBUG    = core.DEBUG
	INFO     = core.INFO
	WARNING  = core.WARNING
	ERROR    = core.ERROR
	CRITICAL = core.CRITICAL
)

var (
	NewLogger      = core.NewLogger
	NewSlogHandler = handlers.NewSlogHandler
	NewZapCore     = handlers.NewZapCore
	NewLogrusHook  = handlers.NewLogrusHook
)
