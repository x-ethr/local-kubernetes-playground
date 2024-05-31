package metadata

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/x-ethr/server/handler/output"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/headers"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func processor(r *http.Request, output chan<- *types.Response, exception chan<- *types.Exception, options *types.Options) {
	const name = "metadata"

	ctx := r.Context()

	tracing := middleware.New().Tracer().Value(ctx)
	version := middleware.New().Version().Value(ctx)
	service := middleware.New().Service().Value(ctx)

	ctx, span := tracing.Start(ctx, name, trace.WithAttributes(attribute.String("component", name)), trace.WithAttributes(telemetry.Resources(ctx, service, version.Service).Attributes()...))

	defer span.End()

	client := otelhttp.DefaultClient
	request, e := http.NewRequestWithContext(ctx, "GET", "http://authorization-service.development.svc.cluster.local:8080", nil)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Error While Calling Internal Authorization Service", Source: e}
		return
	}

	headers.Propagate(r, request)

	response, e := client.Do(request)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Perform Request to Authorization Service", Source: e}
		return
	}

	defer response.Body.Close()

	var structure interface{}
	if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
		slog.ErrorContext(ctx, "Unable to Decode Response Body to Normalized Data Structure", slog.String("error", e.Error()))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to JSON Decode Response from Authorization Service", Source: e}
		return
	}

	path := middleware.New().Path().Value(ctx)

	var payload = map[string]interface{}{
		"path":    path,
		"service": middleware.New().Service().Value(ctx),
		"version": middleware.New().Version().Value(ctx).Service,
		middleware.New().Service().Value(ctx): map[string]interface{}{
			"response": structure,
		},
	}

	span.SetAttributes(attribute.String("path", path))

	output <- &types.Response{Code: http.StatusOK, Payload: payload}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	output.Process(w, r, processor)

	return
}
