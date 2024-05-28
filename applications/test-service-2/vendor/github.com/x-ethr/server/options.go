package server

import (
	"net/http"

	"github.com/x-ethr/server/logging"
)

type Globals struct {
	// Disable determines whether to include [Globals.Middleware] with the caller's handler's own middleware(s). Defaults to false.
	//
	// 	- Note that the following field may be subject to breaking change(s) as additional [Globals] attributes may be added.
	Disable bool

	Logging *logging.Options

	// Middleware to wrap all [http.Handler] implementation(s) with.
	Middleware []func(http.Handler) http.Handler
}

// Options is the configuration structure optionally mutated via the [Variadic] constructor used throughout the package.
type Options struct {
	Globals Globals

	// Metadata will enable [Multiplexer] metadata to get added as a context key during route registration. Defaults to true.
	Metadata bool

	// Middleware to wrap the route's [http.Handler] implementation(s) with. For configuring middleware that should be added to all
	// of a [Mux] [http.Handler] implementation(s), see [Options.Globals], [Globals].
	Middleware []func(http.Handler) http.Handler
}

// Variadic represents a functional constructor for the [Options] type. Typical callers of Variadic won't need to perform
// nil checks as all implementations first construct an [Options] reference using packaged default(s).
type Variadic func(o *Options)

// options represents a default constructor.
func options() *Options {
	return &Options{ // default Options constructor
		Globals: Globals{
			Disable:    false,
			Logging:    logging.Specification(),
			Middleware: make([]func(http.Handler) http.Handler, 0),
		},

		Metadata:   true,
		Middleware: make([]func(http.Handler) http.Handler, 0),
	}
}
