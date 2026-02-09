// Package paths handles Claude Code directory resolution and path operations.
package paths

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/randlee/claude-history/pkg/encoding"
)

// DefaultClaudeDir returns the default Claude configuration directory.
// This is ~/.claude on all platforms.
func DefaultClaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// ProjectsDir returns the path to the projects directory within Claude's config.
// If claudeDir is empty, uses the default ~/.claude location.
func ProjectsDir(claudeDir string) (string, error) {
	if claudeDir == "" {
		var err error
		claudeDir, err = DefaultClaudeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(claudeDir, "projects"), nil
}

// ProjectDir returns the path to a specific project's directory.
// The projectPath can be relative or absolute - it will be resolved to an absolute path.
func ProjectDir(claudeDir string, projectPath string) (string, error) {
	projectsDir, err := ProjectsDir(claudeDir)
	if err != nil {
		return "", err
	}

	// Resolve relative paths to absolute paths
	// Note: Paths starting with "/" are considered absolute for encoding purposes,
	// even on Windows where filepath.IsAbs() would return false.
	// This ensures cross-platform test compatibility where "/test/project"
	// consistently encodes to "-test-project" on all platforms.
	absPath := projectPath
	if !filepath.IsAbs(projectPath) && !strings.HasPrefix(projectPath, "/") {
		absPath, err = filepath.Abs(projectPath)
		if err != nil {
			return "", err
		}
	}

	encoded := encoding.EncodePath(absPath)
	return filepath.Join(projectsDir, encoded), nil
}

// SessionFile returns the path to a session's JSONL file.
func SessionFile(claudeDir string, projectPath string, sessionID string) (string, error) {
	projectDir, err := ProjectDir(claudeDir, projectPath)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, sessionID+".jsonl"), nil
}

// AgentFile returns the path to an agent's JSONL file within a session.
func AgentFile(claudeDir string, projectPath string, sessionID string, agentID string) (string, error) {
	projectDir, err := ProjectDir(claudeDir, projectPath)
	if err != nil {
		return "", err
	}

	// Agent files are in: {projectDir}/{sessionId}/subagents/agent-{agentId}.jsonl
	return filepath.Join(projectDir, sessionID, "subagents", "agent-"+agentID+".jsonl"), nil
}

// SubagentsDir returns the path to the subagents directory for a session.
func SubagentsDir(claudeDir string, projectPath string, sessionID string) (string, error) {
	projectDir, err := ProjectDir(claudeDir, projectPath)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, sessionID, "subagents"), nil
}

// SessionIndexFile returns the path to the sessions-index.json file for a project.
func SessionIndexFile(claudeDir string, projectPath string) (string, error) {
	projectDir, err := ProjectDir(claudeDir, projectPath)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, "sessions-index.json"), nil
}

// ListProjects returns all project directories in the Claude projects folder.
// Returns a map of encoded project name to full path.
func ListProjects(claudeDir string) (map[string]string, error) {
	projectsDir, err := ProjectsDir(claudeDir)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// Skip non-encoded directories (shouldn't normally exist)
			if encoding.IsEncodedPath(name) {
				result[name] = filepath.Join(projectsDir, name)
			}
		}
	}

	return result, nil
}

// ListSessionFiles returns all session JSONL files in a project directory.
// Returns a map of session ID to full file path.
func ListSessionFiles(projectDir string) (map[string]string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	result := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".jsonl") {
			sessionID := strings.TrimSuffix(name, ".jsonl")
			// Session IDs are UUIDs, not encoded paths
			if !encoding.IsEncodedPath(sessionID) && looksLikeUUID(sessionID) {
				result[sessionID] = filepath.Join(projectDir, name)
			}
		}
	}

	return result, nil
}

// ListAgentFiles returns all agent JSONL files in a session's subagents directory.
// Returns a map of agent ID to full file path.
// This function recursively discovers nested agents in subdirectories.
func ListAgentFiles(sessionDir string) (map[string]string, error) {
	subagentsDir := filepath.Join(sessionDir, "subagents")

	// Check if subagents directory exists
	if _, err := os.Stat(subagentsDir); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	result := make(map[string]string)
	err := listAgentFilesRecursive(subagentsDir, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// listAgentFilesRecursive recursively scans for agent JSONL files.
func listAgentFilesRecursive(dir string, result map[string]string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		if entry.IsDir() {
			// Check if this directory has a nested subagents directory
			nestedSubagentsDir := filepath.Join(fullPath, "subagents")
			if _, err := os.Stat(nestedSubagentsDir); err == nil {
				// Recursively scan the nested subagents directory
				if err := listAgentFilesRecursive(nestedSubagentsDir, result); err != nil {
					return err
				}
			}
			continue
		}

		// Process agent JSONL files
		if strings.HasPrefix(name, "agent-") && strings.HasSuffix(name, ".jsonl") {
			agentID := strings.TrimPrefix(strings.TrimSuffix(name, ".jsonl"), "agent-")
			result[agentID] = fullPath
		}
	}

	return nil
}

// looksLikeUUID checks if a string looks like a UUID (has dashes in typical positions).
func looksLikeUUID(s string) bool {
	// UUIDs are 36 chars: 8-4-4-4-12
	if len(s) != 36 {
		return false
	}
	return s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}

// Exists checks if a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
