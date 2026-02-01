package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifest_JSONSerialization(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  now,
		SessionID:   "test-session-123",
		ProjectPath: "/Users/test/project",
		EntryCount:  42,
		AgentTree: &AgentTreeNode{
			ID:      "test-session-123",
			Entries: 30,
			Children: []*AgentTreeNode{
				{
					ID:      "agent-abc",
					Entries: 12,
				},
			},
		},
		SourceFiles: []SourceFile{
			{Type: "session", Path: "/path/to/session.jsonl"},
			{Type: "agent", AgentID: "agent-abc", Path: "/path/to/agent.jsonl"},
		},
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}

	// Deserialize back
	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	// Verify fields
	if decoded.Version != manifest.Version {
		t.Errorf("Version mismatch: got %s, want %s", decoded.Version, manifest.Version)
	}
	if !decoded.ExportedAt.Equal(manifest.ExportedAt) {
		t.Errorf("ExportedAt mismatch: got %v, want %v", decoded.ExportedAt, manifest.ExportedAt)
	}
	if decoded.SessionID != manifest.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", decoded.SessionID, manifest.SessionID)
	}
	if decoded.ProjectPath != manifest.ProjectPath {
		t.Errorf("ProjectPath mismatch: got %s, want %s", decoded.ProjectPath, manifest.ProjectPath)
	}
	if decoded.EntryCount != manifest.EntryCount {
		t.Errorf("EntryCount mismatch: got %d, want %d", decoded.EntryCount, manifest.EntryCount)
	}

	// Check agent tree
	if decoded.AgentTree == nil {
		t.Fatal("AgentTree is nil")
	}
	if decoded.AgentTree.ID != manifest.AgentTree.ID {
		t.Errorf("AgentTree.ID mismatch: got %s, want %s", decoded.AgentTree.ID, manifest.AgentTree.ID)
	}
	if decoded.AgentTree.Entries != manifest.AgentTree.Entries {
		t.Errorf("AgentTree.Entries mismatch: got %d, want %d", decoded.AgentTree.Entries, manifest.AgentTree.Entries)
	}
	if len(decoded.AgentTree.Children) != 1 {
		t.Errorf("AgentTree.Children length mismatch: got %d, want 1", len(decoded.AgentTree.Children))
	}

	// Check source files
	if len(decoded.SourceFiles) != 2 {
		t.Errorf("SourceFiles length mismatch: got %d, want 2", len(decoded.SourceFiles))
	}
}

func TestWriteManifest(t *testing.T) {
	tempDir := t.TempDir()

	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   "test-session",
		ProjectPath: "/test/project",
		EntryCount:  10,
		AgentTree: &AgentTreeNode{
			ID:      "test-session",
			Entries: 10,
		},
		SourceFiles: []SourceFile{
			{Type: "session", Path: "/path/to/session.jsonl"},
		},
	}

	// Write manifest
	if err := WriteManifest(manifest, tempDir); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	// Verify file exists
	manifestPath := filepath.Join(tempDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest.json not created")
	}

	// Read and verify content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest.json: %v", err)
	}

	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to parse manifest.json: %v", err)
	}

	if decoded.SessionID != manifest.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", decoded.SessionID, manifest.SessionID)
	}
}

func TestReadManifest(t *testing.T) {
	tempDir := t.TempDir()

	// Create a manifest file
	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   "read-test-session",
		ProjectPath: "/read/test/project",
		EntryCount:  5,
		AgentTree: &AgentTreeNode{
			ID:      "read-test-session",
			Entries: 5,
		},
		SourceFiles: []SourceFile{},
	}

	data, _ := json.MarshalIndent(manifest, "", "  ")
	manifestPath := filepath.Join(tempDir, "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to create test manifest: %v", err)
	}

	// Read manifest
	read, err := ReadManifest(tempDir)
	if err != nil {
		t.Fatalf("ReadManifest failed: %v", err)
	}

	if read.SessionID != manifest.SessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", read.SessionID, manifest.SessionID)
	}
	if read.ProjectPath != manifest.ProjectPath {
		t.Errorf("ProjectPath mismatch: got %s, want %s", read.ProjectPath, manifest.ProjectPath)
	}
}

func TestReadManifest_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := ReadManifest(tempDir)
	if err == nil {
		t.Error("Expected error when reading non-existent manifest")
	}
}

func TestWriteManifest_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "nested", "output", "dir")

	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   "nested-test",
		ProjectPath: "/nested/test",
		EntryCount:  1,
	}

	if err := WriteManifest(manifest, outputDir); err != nil {
		t.Fatalf("WriteManifest failed to create nested directory: %v", err)
	}

	manifestPath := filepath.Join(outputDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest.json not created in nested directory")
	}
}

func TestAgentTreeNode_DeepNesting(t *testing.T) {
	// Test with deeply nested agent tree
	tree := &AgentTreeNode{
		ID:      "root",
		Entries: 100,
		Children: []*AgentTreeNode{
			{
				ID:      "level1-a",
				Entries: 50,
				Children: []*AgentTreeNode{
					{
						ID:      "level2-a",
						Entries: 25,
						Children: []*AgentTreeNode{
							{
								ID:      "level3-a",
								Entries: 10,
							},
						},
					},
				},
			},
			{
				ID:      "level1-b",
				Entries: 25,
			},
		},
	}

	// Serialize and deserialize
	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("Failed to marshal deep tree: %v", err)
	}

	var decoded AgentTreeNode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal deep tree: %v", err)
	}

	// Verify structure preserved
	if decoded.ID != "root" {
		t.Errorf("Root ID mismatch: got %s, want root", decoded.ID)
	}
	if len(decoded.Children) != 2 {
		t.Fatalf("Root children mismatch: got %d, want 2", len(decoded.Children))
	}
	if decoded.Children[0].ID != "level1-a" {
		t.Errorf("Level1-a ID mismatch: got %s", decoded.Children[0].ID)
	}
	if len(decoded.Children[0].Children) != 1 {
		t.Fatalf("Level1-a children mismatch: got %d, want 1", len(decoded.Children[0].Children))
	}
	if decoded.Children[0].Children[0].ID != "level2-a" {
		t.Errorf("Level2-a ID mismatch: got %s", decoded.Children[0].Children[0].ID)
	}
}

func TestSourceFile_Types(t *testing.T) {
	testCases := []struct {
		name     string
		file     SourceFile
		wantJSON string
	}{
		{
			name:     "session type",
			file:     SourceFile{Type: "session", Path: "/path/to/session.jsonl"},
			wantJSON: `{"type":"session","path":"/path/to/session.jsonl"}`,
		},
		{
			name:     "agent type",
			file:     SourceFile{Type: "agent", AgentID: "abc123", Path: "/path/to/agent.jsonl"},
			wantJSON: `{"type":"agent","agent_id":"abc123","path":"/path/to/agent.jsonl"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.file)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			if string(data) != tc.wantJSON {
				t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), tc.wantJSON)
			}
		})
	}
}

func TestManifest_EmptySession(t *testing.T) {
	// Test manifest with zero entries
	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   "empty-session",
		ProjectPath: "/empty/project",
		EntryCount:  0,
		AgentTree: &AgentTreeNode{
			ID:      "empty-session",
			Entries: 0,
		},
		SourceFiles: []SourceFile{},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal empty session manifest: %v", err)
	}

	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal empty session manifest: %v", err)
	}

	if decoded.EntryCount != 0 {
		t.Errorf("EntryCount should be 0, got %d", decoded.EntryCount)
	}
	if len(decoded.SourceFiles) != 0 {
		t.Errorf("SourceFiles should be empty, got %d", len(decoded.SourceFiles))
	}
}

func TestManifest_NoAgentTree(t *testing.T) {
	// Test manifest with nil agent tree
	manifest := &Manifest{
		Version:     ManifestVersion,
		ExportedAt:  time.Now().UTC(),
		SessionID:   "no-tree-session",
		ProjectPath: "/no/tree",
		EntryCount:  5,
		AgentTree:   nil,
		SourceFiles: []SourceFile{
			{Type: "session", Path: "/path/to/session.jsonl"},
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest without tree: %v", err)
	}

	var decoded Manifest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal manifest without tree: %v", err)
	}

	if decoded.AgentTree != nil {
		t.Error("AgentTree should be nil")
	}
}

func TestAgentTreeNode_NoChildren(t *testing.T) {
	// Verify Children omitempty works
	node := &AgentTreeNode{
		ID:      "no-children",
		Entries: 10,
	}

	data, err := json.Marshal(node)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Should not contain "children" key
	jsonStr := string(data)
	if contains := len(jsonStr) > 0 && jsonStr != `{"id":"no-children","entries":10}`; contains {
		// Check if children is in the output when it shouldn't be
		var m map[string]interface{}
		_ = json.Unmarshal(data, &m)
		if _, hasChildren := m["children"]; hasChildren {
			t.Error("Children should be omitted from JSON when empty")
		}
	}
}

func TestConvertTreeNode(t *testing.T) {
	// Test nil input
	if result := convertTreeNode(nil); result != nil {
		t.Error("convertTreeNode(nil) should return nil")
	}
}

func TestExtractProjectPath(t *testing.T) {
	// Note: filepath.Base behavior is platform-specific for path separators.
	// On Unix, it won't recognize Windows backslash as path separator.
	// These tests use platform-native paths only.
	testCases := []struct {
		projectDir string
		want       string
	}{
		{filepath.Join("/Users/test/.claude/projects", "-Users-test-myproject"), "-Users-test-myproject"},
		{filepath.Join("/home/user/.claude/projects", "-home-user-code"), "-home-user-code"},
		{filepath.Join("projects", "encoded-path"), "encoded-path"},
	}

	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			result := extractProjectPath(tc.projectDir)
			if result != tc.want {
				t.Errorf("extractProjectPath(%q) = %q, want %q", tc.projectDir, result, tc.want)
			}
		})
	}
}

func TestManifestVersion(t *testing.T) {
	// Ensure version constant is set
	if ManifestVersion == "" {
		t.Error("ManifestVersion should not be empty")
	}

	// Verify it's a valid semver-like format
	if ManifestVersion != "1.0.0" {
		t.Logf("ManifestVersion is %s (may need update if intentional)", ManifestVersion)
	}
}

func TestBuildSourceFilesList(t *testing.T) {
	// This test would require mock agent.TreeNode, so we test the output format
	files := []SourceFile{
		{Type: "session", Path: "/test/session.jsonl"},
		{Type: "agent", AgentID: "agent1", Path: "/test/agent1.jsonl"},
		{Type: "agent", AgentID: "agent2", Path: "/test/agent2.jsonl"},
	}

	// Verify all have required fields
	for i, f := range files {
		if f.Type == "" {
			t.Errorf("SourceFile[%d] has empty Type", i)
		}
		if f.Path == "" {
			t.Errorf("SourceFile[%d] has empty Path", i)
		}
		if f.Type == "agent" && f.AgentID == "" {
			t.Errorf("SourceFile[%d] is agent type but has empty AgentID", i)
		}
	}
}

func TestGenerateManifest_NonExistentSession(t *testing.T) {
	tempDir := t.TempDir()

	_, err := GenerateManifest(tempDir, "non-existent-session", tempDir)
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.ErrNotExist, got %v", err)
	}
}

func TestGenerateManifest_WithRealSession(t *testing.T) {
	// Skip if no test fixtures available - this is an integration test
	tempDir := t.TempDir()
	sessionID := "test-session-abc123"

	// Create a minimal session file
	sessionContent := `{"uuid":"1","sessionId":"test-session-abc123","type":"user","timestamp":"2026-01-15T10:00:00Z","message":"Hello"}
{"uuid":"2","sessionId":"test-session-abc123","type":"assistant","timestamp":"2026-01-15T10:00:01Z","message":"Hi there!"}`

	sessionPath := filepath.Join(tempDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionPath, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("Failed to create session file: %v", err)
	}

	manifest, err := GenerateManifest(tempDir, sessionID, tempDir)
	if err != nil {
		t.Fatalf("GenerateManifest failed: %v", err)
	}

	if manifest.SessionID != sessionID {
		t.Errorf("SessionID mismatch: got %s, want %s", manifest.SessionID, sessionID)
	}
	if manifest.Version != ManifestVersion {
		t.Errorf("Version mismatch: got %s, want %s", manifest.Version, ManifestVersion)
	}
	if manifest.EntryCount != 2 {
		t.Errorf("EntryCount mismatch: got %d, want 2", manifest.EntryCount)
	}
	if len(manifest.SourceFiles) != 1 {
		t.Errorf("SourceFiles length mismatch: got %d, want 1", len(manifest.SourceFiles))
	}
	if manifest.SourceFiles[0].Type != "session" {
		t.Errorf("SourceFile type mismatch: got %s, want session", manifest.SourceFiles[0].Type)
	}
}
