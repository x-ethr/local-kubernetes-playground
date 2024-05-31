package registration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"

	"github.com/x-ethr/pg"
	"github.com/x-ethr/server/handler/input"
	"github.com/x-ethr/server/handler/types"
	"github.com/x-ethr/server/middleware"
	"github.com/x-ethr/server/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"user-service/internal/reflection"
	"user-service/internal/token"
	"user-service/models/users"
)

func processor(r *http.Request, input *Body, output chan<- *types.Response, exception chan<- *types.Exception, options *types.Options) {
	const name = "registration"

	ctx := r.Context()

	tracing := middleware.New().Tracer().Value(ctx)
	version := middleware.New().Version().Value(ctx)
	service := middleware.New().Service().Value(ctx)

	ctx, span := tracing.Start(ctx, name, trace.WithAttributes(attribute.String("component", name)), trace.WithAttributes(telemetry.Resources(ctx, service, version.Service).Attributes()...))

	user := &users.CreateParams{Email: input.Email, Username: input.Username}

	dsn := pg.DSN()
	connection, e := pg.Connection(ctx, dsn)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Establish Connection to Database", Source: e}
		return
	}

	defer connection.Release()

	count, e := users.New(connection).Count(ctx, user.Email)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Check if User Exist(s)", Source: e}
		return
	} else if count >= 1 {
		exception <- &types.Exception{Code: http.StatusConflict, Log: "Existing User with Email Address Already Exists", Message: "Account With Email Address Already Exists"}
		return
	}

	var payload = map[string]string{"email": user.Email, "password": input.Password}

	content, e := json.Marshal(payload)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to JSON Marshal Payload to Authorization Service", Source: e}
		return
	}

	var buffer bytes.Buffer
	if size, e := buffer.Write(content); e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Write Payload Content to Buffer", Source: e, Metadata: map[string]interface{}{"size": size}}
		return
	}

	client := otelhttp.DefaultClient
	request, e := http.NewRequestWithContext(ctx, http.MethodPost, "http://authorization-service.development.svc.cluster.local:8080/users", &buffer)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Create Request to Authorization Service", Source: e}
		return
	}

	response, e := client.Do(request)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Error Making Request to Authorization Service", Source: e}
		return
	} else if response.StatusCode >= 400 {
		content, e := io.ReadAll(response.Body)
		if e != nil {
			span.RecordError(e, trace.WithStackTrace(true))
			exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Read Error Response from Authorization Service", Source: e}
			return
		}

		exception <- &types.Exception{Code: response.StatusCode, Log: "Error Making Request to Authorization Service", Source: fmt.Errorf("response: %s", string(content))}
		return
	}

	defer response.Body.Close()

	var structure map[string]interface{}
	if e := json.NewDecoder(response.Body).Decode(&structure); e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Error Decoding Authorization Service's Response Body", Source: e}
		return
	}

	result, e := users.New(connection).Create(ctx, user)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Source: e, Log: "Unable to Create New User"}
		return
	}

	// token , e := jwt.ParseWithClaims(structure["jwt"].(string),

	jwt, e := token.Verify(ctx, structure["token"].(string))
	if e != nil {
		span.RecordError(e)
		exception <- &types.Exception{Source: e, Code: http.StatusForbidden, Log: "Error Verifying JWT Token from Authorization Service"}
		return
	}

	issuer, e := jwt.Claims.GetIssuer()
	if e != nil {
		span.RecordError(e)
		exception <- &types.Exception{Source: e, Code: http.StatusForbidden, Log: "Error Getting Issuer from Authorization Service"}
		return
	} else if issuer != "authorization-service" {
		span.RecordError(e)
		exception <- &types.Exception{Source: fmt.Errorf("unexpected issuer from authorization service: %s", issuer), Code: http.StatusForbidden, Log: fmt.Sprintf("Unexpected Issuer from Authorization Service: %s", issuer)}
		return
	}

	subject, e := jwt.Claims.GetSubject()
	if e != nil {
		span.RecordError(e)
		exception <- &types.Exception{Source: e, Code: http.StatusForbidden, Log: "Error Getting Subject from Authorization Service"}
		return
	} else if subject != "user-service" {
		span.RecordError(e)
		exception <- &types.Exception{Source: fmt.Errorf("unexpected sub from authorization service: %s", issuer), Code: http.StatusForbidden, Log: fmt.Sprintf("Unexpected Subject from Authorization Service: %s", subject)}
		return
	}

	maps.Copy(structure, reflection.Map(result))

	reissue, e := token.Create(ctx, jwt)
	if e != nil {
		span.RecordError(e, trace.WithStackTrace(true))
		exception <- &types.Exception{Code: http.StatusInternalServerError, Log: "Unable to Create New User"}
	}

	output <- &types.Response{Code: http.StatusOK, Payload: structure}

	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	input.Process(w, r, v, processor)

	return
}
