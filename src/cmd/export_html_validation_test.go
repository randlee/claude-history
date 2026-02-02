package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHTML_ValidStructure(t *testing.T) {
	// Create test project and export
	tmpDir, projectDir, projectPath := setupTestProject(t, "html-structure-test")
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

	outputDir := filepath.Join(tmpDir, "export-structure")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse and validate HTML
	indexPath := filepath.Join(outputDir, "index.html")
	doc := parseHTML(t, indexPath)

	// Verify DOCTYPE (it should be present in the raw content)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	if !strings.Contains(string(content), "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE declaration")
	}

	// Verify html tag
	htmlNode := findHTMLNode(doc, "html")
	if htmlNode == nil {
		t.Error("HTML document missing <html> tag")
	}

	// Verify head tag
	headNode := findHTMLNode(doc, "head")
	if headNode == nil {
		t.Fatal("HTML document missing <head> tag")
	}

	// Verify meta charset
	var metaNodes []*html.Node
	findAllHTMLNodes(headNode, "meta", &metaNodes)

	hasCharset := false
	for _, meta := range metaNodes {
		if getAttr(meta, "charset") == "UTF-8" {
			hasCharset = true
			break
		}
	}

	if !hasCharset {
		t.Error("HTML missing <meta charset='UTF-8'>")
	}

	// Verify title tag
	titleNode := findHTMLNode(headNode, "title")
	if titleNode == nil {
		t.Error("HTML missing <title> tag")
	}

	// Verify link to CSS
	var linkNodes []*html.Node
	findAllHTMLNodes(headNode, "link", &linkNodes)

	hasCSS := false
	for _, link := range linkNodes {
		if getAttr(link, "rel") == "stylesheet" {
			hasCSS = true
			break
		}
	}

	if !hasCSS {
		t.Error("HTML missing stylesheet link")
	}

	// Verify body tag
	bodyNode := findHTMLNode(doc, "body")
	if bodyNode == nil {
		t.Error("HTML document missing <body> tag")
	}

	// Verify script tag
	scriptNode := findHTMLNode(bodyNode, "script")
	if scriptNode == nil {
		t.Error("HTML missing <script> tag")
	}
}

func TestHTML_XSSPrevention(t *testing.T) {
	// Create test project with XSS content
	tmpDir, projectDir, projectPath := setupTestProject(t, "xss-test")
	sessionID := createSessionWithXSS(t, projectDir)

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

	outputDir := filepath.Join(tmpDir, "export-xss")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read the HTML output
	indexPath := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	htmlContent := string(content)

	// Verify script tags are escaped
	if strings.Contains(htmlContent, "<script>alert") {
		t.Error("XSS: script tag not properly escaped")
	}

	// Verify img tag with onerror is escaped (the dangerous pattern is <img...onerror=)
	if strings.Contains(htmlContent, "<img") && strings.Contains(htmlContent, " onerror=") {
		// Check if it's really unescaped (not &lt;img)
		if strings.Contains(htmlContent, "<img src=") || strings.Contains(htmlContent, "<img ") {
			t.Error("XSS: img tag with onerror not properly escaped")
		}
	}

	// Verify escaped versions are present
	if !strings.Contains(htmlContent, "&lt;script&gt;") {
		t.Error("XSS content should be HTML entity encoded")
	}

	// Verify other HTML tags are escaped
	if strings.Contains(htmlContent, "<img src=x") {
		t.Error("XSS: img tag not properly escaped")
	}
}

func TestHTML_SpecialCharacters(t *testing.T) {
	// Create test project with special characters
	tmpDir, projectDir, projectPath := setupTestProject(t, "special-chars-test")

	sessionID := "special0-1234-5678-9abc-def012345678"
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"Test with < > & \" ' chars"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":"Response with unicode: \u00e9\u00e8\u00e0 \u4e2d\u6587"}
`, sessionID, sessionID)
	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-chars")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read HTML
	indexPath := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	htmlContent := string(content)

	// Verify special chars are escaped
	if !strings.Contains(htmlContent, "&lt;") {
		t.Error("< should be escaped as &lt;")
	}

	if !strings.Contains(htmlContent, "&gt;") {
		t.Error("> should be escaped as &gt;")
	}

	if !strings.Contains(htmlContent, "&amp;") {
		t.Error("& should be escaped as &amp;")
	}

	if !strings.Contains(htmlContent, "&#34;") && !strings.Contains(htmlContent, "&quot;") {
		t.Error("\" should be escaped")
	}

	// Unicode should be preserved
	if !strings.Contains(htmlContent, "éèà") && !strings.Contains(htmlContent, "\\u00") {
		t.Error("unicode characters should be preserved or escaped")
	}
}

func TestHTML_LongToolInputs(t *testing.T) {
	// Create test with very long tool input
	tmpDir, projectDir, projectPath := setupTestProject(t, "long-tool-test")

	sessionID := "longtool-1234-5678-9abc-def012345678"

	// Create a very long command string
	longCommand := strings.Repeat("echo 'test'; ", 500) // ~5000 chars

	// Create properly formatted message fields (JSON structure directly, not as string)
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"Run a long command"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":[{"type":"text","text":"Running command"},{"type":"tool_use","id":"tool-1","name":"Bash","input":{"command":"%s"}}]}
{"type":"user","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"entry-3","message":[{"type":"tool_result","tool_use_id":"tool-1","content":"Success"}]}
`, sessionID, sessionID, longCommand, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-long")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify HTML is valid despite long content
	indexPath := filepath.Join(outputDir, "index.html")
	doc := parseHTML(t, indexPath)

	if doc == nil {
		t.Error("HTML should be valid even with very long tool inputs")
	}

	// Verify the content is present (even if truncated in display)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	if !strings.Contains(string(content), "echo") {
		t.Error("long tool input should be included in HTML")
	}
}

func TestHTML_ExpandableSections(t *testing.T) {
	// Create test with tool calls
	tmpDir, projectDir, projectPath := setupTestProject(t, "expandable-test")

	sessionID := "expand00-1234-5678-9abc-def012345678"

	// Create properly formatted message fields (JSON structure directly, not as string)
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"Read a file"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":[{"type":"text","text":"Reading file"},{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/test/file.txt"}}]}
{"type":"user","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"entry-3","message":[{"type":"tool_result","tool_use_id":"tool-1","content":"File contents here"}]}
{"type":"queue-operation","timestamp":"2026-02-01T10:01:00Z","sessionId":"%s","uuid":"queue-1","agentId":"test-agent"}
`, sessionID, sessionID, sessionID, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create agent file
	sessionDir := filepath.Join(projectDir, sessionID)
	subagentsDir := filepath.Join(sessionDir, "subagents")
	if err := os.MkdirAll(subagentsDir, 0755); err != nil {
		t.Fatalf("failed to create subagents dir: %v", err)
	}

	agentContent := `{"type":"user","timestamp":"2026-02-01T10:01:05Z","uuid":"agent-1","message":"Agent task"}
`
	agentFile := filepath.Join(subagentsDir, "agent-test-agent.jsonl")
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatalf("failed to create agent file: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-expandable")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read HTML and verify expandable sections
	indexPath := filepath.Join(outputDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}

	htmlContent := string(content)

	// Verify tool call has expandable class/attributes
	if !strings.Contains(htmlContent, "tool-call") {
		t.Error("HTML should contain tool-call class")
	}

	if !strings.Contains(htmlContent, "tool-header") {
		t.Error("HTML should contain tool-header for expansion")
	}

	if !strings.Contains(htmlContent, "tool-body") {
		t.Error("HTML should contain tool-body for content")
	}

	// Verify data attributes for JavaScript
	if !strings.Contains(htmlContent, "data-tool-id") {
		t.Error("HTML should contain data-tool-id attribute")
	}

	// Verify subagent has expandable section
	if !strings.Contains(htmlContent, "subagent") {
		t.Error("HTML should contain subagent class")
	}

	if !strings.Contains(htmlContent, "data-agent-id") {
		t.Error("HTML should contain data-agent-id attribute")
	}

	// Verify onclick handlers for interactivity
	if !strings.Contains(htmlContent, "onclick") {
		t.Error("HTML should contain onclick handlers for expansion")
	}
}

func TestHTML_MultipleEntryTypes(t *testing.T) {
	// Create test with various entry types
	tmpDir, projectDir, projectPath := setupTestProject(t, "entry-types-test")

	sessionID := "entrytpe-1234-5678-9abc-def012345678"
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":"User message"}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":"Assistant message"}
{"type":"system","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"entry-3","message":"System message"}
{"type":"summary","timestamp":"2026-02-01T10:00:15Z","sessionId":"%s","uuid":"entry-4","summary":"Session summary"}
{"type":"file-history-snapshot","timestamp":"2026-02-01T10:00:20Z","sessionId":"%s","uuid":"entry-5"}
{"type":"progress","timestamp":"2026-02-01T10:00:25Z","sessionId":"%s","uuid":"entry-6","message":"Progress update"}
`, sessionID, sessionID, sessionID, sessionID, sessionID, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-types")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Parse HTML and verify all entry types rendered
	indexPath := filepath.Join(outputDir, "index.html")
	doc := parseHTML(t, indexPath)

	// Find all entries
	var entryNodes []*html.Node
	findAllHTMLNodes(doc, "div", &entryNodes)

	// Count entries with different classes
	hasUser := false
	hasAssistant := false
	hasSystem := false

	for _, node := range entryNodes {
		if hasClass(node, "user") {
			hasUser = true
		}
		if hasClass(node, "assistant") {
			hasAssistant = true
		}
		if hasClass(node, "system") {
			hasSystem = true
		}
	}

	if !hasUser {
		t.Error("HTML should contain user entry")
	}
	if !hasAssistant {
		t.Error("HTML should contain assistant entry")
	}
	if !hasSystem {
		t.Error("HTML should contain system entry")
	}
}

func TestHTML_EmptyContentHandling(t *testing.T) {
	// Test entries with missing or empty content
	tmpDir, projectDir, projectPath := setupTestProject(t, "empty-content-test")

	sessionID := "emptyc00-1234-5678-9abc-def012345678"
	sessionContent := fmt.Sprintf(`{"type":"user","timestamp":"2026-02-01T10:00:00Z","sessionId":"%s","uuid":"entry-1","message":""}
{"type":"assistant","timestamp":"2026-02-01T10:00:05Z","sessionId":"%s","uuid":"entry-2","message":""}
{"type":"system","timestamp":"2026-02-01T10:00:10Z","sessionId":"%s","uuid":"entry-3"}
`, sessionID, sessionID, sessionID)

	sessionFile := filepath.Join(projectDir, sessionID+".jsonl")
	if err := os.WriteFile(sessionFile, []byte(sessionContent), 0644); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

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

	outputDir := filepath.Join(tmpDir, "export-empty-content")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// HTML should be valid even with empty content
	indexPath := filepath.Join(outputDir, "index.html")
	doc := parseHTML(t, indexPath)

	if doc == nil {
		t.Error("HTML should be valid even with empty content entries")
	}
}

func TestHTML_CSSAndJSPresent(t *testing.T) {
	// Verify CSS and JavaScript files are created
	tmpDir, projectDir, projectPath := setupTestProject(t, "assets-test")
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

	outputDir := filepath.Join(tmpDir, "export-assets")

	exportSessionID = sessionID
	exportFormat = "html"
	exportOutputDir = outputDir
	claudeDir = tmpDir

	if err := runExport(exportCmd, []string{projectPath}); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify CSS file exists and has content
	cssPath := filepath.Join(outputDir, "static", "style.css")
	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("CSS file not found: %v", err)
	}

	if len(cssContent) == 0 {
		t.Error("CSS file should not be empty")
	}

	// Verify it looks like CSS
	if !strings.Contains(string(cssContent), "{") || !strings.Contains(string(cssContent), "}") {
		t.Error("CSS file should contain valid CSS syntax")
	}

	// Verify JS file exists and has content
	jsPath := filepath.Join(outputDir, "static", "script.js")
	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("JavaScript file not found: %v", err)
	}

	if len(jsContent) == 0 {
		t.Error("JavaScript file should not be empty")
	}

	// Verify it looks like JavaScript (has function keyword)
	if !strings.Contains(string(jsContent), "function") {
		t.Error("JavaScript file should contain function definitions")
	}
}
