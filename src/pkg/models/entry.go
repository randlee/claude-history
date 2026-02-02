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

// ToolUseResult represents the result of a tool use, particularly for agent spawns.
// When status is "async_launched" and AgentID is non-empty, this indicates an agent spawn.
type ToolUseResult struct {
	IsAsync     bool   `json:"isAsync"`
	Status      string `json:"status"`      // "async_launched", "completed", etc.
	AgentID     string `json:"agentId"`     // ID of the spawned agent
	Description string `json:"description"` // Human-readable description of the task
	Prompt      string `json:"prompt"`      // The prompt given to the spawned agent
	OutputFile  string `json:"outputFile"`  // Path to the agent's output file
}

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

	// SourceToolAssistantUUID links this entry to the assistant message that triggered it
	SourceToolAssistantUUID string `json:"sourceToolAssistantUUID,omitempty"`

	// ToolUseResult contains agent spawn information for user entries with tool results
	ToolUseResult *ToolUseResult `json:"toolUseResult,omitempty"`

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
// Deprecated: Agent spawns are now detected via IsAgentSpawn() which checks toolUseResult.
func (e *ConversationEntry) IsQueueOperation() bool {
	return e.Type == EntryTypeQueueOperation
}

// HasToolUseResult returns true if this entry has a toolUseResult field.
func (e *ConversationEntry) HasToolUseResult() bool {
	return e.ToolUseResult != nil
}

// GetToolUseResult returns the ToolUseResult if present, or nil otherwise.
func (e *ConversationEntry) GetToolUseResult() *ToolUseResult {
	return e.ToolUseResult
}

// IsAgentSpawn returns true if this entry represents an agent spawn.
// Agent spawns are recorded in user entries where toolUseResult.status is "async_launched"
// and toolUseResult.agentId is non-empty.
func (e *ConversationEntry) IsAgentSpawn() bool {
	if e.ToolUseResult == nil {
		return false
	}
	return e.ToolUseResult.Status == "async_launched" && e.ToolUseResult.AgentID != ""
}

// GetSpawnedAgentID returns the ID of the spawned agent if this is an agent spawn entry.
// Returns an empty string if this entry is not an agent spawn.
func (e *ConversationEntry) GetSpawnedAgentID() string {
	if !e.IsAgentSpawn() {
		return ""
	}
	return e.ToolUseResult.AgentID
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
