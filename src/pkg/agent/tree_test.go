package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// createToolUseResultEntry creates a JSONL entry with toolUseResult for testing.
// This is the canonical format for agent spawn detection.
func createToolUseResultEntry(uuid, sessionID, agentID, sourceToolAssistantUUID, status string) string {
	entry := map[string]any{
		"uuid":      uuid,
		"sessionId": sessionID,
		"type":      "user",
		"timestamp": "2026-01-15T10:00:00Z",
		"toolUseResult": map[string]any{
			"isAsync":     true,
			"status":      status, // "async_launched" for spawns
			"agentId":     agentID,
			"description": "Test agent spawn",
		},
	}
	if sourceToolAssistantUUID != "" {
		entry["sourceToolAssistantUUID"] = sourceToolAssistantUUID
	}
	data, _ := json.Marshal(entry)
	return string(data) + "\n"
}

// createNonSpawnToolUseResult creates a toolUseResult that is NOT an agent spawn
func createNonSpawnToolUseResult(uuid, sessionID, status string) string {
	entry := map[string]any{
		"uuid":      uuid,
		"sessionId": sessionID,
		"type":      "user",
		"timestamp": "2026-01-15T10:00:00Z",
		"toolUseResult": map[string]any{
			"isAsync": true,
			"status":  status, // Not "async_launched"
		},
	}
	data, _ := json.Marshal(entry)
	return string(data) + "\n"
}

func TestBuildSpawnInfoMap_ToolUseResultFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "test-session-123"

	// Create main session file with toolUseResult-based spawns
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"user-1","type":"user"}` + "\n"
	sessionContent += `{"uuid":"assistant-1","type":"assistant"}` + "\n"
	sessionContent += createToolUseResultEntry("spawn-entry-1", sessionID, "agent-alpha", "assistant-1", "async_launched")
	sessionContent += `{"uuid":"assistant-2","type":"assistant"}` + "\n"
	sessionContent += createToolUseResultEntry("spawn-entry-2", sessionID, "agent-beta", "assistant-2", "async_launched")

	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create session dir but no agents (we're just testing spawn detection)
	sessionDir := filepath.Join(tmpDir, sessionID)
	mustMkdirAll(t, sessionDir)

	// Manually call buildSpawnInfoMap
	spawnMap := buildSpawnInfoMap(sessionFile, sessionDir, nil)

	if len(spawnMap) != 2 {
		t.Fatalf("buildSpawnInfoMap() found %d spawns, want 2", len(spawnMap))
	}

	// Verify agent-alpha spawn info
	if info, ok := spawnMap["agent-alpha"]; !ok {
		t.Error("agent-alpha not found in spawn map")
	} else {
		if info.AgentID != "agent-alpha" {
			t.Errorf("agent-alpha AgentID = %q, want 'agent-alpha'", info.AgentID)
		}
		if info.SpawnUUID != "spawn-entry-1" {
			t.Errorf("agent-alpha SpawnUUID = %q, want 'spawn-entry-1'", info.SpawnUUID)
		}
		if info.ParentUUID != "assistant-1" {
			t.Errorf("agent-alpha ParentUUID = %q, want 'assistant-1'", info.ParentUUID)
		}
	}

	// Verify agent-beta spawn info
	if info, ok := spawnMap["agent-beta"]; !ok {
		t.Error("agent-beta not found in spawn map")
	} else {
		if info.ParentUUID != "assistant-2" {
			t.Errorf("agent-beta ParentUUID = %q, want 'assistant-2'", info.ParentUUID)
		}
	}
}

func TestBuildSpawnInfoMap_IgnoresNonAsyncLaunched(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "test-session-123"

	// Create session with various toolUseResult statuses
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"user-1","type":"user"}` + "\n"
	sessionContent += `{"uuid":"assistant-1","type":"assistant"}` + "\n"
	// This should be detected (async_launched)
	sessionContent += createToolUseResultEntry("spawn-1", sessionID, "agent-valid", "assistant-1", "async_launched")
	// This should NOT be detected (completed status)
	sessionContent += createNonSpawnToolUseResult("not-spawn-1", sessionID, "completed")
	// This should NOT be detected (running status)
	sessionContent += createNonSpawnToolUseResult("not-spawn-2", sessionID, "running")

	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	mustMkdirAll(t, sessionDir)

	spawnMap := buildSpawnInfoMap(sessionFile, sessionDir, nil)

	if len(spawnMap) != 1 {
		t.Errorf("buildSpawnInfoMap() found %d spawns, want 1", len(spawnMap))
	}

	if _, ok := spawnMap["agent-valid"]; !ok {
		t.Error("agent-valid not found in spawn map")
	}
}

func TestBuildSpawnInfoMap_NestedAgentSpawns(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "test-session-123"

	// Create main session - spawns agent-1
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}` + "\n"
	sessionContent += `{"uuid":"main-2","type":"assistant"}` + "\n"
	sessionContent += createToolUseResultEntry("spawn-agent1", sessionID, "agent-1", "main-2", "async_launched")

	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent-1 file which spawns agent-2
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agent1Content := `{"uuid":"a1-user","type":"user"}` + "\n"
	agent1Content += `{"uuid":"a1-assistant","type":"assistant"}` + "\n"
	agent1Content += createToolUseResultEntry("spawn-agent2", sessionID, "agent-2", "agent-1", "async_launched")

	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-1.jsonl"), []byte(agent1Content))

	// Create agent-2 file
	agent2Content := `{"uuid":"a2-user","type":"user"}` + "\n"
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-2.jsonl"), []byte(agent2Content))

	// Discover agents first
	agents, err := DiscoverAgents(sessionDir)
	if err != nil {
		t.Fatalf("DiscoverAgents() error: %v", err)
	}

	// Build spawn map
	spawnMap := buildSpawnInfoMap(sessionFile, sessionDir, agents)

	if len(spawnMap) != 2 {
		t.Fatalf("buildSpawnInfoMap() found %d spawns, want 2", len(spawnMap))
	}

	// Verify agent-1 was spawned from main session
	if info, ok := spawnMap["agent-1"]; !ok {
		t.Error("agent-1 not found in spawn map")
	} else {
		if info.ParentUUID != "main-2" {
			t.Errorf("agent-1 ParentUUID = %q, want 'main-2'", info.ParentUUID)
		}
	}

	// Verify agent-2 was spawned from agent-1
	if info, ok := spawnMap["agent-2"]; !ok {
		t.Error("agent-2 not found in spawn map")
	} else {
		if info.ParentUUID != "agent-1" {
			t.Errorf("agent-2 ParentUUID = %q, want 'agent-1'", info.ParentUUID)
		}
	}
}

func TestBuildSpawnInfoMap_FallbackParentUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "test-session-123"

	// Create main session - spawns agent-1
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}` + "\n"
	sessionContent += `{"uuid":"main-2","type":"assistant"}` + "\n"
	sessionContent += createToolUseResultEntry("spawn-agent1", sessionID, "agent-1", "main-2", "async_launched")

	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent-1 file which spawns agent-2 WITHOUT sourceToolAssistantUUID
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Spawn entry without sourceToolAssistantUUID - should fall back to agent ID
	agent1Content := `{"uuid":"a1-user","type":"user"}` + "\n"
	agent1Content += `{"uuid":"a1-assistant","type":"assistant"}` + "\n"
	agent1Content += createToolUseResultEntry("spawn-agent2", sessionID, "agent-2", "", "async_launched")

	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-1.jsonl"), []byte(agent1Content))

	// Create agent-2 file
	agent2Content := `{"uuid":"a2-user","type":"user"}` + "\n"
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-2.jsonl"), []byte(agent2Content))

	// Discover agents and build spawn map
	agents, _ := DiscoverAgents(sessionDir)
	spawnMap := buildSpawnInfoMap(sessionFile, sessionDir, agents)

	// Verify agent-2's parent falls back to agent-1's ID
	if info, ok := spawnMap["agent-2"]; !ok {
		t.Error("agent-2 not found in spawn map")
	} else {
		if info.ParentUUID != "agent-1" {
			t.Errorf("agent-2 ParentUUID = %q, want 'agent-1' (fallback)", info.ParentUUID)
		}
	}
}

func TestTreeBuilding_MixedScenario(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "mixed-session"

	// Main session:
	// - Spawns agent-with-info (has spawn entry)
	// - Doesn't spawn agent-orphan (no spawn entry in main session)
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}` + "\n"
	sessionContent += `{"uuid":"main-2","type":"assistant"}` + "\n"
	sessionContent += createToolUseResultEntry("spawn-with-info", sessionID, "agent-with-info", "main-2", "async_launched")

	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Agent with spawn info
	agentWithInfoContent := `{"uuid":"awi-1","type":"user"}
{"uuid":"awi-2","type":"assistant"}
`
	// Orphan agent (exists on disk but no spawn entry)
	orphanContent := `{"uuid":"orph-1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-with-info.jsonl"), []byte(agentWithInfoContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-orphan.jsonl"), []byte(orphanContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Both agents should be in the tree
	if len(tree.Children) != 2 {
		t.Errorf("Root has %d children, want 2", len(tree.Children))
	}

	// Verify both are present
	childIDs := make(map[string]bool)
	for _, child := range tree.Children {
		childIDs[child.AgentID] = true
	}
	if !childIDs["agent-with-info"] {
		t.Error("agent-with-info not found in tree")
	}
	if !childIDs["agent-orphan"] {
		t.Error("agent-orphan not found in tree (should attach to root)")
	}
}

func TestFindParentNode_ByAgentID(t *testing.T) {
	nodeMap := map[string]*TreeNode{
		"session-root": {SessionID: "session-root", UUID: "session-root", IsRoot: true},
		"agent-alpha":  {AgentID: "agent-alpha", UUID: "spawn-uuid-alpha"},
		"agent-beta":   {AgentID: "agent-beta", UUID: "spawn-uuid-beta"},
	}
	visited := make(map[string]bool)

	// Find by agent ID
	result := findParentNode(nodeMap, "agent-alpha", "some-other-agent", visited)
	if result == nil {
		t.Fatal("findParentNode() returned nil, want agent-alpha node")
	}
	if result.AgentID != "agent-alpha" {
		t.Errorf("findParentNode() AgentID = %q, want 'agent-alpha'", result.AgentID)
	}
}

func TestFindParentNode_ByUUID(t *testing.T) {
	nodeMap := map[string]*TreeNode{
		"session-root": {SessionID: "session-root", UUID: "session-root", IsRoot: true},
		"agent-alpha":  {AgentID: "agent-alpha", UUID: "spawn-uuid-alpha"},
	}
	visited := make(map[string]bool)

	// Find by UUID field
	result := findParentNode(nodeMap, "spawn-uuid-alpha", "some-other-agent", visited)
	if result == nil {
		t.Fatal("findParentNode() by UUID returned nil, want agent-alpha node")
	}
	if result.AgentID != "agent-alpha" {
		t.Errorf("findParentNode() AgentID = %q, want 'agent-alpha'", result.AgentID)
	}
}

func TestFindParentNode_SelfReference(t *testing.T) {
	nodeMap := map[string]*TreeNode{
		"agent-self": {AgentID: "agent-self", UUID: "uuid-self"},
	}
	visited := make(map[string]bool)

	// Self-reference should return nil
	result := findParentNode(nodeMap, "agent-self", "agent-self", visited)
	if result != nil {
		t.Errorf("findParentNode() with self-reference should return nil, got %+v", result)
	}
}

func TestFindParentNode_NotFound(t *testing.T) {
	nodeMap := map[string]*TreeNode{
		"agent-alpha": {AgentID: "agent-alpha", UUID: "uuid-alpha"},
	}
	visited := make(map[string]bool)

	// Non-existent parent
	result := findParentNode(nodeMap, "nonexistent-parent", "agent-alpha", visited)
	if result != nil {
		t.Errorf("findParentNode() for nonexistent parent should return nil, got %+v", result)
	}
}

func TestFlattenTree(t *testing.T) {
	root := &TreeNode{
		SessionID:  "root",
		IsRoot:     true,
		EntryCount: 5,
		Children: []*TreeNode{
			{
				AgentID:    "child-1",
				EntryCount: 3,
				Children: []*TreeNode{
					{AgentID: "grandchild-1", EntryCount: 2},
				},
			},
			{AgentID: "child-2", EntryCount: 4},
		},
	}

	nodes := FlattenTree(root)

	if len(nodes) != 4 {
		t.Errorf("FlattenTree() returned %d nodes, want 4", len(nodes))
	}

	// Verify all nodes are present
	ids := make(map[string]bool)
	for _, node := range nodes {
		if node.IsRoot {
			ids["root"] = true
		} else {
			ids[node.AgentID] = true
		}
	}

	expected := []string{"root", "child-1", "child-2", "grandchild-1"}
	for _, exp := range expected {
		if !ids[exp] {
			t.Errorf("FlattenTree() missing node: %s", exp)
		}
	}
}

func TestCountTotalEntries_Nested(t *testing.T) {
	root := &TreeNode{
		EntryCount: 10,
		Children: []*TreeNode{
			{
				EntryCount: 5,
				Children: []*TreeNode{
					{EntryCount: 3},
					{EntryCount: 2},
				},
			},
			{EntryCount: 7},
		},
	}

	total := CountTotalEntries(root)
	// 10 + 5 + 3 + 2 + 7 = 27
	if total != 27 {
		t.Errorf("CountTotalEntries() = %d, want 27", total)
	}
}

func TestCountTotalEntries_EmptyTree(t *testing.T) {
	root := &TreeNode{
		EntryCount: 0,
		IsRoot:     true,
	}

	total := CountTotalEntries(root)
	if total != 0 {
		t.Errorf("CountTotalEntries() for empty tree = %d, want 0", total)
	}
}

func TestFlattenTree_Nil(t *testing.T) {
	nodes := FlattenTree(nil)
	if len(nodes) != 0 {
		t.Errorf("FlattenTree(nil) returned %d nodes, want 0", len(nodes))
	}
}
