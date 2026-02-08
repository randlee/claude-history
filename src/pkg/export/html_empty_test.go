package export

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/randlee/claude-history/pkg/models"
	"github.com/randlee/claude-history/pkg/session"
)

// TestNoEmptyMessageBubbles verifies that no empty message bubbles appear in the HTML output.
// Empty bubbles are message-row divs where the message-content div has no visible content
// (no text, no tool calls, nothing rendered).
func TestNoEmptyMessageBubbles(t *testing.T) {
	// Load test session data from github-research
	sessionFile := "/Users/randlee/.claude/projects/-Users-randlee-Documents-github-github-research/8c43ec84-09ad-4dc7-bcf7-17f209e983f0/subagents/agent-aed3b9d.jsonl"

	// Skip test if file doesn't exist (not all environments have this)
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

	t.Logf("Loaded %d entries from session file", len(entries))

	// Count how many entries we expect to render
	expectedRendered := 0
	for _, entry := range entries {
		if hasContent(entry) {
			expectedRendered++
		}
	}
	t.Logf("Expected %d entries to be rendered (hasContent=true)", expectedRendered)

	// Render HTML using RenderQueryResults (same as query command)
	projectPath := "/Users/randlee/Documents/github/github-research"
	sessionID := "8c43ec84-09ad-4dc7-bcf7-17f209e983f0"
	agentID := "aed3b9d"
	sessionFolderPath := filepath.Dir(sessionFile)

	html, err := RenderQueryResults(entries, projectPath, sessionID, sessionFolderPath, agentID, "Orchestrator", "Agent")
	if err != nil {
		t.Fatalf("Failed to render HTML: %v", err)
	}

	// Find all empty message bubbles
	emptyBubbles := findEmptyMessageBubbles(html)

	if len(emptyBubbles) > 0 {
		t.Errorf("Found %d empty message bubbles:", len(emptyBubbles))
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
					if entry.Type == models.EntryTypeAssistant {
						toolCalls := entry.ExtractToolCalls()
						t.Logf("  Entry has tool calls: %d", len(toolCalls))
					}
					t.Logf("")
					break
				}
			}
		}
		t.Fail()
	} else {
		t.Logf("✓ No empty message bubbles found in HTML output")
	}
}

// EmptyBubble represents an empty message bubble found in HTML
type EmptyBubble struct {
	UUID           string
	Role           string
	ContentSnippet string
}

// findEmptyMessageBubbles parses HTML and finds message-row divs with empty message-content.
func findEmptyMessageBubbles(html string) []EmptyBubble {
	var results []EmptyBubble

	// Pattern: <div class="message-row ... data-uuid="...">
	// We want to match message-row divs and extract their UUID and role
	messageRowPattern := regexp.MustCompile(`(?s)<div class="message-row[^"]*"[^>]*data-uuid="([^"]*)"[^>]*>(.*?)</div>\s*</div>\s*</div>`)

	matches := messageRowPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		uuid := match[1]
		rowContent := match[2]

		// Extract role from message-header
		rolePattern := regexp.MustCompile(`<span class="role">([^<]*)</span>`)
		roleMatch := rolePattern.FindStringSubmatch(rowContent)
		role := "unknown"
		if len(roleMatch) > 1 {
			role = roleMatch[1]
		}

		// Extract message-content div
		contentPattern := regexp.MustCompile(`(?s)<div class="message-content">(.*?)</div>`)
		contentMatch := contentPattern.FindStringSubmatch(rowContent)

		if len(contentMatch) > 1 {
			content := contentMatch[1]

			// Check if content is empty or contains only whitespace
			trimmed := strings.TrimSpace(content)

			// Also check if it contains only HTML whitespace elements
			// Remove all HTML tags and check what's left
			withoutTags := stripHTMLTags(trimmed)
			withoutWhitespace := strings.TrimSpace(withoutTags)

			if withoutWhitespace == "" {
				// This is an empty bubble
				snippet := content
				if len(snippet) > 100 {
					snippet = snippet[:100] + "..."
				}

				results = append(results, EmptyBubble{
					UUID:           uuid,
					Role:           role,
					ContentSnippet: snippet,
				})
			}
		}
	}

	return results
}

// stripHTMLTags removes all HTML tags from a string
func stripHTMLTags(html string) string {
	// Remove HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	result := tagPattern.ReplaceAllString(html, "")

	// Decode HTML entities
	result = strings.ReplaceAll(result, "&nbsp;", " ")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	result = strings.ReplaceAll(result, "&#39;", "'")

	return result
}

// TestHasContentFunction tests the hasContent function directly using the actual data structures
func TestHasContentFunction(t *testing.T) {
	// Note: We can't easily construct test entries with the complex Message structure,
	// so we test with real session data in TestNoEmptyMessageBubbles instead.
	// This test serves as documentation of the expected behavior.
	t.Skip("Use TestNoEmptyMessageBubbles for real-world testing")
}
