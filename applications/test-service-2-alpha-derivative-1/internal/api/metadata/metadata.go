package metadata

import (
	"net/http"

	"github.com/x-ethr/server/handler/output"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func processor(w http.ResponseWriter, r *http.Request, output chan<- *types.Response, exception chan<- *types.Exception, options *types.Options) {
	const name = "metadata"

	ctx := r.Context()

	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

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

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	output.Process(w, r, processor)

	return
})
