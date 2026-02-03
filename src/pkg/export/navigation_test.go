package export

import (
	"strings"
	"testing"
)

func TestRenderBreadcrumbs_Default(t *testing.T) {
	html := RenderBreadcrumbs(nil)

	// Should have nav element with proper attributes
	if !strings.Contains(html, `<nav class="breadcrumbs"`) {
		t.Error("Missing breadcrumbs nav element")
	}
	if !strings.Contains(html, `id="breadcrumbs"`) {
		t.Error("Missing breadcrumbs id attribute")
	}
	if !strings.Contains(html, `aria-label="Navigation breadcrumbs"`) {
		t.Error("Missing aria-label attribute")
	}

	// Should have default main session item
	if !strings.Contains(html, "Main Session") {
		t.Error("Missing default Main Session breadcrumb")
	}
	if !strings.Contains(html, `data-agent-id="main"`) {
		t.Error("Missing main agent id")
	}
	if !strings.Contains(html, `aria-current="page"`) {
		t.Error("Missing aria-current for active breadcrumb")
	}
}

func TestRenderBreadcrumbs_Empty(t *testing.T) {
	html := RenderBreadcrumbs([]BreadcrumbItem{})

	// Should fall back to default
	if !strings.Contains(html, "Main Session") {
		t.Error("Empty items should show default Main Session")
	}
}

func TestRenderBreadcrumbs_SingleItem(t *testing.T) {
	items := []BreadcrumbItem{
		{ID: "main", Label: "Main Session"},
	}

	html := RenderBreadcrumbs(items)

	// Should have only active item, no separator
	if !strings.Contains(html, `class="breadcrumb-item active"`) {
		t.Error("Missing active class on single item")
	}
	if strings.Contains(html, `class="breadcrumb-separator"`) {
		t.Error("Should not have separator with single item")
	}
}

func TestRenderBreadcrumbs_MultiplItems(t *testing.T) {
	items := []BreadcrumbItem{
		{ID: "main", Label: "Main Session"},
		{ID: "a12eb64", Label: "a12eb64"},
		{ID: "b34fc89", Label: "b34fc89"},
	}

	html := RenderBreadcrumbs(items)

	// Check all items are present
	if !strings.Contains(html, "Main Session") {
		t.Error("Missing Main Session breadcrumb")
	}
	if !strings.Contains(html, `data-agent-id="a12eb64"`) {
		t.Error("Missing first agent breadcrumb")
	}
	if !strings.Contains(html, `data-agent-id="b34fc89"`) {
		t.Error("Missing second agent breadcrumb")
	}

	// Check separators
	separatorCount := strings.Count(html, `class="breadcrumb-separator"`)
	if separatorCount != 2 {
		t.Errorf("Expected 2 separators, got %d", separatorCount)
	}

	// Check only last item is active
	activeCount := strings.Count(html, `class="breadcrumb-item active"`)
	if activeCount != 1 {
		t.Errorf("Expected 1 active item, got %d", activeCount)
	}

	// Last item should be active
	if !strings.Contains(html, `data-agent-id="b34fc89" aria-current="page"`) {
		t.Error("Last item should have aria-current attribute")
	}
}

func TestRenderBreadcrumbs_XSSPrevention(t *testing.T) {
	items := []BreadcrumbItem{
		{ID: "<script>alert('xss')</script>", Label: "<img src=x onerror=alert(1)>"},
	}

	html := RenderBreadcrumbs(items)

	// Should escape HTML
	if strings.Contains(html, "<script>") {
		t.Error("Unescaped script tag - XSS vulnerability")
	}
	if strings.Contains(html, "<img") {
		t.Error("Unescaped img tag - XSS vulnerability")
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Script tag should be escaped")
	}
}

func TestRenderJumpToParentButton_ValidAgent(t *testing.T) {
	html := RenderJumpToParentButton("a12eb64abc123")

	// Check structure
	if !strings.Contains(html, `class="jump-to-parent-btn"`) {
		t.Error("Missing jump-to-parent-btn class")
	}
	if !strings.Contains(html, `data-agent-id="a12eb64abc123"`) {
		t.Error("Missing data-agent-id attribute")
	}
	if !strings.Contains(html, `type="button"`) {
		t.Error("Missing type attribute")
	}
	if !strings.Contains(html, `title="Jump to parent agent`) {
		t.Error("Missing title attribute")
	}
	if !strings.Contains(html, "Parent") {
		t.Error("Missing button text")
	}
}

func TestRenderJumpToParentButton_EmptyID(t *testing.T) {
	html := RenderJumpToParentButton("")

	if html != "" {
		t.Error("Empty agent ID should return empty string")
	}
}

func TestRenderJumpToParentButton_MainSession(t *testing.T) {
	html := RenderJumpToParentButton("main")

	if html != "" {
		t.Error("Main session should return empty string (no parent)")
	}
}

func TestRenderJumpToParentButton_XSSPrevention(t *testing.T) {
	html := RenderJumpToParentButton("<script>alert('xss')</script>")

	if strings.Contains(html, "<script>") {
		t.Error("Unescaped script tag - XSS vulnerability")
	}
}

func TestRenderAgentContainer_MainSession(t *testing.T) {
	html := RenderAgentContainer("main", 0, "<p>Content</p>")

	// Check structure
	if !strings.Contains(html, `class="agent-container"`) {
		t.Error("Missing agent-container class")
	}
	if !strings.Contains(html, `id="agent-main"`) {
		t.Error("Missing agent id")
	}
	if !strings.Contains(html, "<p>Content</p>") {
		t.Error("Missing inner content")
	}

	// Should NOT have depth attribute for main
	if strings.Contains(html, `data-depth=`) {
		t.Error("Main session should not have depth attribute")
	}

	// Should NOT have jump to parent for main
	if strings.Contains(html, "jump-to-parent-btn") {
		t.Error("Main session should not have jump to parent button")
	}
}

func TestRenderAgentContainer_NestedAgent(t *testing.T) {
	html := RenderAgentContainer("a12eb64", 1, "<p>Nested content</p>")

	// Check structure
	if !strings.Contains(html, `id="agent-a12eb64"`) {
		t.Error("Missing agent id")
	}
	if !strings.Contains(html, `data-depth="1"`) {
		t.Error("Missing depth attribute")
	}

	// Should have jump to parent button
	if !strings.Contains(html, "jump-to-parent-btn") {
		t.Error("Nested agent should have jump to parent button")
	}

	// Check content
	if !strings.Contains(html, "<p>Nested content</p>") {
		t.Error("Missing inner content")
	}
}

func TestRenderAgentContainer_DeepNesting(t *testing.T) {
	html := RenderAgentContainer("deep-agent", 3, "<p>Deep</p>")

	if !strings.Contains(html, `data-depth="3"`) {
		t.Error("Missing depth=3 attribute")
	}
}

func TestRenderAgentContainer_XSSPrevention(t *testing.T) {
	html := RenderAgentContainer("<script>xss</script>", 1, "content")

	if strings.Contains(html, "<script>xss") {
		t.Error("Unescaped script tag in agent ID - XSS vulnerability")
	}
}

func TestRenderNestedSubagentOverlay_BasicStructure(t *testing.T) {
	html := RenderNestedSubagentOverlay("a12eb64abc123def456", 29, 1, nil)

	// Check structure
	if !strings.Contains(html, `class="subagent-overlay agent-overlay collapsible"`) {
		t.Error("Missing subagent-overlay classes")
	}
	if !strings.Contains(html, `data-agent-id="a12eb64abc123def456"`) {
		t.Error("Missing data-agent-id attribute")
	}
	if !strings.Contains(html, `id="agent-a12eb64abc123def456"`) {
		t.Error("Missing id attribute for navigation")
	}
	if !strings.Contains(html, `data-depth="1"`) {
		t.Error("Missing depth attribute")
	}

	// Check header elements
	if !strings.Contains(html, `class="agent-icon"`) {
		t.Error("Missing agent-icon class")
	}
	if !strings.Contains(html, "\xf0\x9f\xa4\x96") { // robot emoji
		t.Error("Missing robot icon")
	}
	if !strings.Contains(html, "Subagent: a12eb64") {
		t.Error("Missing truncated agent ID in title")
	}
	if !strings.Contains(html, "(29 entries)") {
		t.Error("Missing entry count")
	}

	// Check navigation controls
	if !strings.Contains(html, `class="agent-header-controls"`) {
		t.Error("Missing agent-header-controls")
	}
	if !strings.Contains(html, `class="deep-dive-btn"`) {
		t.Error("Missing deep-dive-btn class")
	}
	if !strings.Contains(html, "Deep Dive") {
		t.Error("Missing Deep Dive button text")
	}
	if !strings.Contains(html, `class="jump-to-parent-btn"`) {
		t.Error("Missing jump-to-parent-btn for nested agent")
	}

	// Check copy button
	if !strings.Contains(html, `data-copy-text="a12eb64abc123def456"`) {
		t.Error("Missing full agent ID in copy button")
	}

	// Check content container
	if !strings.Contains(html, `class="subagent-content collapsible-content collapsed"`) {
		t.Error("Missing subagent-content classes")
	}
}

func TestRenderNestedSubagentOverlay_NoJumpToParentAtDepthZero(t *testing.T) {
	html := RenderNestedSubagentOverlay("agent123", 10, 0, nil)

	// Should NOT have jump to parent at depth 0
	if strings.Contains(html, `class="jump-to-parent-btn"`) {
		t.Error("Depth 0 should not have jump to parent button")
	}

	// Should still have deep dive
	if !strings.Contains(html, `class="deep-dive-btn"`) {
		t.Error("Should have deep dive button")
	}
}

func TestRenderNestedSubagentOverlay_WithMetadata(t *testing.T) {
	metadata := map[string]string{
		"Session":  "fbd51e2b",
		"Duration": "5m 32s",
	}

	html := RenderNestedSubagentOverlay("agent123", 10, 1, metadata)

	if !strings.Contains(html, `class="subagent-metadata"`) {
		t.Error("Missing subagent-metadata class")
	}
	if !strings.Contains(html, "Session:") {
		t.Error("Missing Session metadata key")
	}
	if !strings.Contains(html, "fbd51e2b") {
		t.Error("Missing Session metadata value")
	}
}

func TestRenderNestedSubagentOverlay_NoDepthAttribute(t *testing.T) {
	html := RenderNestedSubagentOverlay("agent123", 10, 0, nil)

	// Should NOT have depth attribute when depth is 0
	if strings.Contains(html, `data-depth="0"`) {
		t.Error("Should not have data-depth when depth is 0")
	}
}

func TestRenderNestedSubagentOverlay_ShortAgentID(t *testing.T) {
	html := RenderNestedSubagentOverlay("abc", 5, 1, nil)

	// Short IDs should not be truncated
	if !strings.Contains(html, "Subagent: abc") {
		t.Error("Short agent ID should not be truncated")
	}
}

func TestRenderNestedSubagentOverlay_XSSPrevention(t *testing.T) {
	metadata := map[string]string{
		"<script>": "<script>alert('xss')</script>",
	}

	html := RenderNestedSubagentOverlay("<script>evil</script>", 10, 1, metadata)

	if strings.Contains(html, "<script>") {
		t.Error("Unescaped script tag - XSS vulnerability")
	}
}

func TestGenerateBreadcrumbPath_Empty(t *testing.T) {
	path := GenerateBreadcrumbPath(nil)

	if len(path) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(path))
	}
	if path[0].ID != "main" {
		t.Error("First item should be main")
	}
	if path[0].Label != "Main Session" {
		t.Error("First item label should be Main Session")
	}
}

func TestGenerateBreadcrumbPath_WithAgents(t *testing.T) {
	path := GenerateBreadcrumbPath([]string{"a12eb64abc123", "b34fc89def456"})

	if len(path) != 3 {
		t.Fatalf("Expected 3 items (main + 2 agents), got %d", len(path))
	}

	// First should be main
	if path[0].ID != "main" {
		t.Error("First item should be main")
	}

	// Second should be first agent (truncated)
	if path[1].ID != "a12eb64abc123" {
		t.Error("Second item ID should be full agent ID")
	}
	if path[1].Label != "a12eb64" {
		t.Errorf("Second item label should be truncated, got %q", path[1].Label)
	}

	// Third should be second agent
	if path[2].ID != "b34fc89def456" {
		t.Error("Third item ID should be full agent ID")
	}
}

func TestGenerateBreadcrumbPath_SkipsEmptyAndMain(t *testing.T) {
	path := GenerateBreadcrumbPath([]string{"", "main", "valid-agent"})

	if len(path) != 2 {
		t.Fatalf("Expected 2 items (main + valid-agent), got %d", len(path))
	}

	if path[1].ID != "valid-agent" {
		t.Error("Should skip empty and main IDs")
	}
}

func TestGenerateBreadcrumbPath_ShortAgentID(t *testing.T) {
	path := GenerateBreadcrumbPath([]string{"abc"})

	if len(path) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(path))
	}

	// Short ID should not be truncated
	if path[1].Label != "abc" {
		t.Errorf("Short ID should not be truncated, got %q", path[1].Label)
	}
}

func TestTruncateAgentID_Short(t *testing.T) {
	result := TruncateAgentID("abc")

	if result != "abc" {
		t.Errorf("Short ID should not be truncated, got %q", result)
	}
}

func TestTruncateAgentID_Exact7(t *testing.T) {
	result := TruncateAgentID("1234567")

	if result != "1234567" {
		t.Errorf("7 char ID should not be truncated, got %q", result)
	}
}

func TestTruncateAgentID_Long(t *testing.T) {
	result := TruncateAgentID("a12eb64abc123def456")

	if result != "a12eb64" {
		t.Errorf("Long ID should be truncated to 7 chars, got %q", result)
	}
}

func TestTruncateAgentID_Empty(t *testing.T) {
	result := TruncateAgentID("")

	if result != "" {
		t.Errorf("Empty ID should remain empty, got %q", result)
	}
}

func TestBreadcrumbItem_Structure(t *testing.T) {
	item := BreadcrumbItem{
		ID:    "test-id",
		Label: "Test Label",
	}

	if item.ID != "test-id" {
		t.Errorf("ID mismatch: %q", item.ID)
	}
	if item.Label != "Test Label" {
		t.Errorf("Label mismatch: %q", item.Label)
	}
}

// Integration test: full navigation HTML generation
func TestNavigationIntegration(t *testing.T) {
	// Generate breadcrumbs
	breadcrumbs := RenderBreadcrumbs([]BreadcrumbItem{
		{ID: "main", Label: "Main Session"},
		{ID: "agent1", Label: "agent1"},
	})

	// Generate nested overlay
	overlay := RenderNestedSubagentOverlay("agent1-sub", 15, 2, nil)

	// Generate container
	container := RenderAgentContainer("agent1", 1, overlay)

	// Verify all pieces work together
	if !strings.Contains(breadcrumbs, "Main Session") {
		t.Error("Breadcrumbs missing main session")
	}
	if !strings.Contains(container, "jump-to-parent-btn") {
		t.Error("Container missing jump to parent")
	}
	if !strings.Contains(container, "deep-dive-btn") {
		t.Error("Overlay missing deep dive button")
	}
	if !strings.Contains(container, `data-depth="2"`) {
		t.Error("Nested overlay missing depth attribute")
	}
}

// Benchmark tests
func BenchmarkRenderBreadcrumbs(b *testing.B) {
	items := []BreadcrumbItem{
		{ID: "main", Label: "Main Session"},
		{ID: "agent1", Label: "agent1"},
		{ID: "agent2", Label: "agent2"},
		{ID: "agent3", Label: "agent3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderBreadcrumbs(items)
	}
}

func BenchmarkRenderNestedSubagentOverlay(b *testing.B) {
	metadata := map[string]string{
		"Session":  "fbd51e2b",
		"Duration": "5m 32s",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RenderNestedSubagentOverlay("a12eb64abc123def456", 50, 2, metadata)
	}
}

func BenchmarkGenerateBreadcrumbPath(b *testing.B) {
	agents := []string{
		"a12eb64abc123def456",
		"b34fc89ghi789jkl012",
		"c56mn03pqr345stu678",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateBreadcrumbPath(agents)
	}
}
