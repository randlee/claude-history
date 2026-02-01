package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetStyleCSS(t *testing.T) {
	css := GetStyleCSS()

	if css == "" {
		t.Error("GetStyleCSS returned empty string")
	}

	// Verify it contains expected CSS content
	expectedPatterns := []string{
		".conversation",
		".entry",
		".entry.user",
		".entry.assistant",
		".tool-call",
		".tool-header",
		".tool-body",
		".subagent",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(css, pattern) {
			t.Errorf("CSS missing expected selector: %s", pattern)
		}
	}
}

func TestGetScriptJS(t *testing.T) {
	js := GetScriptJS()

	if js == "" {
		t.Error("GetScriptJS returned empty string")
	}

	// Verify it contains expected JavaScript functions
	expectedFunctions := []string{
		"function toggleTool",
		"function loadAgent",
		"function expandAll",
		"function collapseAll",
	}

	for _, fn := range expectedFunctions {
		if !strings.Contains(js, fn) {
			t.Errorf("JavaScript missing expected function: %s", fn)
		}
	}
}

func TestWriteStaticAssets(t *testing.T) {
	tempDir := t.TempDir()

	err := WriteStaticAssets(tempDir)
	if err != nil {
		t.Fatalf("WriteStaticAssets failed: %v", err)
	}

	// Check static directory was created
	staticDir := filepath.Join(tempDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		t.Error("static directory not created")
	}

	// Check CSS file was written
	cssPath := filepath.Join(staticDir, "style.css")
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		t.Error("style.css not created")
	}

	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read style.css: %v", err)
	}
	if len(cssContent) == 0 {
		t.Error("style.css is empty")
	}

	// Check JS file was written
	jsPath := filepath.Join(staticDir, "script.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Error("script.js not created")
	}

	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read script.js: %v", err)
	}
	if len(jsContent) == 0 {
		t.Error("script.js is empty")
	}
}

func TestWriteStaticAssets_CreatesNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "nested", "output", "path")

	err := WriteStaticAssets(outputDir)
	if err != nil {
		t.Fatalf("WriteStaticAssets failed to create nested directory: %v", err)
	}

	staticDir := filepath.Join(outputDir, "static")
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		t.Error("static directory not created in nested path")
	}
}

func TestWriteStaticAssets_Idempotent(t *testing.T) {
	tempDir := t.TempDir()

	// Write twice - should not error
	if err := WriteStaticAssets(tempDir); err != nil {
		t.Fatalf("First WriteStaticAssets failed: %v", err)
	}

	if err := WriteStaticAssets(tempDir); err != nil {
		t.Fatalf("Second WriteStaticAssets failed: %v", err)
	}

	// Files should still exist and have content
	cssPath := filepath.Join(tempDir, "static", "style.css")
	cssContent, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatalf("Failed to read style.css after double write: %v", err)
	}
	if len(cssContent) == 0 {
		t.Error("style.css is empty after double write")
	}
}

func TestListTemplateFiles(t *testing.T) {
	files, err := ListTemplateFiles()
	if err != nil {
		t.Fatalf("ListTemplateFiles failed: %v", err)
	}

	if len(files) == 0 {
		t.Error("ListTemplateFiles returned empty list")
	}

	// Should contain at least style.css and script.js
	hasCSS := false
	hasJS := false
	for _, f := range files {
		if f == "style.css" {
			hasCSS = true
		}
		if f == "script.js" {
			hasJS = true
		}
	}

	if !hasCSS {
		t.Error("style.css not found in template files")
	}
	if !hasJS {
		t.Error("script.js not found in template files")
	}
}

func TestReadTemplateFile(t *testing.T) {
	testCases := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{"read style.css", "style.css", false},
		{"read script.js", "script.js", false},
		{"read non-existent", "non-existent.txt", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := ReadTemplateFile(tc.file)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(content) == 0 {
					t.Error("Content should not be empty")
				}
			}
		})
	}
}

func TestGetTemplatesFS(t *testing.T) {
	fs := GetTemplatesFS()
	if fs == nil {
		t.Fatal("GetTemplatesFS returned nil")
	}

	// Try to read a file from the FS
	file, err := fs.Open("style.css")
	if err != nil {
		t.Errorf("Failed to open style.css from FS: %v", err)
	} else {
		file.Close()
	}
}

func TestCSSContent_HasResponsiveStyles(t *testing.T) {
	css := GetStyleCSS()

	// Check for media queries
	if !strings.Contains(css, "@media") {
		t.Error("CSS should contain media queries for responsive design")
	}

	// Check for max-width (responsive container)
	if !strings.Contains(css, "max-width") {
		t.Error("CSS should contain max-width for responsive container")
	}
}

func TestCSSContent_HasDarkMode(t *testing.T) {
	css := GetStyleCSS()

	// Check for prefers-color-scheme
	if !strings.Contains(css, "prefers-color-scheme") {
		t.Error("CSS should support dark mode with prefers-color-scheme")
	}
}

func TestCSSContent_HasPrintStyles(t *testing.T) {
	css := GetStyleCSS()

	// Check for print media query
	if !strings.Contains(css, "@media print") {
		t.Error("CSS should contain print styles")
	}
}

func TestJSContent_HasInitFunction(t *testing.T) {
	js := GetScriptJS()

	// Check for init function
	if !strings.Contains(js, "function init") {
		t.Error("JavaScript should contain init function")
	}

	// Check for DOMContentLoaded handling
	if !strings.Contains(js, "DOMContentLoaded") {
		t.Error("JavaScript should handle DOMContentLoaded event")
	}
}

func TestJSContent_HasSearchFunctions(t *testing.T) {
	js := GetScriptJS()

	// Check for search functionality
	searchFunctions := []string{
		"searchEntries",
		"highlightSearch",
		"clearHighlights",
	}

	for _, fn := range searchFunctions {
		if !strings.Contains(js, fn) {
			t.Errorf("JavaScript missing search function: %s", fn)
		}
	}
}

func TestJSContent_HasStatsFunctions(t *testing.T) {
	js := GetScriptJS()

	// Check for getStats function
	if !strings.Contains(js, "getStats") {
		t.Error("JavaScript should contain getStats function")
	}
}

func TestWriteStaticAssets_ContentMatches(t *testing.T) {
	tempDir := t.TempDir()

	err := WriteStaticAssets(tempDir)
	if err != nil {
		t.Fatalf("WriteStaticAssets failed: %v", err)
	}

	// Verify CSS content matches
	cssPath := filepath.Join(tempDir, "static", "style.css")
	writtenCSS, _ := os.ReadFile(cssPath)
	expectedCSS := GetStyleCSS()
	if string(writtenCSS) != expectedCSS {
		t.Error("Written CSS does not match GetStyleCSS() output")
	}

	// Verify JS content matches
	jsPath := filepath.Join(tempDir, "static", "script.js")
	writtenJS, _ := os.ReadFile(jsPath)
	expectedJS := GetScriptJS()
	if string(writtenJS) != expectedJS {
		t.Error("Written JS does not match GetScriptJS() output")
	}
}
