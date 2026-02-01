// Package models defines data structures for Claude Code history entries.
package models

import (
	"encoding/json"
	"time"
)

// EntryType represents the type of a conversation entry.
type EntryType string

const (
	EntryTypeUser           EntryType = "user"
	EntryTypeAssistant      EntryType = "assistant"
	EntryTypeSystem         EntryType = "system"
	EntryTypeQueueOperation EntryType = "queue-operation"
	EntryTypeSummary        EntryType = "summary"
)

// ConversationEntry represents a single entry in a Claude Code session.
type ConversationEntry struct {
	UUID        string          `json:"uuid"`
	SessionID   string          `json:"sessionId"`
	AgentID     string          `json:"agentId,omitempty"`
	IsSidechain bool            `json:"isSidechain"`
	Type        EntryType       `json:"type"`
	ParentUUID  *string         `json:"parentUuid,omitempty"`
	Timestamp   string          `json:"timestamp"`
	Message     json.RawMessage `json:"message,omitempty"`

	// Additional fields that may be present
	CacheBreakpoint bool   `json:"cacheBreakpoint,omitempty"`
	Usertype        string `json:"userType,omitempty"`
}

// GetTimestamp parses and returns the timestamp as a time.Time.
func (e *ConversationEntry) GetTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, e.Timestamp)
}

// IsUser returns true if this is a user message.
func (e *ConversationEntry) IsUser() bool {
	return e.Type == EntryTypeUser
}

// IsAssistant returns true if this is an assistant message.
func (e *ConversationEntry) IsAssistant() bool {
	return e.Type == EntryTypeAssistant
}

// IsSystem returns true if this is a system message.
func (e *ConversationEntry) IsSystem() bool {
	return e.Type == EntryTypeSystem
}

// IsQueueOperation returns true if this is a queue operation (agent spawn).
func (e *ConversationEntry) IsQueueOperation() bool {
	return e.Type == EntryTypeQueueOperation
}

// MessageContent represents the content of a message.
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// Tool use fields
	ToolUseID string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	// Tool result fields
	ToolResultID string          `json:"tool_use_id,omitempty"`
	Content      json.RawMessage `json:"content,omitempty"`
}

// MessageWrapper represents the Claude Code message envelope with role/content.
type MessageWrapper struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// ParseMessageContent parses the message field into structured content.
func (e *ConversationEntry) ParseMessageContent() ([]MessageContent, error) {
	if len(e.Message) == 0 {
		return nil, nil
	}

	// First, try to unwrap the {role, content} envelope
	var wrapper MessageWrapper
	if err := json.Unmarshal(e.Message, &wrapper); err == nil && len(wrapper.Content) > 0 {
		return parseContent(wrapper.Content)
	}

	// Fall back to parsing the message directly
	return parseContent(e.Message)
}

// parseContent parses content that can be a string, object, or array.
func parseContent(data json.RawMessage) ([]MessageContent, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Try as plain string first (most common for user messages)
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return []MessageContent{{Type: "text", Text: text}}, nil
	}

	// Try array (common for assistant messages and tool results)
	var contents []MessageContent
	if err := json.Unmarshal(data, &contents); err == nil {
		return contents, nil
	}

	// Try single object
	var single MessageContent
	if err := json.Unmarshal(data, &single); err == nil {
		return []MessageContent{single}, nil
	}

	return nil, nil
}

// GetTextContent extracts plain text content from the message.
func (e *ConversationEntry) GetTextContent() string {
	contents, err := e.ParseMessageContent()
	if err != nil {
		return ""
	}

	var text string
	for _, c := range contents {
		if c.Type == "text" && c.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += c.Text
		}
		// Also handle direct text content (no type field)
		if c.Type == "" && c.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += c.Text
		}
	}
	return text
}
