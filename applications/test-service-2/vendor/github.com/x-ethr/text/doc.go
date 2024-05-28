// Package text provides string utilities for the american-english language. Additional functions such as Pointer and Dereference
// are also provided.
//
//   - The following package was created for HTTP web-services and interfacing with external cloud-providers and APIs.
//   - For all functions where a [golang.org/x/text/language.Tag] is relevant, [golang.org/x/text/language.AmericanEnglish] is the default value.
//   - Note that most functions contain a special [Variadic] constructor to enable logging, configure a [golang.org/x/text/language.Tag], and other
//     settings.
//   - The package uses [log/slog] for logging purposes. Such configuration is forced and cannot be modified.
package text
