package middleware

import (
	"net/http"

	"github.com/x-ethr/middleware/authentication"
	"github.com/x-ethr/middleware/cors"
	"github.com/x-ethr/middleware/envoy"
	"github.com/x-ethr/middleware/logs"
	"github.com/x-ethr/middleware/name"
	"github.com/x-ethr/middleware/path"
	"github.com/x-ethr/middleware/rip"
	"github.com/x-ethr/middleware/servername"
	"github.com/x-ethr/middleware/state"
	"github.com/x-ethr/middleware/telemetrics"
	"github.com/x-ethr/middleware/timeout"
	"github.com/x-ethr/middleware/tracing"
	"github.com/x-ethr/middleware/versioning"
)

type generic struct{}

func (*generic) Path() path.Implementation {
	return path.New()
}

func (*generic) Version() versioning.Implementation {
	return versioning.New()
}

func (*generic) Service() name.Implementation {
	return name.New()
}

func (*generic) Server() servername.Implementation {
	return servername.New()
}

func (*generic) Timeout() timeout.Implementation {
	return timeout.New()
}

func (*generic) Envoy() envoy.Implementation {
	return envoy.New()
}

func (*generic) Tracer() tracing.Implementation {
	return tracing.New()
}

func (*generic) State() state.Implementation {
	return state.New()
}

func (*generic) Logs() logs.Implementation {
	return logs.New()
}

func (*generic) CORS() cors.Implementation {
	return cors.New()
}

func (*generic) RIP() rip.Implementation {
	return rip.New()
}

func (*generic) Telemetry() telemetrics.Implementation {
	return telemetrics.New()
}

func (*generic) Authentication() authentication.Implementation {
	return authentication.New()
}

type Interface interface {
	Path() path.Implementation                     // Path - See the [path] package for additional details.
	Version() versioning.Implementation            // Version - See the [versioning] package for additional details.
	Service() name.Implementation                  // Service - See the [name] package for additional details.
	Server() servername.Implementation             // Server - See the [servername] package for additional details.
	Timeout() timeout.Implementation               // Timeout - See the [timeout] package for additional details.
	Envoy() envoy.Implementation                   // Envoy - See the [envoy] package for additional details.
	Tracer() tracing.Implementation                // Tracer - See the [tracing] package for additional details.
	State() state.Implementation                   // State - See the [state] package for additional details.
	Logs() logs.Implementation                     // Logs - See the [logs] package for additional details.
	CORS() cors.Implementation                     // CORS - See the [cors] package for additional details.
	RIP() rip.Implementation                       // RIP - See the [rip] package for additional details.
	Telemetry() telemetrics.Implementation         // Telemetry - See the [telemetrics] package for additional details.
	Authentication() authentication.Implementation // Authentication - See the [authentication] package for additional details.
}

var v = &generic{}

func New() Interface {
	return v
}

type Middlewares struct {
	middleware []func(http.Handler) http.Handler
}

func (m *Middlewares) Add(middlewares ...func(http.Handler) http.Handler) {
	if len(middlewares) == 0 {
		return
	}

	m.middleware = append(m.middleware, middlewares...)
}

func (m *Middlewares) Handler(parent http.Handler) (handler http.Handler) {
	var length = len(m.middleware)
	if length == 0 {
		return parent
	}

	// Wrap the end handler with the middleware chain
	handler = m.middleware[len(m.middleware)-1](parent)
	for i := len(m.middleware) - 2; i >= 0; i-- {
		handler = m.middleware[i](handler)
	}

	return
}

func Middleware() *Middlewares {
	return &Middlewares{
		middleware: make([]func(http.Handler) http.Handler, 0),
	}
}
