package headers

import (
	"log/slog"
	"net/http"
	"slices"
	"strings"
)

// Type represents the request's incoming or outgoing context type.
type Type string

const (
	// Incoming represents a contextual, incoming request type.
	Incoming Type = "incoming"
	// Outgoing represents a contextual, incoming request type.
	Outgoing Type = "outgoing"
)

// message represents the string log message deriving from the Type.
//
//   - Internal module usage only.
func (t Type) message() string {
	switch t {
	case Incoming:
		return "Incoming Request Header(s)"
	case Outgoing:
		return "Outgoing Request Header(s)"
	default:
		return "Unknown Request Type"
	}
}

// x represents the string log message deriving from the Type.
//
//   - Internal module usage only.
func (t Type) x() string {
	switch t {
	case Incoming:
		return "Incoming Request X-Header(s)"
	case Outgoing:
		return "Outgoing Request -XHeader(s)"
	default:
		return "Unknown Request Type"
	}
}

// Log logs the information about the HTTP request headers to the standard logger.
// This information contains all request headers along with their first value.
// The whole log message is grouped by the host. Total number of headers is also provided.
//
// This function takes two arguments: a pointer to http.Request and
// a custom Type (either 'Incoming' or 'Outgoing').
// It doesn't return any value.
//
// Using the logger is optional, it doesn't stop the program's execution in case
// logging doesn't work. If the type is neither 'Incoming' nor 'Outgoing',
// a message with 'Unknown Request Type' will be written to the log.
//
// Usage: headers.Log(request, headers.Incoming) - for incoming requests,
// headers.Log(request, headers.Outgoing) - for outgoing requests.
func Log(r *http.Request, t Type) {
	ctx := r.Context()

	exclusions := []string{
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Connection",
		"Cache-Control",
		"Sec-Fetch-Dest",
		"Sec-Fetch-Mode",
		"Sec-Fetch-Site",
		"Upgrade-Insecure-Requests",
	}

	headers := make(map[string]string)
	for header, values := range r.Header {
		if !(slices.Contains(exclusions, header)) && len(values) >= 1 && values[0] != "" {
			headers[header] = values[0]
		}
	}

	slog.Log(ctx, slog.LevelDebug, t.message(), slog.Any("$", headers))
}

func X(r *http.Request, t Type) {
	ctx := r.Context()
	headers := make(map[string]string)
	for header, values := range r.Header {
		if strings.HasPrefix(http.CanonicalHeaderKey(header), "X") && len(values) >= 1 {
			headers[header] = values[0]
		}
	}

	slog.Log(ctx, slog.LevelDebug, t.x(), slog.Any("$", headers))
}
