package markdownchunker

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNewMarkdownChunker(t *testing.T) {
	chunker := NewMarkdownChunker()
	if chunker == nil {
		t.Fatal("NewMarkdownChunker() returned nil")
	}
	if chunker.md == nil {
		t.Error("MarkdownChunker.md is nil")
	}
	if chunker.chunks == nil {
		t.Error("MarkdownChunker.chunks is nil")
	}
	if len(chunker.chunks) != 0 {
		t.Error("MarkdownChunker.chunks should be empty initially")
	}
}

func TestChunkDocument_EmptyContent(t *testing.T) {
	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(""))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty content, got %d", len(chunks))
	}
}

func TestChunkDocument_Headings(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected []Chunk
	}{
		{
			name:     "Single H1",
			markdown: "# Main Title",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "heading",
					Content: "# Main Title",
					Text:    "Main Title",
					Level:   1,
					Metadata: map[string]string{
						"heading_level": "1",
						"level":         "1",
						"word_count":    "2",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 3, EndCol: 13},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "18eda1a7b1be12c1875bb4fdd88c46af74e7340ee70cf69199901b1027d64eb6",
				},
			},
		},
		{
			name:     "Multiple headings",
			markdown: "# H1\n## H2\n### H3",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "heading",
					Content: "# H1",
					Text:    "H1",
					Level:   1,
					Metadata: map[string]string{
						"heading_level": "1",
						"level":         "1",
						"word_count":    "1",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 3, EndCol: 5},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "9177d76a7b3263db70336007126b5bb82a17cc6506b0ed5508ed9d80fd4c641f",
				},
				{
					ID:      1,
					Type:    "heading",
					Content: "## H2",
					Text:    "H2",
					Level:   2,
					Metadata: map[string]string{
						"heading_level": "2",
						"level":         "2",
						"word_count":    "1",
					},
					Position: ChunkPosition{StartLine: 2, EndLine: 2, StartCol: 4, EndCol: 6},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "42940d6aab97a1f6f298b090673106e0da293c4c9b76bf3a4caa39a374439978",
				},
				{
					ID:      2,
					Type:    "heading",
					Content: "### H3",
					Text:    "H3",
					Level:   3,
					Metadata: map[string]string{
						"heading_level": "3",
						"level":         "3",
						"word_count":    "1",
					},
					Position: ChunkPosition{StartLine: 3, EndLine: 3, StartCol: 5, EndCol: 7},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "f5ae32ad45c6026eee1690a56a3bb9a14e4514c59636e179ee516243c42fa853",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if !reflect.DeepEqual(chunks, tt.expected) {
				t.Errorf("ChunkDocument() = %+v, want %+v", chunks, tt.expected)
			}
		})
	}
}

func TestChunkDocument_Paragraphs(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected []Chunk
	}{
		{
			name:     "Single paragraph",
			markdown: "This is a simple paragraph.",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "paragraph",
					Content: "This is a simple paragraph.",
					Text:    "This is a simple paragraph.",
					Level:   0,
					Metadata: map[string]string{
						"word_count": "5",
						"char_count": "27",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 28},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "d1b7f5d3758e33708a120896978625722c73728c39745f9ddc125989a3b5fbf7",
				},
			},
		},
		{
			name:     "Multiple paragraphs",
			markdown: "First paragraph.\n\nSecond paragraph with more words.",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "paragraph",
					Content: "First paragraph.",
					Text:    "First paragraph.",
					Level:   0,
					Metadata: map[string]string{
						"word_count": "2",
						"char_count": "16",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 17},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "98ea01bc109a52fdf7145c10c648e8b27b8ebc877aaa79405f20b044ecfcacaa",
				},
				{
					ID:      1,
					Type:    "paragraph",
					Content: "Second paragraph with more words.",
					Text:    "Second paragraph with more words.",
					Level:   0,
					Metadata: map[string]string{
						"word_count": "5",
						"char_count": "33",
					},
					Position: ChunkPosition{StartLine: 3, EndLine: 3, StartCol: 1, EndCol: 34},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "fcb0540d8bdbaea6c91b5453281abce111259599d3989f08e059512f9f63ef36",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if !reflect.DeepEqual(chunks, tt.expected) {
				t.Errorf("ChunkDocument() = %+v, want %+v", chunks, tt.expected)
			}
		})
	}
}

func TestChunkDocument_CodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected []Chunk
	}{
		{
			name:     "Fenced code block with language",
			markdown: "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "code",
					Content: "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
					Text:    "func main() {\n    fmt.Println(\"Hello\")\n}",
					Level:   0,
					Metadata: map[string]string{
						"language":   "go",
						"line_count": "3",
					},
					Position: ChunkPosition{StartLine: 2, EndLine: 5, StartCol: 1, EndCol: 1},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "1e45e0168c59424b478f84615b76f4f2777e021e3a1cc30acf9284f45d6cb485",
				},
			},
		},
		{
			name:     "Fenced code block without language",
			markdown: "```\nsome code\n```",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "code",
					Content: "```\nsome code\n```",
					Text:    "some code",
					Level:   0,
					Metadata: map[string]string{
						"language":   "",
						"line_count": "1",
					},
					Position: ChunkPosition{StartLine: 2, EndLine: 3, StartCol: 1, EndCol: 1},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "f3e59445f665accdc7b637b3e61d3b44b278e83c079491a1c84f5462f792233f",
				},
			},
		},
		{
			name:     "Indented code block",
			markdown: "    indented code\n    second line",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "code",
					Content: "    indented code\n    second line",
					Text:    "indented code\nsecond line",
					Level:   0,
					Metadata: map[string]string{
						"language":   "",
						"line_count": "2",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 2, StartCol: 5, EndCol: 16},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "fd0881230113c6a11acb15049dee831213d0a22590137880baede071e925e906",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if !reflect.DeepEqual(chunks, tt.expected) {
				t.Errorf("ChunkDocument() = %+v, want %+v", chunks, tt.expected)
			}
		})
	}
}

func TestChunkDocument_Tables(t *testing.T) {
	markdown := `| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.Type != "table" {
		t.Errorf("Expected type 'table', got '%s'", chunk.Type)
	}
	if chunk.Metadata["rows"] != "3" {
		t.Errorf("Expected 3 rows, got %s", chunk.Metadata["rows"])
	}
	if chunk.Metadata["columns"] != "3" {
		t.Errorf("Expected 3 columns, got %s", chunk.Metadata["columns"])
	}
}

func TestChunkDocument_Lists(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected []Chunk
	}{
		{
			name:     "Unordered list",
			markdown: "- Item 1\n- Item 2\n- Item 3",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "list",
					Content: "- Item 1\n- Item 2\n- Item 3",
					Text:    "Item 1 Item 2 Item 3",
					Level:   0,
					Metadata: map[string]string{
						"list_type":  "unordered",
						"item_count": "3",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 1},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "8f510266449d9073201dd709c6c76710da87504e154df95dedc226c15fa59d31",
				},
			},
		},
		{
			name:     "Ordered list",
			markdown: "1. First item\n2. Second item\n3. Third item",
			expected: []Chunk{
				{
					ID:      0,
					Type:    "list",
					Content: "1. First item\n2. Second item\n3. Third item",
					Text:    "First item Second item Third item",
					Level:   0,
					Metadata: map[string]string{
						"list_type":  "ordered",
						"item_count": "3",
					},
					Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 1},
					Links:    []Link{},
					Images:   []Image{},
					Hash:     "f5c423f0239ac59acaf2ad30ccc64854442c09030201f7ab3fae724d2984fb2d",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if !reflect.DeepEqual(chunks, tt.expected) {
				t.Errorf("ChunkDocument() = %+v, want %+v", chunks, tt.expected)
			}
		})
	}
}

func TestChunkDocument_Blockquotes(t *testing.T) {
	markdown := "> This is a blockquote\n> with multiple lines"
	expected := []Chunk{
		{
			ID:      0,
			Type:    "blockquote",
			Content: "> This is a blockquote\n> with multiple lines",
			Text:    "This is a blockquote with multiple lines",
			Level:   0,
			Metadata: map[string]string{
				"word_count": "7",
			},
			Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 1},
			Links:    []Link{},
			Images:   []Image{},
			Hash:     "c056c34a7a84f9716732ea025441f99c941873a14b982fab59e5cd510cb0ef9a",
		},
	}

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("ChunkDocument() = %+v, want %+v", chunks, expected)
	}
}

func TestChunkDocument_ThematicBreaks(t *testing.T) {
	markdown := "---"
	expected := []Chunk{
		{
			ID:      0,
			Type:    "thematic_break",
			Content: "---",
			Text:    "---",
			Level:   0,
			Metadata: map[string]string{
				"type": "horizontal_rule",
			},
			Position: ChunkPosition{StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 1},
			Links:    []Link{},
			Images:   []Image{},
			Hash:     "cb3f91d54eee30e53e35b2b99905f70f169ed549fd78909d3dac2defc9ed8d3b",
		},
	}

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}
	if !reflect.DeepEqual(chunks, expected) {
		t.Errorf("ChunkDocument() = %+v, want %+v", chunks, expected)
	}
}

func TestChunkDocument_ComplexDocument(t *testing.T) {
	markdown := `# Main Title

This is an introduction paragraph.

## Section 1

Here's some content with **bold** and *italic* text.

` + "```python" + `
def hello():
    print("Hello, World!")
` + "```" + `

### Subsection

- Item 1
- Item 2
- Item 3

| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |

> This is a blockquote

---

Final paragraph.`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	expectedTypes := []string{
		"heading",        // # Main Title
		"paragraph",      // This is an introduction paragraph.
		"heading",        // ## Section 1
		"paragraph",      // Here's some content...
		"code",           // Python code block
		"heading",        // ### Subsection
		"list",           // - Item 1, 2, 3
		"table",          // | Column 1 | Column 2 |
		"blockquote",     // > This is a blockquote
		"thematic_break", // ---
		"paragraph",      // Final paragraph.
	}

	if len(chunks) != len(expectedTypes) {
		t.Fatalf("Expected %d chunks, got %d", len(expectedTypes), len(chunks))
	}

	for i, expectedType := range expectedTypes {
		if chunks[i].Type != expectedType {
			t.Errorf("Chunk %d: expected type '%s', got '%s'", i, expectedType, chunks[i].Type)
		}
		if chunks[i].ID != i {
			t.Errorf("Chunk %d: expected ID %d, got %d", i, i, chunks[i].ID)
		}
	}

	// Test specific chunks
	// Heading levels
	if chunks[0].Level != 1 {
		t.Errorf("First heading should be level 1, got %d", chunks[0].Level)
	}
	if chunks[2].Level != 2 {
		t.Errorf("Second heading should be level 2, got %d", chunks[2].Level)
	}
	if chunks[5].Level != 3 {
		t.Errorf("Third heading should be level 3, got %d", chunks[5].Level)
	}

	// Code block language
	if chunks[4].Metadata["language"] != "python" {
		t.Errorf("Code block should have language 'python', got '%s'", chunks[4].Metadata["language"])
	}

	// List type
	if chunks[6].Metadata["list_type"] != "unordered" {
		t.Errorf("List should be unordered, got '%s'", chunks[6].Metadata["list_type"])
	}
	if chunks[6].Metadata["item_count"] != "3" {
		t.Errorf("List should have 3 items, got '%s'", chunks[6].Metadata["item_count"])
	}

	// Table dimensions
	if chunks[7].Metadata["rows"] != "2" {
		t.Errorf("Table should have 2 rows, got '%s'", chunks[7].Metadata["rows"])
	}
	if chunks[7].Metadata["columns"] != "2" {
		t.Errorf("Table should have 2 columns, got '%s'", chunks[7].Metadata["columns"])
	}
}

func TestChunkDocument_EmptyParagraphFiltering(t *testing.T) {
	markdown := "# Title\n\n\n\nContent paragraph"

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Should only have heading and paragraph, empty paragraphs should be filtered
	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks (heading + paragraph), got %d", len(chunks))
	}

	if chunks[0].Type != "heading" {
		t.Errorf("First chunk should be heading, got %s", chunks[0].Type)
	}
	if chunks[1].Type != "paragraph" {
		t.Errorf("Second chunk should be paragraph, got %s", chunks[1].Type)
	}
}

func TestChunkDocument_TextExtraction(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Bold and italic text",
			markdown: "This is **bold** and *italic* text.",
			expected: "This is bold and italic text.",
		},
		{
			name:     "Links",
			markdown: "Visit [Google](https://google.com) for search.",
			expected: "Visit Google for search.",
		},
		{
			name:     "Code spans",
			markdown: "Use `console.log()` to print.",
			expected: "Use console.log() to print.",
		},
		{
			name:     "Mixed formatting",
			markdown: "**Bold** text with `code` and [link](url).",
			expected: "Bold text with code and link.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if len(chunks) != 1 {
				t.Fatalf("Expected 1 chunk, got %d", len(chunks))
			}
			if chunks[0].Text != tt.expected {
				t.Errorf("Expected text '%s', got '%s'", tt.expected, chunks[0].Text)
			}
		})
	}
}

// Benchmark tests
func BenchmarkChunkDocument_Small(b *testing.B) {
	markdown := `# Title
This is a paragraph.
- Item 1
- Item 2`

	chunker := NewMarkdownChunker()

	for b.Loop() {
		_, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunkDocument_Large(b *testing.B) {
	// Create a large markdown document
	var markdown strings.Builder
	for i := range 100 {
		markdown.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
		markdown.WriteString("This is a paragraph with some content.\n\n")
		markdown.WriteString("```go\nfunc example() {\n    return nil\n}\n```\n\n")
		markdown.WriteString("- Item 1\n- Item 2\n- Item 3\n\n")
	}

	chunker := NewMarkdownChunker()
	content := []byte(markdown.String())

	for b.Loop() {
		_, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewMarkdownChunker(b *testing.B) {
	for b.Loop() {
		_ = NewMarkdownChunker()
	}
}

// Additional robust tests for edge cases and error conditions

func TestChunkDocument_InvalidUTF8(t *testing.T) {
	chunker := NewMarkdownChunker()
	// Test with invalid UTF-8 bytes
	invalidUTF8 := []byte{0xff, 0xfe, 0xfd}
	chunks, err := chunker.ChunkDocument(invalidUTF8)
	if err != nil {
		t.Fatalf("ChunkDocument() should handle invalid UTF-8 gracefully, got error: %v", err)
	}
	// goldmark handles invalid UTF-8 by creating content, so we just verify no panic/error
	t.Logf("Invalid UTF-8 handled gracefully, produced %d chunks", len(chunks))
}

func TestChunkDocument_VeryLargeContent(t *testing.T) {
	chunker := NewMarkdownChunker()
	// Create a very large markdown document (1MB+)
	var largeContent strings.Builder
	for i := range 10000 {
		largeContent.WriteString(fmt.Sprintf("# Heading %d\n\nThis is paragraph %d with some content.\n\n", i, i))
	}

	chunks, err := chunker.ChunkDocument([]byte(largeContent.String()))
	if err != nil {
		t.Fatalf("ChunkDocument() should handle large content, got error: %v", err)
	}

	expectedChunks := 20000 // 10000 headings + 10000 paragraphs
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks for large content, got %d", expectedChunks, len(chunks))
	}
}

func TestChunkDocument_NestedStructures(t *testing.T) {
	markdown := "# Main Heading\n\n" +
		"> This is a blockquote with **bold** text and a [link](http://example.com).\n" +
		"> \n" +
		"> It also contains a list:\n" +
		"> - Item 1\n" +
		"> - Item 2\n\n" +
		"## Sub Heading\n\n" +
		"Here's a paragraph with `inline code` and *emphasis*.\n\n" +
		"```python\n" +
		"# This is a code block inside the document\n" +
		"def nested_function():\n" +
		"    return \"nested\"\n" +
		"```\n\n" +
		"- Outer list item 1\n" +
		"  - Nested list item 1\n" +
		"  - Nested list item 2\n" +
		"- Outer list item 2"

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Should handle nested structures gracefully
	expectedTypes := []string{"heading", "blockquote", "heading", "paragraph", "code", "list"}
	if len(chunks) != len(expectedTypes) {
		t.Fatalf("Expected %d chunks, got %d", len(expectedTypes), len(chunks))
	}

	for i, expectedType := range expectedTypes {
		if chunks[i].Type != expectedType {
			t.Errorf("Chunk %d: expected type '%s', got '%s'", i, expectedType, chunks[i].Type)
		}
	}
}

func TestChunkDocument_MalformedMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected int // expected number of chunks
	}{
		{
			name:     "Unclosed code block",
			markdown: "```python\nprint('hello')\n# Missing closing fence",
			expected: 1, // Should be treated as a code block
		},
		{
			name:     "Malformed table",
			markdown: "| Name | Age\n|---\n| John | 25 |\n| Jane",
			expected: 1, // Should still parse as table
		},
		{
			name:     "Mixed heading levels",
			markdown: "# H1\n### H3 (skipped H2)\n##### H5 (skipped H4)",
			expected: 3, // Should handle non-sequential heading levels
		},
		{
			name:     "Empty list items",
			markdown: "- Item 1\n- \n- Item 3",
			expected: 1, // Should handle empty list items
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if len(chunks) != tt.expected {
				t.Errorf("Expected %d chunks, got %d", tt.expected, len(chunks))
			}
		})
	}
}

func TestChunkDocument_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Unicode characters",
			markdown: "# 中文标题\n\n这是一个包含中文的段落。",
			expected: "中文标题",
		},
		{
			name:     "Emoji",
			markdown: "# Hello 👋 World 🌍\n\nThis paragraph has emojis! 🎉",
			expected: "Hello 👋 World 🌍",
		},
		{
			name:     "Special markdown characters",
			markdown: "This text has \\*escaped\\* \\[brackets\\] and \\`backticks\\`.",
			expected: "This text has \\*escaped\\* \\[brackets\\] and \\`backticks\\`.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if len(chunks) == 0 {
				t.Fatal("Expected at least 1 chunk")
			}
			if !strings.Contains(chunks[0].Text, tt.expected) {
				t.Errorf("Expected text to contain '%s', got '%s'", tt.expected, chunks[0].Text)
			}
		})
	}
}

func TestChunkDocument_EdgeCaseContent(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		validate func(t *testing.T, chunks []Chunk)
	}{
		{
			name:     "Only whitespace",
			markdown: "   \n\n\t\t\n   ",
			validate: func(t *testing.T, chunks []Chunk) {
				if len(chunks) != 0 {
					t.Errorf("Expected 0 chunks for whitespace-only content, got %d", len(chunks))
				}
			},
		},
		{
			name:     "Single character",
			markdown: "a",
			validate: func(t *testing.T, chunks []Chunk) {
				if len(chunks) != 1 || chunks[0].Type != "paragraph" {
					t.Errorf("Expected 1 paragraph chunk, got %d chunks", len(chunks))
				}
				if chunks[0].Text != "a" {
					t.Errorf("Expected text 'a', got '%s'", chunks[0].Text)
				}
			},
		},
		{
			name:     "Multiple consecutive empty lines",
			markdown: "# Title\n\n\n\n\n\nParagraph",
			validate: func(t *testing.T, chunks []Chunk) {
				if len(chunks) != 2 {
					t.Errorf("Expected 2 chunks, got %d", len(chunks))
				}
				if chunks[0].Type != "heading" || chunks[1].Type != "paragraph" {
					t.Errorf("Expected heading and paragraph, got %s and %s", chunks[0].Type, chunks[1].Type)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			tt.validate(t, chunks)
		})
	}
}

func TestChunkDocument_MetadataAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		validate func(t *testing.T, chunk Chunk)
	}{
		{
			name:     "Code block line count accuracy",
			markdown: "```go\nline1\nline2\n\nline4\n```",
			validate: func(t *testing.T, chunk Chunk) {
				if chunk.Type != "code" {
					t.Errorf("Expected code chunk, got %s", chunk.Type)
				}
				if chunk.Metadata["line_count"] != "4" {
					t.Errorf("Expected 4 lines, got %s", chunk.Metadata["line_count"])
				}
				if chunk.Metadata["language"] != "go" {
					t.Errorf("Expected language 'go', got %s", chunk.Metadata["language"])
				}
			},
		},
		{
			name:     "Paragraph word count with punctuation",
			markdown: "Hello, world! This is a test sentence with 10 words total.",
			validate: func(t *testing.T, chunk Chunk) {
				if chunk.Type != "paragraph" {
					t.Errorf("Expected paragraph chunk, got %s", chunk.Type)
				}
				if chunk.Metadata["word_count"] != "11" {
					t.Errorf("Expected 11 words, got %s", chunk.Metadata["word_count"])
				}
			},
		},
		{
			name:     "Complex table dimensions",
			markdown: "| A | B | C | D |\n|---|---|---|---|\n| 1 | 2 | 3 | 4 |\n| 5 | 6 | 7 | 8 |\n| 9 | 10| 11| 12|",
			validate: func(t *testing.T, chunk Chunk) {
				if chunk.Type != "table" {
					t.Errorf("Expected table chunk, got %s", chunk.Type)
				}
				if chunk.Metadata["rows"] != "4" {
					t.Errorf("Expected 4 rows, got %s", chunk.Metadata["rows"])
				}
				if chunk.Metadata["columns"] != "4" {
					t.Errorf("Expected 4 columns, got %s", chunk.Metadata["columns"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if len(chunks) == 0 {
				t.Fatal("Expected at least 1 chunk")
			}
			tt.validate(t, chunks[0])
		})
	}
}

func TestChunkDocument_IDSequencing(t *testing.T) {
	markdown := `# Heading 1

Paragraph 1

## Heading 2

Paragraph 2

- List item 1
- List item 2

> Blockquote

---`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Verify IDs are sequential starting from 0
	for i, chunk := range chunks {
		if chunk.ID != i {
			t.Errorf("Chunk %d has ID %d, expected %d", i, chunk.ID, i)
		}
	}
}

func TestChunkDocument_ContentPreservation(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		validate func(t *testing.T, chunk Chunk)
	}{
		{
			name:     "Preserve code block formatting",
			markdown: "```python\ndef hello():\n    print('world')\n```",
			validate: func(t *testing.T, chunk Chunk) {
				expected := "```python\ndef hello():\n    print('world')\n```"
				if chunk.Content != expected {
					t.Errorf("Content not preserved correctly.\nExpected: %q\nGot: %q", expected, chunk.Content)
				}
			},
		},
		{
			name:     "Preserve table formatting",
			markdown: "| Name | Age |\n|------|-----|\n| John | 25  |",
			validate: func(t *testing.T, chunk Chunk) {
				if !strings.Contains(chunk.Content, "|") {
					t.Errorf("Table formatting not preserved in content: %q", chunk.Content)
				}
				if !strings.Contains(chunk.Content, "---") {
					t.Errorf("Table separator not preserved in content: %q", chunk.Content)
				}
			},
		},
		{
			name:     "Preserve blockquote formatting",
			markdown: "> Line 1\n> Line 2",
			validate: func(t *testing.T, chunk Chunk) {
				expected := "> Line 1\n> Line 2"
				if chunk.Content != expected {
					t.Errorf("Blockquote formatting not preserved.\nExpected: %q\nGot: %q", expected, chunk.Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}
			if len(chunks) == 0 {
				t.Fatal("Expected at least 1 chunk")
			}
			tt.validate(t, chunks[0])
		})
	}
}

func TestChunkDocument_ConcurrentAccess(t *testing.T) {
	// Test that multiple chunker instances can be used concurrently
	// Note: The current implementation is not thread-safe for the same instance
	markdown := "# Test\n\nConcurrent access test."

	const numGoroutines = 10
	results := make(chan []Chunk, numGoroutines)
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			// Create a new chunker instance for each goroutine
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(markdown))
			if err != nil {
				errors <- err
				return
			}
			results <- chunks
		}()
	}

	// Collect results
	for range numGoroutines {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent access failed: %v", err)
		case chunks := <-results:
			if len(chunks) != 2 {
				t.Errorf("Expected 2 chunks, got %d", len(chunks))
			}
		}
	}
}

func TestChunkDocument_MemoryEfficiency(t *testing.T) {
	// Test that the chunker doesn't leak memory with repeated use
	chunker := NewMarkdownChunker()
	markdown := "# Test\n\nMemory test paragraph."

	// Process the same content multiple times
	for i := range 1000 {
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error on iteration %d: %v", i, err)
		}
		if len(chunks) != 2 {
			t.Errorf("Expected 2 chunks on iteration %d, got %d", i, len(chunks))
		}

		// Verify chunks slice is reset properly
		if len(chunker.chunks) != 2 {
			t.Errorf("Internal chunks not managed properly on iteration %d", i)
		}
	}
}
