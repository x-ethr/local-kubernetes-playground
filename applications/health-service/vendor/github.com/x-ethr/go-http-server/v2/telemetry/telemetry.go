package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

func Resources(ctx context.Context, service, version string) *resource.Resource {
	options := []resource.Option{
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),         // Discover and provide host information.
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithContainer(),
		resource.WithContainerID(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(service),
			semconv.ServiceNamespaceKey.String(os.Getenv("NAMESPACE")),
			semconv.ServiceVersionKey.String(version),
		),
	}

	instance, e := resource.New(ctx, options...)
	if errors.Is(e, resource.ErrPartialResource) || errors.Is(e, resource.ErrSchemaURLConflict) {
		slog.WarnContext(ctx, "Non-Fatal Open-Telemetry Error", slog.String("error", e.Error()))
	} else if e != nil {
		exception := fmt.Errorf("unable to generate exportable resource: %w", e)
		slog.WarnContext(ctx, "Fatal Open-Telemetry Error", slog.String("error", exception.Error()))
		panic(exception)
	}

	return instance
}

// Setup bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func Setup(ctx context.Context, service, version string, options ...Variadic) (shutdown func(context.Context) error, err error) {
	o := Options()
	for _, option := range options {
		option(o)
	}

	var shutdowns []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdowns {
			err = errors.Join(err, fn(ctx))
		}
		shutdowns = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	catch := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	propagation := propagator(o)
	otel.SetTextMapPropagator(propagation)

	// Set up trace provider.
	tracer, err := traces(ctx, service, version, o)
	if err != nil {
		catch(err)
		return
	}
	shutdowns = append(shutdowns, tracer.Shutdown)
	otel.SetTracerProvider(tracer)

	// Set up meter provider.
	meter, err := metrics(ctx, o)
	if err != nil {
		catch(err)
		return
	}
	shutdowns = append(shutdowns, meter.Shutdown)
	otel.SetMeterProvider(meter)

	// Set up logger provider.
	logger, err := logexporter(ctx, o)
	if err != nil {
		catch(err)
		return
	}

	shutdowns = append(shutdowns, logger.Shutdown)
	global.SetLoggerProvider(logger)

	return
}

func propagator(settings *Settings) propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(settings.Propagators...)
}

func resources(ctx context.Context, service, version string, settings *Settings) (*resource.Resource, error) {
	options := []resource.Option{
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),         // Discover and provide host information.
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithContainer(),
		resource.WithContainerID(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(service),
			semconv.ServiceNamespaceKey.String(os.Getenv("NAMESPACE")),
			semconv.ServiceVersionKey.String(version),
		),
	}

	instance, e := resource.New(ctx, options...)
	if errors.Is(e, resource.ErrPartialResource) || errors.Is(e, resource.ErrSchemaURLConflict) {
		slog.WarnContext(ctx, "Non-Fatal Open-Telemetry Error", slog.String("error", e.Error()))
	} else if e != nil {
		exception := fmt.Errorf("unable to generate resource: %w", e)
		return nil, exception
	}

	return instance, nil
}

func traces(ctx context.Context, service, version string, settings *Settings) (*trace.TracerProvider, error) {
	options := make([]trace.TracerProviderOption, 0)
	options = append(options, trace.WithSampler(trace.AlwaysSample()))

	resources, e := resources(ctx, service, version, settings)
	if e != nil {
		exception := fmt.Errorf("error while attempting to generate resources from traces function: %w", e)
		return nil, exception
	}

	options = append(options, trace.WithResource(resources))

	if settings.Tracer.Local && settings.Tracer.Debugger == nil {
		var e error
		settings.Tracer.Debugger, e = stdouttrace.New(stdouttrace.WithoutTimestamps(), stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stdout))
		if e != nil {
			exception := fmt.Errorf("unable to instantiate local tracer: %w", e)
			return nil, exception
		}
	} else {
		exporter, e := otlptracehttp.New(ctx, settings.Tracer.Options...)
		if e != nil {
			return nil, e
		}

		options = append(options, trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second*30)))

		if settings.Zipkin.Enabled {
			z, e := zipkin.New(settings.Zipkin.URL)
			if e != nil {
				return nil, e
			}

			options = append(options, trace.WithBatcher(z, trace.WithBatchTimeout(time.Second*30)))
		}
	}

	provider := trace.NewTracerProvider(options...)

	return provider, nil
}

func metrics(ctx context.Context, settings *Settings) (*metric.MeterProvider, error) {
	// metricExporter, err := otlpmetrichttp.New(ctx, settings.Metrics.Options...)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// meterProvider := metric.NewMeterProvider(
	// 	metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(30*time.Second))),
	// )
	// return meterProvider, nil

	options := make([]metric.Option, 0)

	if settings.Metrics.Local && settings.Metrics.Debugger == nil {
		var e error
		settings.Metrics.Debugger, e = stdoutmetric.New()
		if e != nil {
			exception := fmt.Errorf("unable to instantiate local metrics exporter: %w", e)
			return nil, exception
		}
	} else {
		exporter, e := otlpmetrichttp.New(ctx, settings.Metrics.Options...)
		if e != nil {
			exception := fmt.Errorf("unable to instantiate primary metrics exporter: %w", e)
			return nil, exception
		}

		options = append(options, metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(30*time.Second))))
	}

	if settings.Metrics.Debugger != nil {
		options = append(options, metric.WithReader(metric.NewPeriodicReader(settings.Metrics.Debugger, metric.WithInterval(10*time.Second))))
	}

	provider := metric.NewMeterProvider(options...)

	return provider, nil
}

func logexporter(ctx context.Context, settings *Settings) (*log.LoggerProvider, error) {
	options := make([]log.LoggerProviderOption, 0)

	if settings.Logs.Local {
		if settings.Logs.Debugger == nil {
			var e error
			settings.Logs.Debugger, e = stdoutlog.New()
			if e != nil {
				exception := fmt.Errorf("unable to instantiate local log exporter: %w", e)
				return nil, exception
			}
		}
	} else {
		exporter, e := otlploghttp.New(ctx, settings.Logs.Options...)
		if e != nil {
			exception := fmt.Errorf("unable to instantiate primary log exporter: %w", e)
			return nil, exception
		}

		options = append(options, log.WithProcessor(log.NewBatchProcessor(exporter)))
	}

	if settings.Logs.Debugger != nil {
		options = append(options, log.WithProcessor(log.NewBatchProcessor(settings.Logs.Debugger)))
	}

	provider := log.NewLoggerProvider(options...)

	return provider, nil
}
