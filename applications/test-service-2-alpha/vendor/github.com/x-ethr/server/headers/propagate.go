package headers

import "net/http"

// Propagate is a function that copies specific HTTP headers from a source request to a target request.
// It uses a pre-defined set of headers to copy, including various tracing and context headers.
// These headers can be used for analyzing network traffic and ensuring that different parts of your
// system can properly communicate with each other. Note that if a header is not present in the source
// request, it will not be added to the target request.
//
// Headers include:
//   - "portal"
//   - "device"
//   - "user"
//   - "travel"
//   - "x-request-id"
//   - "x-b3-traceid"
//   - "x-b3-spanid"
//   - "x-b3-parentspanid"
//   - "x-b3-sampled"
//   - "x-b3-flags"
//   - "x-ot-span-context"
func Propagate(source, target *http.Request) {
	headers := []string{
		http.CanonicalHeaderKey("portal"),
		http.CanonicalHeaderKey("device"),
		http.CanonicalHeaderKey("user"),
		http.CanonicalHeaderKey("travel"),
		http.CanonicalHeaderKey("x-request-id"),
		http.CanonicalHeaderKey("x-b3-traceid"),
		http.CanonicalHeaderKey("x-b3-spanid"),
		http.CanonicalHeaderKey("x-b3-parentspanid"),
		http.CanonicalHeaderKey("x-b3-sampled"),
		http.CanonicalHeaderKey("x-b3-flags"),
		http.CanonicalHeaderKey("x-ot-span-context"),
	}

	for key := range headers {
		header := headers[key]

		assignment := source.Header.Get(header)

		if assignment != "" {
			target.Header.Add(header, source.Header.Get(header))
		}
	}
}
