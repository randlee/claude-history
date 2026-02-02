package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/output"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
	"github.com/randlee/claude-history/pkg/session"
)

var (
	queryStart     string
	queryEnd       string
	queryTypes     string
	querySessionID string
	queryAgentID   string
	queryTools     string // --tool flag
	queryToolMatch string // --tool-match flag
)

// knownTools is used for validation warnings when unknown tool types are specified
var knownTools = map[string]bool{
	"bash": true, "read": true, "write": true, "edit": true,
	"task": true, "glob": true, "grep": true, "webfetch": true,
	"websearch": true, "notebookedit": true, "askuserquestion": true,
}

var queryCmd = &cobra.Command{
	Use:   "query <project-path>",
	Short: "Query conversation history",
	Long: `Query and filter conversation history from Claude Code sessions.

Examples:
  # Query all entries in a project
  claude-history query /path/to/project

  # Filter by date range
  claude-history query /path/to/project --start 2026-01-01 --end 2026-02-01

  # Filter by entry type
  claude-history query /path/to/project --type user,assistant

  # Query specific session
  claude-history query /path/to/project --session 679761ba-80c0-4cd3-a586-cc6a1fc56308

  # Filter by tool type
  claude-history query /path/to/project --tool bash
  claude-history query /path/to/project --tool bash,read,write

  # Filter by tool input pattern
  claude-history query /path/to/project --tool bash --tool-match "git"

  # Output formats
  claude-history query /path/to/project --format json
  claude-history query /path/to/project --format summary`,
	Args: cobra.ExactArgs(1),
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVar(&queryStart, "start", "", "Start date (ISO 8601 format)")
	queryCmd.Flags().StringVar(&queryEnd, "end", "", "End date (ISO 8601 format)")
	queryCmd.Flags().StringVar(&queryTypes, "type", "", "Entry types to include (comma-separated: user,assistant,system)")
	queryCmd.Flags().StringVar(&querySessionID, "session", "", "Filter to specific session ID")
	queryCmd.Flags().StringVar(&queryAgentID, "agent", "", "Filter to specific agent ID")
	queryCmd.Flags().StringVar(&queryTools, "tool", "", "Filter by tool types (comma-separated: bash,read,write)")
	queryCmd.Flags().StringVar(&queryToolMatch, "tool-match", "", "Filter by tool input regex pattern")
}

func runQuery(cmd *cobra.Command, args []string) error {
	projectPath := args[0]
	outputFormat := output.ParseFormat(format)

	// Get the project directory
	projectDir, err := paths.ProjectDir(claudeDir, projectPath)
	if err != nil {
		return err
	}

	if !paths.Exists(projectDir) {
		return fmt.Errorf("project not found: %s", projectPath)
	}

	// Resolve session ID prefix if provided
	var resolvedSessionID string
	if querySessionID != "" {
		resolvedSessionID, err = resolver.ResolveSessionID(claudeDir, projectPath, querySessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve session ID: %w", err)
		}
	}

	// Resolve agent ID prefix if provided
	var resolvedAgentID string
	if queryAgentID != "" {
		if resolvedSessionID == "" {
			return fmt.Errorf("--agent requires --session to be specified")
		}
		resolvedAgentID, err = resolver.ResolveAgentID(claudeDir, projectPath, resolvedSessionID, queryAgentID)
		if err != nil {
			return fmt.Errorf("failed to resolve agent ID: %w", err)
		}
	}

	// Build filter options with resolved IDs
	filterOpts, err := buildFilterOptions(resolvedAgentID)
	if err != nil {
		return err
	}

	// Collect entries
	var allEntries []models.ConversationEntry

	if resolvedSessionID != "" {
		// Query specific session
		entries, err := querySession(projectDir, resolvedSessionID, filterOpts)
		if err != nil {
			return err
		}
		allEntries = entries
	} else {
		// Query all sessions in project
		sessions, err := session.ListSessions(projectDir)
		if err != nil {
			return err
		}

		for _, s := range sessions {
			entries, err := querySession(projectDir, s.ID, filterOpts)
			if err != nil {
				// Skip sessions that can't be read
				continue
			}
			allEntries = append(allEntries, entries...)
		}
	}

	if len(allEntries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries found matching criteria")
		return nil
	}

	return output.WriteEntries(os.Stdout, allEntries, outputFormat)
}

func querySession(projectDir string, sessionID string, opts session.FilterOptions) ([]models.ConversationEntry, error) {
	sessionPath := paths.Exists
	filePath := projectDir + "/" + sessionID + ".jsonl"

	if !sessionPath(filePath) {
		return nil, fmt.Errorf("session file not found: %s", filePath)
	}

	entries, err := session.ReadSession(filePath)
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := session.FilterEntries(entries, opts)

	return filtered, nil
}

func buildFilterOptions(resolvedAgentID string) (session.FilterOptions, error) {
	var opts session.FilterOptions

	// Parse start time
	if queryStart != "" {
		t, err := parseTime(queryStart)
		if err != nil {
			return opts, fmt.Errorf("invalid start date: %v", err)
		}
		opts.StartTime = &t
	}

	// Parse end time
	if queryEnd != "" {
		t, err := parseTime(queryEnd)
		if err != nil {
			return opts, fmt.Errorf("invalid end date: %v", err)
		}
		opts.EndTime = &t
	}

	// Parse types
	if queryTypes != "" {
		types := strings.Split(queryTypes, ",")
		for _, t := range types {
			t = strings.TrimSpace(t)
			switch t {
			case "user":
				opts.Types = append(opts.Types, models.EntryTypeUser)
			case "assistant":
				opts.Types = append(opts.Types, models.EntryTypeAssistant)
			case "system":
				opts.Types = append(opts.Types, models.EntryTypeSystem)
			case "queue-operation":
				opts.Types = append(opts.Types, models.EntryTypeQueueOperation)
			default:
				return opts, fmt.Errorf("unknown entry type: %s", t)
			}
		}
	}

	// Agent ID (use resolved agent ID)
	opts.AgentID = resolvedAgentID

	// Parse tool types
	if queryTools != "" {
		tools := strings.Split(queryTools, ",")
		for _, tool := range tools {
			tool = strings.TrimSpace(tool)
			if tool == "" {
				continue
			}
			// Warn on unknown tools (but still allow them)
			if !knownTools[strings.ToLower(tool)] {
				fmt.Fprintf(os.Stderr, "Warning: unknown tool type: %s\n", tool)
			}
			opts.ToolTypes = append(opts.ToolTypes, tool)
		}
	}

	// Tool match pattern
	opts.ToolMatch = queryToolMatch

	return opts, nil
}

func parseTime(s string) (time.Time, error) {
	// Try various formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", s)
}
