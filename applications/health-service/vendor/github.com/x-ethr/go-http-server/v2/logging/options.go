package logging

import (
	"io"
	"log/slog"
	"os"
)

type Options struct {
	Service  string
	Settings *slog.HandlerOptions
	Logs     Logs

	// Writer represents the output [io.Writer]. Defaults to [os.Stdout] if unspecified.
	Writer io.Writer
}

type Exclusions struct {
	// Headers to exclude from request logging. Defaults include:
	//
	// 	- "Accept"
	// 	- "Accept-Encoding"
	// 	- "Accept-Language"
	// 	- "Connection"
	// 	- "Content-Length"
	// 	- "Content-Type"
	// 	- "Upgrade-Insecure-Requests"
	// 	- "Sec-Fetch-Mode"
	// 	- "Sec-Fetch-Site"
	// 	- "Sec-Fetch-Resource"
	// 	- "Sec-Fetch-User"
	// 	- "Sec-Fetch-Dest"
	// 	- "User-Agent"
	Headers []string
}

type Logs struct {
	Exclusions Exclusions
}

type Variadic func(o *Options)

func Specification() *Options {
	return &Options{
		Service: "",
		Writer:  os.Stdout,
		Settings: &slog.HandlerOptions{
			AddSource:   false,
			Level:       nil,
			ReplaceAttr: nil,
		},
		Logs: Logs{
			Exclusions: Exclusions{
				Headers: []string{
					"Accept",
					"Accept-Encoding",
					"Accept-Language",
					"Connection",
					"Content-Length",
					"Content-Type",
					"Upgrade-Insecure-Requests",
					"Sec-Fetch-Mode",
					"Sec-Fetch-Site",
					"Sec-Fetch-Resource",
					"Sec-Fetch-User",
					"Sec-Fetch-Dest",
					"User-Agent", // @TODO conditionally evaluate later
					"X-Forwarded-Client-Cert",
					"X-Forwarded-For",
					"X-Forwarded-Proto",
					"X-Request-ID",
					"Traceparent",
					"Tracestate",
					"X-Envoy-Attempt-Count",
					"Postman-Token",
				},
			},
		},
	}
}
