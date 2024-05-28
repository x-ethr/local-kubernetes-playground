package color

// Options is the configuration structure optionally mutated via the [Variadic] constructor used throughout the package.
type Options struct {
	Line  bool // Line represents the option to add a newline character to the end of format calls such as [ANSI.Print] and [ANSI.Write]. Defaults to true.
	Space bool // Space represents the option to add a trailing space character to the end of format calls such as [ANSI.Print] and [ANSI.Write]. Defaults to false.
}

// Variadic represents a functional constructor for the [Options] type. Typical callers of Variadic won't need to perform
// nil checks as all implementations first construct an [Options] reference using packaged default(s).
type Variadic func(o *Options)

// options represents a default constructor.
func options() *Options {
	return &Options{ // default Options constructor
		Line: true,
	}
}
