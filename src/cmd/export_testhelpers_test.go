package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/randlee/claude-history/pkg/encoding"
)

// createTestSessionWithAgents creates a realistic test session with the specified number of agents.
// Returns the project directory path, the session ID, and the Claude directory.
func createTestSessionWithAgents(t *testing.T, projectDir string, agentCount int) string {
	t.Helper()

	sessionID := "12345678-1234-1234-1234-123456789abc"

	// Create session content with realistic entries
	// Note: message field contains JSON structure directly (not as a string)
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"Create a test application"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":[{"type":"text","text":"I'll help you create a test application."},{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"mkdir test-app"}}]}
{"type":"user","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"entry-3","message":[{"type":"tool_result","tool_use_id":"tool-1","content":""}]}
`, sessionID, sessionID, sessionID)

	// Add queue-operation entries for spawning agents
	for i := 0; i < agentCount; i++ {
		agentID := fmt.Sprintf("agent-%d", i+1)
		queueEntry := fmt.Sprintf(`{"type":"queue-operation","timestamp":"2026-02-01T10:%02d:00Z","sessionId":"%s","uuid":"queue-%d","agentId":"%s"}
`, i+1, sessionID, i+1, agentID)
		sessionContent += queueEntry
	}

	// Create main session file
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session file: %v", err)
	}

	// Create agent files
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents directory: %v", err)
	}

	for i := 0; i < agentCount; i++ {
		agentID := fmt.Sprintf("agent-%d", i+1)
		agentContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:%02d:05Z","sessionId":"%s","uuid":"%s-entry-1","message":"Task prompt for agent %d"}
{"type":"assistant","timestamp":"2026-02-01T10:%02d:10Z","sessionId":"%s","uuid":"%s-entry-2","message":"Agent %d response"}
`, i+1, sessionID, agentID, i+1, i+1, sessionID, agentID, i+1)

		agentFile := filepath.Join(subagentsDir, fmt.Sprintf("agent-%s.jsonl", agentID))
		if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
			t.Fatalf("failed to create agent file: %v", err)
		}
	}

	return sessionID
}

// createNestedAgentStructure creates a more complex nested agent structure.
// Creates a hierarchy: main session -> agent-parent -> agent-child-1 and agent-child-2.
func createNestedAgentStructure(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")

	// Create parent agent
	parentContent := `{"type":"user","timestamp":"2026-02-01T11:00:00Z","uuid":"parent-entry-1","message":"Parent agent task"}
{"type":"assistant","timestamp":"2026-02-01T11:00:05Z","uuid":"parent-entry-2","message":"Parent agent response"}
{"type":"queue-operation","timestamp":"2026-02-01T11:01:00Z","uuid":"parent-queue-1","agentId":"child-1"}
{"type":"queue-operation","timestamp":"2026-02-01T11:02:00Z","uuid":"parent-queue-2","agentId":"child-2"}
`
	parentFile := filepath.Join(subagentsDir, "agent-parent.jsonl")
	if err := os.WriteFile(parentFile, []byte(parentContent), 0644); err != nil {
		t.Fatalf("failed to create parent agent file: %v", err)
	}

	// Create parent agent directory with nested subagents
	parentAgentDir := filepath.Join(subagentsDir, "agent-parent")
	nestedSubagentsDir := filepath.Join(parentAgentDir, "subagents")
	if err := os.MkdirAll(nestedSubagentsDir, 0755); err != nil {
		t.Fatalf("failed to create nested subagents directory: %v", err)
	}

	// Create child agents
	for i := 1; i <= 2; i++ {
		childID := fmt.Sprintf("child-%d", i)
		childContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T11:%02d:00Z","uuid":"%s-entry-1","message":"Child agent %d task"}
{"type":"assistant","timestamp":"2026-02-01T11:%02d:05Z","uuid":"%s-entry-2","message":"Child agent %d response"}
`, i+2, childID, i, i+2, childID, i)

		childFile := filepath.Join(nestedSubagentsDir, fmt.Sprintf("agent-%s.jsonl", childID))
		if err := os.WriteFile(childFile, []byte(childContent), 0644); err != nil {
			t.Fatalf("failed to create child agent file: %v", err)
		}
	}
}

// verifyHTMLOutput checks that all expected files are present in the output directory.
func verifyHTMLOutput(t *testing.T, outputDir string, expectedAgents int) {
	t.Helper()

	// Check index.html exists
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("index.html not found in %s", outputDir)
	}

	// Check source directory exists
	sourceDir := filepath.Join(outputDir, "source")
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		t.Errorf("source directory not found in %s", outputDir)
	}

	// Check manifest.json exists
	manifestPath := filepath.Join(outputDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("manifest.json not found in %s", outputDir)
	}

	// Check static assets exist
	staticDir := filepath.Join(outputDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		t.Errorf("static directory not found in %s", outputDir)
	}

	cssPath := filepath.Join(staticDir, "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Errorf("style.css not found in %s", staticDir)
	}

	jsPath := filepath.Join(staticDir, "script.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Errorf("script.js not found in %s", staticDir)
	}

	// Check agents directory exists if agents expected
	if expectedAgents > 0 {
		agentsDir := filepath.Join(outputDir, "agents")
		if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
			t.Errorf("agents directory not found in %s", outputDir)
		}

		// Count agent HTML files
		entries, err := os.ReadDir(agentsDir)
		if err != nil {
			t.Errorf("failed to read agents directory: %v", err)
			return
		}

		htmlCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
				htmlCount++
			}
		}

		if htmlCount != expectedAgents {
			t.Errorf("expected %d agent HTML files, found %d", expectedAgents, htmlCount)
		}
	}
}

// parseHTML parses and validates HTML structure.
func parseHTML(t *testing.T, htmlPath string) *html.Node {
	t.Helper()

	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read HTML file %s: %v", htmlPath, err)
	}

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		t.Fatalf("failed to parse HTML from %s: %v", htmlPath, err)
	}

	return doc
}

// findHTMLNode searches for an HTML node by tag name.
func findHTMLNode(node *html.Node, tagName string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && node.Data == tagName {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findHTMLNode(child, tagName); result != nil {
			return result
		}
	}

	return nil
}

// findAllHTMLNodes finds all HTML nodes matching a tag name.
func findAllHTMLNodes(node *html.Node, tagName string, results *[]*html.Node) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode && node.Data == tagName {
		*results = append(*results, node)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		findAllHTMLNodes(child, tagName, results)
	}
}

// hasClass checks if an HTML node has a specific CSS class.
func hasClass(node *html.Node, className string) bool {
	if node == nil {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == className {
					return true
				}
			}
		}
	}

	return false
}

// getAttr gets an attribute value from an HTML node.
func getAttr(node *html.Node, attrName string) string {
	if node == nil {
		return ""
	}

	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

// setupTestProject creates a complete test project structure.
// Returns the temp directory, the project directory, and the encoded project path.
func setupTestProject(t *testing.T, projectName string) (string, string, string) {
	t.Helper()

	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")

	// Create a project path
	projectPath := filepath.Join(tmpDir, projectName)
	encodedName := encoding.EncodePath(projectPath)
	encodedProjectDir := filepath.Join(projectsDir, encodedName)

	if err := os.MkdirAll(encodedProjectDir, 0755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}

	return tmpDir, encodedProjectDir, projectPath
}

// verifyManifestValid checks that a manifest file is valid JSON and contains expected fields.
func verifyManifestValid(t *testing.T, manifestPath string) map[string]interface{} {
	t.Helper()

	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatalf("manifest is not valid JSON: %v", err)
	}

	// Check required fields
	requiredFields := []string{"version", "exported_at", "session_id", "entry_count", "agent_tree", "source_files"}
	for _, field := range requiredFields {
		if _, ok := manifest[field]; !ok {
			t.Errorf("manifest missing required field: %s", field)
		}
	}

	return manifest
}

// createSessionWithXSS creates a test session with potential XSS content.
func createSessionWithXSS(t *testing.T, projectDir string) string {
	t.Helper()

	sessionID := "xsstest0-1234-5678-9abc-def012345678"

	// Create session with malicious content
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"xss-1","message":"<script>alert('XSS')</script>"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"xss-2","message":"Response with <b>HTML</b> and <img src=x onerror=alert('XSS')>"}
{"type":"user","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"xss-3","message":"Content with special chars: & < > \" '"}
`, sessionID, sessionID, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create XSS test session: %v", err)
	}

	return sessionID
}

// createEmptySession creates a session with minimal content (only metadata entries).
func createEmptySession(t *testing.T, projectDir string) string {
	t.Helper()

	sessionID := "emptyses-1234-5678-9abc-def012345678"

	// Create session with only non-message entries
	sessionContent := fmt.Sprintf(`{"type":"file-history-snapshot","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"snap-1"}
`, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create empty session: %v", err)
	}

	return sessionID
}

// createLargeSession creates a session with many entries for performance testing.
func createLargeSession(t *testing.T, projectDir string, entryCount int) string {
	t.Helper()

	sessionID := "largeses-1234-5678-9abc-def012345678"

	var sb strings.Builder
	for i := 0; i < entryCount; i++ {
		entryType := "user"
		if i%2 == 1 {
			entryType = "assistant"
		}

		entry := fmt.Sprintf(`{"type":"%s","timestamp":"2026-02-01T10:%02d:%02d.%03dZ","sessionId":"%s","uuid":"entry-%d","message":"Message %d"}
`, entryType, (i/3600)%24, (i/60)%60, i%60, sessionID, i, i)
		sb.WriteString(entry)
	}

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sb.String()), 0644); err != nil {
		t.Fatalf("failed to create large session: %v", err)
	}

	return sessionID
}
