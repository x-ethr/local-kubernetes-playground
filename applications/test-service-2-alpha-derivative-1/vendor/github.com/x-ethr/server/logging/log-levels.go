package logging

import (
	"log/slog"
	"sync/atomic"
)

const (
	Trace = slog.LevelDebug - 4
	Debug = slog.LevelDebug
	Info  = slog.LevelInfo
	Warn  = slog.LevelWarn
	Error = slog.LevelError
)

var l atomic.Value // default logging level

// Level sets the atomic, global log-level
func Level(v slog.Level) {
	l.Store(v)
}
