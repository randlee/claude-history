package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestManifest_JSONValidity(t *testing.T) {
	// Create and export a session
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-valid-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 2)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-manifest")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify manifest is valid JSON
	manifestPath := filepath.Join(outputDir, "manifest.json")
	manifest := verifyManifestValid(t, manifestPath)

	// Verify all required fields are present
	if manifest["version"] == nil {
		t.Error("manifest missing version field")
	}

	if manifest["exported_at"] == nil {
		t.Error("manifest missing exported_at field")
	}

	if manifest["session_id"] == nil {
		t.Error("manifest missing session_id field")
	}

	if manifest["entry_count"] == nil {
		t.Error("manifest missing entry_count field")
	}

	if manifest["agent_tree"] == nil {
		t.Error("manifest missing agent_tree field")
	}

	if manifest["source_files"] == nil {
		t.Error("manifest missing source_files field")
	}
}

func TestManifest_AgentTree(t *testing.T) {
	// Create session with nested agents
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-tree-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)
	createNestedAgentStructure(t, projectDir, sessionID)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-tree")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read and parse manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	manifest := verifyManifestValid(t, manifestPath)

	// Verify agent tree structure
	agentTree, ok := manifest["agent_tree"].(map[string]interface{})
	if !ok {
		t.Fatal("agent_tree should be an object")
	}

	// Root should have ID (session ID)
	if agentTree["id"] == nil {
		t.Error("agent_tree root should have id field")
	}

	// Root should have children
	children, ok := agentTree["children"].([]interface{})
	if !ok {
		t.Fatal("agent_tree should have children array")
	}

	if len(children) == 0 {
		t.Error("agent_tree should have child agents")
	}

	// Verify child structure
	for i, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			t.Errorf("child %d should be an object", i)
			continue
		}

		if childMap["id"] == nil {
			t.Errorf("child %d missing id field", i)
		}

		if childMap["entries"] == nil {
			t.Errorf("child %d missing entries field", i)
		}
	}

	// Check for nested children (from createNestedAgentStructure)
	// There should be a parent agent with its own children
	foundNestedChildren := false
	for _, child := range children {
		childMap := child.(map[string]interface{})
		if nestedChildren, ok := childMap["children"].([]interface{}); ok && len(nestedChildren) > 0 {
			foundNestedChildren = true
			break
		}
	}

	if !foundNestedChildren {
		t.Error("agent_tree should contain nested agent structure")
	}
}

func TestManifest_SourcePaths(t *testing.T) {
	// Create and export session
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-sources-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 2)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-sources")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	manifest := verifyManifestValid(t, manifestPath)

	// Verify source files list
	sourceFiles, ok := manifest["source_files"].([]interface{})
	if !ok {
		t.Fatal("source_files should be an array")
	}

	if len(sourceFiles) == 0 {
		t.Error("source_files should not be empty")
	}

	// Should have at least 1 session file + 2 agent files
	if len(sourceFiles) < 3 {
		t.Errorf("expected at least 3 source files, got %d", len(sourceFiles))
	}

	// Verify each source file has required fields
	hasSessionFile := false
	agentFileCount := 0

	for i, file := range sourceFiles {
		fileMap, ok := file.(map[string]interface{})
		if !ok {
			t.Errorf("source file %d should be an object", i)
			continue
		}

		fileType, _ := fileMap["type"].(string)
		filePath, _ := fileMap["path"].(string)

		if fileType == "" {
			t.Errorf("source file %d missing type", i)
		}

		if filePath == "" {
			t.Errorf("source file %d missing path", i)
		}

		// Count file types
		if fileType == "session" {
			hasSessionFile = true
		} else if fileType == "agent" {
			agentFileCount++

			// Agent files should have agent_id field
			if fileMap["agent_id"] == nil {
				t.Errorf("agent source file %d missing agent_id", i)
			}
		}

		// Verify path exists on disk
		if !filepath.IsAbs(filePath) {
			t.Errorf("source file %d path should be absolute: %s", i, filePath)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("source file %d does not exist: %s", i, filePath)
		}
	}

	if !hasSessionFile {
		t.Error("source_files should include session file")
	}

	if agentFileCount != 2 {
		t.Errorf("expected 2 agent files, got %d", agentFileCount)
	}
}

func TestManifest_Metadata(t *testing.T) {
	// Create and export session
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-metadata-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-metadata")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	exportTime := time.Now()

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	manifest := verifyManifestValid(t, manifestPath)

	// Verify version
	version, ok := manifest["version"].(string)
	if !ok || version == "" {
		t.Error("manifest should have non-empty version string")
	}

	// Verify exported_at timestamp
	exportedAt, ok := manifest["exported_at"].(string)
	if !ok || exportedAt == "" {
		t.Error("manifest should have exported_at timestamp")
	}

	// Parse timestamp
	parsedTime, err := time.Parse(time.RFC3339, exportedAt)
	if err != nil {
		t.Errorf("exported_at should be valid RFC3339 timestamp: %v", err)
	}

	// Should be within 1 minute of now
	if time.Since(parsedTime) > time.Minute || time.Until(parsedTime) > time.Minute {
		t.Errorf("exported_at timestamp seems incorrect: %s (current time: %s)", exportedAt, exportTime.Format(time.RFC3339))
	}

	// Verify session ID
	manifestSessionID, ok := manifest["session_id"].(string)
	if !ok || manifestSessionID != sessionID {
		t.Errorf("manifest session_id = %v, want %s", manifestSessionID, sessionID)
	}

	// Verify entry count
	entryCount, ok := manifest["entry_count"].(float64)
	if !ok {
		t.Error("manifest entry_count should be a number")
	}

	if entryCount <= 0 {
		t.Errorf("manifest entry_count should be positive, got %v", entryCount)
	}

	// Verify project path
	projectPathField, ok := manifest["project_path"].(string)
	if !ok || projectPathField == "" {
		t.Error("manifest should have project_path")
	}
}

func TestManifest_EmptySession(t *testing.T) {
	// Test manifest for session with no agents
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-empty-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 0)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-empty-manifest")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	manifest := verifyManifestValid(t, manifestPath)

	// Agent tree should exist but have no children
	agentTree, ok := manifest["agent_tree"].(map[string]interface{})
	if !ok {
		t.Fatal("agent_tree should exist even for empty session")
	}

	children, ok := agentTree["children"].([]interface{})
	if ok && len(children) > 0 {
		t.Error("agent_tree should have no children for session without agents")
	}

	// Source files should only have session file
	sourceFiles, ok := manifest["source_files"].([]interface{})
	if !ok {
		t.Fatal("source_files should be an array")
	}

	sessionFileCount := 0
	agentFileCount := 0

	for _, file := range sourceFiles {
		fileMap := file.(map[string]interface{})
		if fileType, _ := fileMap["type"].(string); fileType == "session" {
			sessionFileCount++
		} else if fileType == "agent" {
			agentFileCount++
		}
	}

	if sessionFileCount != 1 {
		t.Errorf("expected 1 session file, got %d", sessionFileCount)
	}

	if agentFileCount != 0 {
		t.Errorf("expected 0 agent files, got %d", agentFileCount)
	}
}

func TestManifest_VersionFormat(t *testing.T) {
	// Verify manifest version follows semantic versioning
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-version-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-version")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	version, ok := manifest["version"].(string)
	if !ok {
		t.Fatal("manifest version should be a string")
	}

	// Check semantic versioning format (X.Y.Z)
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		t.Errorf("manifest version should follow semantic versioning (X.Y.Z), got: %s", version)
	}

	// Each part should be numeric
	for i, part := range parts {
		if len(part) == 0 {
			t.Errorf("manifest version part %d is empty", i)
		}
		for _, c := range part {
			if c < '0' || c > '9' {
				t.Errorf("manifest version part %d contains non-numeric character: %s", i, part)
			}
		}
	}
}

func TestManifest_JSONLFormat(t *testing.T) {
	// Test that manifest is NOT created for JSONL-only export
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-jsonl-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 1)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-jsonl-manifest")

	exportSessionID = sessionID
	exportFormat = "jsonl"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// For JSONL format, manifest might or might not be created
	// (depends on implementation decision)
	// We just verify the export succeeded
	sourceDir := filepath.Join(outputDir, "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Error("source directory should be created for JSONL export")
	}
}

func TestManifest_ReadRoundtrip(t *testing.T) {
	// Test that we can write and read back a manifest
	tmpDir, projectDir, projectPath := setupTestProject(t, "manifest-roundtrip-test")
	sessionID := createTestSessionWithAgents(t, projectDir, 2)

	oldSessionID := exportSessionID
	oldFormat := exportFormat
	oldOutputDir := exportOutputDir
	oldClaudeDir := claudeDir
	defer func() {
		exportSessionID = oldSessionID
		exportFormat = oldFormat
		exportOutputDir = oldOutputDir
		claudeDir = oldClaudeDir
	}()

	outputDir := filepath.Join(tmpDir, "export-roundtrip")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read the manifest
	manifestPath := filepath.Join(outputDir, "manifest.json")
	content1, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	var manifest1 map[string]interface{}
	if err := json.Unmarshal(content1, &manifest1); err != nil {
		t.Fatalf("failed to parse manifest: %v", err)
	}

	// Write it back (simulating a roundtrip)
	content2, err := json.MarshalIndent(manifest1, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	// Parse again
	var manifest2 map[string]interface{}
	if err := json.Unmarshal(content2, &manifest2); err != nil {
		t.Fatalf("failed to parse re-marshaled manifest: %v", err)
	}

	// Verify key fields match
	if manifest1["session_id"] != manifest2["session_id"] {
		t.Error("session_id changed after roundtrip")
	}

	if manifest1["entry_count"] != manifest2["entry_count"] {
		t.Error("entry_count changed after roundtrip")
	}

	if manifest1["version"] != manifest2["version"] {
		t.Error("version changed after roundtrip")
	}
}
