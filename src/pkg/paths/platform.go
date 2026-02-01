package paths

import (
	"path/filepath"
	"runtime"
	"strings"
)

// NormalizePath converts a path to use the OS-appropriate separators
// and cleans it (removes redundant separators, . and ..).
func NormalizePath(path string) string {
	return filepath.Clean(path)
}

// ToAbsolute converts a path to an absolute path.
// If the path is already absolute, it's returned cleaned.
// If relative, it's resolved relative to the current working directory.
func ToAbsolute(path string) (string, error) {
	return filepath.Abs(path)
}

// IsAbsolute checks if a path is absolute.
func IsAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

// IsWindowsPath checks if a path looks like a Windows path
// (has a drive letter like C: or uses backslashes).
func IsWindowsPath(path string) bool {
	if len(path) == 0 {
		return false
	}

	// Check for backslashes
	if strings.Contains(path, "\\") {
		return true
	}

	// Check for drive letter (e.g., C:)
	if len(path) >= 2 {
		firstChar := path[0]
		if (firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') {
			if path[1] == ':' {
				return true
			}
		}
	}

	return false
}

// IsUnixPath checks if a path looks like a Unix path (starts with /).
func IsUnixPath(path string) bool {
	return len(path) > 0 && path[0] == '/'
}

// CurrentOS returns the current operating system identifier.
func CurrentOS() string {
	return runtime.GOOS
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS returns true if running on macOS.
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// PathSeparator returns the OS path separator as a string.
func PathSeparator() string {
	return string(filepath.Separator)
}
