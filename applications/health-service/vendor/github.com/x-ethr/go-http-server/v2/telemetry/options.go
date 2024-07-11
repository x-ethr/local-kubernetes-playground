package telemetry

import (
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
)

type Zipkin struct {
	// URL - Zipkin collector url - defaults to "http://opentelemetry-collector.observability.svc.cluster.local:9441".
	URL string

	// Enabled will enable the Zipkin collector. Default is true.
	Enabled bool
}

type Tracer struct {
	// Options represents [otlptracehttp.Option] configurations.
	//
	// Defaults:
	//
	// 	- otlptracehttp.WithInsecure()
	// 	- otlptracehttp.WithEndpoint("opentelemetry-collector.observability.svc.cluster.local:4318")
	Options []otlptracehttp.Option

	// Debugger configures an additional [stdouttrace.Exporter] if not nil. Defaults nil.
	Debugger *stdouttrace.Exporter

	// Local will prevent an external tracer from getting used as a provider. If true, forces [Tracer.Debugger] configuration. Default is false.
	Local bool
}

type Metrics struct {
	// Options represents [otlpmetrichttp.Option] configurations.
	//
	// Defaults:
	//
	// 	- otlpmetrichttp.WithInsecure()
	//	- otlpmetrichttp.WithEndpoint("opentelemetry-collector.observability.svc.cluster.local:4318")
	Options []otlpmetrichttp.Option

	// Debugger configures an additional [metric.Exporter] if not nil. Defaults nil.
	Debugger metric.Exporter

	// Local will prevent an external metrics provider from getting used. If true, forces [Metrics.Debugger] configuration. Default is false.
	Local bool
}

type Logs struct {
	// Logs represents [otlploghttp.Option] configurations.
	//
	// Defaults:
	//
	// 	- otlploghttp.WithInsecure()
	// 	- otlploghttp.WithEndpoint("http://zipkin.istio-system.svc.cluster.local:9411")
	Options []otlploghttp.Option

	// Debugger configures an additional [stdoutlog.Exporter] if not nil. Defaults nil.
	Debugger *stdoutlog.Exporter

	// Local will prevent an external log exporter from getting used as a processor. If true, forces [Logs.Debugger] configuration. Default is false.
	Local bool
}

type Settings struct {
	// Zipkin represents a zipkin collector.
	Zipkin *Zipkin

	// Tracer represents [otlptracehttp.Option] configurations.
	Tracer *Tracer

	// Metrics represents [otlpmetrichttp.Option] configurations.
	Metrics *Metrics

	// Logs represents [otlploghttp.Option] configurations.
	Logs *Logs

	// Propagators ...
	//
	// Defaults:
	//
	//	- [propagation.TraceContext]
	//	- [propagation.Baggage]
	Propagators []propagation.TextMapPropagator
}

type Variadic func(options *Settings)

func Options() *Settings {
	return &Settings{
		Zipkin: &Zipkin{
			URL:     "http://zipkin.istio-system.svc.cluster.local:9411",
			Enabled: true,
		},
		Metrics: &Metrics{
			Options: []otlpmetrichttp.Option{
				otlpmetrichttp.WithInsecure(),
				otlpmetrichttp.WithEndpoint("opentelemetry-collector.observability.svc.cluster.local:4318"),
			},
		},
		Tracer: &Tracer{
			Options: []otlptracehttp.Option{
				otlptracehttp.WithInsecure(),
				otlptracehttp.WithEndpoint("opentelemetry-collector.observability.svc.cluster.local:4318"),
			},
		},
		Logs: &Logs{
			Options: []otlploghttp.Option{
				otlploghttp.WithInsecure(),
				otlploghttp.WithEndpoint("opentelemetry-collector.observability.svc.cluster.local:4318"),
			},
		},
		Propagators: []propagation.TextMapPropagator{
			propagation.TraceContext{},
			propagation.Baggage{},
		},
	}
}
