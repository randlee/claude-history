// Package agent handles agent discovery and hierarchy operations.
package agent

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
)

// AgentMatch represents a matched agent with metadata.
type AgentMatch struct {
	AgentID      string    `json:"agentId"`
	SessionID    string    `json:"sessionId"`
	ProjectPath  string    `json:"projectPath"`
	JSONLPath    string    `json:"jsonlPath"`
	EntryCount   int       `json:"entryCount"`
	MatchedFiles []string  `json:"matchedFiles,omitempty"`
	MatchedTools []string  `json:"matchedTools,omitempty"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"modified"`
}

// FindAgentsOptions specifies criteria for finding agents.
type FindAgentsOptions struct {
	ExploredPattern string    // Glob pattern for files (Read/Write/Edit tool calls)
	ToolTypes       []string  // Filter by tool types used
	ToolMatch       string    // Regex for tool input matching
	StartTime       time.Time // Filter by start time
	EndTime         time.Time // Filter by end time
	SessionID       string    // Scope to single session
}

// fileOperationTools are tools that operate on files.
var fileOperationTools = map[string]bool{
	"read":          true,
	"write":         true,
	"edit":          true,
	"glob":          true,
	"grep":          true,
	"notebookedit":  true,
}

// FindAgents searches for agents matching the given criteria.
// It scans main session JSONL files and nested subagent files.
func FindAgents(projectDir string, opts FindAgentsOptions) ([]AgentMatch, error) {
	// Validate project directory exists
	if !paths.Exists(projectDir) {
		return nil, &PathError{Path: projectDir, Op: "find agents", Err: "directory does not exist"}
	}

	// List all session files in the project
	sessionFiles, err := paths.ListSessionFiles(projectDir)
	if err != nil {
		return nil, err
	}

	// Initialize as empty slice (not nil) to ensure we return [] not null in JSON
	matches := make([]AgentMatch, 0)

	// If a specific session is requested, filter to just that session
	if opts.SessionID != "" {
		if sessionPath, ok := sessionFiles[opts.SessionID]; ok {
			sessionFiles = map[string]string{opts.SessionID: sessionPath}
		} else {
			// Session not found - return empty results
			return matches, nil
		}
	}

	// Compile regex for tool match if provided
	var toolMatchRe *regexp.Regexp
	if opts.ToolMatch != "" {
		var err error
		toolMatchRe, err = regexp.Compile(opts.ToolMatch)
		if err != nil {
			return nil, &PatternError{Pattern: opts.ToolMatch, Err: "invalid regex"}
		}
	}

	// Process each session
	for sessionID, sessionPath := range sessionFiles {
		sessionDir := filepath.Join(projectDir, sessionID)

		// Check main session file as root "agent"
		mainMatch, err := checkAgentFile(sessionPath, "", sessionID, projectDir, opts, toolMatchRe)
		if err == nil && mainMatch != nil {
			matches = append(matches, *mainMatch)
		}

		// Find all nested agents (including nested subagents)
		nestedMatches, err := findNestedAgents(sessionDir, sessionID, projectDir, opts, toolMatchRe, 0)
		if err == nil {
			matches = append(matches, nestedMatches...)
		}
	}

	// Sort by modified time (most recent first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Modified.After(matches[j].Modified)
	})

	return matches, nil
}

// findNestedAgents recursively finds agents in session subagent directories.
// maxDepth limits recursion to prevent infinite loops (0 = unlimited).
func findNestedAgents(sessionDir, sessionID, projectDir string, opts FindAgentsOptions, toolMatchRe *regexp.Regexp, depth int) ([]AgentMatch, error) {
	// Limit recursion depth to prevent stack overflow (3 levels should be sufficient)
	const maxDepth = 3
	if depth >= maxDepth {
		return nil, nil
	}

	agentFiles, err := paths.ListAgentFiles(sessionDir)
	if err != nil {
		return nil, nil // Directory may not exist, that's okay
	}

	var matches []AgentMatch

	for agentID, agentPath := range agentFiles {
		match, err := checkAgentFile(agentPath, agentID, sessionID, projectDir, opts, toolMatchRe)
		if err == nil && match != nil {
			matches = append(matches, *match)
		}

		// Check for nested subagents (agent that spawned agents)
		nestedDir := filepath.Join(sessionDir, "subagents", "agent-"+agentID)
		nestedMatches, _ := findNestedAgents(nestedDir, sessionID, projectDir, opts, toolMatchRe, depth+1)
		matches = append(matches, nestedMatches...)
	}

	return matches, nil
}

// checkAgentFile checks if an agent file matches the search criteria.
// Returns nil, nil if no match found.
func checkAgentFile(filePath, agentID, sessionID, projectDir string, opts FindAgentsOptions, toolMatchRe *regexp.Regexp) (*AgentMatch, error) {
	var entryCount int
	var firstTimestamp, lastTimestamp time.Time
	var matchedFiles []string
	var matchedTools []string

	fileSet := make(map[string]bool)
	toolSet := make(map[string]bool)
	toolMatchFound := false // Track if any tool input matched the regex

	err := jsonl.ScanInto(filePath, func(entry models.ConversationEntry) error {
		entryCount++

		// Extract timestamp
		ts, err := entry.GetTimestamp()
		if err == nil {
			if firstTimestamp.IsZero() || ts.Before(firstTimestamp) {
				firstTimestamp = ts
			}
			if ts.After(lastTimestamp) {
				lastTimestamp = ts
			}
		}

		// Time range filtering - skip entries outside range
		if !opts.StartTime.IsZero() && ts.Before(opts.StartTime) {
			return nil
		}
		if !opts.EndTime.IsZero() && ts.After(opts.EndTime) {
			return nil
		}

		// Extract tool calls from assistant messages
		if entry.Type == models.EntryTypeAssistant {
			tools := entry.ExtractToolCalls()
			for _, tool := range tools {
				toolNameLower := strings.ToLower(tool.Name)
				toolSet[toolNameLower] = true

				// Check for file operations
				if opts.ExploredPattern != "" && isFileOperationTool(toolNameLower) {
					filePath := extractFilePathFromTool(tool)
					if filePath != "" && matchesFilePattern(filePath, opts.ExploredPattern) {
						fileSet[filePath] = true
					}
				}

				// Check tool match regex
				if toolMatchRe != nil && matchesToolInputRegex(tool, toolMatchRe) {
					toolMatchFound = true
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// No entries found
	if entryCount == 0 {
		return nil, nil
	}

	// Convert sets to slices
	for f := range fileSet {
		matchedFiles = append(matchedFiles, f)
	}
	for t := range toolSet {
		matchedTools = append(matchedTools, t)
	}

	// Apply filters
	if !matchesCriteria(opts, matchedFiles, matchedTools, toolMatchRe, toolMatchFound) {
		return nil, nil
	}

	return &AgentMatch{
		AgentID:      agentID,
		SessionID:    sessionID,
		ProjectPath:  projectDir,
		JSONLPath:    filePath,
		EntryCount:   entryCount,
		MatchedFiles: matchedFiles,
		MatchedTools: matchedTools,
		Created:      firstTimestamp,
		Modified:     lastTimestamp,
	}, nil
}

// matchesCriteria checks if the collected data matches the filter options.
func matchesCriteria(opts FindAgentsOptions, matchedFiles, matchedTools []string, toolMatchRe *regexp.Regexp, toolMatchFound bool) bool {
	// If file pattern specified, must have matching files
	if opts.ExploredPattern != "" && len(matchedFiles) == 0 {
		return false
	}

	// If tool types specified, must have at least one matching tool
	if len(opts.ToolTypes) > 0 {
		hasMatch := false
		toolTypesLower := make(map[string]bool)
		for _, t := range opts.ToolTypes {
			toolTypesLower[strings.ToLower(t)] = true
		}
		for _, tool := range matchedTools {
			if toolTypesLower[strings.ToLower(tool)] {
				hasMatch = true
				break
			}
		}
		if !hasMatch {
			return false
		}
	}

	// If tool match regex specified, must have found a matching tool input
	if toolMatchRe != nil && !toolMatchFound {
		return false
	}

	return true
}

// isFileOperationTool checks if a tool name is a file operation.
func isFileOperationTool(toolName string) bool {
	return fileOperationTools[strings.ToLower(toolName)]
}

// extractFilePathFromTool extracts the file path from a tool's input.
func extractFilePathFromTool(tool models.ToolUse) string {
	if tool.Input == nil {
		return ""
	}

	// Try common field names for file paths
	pathFields := []string{"file_path", "path", "filePath", "file", "filename"}
	for _, field := range pathFields {
		if val, ok := tool.Input[field]; ok {
			if path, ok := val.(string); ok {
				return path
			}
		}
	}

	return ""
}

// matchesToolInputRegex checks if any tool input field matches the regex.
func matchesToolInputRegex(tool models.ToolUse, re *regexp.Regexp) bool {
	if tool.Input == nil {
		return false
	}

	// Check each field in the input
	for _, val := range tool.Input {
		if str, ok := val.(string); ok {
			if re.MatchString(str) {
				return true
			}
		}
	}

	return false
}

// matchesFilePattern checks if a path matches a glob pattern.
// Supports standard glob patterns:
// - * matches any sequence of characters except path separators
// - ** matches any sequence of characters including path separators
// - ? matches any single character
func matchesFilePattern(path, pattern string) bool {
	if pattern == "" {
		return false
	}

	// Normalize path separators
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Handle ** (double star) patterns
	if strings.Contains(pattern, "**") {
		return matchDoubleStarPattern(path, pattern)
	}

	// Use filepath.Match for single-star patterns
	matched, err := filepath.Match(pattern, path)
	if err != nil {
		return false
	}
	if matched {
		return true
	}

	// Also try matching just the filename for patterns without path separators
	if !strings.Contains(pattern, "/") {
		matched, _ = filepath.Match(pattern, filepath.Base(path))
		return matched
	}

	return false
}

// matchDoubleStarPattern handles ** glob patterns.
func matchDoubleStarPattern(path, pattern string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 1 {
		// No **, use regular matching
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// For patterns like "**/*.go", the path should end with the suffix pattern
	if parts[0] == "" && len(parts) == 2 {
		// Pattern starts with **, e.g., "**/*.go"
		suffix := parts[1]
		if strings.HasPrefix(suffix, "/") {
			suffix = suffix[1:]
		}

		// Try to match suffix against any part of the path
		pathParts := strings.Split(path, "/")
		for i := 0; i < len(pathParts); i++ {
			subPath := strings.Join(pathParts[i:], "/")
			matched, _ := filepath.Match(suffix, subPath)
			if matched {
				return true
			}
		}

		// Also try just matching the filename
		if !strings.Contains(suffix, "/") {
			matched, _ := filepath.Match(suffix, filepath.Base(path))
			return matched
		}
	}

	// For patterns like "src/**/file.go"
	if len(parts) >= 2 {
		prefix := parts[0]
		suffix := parts[len(parts)-1]

		// Path must start with prefix (if non-empty)
		if prefix != "" && !strings.HasPrefix(path, prefix) {
			return false
		}

		// Path must end with suffix (if non-empty)
		if suffix != "" {
			if strings.HasPrefix(suffix, "/") {
				suffix = suffix[1:]
			}
			// Try matching suffix
			matched, _ := filepath.Match(suffix, path)
			if matched {
				return true
			}
			// Also try matching just the filename
			if !strings.Contains(suffix, "/") {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}
		}

		return prefix != "" || suffix == ""
	}

	return false
}

// PathError represents an error related to a path operation.
type PathError struct {
	Path string
	Op   string
	Err  string
}

func (e *PathError) Error() string {
	return e.Op + " " + e.Path + ": " + e.Err
}

// PatternError represents an error related to a pattern.
type PatternError struct {
	Pattern string
	Err     string
}

func (e *PatternError) Error() string {
	return "pattern " + e.Pattern + ": " + e.Err
}
