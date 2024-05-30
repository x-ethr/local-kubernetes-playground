package main

import (
	"context"
	"encoding/json"
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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/x-ethr/environment"

	"user-service/internal/api/registration"
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

	mux.Register("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s - main", service))
		span.SetAttributes(attribute.String("workload", service))
		span.SetAttributes(telemetry.Resources(ctx, service, version).Attributes()...)
		span.SetAttributes(attribute.String("component", fmt.Sprintf("%s-%s", service, "example-component")))

		defer span.End()

		channel, exception := make(chan map[string]interface{}, 1), make(chan error)
		var process = func(ctx context.Context, span trace.Span, c chan map[string]interface{}) {
			path := middleware.New().Path().Value(ctx)

			var payload = map[string]interface{}{
				middleware.New().Service().Value(ctx): map[string]interface{}{
					"path":    path,
					"service": middleware.New().Service().Value(ctx),
					"version": middleware.New().Version().Value(ctx).Service,
				},
			}

			span.SetAttributes(attribute.String("path", path))

			c <- payload
		}

		go process(ctx, span, channel)

		select {
		case <-ctx.Done():
			return
		case payload := <-channel:
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Request-ID", r.Header.Get("X-Request-ID"))
			w.WriteHeader(http.StatusOK)

			json.NewEncoder(w).Encode(payload)

			return
		case e := <-exception:
			slog.ErrorContext(ctx, "Error While Processing Request", slog.String("error", e.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

			return
		}
	})

	mux.Register("POST /register", registration.Handler)

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

	environment.Log(ctx, slog.LevelInfo)
}
