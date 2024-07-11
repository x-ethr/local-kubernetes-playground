package logging

import (
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/x-ethr/levels"
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
			variable = levels.Trace
		case "DEBUG":
			variable = levels.Debug
		case "INFO", "LOG", "INFORMATION":
			variable = levels.Info
		case "WARN", "WARNING":
			variable = levels.Warn
		case "ERROR", "EXCEPTION":
			variable = levels.Error
		case "FATAL":
			variable = levels.Fatal
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
