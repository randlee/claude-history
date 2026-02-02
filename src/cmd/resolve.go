package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/output"
	"github.com/randlee/claude-history/pkg/encoding"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
)

var (
	resolveSessionID string
	resolveAgentID   string
)

var resolveCmd = &cobra.Command{
	Use:   "resolve [path]",
	Short: "Resolve paths between filesystem and Claude storage",
	Long: `Resolve converts between filesystem paths and Claude Code's encoded storage paths.

Examples:
  # From filesystem path → encoded project directory
  claude-history resolve /Users/randlee/Documents/github/project

  # From session ID → JSONL file path (requires project path or --session)
  claude-history resolve /path/to/project --session 679761ba-80c0-4cd3-a586-cc6a1fc56308

  # From agent ID → agent JSONL path (requires session)
  claude-history resolve /path/to/project --session <sessionId> --agent a12eb64`,
	Args: cobra.MaximumNArgs(1),
	RunE: runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)

	resolveCmd.Flags().StringVar(&resolveSessionID, "session", "", "Session ID to resolve")
	resolveCmd.Flags().StringVar(&resolveAgentID, "agent", "", "Agent ID to resolve (requires --session)")
}

func runResolve(cmd *cobra.Command, args []string) error {
	// Determine output format
	outputFormat := output.ParseFormat(format)
	if format == "" {
		outputFormat = output.FormatPath
	}

	// If we have an agent ID, we need a session ID
	if resolveAgentID != "" && resolveSessionID == "" {
		return fmt.Errorf("--agent requires --session")
	}

	// Get project path from args or try to resolve from session
	var projectPath string
	if len(args) > 0 {
		projectPath = args[0]
	}

	// If we have a session ID and agent ID, resolve agent file
	if resolveSessionID != "" && resolveAgentID != "" {
		if projectPath == "" {
			return fmt.Errorf("project path required with --session and --agent")
		}

		// Resolve session ID prefix
		fullSessionID, err := resolver.ResolveSessionID(claudeDir, projectPath, resolveSessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve session ID: %w", err)
		}

		// Resolve agent ID prefix
		fullAgentID, err := resolver.ResolveAgentID(claudeDir, projectPath, fullSessionID, resolveAgentID)
		if err != nil {
			return fmt.Errorf("failed to resolve agent ID: %w", err)
		}

		agentPath, err := paths.AgentFile(claudeDir, projectPath, fullSessionID, fullAgentID)
		if err != nil {
			return err
		}

		return outputResult(agentPath, outputFormat)
	}

	// If we have just a session ID, resolve session file
	if resolveSessionID != "" {
		if projectPath == "" {
			// Try to find the session across all projects
			return resolveSessionGlobal(resolveSessionID, outputFormat)
		}

		// Resolve session ID prefix
		fullSessionID, err := resolver.ResolveSessionID(claudeDir, projectPath, resolveSessionID)
		if err != nil {
			return fmt.Errorf("failed to resolve session ID: %w", err)
		}

		sessionPath, err := paths.SessionFile(claudeDir, projectPath, fullSessionID)
		if err != nil {
			return err
		}

		return outputResult(sessionPath, outputFormat)
	}

	// If we have just a project path, resolve to encoded directory
	if projectPath != "" {
		projectDir, err := paths.ProjectDir(claudeDir, projectPath)
		if err != nil {
			return err
		}

		return outputResult(projectDir, outputFormat)
	}

	return fmt.Errorf("provide a path or --session flag")
}

func resolveSessionGlobal(sessionID string, outputFormat output.Format) error {
	// Search all projects for this session
	projects, err := paths.ListProjects(claudeDir)
	if err != nil {
		return err
	}

	for _, projectDir := range projects {
		sessionPath, err := paths.ListSessionFiles(projectDir)
		if err != nil {
			continue
		}

		if path, ok := sessionPath[sessionID]; ok {
			return outputResult(path, outputFormat)
		}
	}

	return fmt.Errorf("session %s not found", sessionID)
}

func outputResult(path string, format output.Format) error {
	switch format {
	case output.FormatJSON:
		result := struct {
			Path   string `json:"path"`
			Exists bool   `json:"exists"`
		}{
			Path:   path,
			Exists: paths.Exists(path),
		}
		return output.WriteJSON(os.Stdout, result)
	default:
		// Check if path exists and provide feedback
		if !paths.Exists(path) {
			fmt.Fprintf(os.Stderr, "Warning: path does not exist\n")
		}
		fmt.Println(path)
	}
	return nil
}

// EncodePathCmd is a utility subcommand for testing path encoding.
var encodePathCmd = &cobra.Command{
	Use:    "encode [path]",
	Short:  "Encode a filesystem path to Claude's format",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		encoded := encoding.EncodePath(args[0])
		fmt.Println(encoded)
	},
}

// DecodePathCmd is a utility subcommand for testing path decoding.
var decodePathCmd = &cobra.Command{
	Use:    "decode [encoded-path]",
	Short:  "Decode a Claude-encoded path back to filesystem format",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		decoded := encoding.DecodePath(args[0], "")
		fmt.Println(decoded)
	},
}

func init() {
	resolveCmd.AddCommand(encodePathCmd)
	resolveCmd.AddCommand(decodePathCmd)
}
