// Package path provides middleware for adding request-specific route context. The following
// package is of utmost importance as the telemetry.Implementation depends on having a path context key
// for wrapping telemetry route handler(s).
//
//   - The Implementation must be added to the stack **before** the telemetry.Implementation middleware.
package path
