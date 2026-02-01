package agent

import (
	"os"
	"path/filepath"
	"testing"
)

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
	sessionDir := filepath.Join(tmpDir, "679761ba-80c0-4cd3-a586-cc6a1fc56308")
	subagentsDir := filepath.Join(sessionDir, "subagents")
	os.MkdirAll(subagentsDir, 0755)

	// Create agent files
	agent1Content := `{"uuid":"1","sessionId":"test","type":"user"}
{"uuid":"2","sessionId":"test","type":"assistant"}
`
	agent2Content := `{"uuid":"1","sessionId":"test","type":"system"}
`
	os.WriteFile(filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agent1Content), 0644)
	os.WriteFile(filepath.Join(subagentsDir, "agent-aprompt_suggestion-abc.jsonl"), []byte(agent2Content), 0644)

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

func TestFindAgentSpawns(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "session.jsonl")

	content := `{"uuid":"1","type":"user"}
{"uuid":"2","type":"queue-operation","agentId":"a12eb64"}
{"uuid":"3","type":"assistant"}
{"uuid":"4","type":"queue-operation","agentId":"a68b8c0"}
`
	os.WriteFile(sessionFile, []byte(content), 0644)

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

func TestBuildTree(t *testing.T) {
	tmpDir := t.TempDir()
	sessionID := "679761ba-80c0-4cd3-a586-cc6a1fc56308"

	// Create main session file
	sessionFile := filepath.Join(tmpDir, sessionID+".jsonl")
	sessionContent := `{"uuid":"1","sessionId":"` + sessionID + `","type":"user"}
{"uuid":"2","sessionId":"` + sessionID + `","type":"assistant"}
{"uuid":"3","sessionId":"` + sessionID + `","type":"queue-operation","agentId":"a12eb64"}
`
	os.WriteFile(sessionFile, []byte(sessionContent), 0644)

	// Create agent file
	sessionDir := filepath.Join(tmpDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	os.MkdirAll(subagentsDir, 0755)

	agentContent := `{"uuid":"a1","type":"user"}
{"uuid":"a2","type":"assistant"}
`
	os.WriteFile(filepath.Join(subagentsDir, "agent-a12eb64.jsonl"), []byte(agentContent), 0644)

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
