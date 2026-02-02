package agent

import (
	"path/filepath"

	"github.com/randlee/claude-history/internal/jsonl"
	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/paths"
)

// TreeNode represents a node in the agent hierarchy tree.
type TreeNode struct {
	AgentID    string      `json:"agentId,omitempty"`
	SessionID  string      `json:"sessionId"`
	FilePath   string      `json:"filePath"`
	EntryCount int         `json:"entryCount"`
	AgentType  string      `json:"agentType,omitempty"`
	IsRoot     bool        `json:"isRoot"`
	Children   []*TreeNode `json:"children,omitempty"`
	ParentUUID string      `json:"parentUuid,omitempty"` // UUID of parent agent or main session
	UUID       string      `json:"uuid,omitempty"`       // UUID of the entry that spawned this agent
}

// SpawnInfo contains information about agent spawn relationships.
type SpawnInfo struct {
	AgentID    string // The ID of the spawned agent
	SpawnUUID  string // UUID of the user entry that contains the spawn result
	ParentUUID string // UUID of the assistant message that triggered the spawn (sourceToolAssistantUUID)
}

// BuildTree constructs an agent hierarchy tree for a session.
// The root node represents the main session, with child nodes for each spawned agent.
// This is a backward-compatible wrapper that calls BuildNestedTree.
func BuildTree(projectDir string, sessionID string) (*TreeNode, error) {
	return BuildNestedTree(projectDir, sessionID)
}

// BuildNestedTree constructs a properly nested agent hierarchy tree for a session.
// It uses toolUseResult from user entries to detect agent spawns and build parent-child relationships.
func BuildNestedTree(projectDir string, sessionID string) (*TreeNode, error) {
	sessionPath := filepath.Join(projectDir, sessionID+".jsonl")
	sessionDir := filepath.Join(projectDir, sessionID)

	// Create root node for the main session
	root := &TreeNode{
		SessionID: sessionID,
		FilePath:  sessionPath,
		IsRoot:    true,
		UUID:      sessionID, // Use session ID as the root's UUID for parent matching
	}

	// Count entries in main session
	if paths.Exists(sessionPath) {
		entries, err := countSessionEntries(sessionPath)
		if err == nil {
			root.EntryCount = entries
		}
	}

	// Find all agents
	agents, err := DiscoverAgents(sessionDir)
	if err != nil {
		// Session may not have any agents, that's OK
		return root, nil
	}

	if len(agents) == 0 {
		return root, nil
	}

	// Build spawn info map from toolUseResult entries in main session and agent files
	spawnInfoMap := buildSpawnInfoMap(sessionPath, sessionDir, agents)

	// Create nodes for all agents
	nodeMap := make(map[string]*TreeNode)
	nodeMap[sessionID] = root // Add root to map for parent lookups
	nodeMap[""] = root        // Empty parent also maps to root

	for _, agent := range agents {
		node := &TreeNode{
			AgentID:    agent.ID,
			SessionID:  agent.SessionID,
			FilePath:   agent.FilePath,
			EntryCount: agent.EntryCount,
			AgentType:  agent.AgentType,
		}

		// Get spawn info for this agent
		if info, ok := spawnInfoMap[agent.ID]; ok {
			node.UUID = info.SpawnUUID
			node.ParentUUID = info.ParentUUID
		}

		nodeMap[agent.ID] = node
	}

	// Track visited nodes for cycle detection
	visited := make(map[string]bool)

	// Build the tree structure by connecting children to parents
	for _, agent := range agents {
		node := nodeMap[agent.ID]
		parent := findParentNode(nodeMap, node.ParentUUID, agent.ID, visited)
		if parent == nil {
			// Orphaned agent - attach to root
			parent = root
		}
		parent.Children = append(parent.Children, node)
	}

	return root, nil
}

// buildSpawnInfoMap extracts spawn information from session and agent files.
// It looks for user entries with toolUseResult where status is "async_launched".
func buildSpawnInfoMap(sessionPath string, sessionDir string, agents []models.Agent) map[string]*SpawnInfo {
	result := make(map[string]*SpawnInfo)

	// Scan main session for agent spawns (user entries with toolUseResult)
	_ = jsonl.ScanInto(sessionPath, func(entry models.ConversationEntry) error {
		if entry.IsAgentSpawn() {
			agentID := entry.GetSpawnedAgentID()
			result[agentID] = &SpawnInfo{
				AgentID:    agentID,
				SpawnUUID:  entry.UUID,
				ParentUUID: entry.SourceToolAssistantUUID,
			}
		}
		return nil
	})

	// Scan each agent file for nested agent spawns
	for _, agent := range agents {
		_ = jsonl.ScanInto(agent.FilePath, func(entry models.ConversationEntry) error {
			if entry.IsAgentSpawn() {
				agentID := entry.GetSpawnedAgentID()
				parentUUID := entry.SourceToolAssistantUUID
				// For nested agents, if parentUUID is empty, the parent is this agent
				if parentUUID == "" {
					parentUUID = agent.ID
				}
				result[agentID] = &SpawnInfo{
					AgentID:    agentID,
					SpawnUUID:  entry.UUID,
					ParentUUID: parentUUID,
				}
			}
			return nil
		})
	}

	return result
}

// findParentNode resolves a sourceToolAssistantUUID to find the parent node.
// It looks up nodes by agent ID or by their UUID field (assistant message UUID).
// Handles circular references by tracking visited nodes.
// Returns nil if parent not found or cycle detected.
func findParentNode(nodeMap map[string]*TreeNode, parentUUID string, currentID string, visited map[string]bool) *TreeNode {
	// Self-referencing check
	if parentUUID == currentID {
		return nil
	}

	// Direct lookup by UUID or agent ID
	if node, ok := nodeMap[parentUUID]; ok {
		// Check for circular reference
		if visited[parentUUID] {
			return nil
		}
		return node
	}

	// Try to find by matching the node's UUID field
	for id, node := range nodeMap {
		if node.UUID == parentUUID {
			// Check for circular reference
			if visited[id] {
				return nil
			}
			return node
		}
	}

	return nil
}

// FindParentAgent resolves parentUuid to find the parent agent.
// This is a public helper for external use.
// Returns nil if parent not found or UUID is empty.
func FindParentAgent(agents map[string]*TreeNode, parentUUID string) *TreeNode {
	if parentUUID == "" {
		return nil
	}

	// Direct lookup by agent ID
	if node, ok := agents[parentUUID]; ok {
		return node
	}

	// Try to find by matching the node's UUID field
	for _, node := range agents {
		if node.UUID == parentUUID {
			return node
		}
	}

	return nil
}

// countSessionEntries counts entries in a session file.
func countSessionEntries(filePath string) (int, error) {
	entries, err := ReadAgentEntries(filePath)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}

// FlattenTree returns all nodes in the tree as a flat slice.
func FlattenTree(root *TreeNode) []*TreeNode {
	var nodes []*TreeNode
	flattenTreeRecursive(root, &nodes)
	return nodes
}

func flattenTreeRecursive(node *TreeNode, nodes *[]*TreeNode) {
	if node == nil {
		return
	}
	*nodes = append(*nodes, node)
	for _, child := range node.Children {
		flattenTreeRecursive(child, nodes)
	}
}

// CountTotalEntries returns the total number of entries across all nodes.
func CountTotalEntries(root *TreeNode) int {
	total := 0
	nodes := FlattenTree(root)
	for _, node := range nodes {
		total += node.EntryCount
	}
	return total
}
