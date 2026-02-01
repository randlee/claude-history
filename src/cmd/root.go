package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	claudeDir string
	format    string
)

var rootCmd = &cobra.Command{
	Use:   "claude-history",
	Short: "Query and traverse Claude Code agent history",
	Long: `claude-history is a CLI tool that maps filesystem paths to Claude Code's
agent history storage, enabling querying and traversal of agent sessions.

It supports:
  - Path resolution between filesystem paths and Claude's encoded storage
  - Querying conversation history with date and type filters
  - Displaying agent hierarchy trees
  - Listing projects and sessions`,
	SilenceUsage: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&claudeDir, "claude-dir", "", "Custom ~/.claude directory location")
	rootCmd.PersistentFlags().StringVar(&format, "format", "", "Output format (json, path, list, summary, ascii, dot)")
}
