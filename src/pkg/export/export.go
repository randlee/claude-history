// Package export handles exporting Claude Code session history.
package export

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/paths"
	"github.com/randlee/claude-history/pkg/resolver"
	"github.com/randlee/claude-history/pkg/session"
)

// ExportResult contains the result of an export operation.
type ExportResult struct {
	// OutputDir is the directory where the export was created.
	OutputDir string `json:"outputDir"`

	// SessionID is the ID of the exported session.
	SessionID string `json:"sessionId"`

	// SourceDir is the path to the source/ subdirectory containing JSONL files.
	SourceDir string `json:"sourceDir"`

	// MainSessionFile is the path to the main session JSONL file.
	MainSessionFile string `json:"mainSessionFile"`

	// AgentFiles is a map of agent ID to copied JSONL file paths.
	AgentFiles map[string]string `json:"agentFiles,omitempty"`

	// TotalAgents is the total number of agents (including nested).
	TotalAgents int `json:"totalAgents"`

	// Errors contains any non-fatal errors encountered during export.
	Errors []string `json:"errors,omitempty"`
}

// ExportOptions configures the export operation.
type ExportOptions struct {
	// OutputDir specifies the output directory. If empty, a temp directory is created.
	OutputDir string

	// ClaudeDir is the custom Claude directory. If empty, uses default ~/.claude.
	ClaudeDir string
}

// ExportSession exports a session's JSONL files to the specified output directory.
// If outputDir in options is empty, generates a temp folder with the session ID and timestamp.
// Supports session ID prefixes (like git) which are automatically resolved to full IDs.
func ExportSession(projectPath, sessionID string, opts ExportOptions) (*ExportResult, error) {
	// Resolve the project directory
	projectDir, err := paths.ProjectDir(opts.ClaudeDir, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project directory: %w", err)
	}

	// Resolve session ID prefix to full ID (supports partial IDs like git)
	resolvedSessionID, err := resolver.ResolveSessionID(projectDir, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve session ID: %w", err)
	}

	// Find the session
	sess, err := session.FindSession(projectDir, resolvedSessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir, err = generateTempPath(resolvedSessionID, sess.Modified)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp path: %w", err)
		}
	}

	// Create output directory structure
	sourceDir := filepath.Join(outputDir, "source")
	agentsDir := filepath.Join(sourceDir, "agents")

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	result := &ExportResult{
		OutputDir:  outputDir,
		SessionID:  resolvedSessionID,
		SourceDir:  sourceDir,
		AgentFiles: make(map[string]string),
	}

	// Copy main session file
	sessionFilePath := filepath.Join(projectDir, resolvedSessionID+".jsonl")
	destSessionFile := filepath.Join(sourceDir, "session.jsonl")
	if err := copyFile(sessionFilePath, destSessionFile); err != nil {
		return nil, fmt.Errorf("failed to copy session file: %w", err)
	}
	result.MainSessionFile = destSessionFile

	// Copy agent files recursively
	sessionDir := filepath.Join(projectDir, resolvedSessionID)
	if err := copyAgentFiles(sessionDir, agentsDir, result); err != nil {
		// Non-fatal: add to errors but continue
		result.Errors = append(result.Errors, fmt.Sprintf("error copying agent files: %v", err))
	}

	return result, nil
}

// generateTempPath creates a temp folder path with the session ID prefix and timestamp.
// Format: {os.TempDir()}/claude-history/{sessionId-prefix-8chars}-{ISO-timestamp}/
func generateTempPath(sessionID string, lastModified time.Time) (string, error) {
	// Get first 8 characters of session ID
	prefix := sessionID
	if len(prefix) > 8 {
		prefix = prefix[:8]
	}

	// Format timestamp as ISO 8601 with dashes (filesystem safe)
	// Use format: 2026-02-01T19-00-22 instead of 2026-02-01T19:00:22
	timestamp := lastModified.Format("2006-01-02T15-04-05")

	// Build path
	folderName := fmt.Sprintf("%s-%s", prefix, timestamp)
	tempDir := os.TempDir()

	return filepath.Join(tempDir, "claude-history", folderName), nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return dstFile.Close()
}

// copyAgentFiles recursively copies all agent JSONL files from a session directory.
// It preserves the nested directory structure for subagents.
func copyAgentFiles(sessionDir, destAgentsDir string, result *ExportResult) error {
	subagentsDir := filepath.Join(sessionDir, "subagents")

	// Check if subagents directory exists
	if !paths.Exists(subagentsDir) {
		return nil // No agents to copy
	}

	return copyAgentFilesRecursive(subagentsDir, destAgentsDir, "", result)
}

// copyAgentFilesRecursive recursively copies agent files, handling nested subagents.
func copyAgentFilesRecursive(srcDir, destDir, parentPath string, result *ExportResult) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())

		if entry.IsDir() {
			// This is an agent directory that may contain nested subagents
			// Directory structure: agent-{id}/ -> subagents/ -> agent-{nested-id}.jsonl
			nestedSubagentsDir := filepath.Join(srcPath, "subagents")
			if paths.Exists(nestedSubagentsDir) {
				// Create corresponding directory in destination
				relPath := entry.Name()
				if parentPath != "" {
					relPath = filepath.Join(parentPath, entry.Name())
				}
				nestedDestDir := filepath.Join(destDir, relPath, "subagents")
				if err := os.MkdirAll(nestedDestDir, 0755); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to create nested dir %s: %v", nestedDestDir, err))
					continue
				}

				// Recursively copy nested agents
				if err := copyAgentFilesRecursive(nestedSubagentsDir, nestedDestDir, "", result); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("error copying nested agents from %s: %v", srcPath, err))
				}
			}
		} else if strings.HasPrefix(entry.Name(), "agent-") && strings.HasSuffix(entry.Name(), ".jsonl") {
			// This is an agent JSONL file
			destPath := filepath.Join(destDir, entry.Name())
			if parentPath != "" {
				destPath = filepath.Join(destDir, parentPath, entry.Name())
			}

			if err := copyFile(srcPath, destPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to copy %s: %v", entry.Name(), err))
				continue
			}

			// Extract agent ID from filename: agent-{id}.jsonl -> {id}
			agentID := strings.TrimPrefix(strings.TrimSuffix(entry.Name(), ".jsonl"), "agent-")
			result.AgentFiles[agentID] = destPath
			result.TotalAgents++
		}
	}

	return nil
}

// GetExportTreeInfo builds tree info for an export, useful for manifest generation.
// This is a convenience wrapper around agent.BuildNestedTree.
func GetExportTreeInfo(projectDir, sessionID string) (*agent.TreeNode, error) {
	return agent.BuildNestedTree(projectDir, sessionID)
}

// CleanupExport removes an export directory.
// Only removes directories under the claude-history temp directory for safety.
func CleanupExport(exportDir string) error {
	// Safety check: only allow cleanup of claude-history directories
	tempBase := filepath.Join(os.TempDir(), "claude-history")

	// Normalize paths for comparison
	absExport, err := filepath.Abs(exportDir)
	if err != nil {
		return fmt.Errorf("failed to resolve export path: %w", err)
	}

	absTempBase, err := filepath.Abs(tempBase)
	if err != nil {
		return fmt.Errorf("failed to resolve temp base path: %w", err)
	}

	// Check if export dir is under the temp base
	if !strings.HasPrefix(absExport, absTempBase+string(filepath.Separator)) {
		return fmt.Errorf("refusing to cleanup directory outside claude-history temp: %s", exportDir)
	}

	return os.RemoveAll(exportDir)
}
