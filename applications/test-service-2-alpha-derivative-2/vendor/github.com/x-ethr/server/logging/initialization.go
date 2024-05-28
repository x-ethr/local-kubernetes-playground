package logging

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

func init() { // USER > LOG_LEVEL > CI > DEFAULT
	var handler = "text"
	if v := os.Getenv("LOGGER"); v != "" {
		switch v {
		case "json":
			handler = "json"
		case "text":
			handler = "text"
		default:
			log.Println("WARNING: invalid LOGGER environment variable - must be \"json\" or \"text\". Defaulting to \"text\"")
		}
	}

	logger.Store(handler)

	Verbose(false)
	if v := strings.ToLower(os.Getenv("VERBOSE")); v == "enabled" || v == "true" || v == "1" {
		Verbose(true)
	}

	var variable = slog.LevelDebug
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" || v == "yes" {
		variable = slog.LevelInfo
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch value := strings.ToUpper(v); value {
		case "TRACE":
			variable = Trace
		case "DEBUG":
			variable = Debug
		case "INFO", "LOG", "INFORMATION":
			variable = Info
		case "WARN", "WARNING":
			variable = Warn
		case "ERROR", "EXCEPTION":
			variable = Error
		}
	}

	l.Store(variable)
	if previous := slog.SetLogLoggerLevel(variable); previous != slog.LevelInfo {
		if verbosity.Load() != nil && *(verbosity.Load()) {
			slog.Warn("User-Defined Global, Non-Default Log Level Found - Reverting Change(s) to Level")
		}

		l.Store(previous)
		slog.SetLogLoggerLevel(previous)
	}

}
