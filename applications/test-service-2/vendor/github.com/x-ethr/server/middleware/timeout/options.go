package timeout

import (
	"time"

	"github.com/x-ethr/server/internal/keystore"
)

type Settings struct {
	// Timeout represents the duration to wait before considering an operation as timed out. If unspecified, or a negative value,
	// a default of 30 seconds is overwritten.
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

type Variadic keystore.Variadic[Settings]

func settings() *Settings {
	return &Settings{
		Timeout: (time.Second * 30),
	}
}
