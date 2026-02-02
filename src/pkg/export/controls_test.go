package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
)

func TestGetControlsJS_ReturnsContent(t *testing.T) {
	content := GetControlsJS()

	if content == "" {
		t.Fatal("GetControlsJS() returned empty string")
	}

	// Check for key functions
	requiredFunctions := []string{
		"expandAllTools",
		"collapseAllTools",
		"toggleAllTools",
		"performSearch",
		"clearSearch",
		"nextMatch",
		"prevMatch",
		"smoothScrollToElement",
		"initKeyboardShortcuts",
		"initControlPanel",
		"initControls",
		"loadState",
		"saveState",
	}

	for _, fn := range requiredFunctions {
		if !strings.Contains(content, fn) {
			t.Errorf("controls.js missing required function: %s", fn)
		}
	}
}

func TestGetControlsJS_HasStorageKey(t *testing.T) {
	content := GetControlsJS()

	if !strings.Contains(content, "STORAGE_KEY") {
		t.Error("controls.js should have STORAGE_KEY constant")
	}
	if !strings.Contains(content, "localStorage") {
		t.Error("controls.js should use localStorage for persistence")
	}
}

func TestGetControlsJS_HasKeyboardShortcuts(t *testing.T) {
	content := GetControlsJS()

	shortcuts := []string{
		"Ctrl",
		"ctrlKey",
		"metaKey",
		"Escape",
		"keydown",
	}

	for _, shortcut := range shortcuts {
		if !strings.Contains(content, shortcut) {
			t.Errorf("controls.js missing keyboard shortcut reference: %s", shortcut)
		}
	}
}

func TestGetControlsJS_HasSearchClasses(t *testing.T) {
	content := GetControlsJS()

	classes := []string{
		"SEARCH_HIGHLIGHT_CLASS",
		"SEARCH_MATCH_CLASS",
		"HIDDEN_BY_SEARCH_CLASS",
		"search-highlight",
		"search-match",
		"hidden-by-search",
	}

	for _, class := range classes {
		if !strings.Contains(content, class) {
			t.Errorf("controls.js missing search class reference: %s", class)
		}
	}
}

func TestGetControlsJS_HasPublicAPI(t *testing.T) {
	content := GetControlsJS()

	// Check for ControlsAPI object
	if !strings.Contains(content, "window.ControlsAPI") {
		t.Error("controls.js should expose ControlsAPI globally")
	}

	// Check for API methods
	apiMethods := []string{
		"expandAll:",
		"collapseAll:",
		"toggleAll:",
		"search:",
		"clearSearch:",
		"nextMatch:",
		"prevMatch:",
		"focusSearch:",
		"scrollTo:",
	}

	for _, method := range apiMethods {
		if !strings.Contains(content, method) {
			t.Errorf("controls.js API missing method: %s", method)
		}
	}
}

func TestGetControlsJS_HasDOMContentLoaded(t *testing.T) {
	content := GetControlsJS()

	if !strings.Contains(content, "DOMContentLoaded") {
		t.Error("controls.js should handle DOMContentLoaded event")
	}
	if !strings.Contains(content, "document.readyState") {
		t.Error("controls.js should check document.readyState")
	}
}

func TestHTMLHeader_ContainsControlPanel(t *testing.T) {
	// Verify control panel HTML structure
	requiredElements := []string{
		`class="page-header"`,
		`class="controls"`,
		`id="expand-all-btn"`,
		`id="collapse-all-btn"`,
		`id="search-box"`,
		`class="search-results"`,
		`class="search-container"`,
		`id="search-prev-btn"`,
		`id="search-next-btn"`,
	}

	for _, elem := range requiredElements {
		if !strings.Contains(htmlHeader, elem) {
			t.Errorf("HTML header missing required element: %s", elem)
		}
	}
}

func TestHTMLHeader_HasAccessibilityAttributes(t *testing.T) {
	accessibilityAttrs := []string{
		`role="toolbar"`,
		`aria-label=`,
		`aria-live="polite"`,
		`aria-hidden="true"`,
	}

	for _, attr := range accessibilityAttrs {
		if !strings.Contains(htmlHeader, attr) {
			t.Errorf("HTML header missing accessibility attribute: %s", attr)
		}
	}
}

func TestHTMLHeader_HasKeyboardShortcutHints(t *testing.T) {
	shortcutHints := []string{
		`data-shortcut="Ctrl+K"`,
		`data-shortcut="Ctrl+F"`,
		`title="Expand all tool calls (Ctrl+K)"`,
		`title="Search messages (Ctrl+F)"`,
	}

	for _, hint := range shortcutHints {
		if !strings.Contains(htmlHeader, hint) {
			t.Errorf("HTML header missing keyboard shortcut hint: %s", hint)
		}
	}
}

func TestHTMLFooter_IncludesControlsScript(t *testing.T) {
	if !strings.Contains(htmlFooter, `<script src="static/controls.js"></script>`) {
		t.Error("HTML footer missing controls.js script tag")
	}
}

func TestHTMLFooter_ScriptOrder(t *testing.T) {
	// controls.js should be loaded after script.js and clipboard.js
	scriptIdx := strings.Index(htmlFooter, "script.js")
	clipboardIdx := strings.Index(htmlFooter, "clipboard.js")
	controlsIdx := strings.Index(htmlFooter, "controls.js")

	if scriptIdx == -1 || clipboardIdx == -1 || controlsIdx == -1 {
		t.Fatal("One or more script tags missing from footer")
	}

	if controlsIdx < scriptIdx || controlsIdx < clipboardIdx {
		t.Error("controls.js should be loaded after script.js and clipboard.js")
	}
}

func TestCSSContent_HasControlStyles(t *testing.T) {
	css := GetStyleCSS()

	requiredSelectors := []string{
		".page-header",
		".controls",
		".controls-group",
		".controls-separator",
		"#search-box",
		".search-container",
		".search-results",
		".search-nav-btn",
	}

	for _, selector := range requiredSelectors {
		if !strings.Contains(css, selector) {
			t.Errorf("CSS missing required selector: %s", selector)
		}
	}
}

func TestCSSContent_HasStickyHeader(t *testing.T) {
	css := GetStyleCSS()

	if !strings.Contains(css, "position: sticky") {
		t.Error("CSS should have sticky positioning for header")
	}
	if !strings.Contains(css, "z-index") {
		t.Error("CSS should set z-index for sticky header")
	}
}

func TestCSSContent_HasSearchHighlightStyles(t *testing.T) {
	css := GetStyleCSS()

	highlightSelectors := []string{
		"mark.search-highlight",
		".search-match",
		".search-active",
		".hidden-by-search",
	}

	for _, selector := range highlightSelectors {
		if !strings.Contains(css, selector) {
			t.Errorf("CSS missing search highlight selector: %s", selector)
		}
	}
}

func TestCSSContent_HasKeyboardShortcutHintStyles(t *testing.T) {
	css := GetStyleCSS()

	if !strings.Contains(css, "[data-shortcut]") {
		t.Error("CSS should style elements with data-shortcut attribute")
	}
	if !strings.Contains(css, "::after") {
		t.Error("CSS should use ::after pseudo-element for shortcut hints")
	}
}

func TestCSSContent_HasResponsiveControlStyles(t *testing.T) {
	css := GetStyleCSS()

	// Check for mobile-specific control styles within media query
	if !strings.Contains(css, "@media (max-width: 768px)") {
		t.Error("CSS should have mobile breakpoint")
	}

	// Check that mobile styles exist for controls
	if !strings.Contains(css, "flex-direction: column") {
		t.Error("CSS should have column layout for mobile controls")
	}
}

func TestCSSContent_HasSearchBoxFocusStyles(t *testing.T) {
	css := GetStyleCSS()

	if !strings.Contains(css, "#search-box:focus") {
		t.Error("CSS should have focus styles for search box")
	}
	if !strings.Contains(css, "box-shadow") {
		t.Error("CSS should have box-shadow for focus indication")
	}
}

func TestWriteStaticAssets_IncludesControlsJS(t *testing.T) {
	tempDir := t.TempDir()

	err := WriteStaticAssets(tempDir)
	if err != nil {
		t.Fatalf("WriteStaticAssets failed: %v", err)
	}

	// Check controls.js was written
	controlsPath := filepath.Join(tempDir, "static", "controls.js")
	if _, err := os.Stat(controlsPath); os.IsNotExist(err) {
		t.Error("controls.js not created")
	}

	controlsContent, err := os.ReadFile(controlsPath)
	if err != nil {
		t.Fatalf("Failed to read controls.js: %v", err)
	}

	if len(controlsContent) == 0 {
		t.Error("controls.js is empty")
	}

	// Verify content matches
	expectedContent := GetControlsJS()
	if string(controlsContent) != expectedContent {
		t.Error("Written controls.js does not match GetControlsJS() output")
	}
}

func TestListTemplateFiles_IncludesControlsJS(t *testing.T) {
	files, err := ListTemplateFiles()
	if err != nil {
		t.Fatalf("ListTemplateFiles failed: %v", err)
	}

	hasControls := false
	for _, f := range files {
		if f == "controls.js" {
			hasControls = true
			break
		}
	}

	if !hasControls {
		t.Error("controls.js not found in template files")
	}
}

func TestReadTemplateFile_ControlsJS(t *testing.T) {
	content, err := ReadTemplateFile("controls.js")
	if err != nil {
		t.Fatalf("Failed to read controls.js: %v", err)
	}

	if len(content) == 0 {
		t.Error("controls.js content should not be empty")
	}

	if !strings.Contains(string(content), "initControls") {
		t.Error("controls.js should contain initControls function")
	}
}

func TestRenderConversation_OutputIncludesControlPanel(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Test message"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	// Check for control panel elements
	controlElements := []string{
		`id="expand-all-btn"`,
		`id="collapse-all-btn"`,
		`id="search-box"`,
		`class="page-header"`,
	}

	for _, elem := range controlElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Rendered HTML missing control element: %s", elem)
		}
	}
}

func TestRenderConversation_OutputIncludesControlsScript(t *testing.T) {
	entries := []models.ConversationEntry{
		{
			UUID:      "uuid-001",
			SessionID: "session-001",
			Type:      models.EntryTypeUser,
			Timestamp: "2026-01-31T10:00:00Z",
			Message:   json.RawMessage(`"Test message"`),
		},
	}

	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("RenderConversation() error = %v", err)
	}

	if !strings.Contains(html, `<script src="static/controls.js"></script>`) {
		t.Error("Rendered HTML missing controls.js script")
	}
}

func TestControlPanel_ButtonTypes(t *testing.T) {
	// All buttons should have type="button" to prevent form submission
	if !strings.Contains(htmlHeader, `type="button"`) {
		t.Error("Control panel buttons should have type=\"button\"")
	}

	buttonCount := strings.Count(htmlHeader, "<button")
	typeCount := strings.Count(htmlHeader, `type="button"`)

	if buttonCount != typeCount {
		t.Errorf("All %d buttons should have type=\"button\", but only %d do", buttonCount, typeCount)
	}
}

func TestControlPanel_SearchInputType(t *testing.T) {
	// Search input should be type="search" for native clear button
	if !strings.Contains(htmlHeader, `type="search"`) {
		t.Error("Search input should have type=\"search\"")
	}
}

func TestControlPanel_PlaceholderText(t *testing.T) {
	if !strings.Contains(htmlHeader, `placeholder="Search messages..."`) {
		t.Error("Search input should have placeholder text")
	}
}

func TestGetControlsJS_HasIIFE(t *testing.T) {
	content := GetControlsJS()

	// Should use IIFE pattern to avoid global pollution
	if !strings.Contains(content, "(function()") {
		t.Error("controls.js should use IIFE pattern")
	}
	if !strings.Contains(content, "'use strict'") {
		t.Error("controls.js should use strict mode")
	}
}

func TestGetControlsJS_HasDebounce(t *testing.T) {
	content := GetControlsJS()

	// Search should be debounced
	if !strings.Contains(content, "debounce") || !strings.Contains(content, "setTimeout") {
		t.Error("controls.js should debounce search input")
	}
}

func TestGetControlsJS_HasHighlightFunction(t *testing.T) {
	content := GetControlsJS()

	if !strings.Contains(content, "highlightTextInElement") {
		t.Error("controls.js should have highlightTextInElement function")
	}
	if !strings.Contains(content, "createTreeWalker") {
		t.Error("controls.js should use TreeWalker for text node traversal")
	}
}

func TestGetControlsJS_HandlesMarkElement(t *testing.T) {
	content := GetControlsJS()

	if !strings.Contains(content, "createElement('mark')") {
		t.Error("controls.js should create mark elements for highlighting")
	}
}

func TestCSSContent_HasDarkModeSearchHighlight(t *testing.T) {
	css := GetStyleCSS()

	// Check for dark mode search highlight within prefers-color-scheme
	darkModeStart := strings.Index(css, "@media (prefers-color-scheme: dark)")
	if darkModeStart == -1 {
		t.Fatal("CSS should have dark mode media query")
	}

	darkModeSection := css[darkModeStart:]
	if !strings.Contains(darkModeSection, "search-highlight") {
		t.Error("CSS should have dark mode styles for search-highlight")
	}
}

func TestControlPanel_HasNavigationButtons(t *testing.T) {
	// Check for previous/next navigation buttons
	if !strings.Contains(htmlHeader, `id="search-prev-btn"`) {
		t.Error("Control panel should have previous search button")
	}
	if !strings.Contains(htmlHeader, `id="search-next-btn"`) {
		t.Error("Control panel should have next search button")
	}
	if !strings.Contains(htmlHeader, `class="search-nav-btn"`) {
		t.Error("Navigation buttons should have search-nav-btn class")
	}
}

func TestControlPanel_NavigationButtonTitles(t *testing.T) {
	if !strings.Contains(htmlHeader, `title="Previous match`) {
		t.Error("Previous button should have title attribute")
	}
	if !strings.Contains(htmlHeader, `title="Next match`) {
		t.Error("Next button should have title attribute")
	}
}

func TestGetControlsJS_HasNavigateFunctions(t *testing.T) {
	content := GetControlsJS()

	if !strings.Contains(content, "navigateToMatch") {
		t.Error("controls.js should have navigateToMatch function")
	}
	if !strings.Contains(content, "currentSearchIndex") {
		t.Error("controls.js should track current search index")
	}
	if !strings.Contains(content, "currentMatches") {
		t.Error("controls.js should track current matches")
	}
}

func TestGetControlsJS_WrapAroundNavigation(t *testing.T) {
	content := GetControlsJS()

	// Should handle wrap-around when navigating past end or before start
	if !strings.Contains(content, "currentMatches.length - 1") {
		t.Error("controls.js should wrap to last match when going before first")
	}
}
