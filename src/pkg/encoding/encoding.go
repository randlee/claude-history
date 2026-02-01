// Package encoding handles path encoding/decoding for Claude Code storage.
//
// Claude Code encodes filesystem paths for use as directory names by replacing
// special characters with dashes. This creates a flat directory structure
// under ~/.claude/projects/
package encoding

import (
	"runtime"
	"strings"
)

// EncodePath converts an absolute filesystem path to Claude's encoded format.
//
// Encoding rules:
//   - ':' → '-' (Windows drive letter)
//   - '\' → '-' (Windows path separator)
//   - '/' → '-' (Unix path separator)
//   - '.' → '-' (dots in path components)
//
// Examples:
//
//	/home/user/projects/my-app → -home-user-projects-my-app
//	C:\Users\JohnDoe\projects  → C--Users-JohnDoe-projects
func EncodePath(absPath string) string {
	result := absPath

	// Replace all special characters with dash
	// Order matters: do colon before slashes
	result = strings.ReplaceAll(result, ":", "-")  // Windows drive letter
	result = strings.ReplaceAll(result, "\\", "-") // Windows separator
	result = strings.ReplaceAll(result, "/", "-")  // Unix separator
	result = strings.ReplaceAll(result, ".", "-")  // Dots in paths

	return result
}

// DecodePath attempts to convert an encoded path back to a filesystem path.
//
// Note: Perfect round-trip decoding is impossible since '.', '/', '\', and ':'
// all encode to '-'. This function uses heuristics based on the target OS.
// For accurate decoding, use sessions-index.json which stores the original
// projectPath.
//
// If targetOS is empty, it defaults to the current runtime OS.
func DecodePath(encoded string, targetOS string) string {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}

	if targetOS == "windows" {
		// Assume format: C--path-components (drive letter followed by --)
		if len(encoded) >= 2 && encoded[1] == '-' {
			// Check if first char is a drive letter (A-Z)
			firstChar := encoded[0]
			if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
				// Restore drive letter: C- → C:
				rest := encoded[2:]
				return string(firstChar) + ":" + strings.ReplaceAll(rest, "-", "\\")
			}
		}
		// Fallback: just replace dashes with backslashes
		return strings.ReplaceAll(encoded, "-", "\\")
	}

	// Unix (macOS, Linux): assume leading - was /
	if strings.HasPrefix(encoded, "-") {
		return "/" + strings.ReplaceAll(encoded[1:], "-", "/")
	}

	// Fallback: replace dashes with forward slashes
	return strings.ReplaceAll(encoded, "-", "/")
}

// IsEncodedPath checks if a string looks like an encoded Claude path.
// Encoded paths typically start with '-' (Unix) or a drive letter followed by '--' (Windows).
func IsEncodedPath(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Unix-style: starts with dash
	if strings.HasPrefix(s, "-") {
		return true
	}

	// Windows-style: starts with drive letter followed by --
	if len(s) >= 3 {
		firstChar := s[0]
		if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
			if s[1] == '-' && s[2] == '-' {
				return true
			}
		}
	}

	return false
}
