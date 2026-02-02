package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/output"
	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
	"github.com/randlee/claude-history/pkg/session"
)

var (
	treeSessionID string
	treeDepth     int
)

var treeCmd = &cobra.Command{
	Use:   "tree <project-path>",
	Short: "Display agent hierarchy tree",
	Long: `Display the agent hierarchy for a Claude Code session.

Shows the main conversation and all spawned agents in a tree structure.

Examples:
  # Show tree for most recent session
  claude-history tree /path/to/project

  # Show tree for specific session
  claude-history tree /path/to/project --session 679761ba-80c0-4cd3-a586-cc6a1fc56308

  # Output formats
  claude-history tree /path/to/project --format ascii   # Default: ASCII art
  claude-history tree /path/to/project --format json    # JSON structure
  claude-history tree /path/to/project --format dot     # GraphViz DOT format`,
	Args: cobra.ExactArgs(1),
	RunE: runTree,
}

func init() {
	rootCmd.AddCommand(treeCmd)

	treeCmd.Flags().StringVar(&treeSessionID, "session", "", "Session ID to display")
	treeCmd.Flags().IntVar(&treeDepth, "depth", 0, "Maximum tree depth (0 = unlimited)")
}

func runTree(cmd *cobra.Command, args []string) error {
	projectPath := args[0]

	// Determine output format
	outputFormat := output.ParseFormat(format)
	if format == "" {
		outputFormat = output.FormatASCII
	}

	// Get the project directory
	projectDir, err := paths.ProjectDir(claudeDir, projectPath)
	if err != nil {
		return err
	}

	if !paths.Exists(projectDir) {
		return fmt.Errorf("project not found: %s", projectPath)
	}

	// Get session ID
	sessionID := treeSessionID
	if sessionID == "" {
		// Use most recent session
		sessions, err := session.ListSessions(projectDir)
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			return fmt.Errorf("no sessions found in project")
		}
		sessionID = sessions[0].ID
	} else {
		// Resolve session ID prefix
		resolvedSessionID, err := resolver.ResolveSessionID(projectPath, sessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve session ID: %w", err)
		}
		sessionID = resolvedSessionID
	}

	// Build the tree
	tree, err := agent.BuildTree(projectDir, sessionID)
	if err != nil {
		return err
	}

	// Write output
	return output.WriteTree(os.Stdout, tree, outputFormat)
}
