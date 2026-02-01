package encoding

import (
	"testing"
)

func TestEncodePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Unix simple path",
			input:    "/home/user/projects/my-app",
			expected: "-home-user-projects-my-app",
		},
		{
			name:     "Unix path with hidden folder",
			input:    "/home/user/.config/settings",
			expected: "-home-user--config-settings",
		},
		{
			name:     "macOS path",
			input:    "/Users/randlee/Documents/github/github-research",
			expected: "-Users-randlee-Documents-github-github-research",
		},
		{
			name:     "Windows simple path",
			input:    "C:\\Users\\JohnDoe\\projects\\my-app",
			expected: "C--Users-JohnDoe-projects-my-app",
		},
		{
			name:     "Windows path with hidden folder",
			input:    "C:\\Users\\JohnDoe\\.config\\settings",
			expected: "C--Users-JohnDoe--config-settings",
		},
		{
			name:     "Path with dots in filename",
			input:    "/home/user/file.name.txt",
			expected: "-home-user-file-name-txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodePath(tt.input)
			if result != tt.expected {
				t.Errorf("EncodePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDecodePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		targetOS string
		expected string
	}{
		{
			name:     "Unix encoded path",
			input:    "-home-user-projects-my-app",
			targetOS: "linux",
			expected: "/home/user/projects/my/app",
		},
		{
			name:     "macOS encoded path",
			input:    "-Users-randlee-Documents-github",
			targetOS: "darwin",
			expected: "/Users/randlee/Documents/github",
		},
		{
			name:     "Windows encoded path",
			input:    "C--Users-JohnDoe-projects",
			targetOS: "windows",
			expected: "C:\\Users\\JohnDoe\\projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodePath(tt.input, tt.targetOS)
			if result != tt.expected {
				t.Errorf("DecodePath(%q, %q) = %q, want %q", tt.input, tt.targetOS, result, tt.expected)
			}
		})
	}
}

func TestIsEncodedPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Unix encoded path",
			input:    "-home-user-projects",
			expected: true,
		},
		{
			name:     "Windows encoded path",
			input:    "C--Users-JohnDoe",
			expected: true,
		},
		{
			name:     "Unix absolute path",
			input:    "/home/user/projects",
			expected: false,
		},
		{
			name:     "Windows absolute path",
			input:    "C:\\Users\\JohnDoe",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "UUID",
			input:    "679761ba-80c0-4cd3-a586-cc6a1fc56308",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncodedPath(tt.input)
			if result != tt.expected {
				t.Errorf("IsEncodedPath(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
