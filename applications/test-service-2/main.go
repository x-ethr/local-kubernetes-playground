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
	"github.com/x-ethr/server/headers"
	"github.com/x-ethr/server/logging"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/middleware/name"
	"github.com/x-ethr/server/middleware/servername"
	"github.com/x-ethr/server/middleware/timeout"
	"github.com/x-ethr/server/middleware/versioning"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

var (
	tracer = otel.Tracer(service)
)

func main() {
	// Create an instance of the custom handler
	mux := server.New()

	mux.Middleware(middleware.New().Path().Middleware)
	mux.Middleware(middleware.New().Envoy().Middleware)
	mux.Middleware(middleware.New().Timeout().Configuration(func(options *timeout.Settings) {
		options.Timeout = 30 * time.Second
	}).Middleware)

	mux.Middleware(middleware.New().Server().Configuration(func(options *servername.Settings) {
		options.Server = header
	}).Middleware)

	mux.Middleware(middleware.New().Service().Configuration(func(options *name.Settings) {
		options.Service = service
	}).Middleware)

	mux.Middleware(middleware.New().Version().Configuration(func(options *versioning.Settings) {

		options.Version.Service = version
	}).Middleware)

	mux.Middleware(middleware.New().Telemetry().Middleware)

	mux.Register("GET /health", server.Health)

	mux.Register("GET /", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s - main", service))
		span.SetAttributes(attribute.String("workload", service))
		span.SetAttributes(telemetry.Resources(ctx, service, version).Attributes()...)
		span.SetAttributes(attribute.String("component", fmt.Sprintf("%s-%s", service, "example-component")))

		defer span.End()

		channel := make(chan map[string]interface{}, 1)
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
		}
	})

	mux.Register("GET /alpha", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		attributes := trace.WithAttributes(telemetry.Resources(ctx, service, version).Attributes()...)
		ctx, span := tracer.Start(ctx, fmt.Sprintf("%s - main", service), trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attribute.String("workload", service), attribute.String("component", fmt.Sprintf("%s-%s", service, "example-component"))), attributes)
		defer span.End()

		channel, exception := make(chan map[string]interface{}, 1), make(chan error)
		var process = func(ctx context.Context, span trace.Span, c chan map[string]interface{}) {
			client := otelhttp.DefaultClient
			request, e := http.NewRequestWithContext(ctx, "GET", "http://test-service-2-alpha.development.svc.cluster.local:8080", nil)
			if e != nil {
				slog.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
				exception <- e
				return
			}

			headers.Propagate(r, request)

			// response, e := otelhttp.Get(ctx, "http://test-service-2-alpha.development.svc.cluster.local:8080")
			response, e := client.Do(request)
			if e != nil {
				slog.ErrorContext(ctx, "Error Generating Response from Internal Service", slog.String("error", e.Error()))
				exception <- e
				return
			}

			defer response.Body.Close()

			var structure interface{}
			if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
				slog.ErrorContext(ctx, "Unable to Decode Response Body to Normalized Data Structure", slog.String("error", e.Error()))
				exception <- e
				return
			}

			path := middleware.New().Path().Value(ctx)
			var payload = map[string]interface{}{
				middleware.New().Service().Value(ctx): map[string]interface{}{
					"path":    path,
					"service": middleware.New().Service().Value(ctx),
					"version": middleware.New().Version().Value(ctx).Service,

					"response": structure,
				},
			}

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

	if service == "service" && os.Getenv("CI") != "true" {
		_, file, _, ok := runtime.Caller(0)
		if ok {
			service = filepath.Base(filepath.Dir(file))
		}
	}

	handler := logging.Logger(func(o *logging.Options) { o.Service = service })
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
