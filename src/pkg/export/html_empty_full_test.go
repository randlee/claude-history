package export

import (
	"os"
	"testing"

	"github.com/randlee/claude-history/pkg/session"
)

// TestNoEmptyMessageBubblesFullExport verifies that no empty message bubbles appear
// in the HTML output when using RenderConversation (full export, not query).
func TestNoEmptyMessageBubblesFullExport(t *testing.T) {
	// Load test session data from github-research main session
	sessionFile := "/Users/randlee/.claude/projects/-Users-randlee-Documents-github-github-research/8c43ec84-09ad-4dc7-bcf7-17f209e983f0.jsonl"

	// Skip test if file doesn't exist
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Skipf("Test session file not found: %s", sessionFile)
	}

	// Parse entries from session file
	entries, err := session.ReadSession(sessionFile)
	if err != nil {
		t.Fatalf("Failed to parse session file: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("No entries found in session file")
	}

	t.Logf("Loaded %d entries from main session file", len(entries))

	// Count how many entries we expect to render
	expectedRendered := 0
	for _, entry := range entries {
		if hasContent(entry) {
			expectedRendered++
		}
	}
	t.Logf("Expected %d entries to be rendered (hasContent=true)", expectedRendered)

	// Render HTML using RenderConversation (same as export command)
	html, err := RenderConversation(entries, nil)
	if err != nil {
		t.Fatalf("Failed to render HTML: %v", err)
	}

	// Find all empty message bubbles
	emptyBubbles := findEmptyMessageBubbles(html)

	if len(emptyBubbles) > 0 {
		t.Errorf("Found %d empty message bubbles in full export:", len(emptyBubbles))
		for i, bubble := range emptyBubbles {
			if i >= 5 {
				t.Logf("  ... and %d more", len(emptyBubbles)-5)
				break
			}
			t.Logf("  UUID: %s", bubble.UUID)
			t.Logf("  Role: %s", bubble.Role)
			t.Logf("  Content snippet: %q", bubble.ContentSnippet)

			// Find the corresponding entry
			for _, entry := range entries {
				if entry.UUID == bubble.UUID {
					t.Logf("  Entry type: %s", entry.Type)
					textContent := entry.GetTextContent()
					t.Logf("  Entry text content: %q (len=%d)", textContent, len(textContent))
					t.Logf("  hasContent() returned: %v", hasContent(entry))
					toolResults := entry.ExtractToolResults()
					t.Logf("  Entry has tool results: %d", len(toolResults))
					t.Logf("")
					break
				}
			}
		}
		t.Fail()
	} else {
		t.Logf("✓ No empty message bubbles found in full export HTML output")
	}
}
