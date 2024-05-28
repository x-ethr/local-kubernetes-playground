package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
)

// Host represents the hostname for routing HTTP requests. It is a string type.
type Host string

// Method represents an HTTP method used for routing HTTP requests.
type Method string

type Path string

// chain builds a http.Handler composed of an inline middleware stack and endpoint handler in the order they are passed.
func chain(options *Options, endpoint http.Handler, m *Multiplexer) (handler http.Handler) {
	var metadata = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			{
				ctx = context.WithValue(ctx, "multiplexer-metadata", m)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	if options.Metadata {
		handler = metadata(handler)
	}

	middlewares := make([]func(http.Handler) http.Handler, 0)
	middlewares = append(middlewares, options.Globals.Middleware...)
	middlewares = append(middlewares, options.Middleware...)

	var length = len(middlewares)
	if length == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	handler = middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return
}

// New creates a new instance of the Mux struct with the provided settings. It constructs an options struct
// using the `options` function, applies the optional settings to it, and returns a pointer to the Mux struct.
// The Mux struct contains mutex for safe concurrent access, options for configuration, and maps to store registered routes.
// The hosts map is initialized as an empty map, and the routes map is initialized as an empty map.
// The returned Mux instance is ready to handle HTTP requests.
func New(settings ...Variadic) *Mux {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	mux := &Mux{options: o, hosts: make(map[Host]map[Path]map[Method]*Multiplexer), routes: make(map[Path]map[Method]*Multiplexer)}

	return mux
}

// Multiplexer represents a type that handles HTTP requests by routing them based on the provided host, path, and method.
// It contains fields for the host, path, method, pattern, and the underlying http.ServeMux used for routing.
type Multiplexer struct {
	Host   string `json:"host" yaml:"host"`
	Path   string `json:"path" yaml:"path"`
	Method string `json:"method" yaml:"method"`

	Pattern string `json:"pattern" yaml:"pattern"`

	mux *http.ServeMux `json:"-" yaml:"-"`
}

// Mux represents a multiplexer that handles HTTP requests. It routes requests based on the provided host, path, and method.
// The Mux struct contains mutex for safe concurrent access, options for configuration, and maps to store registered routes.
type Mux struct {
	mutex   sync.RWMutex
	options *Options

	hosts  map[Host]map[Path]map[Method]*Multiplexer
	routes map[Path]map[Method]*Multiplexer
}

func hostname(host string) string {
	// If no port on host, return unchanged
	if !strings.Contains(host, ":") {
		return host
	}

	v, port, e := net.SplitHostPort(host)
	if e != nil {
		fmt.Println("error host, port, e", v, port, e)
		return host
	}

	return v
}

func (mu *Mux) Register(pattern string, function http.HandlerFunc, settings ...Variadic) {
	mu.Handle(pattern, function, settings...)
}

// search searches for a multiplexer in the routes or hosts map of the multiplexer based on the given host, path, and method.
//
//   - If a host is provided, it checks if the hosts map has been initialized and creates the necessary maps if not.
//
//   - The function then retrieves the multiplexer from the hosts map based on the host, path, and method.
//
//   - If the multiplexer is not found, the found variable is set to false.
//
//   - If no host is provided, it checks if the routes map has been initialized and creates the necessary maps if not.
//
//   - The function then retrieves the multiplexer from the routes map based on the path and method.
//
//   - If the multiplexer is not found, the found variable is set to false.
//
// The function returns the found multiplexer and a boolean indicating whether the multiplexer was found or not.
func (mu *Mux) search(host Host, path Path, method Method) (multiplexer *Multiplexer, found bool) {
	mu.mutex.Lock()
	defer mu.mutex.Unlock()

	switch {
	case host != "":
		if mu.hosts == nil {
			mu.hosts = make(map[Host]map[Path]map[Method]*Multiplexer)
			mu.hosts[host] = make(map[Path]map[Method]*Multiplexer)
			mu.hosts[host][path] = make(map[Method]*Multiplexer)
		} else if mu.hosts[host] == nil {
			mu.hosts[host] = make(map[Path]map[Method]*Multiplexer)
			mu.hosts[host][path] = make(map[Method]*Multiplexer)
		} else if mu.hosts[host][path] == nil {
			mu.hosts[host][path] = make(map[Method]*Multiplexer)
		}

		multiplexer, found = mu.hosts[host][path][method]

		if !(found) {
			// fmt.Println("multiplexer not found", "host", host, "path", path, "method", method)
		}
	default:
		if mu.routes == nil {
			mu.routes = make(map[Path]map[Method]*Multiplexer)
			mu.routes[path] = make(map[Method]*Multiplexer)
		} else if mu.routes[path] == nil {
			mu.routes[path] = make(map[Method]*Multiplexer)
		}

		multiplexer, found = mu.routes[path][method]

		if !(found) {
			// fmt.Println("multiplexer not found", "path", path, "method", method)
		}
	}

	return
}

// Handle registers a pattern and its corresponding handler function to the multiplexer.
// It also accepts optional settings that can be applied to the multiplexer.
//
//   - The pattern must follow the format "[METHOD ][HOST]/[PATH]" or "[METHOD ]/[PATH]".
//   - If the pattern is invalid, a panic will be raised.
//   - The multiplexer metadata will be updated with the registered pattern.
//   - The handler function will be wrapped with the specified middleware functions before being executed.
//   - The registered pattern and handler function are used to construct the pattern string.
//   - The multiplexer is then added to the routes or hosts map of the multiplexer based on the presence of the host.
//
// This method is used to register and handle HTTP requests in the multiplexer.
func (mu *Mux) Handle(pattern string, function http.Handler, settings ...Variadic) {
	var o = options()
	if !(mu.options.Globals.Disable) {
		o = mu.options
	}

	for _, configuration := range settings {
		configuration(o)
	}

	var host string
	method, partial, valid := strings.Cut(pattern, " ")
	if !(valid) {
		panic("invalid pattern - pattern must follow [METHOD ][HOST]/[PATH] or [METHOD ]/[PATH]")
	}

	path := partial
	if !(strings.Contains(partial, "/")) {
		host, path = partial, "/"
	} else if !(strings.HasPrefix("/", partial)) {
		partials := strings.SplitN(partial, "/", 2)

		if len(partials) == 0 {
			panic("invalid pattern - pattern must include a path or a hostname")
		} else if len(partials) == 1 {
			host, path = partials[0], "/"
		} else {
			host, path = partials[0], fmt.Sprintf("/%s", partials[1])
		}
	}

	if v, exists := mu.search(Host(host), Path(path), Method(method)); exists {
		message := fmt.Sprintf("duplicate mux registered for host (%s), path (%s), method (%s): \"%s\"", host, path, method, v.Pattern)
		e := errors.New(message)
		panic(e)
	}

	multiplexer := &Multiplexer{
		Host:   host,
		Path:   path,
		Method: method,

		Pattern: pattern,

		mux: http.NewServeMux(),
	}

	handler := chain(o, function, multiplexer)

	pattern = fmt.Sprintf("%s %s", method, path)
	if host != "" {
		pattern = fmt.Sprintf("%s %s%s", method, host, path)
	}

	multiplexer.mux.Handle(pattern, handler)

	mu.mutex.Lock()
	defer mu.mutex.Unlock()

	if host == "" {
		mu.routes[Path(path)][Method(method)] = multiplexer
		return
	}

	mu.hosts[Host(host)][Path(path)][Method(method)] = multiplexer
}

func (mu *Mux) Handler(r *http.Request) (http.Handler, string) {
	host, path, method := hostname(r.Host), r.URL.EscapedPath(), r.Method

	var multiplexer *Multiplexer
	if partial, found := mu.search("", Path(path), Method(method)); found {
		multiplexer = partial
	} else if full, ok := mu.search(Host(host), Path(path), Method(method)); ok {
		multiplexer = full
	}

	return multiplexer.mux.Handler(r)
}

func (mu *Mux) Pattern(r *http.Request) string {
	_, pattern := mu.Handler(r)

	return pattern
}

// Middleware adds the provided middleware functions to the global middleware stack.
// The middleware functions are applied in the order they are passed.
func (mu *Mux) Middleware(middlewares ...func(http.Handler) http.Handler) {
	if len(middlewares) == 0 {
		return
	}

	mu.options.Globals.Middleware = append(mu.options.Globals.Middleware, middlewares...)

	// slog.Info("Middleware(s)", slog.Int("count", len(mu.options.Globals.Middleware)))
}

func (mu *Mux) log(ctx context.Context, host, method, path string, headers http.Header) {
	mapping := make(map[string]string)

	for k, v := range headers {
		if slices.Contains(mu.options.Globals.Logging.Logs.Exclusions.Headers, k) {
			continue
		} else if http.CanonicalHeaderKey(k) == "Cookie" {
			value := strings.ToLower(v[0])
			if strings.HasPrefix(value, "goland") || strings.HasPrefix(value, "webstorm") {
				v[0] = "IDE-[...]"
			}
		} else if http.CanonicalHeaderKey(k) == "User-Agent" {
			partials := strings.Split(v[0], " ")
			if len(partials) == 0 {
				continue
			}

			v[0] = partials[0]
		}

		mapping[k] = strings.Join(v, ", ")
	}

	slog.InfoContext(ctx, "HTTP(s) Request", slog.Group("$", slog.String("path", path), slog.String("method", method), slog.String("host", host)), slog.Any("headers", mapping))
}

// ServeHTTP handles an HTTP request by routing it to the appropriate handler based on the provided host, path, and method.
// It first retrieves the context from the request and extracts the method, path, and host information.
//
// It then logs the request using the log method of the multiplexer.
//
// Next, it retrieves the handler from the context and checks its type.
// If the handler is of type *Mux, it performs a search for the multiplexer in the routes and hosts maps of the multiplexer
// based on the path and method.
//
// If the multiplexer is not found, it returns a 404 Not Found response.
//
// If the pattern is empty, it raises a 404 Not Found response, otherwise it serves the request using the found handler.
//
// If the handler is of type *http.ServeMux, it retrieves the handler and pattern using the Handler method of the *http.ServeMux.
// It then performs a search for the multiplexer in the routes and hosts maps of the multiplexer based on the path and method.
//
// If the multiplexer is not found, it returns a 404 Not Found response.
//
// If the pattern is empty, it raises a 404 Not Found response, otherwise it serves the request using the found handler.
//
// If the handler is of any other type, it raises a panic with the message "invalid handler".
func (mu *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	method, path, host := r.Method, r.URL.EscapedPath(), hostname(r.Host)

	mu.log(ctx, host, method, path, r.Header)

	server := ctx.Value(http.ServerContextKey).(*http.Server)
	handler := server.Handler

	switch handler.(type) {
	case *Mux:
		var multiplexer *Multiplexer
		if partial, found := mu.search("", Path(path), Method(method)); found {
			multiplexer = partial
		} else if full, ok := mu.search(Host(host), Path(path), Method(method)); ok {
			multiplexer = full
		}

		if multiplexer == nil {
			http.NotFound(w, r)
			return
		}

		handle, pattern := multiplexer.mux.Handler(r)

		if pattern == "" {
			handle.ServeHTTP(w, r) // will be a 404 not found mux
			return
		}

		handle.ServeHTTP(w, r)
		return
	case *http.ServeMux:
		handle, pattern := (handler.(*http.ServeMux)).Handler(r)
		if pattern == "" {
			handle.ServeHTTP(w, r) // will be a 404 not found mux
			return
		}

		method, path, host = r.Method, r.URL.EscapedPath(), hostname(r.Host)

		var multiplexer *Multiplexer
		if partial, found := mu.search("", Path(path), Method(method)); found {
			multiplexer = partial
		} else if full, ok := mu.search(Host(host), Path(path), Method(method)); ok {
			multiplexer = full
		}

		if multiplexer == nil {
			http.NotFound(w, r)
			return
		}

		handle, pattern = multiplexer.mux.Handler(r)

		if pattern == "" {
			handle.ServeHTTP(w, r) // will be a 404 not found mux
			return
		}

		handle.ServeHTTP(w, r)
		return

	default:
		panic("invalid handler")
	}
}
