package logging

import (
	"log/slog"
	"sync/atomic"
)

var l atomic.Value // default logging level

// Level sets the atomic, global log-level
func Level(v slog.Level) {
	l.Store(v)
}
