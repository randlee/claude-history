package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
)

var (
	findExplored  string
	findTools     string
	findToolMatch string
	findStart     string
	findEnd       string
	findSessionID string
)

var findAgentCmd = &cobra.Command{
	Use:   "find-agent <project-path>",
	Short: "Find agents matching search criteria",
	Long: `Find agents (main sessions and subagents) matching specified criteria.

Search by files explored, tool usage, time range, or combinations of these.

Examples:
  # Find agents that explored a file
  claude-history find-agent /path --explored "src/*.go"
  claude-history find-agent /path --explored "**/*.go"

  # Find agents in time range
  claude-history find-agent /path --start 2026-01-30 --end 2026-02-01

  # Find agents by tool usage
  claude-history find-agent /path --tool bash,read
  claude-history find-agent /path --tool-match "db\.go"

  # Combine filters
  claude-history find-agent /path --start 2026-01-30 --explored "*.go" --tool bash

  # Scope to single session
  claude-history find-agent /path --session abc123 --explored "*.go"

  # JSON output for scripting
  claude-history find-agent /path --explored "*.go" --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runFindAgent,
}

func init() {
	rootCmd.AddCommand(findAgentCmd)

	findAgentCmd.Flags().StringVar(&findExplored, "explored", "", "Glob pattern for files explored (Read/Write/Edit)")
	findAgentCmd.Flags().StringVar(&findTools, "tool", "", "Filter by tool types (comma-separated: bash,read,write)")
	findAgentCmd.Flags().StringVar(&findToolMatch, "tool-match", "", "Regex pattern for tool input")
	findAgentCmd.Flags().StringVar(&findStart, "start", "", "Start time filter (RFC3339 or YYYY-MM-DD)")
	findAgentCmd.Flags().StringVar(&findEnd, "end", "", "End time filter (RFC3339 or YYYY-MM-DD)")
	findAgentCmd.Flags().StringVar(&findSessionID, "session", "", "Scope to single session ID")
}

func runFindAgent(cmd *cobra.Command, args []string) error {
	projectPath := args[0]

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
	if findSessionID != "" {
		resolvedSessionID, err = resolver.ResolveSessionID(projectPath, findSessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve session ID: %w", err)
		}
	}

	// Build find options with resolved session ID
	opts, err := buildFindOptions(resolvedSessionID)
	if err != nil {
		return err
	}

	// Find matching agents
	matches, err := agent.FindAgents(projectDir, opts)
	if err != nil {
		return fmt.Errorf("error searching agents: %w", err)
	}

	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "No agents found matching criteria")
		return nil
	}

	// Output results based on format
	return outputAgentMatches(matches, format)
}

func buildFindOptions(resolvedSessionID string) (agent.FindAgentsOptions, error) {
	var opts agent.FindAgentsOptions

	// Explored pattern
	opts.ExploredPattern = findExplored

	// Parse tool types
	if findTools != "" {
		tools := strings.Split(findTools, ",")
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
	opts.ToolMatch = findToolMatch

	// Parse start time
	if findStart != "" {
		t, err := parseFindTime(findStart)
		if err != nil {
			return opts, fmt.Errorf("invalid start date: %v", err)
		}
		opts.StartTime = t
	}

	// Parse end time
	if findEnd != "" {
		t, err := parseFindTime(findEnd)
		if err != nil {
			return opts, fmt.Errorf("invalid end date: %v", err)
		}
		opts.EndTime = t
	}

	// Session ID (use resolved session ID)
	opts.SessionID = resolvedSessionID

	return opts, nil
}

func parseFindTime(s string) (time.Time, error) {
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

func outputAgentMatches(matches []agent.AgentMatch, outputFormat string) error {
	switch strings.ToLower(outputFormat) {
	case "json":
		return outputAgentMatchesJSON(matches)
	default:
		return outputAgentMatchesList(matches)
	}
}

// agentMatchJSON is the JSON output structure for agent matches.
type agentMatchJSON struct {
	Agents []agentMatchEntry `json:"agents"`
}

type agentMatchEntry struct {
	AgentID      string   `json:"agentId"`
	SessionID    string   `json:"sessionId"`
	JSONLPath    string   `json:"jsonlPath"`
	EntryCount   int      `json:"entryCount"`
	MatchedFiles []string `json:"matchedFiles,omitempty"`
	MatchedTools []string `json:"matchedTools,omitempty"`
	Created      string   `json:"created"`
}

func outputAgentMatchesJSON(matches []agent.AgentMatch) error {
	output := agentMatchJSON{
		Agents: make([]agentMatchEntry, 0, len(matches)),
	}

	for _, m := range matches {
		entry := agentMatchEntry{
			AgentID:      m.AgentID,
			SessionID:    m.SessionID,
			JSONLPath:    m.JSONLPath,
			EntryCount:   m.EntryCount,
			MatchedFiles: m.MatchedFiles,
			MatchedTools: m.MatchedTools,
			Created:      m.Created.Format(time.RFC3339),
		}
		// Handle main sessions (empty agent ID)
		if entry.AgentID == "" {
			entry.AgentID = "(main)"
		}
		output.Agents = append(output.Agents, entry)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputAgentMatchesList(matches []agent.AgentMatch) error {
	fmt.Printf("Found %d matching agents:\n\n", len(matches))

	for _, m := range matches {
		// Agent ID display
		agentDisplay := m.AgentID
		if agentDisplay == "" {
			agentDisplay = "(main session)"
		}

		fmt.Printf("Agent: %s\n", agentDisplay)
		fmt.Printf("  Session: %s\n", m.SessionID)
		fmt.Printf("  Path: %s\n", m.JSONLPath)
		fmt.Printf("  Entries: %d\n", m.EntryCount)

		// Files (if any)
		if len(m.MatchedFiles) > 0 {
			// Sort files for consistent output
			sortedFiles := make([]string, len(m.MatchedFiles))
			copy(sortedFiles, m.MatchedFiles)
			sort.Strings(sortedFiles)
			fmt.Printf("  Files: %s\n", strings.Join(sortedFiles, ", "))
		}

		// Tools (if any)
		if len(m.MatchedTools) > 0 {
			// Sort tools for consistent output, capitalize first letter
			sortedTools := make([]string, len(m.MatchedTools))
			copy(sortedTools, m.MatchedTools)
			sort.Strings(sortedTools)
			capitalizedTools := make([]string, len(sortedTools))
			for i, t := range sortedTools {
				capitalizedTools[i] = capitalizeFirst(t)
			}
			fmt.Printf("  Tools: %s\n", strings.Join(capitalizedTools, ", "))
		}

		// Created time
		fmt.Printf("  Created: %s\n", m.Created.Format(time.RFC3339))
		fmt.Println()
	}

	return nil
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
