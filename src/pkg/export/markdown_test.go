package export

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_EmptyString(t *testing.T) {
	result := RenderMarkdown("", "")
	if result != "" {
		t.Errorf("RenderMarkdown('') = %q, want empty string", result)
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	result := RenderMarkdown("Hello, world!", "")
	if !strings.Contains(result, "Hello, world!") {
		t.Errorf("RenderMarkdown should preserve plain text, got %q", result)
	}
}

func TestRenderMarkdown_Headers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "h1",
			input:    "# Header 1",
			expected: `<h1 class="md-h1">Header 1</h1>`,
		},
		{
			name:     "h2",
			input:    "## Header 2",
			expected: `<h2 class="md-h2">Header 2</h2>`,
		},
		{
			name:     "h3",
			input:    "### Header 3",
			expected: `<h3 class="md-h3">Header 3</h3>`,
		},
		{
			name:     "h4",
			input:    "#### Header 4",
			expected: `<h4 class="md-h4">Header 4</h4>`,
		},
		{
			name:     "h5",
			input:    "##### Header 5",
			expected: `<h5 class="md-h5">Header 5</h5>`,
		},
		{
			name:     "h6",
			input:    "###### Header 6",
			expected: `<h6 class="md-h6">Header 6</h6>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMarkdown(tt.input, "")
			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderMarkdown(%q) = %q, want to contain %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRenderMarkdown_MultipleHeaders(t *testing.T) {
	input := `# Title
## Section 1
### Subsection
## Section 2`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<h1 class="md-h1">Title</h1>`) {
		t.Error("Missing h1")
	}
	if !strings.Contains(result, `<h2 class="md-h2">Section 1</h2>`) {
		t.Error("Missing first h2")
	}
	if !strings.Contains(result, `<h3 class="md-h3">Subsection</h3>`) {
		t.Error("Missing h3")
	}
	if !strings.Contains(result, `<h2 class="md-h2">Section 2</h2>`) {
		t.Error("Missing second h2")
	}
}

func TestRenderMarkdown_UnorderedList(t *testing.T) {
	input := `- Item 1
- Item 2
- Item 3`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing ul opening tag")
	}
	if !strings.Contains(result, `<li>Item 1</li>`) {
		t.Error("Missing first list item")
	}
	if !strings.Contains(result, `<li>Item 2</li>`) {
		t.Error("Missing second list item")
	}
	if !strings.Contains(result, `<li>Item 3</li>`) {
		t.Error("Missing third list item")
	}
	if !strings.Contains(result, `</ul>`) {
		t.Error("Missing ul closing tag")
	}
}

func TestRenderMarkdown_OrderedList(t *testing.T) {
	input := `1. First
2. Second
3. Third`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ol class="md-ol">`) {
		t.Error("Missing ol opening tag")
	}
	if !strings.Contains(result, `<li>First</li>`) {
		t.Error("Missing first list item")
	}
	if !strings.Contains(result, `<li>Second</li>`) {
		t.Error("Missing second list item")
	}
	if !strings.Contains(result, `<li>Third</li>`) {
		t.Error("Missing third list item")
	}
	if !strings.Contains(result, `</ol>`) {
		t.Error("Missing ol closing tag")
	}
}

func TestRenderMarkdown_NestedList(t *testing.T) {
	input := `- Parent 1
  - Child 1
  - Child 2
- Parent 2`

	result := RenderMarkdown(input, "")

	// Should have nested structure
	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing outer ul")
	}
	if !strings.Contains(result, `<li>Parent 1</li>`) {
		t.Error("Missing Parent 1")
	}
	if !strings.Contains(result, `<li>Child 1</li>`) {
		t.Error("Missing Child 1")
	}
	if !strings.Contains(result, `<li>Child 2</li>`) {
		t.Error("Missing Child 2")
	}
	if !strings.Contains(result, `<li>Parent 2</li>`) {
		t.Error("Missing Parent 2")
	}

	// Count ul tags to verify nesting
	ulCount := strings.Count(result, `<ul class="md-ul">`)
	if ulCount < 2 {
		t.Errorf("Expected at least 2 ul tags for nested list, got %d", ulCount)
	}
}

func TestRenderMarkdown_MixedListTypes(t *testing.T) {
	input := `- Unordered item
1. Ordered item`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing ul tag")
	}
	if !strings.Contains(result, `<ol class="md-ol">`) {
		t.Error("Missing ol tag")
	}
}

func TestRenderMarkdown_TaskList_Unchecked(t *testing.T) {
	input := "- [ ] Todo item"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ul class="md-task-list">`) {
		t.Error("Missing task list container")
	}
	if !strings.Contains(result, `<li class="task-item">`) {
		t.Error("Missing task-item class")
	}
	if !strings.Contains(result, `<input type="checkbox" disabled>`) {
		t.Error("Missing unchecked checkbox")
	}
	if !strings.Contains(result, "Todo item") {
		t.Error("Missing task text")
	}
}

func TestRenderMarkdown_TaskList_Checked(t *testing.T) {
	input := "- [x] Done item"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ul class="md-task-list">`) {
		t.Error("Missing task list container")
	}
	if !strings.Contains(result, `<input type="checkbox" checked disabled>`) {
		t.Error("Missing checked checkbox")
	}
	if !strings.Contains(result, "Done item") {
		t.Error("Missing task text")
	}
}

func TestRenderMarkdown_TaskList_Mixed(t *testing.T) {
	input := `- [ ] Todo 1
- [x] Done
- [ ] Todo 2`

	result := RenderMarkdown(input, "")

	// Count checkboxes
	uncheckedCount := strings.Count(result, `<input type="checkbox" disabled>`)
	checkedCount := strings.Count(result, `<input type="checkbox" checked disabled>`)

	if uncheckedCount != 2 {
		t.Errorf("Expected 2 unchecked checkboxes, got %d", uncheckedCount)
	}
	if checkedCount != 1 {
		t.Errorf("Expected 1 checked checkbox, got %d", checkedCount)
	}
}

func TestRenderMarkdown_CodeBlock_WithLanguage(t *testing.T) {
	input := "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<div class="code-block language-go">`) {
		t.Error("Missing code-block container with language class")
	}
	if !strings.Contains(result, `<span class="language-badge">go</span>`) {
		t.Error("Missing language badge")
	}
	if !strings.Contains(result, `<button class="copy-code-btn"`) {
		t.Error("Missing copy button")
	}
	if !strings.Contains(result, `<pre class="code-content">`) {
		t.Error("Missing pre element")
	}
	if !strings.Contains(result, `func main()`) {
		t.Error("Missing code content")
	}
}

func TestRenderMarkdown_CodeBlock_NoLanguage(t *testing.T) {
	input := "```\nplain code\n```"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<div class="code-block">`) {
		t.Error("Missing code-block container")
	}
	if !strings.Contains(result, `<span class="language-badge">text</span>`) {
		t.Error("Missing default language badge")
	}
	if !strings.Contains(result, "plain code") {
		t.Error("Missing code content")
	}
}

func TestRenderMarkdown_CodeBlock_MultipleLanguages(t *testing.T) {
	input := "```bash\necho hello\n```\n\nSome text\n\n```python\nprint('world')\n```"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `language-bash`) {
		t.Error("Missing bash language class")
	}
	if !strings.Contains(result, `language-python`) {
		t.Error("Missing python language class")
	}
	if !strings.Contains(result, `<span class="language-badge">bash</span>`) {
		t.Error("Missing bash badge")
	}
	if !strings.Contains(result, `<span class="language-badge">python</span>`) {
		t.Error("Missing python badge")
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	input := "Use the `fmt.Println` function"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<code class="inline-code">fmt.Println</code>`) {
		t.Errorf("Missing inline code, got %q", result)
	}
}

func TestRenderMarkdown_InlineCode_Multiple(t *testing.T) {
	input := "Use `foo` and `bar` functions"

	result := RenderMarkdown(input, "")

	if strings.Count(result, `<code class="inline-code">`) != 2 {
		t.Error("Expected 2 inline code elements")
	}
	if !strings.Contains(result, `<code class="inline-code">foo</code>`) {
		t.Error("Missing first inline code")
	}
	if !strings.Contains(result, `<code class="inline-code">bar</code>`) {
		t.Error("Missing second inline code")
	}
}

func TestRenderMarkdown_Links(t *testing.T) {
	input := "Visit [Google](https://google.com) for more info"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<a href="https://google.com" class="md-link">Google</a>`) {
		t.Errorf("Missing link, got %q", result)
	}
}

func TestRenderMarkdown_Links_Multiple(t *testing.T) {
	input := "[Link 1](https://example.com) and [Link 2](https://test.com)"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<a href="https://example.com" class="md-link">Link 1</a>`) {
		t.Error("Missing first link")
	}
	if !strings.Contains(result, `<a href="https://test.com" class="md-link">Link 2</a>`) {
		t.Error("Missing second link")
	}
}

func TestRenderMarkdown_Images(t *testing.T) {
	input := "![Alt text](https://example.com/image.png)"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<img src="https://example.com/image.png" alt="Alt text" class="md-image">`) {
		t.Errorf("Missing image, got %q", result)
	}
}

func TestRenderMarkdown_Blockquote(t *testing.T) {
	input := "> This is a quote"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<blockquote class="md-blockquote">`) {
		t.Error("Missing blockquote opening tag")
	}
	if !strings.Contains(result, "This is a quote") {
		t.Error("Missing quote content")
	}
	if !strings.Contains(result, `</blockquote>`) {
		t.Error("Missing blockquote closing tag")
	}
}

func TestRenderMarkdown_Blockquote_Multiline(t *testing.T) {
	input := `> Line 1
> Line 2
> Line 3`

	result := RenderMarkdown(input, "")

	// Should be a single blockquote
	if strings.Count(result, `<blockquote class="md-blockquote">`) != 1 {
		t.Error("Expected single blockquote for consecutive lines")
	}
	if !strings.Contains(result, "Line 1") {
		t.Error("Missing Line 1")
	}
	if !strings.Contains(result, "Line 2") {
		t.Error("Missing Line 2")
	}
	if !strings.Contains(result, "Line 3") {
		t.Error("Missing Line 3")
	}
}

func TestRenderMarkdown_Table_Basic(t *testing.T) {
	input := `| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<table class="md-table">`) {
		t.Error("Missing table tag")
	}
	if !strings.Contains(result, `<thead>`) {
		t.Error("Missing thead tag")
	}
	if !strings.Contains(result, `<th>Header 1</th>`) {
		t.Error("Missing header 1")
	}
	if !strings.Contains(result, `<th>Header 2</th>`) {
		t.Error("Missing header 2")
	}
	if !strings.Contains(result, `<tbody>`) {
		t.Error("Missing tbody tag")
	}
	if !strings.Contains(result, `<td>Cell 1</td>`) {
		t.Error("Missing cell 1")
	}
	if !strings.Contains(result, `<td>Cell 2</td>`) {
		t.Error("Missing cell 2")
	}
}

func TestRenderMarkdown_Table_MultipleRows(t *testing.T) {
	input := `| Name | Age |
|------|-----|
| Alice | 30 |
| Bob | 25 |`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<th>Name</th>`) {
		t.Error("Missing Name header")
	}
	if !strings.Contains(result, `<th>Age</th>`) {
		t.Error("Missing Age header")
	}
	if !strings.Contains(result, `<td>Alice</td>`) {
		t.Error("Missing Alice")
	}
	if !strings.Contains(result, `<td>30</td>`) {
		t.Error("Missing 30")
	}
	if !strings.Contains(result, `<td>Bob</td>`) {
		t.Error("Missing Bob")
	}
	if !strings.Contains(result, `<td>25</td>`) {
		t.Error("Missing 25")
	}

	// Count rows
	trCount := strings.Count(result, `<tr>`)
	if trCount != 3 { // 1 header + 2 body rows
		t.Errorf("Expected 3 tr elements, got %d", trCount)
	}
}

func TestRenderMarkdown_HorizontalRule(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"dashes", "---"},
		{"asterisks", "***"},
		{"underscores", "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderMarkdown(tt.input, "")
			if !strings.Contains(result, `<hr class="md-hr">`) {
				t.Errorf("Missing hr for %s, got %q", tt.name, result)
			}
		})
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	input := "This is **bold** text"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<strong>bold</strong>`) {
		t.Errorf("Missing bold, got %q", result)
	}
}

func TestRenderMarkdown_Italic(t *testing.T) {
	input := "This is *italic* text"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<em>italic</em>`) {
		t.Errorf("Missing italic, got %q", result)
	}
}

func TestRenderMarkdown_BoldAndItalic(t *testing.T) {
	input := "This is **bold** and *italic* text"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<strong>bold</strong>`) {
		t.Error("Missing bold")
	}
	if !strings.Contains(result, `<em>italic</em>`) {
		t.Error("Missing italic")
	}
}

func TestRenderMarkdown_MixedContent(t *testing.T) {
	input := `# My Document

This is a paragraph with **bold** and *italic* text.

## Code Example

Here is some code:

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```" + `

## List

- Item 1
- Item 2

## Table

| A | B |
|---|---|
| 1 | 2 |

> A quote

---

The end.`

	result := RenderMarkdown(input, "")

	// Verify all major elements are present
	if !strings.Contains(result, `<h1 class="md-h1">My Document</h1>`) {
		t.Error("Missing h1")
	}
	if !strings.Contains(result, `<strong>bold</strong>`) {
		t.Error("Missing bold")
	}
	if !strings.Contains(result, `<em>italic</em>`) {
		t.Error("Missing italic")
	}
	if !strings.Contains(result, `<div class="code-block language-go">`) {
		t.Error("Missing code block")
	}
	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing list")
	}
	if !strings.Contains(result, `<table class="md-table">`) {
		t.Error("Missing table")
	}
	if !strings.Contains(result, `<blockquote class="md-blockquote">`) {
		t.Error("Missing blockquote")
	}
	if !strings.Contains(result, `<hr class="md-hr">`) {
		t.Error("Missing horizontal rule")
	}
}

func TestRenderMarkdown_HTMLEscaping(t *testing.T) {
	input := "This has <script>alert('xss')</script> in it"

	result := RenderMarkdown(input, "")

	if strings.Contains(result, "<script>") {
		t.Error("XSS vulnerability: script tag not escaped")
	}
	if !strings.Contains(result, "&lt;script&gt;") {
		t.Error("Script tag should be HTML escaped")
	}
}

func TestRenderMarkdown_HTMLEscaping_InCodeBlock(t *testing.T) {
	input := "```html\n<div class=\"container\"></div>\n```"

	result := RenderMarkdown(input, "")

	if strings.Contains(result, `<div class="container">`) && !strings.Contains(result, "&lt;div") {
		t.Error("HTML in code block should be escaped")
	}
}

func TestRenderMarkdown_SpecialCharacters(t *testing.T) {
	input := "Ampersand: & Less than: < Greater than: >"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, "&amp;") {
		t.Error("Ampersand should be escaped")
	}
	if !strings.Contains(result, "&lt;") {
		t.Error("Less than should be escaped")
	}
	if !strings.Contains(result, "&gt;") {
		t.Error("Greater than should be escaped")
	}
}

func TestExtractCodeBlocks_NoBlocks(t *testing.T) {
	content := "No code blocks here"

	blocks := ExtractCodeBlocks(content)

	if len(blocks) != 0 {
		t.Errorf("Expected 0 code blocks, got %d", len(blocks))
	}
}

func TestExtractCodeBlocks_SingleBlock(t *testing.T) {
	content := "```go\nfunc main() {}\n```"

	blocks := ExtractCodeBlocks(content)

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 code block, got %d", len(blocks))
	}

	if blocks[0].Language != "go" {
		t.Errorf("Expected language 'go', got %q", blocks[0].Language)
	}
	if !strings.Contains(blocks[0].Code, "func main()") {
		t.Errorf("Code should contain 'func main()', got %q", blocks[0].Code)
	}
}

func TestExtractCodeBlocks_MultipleBlocks(t *testing.T) {
	content := "```bash\necho hello\n```\n\n```python\nprint('hi')\n```"

	blocks := ExtractCodeBlocks(content)

	if len(blocks) != 2 {
		t.Fatalf("Expected 2 code blocks, got %d", len(blocks))
	}

	if blocks[0].Language != "bash" {
		t.Errorf("First block language should be 'bash', got %q", blocks[0].Language)
	}
	if blocks[1].Language != "python" {
		t.Errorf("Second block language should be 'python', got %q", blocks[1].Language)
	}
}

func TestExtractCodeBlocks_NoLanguage(t *testing.T) {
	content := "```\nplain text\n```"

	blocks := ExtractCodeBlocks(content)

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 code block, got %d", len(blocks))
	}

	if blocks[0].Language != "" {
		t.Errorf("Expected empty language, got %q", blocks[0].Language)
	}
}

func TestExtractCodeBlocks_Positions(t *testing.T) {
	content := "before\n```go\ncode\n```\nafter"

	blocks := ExtractCodeBlocks(content)

	if len(blocks) != 1 {
		t.Fatalf("Expected 1 code block, got %d", len(blocks))
	}

	if blocks[0].StartPos == 0 {
		t.Error("StartPos should not be 0")
	}
	if blocks[0].EndPos <= blocks[0].StartPos {
		t.Error("EndPos should be greater than StartPos")
	}
}

func TestRenderMarkdown_PreservesCodeBlockContent(t *testing.T) {
	// Code blocks should preserve their content exactly (except for HTML escaping)
	input := "```go\n// Comment with **markdown**\nfunc test() {\n    x := 1 + 2\n}\n```"

	result := RenderMarkdown(input, "")

	// The markdown syntax inside code block should NOT be converted
	if strings.Contains(result, "<strong>markdown</strong>") {
		t.Error("Markdown inside code block should not be rendered")
	}
	// But the text should be there
	if !strings.Contains(result, "**markdown**") {
		t.Error("Code content should be preserved")
	}
}

func TestRenderMarkdown_InlineCodeNotNestedInCodeBlock(t *testing.T) {
	input := "```bash\necho `date`\n```"

	result := RenderMarkdown(input, "")

	// Backticks inside code block should not create inline code
	inlineCodeCount := strings.Count(result, `<code class="inline-code">`)
	if inlineCodeCount > 0 {
		t.Error("Inline code should not be created inside code blocks")
	}
}

func TestRenderMarkdown_EmptyCodeBlock(t *testing.T) {
	input := "```\n```"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<div class="code-block">`) {
		t.Error("Empty code block should still render container")
	}
}

func TestRenderMarkdown_ListAfterParagraph(t *testing.T) {
	input := `Some text before.

- Item 1
- Item 2

Some text after.`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, "Some text before.") {
		t.Error("Missing text before list")
	}
	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing list")
	}
	if !strings.Contains(result, "Some text after.") {
		t.Error("Missing text after list")
	}
}

func TestRenderMarkdown_TableNotAtStart(t *testing.T) {
	input := `Introduction text.

| A | B |
|---|---|
| 1 | 2 |

Conclusion text.`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, "Introduction text.") {
		t.Error("Missing introduction")
	}
	if !strings.Contains(result, `<table class="md-table">`) {
		t.Error("Missing table")
	}
	if !strings.Contains(result, "Conclusion text.") {
		t.Error("Missing conclusion")
	}
}

func TestRenderMarkdown_AsteriskListItems(t *testing.T) {
	input := `* Item 1
* Item 2
* Item 3`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ul class="md-ul">`) {
		t.Error("Missing ul for asterisk list")
	}
	if !strings.Contains(result, `<li>Item 1</li>`) {
		t.Error("Missing Item 1")
	}
}

func TestRenderMarkdown_LargeOrderedListNumbers(t *testing.T) {
	input := `10. Item 10
11. Item 11
100. Item 100`

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `<ol class="md-ol">`) {
		t.Error("Missing ol tag")
	}
	if !strings.Contains(result, `<li>Item 10</li>`) {
		t.Error("Missing Item 10")
	}
	if !strings.Contains(result, `<li>Item 100</li>`) {
		t.Error("Missing Item 100")
	}
}

func TestRenderMarkdown_LinkWithSpecialChars(t *testing.T) {
	input := "[Link with spaces](https://example.com/path?q=test&foo=bar)"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `href="https://example.com/path?q=test&amp;foo=bar"`) {
		t.Errorf("Link URL should have escaped ampersand, got %q", result)
	}
}

func TestRenderMarkdown_ImageWithEmptyAlt(t *testing.T) {
	input := "![](https://example.com/image.png)"

	result := RenderMarkdown(input, "")

	if !strings.Contains(result, `alt=""`) {
		t.Error("Image should have empty alt attribute")
	}
}

func TestRenderMarkdown_NewlinesConvertedToBr(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3"

	result := RenderMarkdown(input, "")

	// Should have <br> tags for line breaks
	brCount := strings.Count(result, "<br>")
	if brCount < 2 {
		t.Errorf("Expected at least 2 <br> tags, got %d", brCount)
	}
}

func TestRenderMarkdown_NoExtraBrAfterBlockElements(t *testing.T) {
	input := `# Header

Paragraph`

	result := RenderMarkdown(input, "")

	// Should not have <br> immediately after </h1>
	if strings.Contains(result, "</h1><br>") || strings.Contains(result, "</h1>\n<br>") {
		t.Error("Should not have <br> after block elements")
	}
}

// Benchmark tests
func BenchmarkRenderMarkdown_Simple(b *testing.B) {
	input := "Hello **world**"
	for i := 0; i < b.N; i++ {
		RenderMarkdown(input, "")
	}
}

func BenchmarkRenderMarkdown_Complex(b *testing.B) {
	input := `# Title

This is a **bold** paragraph with *italic* text.

## Code

` + "```go\nfunc main() {}\n```" + `

## List

- Item 1
- Item 2
  - Nested

| A | B |
|---|---|
| 1 | 2 |

> Quote

---

[Link](https://example.com)
`
	for i := 0; i < b.N; i++ {
		RenderMarkdown(input, "")
	}
}

func BenchmarkExtractCodeBlocks(b *testing.B) {
	input := "```go\nfunc main() {}\n```\n\n```bash\necho hello\n```"
	for i := 0; i < b.N; i++ {
		ExtractCodeBlocks(input)
	}
}
