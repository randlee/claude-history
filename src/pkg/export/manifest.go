// Package export handles HTML export of Claude Code sessions.
package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
)

// ManifestVersion is the current version of the manifest format.
const ManifestVersion = "1.0.0"

// Manifest contains metadata about an exported session.
type Manifest struct {
	Version     string          `json:"version"`
	ExportedAt  time.Time       `json:"exported_at"`
	SessionID   string          `json:"session_id"`
	ProjectPath string          `json:"project_path"`
	EntryCount  int             `json:"entry_count"`
	AgentTree   *AgentTreeNode  `json:"agent_tree"`
	SourceFiles []SourceFile    `json:"source_files"`
}

// AgentTreeNode represents a node in the agent hierarchy for the manifest.
type AgentTreeNode struct {
	ID       string           `json:"id"`
	Entries  int              `json:"entries"`
	Children []*AgentTreeNode `json:"children,omitempty"`
}

// SourceFile describes a source JSONL file included in the export.
type SourceFile struct {
	Type    string `json:"type"` // "session" or "agent"
	AgentID string `json:"agent_id,omitempty"`
	Path    string `json:"path"`
}

// GenerateManifest creates a manifest for a session export.
// projectDir is the full path to the Claude project directory.
// sessionID is the session identifier.
// outputDir is the directory where the export will be written.
func GenerateManifest(projectDir, sessionID, outputDir string) (*Manifest, error) {
	sessionPath := filepath.Join(projectDir, sessionID+".jsonl")

	// Verify session file exists
	if !paths.Exists(sessionPath) {
		return nil, os.ErrNotExist
	}

	// Build the agent tree
	tree, err := agent.BuildTree(projectDir, sessionID)
	if err != nil {
		return nil, err
	}

	// Convert agent tree to manifest format
	agentTree := convertTreeNode(tree)

	// Calculate total entry count
	totalEntries := agent.CountTotalEntries(tree)

	// Build source files list
	sourceFiles := buildSourceFilesList(tree, projectDir, sessionID)

	// Determine project path (decode from directory name)
	projectPath := extractProjectPath(projectDir)

	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   sessionID,
		ProjectPath: projectPath,
		EntryCount:  totalEntries,
		AgentTree:   agentTree,
		SourceFiles: sourceFiles,
	}

	return manifest, nil
}

// WriteManifest writes a manifest to the output directory as manifest.json.
func WriteManifest(manifest *Manifest, outputDir string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	manifestPath := filepath.Join(outputDir, "manifest.json")

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

// ReadManifest reads a manifest from an export directory.
func ReadManifest(outputDir string) (*Manifest, error) {
	manifestPath := filepath.Join(outputDir, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// convertTreeNode converts an agent.TreeNode to an AgentTreeNode.
func convertTreeNode(node *agent.TreeNode) *AgentTreeNode {
	if node == nil {
		return nil
	}

	result := &AgentTreeNode{
		ID:      node.AgentID,
		Entries: node.EntryCount,
	}

	// For root node, use session ID as the ID
	if node.IsRoot {
		result.ID = node.SessionID
	}

	// Convert children
	if len(node.Children) > 0 {
		result.Children = make([]*AgentTreeNode, len(node.Children))
		for i, child := range node.Children {
			result.Children[i] = convertTreeNode(child)
		}
	}

	return result
}

// buildSourceFilesList creates the list of source files from the tree.
func buildSourceFilesList(tree *agent.TreeNode, projectDir, sessionID string) []SourceFile {
	var files []SourceFile

	// Add main session file
	sessionPath := filepath.Join(projectDir, sessionID+".jsonl")
	if paths.Exists(sessionPath) {
		files = append(files, SourceFile{
			Type: "session",
			Path: sessionPath,
		})
	}

	// Add agent files recursively
	addAgentFiles(tree, &files)

	return files
}

// addAgentFiles recursively adds agent files to the list.
func addAgentFiles(node *agent.TreeNode, files *[]SourceFile) {
	if node == nil {
		return
	}

	// Add this node's file if it's an agent (not root)
	if !node.IsRoot && node.AgentID != "" && node.FilePath != "" {
		*files = append(*files, SourceFile{
			Type:    "agent",
			AgentID: node.AgentID,
			Path:    node.FilePath,
		})
	}

	// Process children
	for _, child := range node.Children {
		addAgentFiles(child, files)
	}
}

// extractProjectPath attempts to decode the project path from the directory name.
func extractProjectPath(projectDir string) string {
	// The project directory name is the encoded path
	dirName := filepath.Base(projectDir)

	// Try to decode it using the encoding package
	// For now, just return the directory name as-is
	// The actual decoding would require importing the encoding package
	return dirName
}

// CountSessionEntries counts entries in a session file.
func CountSessionEntries(filePath string) (int, error) {
	return jsonl.CountLines(filePath)
}

// GetAgentEntryCount returns the entry count for an agent.
func GetAgentEntryCount(filePath string) (int, error) {
	count := 0
	err := jsonl.ScanInto(filePath, func(entry models.ConversationEntry) error {
		count++
		return nil
	})
	return count, err
}
