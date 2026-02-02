// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

// RenderConversation generates a complete HTML page for a conversation.
// entries contains the conversation history, agents contains the agent hierarchy.
func RenderConversation(entries []models.ConversationEntry, agents []*agent.TreeNode) (string, error) {
	var sb strings.Builder

	// Write HTML header
	sb.WriteString(htmlHeader)

	// Write conversation entries
	sb.WriteString(`<div class="conversation">` + "\n")

	// Build a map of agent IDs to entry counts for subagent display
	agentMap := buildAgentMap(agents)

	// Track tool results for matching with tool calls
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		entryHTML := renderEntry(entry, toolResults)
		sb.WriteString(entryHTML)

		// Check if this entry spawned a subagent
		if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
			subagentHTML := renderSubagentPlaceholder(entry.AgentID, agentMap)
			sb.WriteString(subagentHTML)
		}
	}

	sb.WriteString("</div>\n")

	// Write HTML footer
	sb.WriteString(htmlFooter)

	return sb.String(), nil
}

// RenderAgentFragment generates an HTML fragment for a subagent's conversation.
// This is used for lazy loading subagent content.
func RenderAgentFragment(agentID string, entries []models.ConversationEntry) (string, error) {
	var sb strings.Builder

	// Track tool results for this agent's entries
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		entryHTML := renderEntry(entry, toolResults)
		sb.WriteString(entryHTML)
	}

	return sb.String(), nil
}

// renderEntry renders a single conversation entry as HTML.
func renderEntry(entry models.ConversationEntry, toolResults map[string]models.ToolResult) string {
	var sb strings.Builder

	entryClass := getEntryClass(entry.Type)
	timestamp := formatTimestamp(entry.Timestamp)

	sb.WriteString(fmt.Sprintf(`<div class="entry %s" data-uuid="%s">`, entryClass, escapeHTML(entry.UUID)))
	sb.WriteString("\n")

	// Entry header with timestamp and type
	sb.WriteString(`  <div class="entry-header">`)
	sb.WriteString(fmt.Sprintf(`<span class="timestamp">%s</span>`, escapeHTML(timestamp)))
	sb.WriteString(fmt.Sprintf(` <span class="type">%s</span>`, escapeHTML(string(entry.Type))))
	if entry.AgentID != "" {
		sb.WriteString(fmt.Sprintf(` <span class="agent-id">[%s]</span>`, escapeHTML(entry.AgentID)))
	}
	sb.WriteString("</div>\n")

	// Entry content
	sb.WriteString(`  <div class="content">`)

	// Get text content
	textContent := entry.GetTextContent()
	if textContent != "" {
		sb.WriteString(fmt.Sprintf(`<div class="text">%s</div>`, escapeHTML(textContent)))
	}

	// Render tool calls for assistant messages
	if entry.Type == models.EntryTypeAssistant {
		tools := entry.ExtractToolCalls()
		for _, tool := range tools {
			toolResult, hasResult := toolResults[tool.ID]
			toolHTML := renderToolCall(tool, toolResult, hasResult)
			sb.WriteString(toolHTML)
		}
	}

	sb.WriteString("</div>\n")
	sb.WriteString("</div>\n")

	return sb.String()
}

// renderToolCall renders a single tool call as an expandable HTML section.
func renderToolCall(tool models.ToolUse, result models.ToolResult, hasResult bool) string {
	var sb strings.Builder

	toolSummary := formatToolSummary(tool)

	sb.WriteString(fmt.Sprintf(`<div class="tool-call" data-tool-id="%s">`, escapeHTML(tool.ID)))
	sb.WriteString("\n")

	// Collapsible header
	sb.WriteString(fmt.Sprintf(`  <div class="tool-header" onclick="toggleTool(this)">%s</div>`, escapeHTML(toolSummary)))
	sb.WriteString("\n")

	// Hidden body with input and output
	sb.WriteString(`  <div class="tool-body hidden">`)
	sb.WriteString("\n")

	// Tool input
	inputJSON := formatToolInput(tool.Input)
	sb.WriteString(fmt.Sprintf(`    <pre class="tool-input">%s</pre>`, escapeHTML(inputJSON)))
	sb.WriteString("\n")

	// Tool output (if available)
	if hasResult {
		outputClass := "tool-output"
		if result.IsError {
			outputClass = "tool-output error"
		}
		sb.WriteString(fmt.Sprintf(`    <pre class="%s">%s</pre>`, outputClass, escapeHTML(result.Content)))
		sb.WriteString("\n")
	}

	sb.WriteString("  </div>\n")
	sb.WriteString("</div>\n")

	return sb.String()
}

// renderSubagentPlaceholder renders a placeholder for a subagent section.
func renderSubagentPlaceholder(agentID string, agentMap map[string]int) string {
	var sb strings.Builder

	entryCount := agentMap[agentID]
	shortID := agentID
	if len(shortID) > 7 {
		shortID = shortID[:7]
	}

	sb.WriteString(fmt.Sprintf(`<div class="subagent" data-agent-id="%s">`, escapeHTML(agentID)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  <div class="subagent-header" onclick="loadAgent(this)">Subagent: %s (%d entries)</div>`, escapeHTML(shortID), entryCount))
	sb.WriteString("\n")
	sb.WriteString(`  <div class="subagent-content"></div>`)
	sb.WriteString("\n")
	sb.WriteString("</div>\n")

	return sb.String()
}

// escapeHTML escapes a string to prevent XSS attacks.
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

// getEntryClass returns the CSS class for an entry type.
func getEntryClass(entryType models.EntryType) string {
	switch entryType {
	case models.EntryTypeUser:
		return "user"
	case models.EntryTypeAssistant:
		return "assistant"
	case models.EntryTypeSystem:
		return "system"
	case models.EntryTypeQueueOperation:
		return "queue-operation"
	case models.EntryTypeSummary:
		return "summary"
	default:
		return "unknown"
	}
}

// formatTimestamp formats a timestamp for display.
func formatTimestamp(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("15:04:05")
}

// formatToolSummary creates a summary string for a tool call header.
func formatToolSummary(tool models.ToolUse) string {
	displayValue := extractToolDisplayValue(tool.Name, tool.Input)
	if displayValue == "" {
		return fmt.Sprintf("[%s]", tool.Name)
	}

	// Truncate if too long
	const maxLen = 60
	if len(displayValue) > maxLen {
		displayValue = displayValue[:maxLen-3] + "..."
	}

	return fmt.Sprintf("[%s] %s", tool.Name, displayValue)
}

// extractToolDisplayValue extracts the most relevant display value from tool input.
func extractToolDisplayValue(toolName string, input map[string]any) string {
	if input == nil {
		return ""
	}

	switch toolName {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return cmd
		}
	case "Read":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Write":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Edit":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "Grep":
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
	case "Glob":
		if pattern, ok := input["pattern"].(string); ok {
			return pattern
		}
	case "Task":
		if desc, ok := input["description"].(string); ok {
			return desc
		}
		if prompt, ok := input["prompt"].(string); ok {
			return prompt
		}
	case "WebFetch":
		if url, ok := input["url"].(string); ok {
			return url
		}
	case "WebSearch":
		if query, ok := input["query"].(string); ok {
			return query
		}
	}

	return ""
}

// formatToolInput formats tool input as indented JSON.
func formatToolInput(input map[string]any) string {
	if input == nil {
		return "{}"
	}

	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(data)
}

// buildAgentMap creates a map of agent IDs to entry counts from the agent tree.
func buildAgentMap(agents []*agent.TreeNode) map[string]int {
	result := make(map[string]int)

	var walk func(nodes []*agent.TreeNode)
	walk = func(nodes []*agent.TreeNode) {
		for _, node := range nodes {
			if node.AgentID != "" {
				result[node.AgentID] = node.EntryCount
			}
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}

	walk(agents)
	return result
}

// buildToolResultsMap creates a map of tool use IDs to their results.
// This allows matching tool calls with their corresponding results.
func buildToolResultsMap(entries []models.ConversationEntry) map[string]models.ToolResult {
	result := make(map[string]models.ToolResult)

	for _, entry := range entries {
		if entry.Type == models.EntryTypeUser {
			results := entry.ExtractToolResults()
			for _, r := range results {
				result[r.ToolUseID] = r
			}
		}
	}

	return result
}

// HTML template constants
const htmlHeader = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Claude Conversation Export</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
`

const htmlFooter = `    <script src="static/script.js"></script>
</body>
</html>
`
