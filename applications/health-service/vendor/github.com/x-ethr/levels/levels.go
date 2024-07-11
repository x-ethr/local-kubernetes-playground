package levels

import (
	"log/slog"
)

const (
	Trace = slog.LevelDebug - 4
	Debug = slog.LevelDebug
	Info  = slog.LevelInfo
	Warn  = slog.LevelWarn
	Error = slog.LevelError
	Fatal = slog.LevelError + 4
)

// String returns the corresponding slog.Level value based on the provided level string.
// If the level string matches any of the predefined levels, the corresponding slog.Level value is returned.
// If no match is found, slog.LevelDebug is returned as the default value.
//
// Expected matches are:
//
//   - TRACE
//   - DEBUG
//   - INFO
//   - WARN
//   - ERROR
//   - FATAL
func String(level string) (value slog.Level) {
	value = slog.LevelDebug

	switch level {
	case "TRACE":
		value = Trace
	case "DEBUG":
		value = Debug
	case "INFO":
		value = Info
	case "WARN":
		value = Warn
	case "ERROR":
		value = Error
	case "FATAL":
		value = Fatal
	}

	return
}
