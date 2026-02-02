// Package cmd provides CLI commands for claude-history.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/export"
	"github.com/randlee/claude-history/pkg/models"
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

	// Export JSONL files
	opts = export.ExportOptions{
		OutputDir: outputDir,
		ClaudeDir: claudeDir,
	}
	result2, err := export.ExportSession(projectPath, exportSessionID, opts)
	if err != nil {
		return fmt.Errorf("failed to export session: %w", err)
	}

	// Report any non-fatal errors
	if len(result2.Errors) > 0 {
		for _, errMsg := range result2.Errors {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", errMsg)
		}
	}

	fmt.Fprintf(os.Stderr, "✓ JSONL files exported (%d agents)\n", result.TotalAgents)

	// If HTML format requested, generate HTML pages
	if exportFormat == "html" {
		if err := renderHTML(result, projectPath, projectDir, exportSessionID); err != nil {
			// Non-fatal: JSONL files are already exported
			fmt.Fprintf(os.Stderr, "Warning: HTML rendering failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "✓ HTML export completed\n")
		}
	}

	// Print the output location (stdout for scripting)
	fmt.Println(outputDir)

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

// renderHTML generates HTML pages for the exported session.
func renderHTML(result *export.ExportResult, projectPath, projectDir, sessionID string) error {
	// 1. Read main session entries
	entries, err := jsonl.ReadAll[models.ConversationEntry](result.MainSessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session: %w", err)
	}

	// 2. Build agent tree
	agentTree, err := agent.BuildNestedTree(projectDir, sessionID)
	if err != nil {
		return fmt.Errorf("failed to build agent tree: %w", err)
	}

	// Convert tree to slice for RenderConversation
	var agentNodes []*agent.TreeNode
	if agentTree != nil && len(agentTree.Children) > 0 {
		agentNodes = agentTree.Children
	}

	// 3. Render main conversation HTML
	htmlContent, err := export.RenderConversation(entries, agentNodes)
	if err != nil {
		return fmt.Errorf("failed to render conversation: %w", err)
	}

	// 4. Write index.html
	indexPath := filepath.Join(result.OutputDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %w", err)
	}

	// 5. Render agent fragments
	if err := renderAgentFragments(result, agentTree); err != nil {
		// Non-fatal: log warning and continue
		fmt.Fprintf(os.Stderr, "Warning: some agent fragments failed: %v\n", err)
	}

	// 6. Write static assets (CSS, JS)
	if err := export.WriteStaticAssets(result.OutputDir); err != nil {
		return fmt.Errorf("failed to write static assets: %w", err)
	}

	// 7. Generate and write manifest.json
	manifest, err := export.GenerateManifest(projectDir, sessionID, result.OutputDir)
	if err != nil {
		// Non-fatal: log warning
		fmt.Fprintf(os.Stderr, "Warning: failed to generate manifest: %v\n", err)
	} else {
		if err := export.WriteManifest(manifest, result.OutputDir); err != nil {
			// Non-fatal: log warning
			fmt.Fprintf(os.Stderr, "Warning: failed to write manifest: %v\n", err)
		}
	}

	return nil
}

// renderAgentFragments renders HTML fragments for each agent.
func renderAgentFragments(result *export.ExportResult, agentTree *agent.TreeNode) error {
	// Create agents/ directory
	agentsDir := filepath.Join(result.OutputDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return err
	}

	// Render each agent
	var errors []string
	for agentID, agentFile := range result.AgentFiles {
		// Read agent entries
		entries, err := jsonl.ReadAll[models.ConversationEntry](agentFile)
		if err != nil {
			errors = append(errors, fmt.Sprintf("agent %s: %v", truncateAgentID(agentID), err))
			continue
		}

		// Render agent fragment
		htmlContent, err := export.RenderAgentFragment(agentID, entries)
		if err != nil {
			errors = append(errors, fmt.Sprintf("agent %s: %v", truncateAgentID(agentID), err))
			continue
		}

		// Write agent HTML
		agentPath := filepath.Join(agentsDir, truncateAgentID(agentID)+".html")
		if err := os.WriteFile(agentPath, []byte(htmlContent), 0644); err != nil {
			errors = append(errors, fmt.Sprintf("agent %s: %v", truncateAgentID(agentID), err))
			continue
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d agent(s) failed: %s", len(errors), strings.Join(errors, ", "))
	}
	return nil
}

// truncateAgentID returns the first 8 characters of an agent ID for display.
func truncateAgentID(agentID string) string {
	if len(agentID) > 8 {
		return agentID[:8]
	}
	return agentID
}
