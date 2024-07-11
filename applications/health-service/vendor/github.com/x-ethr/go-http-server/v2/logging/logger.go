package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/x-ethr/color"
	"github.com/x-ethr/levels"
)

func caller() runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, "github.com/x-ethr/server") && !(strings.Contains(frame.Function, "log/slog")) {
			return frame
		}
		if !more {
			break
		}
	}
	return frame
}

func frame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

// logger represents an atomic pointer to enumeration Format - must be one of ( JSON | Text | Default ).
var logger atomic.Value // atomic.Pointer[Type]

// Format will update the atomic logger value of which the slog.Handler is constructed. Valid values are "json" | "text" | "JSON" | "Text" | "TEXT". Defaults to "text".
func Format(v string) {
	switch v {
	case "json", "JSON":
		logger.Store("json")
	case "text", "Text", "TEXT":
		logger.Store("text")
	default:
		logger.Store("text")
	}
}

var verbosity atomic.Pointer[bool]

// Verbose enables an atomic.Pointer to log internal Handler log messages.
func Verbose(v bool) {
	verbosity.Store(&v)
}

type Handler struct {
	slog.Handler

	service  string
	settings *slog.HandlerOptions

	writer io.Writer

	t    string
	text *slog.TextHandler
	json *slog.JSONHandler

	logger *log.Logger

	group      string
	attributes []slog.Attr
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	h.logger.SetOutput(h.writer)

	minimum := l.Load().(slog.Level)
	if h.settings == nil {
		h.settings = &slog.HandlerOptions{}
	} else if h.settings.Level != nil {
		minimum = h.settings.Level.Level()
	}

	enabled := level >= minimum
	if verbosity.Load() != nil && *(verbosity.Load()) {
		fmt.Fprintf(os.Stdout, "Evaluating Logger Enablement - Atomic Log Level: %s, Caller's Log Level: %s, Enabled: %v \n", minimum, level, enabled)
	}

	return enabled
}

func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(h.attributes...)

	if h.t == "json" {
		return h.json.Handle(ctx, record)
	}

	var level string

	switch record.Level {
	case levels.Trace:
		level = fmt.Sprintf("%s:", color.Color().Default("TRACE"))
	case levels.Debug:
		level = fmt.Sprintf("%s:", color.Color().Magenta("DEBUG"))
	case levels.Info:
		level = fmt.Sprintf("%s:", color.Color().Green("INFO"))
	case levels.Warn:
		level = fmt.Sprintf("%s:", color.Color().Yellow("WARN"))
	case levels.Error:
		level = fmt.Sprintf("%s:", color.Color().Red("ERROR"))
	case levels.Fatal:
		level = fmt.Sprintf("%s:", color.Color().Red("FATAL"))
	default:
		return fmt.Errorf("invalid, unknown level: %s", record.Level.String())
	}

	fields := make(map[string]interface{}, record.NumAttrs())

	var evaluate = func(a slog.Attr) bool {
		if kind := a.Value.Kind(); kind == slog.KindGroup {
			group := a.Value.Group()
			mapping := make(map[string]interface{}, len(group))

			for root := range group {
				attribute := group[root]
				value := attribute.Value.Any()
				if kind := attribute.Value.Kind(); kind == slog.KindGroup {
					child := attribute.Value.Group()
					children := make(map[string]interface{}, len(child))
					for parent := range attribute.Value.Group() {
						sibling := child[parent]
						assignment := sibling.Value.Any()
						if kind := sibling.Value.Kind(); kind == slog.KindGroup {
							final := sibling.Value.Group()
							nesting := make(map[string]interface{}, len(final))
							for index := range final {
								nest := final[index]
								v := nest.Value.Any()

								if kind := nest.Value.Kind(); kind == slog.KindGroup {
									nesting[nest.Key] = nest.Value.String()
								} else {
									nesting[nest.Key] = v
								}
							}

							children[sibling.Key] = nesting
						} else {
							children[sibling.Key] = assignment
						}
					}

					mapping[attribute.Key] = children
				} else {
					mapping[attribute.Key] = value
				}
			}

			fields[a.Key] = mapping

			return true
		}

		value := a.Value.Any()
		switch value.(type) {
		case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
			fields[a.Key] = value
		default:
			output, e := json.Marshal(a.Value.Any())
			if e != nil {
				relevant := caller()

				line, f, file := relevant.Line, relevant.Func, relevant.File

				fmt.Fprintf(os.Stderr, "ERROR - (%d) (%s) (%s) Unable to Marshal Logging Attribute (%s): %s - %v\n", line, file, f.Name(), a.Key, a.Value.String(), e)

				return false
			}

			if e := json.Unmarshal(output, &value); e != nil {
				relevant := caller()

				line, f, file := relevant.Line, relevant.Func, relevant.File

				fmt.Fprintf(os.Stderr, "ERROR - (%d) (%s) (%s) Unable to Unmarshal Logging Attribute (%s): %s\n", line, file, f.Name(), a.Key, a.Value.String())

				return false
			}

			fields[a.Key] = value
			if a.Key == "$" && fields[a.Key] != nil { // special key to denote the mapping should be inlined.
				if v, ok := value.(map[string]interface{}); ok {
					fields = v
				}
			}
		}

		return true
	}

	record.Attrs(evaluate)
	for index := range h.attributes {
		attribute := h.attributes[index]
		evaluate(attribute)
	}

	format := record.Time.Format("[Jan 02 15:04:05.000]")
	message := color.Color().Cyan(record.Message).String()

	if service := h.service; service != "" && h.logger.Prefix() == "" {
		literal := color.Color().Bold(color.Color().Red(service).String()).String()

		prefix := fmt.Sprintf("%s ", literal)
		if group := h.group; group != "" {
			prefix = fmt.Sprintf("%s (%s) ", literal, group)
		}

		h.logger.SetPrefix(prefix)
	} else if group := h.group; group != "" {
		literal := color.Color().Bold(color.Color().Red(group).String()).String()

		prefix := fmt.Sprintf("%s ", literal)

		h.logger.SetPrefix(prefix)
	}

	var buffer []byte
	if record.Message == "HTTP(s) Request" || record.Message == "Middleware" || record.Message == "Response" {
		var e error

		buffer, e = json.Marshal(fields)
		if e != nil {
			e = fmt.Errorf("failed to marshal fields to json: %v", e)
			return e
		}

		partials := bytes.Split(buffer, []byte(":"))
		buffer = bytes.Join(partials, []byte(": "))

		partials = bytes.Split(buffer, []byte(","))
		buffer = bytes.Join(partials, []byte(", "))

		partials = bytes.Split(buffer, []byte("{"))
		buffer = bytes.Join(partials, []byte("{ "))

		partials = bytes.Split(buffer, []byte("}"))
		buffer = bytes.Join(partials, []byte(" }"))
	} else {
		var e error

		buffer, e = json.MarshalIndent(fields, "", "    ")
		if e != nil {
			e = fmt.Errorf("failed to marshal fields to json: %v", e)
			return e
		}
	}

	h.logger.Println(color.Color().Dim(format), level, message, color.Color().White(string(buffer)))

	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		text:       slog.NewTextHandler(h.writer, h.settings).WithAttrs(attrs).(*slog.TextHandler),
		json:       slog.NewJSONHandler(h.writer, h.settings).WithAttrs(attrs).(*slog.JSONHandler),
		writer:     h.writer,
		service:    h.service,
		settings:   h.settings,
		logger:     h.logger,
		attributes: attrs,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		text:     slog.NewTextHandler(h.writer, h.settings).WithGroup(name).(*slog.TextHandler),
		json:     slog.NewJSONHandler(h.writer, h.settings).WithGroup(name).(*slog.JSONHandler),
		writer:   h.writer,
		service:  h.service,
		settings: h.settings,
		logger:   h.logger,
		group:    name,
	}
}

func Logger(settings ...Variadic) slog.Handler {
	var o = Specification()
	for _, configuration := range settings {
		configuration(o)
	}

	var instantiation = &Handler{
		text:       slog.NewTextHandler(o.Writer, o.Settings),
		json:       slog.NewJSONHandler(o.Writer, o.Settings),
		writer:     o.Writer,
		service:    o.Service,
		settings:   o.Settings,
		logger:     log.New(o.Writer, "", 0),
		attributes: make([]slog.Attr, 0),
	}

	if logger.Load() != nil && logger.Load().(string) == "json" {
		instantiation.t = "json"
		instantiation.Handler = instantiation.json
	} else {
		instantiation.t = "text"
		instantiation.Handler = instantiation.text
	}

	return instantiation
}
