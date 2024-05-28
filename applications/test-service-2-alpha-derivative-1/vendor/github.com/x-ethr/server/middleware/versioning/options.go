package versioning

import (
	"github.com/x-ethr/server/internal/keystore"
)

type Settings struct {
	// The `Version` struct represents the version information of a service or API. It has two fields: `API` and `Service`.
	Version Version `json:"version" yaml:"version"`
}

type Variadic keystore.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
