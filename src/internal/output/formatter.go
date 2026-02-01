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
func WriteEntries(w io.Writer, entries []models.ConversationEntry, format Format) error {
	switch format {
	case FormatJSON:
		return WriteJSON(w, entries)
	case FormatSummary:
		return writeEntrySummary(w, entries)
	default:
		return writeEntryList(w, entries)
	}
}

func writeEntryList(w io.Writer, entries []models.ConversationEntry) error {
	for _, e := range entries {
		ts, _ := e.GetTimestamp()
		text := e.GetTextContent()
		if len(text) > 100 {
			text = text[:100] + "..."
		}
		text = strings.ReplaceAll(text, "\n", " ")
		fmt.Fprintf(w, "[%s] %s: %s\n", ts.Format("15:04:05"), e.Type, text)
	}
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
