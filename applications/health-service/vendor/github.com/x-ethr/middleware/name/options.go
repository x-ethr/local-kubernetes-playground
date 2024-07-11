package name

import "github.com/x-ethr/middleware/types"

type Settings struct {
	// Service represents a string field in the Settings struct. It is used to configure the service name in middleware configurations (X-Service-Name) [http.Header].
	Service string `json:"service" yaml:"service"`
}

type Variadic types.Variadic[Settings]

func settings() *Settings {
	return &Settings{}
}
