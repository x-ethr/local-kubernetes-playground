package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/x-ethr/server"
	"github.com/x-ethr/server/logging"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/middleware/name"
	"github.com/x-ethr/server/middleware/servername"
	"github.com/x-ethr/server/middleware/timeout"
	"github.com/x-ethr/server/middleware/versioning"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/otel"

	"authorization-service/internal/api/user/create"
)

// header is a dynamically linked string value - defaults to "server" - which represents the server name.
var header string = "server"

// service is a dynamically linked string value - defaults to "service" - which represents the service name.
var service string = "service"

// version is a dynamically linked string value - defaults to "development" - which represents the service's version.
var version string = "development" // production builds have version dynamically linked

var prefix = map[string]string{
	(version): "v1", // default version prefix
}

// ctx, cancel represent the server's runtime context and cancellation handler.
var ctx, cancel = context.WithCancel(context.Background())

// port represents a cli flag that sets the server listening port
var port = flag.String("port", "8080", "Server Listening Port.")

var logger *slog.Logger

var (
	tracer = otel.Tracer(service)
)

func main() {
	// Create an instance of the custom handler
	mux := server.New()

	mux.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "logger", logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	mux.Middleware(middleware.New().Path().Middleware)
	mux.Middleware(middleware.New().Envoy().Middleware)
	mux.Middleware(middleware.New().Timeout().Configuration(func(options *timeout.Settings) { options.Timeout = 30 * time.Second }).Middleware)
	mux.Middleware(middleware.New().Server().Configuration(func(options *servername.Settings) { options.Server = header }).Middleware)
	mux.Middleware(middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = service }).Middleware)
	mux.Middleware(middleware.New().Version().Configuration(func(options *versioning.Settings) { options.Version.Service = version }).Middleware)
	mux.Middleware(middleware.New().Telemetry().Middleware)

	mux.Register("POST /users", create.Handler)

	mux.Register("GET /health", server.Health)

	// Start the HTTP server
	slog.Info("Starting Server ...", slog.String("local", fmt.Sprintf("http://localhost:%s", *(port))))

	api := server.Server(ctx, mux, *port)

	// Issue Cancellation Handler
	server.Interrupt(ctx, cancel, api)

	// Telemetry Setup
	shutdown, e := telemetry.Setup(ctx, service, version, func(options *telemetry.Settings) {
		if version == "development" && os.Getenv("CI") == "" {
			options.Zipkin.Enabled = false

			options.Tracer.Local = true
			options.Metrics.Local = true
			options.Logs.Local = true
		}
	})
	if e != nil {
		panic(e)
	}

	defer func() {
		e = errors.Join(e, shutdown(ctx))
	}()

	// <-- Blocking
	if e := api.ListenAndServe(); e != nil && !(errors.Is(e, http.ErrServerClosed)) {
		slog.ErrorContext(ctx, "Error During Server's Listen & Serve Call ...", slog.String("error", e.Error()))

		os.Exit(100)
	}

	// --> Exit
	{
		slog.InfoContext(ctx, "Graceful Shutdown Complete")

		// Waiter
		<-ctx.Done()
	}
}

func init() {
	flag.Parse()

	level := slog.Level(-8)
	if os.Getenv("CI") == "true" {
		level = slog.LevelDebug
	}

	logging.Level(level)
	slog.SetLogLoggerLevel(level)
	if service == "service" && os.Getenv("CI") != "true" {
		_, file, _, ok := runtime.Caller(0)
		if ok {
			service = filepath.Base(filepath.Dir(file))
		}
	}

	handler := logging.Logger(func(o *logging.Options) { o.Service = service })
	logger = slog.New(handler)
	slog.SetDefault(logger)
}