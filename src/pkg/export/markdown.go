// Package export provides HTML export functionality for Claude Code conversation history.
package export

import (
	"fmt"
	"regexp"
	"strings"
)

// CodeBlock represents a fenced code block extracted from markdown.
type CodeBlock struct {
	Language string // The language tag (e.g., "go", "bash", "python")
	Code     string // The code content (without the fence markers)
	StartPos int    // Start position in the original text
	EndPos   int    // End position in the original text
}

// Regular expression patterns for markdown parsing
var (
	// Code blocks: ```lang\ncode\n```
	codeBlockRe = regexp.MustCompile("(?s)```(\\w*)\\n?(.*?)```")

	// Inline code: `code`
	inlineCodeRe = regexp.MustCompile("`([^`\n]+)`")

	// Headers: # through ######
	h1Re = regexp.MustCompile(`(?m)^# (.+)$`)
	h2Re = regexp.MustCompile(`(?m)^## (.+)$`)
	h3Re = regexp.MustCompile(`(?m)^### (.+)$`)
	h4Re = regexp.MustCompile(`(?m)^#### (.+)$`)
	h5Re = regexp.MustCompile(`(?m)^##### (.+)$`)
	h6Re = regexp.MustCompile(`(?m)^###### (.+)$`)

	// Bold and italic
	boldRe   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	italicRe = regexp.MustCompile(`\*([^*]+)\*`)

	// Links: [text](url)
	linkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	// Images: ![alt](url)
	imageRe = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	// Blockquotes: > text
	blockquoteRe = regexp.MustCompile(`(?m)^> (.+)$`)

	// Horizontal rules: ---, ***, ___
	hrRe = regexp.MustCompile(`(?m)^(---|\*\*\*|___)$`)

	// Task lists: - [ ] or - [x]
	taskUncheckedRe = regexp.MustCompile(`(?m)^(\s*)- \[ \] (.+)$`)
	taskCheckedRe   = regexp.MustCompile(`(?m)^(\s*)- \[x\] (.+)$`)

	// Unordered lists: - item or * item
	ulItemRe = regexp.MustCompile(`(?m)^(\s*)[-*] (.+)$`)

	// Ordered lists: 1. item
	olItemRe = regexp.MustCompile(`(?m)^(\s*)\d+\. (.+)$`)

	// Table patterns
	tableSeparatorRe = regexp.MustCompile(`^[\s|:-]+$`)
)

// ExtractCodeBlocks finds all fenced code blocks in the markdown text.
// Returns a slice of CodeBlock structs with language, code content, and positions.
func ExtractCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock

	matches := codeBlockRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		if len(match) >= 6 {
			// match[0]:match[1] = full match
			// match[2]:match[3] = language
			// match[4]:match[5] = code content
			lang := ""
			if match[2] != -1 && match[3] != -1 {
				lang = content[match[2]:match[3]]
			}
			code := ""
			if match[4] != -1 && match[5] != -1 {
				code = content[match[4]:match[5]]
			}
			// Trim trailing newline from code
			code = strings.TrimSuffix(code, "\n")

			blocks = append(blocks, CodeBlock{
				Language: lang,
				Code:     code,
				StartPos: match[0],
				EndPos:   match[1],
			})
		}
	}

	return blocks
}

// RenderMarkdown converts markdown text to HTML.
// Supports: headers (h1-h6), lists (ordered, unordered, nested), tables, blockquotes,
// code blocks (fenced and inline), links, images, bold, italic, task lists, and horizontal rules.
// Code blocks are rendered with language badges and copy buttons for enhanced UX.
// All plain text is HTML-escaped to prevent XSS attacks.
func RenderMarkdown(content string) string {
	if content == "" {
		return ""
	}

	// Store code blocks and replace with placeholders to protect them during processing
	codeBlocks := ExtractCodeBlocks(content)
	codeBlockPlaceholders := make(map[string]string)
	result := content

	// Replace code blocks with placeholders (process in reverse to maintain positions)
	for i := len(codeBlocks) - 1; i >= 0; i-- {
		block := codeBlocks[i]
		placeholder := fmt.Sprintf("\x00CODE_BLOCK_%d\x00", i)
		codeBlockPlaceholders[placeholder] = renderCodeBlock(block)
		result = result[:block.StartPos] + placeholder + result[block.EndPos:]
	}

	// Protect inline code and replace with placeholders
	inlineCodePlaceholders := make(map[string]string)
	inlineIdx := 0
	result = inlineCodeRe.ReplaceAllStringFunc(result, func(match string) string {
		code := inlineCodeRe.FindStringSubmatch(match)[1]
		placeholder := fmt.Sprintf("\x00INLINE_CODE_%d\x00", inlineIdx)
		inlineCodePlaceholders[placeholder] = `<code class="inline-code">` + escapeHTML(code) + `</code>`
		inlineIdx++
		return placeholder
	})

	// Process images before links (images have ! prefix)
	// Store rendered images in placeholders to protect URLs from escaping
	imagePlaceholders := make(map[string]string)
	imageIdx := 0
	result = imageRe.ReplaceAllStringFunc(result, func(match string) string {
		parts := imageRe.FindStringSubmatch(match)
		if len(parts) >= 3 {
			placeholder := fmt.Sprintf("\x00IMAGE_%d\x00", imageIdx)
			imagePlaceholders[placeholder] = `<img src="` + escapeHTML(parts[2]) + `" alt="` + escapeHTML(parts[1]) + `" class="md-image">`
			imageIdx++
			return placeholder
		}
		return match
	})

	// Process links and store in placeholders
	linkPlaceholders := make(map[string]string)
	linkIdx := 0
	result = linkRe.ReplaceAllStringFunc(result, func(match string) string {
		parts := linkRe.FindStringSubmatch(match)
		if len(parts) >= 3 {
			placeholder := fmt.Sprintf("\x00LINK_%d\x00", linkIdx)
			linkPlaceholders[placeholder] = `<a href="` + escapeHTML(parts[2]) + `" class="md-link">` + escapeHTML(parts[1]) + `</a>`
			linkIdx++
			return placeholder
		}
		return match
	})

	// Process tables (before escaping so we can detect the | delimiters)
	result = processMarkdownTables(result)

	// Process lists (before escaping so we can detect the - and * markers)
	result = processMarkdownLists(result)

	// Process blockquotes (before escaping so we can detect the > marker)
	result = processBlockquotes(result)

	// Process horizontal rules (before escaping)
	result = hrRe.ReplaceAllString(result, "\x00HR\x00")

	// Process headers (before escaping so we can detect the # markers)
	// Note: We don't escape here - escapeRemainingText() will handle it later
	result = h6Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h6Re.FindStringSubmatch(match)
		return `<h6 class="md-h6">` + parts[1] + `</h6>`
	})
	result = h5Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h5Re.FindStringSubmatch(match)
		return `<h5 class="md-h5">` + parts[1] + `</h5>`
	})
	result = h4Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h4Re.FindStringSubmatch(match)
		return `<h4 class="md-h4">` + parts[1] + `</h4>`
	})
	result = h3Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h3Re.FindStringSubmatch(match)
		return `<h3 class="md-h3">` + parts[1] + `</h3>`
	})
	result = h2Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h2Re.FindStringSubmatch(match)
		return `<h2 class="md-h2">` + parts[1] + `</h2>`
	})
	result = h1Re.ReplaceAllStringFunc(result, func(match string) string {
		parts := h1Re.FindStringSubmatch(match)
		return `<h1 class="md-h1">` + parts[1] + `</h1>`
	})

	// Process bold and italic (without escaping - escapeRemainingText handles it)
	result = boldRe.ReplaceAllStringFunc(result, func(match string) string {
		parts := boldRe.FindStringSubmatch(match)
		return `<strong>` + parts[1] + `</strong>`
	})
	result = italicRe.ReplaceAllStringFunc(result, func(match string) string {
		parts := italicRe.FindStringSubmatch(match)
		return `<em>` + parts[1] + `</em>`
	})

	// Now escape any remaining plain text that wasn't processed
	// We need to be careful not to escape our placeholders or HTML tags we've already created
	result = escapeRemainingText(result)

	// Restore horizontal rule placeholder
	result = strings.ReplaceAll(result, "\x00HR\x00", `<hr class="md-hr">`)

	// Convert remaining newlines to <br> for proper display (but not after block elements)
	result = convertNewlinesToBr(result)

	// Restore all placeholders
	for placeholder, html := range imagePlaceholders {
		result = strings.ReplaceAll(result, placeholder, html)
	}
	for placeholder, html := range linkPlaceholders {
		result = strings.ReplaceAll(result, placeholder, html)
	}
	for placeholder, html := range inlineCodePlaceholders {
		result = strings.ReplaceAll(result, placeholder, html)
	}
	for placeholder, html := range codeBlockPlaceholders {
		result = strings.ReplaceAll(result, placeholder, html)
	}

	return result
}

// escapeRemainingText escapes HTML in text that hasn't been processed as markdown.
// It preserves HTML tags that we've already created and placeholder markers.
func escapeRemainingText(content string) string {
	var result strings.Builder
	i := 0
	for i < len(content) {
		// Check for placeholder markers (null byte)
		if content[i] == '\x00' {
			// Find the end of the placeholder
			end := strings.Index(content[i+1:], "\x00")
			if end != -1 {
				result.WriteString(content[i : i+end+2])
				i += end + 2
				continue
			}
		}

		// Check for HTML tags we've created
		if content[i] == '<' {
			// Find the end of the tag
			tagEnd := strings.Index(content[i:], ">")
			if tagEnd != -1 {
				// Check if this looks like a valid HTML tag
				tagContent := content[i : i+tagEnd+1]
				if isValidHTMLTag(tagContent) {
					result.WriteString(tagContent)
					i += tagEnd + 1
					continue
				}
			}
		}

		// Escape this character if it's a special HTML character
		switch content[i] {
		case '&':
			result.WriteString("&amp;")
		case '<':
			result.WriteString("&lt;")
		case '>':
			result.WriteString("&gt;")
		case '"':
			result.WriteString("&#34;")
		case '\'':
			result.WriteString("&#39;")
		default:
			result.WriteByte(content[i])
		}
		i++
	}
	return result.String()
}

// isValidHTMLTag checks if a string looks like an HTML tag we've generated.
func isValidHTMLTag(tag string) bool {
	if len(tag) < 3 {
		return false
	}
	if tag[0] != '<' || tag[len(tag)-1] != '>' {
		return false
	}

	// List of tags we generate
	validTags := []string{
		"<h1", "<h2", "<h3", "<h4", "<h5", "<h6",
		"</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>",
		"<ul", "</ul>", "<ol", "</ol>", "<li", "</li>",
		"<table", "</table>", "<thead>", "</thead>", "<tbody>", "</tbody>",
		"<tr>", "</tr>", "<th>", "</th>", "<td>", "</td>",
		"<blockquote", "</blockquote>",
		"<strong>", "</strong>", "<em>", "</em>",
		"<input",
	}

	for _, valid := range validTags {
		if strings.HasPrefix(tag, valid) {
			return true
		}
	}
	return false
}

// renderCodeBlock renders a fenced code block with language badge and copy button.
func renderCodeBlock(block CodeBlock) string {
	var sb strings.Builder

	languageClass := ""
	languageDisplay := "text"
	if block.Language != "" {
		languageClass = " language-" + escapeHTML(block.Language)
		languageDisplay = escapeHTML(block.Language)
	}

	sb.WriteString(`<div class="code-block` + languageClass + `">`)
	sb.WriteString(`<div class="code-header">`)
	sb.WriteString(`<span class="language-badge">` + languageDisplay + `</span>`)
	sb.WriteString(`<button class="copy-code-btn" onclick="copyCode(this)" title="Copy code">Copy</button>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`<pre class="code-content"><code>` + escapeHTML(block.Code) + `</code></pre>`)
	sb.WriteString(`</div>`)

	return sb.String()
}

// processMarkdownTables converts markdown tables to HTML tables.
func processMarkdownTables(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inTable := false
	var tableRows []string

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check if this line is a table row
		if strings.HasPrefix(trimmedLine, "|") && strings.HasSuffix(trimmedLine, "|") {
			if !inTable {
				inTable = true
				tableRows = []string{}
			}
			tableRows = append(tableRows, line)
		} else {
			// Not a table row
			if inTable {
				// End of table, render it
				tableHTML := renderTable(tableRows)
				result = append(result, tableHTML)
				inTable = false
				tableRows = nil
			}
			result = append(result, line)
		}

		// Handle end of input
		if i == len(lines)-1 && inTable {
			tableHTML := renderTable(tableRows)
			result = append(result, tableHTML)
		}
	}

	return strings.Join(result, "\n")
}

// renderTable converts table rows to HTML table.
func renderTable(rows []string) string {
	if len(rows) < 2 {
		// Not a valid table (needs header + separator at minimum)
		return strings.Join(rows, "\n")
	}

	var sb strings.Builder
	sb.WriteString(`<table class="md-table">`)

	for i, row := range rows {
		cells := parseTableRow(row)
		if len(cells) == 0 {
			continue
		}

		// Check if this is the separator row
		if i == 1 && isTableSeparator(row) {
			continue
		}

		if i == 0 {
			// Header row
			sb.WriteString(`<thead><tr>`)
			for _, cell := range cells {
				// Don't escape here - escapeRemainingText() will handle it
				sb.WriteString(`<th>` + strings.TrimSpace(cell) + `</th>`)
			}
			sb.WriteString(`</tr></thead><tbody>`)
		} else {
			// Body row
			sb.WriteString(`<tr>`)
			for _, cell := range cells {
				// Don't escape here - escapeRemainingText() will handle it
				sb.WriteString(`<td>` + strings.TrimSpace(cell) + `</td>`)
			}
			sb.WriteString(`</tr>`)
		}
	}

	sb.WriteString(`</tbody></table>`)
	return sb.String()
}

// parseTableRow extracts cells from a markdown table row.
func parseTableRow(row string) []string {
	// Remove leading/trailing pipes and split
	trimmed := strings.Trim(strings.TrimSpace(row), "|")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "|")
}

// isTableSeparator checks if a row is a table separator (|---|---|).
func isTableSeparator(row string) bool {
	trimmed := strings.Trim(strings.TrimSpace(row), "|")
	return tableSeparatorRe.MatchString(trimmed)
}

// processMarkdownLists processes unordered, ordered, and task lists.
func processMarkdownLists(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var listStack []string // stack of "ul", "ol", or "task"
	var indentStack []int  // stack of indent levels

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Calculate indent level
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Check for task list items first (they're a special case of unordered lists)
		// Don't escape content here - escapeRemainingText() will handle it
		if taskMatch := taskUncheckedRe.FindStringSubmatch(line); taskMatch != nil {
			result = handleListItem(result, &listStack, &indentStack, indent, "task", `<li class="task-item"><input type="checkbox" disabled> `+taskMatch[2]+`</li>`)
			continue
		}
		if taskMatch := taskCheckedRe.FindStringSubmatch(line); taskMatch != nil {
			result = handleListItem(result, &listStack, &indentStack, indent, "task", `<li class="task-item"><input type="checkbox" checked disabled> `+taskMatch[2]+`</li>`)
			continue
		}

		// Check for unordered list items
		if ulMatch := ulItemRe.FindStringSubmatch(line); ulMatch != nil && !strings.HasPrefix(trimmedLine, "- [ ]") && !strings.HasPrefix(trimmedLine, "- [x]") {
			result = handleListItem(result, &listStack, &indentStack, indent, "ul", `<li>`+ulMatch[2]+`</li>`)
			continue
		}

		// Check for ordered list items
		if olMatch := olItemRe.FindStringSubmatch(line); olMatch != nil {
			result = handleListItem(result, &listStack, &indentStack, indent, "ol", `<li>`+olMatch[2]+`</li>`)
			continue
		}

		// Not a list item - close all open lists
		for len(listStack) > 0 {
			listType := listStack[len(listStack)-1]
			listStack = listStack[:len(listStack)-1]
			indentStack = indentStack[:len(indentStack)-1]
			result = append(result, closeListTag(listType))
		}
		result = append(result, line)
	}

	// Close any remaining open lists
	for len(listStack) > 0 {
		listType := listStack[len(listStack)-1]
		listStack = listStack[:len(listStack)-1]
		result = append(result, closeListTag(listType))
	}

	return strings.Join(result, "\n")
}

// handleListItem handles adding a list item and managing the list stack.
func handleListItem(result []string, listStack *[]string, indentStack *[]int, indent int, listType string, itemHTML string) []string {
	// If stack is empty, start a new list
	if len(*listStack) == 0 {
		*listStack = append(*listStack, listType)
		*indentStack = append(*indentStack, indent)
		result = append(result, openListTag(listType))
		result = append(result, itemHTML)
		return result
	}

	currentIndent := (*indentStack)[len(*indentStack)-1]
	currentType := (*listStack)[len(*listStack)-1]

	if indent > currentIndent {
		// Nested list - start new list
		*listStack = append(*listStack, listType)
		*indentStack = append(*indentStack, indent)
		result = append(result, openListTag(listType))
		result = append(result, itemHTML)
	} else if indent < currentIndent {
		// Dedented - close lists until we match or exceed
		for len(*listStack) > 0 && (*indentStack)[len(*indentStack)-1] > indent {
			closingType := (*listStack)[len(*listStack)-1]
			*listStack = (*listStack)[:len(*listStack)-1]
			*indentStack = (*indentStack)[:len(*indentStack)-1]
			result = append(result, closeListTag(closingType))
		}
		// Check if we need to switch list types
		if len(*listStack) > 0 && (*listStack)[len(*listStack)-1] != listType {
			closingType := (*listStack)[len(*listStack)-1]
			*listStack = (*listStack)[:len(*listStack)-1]
			*indentStack = (*indentStack)[:len(*indentStack)-1]
			result = append(result, closeListTag(closingType))
			*listStack = append(*listStack, listType)
			*indentStack = append(*indentStack, indent)
			result = append(result, openListTag(listType))
		}
		result = append(result, itemHTML)
	} else {
		// Same indent level
		if currentType != listType {
			// Different list type at same level - close current and open new
			*listStack = (*listStack)[:len(*listStack)-1]
			*indentStack = (*indentStack)[:len(*indentStack)-1]
			result = append(result, closeListTag(currentType))
			*listStack = append(*listStack, listType)
			*indentStack = append(*indentStack, indent)
			result = append(result, openListTag(listType))
		}
		result = append(result, itemHTML)
	}

	return result
}

// openListTag returns the opening tag for a list type.
func openListTag(listType string) string {
	switch listType {
	case "ul":
		return `<ul class="md-ul">`
	case "ol":
		return `<ol class="md-ol">`
	case "task":
		return `<ul class="md-task-list">`
	default:
		return `<ul>`
	}
}

// closeListTag returns the closing tag for a list type.
func closeListTag(listType string) string {
	switch listType {
	case "ol":
		return `</ol>`
	default:
		return `</ul>`
	}
}

// processBlockquotes converts blockquote lines to HTML.
func processBlockquotes(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlockquote := false

	for _, line := range lines {
		if match := blockquoteRe.FindStringSubmatch(line); match != nil {
			if !inBlockquote {
				result = append(result, `<blockquote class="md-blockquote">`)
				inBlockquote = true
			}
			// Don't escape here - escapeRemainingText() will handle it
			result = append(result, match[1])
		} else {
			if inBlockquote {
				result = append(result, `</blockquote>`)
				inBlockquote = false
			}
			result = append(result, line)
		}
	}

	if inBlockquote {
		result = append(result, `</blockquote>`)
	}

	return strings.Join(result, "\n")
}

// convertNewlinesToBr converts newlines to <br> tags, but preserves block element structure.
func convertNewlinesToBr(content string) string {
	// Don't add <br> after block elements
	blockEndings := []string{
		"</h1>", "</h2>", "</h3>", "</h4>", "</h5>", "</h6>",
		"</ul>", "</ol>", "</li>", "</table>", "</tr>", "</blockquote>",
		"</div>", "</pre>", "\x00HR\x00",
	}

	lines := strings.Split(content, "\n")
	var result []string
	consecutiveEmpty := 0

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Track consecutive empty lines and collapse them
		if trimmedLine == "" {
			consecutiveEmpty++
			// Skip this empty line if we've already added one (collapse multiple empties)
			if consecutiveEmpty > 1 {
				continue
			}
		} else {
			consecutiveEmpty = 0
		}

		result = append(result, line)

		// Don't add <br> after the last line or after block elements
		if i < len(lines)-1 {
			isBlockEnd := false
			for _, ending := range blockEndings {
				if strings.HasSuffix(trimmedLine, ending) {
					isBlockEnd = true
					break
				}
			}

			// Also don't add <br> before block elements
			nextTrimmed := strings.TrimSpace(lines[i+1])
			isNextBlockStart := strings.HasPrefix(nextTrimmed, "<h") ||
				strings.HasPrefix(nextTrimmed, "<ul") ||
				strings.HasPrefix(nextTrimmed, "<ol") ||
				strings.HasPrefix(nextTrimmed, "<table") ||
				strings.HasPrefix(nextTrimmed, "<blockquote") ||
				strings.HasPrefix(nextTrimmed, "<div") ||
				strings.HasPrefix(nextTrimmed, "\x00HR\x00") ||
				strings.HasPrefix(nextTrimmed, "\x00CODE_BLOCK")

			// Don't add <br> between empty lines either (prevents double spacing)
			if trimmedLine == "" || isBlockEnd || isNextBlockStart {
				continue
			}

			// Add <br> for regular line breaks
			result[len(result)-1] = line + "<br>"
		}
	}

	return strings.Join(result, "")
}
