package text

import (
	"log/slog"

	"golang.org/x/text/cases"
)

// Title will cast a string to title casing. If no options are provided, [cases.NoLower] will be appended by default.
func Title(v string, settings ...Variadic) string {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	if len(o.Options) == 0 {
		o.Options = append(o.Options, cases.NoLower)
	}

	if v == "" && o.Log {
		slog.Warn("Empty String Provided as Value - Returning Unmodified Value")
		return ""
	}

	casing := cases.Title(o.Language, o.Options...)

	return casing.String(v)
}

// Lowercase will cast a string to all-lower casing.
func Lowercase(v string, settings ...Variadic) string {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	if v == "" && o.Log {
		slog.Warn("Empty String Provided as Value - Returning Unmodified Value")
		return ""
	}

	casing := cases.Lower(o.Language, o.Options...)

	return casing.String(v)
}

// Uppercase will cast a string to an uppercase casing.
func Uppercase(v string, settings ...Variadic) string {
	var o = options()
	for _, configuration := range settings {
		configuration(o)
	}

	if v == "" && o.Log {
		slog.Warn("Empty String Provided as Value - Returning Unmodified Value")
		return ""
	}

	casing := cases.Upper(o.Language, o.Options...)

	return casing.String(v)
}
