// Package version provides the single source of truth for the application version.
package version

// Version is the application version string.
// This is used for development builds and can be overridden by GoReleaser
// via ldflags during release builds.
const Version = "0.3.0"
