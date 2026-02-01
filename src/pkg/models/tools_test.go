package models

import (
	"encoding/json"
	"testing"
)

func TestExtractToolCalls_Bash(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Let me run that command."},
				{"type": "tool_use", "id": "toolu_01ABC", "name": "Bash", "input": {"command": "ls -la", "description": "List files"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].ID != "toolu_01ABC" {
		t.Errorf("Tool ID = %q, want %q", tools[0].ID, "toolu_01ABC")
	}
	if tools[0].Name != "Bash" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Bash")
	}
	if tools[0].Input["command"] != "ls -la" {
		t.Errorf("Tool Input[command] = %v, want %q", tools[0].Input["command"], "ls -la")
	}
}

func TestExtractToolCalls_Read(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_02DEF", "name": "Read", "input": {"file_path": "/path/to/file.go", "offset": 0, "limit": 100}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Read" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Read")
	}
	if tools[0].Input["file_path"] != "/path/to/file.go" {
		t.Errorf("Tool Input[file_path] = %v, want %q", tools[0].Input["file_path"], "/path/to/file.go")
	}
	// JSON numbers are float64
	if tools[0].Input["offset"] != float64(0) {
		t.Errorf("Tool Input[offset] = %v, want 0", tools[0].Input["offset"])
	}
}

func TestExtractToolCalls_Write(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_03GHI", "name": "Write", "input": {"file_path": "/path/to/new.go", "content": "package main\n\nfunc main() {}"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Write" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Write")
	}
	if tools[0].Input["content"] != "package main\n\nfunc main() {}" {
		t.Errorf("Tool Input[content] = %v", tools[0].Input["content"])
	}
}

func TestExtractToolCalls_Edit(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_04JKL", "name": "Edit", "input": {"file_path": "/path/to/file.go", "old_string": "foo", "new_string": "bar"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Edit" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Edit")
	}
	if tools[0].Input["old_string"] != "foo" {
		t.Errorf("Tool Input[old_string] = %v, want %q", tools[0].Input["old_string"], "foo")
	}
	if tools[0].Input["new_string"] != "bar" {
		t.Errorf("Tool Input[new_string] = %v, want %q", tools[0].Input["new_string"], "bar")
	}
}

func TestExtractToolCalls_Task(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_05MNO", "name": "Task", "input": {"description": "Explore the codebase", "prompt": "Find all Go files"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Task" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Task")
	}
	if tools[0].Input["description"] != "Explore the codebase" {
		t.Errorf("Tool Input[description] = %v", tools[0].Input["description"])
	}
}

func TestExtractToolCalls_Glob(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_06PQR", "name": "Glob", "input": {"pattern": "**/*.go", "path": "/project"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Glob" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Glob")
	}
	if tools[0].Input["pattern"] != "**/*.go" {
		t.Errorf("Tool Input[pattern] = %v, want %q", tools[0].Input["pattern"], "**/*.go")
	}
}

func TestExtractToolCalls_Grep(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_07STU", "name": "Grep", "input": {"pattern": "func.*Test", "path": "/project/src", "glob": "*.go"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Grep" {
		t.Errorf("Tool Name = %q, want %q", tools[0].Name, "Grep")
	}
	if tools[0].Input["pattern"] != "func.*Test" {
		t.Errorf("Tool Input[pattern] = %v, want %q", tools[0].Input["pattern"], "func.*Test")
	}
}

func TestExtractToolCalls_MultipleTools(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Let me check the files."},
				{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/file1.go"}},
				{"type": "tool_use", "id": "toolu_02", "name": "Read", "input": {"file_path": "/file2.go"}},
				{"type": "tool_use", "id": "toolu_03", "name": "Bash", "input": {"command": "go test"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 3 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 3", len(tools))
	}

	if tools[0].Name != "Read" || tools[0].ID != "toolu_01" {
		t.Errorf("Tool 0: Name=%q ID=%q, want Read/toolu_01", tools[0].Name, tools[0].ID)
	}
	if tools[1].Name != "Read" || tools[1].ID != "toolu_02" {
		t.Errorf("Tool 1: Name=%q ID=%q, want Read/toolu_02", tools[1].Name, tools[1].ID)
	}
	if tools[2].Name != "Bash" || tools[2].ID != "toolu_03" {
		t.Errorf("Tool 2: Name=%q ID=%q, want Bash/toolu_03", tools[2].Name, tools[2].ID)
	}
}

func TestExtractToolCalls_NonAssistantEntry(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{"type": "tool_result", "tool_use_id": "toolu_01", "content": "result"}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 0 {
		t.Errorf("ExtractToolCalls() on user entry returned %d tools, want 0", len(tools))
	}
}

func TestExtractToolCalls_EmptyContent(t *testing.T) {
	entry := ConversationEntry{
		Type:    EntryTypeAssistant,
		Message: json.RawMessage(`{}`),
	}

	tools := entry.ExtractToolCalls()

	if tools != nil && len(tools) != 0 {
		t.Errorf("ExtractToolCalls() on empty content returned %d tools, want nil or empty", len(tools))
	}
}

func TestExtractToolCalls_NoToolUse(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Just a text response with no tools."}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 0 {
		t.Errorf("ExtractToolCalls() returned %d tools, want 0", len(tools))
	}
}

func TestExtractToolResults_SingleResult(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{"type": "tool_result", "tool_use_id": "toolu_01ABC", "content": "Command executed successfully"}
			]
		}`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 1 {
		t.Fatalf("ExtractToolResults() returned %d results, want 1", len(results))
	}

	if results[0].ToolUseID != "toolu_01ABC" {
		t.Errorf("ToolUseID = %q, want %q", results[0].ToolUseID, "toolu_01ABC")
	}
	if results[0].Content != "Command executed successfully" {
		t.Errorf("Content = %q, want %q", results[0].Content, "Command executed successfully")
	}
	if results[0].IsError {
		t.Errorf("IsError = true, want false")
	}
}

func TestExtractToolResults_ErrorResult(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{"type": "tool_result", "tool_use_id": "toolu_02DEF", "content": "Error: file not found", "is_error": true}
			]
		}`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 1 {
		t.Fatalf("ExtractToolResults() returned %d results, want 1", len(results))
	}

	if results[0].ToolUseID != "toolu_02DEF" {
		t.Errorf("ToolUseID = %q, want %q", results[0].ToolUseID, "toolu_02DEF")
	}
	if results[0].Content != "Error: file not found" {
		t.Errorf("Content = %q, want %q", results[0].Content, "Error: file not found")
	}
	if !results[0].IsError {
		t.Errorf("IsError = false, want true")
	}
}

func TestExtractToolResults_MultipleResults(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{"type": "tool_result", "tool_use_id": "toolu_01", "content": "Result 1"},
				{"type": "tool_result", "tool_use_id": "toolu_02", "content": "Result 2"},
				{"type": "tool_result", "tool_use_id": "toolu_03", "content": "Error", "is_error": true}
			]
		}`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 3 {
		t.Fatalf("ExtractToolResults() returned %d results, want 3", len(results))
	}

	if results[0].ToolUseID != "toolu_01" || results[0].Content != "Result 1" || results[0].IsError {
		t.Errorf("Result 0 mismatch: %+v", results[0])
	}
	if results[1].ToolUseID != "toolu_02" || results[1].Content != "Result 2" || results[1].IsError {
		t.Errorf("Result 1 mismatch: %+v", results[1])
	}
	if results[2].ToolUseID != "toolu_03" || results[2].Content != "Error" || !results[2].IsError {
		t.Errorf("Result 2 mismatch: %+v", results[2])
	}
}

func TestExtractToolResults_NonUserEntry(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "ls"}}
			]
		}`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 0 {
		t.Errorf("ExtractToolResults() on assistant entry returned %d results, want 0", len(results))
	}
}

func TestExtractToolResults_ArrayContent(t *testing.T) {
	// Some tool results have content as an array of content blocks
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`{
			"role": "user",
			"content": [
				{"type": "tool_result", "tool_use_id": "toolu_01", "content": [{"type": "text", "text": "Line 1"}, {"type": "text", "text": "Line 2"}]}
			]
		}`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 1 {
		t.Fatalf("ExtractToolResults() returned %d results, want 1", len(results))
	}

	if results[0].Content != "Line 1\nLine 2" {
		t.Errorf("Content = %q, want %q", results[0].Content, "Line 1\nLine 2")
	}
}

func TestHasToolCall_CaseInsensitive(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "ls"}}
			]
		}`),
	}

	tests := []struct {
		toolName string
		want     bool
	}{
		{"Bash", true},
		{"bash", true},
		{"BASH", true},
		{"BaSh", true},
		{"Read", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := entry.HasToolCall(tt.toolName)
			if got != tt.want {
				t.Errorf("HasToolCall(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestHasToolCall_MultipleTools(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/file.go"}},
				{"type": "tool_use", "id": "toolu_02", "name": "Grep", "input": {"pattern": "test"}}
			]
		}`),
	}

	if !entry.HasToolCall("Read") {
		t.Error("HasToolCall(Read) = false, want true")
	}
	if !entry.HasToolCall("grep") {
		t.Error("HasToolCall(grep) = false, want true")
	}
	if entry.HasToolCall("Bash") {
		t.Error("HasToolCall(Bash) = true, want false")
	}
}

func TestHasToolCall_NonAssistantEntry(t *testing.T) {
	entry := ConversationEntry{
		Type:    EntryTypeUser,
		Message: json.RawMessage(`"test message"`),
	}

	if entry.HasToolCall("Bash") {
		t.Error("HasToolCall() on user entry = true, want false")
	}
}

func TestMatchesToolInput_SimplePattern(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "git status", "description": "Check git status"}}
			]
		}`),
	}

	tests := []struct {
		pattern string
		want    bool
	}{
		{"git status", true},
		{"git", true},
		{"status", true},
		{"git.*status", true},
		{"npm install", false},
		{"GIT STATUS", false}, // Regex is case-sensitive by default
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := entry.MatchesToolInput(tt.pattern)
			if got != tt.want {
				t.Errorf("MatchesToolInput(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchesToolInput_CaseInsensitivePattern(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "git status"}}
			]
		}`),
	}

	// Using (?i) for case-insensitive matching
	if !entry.MatchesToolInput("(?i)GIT STATUS") {
		t.Error("MatchesToolInput((?i)GIT STATUS) = false, want true")
	}
}

func TestMatchesToolInput_FilePath(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/Users/test/project/main.go"}}
			]
		}`),
	}

	tests := []struct {
		pattern string
		want    bool
	}{
		{"main.go", true},
		{"/project/", true},
		{"\\.go\"", true}, // Match .go followed by quote (end of JSON string value)
		{"package.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := entry.MatchesToolInput(tt.pattern)
			if got != tt.want {
				t.Errorf("MatchesToolInput(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestMatchesToolInput_MultipleTools(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Read", "input": {"file_path": "/file1.go"}},
				{"type": "tool_use", "id": "toolu_02", "name": "Bash", "input": {"command": "go test ./..."}}
			]
		}`),
	}

	// Should match if ANY tool input matches
	if !entry.MatchesToolInput("file1.go") {
		t.Error("MatchesToolInput(file1.go) = false, want true")
	}
	if !entry.MatchesToolInput("go test") {
		t.Error("MatchesToolInput(go test) = false, want true")
	}
}

func TestMatchesToolInput_InvalidRegex(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "ls"}}
			]
		}`),
	}

	// Invalid regex should return false
	if entry.MatchesToolInput("[invalid") {
		t.Error("MatchesToolInput([invalid) = true, want false for invalid regex")
	}
}

func TestMatchesToolInput_EmptyInput(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {}}
			]
		}`),
	}

	// Empty input should not match anything (except empty pattern which matches empty JSON)
	if entry.MatchesToolInput("something") {
		t.Error("MatchesToolInput(something) on empty input = true, want false")
	}
}

func TestMatchesToolInput_NoInput(t *testing.T) {
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "id": "toolu_01", "name": "SomeTool"}
			]
		}`),
	}

	if entry.MatchesToolInput("anything") {
		t.Error("MatchesToolInput() with no input = true, want false")
	}
}

func TestMatchesToolInput_NonAssistantEntry(t *testing.T) {
	entry := ConversationEntry{
		Type:    EntryTypeUser,
		Message: json.RawMessage(`"test message"`),
	}

	if entry.MatchesToolInput("test") {
		t.Error("MatchesToolInput() on user entry = true, want false")
	}
}

func TestExtractToolCalls_MalformedJSON(t *testing.T) {
	entry := ConversationEntry{
		Type:    EntryTypeAssistant,
		Message: json.RawMessage(`{not valid json`),
	}

	tools := entry.ExtractToolCalls()

	if tools != nil && len(tools) != 0 {
		t.Errorf("ExtractToolCalls() on malformed JSON returned %d tools, want nil or empty", len(tools))
	}
}

func TestExtractToolResults_MalformedJSON(t *testing.T) {
	entry := ConversationEntry{
		Type:    EntryTypeUser,
		Message: json.RawMessage(`{not valid json`),
	}

	results := entry.ExtractToolResults()

	if results != nil && len(results) != 0 {
		t.Errorf("ExtractToolResults() on malformed JSON returned %d results, want nil or empty", len(results))
	}
}

func TestExtractToolCalls_MissingFields(t *testing.T) {
	// Tool use with missing id
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`{
			"role": "assistant",
			"content": [
				{"type": "tool_use", "name": "Bash", "input": {"command": "ls"}}
			]
		}`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].ID != "" {
		t.Errorf("Tool ID = %q, want empty string", tools[0].ID)
	}
	if tools[0].Name != "Bash" {
		t.Errorf("Tool Name = %q, want Bash", tools[0].Name)
	}
}

func TestExtractToolCalls_DirectContentArray(t *testing.T) {
	// Content without message wrapper envelope
	entry := ConversationEntry{
		Type: EntryTypeAssistant,
		Message: json.RawMessage(`[
			{"type": "text", "text": "Let me check."},
			{"type": "tool_use", "id": "toolu_01", "name": "Bash", "input": {"command": "ls"}}
		]`),
	}

	tools := entry.ExtractToolCalls()

	if len(tools) != 1 {
		t.Fatalf("ExtractToolCalls() returned %d tools, want 1", len(tools))
	}

	if tools[0].Name != "Bash" {
		t.Errorf("Tool Name = %q, want Bash", tools[0].Name)
	}
}

func TestExtractToolResults_DirectContentArray(t *testing.T) {
	// Content without message wrapper envelope
	entry := ConversationEntry{
		Type: EntryTypeUser,
		Message: json.RawMessage(`[
			{"type": "tool_result", "tool_use_id": "toolu_01", "content": "Success"}
		]`),
	}

	results := entry.ExtractToolResults()

	if len(results) != 1 {
		t.Fatalf("ExtractToolResults() returned %d results, want 1", len(results))
	}

	if results[0].Content != "Success" {
		t.Errorf("Content = %q, want Success", results[0].Content)
	}
}
