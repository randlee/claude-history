// Package output provides output formatting for CLI display.
//
//nolint:errcheck // CLI output errors are unrecoverable - writing to stdout/stderr
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/randlee/claude-history/pkg/models"
)

// Format represents an output format type.
type Format string

const (
	FormatJSON    Format = "json"
	FormatList    Format = "list"
	FormatSummary Format = "summary"
	FormatASCII   Format = "ascii"
	FormatDOT     Format = "dot"
	FormatPath    Format = "path"
)

// ParseFormat parses a format string, returning FormatList as default.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "list":
		return FormatList
	case "summary":
		return FormatSummary
	case "ascii":
		return FormatASCII
	case "dot":
		return FormatDOT
	case "path":
		return FormatPath
	default:
		return FormatList
	}
}

// WriteJSON writes data as formatted JSON.
func WriteJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteList writes items as a simple list.
func WriteList(w io.Writer, items []string) {
	for _, item := range items {
		fmt.Fprintln(w, item)
	}
}

// WriteSessions writes sessions in list format.
func WriteSessions(w io.Writer, sessions []models.Session, format Format) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, sessions)
	default:
		for _, s := range sessions {
			modified := s.Modified.Format(time.RFC3339)
			prompt := s.FirstPrompt
			if len(prompt) > 50 {
				prompt = prompt[:50] + "..."
			}
			fmt.Fprintf(w, "%s  %s  %d msgs  %s\n", s.ID, modified, s.MessageCount, prompt)
		}
	}
	return nil
}

// WriteProjects writes projects in list format.
func WriteProjects(w io.Writer, projects []models.Project, format Format) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, projects)
	default:
		for _, p := range projects {
			fmt.Fprintf(w, "%s\n  Path: %s\n", p.Name, p.ProjectPath)
		}
	}
	return nil
}

// WriteEntries writes conversation entries.
func WriteEntries(w io.Writer, entries []models.ConversationEntry, format Format, limit int) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, entries)
	case FormatSummary:
		return writeEntrySummary(w, entries)
	default:
		return writeEntryList(w, entries, limit)
	}
}

func writeEntryList(w io.Writer, entries []models.ConversationEntry, limit int) error {
	// Filter out entries with no text content first
	var textEntries []models.ConversationEntry
	for _, e := range entries {
		if e.GetTextContent() != "" {
			textEntries = append(textEntries, e)
		}
	}

	if len(textEntries) == 0 {
		return nil
	}

	// Default mode (limit=100): Show preview format
	if limit == 100 && len(textEntries) > 2 {
		return writeEntryPreview(w, textEntries)
	}

	// Full output mode (limit=0) or custom limit: Show all entries
	for _, e := range textEntries {
		ts, _ := e.GetTimestamp()
		text := e.GetTextContent()

		// Apply truncation if limit > 0
		if limit > 0 && len(text) > limit {
			text = text[:limit] + "..."
		}
		text = strings.ReplaceAll(text, "\n", " ")
		fmt.Fprintf(w, "[%s] %s: %s\n", ts.Format("15:04:05"), e.Type, text)
	}
	return nil
}

// writeEntryPreview shows first entry, count, and last entry with preview
func writeEntryPreview(w io.Writer, entries []models.ConversationEntry) error {
	first := entries[0]
	last := entries[len(entries)-1]

	// Show first entry
	firstTS, _ := first.GetTimestamp()
	firstText := first.GetTextContent()
	if len(firstText) > 100 {
		firstText = firstText[:100] + "..."
	}
	firstText = strings.ReplaceAll(firstText, "\n", " ")
	fmt.Fprintf(w, "[%s] %s: %s\n", firstTS.Format("15:04:05"), first.Type, firstText)

	// Show count of middle entries
	if len(entries) > 2 {
		fmt.Fprintf(w, "... (%d more entries) ...\n", len(entries)-2)
	}

	// Show last entry with first 10 lines of full text
	lastTS, _ := last.GetTimestamp()
	lastText := last.GetTextContent()

	fmt.Fprintf(w, "[%s] %s: ", lastTS.Format("15:04:05"), last.Type)

	// Split into lines and show first 10
	lines := strings.Split(lastText, "\n")
	previewLines := 10
	if len(lines) < previewLines {
		previewLines = len(lines)
	}

	for i := 0; i < previewLines; i++ {
		if i == 0 {
			fmt.Fprintf(w, "%s\n", lines[i])
		} else {
			fmt.Fprintf(w, "                     %s\n", lines[i])
		}
	}

	if len(lines) > previewLines {
		fmt.Fprintf(w, "                     ... (%d of %d lines shown - TRUNCATED)\n", previewLines, len(lines))
	}

	// Show help message with exact command
	fmt.Fprintf(w, "\n⚠️  Text format truncates long outputs. For full content:\n")
	fmt.Fprintf(w, "   --format json | jq -r '.[-1].message.content[0].text'\n")

	return nil
}

func writeEntrySummary(w io.Writer, entries []models.ConversationEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(w, "No entries")
		return nil
	}

	// Count by type
	typeCounts := make(map[models.EntryType]int)
	for _, e := range entries {
		typeCounts[e.Type]++
	}

	// Time range
	var firstTime, lastTime time.Time
	if ts, err := entries[0].GetTimestamp(); err == nil {
		firstTime = ts
	}
	if ts, err := entries[len(entries)-1].GetTimestamp(); err == nil {
		lastTime = ts
	}

	fmt.Fprintf(w, "Total entries: %d\n", len(entries))
	fmt.Fprintf(w, "Time range: %s to %s\n", firstTime.Format(time.RFC3339), lastTime.Format(time.RFC3339))
	fmt.Fprintln(w, "\nBreakdown by type:")
	for t, count := range typeCounts {
		fmt.Fprintf(w, "  %s: %d\n", t, count)
	}

	return nil
}

// WritePath writes a single path.
func WritePath(w io.Writer, path string) {
	fmt.Fprintln(w, path)
}

// ToolUse represents a tool invocation for formatting (matches pkg/models.ToolUse).
type ToolUse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// maxToolInputLength is the maximum length for tool input display before truncation.
const maxToolInputLength = 80

// FormatToolCall formats a single tool call for display.
// Returns formatted string like "[Bash] git status" or "[Read] /path/to/file.go".
func FormatToolCall(toolName string, input map[string]any) string {
	displayValue := extractToolDisplayValue(toolName, input)
	if displayValue == "" {
		return fmt.Sprintf("[%s]", toolName)
	}

	// Truncate if needed
	if len(displayValue) > maxToolInputLength {
		displayValue = displayValue[:maxToolInputLength-3] + "..."
	}

	return fmt.Sprintf("[%s] %s", toolName, displayValue)
}

// FormatToolCalls formats multiple tool calls for list output.
// Each tool is formatted on a separate line.
func FormatToolCalls(tools []ToolUse) string {
	if len(tools) == 0 {
		return ""
	}

	var lines []string
	for _, tool := range tools {
		lines = append(lines, FormatToolCall(tool.Name, tool.Input))
	}
	return strings.Join(lines, "\n")
}

// FormatToolSummary creates a short summary of tools used in an entry.
// Returns a compact format like "[Bash, Read, Write]" for multiple tools.
func FormatToolSummary(tools []ToolUse) string {
	if len(tools) == 0 {
		return ""
	}

	if len(tools) == 1 {
		return FormatToolCall(tools[0].Name, tools[0].Input)
	}

	// Multiple tools - just list the names
	var names []string
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return fmt.Sprintf("[%s]", strings.Join(names, ", "))
}

// extractToolDisplayValue extracts the most relevant display value from tool input.
func extractToolDisplayValue(toolName string, input map[string]any) string {
	if input == nil {
		return ""
	}

	// Tool-specific extraction
	switch toolName {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return cmd
		}
	case "Read":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Write":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Edit":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Grep":
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
	case "Glob":
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
	case "Task":
		// Try description first, then prompt
		if desc, ok := input["description"].(string); ok {
			return desc
		}
		if prompt, ok := input["prompt"].(string); ok {
			return prompt
		}
	}

	// Default: JSON serialize the input
	return serializeInput(input)
}

// serializeInput converts input map to a compact JSON string for display.
func serializeInput(input map[string]any) string {
	if len(input) == 0 {
		return ""
	}

	data, err := json.Marshal(input)
	if err != nil {
		return ""
	}
	return string(data)
}
