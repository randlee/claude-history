// Package cmd provides CLI commands for claude-history.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/pkg/export"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/session"
)

var (
	exportSessionID string
	exportOutputDir string
	exportFormat    string
)

var exportCmd = &cobra.Command{
	Use:   "export [project-path]",
	Short: "Export session to HTML or JSONL format",
	Long: `Export a Claude Code session to a shareable format.

HTML format creates a standalone folder with:
- index.html: Main conversation view
- agents/*.html: Lazy-loaded subagent content
- source/*.jsonl: Original session files for resurrection
- manifest.json: Metadata and tree structure
- style.css, script.js: Static assets

JSONL format copies only the source files.

Examples:
  # Export to HTML (default format)
  claude-history export /path/to/project --session abc123

  # Export to specific folder
  claude-history export /path/to/project --session abc123 --output ./my-export/

  # Export just JSONL (smaller, for backup/restore)
  claude-history export /path/to/project --session abc123 --format jsonl`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportSessionID, "session", "s", "", "Session ID (required)")
	exportCmd.Flags().StringVarP(&exportOutputDir, "output", "o", "", "Output directory (auto-generated if not specified)")
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "html", "Export format: html or jsonl")
	_ = exportCmd.MarkFlagRequired("session")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Get project path (default to current directory)
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}

	// Resolve to absolute path if needed
	if !filepath.IsAbs(projectPath) {
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return fmt.Errorf("failed to resolve project path: %w", err)
		}
		projectPath = absPath
	}

	// Validate format
	if exportFormat != "html" && exportFormat != "jsonl" {
		return fmt.Errorf("invalid format: %s (must be 'html' or 'jsonl')", exportFormat)
	}

	// Get the project directory in Claude's storage
	projectDir, err := paths.ProjectDir(claudeDir, projectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project directory: %w", err)
	}

	if !paths.Exists(projectDir) {
		return fmt.Errorf("project not found: %s", projectPath)
	}

	// Validate session exists
	sessionFile := filepath.Join(projectDir, exportSessionID+".jsonl")
	if !paths.Exists(sessionFile) {
		return fmt.Errorf("session not found: %s", exportSessionID)
	}

	// Get session info for display
	sessionInfo, err := session.GetSessionInfo(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session: %w", err)
	}

	// Generate output directory if not specified
	outputDir := exportOutputDir
	if outputDir == "" {
		outputDir = generateTempExportPath(exportSessionID)
	}

	// Resolve output directory to absolute path
	if !filepath.IsAbs(outputDir) {
		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("failed to resolve output path: %w", err)
		}
		outputDir = absPath
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Prepare export options
	opts := export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}

	// Call export
	result, err := export.ExportSession(projectPath, exportSessionID, opts)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Report export parameters
	fmt.Fprintf(os.Stderr, "Exporting session %s\n", exportSessionID[:8])
	fmt.Fprintf(os.Stderr, "  Project: %s\n", projectPath)
	fmt.Fprintf(os.Stderr, "  Format: %s\n", exportFormat)
	fmt.Fprintf(os.Stderr, "  Output: %s\n", result.OutputDir)
	if sessionInfo.FirstPrompt != "" {
		fmt.Fprintf(os.Stderr, "  First prompt: %s\n", truncateString(sessionInfo.FirstPrompt, 60))
	}
	fmt.Fprintf(os.Stderr, "  Total agents: %d\n", result.TotalAgents)
	fmt.Fprintln(os.Stderr)

	// For HTML format: generate HTML (Phase 8b will implement this)
	if exportFormat == "html" {
		fmt.Fprintln(os.Stderr, "Warning: HTML export not yet implemented, JSONL files copied only")
	}

	// Print success message
	fmt.Fprintf(os.Stderr, "✓ Export created at: %s\n", result.OutputDir)

	// Print warnings if any
	if len(result.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nWarnings encountered:")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
	}

	// Output directory path for scripting (last line)
	fmt.Println(result.OutputDir)

	return nil
}

// generateTempExportPath creates a temporary export path based on session ID and timestamp.
// Format: {tempdir}/claude-history/{sessionId[:8]}-{timestamp}/
func generateTempExportPath(sessionID string) string {
	prefix := sessionID
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	return filepath.Join(os.TempDir(), "claude-history", fmt.Sprintf("%s-%s", prefix, timestamp))
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
