package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// mustMkdirAll creates directories or fails the test
func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0750); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", path, err)
	}
}

// mustWriteFile writes a file or fails the test
func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", path, err)
	}
}

// createAgentSpawnEntryModern creates a modern-format agent spawn entry (user with toolUseResult).
// This matches the real Claude Code format where agent spawns are recorded as user entries
// with toolUseResult containing agent spawn information.
func createAgentSpawnEntryModern(sessionID, uuid, parentUUID, toolUseID, agentID, description string) string {
	return `{"type":"user","sessionId":"` + sessionID + `","uuid":"` + uuid + `","parentUuid":"` + parentUUID + `","sourceToolAssistantUUID":"` + parentUUID + `","message":[{"type":"tool_result","tool_use_id":"` + toolUseID + `","content":[]}],"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"` + agentID + `","description":"` + description + `","prompt":"` + description + `","outputFile":"/tmp/claude/tasks/` + agentID + `.output"}}
`
}

// createAssistantWithTaskTool creates an assistant entry with a Task tool_use.
func createAssistantWithTaskTool(sessionID, uuid, toolUseID, prompt string) string {
	return `{"type":"assistant","sessionId":"` + sessionID + `","uuid":"` + uuid + `","message":[{"type":"text","text":"I'll spawn an agent."},{"type":"tool_use","id":"` + toolUseID + `","name":"Task","input":{"prompt":"` + prompt + `"}}]}
`
}

// createAgentFileEntry creates an entry for an agent file with proper agentId field.
func createAgentFileEntry(sessionID, agentID, uuid, entryType, message string) string {
	if entryType == "user" {
		return `{"type":"user","sessionId":"` + sessionID + `","agentId":"` + agentID + `","uuid":"` + uuid + `","parentUuid":null,"message":"` + message + `"}
`
	}
	return `{"type":"assistant","sessionId":"` + sessionID + `","agentId":"` + agentID + `","uuid":"` + uuid + `","message":[{"type":"text","text":"` + message + `"}]}
`
}

func TestParseAgentType(t *testing.T) {
	tests := []struct {
		agentID  string
		expected string
	}{
		{"a12eb64", ""},
		{"aprompt_suggestion-abc123", "prompt_suggestion"},
		{"aexplore-def456", "explore"},
		{"random", ""},
	}

	for _, tt := range tests {
		t.Run(tt.agentID, func(t *testing.T) {
			result := parseAgentType(tt.agentID)
			if result != tt.expected {
				t.Errorf("parseAgentType(%q) = %q, want %q", tt.agentID, result, tt.expected)
			}
		})
	}
}

func TestDiscoverAgents(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Create agent files with proper agentId field (matching real Claude Code format)
	agent1Content := `{"uuid":"1","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"user","parentUuid":null,"message":"Agent task"}
{"uuid":"2","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"assistant","message":[{"type":"text","text":"Agent response"}]}
`
	agent2Content := `{"uuid":"1","sessionId":"` + sessionID + `","agentId":"aprompt_suggestion-abc","type":"system","message":"System event"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agent1Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-aprompt_suggestion-abc.jsonl"), []byte(agent2Content))

	agents, err := DiscoverAgents(sessionDir)
	if err != nil {
		t.Fatalf("DiscoverAgents() error: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("DiscoverAgents() returned %d agents, want 2", len(agents))
	}

	// Check that we found the right agents
	foundA12 := false
	foundPrompt := false
	for _, agent := range agents {
		if agent.ID == "a12eb64" {
			foundA12 = true
			if agent.EntryCount != 2 {
				t.Errorf("Agent a12eb64 entry count = %d, want 2", agent.EntryCount)
			}
		}
		if agent.ID == "aprompt_suggestion-abc" {
			foundPrompt = true
			if agent.AgentType != "prompt_suggestion" {
				t.Errorf("Agent type = %q, want 'prompt_suggestion'", agent.AgentType)
			}
		}
	}

	if !foundA12 {
		t.Error("Agent a12eb64 not found")
	}
	if !foundPrompt {
		t.Error("Agent aprompt_suggestion-abc not found")
	}
}

// TestFindAgentSpawns tests detection using LEGACY queue-operation format.
// This is the original Claude Code format where agent spawns are recorded as queue-operation entries.
// Note: Modern Claude Code uses user entries with toolUseResult for agent spawning,
// but the code still supports queue-operation for backward compatibility.
func TestFindAgentSpawns(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// Legacy format: queue-operation entries with agentId
	content := `{"uuid":"1","type":"user"}
{"uuid":"2","type":"queue-operation","agentId":"a12eb64"}
{"uuid":"3","type":"assistant"}
{"uuid":"4","type":"queue-operation","agentId":"a68b8c0"}
`
	mustWriteFile(t, sessionFile, []byte(content))

	spawns, err := FindAgentSpawns(sessionFile)
	if err != nil {
		t.Fatalf("FindAgentSpawns() error: %v", err)
	}

	if len(spawns) != 2 {
		t.Errorf("FindAgentSpawns() found %d spawns, want 2", len(spawns))
	}

	if spawns["a12eb64"] != "2" {
		t.Errorf("Spawn for a12eb64 = %q, want '2'", spawns["a12eb64"])
	}

	if spawns["a68b8c0"] != "4" {
		t.Errorf("Spawn for a68b8c0 = %q, want '4'", spawns["a68b8c0"])
	}
}

// TestFindAgentSpawns_BothFormats tests detection with both legacy and modern formats present.
// Real Claude Code may emit both formats for compatibility, so we should handle both.
func TestFindAgentSpawns_BothFormats(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	// Content includes both modern (user with toolUseResult) and legacy (queue-operation) formats
	// The current implementation only detects queue-operation, but the test data shows both formats
	// so future implementations can detect modern format when data model is updated.
	content := `{"uuid":"1","type":"user","sessionId":"` + sessionID + `"}
{"uuid":"2","type":"assistant","sessionId":"` + sessionID + `","message":[{"type":"text","text":"I'll spawn an agent."},{"type":"tool_use","id":"toolu_01","name":"Task","input":{"prompt":"Task 1"}}]}
{"uuid":"3","type":"user","sessionId":"` + sessionID + `","parentUuid":"2","sourceToolAssistantUUID":"2","message":[{"type":"tool_result","tool_use_id":"toolu_01","content":[]}],"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a12eb64","description":"Task 1","prompt":"Task 1","outputFile":"/tmp/claude/tasks/a12eb64.output"}}
{"uuid":"4","type":"queue-operation","sessionId":"` + sessionID + `","agentId":"a12eb64"}
{"uuid":"5","type":"assistant","sessionId":"` + sessionID + `"}
`
	mustWriteFile(t, sessionFile, []byte(content))

	spawns, err := FindAgentSpawns(sessionFile)
	if err != nil {
		t.Fatalf("FindAgentSpawns() error: %v", err)
	}

	// Current implementation detects via queue-operation
	if len(spawns) != 1 {
		t.Errorf("FindAgentSpawns() found %d spawns, want 1 (from queue-operation)", len(spawns))
	}

	if spawns["a12eb64"] != "4" {
		t.Errorf("Spawn for a12eb64 = %q, want '4'", spawns["a12eb64"])
	}
}

// TestBuildTree tests tree building with legacy queue-operation spawn format.
func TestBuildTree(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file with legacy queue-operation format
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"` + sessionID + `","type":"user"}
{"uuid":"2","sessionId":"` + sessionID + `","type":"assistant"}
{"uuid":"3","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent file with proper agentId field
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agentContent := `{"uuid":"a1","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"user","parentUuid":null}
{"uuid":"a2","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"assistant"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agentContent))

	tree, err := BuildTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root node should have IsRoot=true")
	}

	if tree.SessionID != sessionID {
		t.Errorf("Root SessionID = %q, want %q", tree.SessionID, sessionID)
	}

	if tree.EntryCount != 3 {
		t.Errorf("Root entry count = %d, want 3", tree.EntryCount)
	}

	if len(tree.Children) != 1 {
		t.Errorf("Root has %d children, want 1", len(tree.Children))
	}

	if len(tree.Children) > 0 {
		child := tree.Children[0]
		if child.AgentID != "a12eb64" {
			t.Errorf("Child AgentID = %q, want 'a12eb64'", child.AgentID)
		}
		if child.EntryCount != 2 {
			t.Errorf("Child entry count = %d, want 2", child.EntryCount)
		}
	}
}

// TestBuildTree_ModernFormat tests tree building with session data in modern Claude Code format.
// Modern format uses user entries with toolUseResult for agent spawning, but also includes
// queue-operation entries for backward compatibility.
func TestBuildTree_ModernFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file with BOTH modern and legacy formats
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"` + sessionID + `","type":"user","message":"Hello"}
{"uuid":"2","sessionId":"` + sessionID + `","type":"assistant","message":[{"type":"text","text":"I'll help."},{"type":"tool_use","id":"toolu_01","name":"Task","input":{"prompt":"Explore"}}]}
{"uuid":"3","sessionId":"` + sessionID + `","type":"user","parentUuid":"2","sourceToolAssistantUUID":"2","message":[{"type":"tool_result","tool_use_id":"toolu_01","content":[]}],"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a12eb64","description":"Explore","prompt":"Explore","outputFile":"/tmp/claude/tasks/a12eb64.output"}}
{"uuid":"4","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent file
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agentContent := `{"uuid":"a1","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"user","parentUuid":null,"message":"Explore"}
{"uuid":"a2","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"assistant","message":[{"type":"text","text":"Exploring..."}]}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agentContent))

	tree, err := BuildTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root node should have IsRoot=true")
	}

	// With both modern and legacy entries, we have 4 entries in main session
	if tree.EntryCount != 4 {
		t.Errorf("Root entry count = %d, want 4", tree.EntryCount)
	}

	if len(tree.Children) != 1 {
		t.Errorf("Root has %d children, want 1", len(tree.Children))
	}

	if len(tree.Children) > 0 {
		child := tree.Children[0]
		if child.AgentID != "a12eb64" {
			t.Errorf("Child AgentID = %q, want 'a12eb64'", child.AgentID)
		}
	}
}

func TestCountTotalEntries(t *testing.T) {
	root := &TreeNode{
		EntryCount: 10,
		Children: []*TreeNode{
			{EntryCount: 5},
			{EntryCount: 3, Children: []*TreeNode{
				{EntryCount: 2},
			}},
		},
	}

	total := CountTotalEntries(root)
	if total != 20 {
		t.Errorf("CountTotalEntries() = %d, want 20", total)
	}
}

// TestBuildNestedTree_SingleLevel tests single-level nesting with legacy queue-operation format.
func TestBuildNestedTree_SingleLevel(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file with queue-operations spawning agents
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","sessionId":"` + sessionID + `","type":"user"}
{"uuid":"main-2","sessionId":"` + sessionID + `","type":"assistant"}
{"uuid":"spawn-1","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64","parentUuid":"main-2"}
{"uuid":"spawn-2","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a68b8c0","parentUuid":"main-2"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent files with proper agentId field
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agent1Content := `{"uuid":"a1-1","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"user","parentUuid":null}
{"uuid":"a1-2","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"assistant"}
`
	agent2Content := `{"uuid":"a2-1","sessionId":"` + sessionID + `","agentId":"a68b8c0","type":"user","parentUuid":null}
{"uuid":"a2-2","sessionId":"` + sessionID + `","agentId":"a68b8c0","type":"assistant"}
{"uuid":"a2-3","sessionId":"` + sessionID + `","agentId":"a68b8c0","type":"assistant"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agent1Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a68b8c0.jsonl"), []byte(agent2Content))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root node should have IsRoot=true")
	}

	if tree.EntryCount != 4 {
		t.Errorf("Root entry count = %d, want 4", tree.EntryCount)
	}

	if len(tree.Children) != 2 {
		t.Errorf("Root has %d children, want 2", len(tree.Children))
	}

	// Verify both agents are direct children of root
	childIDs := make(map[string]bool)
	for _, child := range tree.Children {
		childIDs[child.AgentID] = true
	}
	if !childIDs["a12eb64"] || !childIDs["a68b8c0"] {
		t.Errorf("Expected children a12eb64 and a68b8c0, got %v", childIDs)
	}
}

// TestBuildNestedTree_SingleLevel_ModernFormat tests single-level nesting with modern toolUseResult format.
func TestBuildNestedTree_SingleLevel_ModernFormat(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file with BOTH modern and legacy formats for agent spawning
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","sessionId":"` + sessionID + `","type":"user","message":"Hello"}
{"uuid":"main-2","sessionId":"` + sessionID + `","type":"assistant","message":[{"type":"text","text":"I'll spawn agents."},{"type":"tool_use","id":"toolu_01","name":"Task","input":{"prompt":"Task 1"}}]}
{"uuid":"spawn-1-modern","sessionId":"` + sessionID + `","type":"user","parentUuid":"main-2","sourceToolAssistantUUID":"main-2","message":[{"type":"tool_result","tool_use_id":"toolu_01","content":[]}],"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a12eb64","description":"Task 1","prompt":"Task 1","outputFile":"/tmp/claude/tasks/a12eb64.output"}}
{"uuid":"spawn-1","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64","parentUuid":"main-2"}
{"uuid":"main-3","sessionId":"` + sessionID + `","type":"assistant","message":[{"type":"tool_use","id":"toolu_02","name":"Task","input":{"prompt":"Task 2"}}]}
{"uuid":"spawn-2-modern","sessionId":"` + sessionID + `","type":"user","parentUuid":"main-3","sourceToolAssistantUUID":"main-3","message":[{"type":"tool_result","tool_use_id":"toolu_02","content":[]}],"toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a68b8c0","description":"Task 2","prompt":"Task 2","outputFile":"/tmp/claude/tasks/a68b8c0.output"}}
{"uuid":"spawn-2","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a68b8c0","parentUuid":"main-3"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create agent files with proper agentId field
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agent1Content := `{"uuid":"a1-1","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"user","parentUuid":null,"message":"Task 1"}
{"uuid":"a1-2","sessionId":"` + sessionID + `","agentId":"a12eb64","type":"assistant","message":[{"type":"text","text":"Working on task 1"}]}
`
	agent2Content := `{"uuid":"a2-1","sessionId":"` + sessionID + `","agentId":"a68b8c0","type":"user","parentUuid":null,"message":"Task 2"}
{"uuid":"a2-2","sessionId":"` + sessionID + `","agentId":"a68b8c0","type":"assistant","message":[{"type":"text","text":"Working on task 2"}]}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agent1Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a68b8c0.jsonl"), []byte(agent2Content))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root node should have IsRoot=true")
	}

	// Main session has 7 entries: user, assistant, spawn-modern, queue-op, assistant, spawn-modern, queue-op
	if tree.EntryCount != 7 {
		t.Errorf("Root entry count = %d, want 7", tree.EntryCount)
	}

	if len(tree.Children) != 2 {
		t.Errorf("Root has %d children, want 2", len(tree.Children))
	}

	// Verify both agents are direct children of root
	childIDs := make(map[string]bool)
	for _, child := range tree.Children {
		childIDs[child.AgentID] = true
	}
	if !childIDs["a12eb64"] || !childIDs["a68b8c0"] {
		t.Errorf("Expected children a12eb64 and a68b8c0, got %v", childIDs)
	}
}

func TestBuildNestedTree_TwoLevelsDeep(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file - spawns agent-a12eb64
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","sessionId":"` + sessionID + `","type":"user"}
{"uuid":"spawn-a12","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Create session directory and subagents
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Agent a12eb64 spawns agent nested-child
	agent1Content := `{"uuid":"a1-1","type":"user"}
{"uuid":"spawn-nested","type":"queue-operation","agentId":"nested-child","parentUuid":"a12eb64"}
{"uuid":"a1-2","type":"assistant"}
`
	nestedContent := `{"uuid":"n1","type":"user"}
{"uuid":"n2","type":"assistant"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agent1Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-nested-child.jsonl"), []byte(nestedContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Root should have 1 direct child (a12eb64)
	if len(tree.Children) != 1 {
		t.Fatalf("Root has %d children, want 1", len(tree.Children))
	}

	child := tree.Children[0]
	if child.AgentID != "a12eb64" {
		t.Errorf("First child AgentID = %q, want 'a12eb64'", child.AgentID)
	}

	// a12eb64 should have 1 child (nested-child)
	if len(child.Children) != 1 {
		t.Fatalf("Agent a12eb64 has %d children, want 1", len(child.Children))
	}

	grandchild := child.Children[0]
	if grandchild.AgentID != "nested-child" {
		t.Errorf("Grandchild AgentID = %q, want 'nested-child'", grandchild.AgentID)
	}
	if grandchild.EntryCount != 2 {
		t.Errorf("Grandchild entry count = %d, want 2", grandchild.EntryCount)
	}
}

func TestBuildNestedTree_ThreeLevelsDeep(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session spawns level1
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-l1","type":"queue-operation","agentId":"level1"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// level1 spawns level2
	level1Content := `{"uuid":"l1-1","type":"user"}
{"uuid":"spawn-l2","type":"queue-operation","agentId":"level2","parentUuid":"level1"}
`
	// level2 spawns level3
	level2Content := `{"uuid":"l2-1","type":"user"}
{"uuid":"spawn-l3","type":"queue-operation","agentId":"level3","parentUuid":"level2"}
`
	level3Content := `{"uuid":"l3-1","type":"user"}
{"uuid":"l3-2","type":"assistant"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-level1.jsonl"), []byte(level1Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-level2.jsonl"), []byte(level2Content))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-level3.jsonl"), []byte(level3Content))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Verify three levels of nesting
	if len(tree.Children) != 1 {
		t.Fatalf("Root has %d children, want 1", len(tree.Children))
	}
	if tree.Children[0].AgentID != "level1" {
		t.Errorf("Level 1 agent = %q, want 'level1'", tree.Children[0].AgentID)
	}

	level1 := tree.Children[0]
	if len(level1.Children) != 1 {
		t.Fatalf("level1 has %d children, want 1", len(level1.Children))
	}
	if level1.Children[0].AgentID != "level2" {
		t.Errorf("Level 2 agent = %q, want 'level2'", level1.Children[0].AgentID)
	}

	level2 := level1.Children[0]
	if len(level2.Children) != 1 {
		t.Fatalf("level2 has %d children, want 1", len(level2.Children))
	}
	if level2.Children[0].AgentID != "level3" {
		t.Errorf("Level 3 agent = %q, want 'level3'", level2.Children[0].AgentID)
	}
}

func TestBuildNestedTree_MultipleChildrenSameLevel(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session spawns 3 agents
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-a","type":"queue-operation","agentId":"agent-a"}
{"uuid":"spawn-b","type":"queue-operation","agentId":"agent-b"}
{"uuid":"spawn-c","type":"queue-operation","agentId":"agent-c"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agentContent := `{"uuid":"1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-a.jsonl"), []byte(agentContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-b.jsonl"), []byte(agentContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-c.jsonl"), []byte(agentContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	if len(tree.Children) != 3 {
		t.Errorf("Root has %d children, want 3", len(tree.Children))
	}

	childIDs := make(map[string]bool)
	for _, child := range tree.Children {
		childIDs[child.AgentID] = true
	}
	expected := []string{"agent-a", "agent-b", "agent-c"}
	for _, exp := range expected {
		if !childIDs[exp] {
			t.Errorf("Missing expected child: %s", exp)
		}
	}
}

func TestBuildNestedTree_OrphanedAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session with no queue-operation for the orphan
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-known","type":"queue-operation","agentId":"known-agent"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Create both agents - orphan has no spawn record and parentUuid points to non-existent agent
	knownContent := `{"uuid":"k1","type":"user"}
`
	// Orphan claims a parent that doesn't exist
	orphanContent := `{"uuid":"o1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-known-agent.jsonl"), []byte(knownContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-orphan-agent.jsonl"), []byte(orphanContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Both agents should be attached to root (orphan has no parent info)
	if len(tree.Children) != 2 {
		t.Errorf("Root has %d children, want 2 (known + orphan)", len(tree.Children))
	}

	childIDs := make(map[string]bool)
	for _, child := range tree.Children {
		childIDs[child.AgentID] = true
	}
	if !childIDs["known-agent"] {
		t.Error("known-agent not found in children")
	}
	if !childIDs["orphan-agent"] {
		t.Error("orphan-agent not found in children (should attach to root)")
	}
}

func TestBuildNestedTree_EmptySession(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create empty session file
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	mustWriteFile(t, sessionFile, []byte(""))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root node should have IsRoot=true")
	}
	if len(tree.Children) != 0 {
		t.Errorf("Empty session should have 0 children, got %d", len(tree.Children))
	}
	if tree.EntryCount != 0 {
		t.Errorf("Empty session entry count = %d, want 0", tree.EntryCount)
	}
}

func TestBuildNestedTree_SessionWithNoSubagents(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create session file with entries but no queue-operations
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","type":"user"}
{"uuid":"2","type":"assistant"}
{"uuid":"3","type":"user"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Don't create subagents directory

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	if !tree.IsRoot {
		t.Error("Root should have IsRoot=true")
	}
	if tree.EntryCount != 3 {
		t.Errorf("Root entry count = %d, want 3", tree.EntryCount)
	}
	if len(tree.Children) != 0 {
		t.Errorf("Should have 0 children, got %d", len(tree.Children))
	}
}

func TestFindParentAgent(t *testing.T) {
	agents := map[string]*TreeNode{
		"agent-a": {AgentID: "agent-a", UUID: "uuid-a"},
		"agent-b": {AgentID: "agent-b", UUID: "uuid-b"},
		"agent-c": {AgentID: "agent-c", UUID: "uuid-c"},
	}

	tests := []struct {
		name       string
		parentUUID string
		wantID     string
		wantNil    bool
	}{
		{
			name:       "find by agent ID",
			parentUUID: "agent-a",
			wantID:     "agent-a",
		},
		{
			name:       "find by UUID",
			parentUUID: "uuid-b",
			wantID:     "agent-b",
		},
		{
			name:       "parent not found",
			parentUUID: "nonexistent",
			wantNil:    true,
		},
		{
			name:       "empty UUID returns nil",
			parentUUID: "",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindParentAgent(agents, tt.parentUUID)
			if tt.wantNil {
				if result != nil {
					t.Errorf("FindParentAgent() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Fatalf("FindParentAgent() = nil, want agent %q", tt.wantID)
				}
				if result.AgentID != tt.wantID {
					t.Errorf("FindParentAgent().AgentID = %q, want %q", result.AgentID, tt.wantID)
				}
			}
		})
	}
}

func TestBuildNestedTree_CircularReference(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session spawns agent-a
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-a","type":"queue-operation","agentId":"agent-a"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// agent-a spawns agent-b with parentUuid = agent-a
	agentAContent := `{"uuid":"a1","type":"user"}
{"uuid":"spawn-b","type":"queue-operation","agentId":"agent-b","parentUuid":"agent-a"}
`
	// agent-b tries to spawn with parentUuid = agent-a (circular back to grandparent)
	// This shouldn't cause infinite loops
	agentBContent := `{"uuid":"b1","type":"user"}
{"uuid":"spawn-c","type":"queue-operation","agentId":"agent-c","parentUuid":"agent-a"}
`
	agentCContent := `{"uuid":"c1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-a.jsonl"), []byte(agentAContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-b.jsonl"), []byte(agentBContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-c.jsonl"), []byte(agentCContent))

	// This should complete without infinite loops
	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Verify tree was built (exact structure depends on cycle handling)
	if tree == nil {
		t.Fatal("BuildNestedTree() returned nil")
	}

	// All agents should be in the tree somewhere
	total := CountTotalEntries(tree)
	// main(2) + agent-a(2) + agent-b(2) + agent-c(1) = 7
	if total != 7 {
		t.Errorf("Total entries = %d, want 7", total)
	}
}

func TestBuildNestedTree_SelfReferencingAgent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session spawns self-ref agent
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-self","type":"queue-operation","agentId":"self-ref"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	// Agent with self-reference in parentUuid
	selfRefContent := `{"uuid":"s1","type":"user"}
{"uuid":"spawn-nested","type":"queue-operation","agentId":"nested","parentUuid":"self-ref"}
`
	nestedContent := `{"uuid":"n1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-self-ref.jsonl"), []byte(selfRefContent))
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-nested.jsonl"), []byte(nestedContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// self-ref should be child of root
	if len(tree.Children) != 1 {
		t.Fatalf("Root has %d children, want 1", len(tree.Children))
	}

	selfRef := tree.Children[0]
	if selfRef.AgentID != "self-ref" {
		t.Errorf("First child = %q, want 'self-ref'", selfRef.AgentID)
	}

	// nested should be child of self-ref (normal parent-child)
	if len(selfRef.Children) != 1 {
		t.Fatalf("self-ref has %d children, want 1", len(selfRef.Children))
	}
	if selfRef.Children[0].AgentID != "nested" {
		t.Errorf("Nested agent = %q, want 'nested'", selfRef.Children[0].AgentID)
	}
}

func TestBuildNestedTree_InvalidParentUUID(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Main session spawns agent with invalid parentUuid format
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-1","type":"queue-operation","agentId":"agent-1","parentUuid":"!!!invalid!!!"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	mustMkdirAll(t, subagentsDir)

	agentContent := `{"uuid":"a1","type":"user"}
`
	mustWriteFile(t, filepath.Join(subagentsDir, "agent-agent-1.jsonl"), []byte(agentContent))

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Agent with invalid parent should be attached to root
	if len(tree.Children) != 1 {
		t.Fatalf("Root has %d children, want 1", len(tree.Children))
	}
	if tree.Children[0].AgentID != "agent-1" {
		t.Errorf("Child agent = %q, want 'agent-1'", tree.Children[0].AgentID)
	}
}

func TestBuildNestedTree_MissingSubagentsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create session file with queue-operation
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"main-1","type":"user"}
{"uuid":"spawn-1","type":"queue-operation","agentId":"phantom-agent"}
`
	mustWriteFile(t, sessionFile, []byte(sessionContent))

	// Don't create the session directory or subagents directory

	tree, err := BuildNestedTree(tmpDir, sessionID)
	if err != nil {
		t.Fatalf("BuildNestedTree() error: %v", err)
	}

	// Should handle missing directory gracefully - no children since files don't exist
	if len(tree.Children) != 0 {
		t.Errorf("Root has %d children, want 0 (subagents dir missing)", len(tree.Children))
	}
}
