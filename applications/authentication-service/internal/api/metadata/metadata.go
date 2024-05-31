package metadata

import (
	"net/http"

	"github.com/x-ethr/server/handler/output"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
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

	path := middleware.New().Path().Value(ctx)

	var payload = map[string]interface{}{
		middleware.New().Service().Value(ctx): map[string]interface{}{
			"path":    path,
			"service": middleware.New().Service().Value(ctx),
			"version": middleware.New().Version().Value(ctx).Service,
		},
	}

	span.SetAttributes(attribute.String("path", path))

	output <- &types.Response{Code: http.StatusOK, Payload: payload}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	output.Process(w, r, processor)

	return
}
