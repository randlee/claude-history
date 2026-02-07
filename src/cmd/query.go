package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/output"
	"github.com/randlee/claude-history/pkg/export"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
	"github.com/randlee/claude-history/pkg/session"
)

var (
	queryStart         string
	queryEnd           string
	queryTypes         string
	querySessionID     string
	queryAgentID       string
	queryTools         string // --tool flag
	queryToolMatch     string // --tool-match flag
	queryIncludeAgents bool   // --include-agents flag
	queryLimit         int    // --limit flag for text truncation (0 = no truncation)
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

  # Query specific agent (reads agent's JSONL file directly)
  claude-history query /path/to/project --session <session-id> --agent <agent-id>

  # Query session including all subagent entries
  claude-history query /path/to/project --session <session-id> --include-agents

  # Filter by tool type
  claude-history query /path/to/project --tool bash
  claude-history query /path/to/project --tool bash,read,write

  # Filter by tool input pattern
  claude-history query /path/to/project --tool bash --tool-match "git"

  # Output formats
  claude-history query /path/to/project --format json
  claude-history query /path/to/project --format summary
  claude-history query /path/to/project --format html

  # Control text truncation
  claude-history query /path/to/project --limit 0        # No truncation (full content)
  claude-history query /path/to/project --limit 500      # Truncate at 500 chars
  claude-history query /path/to/project --type assistant --limit 0  # Full assistant responses

Agent Queries:
  When --agent is specified, the command reads the agent's JSONL file directly
  instead of filtering the main session file. This provides accurate results
  for agent-specific queries, as agent entries are stored in separate files.

  When --include-agents is specified, entries from all subagents are included
  in the query results, recursively gathering entries from nested agents.`,
	Args: cobra.ExactArgs(1),
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVar(&queryStart, "start", "", "Start date (ISO 8601 format)")
	queryCmd.Flags().StringVar(&queryEnd, "end", "", "End date (ISO 8601 format)")
	queryCmd.Flags().StringVar(&queryTypes, "type", "", "Entry types to include (comma-separated: user,assistant,system)")
	queryCmd.Flags().StringVar(&querySessionID, "session", "", "Filter to specific session ID")
	queryCmd.Flags().StringVar(&queryAgentID, "agent", "", "Query specific agent (reads agent's JSONL file directly)")
	queryCmd.Flags().StringVar(&queryTools, "tool", "", "Filter by tool types (comma-separated: bash,read,write)")
	queryCmd.Flags().StringVar(&queryToolMatch, "tool-match", "", "Filter by tool input regex pattern")
	queryCmd.Flags().BoolVar(&queryIncludeAgents, "include-agents", false, "Include entries from all subagents")
	queryCmd.Flags().IntVar(&queryLimit, "limit", 100, "Maximum characters per entry in text format (0 = no limit)")
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
		resolvedSessionID, err = resolver.ResolveSessionID(projectDir, querySessionID)
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
		resolvedAgentID, err = resolver.ResolveAgentID(projectDir, resolvedSessionID, queryAgentID)
		if err != nil {
			return fmt.Errorf("failed to resolve agent ID: %w", err)
		}
	}

	// Validate flag combinations
	if queryIncludeAgents && resolvedAgentID != "" {
		return fmt.Errorf("--include-agents and --agent cannot be used together")
	}

	// Build filter options (don't pass agent ID since we read agent file directly)
	filterOpts, err := buildFilterOptions("")
	if err != nil {
		return err
	}

	// Collect entries
	var allEntries []models.ConversationEntry

	if resolvedSessionID != "" {
		if resolvedAgentID != "" {
			// Query specific agent - read agent's JSONL file directly
			entries, err := queryAgentFile(projectDir, resolvedSessionID, resolvedAgentID, filterOpts)
			if err != nil {
				return err
			}
			allEntries = entries
		} else if queryIncludeAgents {
			// Query session including all subagent entries
			entries, err := querySessionWithAgents(projectDir, resolvedSessionID, filterOpts)
			if err != nil {
				return err
			}
			allEntries = entries
		} else {
			// Query main session file only
			entries, err := querySession(projectDir, resolvedSessionID, filterOpts)
			if err != nil {
				return err
			}
			allEntries = entries
		}
	} else {
		// Query all sessions in project
		sessions, err := session.ListSessions(projectDir)
		if err != nil {
			return err
		}

		for _, s := range sessions {
			var entries []models.ConversationEntry
			var queryErr error

			if queryIncludeAgents {
				entries, queryErr = querySessionWithAgents(projectDir, s.ID, filterOpts)
			} else {
				entries, queryErr = querySession(projectDir, s.ID, filterOpts)
			}
			if queryErr != nil {
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

	// Handle HTML format specially - generate and open HTML file
	if outputFormat == output.FormatHTML {
		// Build session folder path if we have a session ID
		sessionFolderPath := ""
		if resolvedSessionID != "" {
			sessionFolderPath = filepath.Join(projectDir, resolvedSessionID)
		}

		htmlFile, err := generateQueryHTML(projectPath, sessionFolderPath, allEntries, resolvedSessionID, resolvedAgentID)
		if err != nil {
			return fmt.Errorf("failed to generate HTML: %w", err)
		}
		fmt.Printf("HTML generated: %s\n", htmlFile)

		// Open in browser
		if err := openBrowser(htmlFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
		}
		return nil
	}

	return output.WriteEntries(os.Stdout, allEntries, outputFormat, queryLimit)
}

func querySession(projectDir string, sessionID string, opts session.FilterOptions) ([]models.ConversationEntry, error) {
	filePath := filepath.Join(projectDir, sessionID+".jsonl")

	if !paths.Exists(filePath) {
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

// getAgentPath returns the path to an agent's JSONL file.
// It first checks the standard location, then falls back to recursive search for nested agents.
func getAgentPath(projectDir, sessionID, agentID string) (string, error) {
	sessionDir := filepath.Join(projectDir, sessionID)

	// Try standard location first
	agentPath := filepath.Join(sessionDir, "subagents", "agent-"+agentID+".jsonl")
	if paths.Exists(agentPath) {
		return agentPath, nil
	}

	// Fallback: recursive search for nested agents
	agentFiles, err := paths.ListAgentFiles(sessionDir)
	if err == nil {
		if path, ok := agentFiles[agentID]; ok {
			return path, nil
		}
	}

	return "", fmt.Errorf("agent not found: %s", agentID)
}

// queryAgentFile reads and queries an agent's JSONL file directly.
func queryAgentFile(projectDir, sessionID, agentID string, opts session.FilterOptions) ([]models.ConversationEntry, error) {
	agentPath, err := getAgentPath(projectDir, sessionID, agentID)
	if err != nil {
		return nil, err
	}

	entries, err := session.ReadSession(agentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent file: %w", err)
	}

	// Apply filters
	filtered := session.FilterEntries(entries, opts)

	return filtered, nil
}

// querySessionWithAgents queries the main session and all subagent files.
func querySessionWithAgents(projectDir, sessionID string, opts session.FilterOptions) ([]models.ConversationEntry, error) {
	var allEntries []models.ConversationEntry

	// First, query the main session file
	mainEntries, err := querySession(projectDir, sessionID, opts)
	if err != nil {
		return nil, err
	}
	allEntries = append(allEntries, mainEntries...)

	// Then, query all agent files
	sessionDir := filepath.Join(projectDir, sessionID)
	agentFiles, err := paths.ListAgentFiles(sessionDir)
	if err != nil {
		// No agents or error listing - just return main entries
		return allEntries, nil
	}

	for _, agentPath := range agentFiles {
		entries, err := session.ReadSession(agentPath)
		if err != nil {
			// Skip agents that can't be read
			continue
		}

		// Apply filters
		filtered := session.FilterEntries(entries, opts)
		allEntries = append(allEntries, filtered...)
	}

	return allEntries, nil
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

// generateQueryHTML generates an HTML file for query results and returns the file path.
func generateQueryHTML(projectPath, sessionFolderPath string, entries []models.ConversationEntry, sessionID, agentID string) (string, error) {
	// Create temp file with descriptive name
	var fileName string
	if agentID != "" {
		// Truncate agent ID safely
		truncated := agentID
		if len(truncated) > 8 {
			truncated = truncated[:8]
		}
		fileName = fmt.Sprintf("query-%s.html", truncated)
	} else if sessionID != "" {
		// Truncate session ID safely
		truncated := sessionID
		if len(truncated) > 8 {
			truncated = truncated[:8]
		}
		fileName = fmt.Sprintf("query-%s.html", truncated)
	} else {
		fileName = "query-results.html"
	}
	tmpFile := filepath.Join(os.TempDir(), fileName)

	// Determine role labels based on context
	userLabel := "User"
	assistantLabel := "Assistant"
	if agentID != "" {
		// For subagent queries, use Orchestrator/Agent labels
		userLabel = "Orchestrator"
		assistantLabel = "Agent"
	}

	// Render entries as HTML using export package
	htmlContent, err := export.RenderQueryResults(entries, projectPath, sessionID, sessionFolderPath, agentID, userLabel, assistantLabel)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(tmpFile, []byte(htmlContent), 0644); err != nil {
		return "", err
	}

	return tmpFile, nil
}

// openBrowser opens a URL or file path in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
