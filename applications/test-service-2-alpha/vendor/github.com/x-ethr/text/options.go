package text

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Options is the configuration structure optionally mutated via the [Variadic] constructor used throughout the package.
type Options struct {
	// Log represents an optional flag that will log when potential, unexpected behavior could occur. E.g.
	// when using the [Dereference] function, log a warning that the pointer was nil.
	Log bool

	// Language represents the language tag used in the [Options] struct. [Pointer] defaults to [language.AmericanEnglish].
	//
	//   - See [language.Tag] for type information.
	Language language.Tag

	// Options represents an array of [cases.Option]. These are only applicable to certain casing functions.
	Options []cases.Option
}

// Variadic represents a functional constructor for the [Options] type. Typical callers of Variadic won't need to perform
// nil checks as all implementations first construct an [Options] reference using packaged default(s).
type Variadic func(o *Options)

// options represents a default constructor.
func options() *Options {
	return &Options{ // default Options constructor
		Language: language.AmericanEnglish,
	}
}
