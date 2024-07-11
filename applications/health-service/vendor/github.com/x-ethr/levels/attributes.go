package levels

import "log/slog"

// Attributes returns the updated attributes based on the provided groups and attribute.
//
//   - Handles updating the [slog.LevelKey] according to the [levels] package's additional custom log-levels: [Trace] and [Fatal]
//   - Usage is to be expected with [slog.HandlerOptions.ReplaceAttr].
func Attributes(groups []string, a slog.Attr) slog.Attr {
	if a.Key == (slog.LevelKey) {
		switch a.Value.String() {
		case "DEBUG-4":
			a.Value = slog.StringValue("TRACE")
		case "ERROR+4":
			a.Value = slog.StringValue("FATAL")
		}
	}

	return a
}
