// Package models defines data structures for Claude Code history entries.
package models

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ToolUse represents a tool call in an assistant message.
type ToolUse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// ExtractToolCalls extracts tool_use blocks from assistant message content.
// Returns an empty slice if the entry is not an assistant message or has no tool calls.
func (e *ConversationEntry) ExtractToolCalls() []ToolUse {
	if e.Type != EntryTypeAssistant {
		return nil
	}

	contents, err := e.ParseMessageContent()
	if err != nil {
		return nil
	}

	var tools []ToolUse
	for _, c := range contents {
		if c.Type != "tool_use" {
			continue
		}

		tool := ToolUse{
			ID:   c.ToolUseID,
			Name: c.Name,
		}

		// Parse the input field if present
		if len(c.Input) > 0 {
			var input map[string]any
			if err := json.Unmarshal(c.Input, &input); err == nil {
				tool.Input = input
			}
		}

		tools = append(tools, tool)
	}

	return tools
}

// ExtractToolResults extracts tool results from user message content.
// User messages with tool results have content as an array of tool_result objects.
// Returns an empty slice if the entry is not a user message or has no tool results.
func (e *ConversationEntry) ExtractToolResults() []ToolResult {
	if e.Type != EntryTypeUser {
		return nil
	}

	contents, err := e.ParseMessageContent()
	if err != nil {
		return nil
	}

	var results []ToolResult
	for _, c := range contents {
		if c.Type != "tool_result" {
			continue
		}

		result := ToolResult{
			ToolUseID: c.ToolResultID,
		}

		// Parse content - can be string or array
		if len(c.Content) > 0 {
			// Try as string first
			var contentStr string
			if err := json.Unmarshal(c.Content, &contentStr); err == nil {
				result.Content = contentStr
			} else {
				// Try as array of content blocks
				var contentBlocks []struct {
					Type string `json:"type"`
					Text string `json:"text,omitempty"`
				}
				if err := json.Unmarshal(c.Content, &contentBlocks); err == nil {
					var texts []string
					for _, block := range contentBlocks {
						if block.Text != "" {
							texts = append(texts, block.Text)
						}
					}
					result.Content = strings.Join(texts, "\n")
				}
			}
		}

		// Check for is_error field in the original content
		// We need to re-parse to get is_error since MessageContent doesn't have it
		result.IsError = extractIsError(e.Message, c.ToolResultID)

		results = append(results, result)
	}

	return results
}

// extractIsError checks if a tool result has is_error set to true.
func extractIsError(message json.RawMessage, toolUseID string) bool {
	if len(message) == 0 {
		return false
	}

	// First unwrap the message envelope if present
	var wrapper struct {
		Content json.RawMessage `json:"content"`
	}
	contentData := message
	if err := json.Unmarshal(message, &wrapper); err == nil && len(wrapper.Content) > 0 {
		contentData = wrapper.Content
	}

	// Parse as array of tool results
	var results []struct {
		Type      string `json:"type"`
		ToolUseID string `json:"tool_use_id"`
		IsError   bool   `json:"is_error"`
	}
	if err := json.Unmarshal(contentData, &results); err != nil {
		return false
	}

	for _, r := range results {
		if r.Type == "tool_result" && r.ToolUseID == toolUseID {
			return r.IsError
		}
	}

	return false
}

// HasToolCall checks if the entry has a tool call with the specified name.
// The comparison is case-insensitive.
func (e *ConversationEntry) HasToolCall(toolName string) bool {
	tools := e.ExtractToolCalls()
	toolNameLower := strings.ToLower(toolName)

	for _, tool := range tools {
		if strings.ToLower(tool.Name) == toolNameLower {
			return true
		}
	}

	return false
}

// MatchesToolInput checks if any tool input matches the given regex pattern.
// The input map is serialized to JSON and matched against the pattern.
// Returns false if the pattern is invalid or no tool inputs match.
func (e *ConversationEntry) MatchesToolInput(pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}

	tools := e.ExtractToolCalls()
	for _, tool := range tools {
		if tool.Input == nil {
			continue
		}

		// Serialize input to JSON for pattern matching
		inputJSON, err := json.Marshal(tool.Input)
		if err != nil {
			continue
		}

		if re.Match(inputJSON) {
			return true
		}
	}

	return false
}
