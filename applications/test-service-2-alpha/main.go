package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"maps"
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
	"github.com/x-ethr/server/middleware/tracing"
	"github.com/x-ethr/server/middleware/versioning"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// header is a dynamically linked string value - defaults to "server" - which represents the server name.
var header string = "server"

// service is a dynamically linked string value - defaults to "service" - which represents the service name.
var service string = "service"

// version is a dynamically linked string value - defaults to "development" - which represents the service's version.
var version string = "development" // production builds have version dynamically linked

// ctx, cancel represent the server's runtime context and cancellation handler.
var ctx, cancel = context.WithCancel(context.Background())

// port represents a cli flag that sets the server listening port
var port = flag.String("port", "8080", "Server Listening Port.")

var (
	tracer = otel.Tracer(service)
)

var logger *slog.Logger

func main() {
	middlewares := server.Middleware()

	middlewares.Add(middleware.New().Path().Middleware)
	middlewares.Add(middleware.New().Envoy().Middleware)
	middlewares.Add(middleware.New().Timeout().Configuration(func(options *timeout.Settings) { options.Timeout = 30 * time.Second }).Middleware)
	middlewares.Add(middleware.New().Server().Configuration(func(options *servername.Settings) { options.Server = header }).Middleware)
	middlewares.Add(middleware.New().Service().Configuration(func(options *name.Settings) { options.Service = service }).Middleware)
	middlewares.Add(middleware.New().Version().Configuration(func(options *versioning.Settings) { options.Version.Service = version }).Middleware)
	middlewares.Add(middleware.New().Telemetry().Middleware)

	middlewares.Add(middleware.New().Tracer().Configuration(func(options *tracing.Settings) { options.Tracer = tracer }).Middleware)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.Health)

	mux.Handle("GET /", otelhttp.WithRouteTag("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const name = "metadata"

		ctx := r.Context()

		service := middleware.New().Service().Value(ctx)
		ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

		defer span.End()

		channel := make(chan map[string]interface{}, 1)

		var process = func(ctx context.Context, c chan map[string]interface{}) {
			var a, b map[string]interface{}

			client := otelhttp.DefaultClient

			{
				request, e := http.NewRequestWithContext(ctx, "GET", "http://test-service-2-alpha-derivative-1.development.svc.cluster.local:8080", nil)
				if e != nil {
					slog.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
					http.Error(w, "error while generating GET request to internal service", http.StatusInternalServerError)
					return
				}

				headers.Propagate(r, request)

				// response, e := otelhttp.Get(ctx, "http://test-service-2-alpha-derivative-1.development.svc.cluster.local:8080")
				response, e := client.Do(request)
				if e != nil {
					slog.ErrorContext(ctx, "Error Generating Response from Internal Service", slog.String("error", e.Error()))
					http.Error(w, "exception generating response from internal service", http.StatusInternalServerError)
					return
				}

				if e := json.NewDecoder(response.Body).Decode(&a); e != nil {
					http.Error(w, "unable to decode response body to normalized data structure", http.StatusInternalServerError)
					return
				}
			}

			{
				request, e := http.NewRequestWithContext(ctx, "GET", "http://test-service-2-alpha-derivative-2.development.svc.cluster.local:8080", nil)
				if e != nil {
					slog.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
					http.Error(w, "error while generating GET request to internal service", http.StatusInternalServerError)
					return
				}

				headers.Propagate(r, request)

				// response, e := otelhttp.Get(ctx, "http://test-service-2-alpha-derivative-2.development.svc.cluster.local:8080")
				response, e := client.Do(request)
				if e := json.NewDecoder(response.Body).Decode(&b); e != nil {
					http.Error(w, "unable to decode response body to normalized data structure", http.StatusInternalServerError)
					return
				}

				defer response.Body.Close()
			}

			var response = make(map[string]interface{})

			maps.Copy(response, a)
			maps.Copy(response, b)

			var payload = map[string]interface{}{
				middleware.New().Service().Value(ctx): map[string]interface{}{
					"service": middleware.New().Service().Value(ctx),
					"version": middleware.New().Version().Value(ctx).Service,

					"response": response,
				},
			}

			c <- payload
		}

		go process(ctx, channel)

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
	})))

	// Start the HTTP server
	slog.Info("Starting Server ...", slog.String("local", fmt.Sprintf("http://localhost:%s", *(port))))

	api := server.Server(ctx, mux, middlewares, *port)

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
