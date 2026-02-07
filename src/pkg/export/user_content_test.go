package export

import (
	"strings"
	"testing"
)

func TestFormatUserContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantContains []string
		wantNotContains []string
	}{
		{
			name:  "empty string",
			input: "",
			wantContains: []string{},
		},
		{
			name:  "plain text without XML",
			input: "Hello world",
			wantContains: []string{"Hello world"},
			wantNotContains: []string{"xml-tag-block"},
		},
		{
			name:  "bash-stdout with content",
			input: "<bash-stdout>beads\nbeads-state-machine\nclaude-code-viewer</bash-stdout>",
			wantContains: []string{
				`xml-tag-block`,
				`&lt;bash-stdout&gt;`,
				`&lt;/bash-stdout&gt;`,
				`xml-tag-content`,
				"beads",
				"beads-state-machine",
				"claude-code-viewer",
			},
			wantNotContains: []string{},
		},
		{
			name:  "bash-stderr empty (should be hidden)",
			input: "<bash-stderr></bash-stderr>",
			wantContains: []string{},
			wantNotContains: []string{"bash-stderr"},
		},
		{
			name:  "bash-stderr with only whitespace (should be hidden)",
			input: "<bash-stderr>   \n  </bash-stderr>",
			wantContains: []string{},
			wantNotContains: []string{"bash-stderr"},
		},
		{
			name:  "mixed stdout and empty stderr",
			input: "<bash-stdout>output text</bash-stdout><bash-stderr></bash-stderr>",
			wantContains: []string{
				"bash-stdout",
				"output text",
			},
			wantNotContains: []string{"bash-stderr"},
		},
		{
			name:  "multiple non-empty tags",
			input: "<bash-stdout>stdout content</bash-stdout><bash-stderr>error message</bash-stderr>",
			wantContains: []string{
				"bash-stdout",
				"stdout content",
				"bash-stderr",
				"error message",
			},
		},
		{
			name:  "XML with text before and after",
			input: "Running command:\n<bash-stdout>result</bash-stdout>\nDone",
			wantContains: []string{
				"Running command:",
				"bash-stdout",
				"result",
				"Done",
			},
		},
		{
			name:  "HTML entities in content are escaped",
			input: "<bash-stdout><script>alert('xss')</script></bash-stdout>",
			wantContains: []string{
				"bash-stdout",
				"&lt;script&gt;",
				"alert(&#39;xss&#39;)",
				"&lt;/script&gt;",
			},
			wantNotContains: []string{"<script>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatUserContent(tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("formatUserContent() result does not contain %q\nGot: %s", want, result)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("formatUserContent() result should not contain %q\nGot: %s", notWant, result)
				}
			}
		})
	}
}

func TestFormatUserContentPreservesLineBreaks(t *testing.T) {
	input := "<bash-stdout>line1\nline2\nline3</bash-stdout>"
	result := formatUserContent(input)

	// Should preserve newlines in content
	if !strings.Contains(result, "line1\nline2\nline3") {
		t.Errorf("formatUserContent() should preserve newlines in content")
	}
}

func TestFormatUserContentMultipleTags(t *testing.T) {
	input := `<bash-stdout>output</bash-stdout><bash-stderr></bash-stderr><other-tag>data</other-tag>`
	result := formatUserContent(input)

	// Should show stdout
	if !strings.Contains(result, "bash-stdout") || !strings.Contains(result, "output") {
		t.Errorf("formatUserContent() should show bash-stdout with content")
	}

	// Should hide empty stderr
	if strings.Contains(result, "bash-stderr") {
		t.Errorf("formatUserContent() should hide empty bash-stderr")
	}

	// Should show other-tag
	if !strings.Contains(result, "other-tag") || !strings.Contains(result, "data") {
		t.Errorf("formatUserContent() should show other-tag with content")
	}
}
