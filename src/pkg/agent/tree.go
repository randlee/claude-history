package agent

import (
	"path/filepath"

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
}

// BuildTree constructs an agent hierarchy tree for a session.
// The root node represents the main session, with child nodes for each spawned agent.
func BuildTree(projectDir string, sessionID string) (*TreeNode, error) {
	sessionPath := filepath.Join(projectDir, sessionID+".jsonl")
	sessionDir := filepath.Join(projectDir, sessionID)

	// Create root node for the main session
	root := &TreeNode{
		SessionID: sessionID,
		FilePath:  sessionPath,
		IsRoot:    true,
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

	// Find which queue-operations spawned which agents
	spawns, _ := FindAgentSpawns(sessionPath)

	// Build child nodes for each agent
	// For now, we create a flat tree (all agents are direct children of root)
	// A more sophisticated version would build a proper hierarchy based on parentUuid chains
	for _, agent := range agents {
		child := &TreeNode{
			AgentID:    agent.ID,
			SessionID:  agent.SessionID,
			FilePath:   agent.FilePath,
			EntryCount: agent.EntryCount,
			AgentType:  agent.AgentType,
		}

		// Record the spawn relationship
		if spawnUUID, ok := spawns[agent.ID]; ok {
			_ = spawnUUID // Could be used for more detailed tree building
		}

		root.Children = append(root.Children, child)
	}

	return root, nil
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
