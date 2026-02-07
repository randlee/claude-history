// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"encoding/json"
	"fmt"
	"html"
	"path/filepath"
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
	SessionFolderPath  string // Full path to session folder (for file:// links)
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

// RenderQueryResults generates a simplified HTML page for query results.
// This is used by the query command to display filtered conversation entries.
// Unlike RenderConversation, this does not include agent tree navigation or lazy-loading features.
// userLabel and assistantLabel specify the role names to use (e.g., "User"/"Assistant" or "Orchestrator"/"Agent").
// sessionFolderPath is the absolute path to the session folder (optional, used for file:// links).
// agentID is the agent ID if this is a subagent query (used to determine page title and correct agent ID display).
func RenderQueryResults(entries []models.ConversationEntry, projectPath, sessionID, sessionFolderPath, agentID, userLabel, assistantLabel string) (string, error) {
	var sb strings.Builder

	// Compute basic stats from entries
	stats := ComputeSessionStats(entries, nil)
	stats.ProjectPath = projectPath
	stats.SessionID = sessionID
	stats.SessionFolderPath = sessionFolderPath

	// Determine page title based on whether this is a subagent query
	pageTitle := "Query Results"
	if agentID != "" {
		pageTitle = "Subagent Session"
	}

	// Build session folder link if we have a path
	sessionFolderName := extractSessionFolderName(stats.SessionFolderPath)
	if sessionFolderName == "" && projectPath != "" {
		sessionFolderName = extractSessionFolderName(projectPath)
	}
	sessionFolderLink := ""
	if stats.SessionFolderPath != "" {
		sessionFolderLink = renderFileLink(stats.SessionFolderPath, sessionFolderName, "folder-link")
	} else if sessionFolderName != "" {
		sessionFolderLink = escapeHTML(sessionFolderName)
	}

	// Write HTML doctype and head
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>`)
	sb.WriteString(escapeHTML(pageTitle))
	sb.WriteString(`</title>
    <style>`)
	sb.WriteString(GetStyleCSS())
	sb.WriteString(`
    </style>
</head>
<body>
`)

	// Write simplified header with dynamic title
	sb.WriteString(`<header class="page-header">
    <h1>`)
	sb.WriteString(escapeHTML(pageTitle))
	if sessionFolderLink != "" {
		sb.WriteString(`: `)
		sb.WriteString(sessionFolderLink)
	}
	sb.WriteString(`</h1>
    <div class="session-metadata">
`)

	// Session ID if available
	if sessionID != "" {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Session: %s</span>
`, renderSessionIDWithCopy(sessionID, projectPath, agentID)))
	}

	// Entry counts
	sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Entries: %d</span>
`, len(entries)))

	// Message type breakdown
	if stats.UserMessages > 0 || stats.AssistantMessages > 0 {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">%s: %d | %s: %d</span>
`, escapeHTML(userLabel), stats.UserMessages, escapeHTML(assistantLabel), stats.AssistantMessages))
	}

	sb.WriteString(`    </div>
</header>
`)

	// Write conversation entries
	sb.WriteString(`<div class="conversation">
`)

	// Track tool results for matching with tool calls
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		// Skip entries with no meaningful content
		if !hasContent(entry) {
			continue
		}

		entryHTML := renderEntry(entry, toolResults, projectPath, sessionID, agentID, userLabel, assistantLabel)
		sb.WriteString(entryHTML)
	}

	sb.WriteString("</div>\n")

	// Write simplified footer
	sb.WriteString(`<footer class="page-footer">
    <div class="footer-info">
        <p>Generated by <strong>claude-history</strong> query command</p>
`)
	if projectPath != "" {
		sb.WriteString(fmt.Sprintf(`        <p>Project: <code>%s</code></p>
`, escapeHTML(projectPath)))
	}
	sb.WriteString(`    </div>
</footer>
`)

	// Write JavaScript for interactivity
	sb.WriteString(`<script>`)
	sb.WriteString(GetScriptJS())
	sb.WriteString(`
</script>
<script>`)
	sb.WriteString(GetClipboardJS())
	sb.WriteString(`
</script>
</body>
</html>
`)

	return sb.String(), nil
}

// RenderConversationWithStats generates a complete HTML page for a conversation with session statistics.
// entries contains the conversation history, agents contains the agent hierarchy,
// stats contains optional session statistics for the header (if nil, stats are computed from entries/agents).
// This function uses "User" and "Assistant" as role labels for full session exports.
func RenderConversationWithStats(entries []models.ConversationEntry, agents []*agent.TreeNode, stats *SessionStats) (string, error) {
	var sb strings.Builder

	// Calculate stats if not provided
	if stats == nil {
		stats = ComputeSessionStats(entries, agents)
	}

	// Build a map of agent IDs to entry counts for subagent display and tooltip
	agentMap := buildAgentMap(agents)

	// Write HTML header with metadata and agent details
	sb.WriteString(renderHTMLHeader(stats, agentMap))

	// Write conversation entries
	sb.WriteString(`<div class="conversation">` + "\n")

	// Track tool results for matching with tool calls
	toolResults := buildToolResultsMap(entries)

	for _, entry := range entries {
		// Skip entries with no meaningful content
		if !hasContent(entry) {
			// Still render subagent placeholder if this entry spawned one
			if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
				subagentHTML := renderSubagentPlaceholder(entry.AgentID, agentMap, stats.SessionID, stats.ProjectPath)
				sb.WriteString(subagentHTML)
			}
			continue
		}

		// For full conversation exports, pass empty strings for sessionID/agentID (not a filtered query)
		entryHTML := renderEntry(entry, toolResults, stats.ProjectPath, "", "", "User", "Assistant")
		sb.WriteString(entryHTML)

		// Check if this entry spawned a subagent
		if entry.Type == models.EntryTypeQueueOperation && entry.AgentID != "" {
			subagentHTML := renderSubagentPlaceholder(entry.AgentID, agentMap, stats.SessionID, stats.ProjectPath)
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

// truncateID truncates an ID to the specified length.
// Used for displaying shortened IDs in the UI while preserving full IDs in copy operations.
// This prevents ID collision issues (birthday paradox) by keeping full IDs in clipboard.
func truncateID(id string, length int) string {
	if len(id) <= length {
		return id
	}
	return id[:length]
}

// TruncateSessionID returns a truncated session ID for display (first 8 chars).
// Deprecated: Use truncateID instead for consistency.
func TruncateSessionID(sessionID string) string {
	return truncateID(sessionID, 8)
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

		// RenderAgentFragment doesn't have access to ProjectPath or session context
		// Use "User"/"Assistant" labels for agent fragments (they're viewed in context of the full export)
		// Pass empty strings for sessionID/agentID since this is used for lazy-loaded fragments
		entryHTML := renderEntry(entry, toolResults, "", "", "", "User", "Assistant")
		sb.WriteString(entryHTML)
	}

	return sb.String(), nil
}

// hasContent checks if an entry has meaningful content worth rendering.
// Returns false for empty messages, true if the entry has text, tool calls, or tool results.
func hasContent(entry models.ConversationEntry) bool {
	// Check for text content with aggressive whitespace trimming
	textContent := entry.GetTextContent()
	trimmedText := strings.TrimSpace(textContent)

	// Check if text contains only whitespace characters (newlines, tabs, spaces)
	// Even if there's "content", if it's all whitespace, skip it
	if trimmedText != "" {
		// Additional check: is the trimmed text just repeated newlines or spaces?
		if strings.Trim(trimmedText, "\n\r\t ") != "" {
			return true
		}
	}

	// For assistant messages with tool calls but NO text:
	// Keep them so we can display them with "TOOL: X" header.
	// The renderEntry function will format these specially.
	if entry.Type == models.EntryTypeAssistant {
		tools := entry.ExtractToolCalls()
		if len(tools) > 0 {
			// Keep tool-only messages - they'll get special formatting
			return true
		}
	}

	// For user messages, tool results are NOT rendered in the HTML output.
	// Tool results only appear paired with tool calls in assistant messages.
	// Therefore, user messages with ONLY tool results (no text) should be filtered out.
	// This prevents empty message bubbles from appearing in the conversation.

	// Note: We don't check for tool results in user messages here because:
	// 1. Tool results in user messages are not displayed in the HTML output
	// 2. If a user message has actual text content, it's already caught above
	// 3. User messages with only tool results would create empty bubbles

	// Queue operations without content can be skipped (we still render subagent placeholder)
	// Summary entries without content should be skipped
	// Other entry types with no text should be skipped
	return false
}

// renderEntry renders a single conversation entry as HTML using the chat bubble layout.
// projectPath is used for generating CLI commands in task notifications (can be empty string if not available).
// sessionID and agentID are used to determine the correct agent ID to display in message headers:
//   - For main session queries (agentID == ""): show entry.AgentID if present
//   - For subagent queries (agentID != ""):
//     - ORCHESTRATOR/User messages: show sessionID (parent)
//     - AGENT/Assistant messages: show agentID (subagent)
// userLabel and assistantLabel specify the role names to display (e.g., "User"/"Assistant" or "Orchestrator"/"Agent").
func renderEntry(entry models.ConversationEntry, toolResults map[string]models.ToolResult, projectPath, sessionID, agentID, userLabel, assistantLabel string) string {
	var sb strings.Builder

	// Get text content
	textContent := entry.GetTextContent()

	// Detect task-notification blocks and render with flattened structure
	isTaskNotif := entry.Type == models.EntryTypeUser && strings.Contains(textContent, "<task-notification>")
	if isTaskNotif {
		taskNotif := parseTaskNotification(textContent)
		return renderFlatTaskNotification(taskNotif, entry, projectPath)
	}

	entryType := entry.Type
	roleLabel := getRoleLabel(entry.Type, userLabel, assistantLabel)
	entryClass := getEntryClass(entryType)
	timestamp := formatTimestampReadable(entry.Timestamp)

	// Check if this is a tool-only message (assistant message with no text, only tool calls)
	hasText := strings.TrimSpace(textContent) != ""
	toolCalls := entry.ExtractToolCalls()
	hasTools := len(toolCalls) > 0
	isToolOnly := entry.Type == models.EntryTypeAssistant && !hasText && hasTools

	// Build tool summary for header if this is a tool-only message
	toolSummary := ""
	if isToolOnly && len(toolCalls) > 0 {
		primaryTool := toolCalls[0]
		roleLabel = fmt.Sprintf("TOOL: %s", primaryTool.Name)

		// Extract display value for common tools
		displayValue := extractToolDisplayValue(primaryTool.Name, primaryTool.Input)
		if displayValue != "" {
			// Truncate if too long for inline display
			const maxInlineLen = 60
			if len(displayValue) > maxInlineLen {
				displayValue = displayValue[:maxInlineLen-3] + "..."
			}
			toolSummary = displayValue
		}
	}

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

	// Apply special styling for tool-only messages
	roleClass := "role"
	if isToolOnly {
		roleClass = "role tool-only-label"
	}
	sb.WriteString(fmt.Sprintf(`<span class="%s">%s</span>`, roleClass, escapeHTML(roleLabel)))

	// Add inline tool summary if present
	if toolSummary != "" {
		sb.WriteString(fmt.Sprintf(`<span class="tool-summary-inline">%s</span>`, escapeHTML(toolSummary)))
	}

	// Determine which agent ID to display
	displayAgentID := determineDisplayAgentID(entry, sessionID, agentID)
	if displayAgentID != "" {
		sb.WriteString(renderAgentIDWithCopy(entry, displayAgentID, sessionID, agentID, projectPath, roleLabel))
	}

	sb.WriteString(fmt.Sprintf(` <span class="timestamp">%s</span>`, escapeHTML(timestamp)))
	sb.WriteString("</div>\n")

	// Message content
	sb.WriteString(`    <div class="message-content">`)

	if textContent != "" {
		if entry.Type == models.EntryTypeAssistant {
			// Apply markdown rendering for assistant messages (with file path detection)
			sb.WriteString(fmt.Sprintf(`<div class="text markdown-content">%s</div>`, RenderMarkdown(textContent, projectPath)))
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

// determineDisplayAgentID determines which agent ID should be displayed for a message.
// For main session queries (agentID == ""), it returns entry.AgentID.
// For subagent queries (agentID != ""):
//   - User/Orchestrator messages: return sessionID (the parent orchestrator)
//   - Assistant/Agent messages: return agentID (the subagent responding)
func determineDisplayAgentID(entry models.ConversationEntry, sessionID, agentID string) string {
	// If this isn't a subagent query, use the entry's agent ID
	if agentID == "" {
		return entry.AgentID
	}

	// For subagent queries, determine based on entry type
	switch entry.Type {
	case models.EntryTypeUser:
		// User/Orchestrator messages come from the parent session
		return sessionID
	case models.EntryTypeAssistant:
		// Assistant/Agent messages come from the subagent
		return agentID
	default:
		// For other types, use entry's agent ID if available
		return entry.AgentID
	}
}

// buildSessionCopyContext builds the full context string for copying session information.
// This includes the session ID, project path, and a CLI command to query the session.
func buildSessionCopyContext(sessionID, projectPath, agentID string) string {
	if sessionID == "" {
		return sessionID
	}

	var sb strings.Builder

	// Session identification
	sb.WriteString(fmt.Sprintf("Session: %s\n", sessionID))

	// Add project path if available
	if projectPath != "" {
		sb.WriteString(fmt.Sprintf("Project: %s\n", projectPath))
	}

	// Build CLI command
	pathArg := projectPath
	if pathArg == "" {
		pathArg = "/path/to/project"
	}

	if agentID != "" {
		// If viewing a subagent, include agent ID in command
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s --agent %s", pathArg, sessionID, agentID))
	} else {
		// Otherwise, just session query
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s", pathArg, sessionID))
	}

	return sb.String()
}

// buildAgentIDCopyContext builds the full context string for copying agent ID information.
// This includes the role, agent ID, and a CLI command to query that specific message stream.
func buildAgentIDCopyContext(entry models.ConversationEntry, displayAgentID, sessionID, agentID, projectPath, roleLabel string) string {
	if displayAgentID == "" {
		return ""
	}

	var sb strings.Builder

	// Message identification
	sb.WriteString(fmt.Sprintf("%s Message: %s\n", roleLabel, displayAgentID))

	// Add session context
	if sessionID != "" {
		sb.WriteString(fmt.Sprintf("Session: %s\n", sessionID))
	}

	// Add project path if available
	if projectPath != "" {
		sb.WriteString(fmt.Sprintf("Project: %s\n", projectPath))
	}

	// Build CLI command
	pathArg := projectPath
	if pathArg == "" {
		pathArg = "/path/to/project"
	}

	// Determine which query command to suggest
	if agentID != "" && entry.Type == models.EntryTypeUser {
		// For orchestrator messages in a subagent query, suggest querying the main session
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s", pathArg, sessionID))
	} else if agentID != "" && entry.Type == models.EntryTypeAssistant {
		// For agent messages in a subagent query, suggest querying this subagent
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s --agent %s", pathArg, sessionID, agentID))
	} else if sessionID != "" && displayAgentID != sessionID {
		// For messages with a different agent ID in a main session query
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s --agent %s", pathArg, sessionID, displayAgentID))
	} else if sessionID != "" {
		// For main session messages
		sb.WriteString(fmt.Sprintf("claude-history query %s --session %s", pathArg, sessionID))
	}

	return sb.String()
}

// getRoleLabel returns a human-readable label for the entry type.
// userLabel and assistantLabel specify custom role names (e.g., "Orchestrator"/"Agent" for subagent contexts).
func getRoleLabel(entryType models.EntryType, userLabel, assistantLabel string) string {
	switch entryType {
	case models.EntryTypeUser:
		return userLabel
	case models.EntryTypeAssistant:
		return assistantLabel
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
// sessionID and projectPath are used to build the full copy context with CLI commands.
func renderSubagentPlaceholder(agentID string, agentMap map[string]int, sessionID, projectPath string) string {
	var sb strings.Builder

	entryCount := agentMap[agentID]
	shortID := truncateID(agentID, 7)

	sb.WriteString(fmt.Sprintf(`<div class="subagent collapsible collapsed" data-agent-id="%s">`, escapeHTML(agentID)))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`  <div class="subagent-header collapsible-trigger" onclick="loadAgent(this)"><span class="subagent-title">Subagent: %s</span> <span class="subagent-meta">(%d entries)</span>%s<span class="chevron down">▼</span></div>`,
		escapeHTML(shortID),
		entryCount,
		renderSubagentBadgeWithCopy(agentID, sessionID, projectPath)))
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

// ============================================================================
// Helper Functions for DRY (Don't Repeat Yourself)
// ============================================================================
// These functions provide a single source of truth for common HTML rendering
// patterns, ensuring consistency across the codebase.

// renderSessionIDWithCopy renders a session ID badge with truncated display and copy button.
// The display shows only the first 8 chars for clean UI, but the copy button includes
// full context (session ID, project path, and CLI command) to prevent ID collisions.
func renderSessionIDWithCopy(sessionID, projectPath, agentID string) string {
	if sessionID == "" {
		return ""
	}

	truncatedID := truncateID(sessionID, 8)
	copyContext := buildSessionCopyContext(sessionID, projectPath, agentID)

	return fmt.Sprintf(`<code>%s</code>%s`,
		escapeHTML(truncatedID),
		renderCopyButton(copyContext, "session-id", "Copy session details"))
}

// renderAgentIDWithCopy renders an agent ID badge with truncated display and copy button.
// The display shows only the first 8 chars for clean UI, but the copy button includes
// full context (role, agent ID, session, and CLI command) to prevent ID collisions.
func renderAgentIDWithCopy(entry models.ConversationEntry, displayAgentID, sessionID, agentID, projectPath, roleLabel string) string {
	if displayAgentID == "" {
		return ""
	}

	truncatedID := truncateID(displayAgentID, 8)
	copyContext := buildAgentIDCopyContext(entry, displayAgentID, sessionID, agentID, projectPath, roleLabel)

	return fmt.Sprintf(`<span class="agent-id-badge">%s%s</span>`,
		escapeHTML(truncatedID),
		renderCopyButton(copyContext, "agent-id", "Copy agent details"))
}

// renderSubagentBadgeWithCopy renders a subagent placeholder badge with copy button.
// Used in subagent placeholder sections to show agent ID and provide copy functionality.
// The copy button includes the full agent ID to prevent collisions.
func renderSubagentBadgeWithCopy(agentID, sessionID, projectPath string) string {
	if agentID == "" {
		return ""
	}

	// Build full copy context for subagent
	var copyText strings.Builder
	copyText.WriteString(fmt.Sprintf("Subagent: %s\n", agentID))

	if sessionID != "" {
		copyText.WriteString(fmt.Sprintf("Session: %s\n", sessionID))
	}

	if projectPath != "" {
		copyText.WriteString(fmt.Sprintf("Project: %s\n", projectPath))
	}

	// Build CLI command
	pathArg := projectPath
	if pathArg == "" {
		pathArg = "/path/to/project"
	}

	if sessionID != "" {
		copyText.WriteString(fmt.Sprintf("claude-history query %s --session %s --agent %s", pathArg, sessionID, agentID))
	}

	return renderCopyButton(copyText.String(), "agent-id", "Copy agent details")
}

// renderFileLink renders a clickable file:// link for opening files in Finder/Explorer.
func renderFileLink(path, displayText, cssClass string) string {
	if path == "" {
		return escapeHTML(displayText)
	}

	fileURL := buildFileURL(path)
	return fmt.Sprintf(`<a href="%s" class="%s" title="Open in Finder/Explorer">%s</a>`,
		escapeHTML(fileURL), escapeHTML(cssClass), escapeHTML(displayText))
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
// agentDetails is an optional map of agent IDs to message counts for the interactive tooltip.
func renderHTMLHeader(stats *SessionStats, agentDetails map[string]int) string {
	var sb strings.Builder

	// Build session folder link if we have a path
	sessionFolderName := ""
	sessionFolderLink := ""
	if stats != nil {
		sessionFolderName = extractSessionFolderName(stats.SessionFolderPath)
		if sessionFolderName == "" && stats.ProjectPath != "" {
			sessionFolderName = extractSessionFolderName(stats.ProjectPath)
		}
		if stats.SessionFolderPath != "" {
			sessionFolderLink = renderFileLink(stats.SessionFolderPath, sessionFolderName, "folder-link")
		} else if sessionFolderName != "" {
			sessionFolderLink = escapeHTML(sessionFolderName)
		}
	}

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Claude Code Session</title>
    <link rel="stylesheet" href="static/style.css">
</head>
<body>
<header class="page-header">
    <h1>Claude Code Session`)
	if sessionFolderLink != "" {
		sb.WriteString(`: `)
		sb.WriteString(sessionFolderLink)
	}
	sb.WriteString(`</h1>
    <div class="session-metadata">
`)

	// Session ID with copy button
	if stats != nil && stats.SessionID != "" {
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">Session: %s</span>
`, renderSessionIDWithCopy(stats.SessionID, stats.ProjectPath, "")))
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

	// Enhanced message statistics with interactive agent tooltip
	if stats != nil {
		// Encode agent details as JSON for JavaScript
		agentDetailsJSON := "{}"
		if len(agentDetails) > 0 {
			jsonBytes, err := json.Marshal(agentDetails)
			if err == nil {
				agentDetailsJSON = string(jsonBytes)
			}
		}

		// Build the statistics line with interactive agent tooltip
		sb.WriteString(fmt.Sprintf(`        <span class="meta-item">User: %d | Assistant: %d | `, stats.UserMessages, stats.AssistantMessages))

		// Add interactive agent stats span if there are agents
		if stats.AgentCount > 0 {
			sb.WriteString(fmt.Sprintf(`<span class="agent-stats-interactive" data-session-id="%s" data-agent-details='%s' title="Click to copy agent list">Subagents[%d]: %d messages</span>`,
				escapeHTML(stats.SessionID),
				escapeHTML(agentDetailsJSON),
				stats.AgentCount,
				stats.TotalAgentMessages))
		} else {
			sb.WriteString(fmt.Sprintf(`Subagents[%d]: %d messages`, stats.AgentCount, stats.TotalAgentMessages))
		}

		sb.WriteString("</span>\n")
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
    <script src="static/agent-tooltip.js"></script>
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

// renderFlatTaskNotification renders a task notification with flattened structure (2-level DOM).
// Returns a standalone notification-row div, not wrapped in message-row/bubble structure.
// projectPath is used for generating CLI commands (can be empty string if not available).
func renderFlatTaskNotification(taskNotif *TaskNotificationData, entry models.ConversationEntry, projectPath string) string {
	if taskNotif == nil {
		// Fallback to empty string
		return ""
	}

	var sb strings.Builder

	// Status icon and class
	statusIcon := "⏳"
	statusClass := "running"
	switch taskNotif.Status {
	case "completed":
		statusIcon = "✓"
		statusClass = "completed"
	case "failed", "error":
		statusIcon = "✗"
		statusClass = "failed"
	}

	// Build CLI command for tooltip (if we have session + agent info)
	cliCommand := ""
	if entry.SessionID != "" && taskNotif.TaskID != "" {
		// Use projectPath if available, otherwise use placeholder
		pathArg := projectPath
		if pathArg == "" {
			pathArg = "<project-path>"
		}
		cliCommand = fmt.Sprintf("claude-history query %s --session %s --agent %s",
			pathArg, entry.SessionID, taskNotif.TaskID)
	}

	// Main notification row
	sb.WriteString(fmt.Sprintf(`<div class="notification-row %s" data-uuid="%s">`, statusClass, escapeHTML(entry.UUID)))
	sb.WriteString("\n")

	// Collapsible header (single line)
	sb.WriteString(`  <div class="notification-header" aria-expanded="true">`)
	sb.WriteString("\n")

	// Collapse toggle
	sb.WriteString(`    <button class="collapse-toggle" aria-label="Toggle notification">▼</button>`)
	sb.WriteString("\n")

	// Notification type
	sb.WriteString(`    <span class="notification-type">Subagent</span>`)
	sb.WriteString("\n")

	// Summary/description with status icon
	sb.WriteString(fmt.Sprintf(`    <span class="notification-summary">%s %s</span>`,
		statusIcon, escapeHTML(taskNotif.Summary)))
	sb.WriteString("\n")

	// Agent/Task ID badge with tooltip and copy button
	if taskNotif.TaskID != "" {
		tooltipText := cliCommand
		if tooltipText == "" {
			tooltipText = "Agent ID: " + taskNotif.TaskID
		}

		// Build full copy text with context
		copyText := ""
		if cliCommand != "" {
			copyText = fmt.Sprintf("Subagent \"%s\" %s\n%s",
				taskNotif.Summary,
				taskNotif.TaskID,
				cliCommand)
		} else {
			copyText = fmt.Sprintf("Subagent \"%s\" %s\nAgent ID: %s",
				taskNotif.Summary,
				taskNotif.TaskID,
				taskNotif.TaskID)
		}

		truncatedID := truncateID(taskNotif.TaskID, 8)
		sb.WriteString(fmt.Sprintf(`    <span class="agent-id-badge" data-full-id="%s" title="%s">`,
			escapeHTML(taskNotif.TaskID), escapeHTML(tooltipText)))
		sb.WriteString(escapeHTML(truncatedID))
		sb.WriteString(renderCopyButton(copyText, "agent-notification", "Copy agent details"))
		sb.WriteString(`</span>`)
		sb.WriteString("\n")
	}

	// Timestamp
	if entry.Timestamp != "" {
		sb.WriteString(fmt.Sprintf(`    <span class="timestamp">%s</span>`,
			formatTimestampReadable(entry.Timestamp)))
		sb.WriteString("\n")
	}

	sb.WriteString(`  </div>`)
	sb.WriteString("\n")

	// Collapsible content
	if taskNotif.Result != "" {
		sb.WriteString(`  <div class="notification-content">`)
		sb.WriteString("\n")

		sb.WriteString(fmt.Sprintf(`    <div class="notification-result">%s</div>`,
			RenderMarkdown(taskNotif.Result, projectPath)))
		sb.WriteString("\n")

		sb.WriteString(`  </div>`)
		sb.WriteString("\n")
	}

	sb.WriteString(`</div>`)
	sb.WriteString("\n")

	return sb.String()
}

// extractSessionFolderName extracts the last component of a path (session folder name).
// For example: "/Users/name/project" -> "project"
// Windows paths like "C:\Users\name\project" -> "project"
func extractSessionFolderName(path string) string {
	if path == "" {
		return ""
	}
	// Use filepath.Base to get the last component (cross-platform)
	return filepath.Base(path)
}

// buildFileURL builds a file:// URL from an absolute path.
// Uses forward slashes for consistency across platforms.
func buildFileURL(path string) string {
	if path == "" {
		return ""
	}
	// Convert backslashes to forward slashes for Windows
	path = strings.ReplaceAll(path, "\\", "/")
	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "file://" + path
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
