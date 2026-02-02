// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"fmt"
	"strings"

	"github.com/randlee/claude-history/pkg/models"
)

// ToolInfo contains metadata about a tool type for rendering.
type ToolInfo struct {
	Icon        string // Emoji icon for the tool
	Hint        string // Human-readable description
	ColorClass  string // CSS class for color coding
	DisplayName string // Display name (may differ from tool name)
}

// toolRegistry maps tool names to their display metadata.
var toolRegistry = map[string]ToolInfo{
	"Bash":         {Icon: "\xf0\x9f\x94\xa7", Hint: "command execution", ColorClass: "tool-bash", DisplayName: "Bash"},
	"Read":         {Icon: "\xf0\x9f\x93\x84", Hint: "file read", ColorClass: "tool-read", DisplayName: "Read"},
	"Write":        {Icon: "\xf0\x9f\x93\x9d", Hint: "file write", ColorClass: "tool-write", DisplayName: "Write"},
	"Edit":         {Icon: "\xe2\x9c\x8f\xef\xb8\x8f", Hint: "file edit", ColorClass: "tool-edit", DisplayName: "Edit"},
	"Grep":         {Icon: "\xf0\x9f\x94\x8d", Hint: "content search", ColorClass: "tool-grep", DisplayName: "Grep"},
	"Glob":         {Icon: "\xf0\x9f\x93\x81", Hint: "file pattern matching", ColorClass: "tool-glob", DisplayName: "Glob"},
	"Task":         {Icon: "\xf0\x9f\xa4\x96", Hint: "spawn subagent", ColorClass: "tool-task", DisplayName: "Task"},
	"WebFetch":     {Icon: "\xf0\x9f\x8c\x90", Hint: "fetch URL", ColorClass: "tool-webfetch", DisplayName: "WebFetch"},
	"WebSearch":    {Icon: "\xf0\x9f\x94\x8e", Hint: "web search", ColorClass: "tool-websearch", DisplayName: "WebSearch"},
	"NotebookEdit": {Icon: "\xf0\x9f\x93\x93", Hint: "notebook edit", ColorClass: "tool-notebook", DisplayName: "NotebookEdit"},
	"TodoRead":     {Icon: "\xf0\x9f\x93\x8b", Hint: "read todos", ColorClass: "tool-todo", DisplayName: "TodoRead"},
	"TodoWrite":    {Icon: "\xf0\x9f\x93\x8b", Hint: "write todos", ColorClass: "tool-todo", DisplayName: "TodoWrite"},
	"ToolSearch":   {Icon: "\xf0\x9f\x94\xa7", Hint: "search tools", ColorClass: "tool-search", DisplayName: "ToolSearch"},
	"Skill":        {Icon: "\xf0\x9f\x8e\xaf", Hint: "invoke skill", ColorClass: "tool-skill", DisplayName: "Skill"},
}

// GetToolInfo returns the ToolInfo for a given tool name.
// If the tool is not in the registry, returns a default info.
func GetToolInfo(toolName string) ToolInfo {
	if info, ok := toolRegistry[toolName]; ok {
		return info
	}
	// Default for unknown tools
	return ToolInfo{
		Icon:        "\xf0\x9f\x94\xa7", // wrench emoji
		Hint:        "tool",
		ColorClass:  "tool-unknown",
		DisplayName: toolName,
	}
}

// RenderToolOverlay renders a tool call as an expandable overlay with color coding.
func RenderToolOverlay(tool models.ToolUse, result models.ToolResult, hasResult bool) string {
	var sb strings.Builder

	info := GetToolInfo(tool.Name)

	// Calculate character counts
	inputJSON := formatToolInput(tool.Input)
	inputCharCount := len(inputJSON)
	outputCharCount := 0
	if hasResult {
		outputCharCount = len(result.Content)
	}

	// Main overlay container with tool-specific color class
	sb.WriteString(fmt.Sprintf(`<div class="tool-overlay %s collapsible" data-tool-id="%s">`,
		escapeHTML(info.ColorClass), escapeHTML(tool.ID)))
	sb.WriteString("\n")

	// Header (collapsible trigger)
	sb.WriteString(`  <div class="tool-header collapsible-trigger" onclick="toggleToolOverlay(this)">`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`    <span class="tool-icon">%s</span>`, info.Icon))
	sb.WriteString(fmt.Sprintf(`<span class="tool-name">%s</span>`, escapeHTML(info.DisplayName)))
	sb.WriteString(fmt.Sprintf(`<span class="tool-hint">(%s)</span>`, escapeHTML(info.Hint)))

	// Character count hint
	sb.WriteString(fmt.Sprintf(`<span class="tool-char-count">%s</span>`, formatCharCount(inputCharCount, outputCharCount)))

	// Copy button for tool ID
	sb.WriteString(renderCopyButton(tool.ID, "tool-id", "Copy tool ID"))

	// File path copy button for file-related tools
	filePath := extractFilePath(tool.Name, tool.Input)
	if filePath != "" {
		sb.WriteString(renderCopyButton(filePath, "file-path", "Copy file path"))
	}

	// Chevron indicator
	sb.WriteString(`<span class="chevron down">`)
	sb.WriteString("\xe2\x96\xbc") // down arrow
	sb.WriteString(`</span>`)

	sb.WriteString("\n  </div>\n")

	// Body (collapsible content) - starts hidden
	sb.WriteString(`  <div class="tool-body collapsible-content collapsed">`)
	sb.WriteString("\n")

	// Tool ID section
	sb.WriteString(`    <div class="tool-section">`)
	sb.WriteString("\n")
	sb.WriteString(`      <h4>Tool ID</h4>`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`      <code class="tool-id-value">%s%s</code>`,
		escapeHTML(tool.ID),
		renderCopyButton(tool.ID, "tool-id", "Copy tool ID")))
	sb.WriteString("\n    </div>\n")

	// Input section
	sb.WriteString(`    <div class="tool-section">`)
	sb.WriteString("\n")
	sb.WriteString(`      <h4>Input</h4>`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`      <pre class="tool-input"><code>%s</code>%s</pre>`,
		escapeHTML(inputJSON),
		renderCopyButton(inputJSON, "tool-input", "Copy input")))
	sb.WriteString("\n    </div>\n")

	// Output section (if available)
	if hasResult {
		outputClass := "tool-output"
		if result.IsError {
			outputClass = "tool-output error"
		}
		sb.WriteString(`    <div class="tool-section">`)
		sb.WriteString("\n")
		sb.WriteString(`      <h4>Output</h4>`)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf(`      <pre class="%s"><code>%s</code>%s</pre>`,
			outputClass,
			escapeHTML(result.Content),
			renderCopyButton(result.Content, "tool-output", "Copy output")))
		sb.WriteString("\n    </div>\n")
	}

	sb.WriteString("  </div>\n")
	sb.WriteString("</div>\n")

	return sb.String()
}

// RenderSubagentOverlay renders a subagent placeholder as an expandable overlay.
func RenderSubagentOverlay(agentID string, entryCount int, metadata map[string]string) string {
	var sb strings.Builder

	shortID := agentID
	if len(shortID) > 7 {
		shortID = shortID[:7]
	}

	// Main overlay container
	sb.WriteString(fmt.Sprintf(`<div class="subagent-overlay agent-overlay collapsible" data-agent-id="%s">`,
		escapeHTML(agentID)))
	sb.WriteString("\n")

	// Header (collapsible trigger)
	sb.WriteString(`  <div class="subagent-header overlay-header collapsible-trigger" onclick="loadAgent(this)">`)
	sb.WriteString("\n")
	sb.WriteString(`    <span class="agent-icon">`)
	sb.WriteString("\xf0\x9f\xa4\x96") // robot emoji
	sb.WriteString(`</span>`)
	sb.WriteString(fmt.Sprintf(`<span class="subagent-title">Subagent: %s</span>`, escapeHTML(shortID)))
	sb.WriteString(fmt.Sprintf(`<span class="subagent-meta">(%d entries)</span>`, entryCount))

	// Copy button for agent ID
	sb.WriteString(renderCopyButton(agentID, "agent-id", "Copy agent ID"))

	// Deep Dive button
	sb.WriteString(fmt.Sprintf(`<button class="deep-dive-btn" onclick="deepDiveAgent('%s', event)" title="Open agent conversation">`,
		escapeHTML(agentID)))
	sb.WriteString(`Deep Dive</button>`)

	// Chevron indicator
	sb.WriteString(`<span class="chevron down">`)
	sb.WriteString("\xe2\x96\xbc") // down arrow
	sb.WriteString(`</span>`)

	sb.WriteString("\n  </div>\n")

	// Metadata section (if provided)
	if len(metadata) > 0 {
		sb.WriteString(`  <div class="subagent-metadata">`)
		sb.WriteString("\n")
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf(`    <span class="meta-item"><strong>%s:</strong> %s</span>`,
				escapeHTML(key), escapeHTML(value)))
			sb.WriteString("\n")
		}
		sb.WriteString("  </div>\n")
	}

	// Content container (lazy loaded)
	sb.WriteString(`  <div class="subagent-content collapsible-content collapsed"></div>`)
	sb.WriteString("\n")

	sb.WriteString("</div>\n")

	return sb.String()
}

// ThinkingBlock represents a thinking block from assistant content.
type ThinkingBlock struct {
	Content string
}

// RenderThinkingOverlay renders a thinking block as an expandable overlay.
func RenderThinkingOverlay(thinking ThinkingBlock) string {
	var sb strings.Builder

	charCount := len(thinking.Content)

	// Main overlay container
	sb.WriteString(`<div class="thinking-overlay collapsible">`)
	sb.WriteString("\n")

	// Header (collapsible trigger)
	sb.WriteString(`  <div class="thinking-header overlay-header collapsible-trigger" onclick="toggleThinking(this)">`)
	sb.WriteString("\n")
	sb.WriteString(`    <span class="thinking-icon">`)
	sb.WriteString("\xf0\x9f\x92\xa1") // lightbulb emoji
	sb.WriteString(`</span>`)
	sb.WriteString(`<span class="thinking-title">Thinking</span>`)
	sb.WriteString(fmt.Sprintf(`<span class="thinking-char-count">(%s)</span>`, formatSingleCharCount(charCount)))

	// Chevron indicator
	sb.WriteString(`<span class="chevron down">`)
	sb.WriteString("\xe2\x96\xbc") // down arrow
	sb.WriteString(`</span>`)

	sb.WriteString("\n  </div>\n")

	// Body (collapsible content) - starts hidden
	sb.WriteString(`  <div class="thinking-body collapsible-content collapsed">`)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`    <pre class="thinking-content"><code>%s</code></pre>`,
		escapeHTML(thinking.Content)))
	sb.WriteString("\n  </div>\n")

	sb.WriteString("</div>\n")

	return sb.String()
}

// ExtractThinkingBlocks extracts thinking blocks from assistant message content.
// Thinking blocks appear as content blocks with type "thinking".
func ExtractThinkingBlocks(entry models.ConversationEntry) []ThinkingBlock {
	if entry.Type != models.EntryTypeAssistant {
		return nil
	}

	contents, err := entry.ParseMessageContent()
	if err != nil {
		return nil
	}

	var blocks []ThinkingBlock
	for _, c := range contents {
		if c.Type == "thinking" && c.Text != "" {
			blocks = append(blocks, ThinkingBlock{Content: c.Text})
		}
	}

	return blocks
}

// formatCharCount formats input and output character counts for display.
func formatCharCount(inputChars, outputChars int) string {
	if outputChars == 0 {
		return fmt.Sprintf("%s in", formatSize(inputChars))
	}
	return fmt.Sprintf("%s in / %s out", formatSize(inputChars), formatSize(outputChars))
}

// formatSingleCharCount formats a single character count for display.
func formatSingleCharCount(chars int) string {
	return formatSize(chars)
}

// formatSize formats a character count with appropriate units.
func formatSize(chars int) string {
	if chars < 1000 {
		return fmt.Sprintf("%d chars", chars)
	}
	if chars < 10000 {
		return fmt.Sprintf("%.1fK chars", float64(chars)/1000)
	}
	return fmt.Sprintf("%.0fK chars", float64(chars)/1000)
}
