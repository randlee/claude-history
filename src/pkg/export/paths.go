// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Regular expressions for detecting file paths
// Note: We do NOT allow spaces in paths to avoid false positives.
// If users have files with spaces, they can reference them with quotes or explicit links.
var (
	// Absolute Unix paths: /path/to/file.ext
	// Components: alphanumeric, underscore, dot, dash, tilde (NO SPACES)
	// Tilde is included to support 8.3 short filenames and home directories
	unixAbsPathRe = regexp.MustCompile(`(?:^|\s)(/(?:[a-zA-Z0-9_.~-]+/)*[a-zA-Z0-9_.~-]+\.[a-zA-Z0-9]+)(?:\s|$|[,;:)])`)

	// Absolute Windows paths: C:\path\to\file.ext or C:/path/to/file.ext
	// Tilde is included to support 8.3 short filenames (e.g., C:\Users\RUNNER~1\)
	winAbsPathRe = regexp.MustCompile(`(?:^|\s)([A-Za-z]:[/\\](?:[a-zA-Z0-9_.~-]+[/\\])*[a-zA-Z0-9_.~-]+\.[a-zA-Z0-9]+)(?:\s|$|[,;:)])`)

	// Relative paths with ./  or ../ prefix: ./src/file.go, ../pkg/export.go
	relPathPrefixRe = regexp.MustCompile(`(?:^|\s)(\.\.?/(?:[a-zA-Z0-9_.~-]+/)*[a-zA-Z0-9_.~-]+(?:\.[a-zA-Z0-9]+)?)(?:\s|$|[,;:)])`)

	// Relative paths without prefix but with directory separators: src/file.go, pkg/export/html.go
	relPathRe = regexp.MustCompile(`(?:^|\s)([a-zA-Z0-9_.~-]+/(?:[a-zA-Z0-9_.~-]+/)*[a-zA-Z0-9_.~-]+\.[a-zA-Z0-9]+)(?:\s|$|[,;:)])`)

	// Simple filename with extension (no path separators): test.go, main.py
	simpleFilenameRe = regexp.MustCompile(`(?:^|\s)([a-zA-Z0-9_~][a-zA-Z0-9_.~-]*\.[a-zA-Z0-9]+)(?:\s|$|[,;:.)])`)
)

// makePathsClickableWithPlaceholders scans text for file paths and converts them to placeholders.
// The placeholders map is populated with the link HTML for each placeholder.
// Only paths that exist on disk are converted to links.
// projectPath is used to resolve relative paths (can be empty for absolute-only detection).
// This function should be called AFTER code blocks are protected but BEFORE final HTML escaping.
func makePathsClickableWithPlaceholders(text string, projectPath string, placeholders *map[string]string, startIdx *int) string {
	if text == "" {
		return text
	}

	result := makePathsClickable(text, projectPath, func(pathStr string, linkHTML string, start int, end int) string {
		placeholder := fmt.Sprintf("\x00PATH_%d\x00", *startIdx)
		(*placeholders)[placeholder] = linkHTML
		*startIdx++
		return placeholder
	})

	return result
}

// makePathsClickable scans text for file paths and converts them to clickable file:// links.
// Only paths that exist on disk are converted to links.
// projectPath is used to resolve relative paths (can be empty for absolute-only detection).
// replacer is a function that takes (pathStr, linkHTML, start, end) and returns replacement text.
// This function should be called AFTER code blocks are protected but BEFORE final HTML escaping.
func makePathsClickable(text string, projectPath string, replacer func(string, string, int, int) string) string {
	if text == "" {
		return text
	}

	// Track replacements to avoid overlapping matches
	// We'll collect all matches first, then apply them in reverse order (longest first, rightmost first)
	type pathMatch struct {
		start      int
		end        int
		matchedStr string
		linkHTML   string
	}
	var matches []pathMatch

	// Helper to check if a path exists and create link HTML
	checkAndCreateLink := func(pathStr string, isAbsolute bool) (string, bool) {
		absPath := pathStr
		if !isAbsolute && projectPath != "" {
			absPath = filepath.Join(projectPath, pathStr)
		}

		// Normalize path (clean up .. and .)
		absPath = filepath.Clean(absPath)

		// Check if file exists
		if _, err := os.Stat(absPath); err == nil {
			// File exists, create link
			fileURL := buildFileURL(absPath)
			// Escape the displayed path text (not the URL)
			linkHTML := fmt.Sprintf(`<a href="%s" class="file-link" title="Open file: %s">%s</a>`,
				fileURL,
				escapeHTML(absPath),
				escapeHTML(pathStr))
			return linkHTML, true
		}
		return "", false
	}

	// Scan for Unix absolute paths
	for _, match := range unixAbsPathRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) >= 4 {
			pathStart := match[2]
			pathEnd := match[3]
			pathStr := text[pathStart:pathEnd]

			if linkHTML, ok := checkAndCreateLink(pathStr, true); ok {
				matches = append(matches, pathMatch{
					start:      pathStart,
					end:        pathEnd,
					matchedStr: pathStr,
					linkHTML:   linkHTML,
				})
			}
		}
	}

	// Scan for Windows absolute paths
	for _, match := range winAbsPathRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) >= 4 {
			pathStart := match[2]
			pathEnd := match[3]
			pathStr := text[pathStart:pathEnd]

			if linkHTML, ok := checkAndCreateLink(pathStr, true); ok {
				matches = append(matches, pathMatch{
					start:      pathStart,
					end:        pathEnd,
					matchedStr: pathStr,
					linkHTML:   linkHTML,
				})
			}
		}
	}

	// Scan for relative paths with ./ or ../ prefix
	for _, match := range relPathPrefixRe.FindAllStringSubmatchIndex(text, -1) {
		if len(match) >= 4 {
			pathStart := match[2]
			pathEnd := match[3]
			pathStr := text[pathStart:pathEnd]

			if linkHTML, ok := checkAndCreateLink(pathStr, false); ok {
				matches = append(matches, pathMatch{
					start:      pathStart,
					end:        pathEnd,
					matchedStr: pathStr,
					linkHTML:   linkHTML,
				})
			}
		}
	}

	// Scan for relative paths without prefix (most conservative)
	// Skip if projectPath is empty (can't resolve)
	if projectPath != "" {
		for _, match := range relPathRe.FindAllStringSubmatchIndex(text, -1) {
			if len(match) >= 4 {
				pathStart := match[2]
				pathEnd := match[3]
				pathStr := text[pathStart:pathEnd]

				// Skip if this looks like a URL (http://example.com/path)
				if strings.Contains(pathStr, "://") {
					continue
				}

				if linkHTML, ok := checkAndCreateLink(pathStr, false); ok {
					matches = append(matches, pathMatch{
						start:      pathStart,
						end:        pathEnd,
						matchedStr: pathStr,
						linkHTML:   linkHTML,
					})
				}
			}
		}

		// Scan for simple filenames (no path separators)
		// Only when projectPath is available
		for _, match := range simpleFilenameRe.FindAllStringSubmatchIndex(text, -1) {
			if len(match) >= 4 {
				pathStart := match[2]
				pathEnd := match[3]
				pathStr := text[pathStart:pathEnd]

				// Skip common false positives
				// Skip if looks like a URL (example.com)
				if strings.Contains(pathStr, "://") {
					continue
				}

				// Skip if starts with www. (likely a domain)
				if strings.HasPrefix(pathStr, "www.") {
					continue
				}

				if linkHTML, ok := checkAndCreateLink(pathStr, false); ok {
					matches = append(matches, pathMatch{
						start:      pathStart,
						end:        pathEnd,
						matchedStr: pathStr,
						linkHTML:   linkHTML,
					})
				}
			}
		}
	}

	// If no matches, return original text
	if len(matches) == 0 {
		return text
	}

	// Sort matches by position (rightmost first to preserve indices)
	// This is a simple bubble sort since we expect few matches
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].start < matches[j].start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Remove overlapping matches (keep the first one encountered, which is rightmost)
	filtered := []pathMatch{}
	lastEnd := len(text) + 1
	for _, m := range matches {
		if m.end < lastEnd {
			filtered = append(filtered, m)
			lastEnd = m.start
		}
	}

	// Apply replacements from right to left
	result := text
	for _, m := range filtered {
		replacement := replacer(m.matchedStr, m.linkHTML, m.start, m.end)
		result = result[:m.start] + replacement + result[m.end:]
	}

	return result
}
