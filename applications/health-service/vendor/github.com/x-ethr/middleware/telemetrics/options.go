package telemetrics

import (
	"log/slog"

	"github.com/x-ethr/middleware/types"
)

type Settings struct {
	// Headers includes telemetry-specific header(s) to store in a context key as derived from an http(s) request.
	//
	// Default(s):
	//
	// 	- "portal",
	// 	- "device",
	// 	- "user",
	// 	- "travel",
	// 	- "traceparent",
	// 	- "tracestate",
	// 	- "x-cloud-trace-context",
	// 	- "sw8",
	// 	- "user-agent",
	// 	- "cookie",
	// 	- "authorization",
	// 	- "jwt",
	// 	- "x-request-id",
	// 	- "x-b3-traceid",
	// 	- "x-b3-spanid",
	// 	- "x-b3-parentspanid",
	// 	- "x-b3-sampled",
	// 	- "x-b3-flags",
	// 	- "x-ot-span-context",
	// 	- "x-api-version",
	Headers []string

	Level slog.Leveler // Level represents a [log/slog] log level - defaults to (slog.LevelDebug - 4) (trace)
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{
		Level: (slog.LevelDebug - 4),
		Headers: []string{
			"portal",
			"device",
			"user",
			"travel",
			"traceparent",
			"tracestate",
			"x-cloud-trace-context",
			"sw8",
			"user-agent",
			"cookie",
			"authorization",
			"jwt",
			"x-request-id",
			"x-b3-traceid",
			"x-b3-spanid",
			"x-b3-parentspanid",
			"x-b3-sampled",
			"x-b3-flags",
			"x-ot-span-context",
			"x-api-version",
		},
	}
}
