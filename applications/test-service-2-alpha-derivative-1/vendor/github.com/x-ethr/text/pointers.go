package text

import (
	"log/slog"
)

// Pointer creates a pointer to the specified string value.
//
// The function accepts optional settings of type [Variadic] that configures the behavior of the pointer creation.
// If the Log option in the settings is true and the value is an empty string, a warning log will be printed.
//
// The function returns a pointer to the string parameter.
func Pointer(v string, settings ...Variadic) *string {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	if o.Log && v == "" {
		slog.Warn("Value is Empty String")
	}

	return &v
}

// Dereference dereferences the given string pointer and returns the string value it points to.
//
// The function accepts optional settings of type [Variadic] that configure the behavior of dereferencing.
// If the pointer is nil and the [Options.Log] directive is true, a warning log will be printed
// and an empty string will be returned.
//
// The function returns the string value of the pointer.
func Dereference(v *string, settings ...Variadic) string {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	if v == nil {
		if o.Log {
			slog.Warn("String nil Pointer - Returning Empty String")
		}

		return ""
	}

	return *(v)
}
