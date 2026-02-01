// Package agent handles agent discovery and hierarchy operations.
package agent

import (
	"path/filepath"
	"strings"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
)

// DiscoverAgents finds all agent files for a session.
func DiscoverAgents(sessionDir string) ([]models.Agent, error) {
	agentFiles, err := paths.ListAgentFiles(sessionDir)
	if err != nil {
		return nil, err
	}

	var agents []models.Agent
	for agentID, filePath := range agentFiles {
		agent := models.Agent{
			ID:       agentID,
			FilePath: filePath,
		}

		// Determine agent type from filename
		agent.AgentType = parseAgentType(agentID)

		// Count entries in the agent file
		count, err := jsonl.CountLines(filePath)
		if err == nil {
			agent.EntryCount = count
		}

		// Try to get session ID from first entry
		_ = jsonl.ScanInto(filePath, func(entry models.ConversationEntry) error {
			agent.SessionID = entry.SessionID
			return StopIteration // Stop after first entry
		})

		agents = append(agents, agent)
	}

	return agents, nil
}

// StopIteration is a sentinel error to stop scanning early.
var StopIteration = &stopIterationError{}

type stopIterationError struct{}

func (e *stopIterationError) Error() string { return "stop iteration" }

// parseAgentType extracts the agent type from an agent ID.
// Examples:
//
//	"a12eb64" -> ""
//	"aprompt_suggestion-abc123" -> "prompt_suggestion"
func parseAgentType(agentID string) string {
	// Agent IDs can have prefixes like "aprompt_suggestion-"
	if strings.HasPrefix(agentID, "aprompt_suggestion-") {
		return "prompt_suggestion"
	}
	if strings.HasPrefix(agentID, "aexplore-") {
		return "explore"
	}
	// Default agents (just hex IDs like "a12eb64") have no specific type
	return ""
}

// FindAgentSpawns scans a session file to find queue-operation entries that spawn agents.
// Returns a map of agent ID to the UUID of the entry that spawned it.
func FindAgentSpawns(sessionFilePath string) (map[string]string, error) {
	spawns := make(map[string]string)

	err := jsonl.ScanInto(sessionFilePath, func(entry models.ConversationEntry) error {
		if entry.Type == models.EntryTypeQueueOperation {
			// Queue operations that spawn agents have an agentId
			if entry.AgentID != "" {
				spawns[entry.AgentID] = entry.UUID
			}
		}
		return nil
	})

	return spawns, err
}

// GetAgent finds a specific agent by ID in a session directory.
func GetAgent(sessionDir string, agentID string) (*models.Agent, error) {
	// Try both formats: with and without "agent-" prefix
	subagentsDir := filepath.Join(sessionDir, "subagents")

	var filePath string
	testPaths := []string{
		filepath.Join(subagentsDir, "agent-"+agentID+".jsonl"),
		filepath.Join(subagentsDir, agentID+".jsonl"),
	}

	for _, p := range testPaths {
		if paths.Exists(p) {
			filePath = p
			break
		}
	}

	if filePath == "" {
		return nil, nil
	}

	agent := &models.Agent{
		ID:        agentID,
		FilePath:  filePath,
		AgentType: parseAgentType(agentID),
	}

	count, err := jsonl.CountLines(filePath)
	if err == nil {
		agent.EntryCount = count
	}

	return agent, nil
}

// ReadAgentEntries reads all entries from an agent's JSONL file.
func ReadAgentEntries(filePath string) ([]models.ConversationEntry, error) {
	return jsonl.ReadAll[models.ConversationEntry](filePath)
}
