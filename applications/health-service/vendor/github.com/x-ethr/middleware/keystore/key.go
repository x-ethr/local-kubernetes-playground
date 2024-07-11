package keystore

// Key represents a constant context-key string value.
type Key string

// String represents the string value of the key. When using with [context.Context], do not use
// the string representation.
func (k Key) String() string {
	return string(k)
}

// Store represents the interface that providers all package-specific context, context keys.
type Store interface {
	// Path represents the context.Context key: "path". See [path.Implementation] for the middleware.
	Path() Key

	// Service represents the context.Context key: "service". See [name.Implementation] for the middleware.
	Service() Key

	// Version represents the context.Context key: "version". See [versioning.Implementation] for the middleware.
	//
	//   - Used for configuring middleware that adds versioning information to both context keys and response headers.
	Version() Key

	// Server represents the context.Context key: "server". See [servername.Implementation] for the middleware.
	//
	//   - Used for configuring middleware that sets the "Server" response header.
	Server() Key

	// Logs represents the context.Context key: "logs". See [logs.Implementation] for the middleware.
	//
	//   - Used for configuring middleware that sets a [log/slog] context value.
	Logs() Key

	// Timeout represents the context.Context key: "timeout". See [timeout.Implementation] for the middleware.
	//
	//   - Used for configuring middleware that sets the "Server" response header.
	Timeout() Key

	// Envoy represents the context.Context key: "envoy". See [envoy.Implementation] for the middleware.
	//
	//	- Used for storing headers in middleware:
	//		- X-Envoy-Original-Path
	//		- X-Envoy-Internal
	//		- X-Envoy-Attempt-Count
	Envoy() Key

	// Tracer represents the context.Context key: "tracer". See [tracing.Implementation] for the middleware.
	Tracer() Key

	// State represents the context.Context key: "state". See [state.Implementation] for the middleware.
	State() Key

	// CORS represents the context.Context key: "cors". See [cors.Implementation] for the middleware.
	CORS() Key

	// RIP represents the context.Context key: "real-ip". See [rip.Implementation] for the middleware.
	RIP() Key

	// Telemetry represents the context.Context key: "telemetry". See [telemetry.Implementation] for the middleware.
	Telemetry() Key

	// Authentication adds jwt-enforced middleware and a context key containing JWT-related claims data. See [authentication.Implementation] for the middleware.
	Authentication() Key
}

type store struct{}

func (s store) Path() Key {
	return "path"
}

func (s store) Service() Key {
	return "service"
}

func (s store) Version() Key {
	return "version"
}

func (s store) Server() Key {
	return "server"
}

func (s store) Timeout() Key {
	return "timeout"
}

func (s store) Envoy() Key { return "envoy" }

func (s store) Tracer() Key { return "tracer" }

func (s store) State() Key { return "state" }

func (s store) Logs() Key { return "logs" }

func (s store) CORS() Key { return "cors" }

func (s store) RIP() Key { return "real-ip" }

func (s store) Telemetry() Key { return "telemetry" }

func (s store) Authentication() Key { return "authentication" }

var s = store{}

func Keys() Store {
	return s
}
