package alpha

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/x-ethr/server"
	"github.com/x-ethr/server/headers"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"github.com/x-ethr/server/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var handle server.Handle = func(x *types.CTX) {
	const name = "login"

	ctx := x.Request().Context()

	labeler := telemetry.Labeler(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	logger := slog.With(slog.String("name", name))

	client := otelhttp.DefaultClient
	request, e := http.NewRequestWithContext(ctx, "GET", "http://test-service-2-alpha.development.svc.cluster.local:8080", nil)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		logger.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Error While Calling Internal Service", Source: e})
		return
	}

	headers.Propagate(x.Request(), request)

	// response, e := otelhttp.Get(ctx, "http://test-service-2-alpha.development.svc.cluster.local:8080")
	response, e := client.Do(request)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		logger.ErrorContext(ctx, "Error While Calling Internal Service", slog.String("error", e.Error()))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Error While Calling Internal Service", Source: e})
		return
	}

	defer response.Body.Close()

	var structure interface{}
	if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
		labeler.Add(attribute.Bool("error", true))
		logger.ErrorContext(ctx, "Unable to Decode Response Body to Normalized Data Structure", slog.String("error", e.Error()))
		x.Error(&types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Marshal Response Data Structure", Source: e})
		return
	}

	var payload = map[string]interface{}{
		middleware.New().Service().Value(ctx): map[string]interface{}{
			"path":    middleware.New().Path().Value(ctx),
			"service": middleware.New().Service().Value(ctx),
			"version": middleware.New().Version().Value(ctx).Service,

			"response": structure,
		},
	}

	x.Complete(&types.Response{Status: http.StatusOK, Payload: payload})

}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	server.Process(w, r, handle)

	return
})
