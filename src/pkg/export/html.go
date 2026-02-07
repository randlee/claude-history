// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/randlee/claude-history/pkg/agent"
	"github.com/randlee/claude-history/pkg/models"
)

// SessionStats contains statistics about a session for display in the header.
type SessionStats struct {
	SessionID          string // Full session ID
	ProjectPath        string // Project directory path
	ExportTime         string // Formatted export timestamp (kept for backward compat, not displayed)
	SessionStart       string // First entry timestamp (formatted for display)
	SessionEnd         string // Last entry timestamp (formatted for display)
	Duration           string // Human-readable duration (e.g., "2h 35m")
	MessageCount       int    // Count of user + assistant messages (deprecated, kept for backward compat)
	UserMessages       int    // Count of user messages
	AssistantMessages  int    // Count of assistant messages (main session only)
	SubagentMessages   int    // Count of all subagent messages
	AgentCount         int    // Count of subagents
	TotalAgentMessages int    // Total messages across all subagents
	ToolCallCount      int    // Count of tool calls
}

// ExportFormatVersion is the current version of the export format.
const ExportFormatVersion = "2.0"

// RenderConversation generates a complete HTML page for a conversation.
// entries contains the conversation history, agents contains the agent hierarchy.
func RenderConversation(entries []models.ConversationEntry, agents []*agent.TreeNode) (string, error) {
	return RenderConversationWithStats(entries, agents, nil)
}

// RenderConversationWithStats generates a complete HTML page for a conversation with session statistics.
// entries contains the conversation history, agents contains the agent hierarchy,
// stats contains optional session statistics for the header (if nil, stats are computed from entries/agents).
func RenderConversationWithStats(entries []models.ConversationEntry, agents []*agent.TreeNode, stats *SessionStats) (string, error) {
	var sb strings.Builder

	// Calculate stats if not provided
	if stats == nil {
		stats = ComputeSessionStats(entries, agents)
	}

	// Write HTML header with metadata
	sb.WriteString(renderHTMLHeader(stats))

	// Write conversation entries
	sb.WriteString(`<div class="conversation">` + "\n")

	// Build a map of agent IDs to entry counts for subagent display
	agentMap := buildAgentMap(agents)

	// Track tool results for matching with tool calls
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		// Skip entries with no meaningful content
		if !hasContent(entry) {
			// Still render subagent placeholder if this entry spawned one
			if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
				subagentHTML := renderSubagentPlaceholder(entry.AgentID, agentMap)
				sb.WriteString(subagentHTML)
			}
			continue
		}

		entryHTML := renderEntry(entry, toolResults)
		sb.WriteString(entryHTML)

		// Check if this entry spawned a subagent
		if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
			subagentHTML := renderSubagentPlaceholder(entry.AgentID, agentMap)
			sb.WriteString(subagentHTML)
		}
	}

	sb.WriteString("</div>\n")

	// Write HTML footer with info and keyboard shortcuts
	sb.WriteString(renderHTMLFooter(stats))

	return sb.String(), nil
}

// ComputeSessionStats calculates statistics from entries and agents.
func ComputeSessionStats(entries []models.ConversationEntry, agents []*agent.TreeNode) *SessionStats {
	stats := &SessionStats{
		ExportTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Get session start/end times from first/last entries with timestamps
	if len(entries) > 0 {
		// Find first entry with a timestamp
		var firstTime time.Time
		for _, entry := range entries {
			if entry.Timestamp != "" {
				if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
					firstTime = t
					stats.SessionStart = firstTime.Format("2006-01-02 15:04")
					break
				}
			}
		}

		// Find last entry with a timestamp (search backwards)
		var lastTime time.Time
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i].Timestamp != "" {
				if t, err := time.Parse(time.RFC3339Nano, entries[i].Timestamp); err == nil {
					lastTime = t
					stats.SessionEnd = lastTime.Format("2006-01-02 15:04")
					break
				}
			}
		}

		// Calculate duration if we have both timestamps
		if !firstTime.IsZero() && !lastTime.IsZero() {
			duration := lastTime.Sub(firstTime)
			stats.Duration = formatDuration(duration)
		}
	}

	// Count messages by type
	for _, entry := range entries {
		switch entry.Type {
		case models.EntryTypeUser:
			stats.UserMessages++
			stats.MessageCount++ // Keep for backward compat
		case models.EntryTypeAssistant:
			stats.AssistantMessages++
			stats.MessageCount++ // Keep for backward compat
			// Count tool calls from assistant messages
			tools := entry.ExtractToolCalls()
			stats.ToolCallCount += len(tools)
		}
		// Extract session ID from first entry if available
		if stats.SessionID == "" && entry.SessionID != "" {
			stats.SessionID = entry.SessionID
		}
	}

	// Count agents and subagent messages
	if len(agents) > 0 {
		agentMap := buildAgentMap(agents)
		stats.AgentCount = len(agentMap)

		// Sum all subagent entry counts
		for _, count := range agentMap {
			stats.TotalAgentMessages += count
		}
		stats.SubagentMessages = stats.TotalAgentMessages
	}

	return stats
}

// formatDuration formats a duration into a human-readable string.
// Examples: "2h 35m", "45m", "30s"
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	seconds := int(d.Seconds())
	return fmt.Sprintf("%ds", seconds)
}

// TruncateSessionID returns a truncated session ID for display (first 8 chars).
func TruncateSessionID(sessionID string) string {
	if len(sessionID) > 8 {
		return sessionID[:8]
	}
	return sessionID
}

// RenderAgentFragment generates an HTML fragment for a subagent's conversation.
// This is used for lazy loading subagent content.
func RenderAgentFragment(agentID string, entries []models.ConversationEntry) (string, error) {
	var sb strings.Builder

	// Track tool results for this agent's entries
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		// Skip entries with no meaningful content
		if !hasContent(entry) {
			continue
		}

		entryHTML := renderEntry(entry, toolResults)
		sb.WriteString(entryHTML)
	}

	return sb.String(), nil
}

// hasContent checks if an entry has meaningful content worth rendering.
// Returns false for empty messages, true if the entry has text, tool calls, or other content.
func hasContent(entry models.ConversationEntry) bool {
	// Check for text content (trim whitespace to detect empty/whitespace-only messages)
	textContent := entry.GetTextContent()
	if strings.TrimSpace(textContent) != "" {
		return true
	}

	// Check for tool calls in assistant messages
	if entry.Type == models.EntryTypeAssistant {
		tools := entry.ExtractToolCalls()
		if len(tools) > 0 {
			return true
		}
	}

	// Queue operations without content can be skipped (we still render subagent placeholder)
	// Summary entries without content should be skipped
	// Other entry types with no text should be skipped
	return false
}

// renderEntry renders a single conversation entry as HTML using the chat bubble layout.
func renderEntry(entry models.ConversationEntry, toolResults map[string]models.ToolResult) string {
	var sb strings.Builder

	// Get text content
	textContent := entry.GetTextContent()

	// Detect task-notification blocks and override entry type/styling
	isTaskNotif := entry.Type == models.EntryTypeUser && strings.Contains(textContent, "<task-notification>")

	entryType := entry.Type
	roleLabel := getRoleLabel(entry.Type)

	// Override for task-notification
	if isTaskNotif {
		entryType = models.EntryTypeSystem
		roleLabel = "Agent Notification"
	}

	entryClass := getEntryClass(entryType)
	timestamp := formatTimestampReadable(entry.Timestamp)

	// Message row with alignment based on type
	sb.WriteString(fmt.Sprintf(`<div class="message-row %s" data-uuid="%s">`, entryClass, escapeHTML(entry.UUID)))
	sb.WriteString("\n")

	// Avatar placeholder
	sb.WriteString(fmt.Sprintf(`  <div class="avatar %s" aria-hidden="true"></div>`, entryClass))
	sb.WriteString("\n")

	// Message bubble
	sb.WriteString(`  <div class="message-bubble">`)
	sb.WriteString("\n")

	// Message header with role and timestamp
	sb.WriteString(`    <div class="message-header">`)
	sb.WriteString(fmt.Sprintf(`<span class="role">%s</span>`, escapeHTML(roleLabel)))
	if entry.AgentID != "" {
		sb.WriteString(fmt.Sprintf(` <span class="agent-id">[%s]%s</span>`,
			escapeHTML(entry.AgentID),
			renderCopyButton(entry.AgentID, "agent-id", "Copy agent ID")))
	}
	sb.WriteString(fmt.Sprintf(` <span class="timestamp">%s</span>`, escapeHTML(timestamp)))
	sb.WriteString("</div>\n")

	// Message content
	sb.WriteString(`    <div class="message-content">`)

	if textContent != "" {
		// Special handling for task-notification blocks
		if isTaskNotif {
			sb.WriteString(renderTaskNotification(textContent))
		} else if entry.Type == models.EntryTypeAssistant {
			// Apply markdown rendering for assistant messages
			sb.WriteString(fmt.Sprintf(`<div class="text markdown-content">%s</div>`, RenderMarkdown(textContent)))
		} else {
			// Regular user message - format XML tags for better display
			sb.WriteString(fmt.Sprintf(`<div class="text user-content">%s</div>`, formatUserContent(textContent)))
		}
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

	sb.WriteString("</div>\n")   // Close message-content
	sb.WriteString("  </div>\n") // Close message-bubble
	sb.WriteString("</div>\n")   // Close message-row

	return sb.String()
}

// getRoleLabel returns a human-readable label for the entry type.
func getRoleLabel(entryType models.EntryType) string {
	switch entryType {
	case models.EntryTypeUser:
		return "User"
	case models.EntryTypeAssistant:
		return "Assistant"
	case models.EntryTypeSystem:
		return "System"
	case models.EntryTypeQueueOperation:
		return "Agent Task"
	case models.EntryTypeSummary:
		return "Summary"
	default:
		return string(entryType)
	}
}

// formatTimestampReadable formats a timestamp for display as a readable time (e.g., "2:30 PM").
func formatTimestampReadable(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("3:04 PM")
}

// renderToolCall renders a single tool call as an expandable HTML section.
func renderToolCall(tool models.ToolUse, result models.ToolResult, hasResult bool) string {
	var sb strings.Builder

	toolSummary := formatToolSummary(tool)

	sb.WriteString(fmt.Sprintf(`<div class="tool-call collapsible collapsed" data-tool-id="%s">`, escapeHTML(tool.ID)))
	sb.WriteString("\n")

	// Collapsible header with tool ID copy button and chevron
	sb.WriteString(fmt.Sprintf(`  <div class="tool-header collapsible-trigger" onclick="toggleTool(this)"><span class="tool-summary">%s</span>`,
		escapeHTML(toolSummary)))
	sb.WriteString(fmt.Sprintf(`<span class="tool-id">%s</span>`, renderCopyButton(tool.ID, "tool-id", "Copy tool ID")))

	// Add file path copy button for file-related tools
	filePath := extractFilePath(tool.Name, tool.Input)
	if filePath != "" {
		sb.WriteString(fmt.Sprintf(`<span class="file-path-btn">%s</span>`,
			renderCopyButton(filePath, "file-path", "Copy file path")))
	}

	// Add chevron indicator
	sb.WriteString(`<span class="chevron down">▼</span>`)

	sb.WriteString("</div>\n")

	// Hidden body with input and output (starts collapsed)
	sb.WriteString(`  <div class="tool-body hidden collapsible-content collapsed">`)
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

	sb.WriteString(fmt.Sprintf(`<div class="subagent collapsible collapsed" data-agent-id="%s">`, escapeHTML(agentID)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  <div class="subagent-header collapsible-trigger" onclick="loadAgent(this)"><span class="subagent-title">Subagent: %s</span> <span class="subagent-meta">(%d entries)</span>%s<span class="chevron down">▼</span></div>`,
		escapeHTML(shortID),
		entryCount,
		renderCopyButton(agentID, "agent-id", "Copy agent ID")))
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

// renderCopyButton generates HTML for a copy-to-clipboard button.
// text is the value to copy, copyType indicates what kind of value it is (for styling/tracking),
// and tooltip is the hover text shown to the user.
func renderCopyButton(text, copyType, tooltip string) string {
	if text == "" {
		return ""
	}
	return fmt.Sprintf(
		`<button class="copy-btn" data-copy-text="%s" data-copy-type="%s" title="%s"><span class="copy-icon">&#128203;</span></button>`,
		escapeHTML(text),
		escapeHTML(copyType),
		escapeHTML(tooltip),
	)
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

// extractFilePath extracts the file path from tool input for file-related tools.
// Returns empty string for non-file tools or if no file path is present.
func extractFilePath(toolName string, input map[string]any) string {
	if input == nil {
		return ""
	}

	switch toolName {
	case "Read", "Write", "Edit":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "NotebookEdit":
		if path, ok := input["notebook_path"].(string); ok {
			return path
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

// renderHTMLHeader generates the HTML header with session metadata.
func renderHTMLHeader(stats *SessionStats) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Claude Conversation Export</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
<header class="page-header">
    <h1>Claude Code Session</h1>
    <div class="session-metadata">
`)

	// Session ID with copy button
	if stats != nil && stats.SessionID != "" {
		truncatedID := TruncateSessionID(stats.SessionID)
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Session: <code>%s</code>%s</span>
`,
			escapeHTML(truncatedID),
			renderCopyButton(stats.SessionID, "session-id", "Copy full session ID")))
	}

	// Project path
	if stats != nil && stats.ProjectPath != "" {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Project: <code>%s</code></span>
`, escapeHTML(stats.ProjectPath)))
	}

	// Session start time
	if stats != nil && stats.SessionStart != "" {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Started: %s</span>
`, escapeHTML(stats.SessionStart)))
	}

	// Session duration
	if stats != nil && stats.Duration != "" {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Duration: %s</span>
`, escapeHTML(stats.Duration)))
	}

	// Enhanced message statistics
	if stats != nil {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">User: %d | Assistant: %d | Subagents[%d]: %d messages</span>
`, stats.UserMessages, stats.AssistantMessages, stats.AgentCount, stats.TotalAgentMessages))
	}

	// Tool call count
	if stats != nil {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Tools: %d calls</span>
`, stats.ToolCallCount))
	}

	sb.WriteString(`    </div>
    <div class="controls" role="toolbar" aria-label="Conversation controls">
        <div class="controls-group">
            <button id="expand-all-btn" type="button" data-shortcut="Ctrl+K" title="Expand all tool calls (Ctrl+K)">Expand All</button>
            <button id="collapse-all-btn" type="button" title="Collapse all tool calls">Collapse All</button>
        </div>
        <div class="controls-separator" aria-hidden="true"></div>
        <div class="search-container">
            <input type="search" id="search-box" placeholder="Search messages..." aria-label="Search messages" data-shortcut="Ctrl+F" title="Search messages (Ctrl+F)">
            <button id="search-prev-btn" type="button" class="search-nav-btn" title="Previous match (Shift+Enter)" aria-label="Previous match">&lt;</button>
            <button id="search-next-btn" type="button" class="search-nav-btn" title="Next match (Enter)" aria-label="Next match">&gt;</button>
            <span class="search-results" aria-live="polite"></span>
        </div>
    </div>
    <nav class="breadcrumbs" id="breadcrumbs" aria-label="Navigation breadcrumbs">
        <a href="#main" class="breadcrumb-item active" data-agent-id="main" aria-current="page">Main Session</a>
    </nav>
</header>
`)

	return sb.String()
}

// renderHTMLFooter generates the HTML footer with export info and keyboard shortcuts.
func renderHTMLFooter(stats *SessionStats) string {
	var sb strings.Builder

	sb.WriteString(`<footer class="page-footer">
    <div class="footer-info">
        <p>Exported from <strong>claude-history</strong> CLI</p>
`)
	sb.WriteString(fmt.Sprintf(`        <p>Export format version: %s</p>
`, ExportFormatVersion))

	// Source path with copy button if available
	if stats != nil && stats.ProjectPath != "" {
		sourcePath := fmt.Sprintf("~/.claude/projects/%s", escapeHTML(stats.ProjectPath))
		sb.WriteString(fmt.Sprintf(`        <p>Source: <code>%s</code>%s</p>
`, sourcePath, renderCopyButton(stats.ProjectPath, "source-path", "Copy source path")))
	}

	sb.WriteString(`    </div>
    <div class="footer-help">
        <details>
            <summary>Keyboard Shortcuts</summary>
            <ul>
                <li><kbd>Ctrl</kbd>+<kbd>K</kbd> - Expand/Collapse All</li>
                <li><kbd>Ctrl</kbd>+<kbd>F</kbd> - Search</li>
                <li><kbd>Esc</kbd> - Clear Search</li>
            </ul>
        </details>
    </div>
</footer>
    <script src="static/script.js"></script>
    <script src="static/clipboard.js"></script>
    <script src="static/controls.js"></script>
    <script src="static/navigation.js"></script>
</body>
</html>
`)

	return sb.String()
}

// htmlHeader is kept for backward compatibility with older tests.
// Deprecated: Use renderHTMLHeader() instead.
var htmlHeader = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Claude Conversation Export</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
<header class="page-header">
    <h1>Claude Code Session</h1>
    <div class="session-metadata">
    </div>
    <div class="controls" role="toolbar" aria-label="Conversation controls">
        <div class="controls-group">
            <button id="expand-all-btn" type="button" data-shortcut="Ctrl+K" title="Expand all tool calls (Ctrl+K)">Expand All</button>
            <button id="collapse-all-btn" type="button" title="Collapse all tool calls">Collapse All</button>
        </div>
        <div class="controls-separator" aria-hidden="true"></div>
        <div class="search-container">
            <input type="search" id="search-box" placeholder="Search messages..." aria-label="Search messages" data-shortcut="Ctrl+F" title="Search messages (Ctrl+F)">
            <button id="search-prev-btn" type="button" class="search-nav-btn" title="Previous match (Shift+Enter)" aria-label="Previous match">&lt;</button>
            <button id="search-next-btn" type="button" class="search-nav-btn" title="Next match (Enter)" aria-label="Next match">&gt;</button>
            <span class="search-results" aria-live="polite"></span>
        </div>
    </div>
    <nav class="breadcrumbs" id="breadcrumbs" aria-label="Navigation breadcrumbs">
        <a href="#main" class="breadcrumb-item active" data-agent-id="main" aria-current="page">Main Session</a>
    </nav>
</header>
`

// htmlFooter is kept for backward compatibility with older tests.
// Deprecated: Use renderHTMLFooter() instead.
var htmlFooter = `<footer class="page-footer">
    <div class="footer-info">
        <p>Exported from <strong>claude-history</strong> CLI</p>
    </div>
</footer>
    <script src="static/script.js"></script>
    <script src="static/clipboard.js"></script>
    <script src="static/controls.js"></script>
    <script src="static/navigation.js"></script>
</body>
</html>
`

// TaskNotificationData holds parsed data from a task-notification XML block.
type TaskNotificationData struct {
	TaskID  string
	Status  string
	Summary string
	Result  string
}

// parseTaskNotification extracts structured data from a task-notification XML block.
func parseTaskNotification(content string) *TaskNotificationData {
	if !strings.Contains(content, "<task-notification>") {
		return nil
	}

	data := &TaskNotificationData{}

	// Extract task-id
	if matches := regexp.MustCompile(`<task-id>(.*?)</task-id>`).FindStringSubmatch(content); len(matches) > 1 {
		data.TaskID = strings.TrimSpace(matches[1])
	}

	// Extract status
	if matches := regexp.MustCompile(`<status>(.*?)</status>`).FindStringSubmatch(content); len(matches) > 1 {
		data.Status = strings.TrimSpace(matches[1])
	}

	// Extract summary
	if matches := regexp.MustCompile(`<summary>(.*?)</summary>`).FindStringSubmatch(content); len(matches) > 1 {
		data.Summary = strings.TrimSpace(matches[1])
	}

	// Extract result (may contain newlines, use (?s) for dot-all mode)
	if matches := regexp.MustCompile(`(?s)<result>(.*?)</result>`).FindStringSubmatch(content); len(matches) > 1 {
		data.Result = strings.TrimSpace(matches[1])
	}

	return data
}

// renderTaskNotification renders a task-notification block with special formatting.
func renderTaskNotification(content string) string {
	data := parseTaskNotification(content)
	if data == nil {
		// Fallback to escaped content if parsing fails
		return fmt.Sprintf(`<div class="text">%s</div>`, escapeHTML(content))
	}

	var sb strings.Builder

	sb.WriteString(`<div class="task-notification">`)

	// Header with icon and status
	statusIcon := "✓"
	statusClass := "completed"
	switch data.Status {
	case "failed", "error":
		statusIcon = "✗"
		statusClass = "failed"
	case "running":
		statusIcon = "⏳"
		statusClass = "running"
	}

	sb.WriteString(fmt.Sprintf(`<div class="task-notification-header status-%s">`, statusClass))
	sb.WriteString(fmt.Sprintf(`<span class="status-icon">%s</span> `, statusIcon))
	sb.WriteString(fmt.Sprintf(`<span class="summary">%s</span>`, escapeHTML(data.Summary)))

	// Add task ID badge if present
	if data.TaskID != "" {
		sb.WriteString(fmt.Sprintf(` <span class="task-id-badge" title="Task ID">%s</span>`,
			escapeHTML(data.TaskID)))
	}

	sb.WriteString("</div>\n")

	// Result content (collapsible if long)
	if data.Result != "" {
		isLong := len(data.Result) > 300

		if isLong {
			// Make it collapsible
			sb.WriteString(`<details class="task-notification-result">`)
			sb.WriteString(`<summary>View result</summary>`)
			sb.WriteString(fmt.Sprintf(`<div class="task-result-content">%s</div>`,
				escapeHTML(data.Result)))
			sb.WriteString(`</details>`)
		} else {
			// Show inline
			sb.WriteString(fmt.Sprintf(`<div class="task-result-content">%s</div>`,
				escapeHTML(data.Result)))
		}
	}

	sb.WriteString("</div>\n")

	return sb.String()
}

// formatUserContent formats user message content, processing XML-like tags for better display.
// This improves readability of bash-stdout, bash-stderr, and other tool result XML blocks in USER INPUT messages.
// Empty tags are hidden, and non-empty tags are wrapped in styled divs with proper spacing.
func formatUserContent(content string) string {
	if content == "" {
		return ""
	}

	// Go's regexp doesn't support backreferences, so we match both tags and verify they match
	// Pattern: <tag-name>content</any-tag-name>
	// Use (?s) flag to make . match newlines
	tagPattern := regexp.MustCompile(`(?s)<([a-z][a-z0-9\-]*)((?:\s+[^>]*)?)>(.*?)</([a-z][a-z0-9\-]*)>`)

	// Find all XML-like tag blocks
	matches := tagPattern.FindAllStringSubmatch(content, -1)
	matchIndices := tagPattern.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		// No XML tags found, just escape and return
		return escapeHTML(content)
	}

	var result strings.Builder
	lastEnd := 0

	for i, match := range matches {
		matchIndex := matchIndices[i]
		openingTag := match[1] // Captured opening tag name
		closingTag := match[4] // Captured closing tag name
		tagContent := match[3] // Content between tags

		// Only process if opening and closing tags match
		if openingTag != closingTag {
			continue
		}

		// Add any text before this tag
		if matchIndex[0] > lastEnd {
			beforeText := content[lastEnd:matchIndex[0]]
			result.WriteString(escapeHTML(beforeText))
		}

		// Skip empty tags (e.g., <bash-stderr></bash-stderr>)
		if strings.TrimSpace(tagContent) == "" {
			lastEnd = matchIndex[1]
			continue
		}

		// Render non-empty tags with proper formatting
		// Tag names are escaped, content is escaped separately
		result.WriteString(fmt.Sprintf(`<div class="xml-tag-block">&lt;%s&gt;`, escapeHTML(openingTag)))
		result.WriteString(fmt.Sprintf(`<div class="xml-tag-content">%s</div>`, escapeHTML(tagContent)))
		result.WriteString(fmt.Sprintf(`&lt;/%s&gt;</div>`, escapeHTML(openingTag)))

		lastEnd = matchIndex[1]
	}

	// Add any remaining text after the last tag
	if lastEnd < len(content) {
		result.WriteString(escapeHTML(content[lastEnd:]))
	}

	return result.String()
}
