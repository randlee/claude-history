// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"fmt"
	"strings"
)

// BreadcrumbItem represents a single item in the breadcrumb trail.
type BreadcrumbItem struct {
	ID    string // The agent ID or "main" for main session
	Label string // Display label for the breadcrumb
}

// RenderBreadcrumbs generates HTML for a breadcrumb navigation trail.
// The last item is marked as active.
func RenderBreadcrumbs(items []BreadcrumbItem) string {
	if len(items) == 0 {
		// Default breadcrumb with just main session
		items = []BreadcrumbItem{{ID: "main", Label: "Main Session"}}
	}

	var sb strings.Builder

	sb.WriteString(`<nav class="breadcrumbs" id="breadcrumbs" aria-label="Navigation breadcrumbs">`)
	sb.WriteString("\n")

	for i, item := range items {
		isLast := i == len(items)-1

		if isLast {
			// Active item (current location)
			sb.WriteString(fmt.Sprintf(`  <a href="#%s" class="breadcrumb-item active" data-agent-id="%s" aria-current="page">%s</a>`,
				escapeHTML(item.ID),
				escapeHTML(item.ID),
				escapeHTML(item.Label)))
		} else {
			// Clickable navigation item
			sb.WriteString(fmt.Sprintf(`  <a href="#%s" class="breadcrumb-item" data-agent-id="%s">%s</a>`,
				escapeHTML(item.ID),
				escapeHTML(item.ID),
				escapeHTML(item.Label)))
		}
		sb.WriteString("\n")

		// Add separator except after last item
		if !isLast {
			sb.WriteString(`  <span class="breadcrumb-separator" aria-hidden="true">&#8250;</span>`)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("</nav>\n")

	return sb.String()
}

// RenderJumpToParentButton generates HTML for a "Jump to Parent" button.
// agentID is the current agent's ID, used to determine the parent.
func RenderJumpToParentButton(agentID string) string {
	if agentID == "" || agentID == "main" {
		return "" // No parent for main session
	}

	return fmt.Sprintf(
		`<button class="jump-to-parent-btn" data-agent-id="%s" title="Jump to parent agent (Alt+Up)" type="button">Parent</button>`,
		escapeHTML(agentID),
	)
}

// RenderAgentContainer generates HTML wrapper for an agent section with proper ID for navigation.
// agentID is the agent's unique identifier.
// depth indicates the nesting level (0 for main session, 1+ for nested agents).
// content is the inner HTML content.
func RenderAgentContainer(agentID string, depth int, content string) string {
	var sb strings.Builder

	// Determine CSS classes based on depth
	depthClass := ""
	if depth > 0 {
		depthClass = fmt.Sprintf(` data-depth="%d"`, depth)
	}

	sb.WriteString(fmt.Sprintf(`<div class="agent-container" id="agent-%s"%s>`,
		escapeHTML(agentID),
		depthClass))
	sb.WriteString("\n")

	// Add jump to parent if not main session
	if agentID != "main" && depth > 0 {
		sb.WriteString(`  <div class="agent-header-controls">`)
		sb.WriteString("\n")
		sb.WriteString("    ")
		sb.WriteString(RenderJumpToParentButton(agentID))
		sb.WriteString("\n")
		sb.WriteString("  </div>\n")
	}

	sb.WriteString(content)
	sb.WriteString("\n</div>\n")

	return sb.String()
}

// RenderNestedSubagentOverlay generates HTML for a subagent overlay with navigation support.
// This is an enhanced version of RenderSubagentOverlay that includes navigation elements.
func RenderNestedSubagentOverlay(agentID string, entryCount int, depth int, metadata map[string]string) string {
	var sb strings.Builder

	shortID := agentID
	if len(shortID) > 7 {
		shortID = shortID[:7]
	}

	// Determine depth attribute
	depthAttr := ""
	if depth > 0 {
		depthAttr = fmt.Sprintf(` data-depth="%d"`, depth)
	}

	// Main overlay container with navigation attributes
	sb.WriteString(fmt.Sprintf(`<div class="subagent-overlay agent-overlay collapsible" data-agent-id="%s" id="agent-%s"%s>`,
		escapeHTML(agentID),
		escapeHTML(agentID),
		depthAttr))
	sb.WriteString("\n")

	// Header (collapsible trigger)
	sb.WriteString(`  <div class="subagent-header overlay-header collapsible-trigger">`)
	sb.WriteString("\n")
	sb.WriteString(`    <span class="agent-icon">`)
	sb.WriteString("\xf0\x9f\xa4\x96") // robot emoji
	sb.WriteString(`</span>`)
	sb.WriteString(fmt.Sprintf(`<span class="subagent-title">Subagent: %s</span>`, escapeHTML(shortID)))
	sb.WriteString(fmt.Sprintf(`<span class="subagent-meta">(%d entries)</span>`, entryCount))

	// Copy button for agent ID
	sb.WriteString(renderCopyButton(agentID, "agent-id", "Copy agent ID"))

	// Agent header controls (Deep Dive and Jump to Parent)
	sb.WriteString(`<div class="agent-header-controls">`)

	// Jump to Parent button (only if nested)
	if depth > 0 {
		sb.WriteString(RenderJumpToParentButton(agentID))
	}

	// Deep Dive button
	sb.WriteString(fmt.Sprintf(`<button class="deep-dive-btn" onclick="deepDiveAgent('%s', event)" title="Open agent conversation">`,
		escapeHTML(agentID)))
	sb.WriteString(`Deep Dive</button>`)

	sb.WriteString(`</div>`) // Close agent-header-controls

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

// GenerateBreadcrumbPath creates a breadcrumb path from a list of agent IDs.
// The first item is always the main session, followed by any nested agents.
func GenerateBreadcrumbPath(agentIDs []string) []BreadcrumbItem {
	items := []BreadcrumbItem{
		{ID: "main", Label: "Main Session"},
	}

	for _, id := range agentIDs {
		if id == "" || id == "main" {
			continue
		}

		// Truncate ID for display
		label := id
		if len(label) > 7 {
			label = label[:7]
		}

		items = append(items, BreadcrumbItem{
			ID:    id,
			Label: label,
		})
	}

	return items
}

// TruncateAgentID truncates an agent ID for display purposes.
// Returns the first 7 characters if longer.
func TruncateAgentID(agentID string) string {
	if len(agentID) > 7 {
		return agentID[:7]
	}
	return agentID
}
