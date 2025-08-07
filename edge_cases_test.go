package markdownchunker

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// TestEdgeCases tests various edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		verify   func(t *testing.T, chunks []Chunk, err error)
	}{
		{
			name:     "Empty string",
			markdown: "",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) != 0 {
					t.Errorf("Expected 0 chunks for empty string, got %d", len(chunks))
				}
			},
		},
		{
			name:     "Only whitespace",
			markdown: "   \n\n\t\t\n   ",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) != 0 {
					t.Errorf("Expected 0 chunks for whitespace-only content, got %d", len(chunks))
				}
			},
		},
		{
			name:     "Single character",
			markdown: "a",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) != 1 {
					t.Errorf("Expected 1 chunk for single character, got %d", len(chunks))
				}
				if len(chunks) > 0 && chunks[0].Type != "paragraph" {
					t.Errorf("Expected paragraph chunk, got %s", chunks[0].Type)
				}
			},
		},
		{
			name:     "Unicode characters",
			markdown: "# 中文标题\n\n这是一个包含中文的段落。\n\n```go\n// 中文注释\nfmt.Println(\"你好世界\")\n```",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) < 3 {
					t.Errorf("Expected at least 3 chunks, got %d", len(chunks))
				}
				// Verify Unicode is preserved
				foundUnicode := false
				for _, chunk := range chunks {
					if strings.Contains(chunk.Text, "中文") || strings.Contains(chunk.Text, "你好世界") {
						foundUnicode = true
						break
					}
				}
				if !foundUnicode {
					t.Error("Expected to find Unicode characters in chunks")
				}
			},
		},
		{
			name:     "Emoji characters",
			markdown: "# Hello 👋 World 🌍\n\nThis paragraph has emojis! 🎉 🚀 ✨",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(chunks) == 0 {
					t.Error("Expected some chunks")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			chunks, err := chunker.ChunkDocument([]byte(tt.markdown))
			tt.verify(t, chunks, err)
		})
	}
}

// TestMalformedMarkdown tests handling of malformed markdown
func TestMalformedMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		generateDoc func() []byte
		maxTime     time.Duration
		verify      func(t *testing.T, chunks []Chunk, err error)
	}{
		{
			name: "Empty document",
			generateDoc: func() []byte {
				return []byte("")
			},
			maxTime: 1 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Only whitespace",
			generateDoc: func() []byte {
				return []byte("   \n\n\t\t\n   ")
			},
			maxTime: 1 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Single character",
			generateDoc: func() []byte {
				return []byte("a")
			},
			maxTime: 1 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Very long single line",
			generateDoc: func() []byte {
				return []byte(strings.Repeat("a", 1000000))
			},
			maxTime: 5 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Many empty lines",
			generateDoc: func() []byte {
				return []byte(strings.Repeat("\n", 10000))
			},
			maxTime: 2 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Unicode content",
			generateDoc: func() []byte {
				return []byte("# 测试文档\n\n这是一个包含中文的段落。\n\n```python\n# 代码注释\nprint('你好世界')\n```\n\n| 名称 | 值 |\n|------|----|\n| 测试 | 123 |")
			},
			maxTime: 2 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Deeply nested structure",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 1; i <= 6; i++ {
					doc.WriteString(strings.Repeat("#", i))
					doc.WriteString(fmt.Sprintf(" Heading Level %d\n\n", i))
					doc.WriteString("Content for this level.\n\n")
				}
				return []byte(doc.String())
			},
			maxTime: 2 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Mixed content types",
			generateDoc: func() []byte {
				return []byte(`# Mixed Content

Paragraph with **bold** and *italic*.

` + "```go" + `
func test() {}
` + "```" + `

| A | B |
|---|---|
| 1 | 2 |

> Quote

- List item

1. Numbered item

---

Final paragraph.`)
			},
			maxTime: 2 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		{
			name: "Repeated identical content",
			generateDoc: func() []byte {
				content := "# Same Heading\n\nSame paragraph content.\n\n```go\nfunc same() {}\n```\n\n"
				return []byte(strings.Repeat(content, 1000))
			},
			maxTime: 2 * time.Second,
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if len(chunks) == 0 {
					t.Error("Expected some chunks")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.generateDoc()
			chunker := NewMarkdownChunker()

			start := time.Now()
			chunks, err := chunker.ChunkDocument(doc)
			elapsed := time.Since(start)

			if elapsed > tt.maxTime {
				t.Errorf("Processing took too long: %v (max: %v)", elapsed, tt.maxTime)
			}

			if tt.verify != nil {
				tt.verify(t, chunks, err)
			}
		})
	}
}

// TestBoundaryConditions tests boundary conditions and limits
func TestBoundaryConditions(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() ([]byte, *ChunkerConfig)
		verify func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker)
	}{
		{
			name: "Maximum heading level",
			setup: func() ([]byte, *ChunkerConfig) {
				markdown := "###### Level 6 Heading\n\n####### Level 7 (invalid)\n\n######## Level 8 (invalid)"
				return []byte(markdown), DefaultConfig()
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				headingCount := 0
				for _, chunk := range chunks {
					if chunk.Type == "heading" {
						headingCount++
						if chunk.Level > 6 {
							t.Errorf("Heading level should not exceed 6, got %d", chunk.Level)
						}
					}
				}
				// Should have at least one valid heading
				if headingCount == 0 {
					t.Error("Expected at least one valid heading")
				}
			},
		},
		{
			name: "Very large table",
			setup: func() ([]byte, *ChunkerConfig) {
				var markdown strings.Builder
				// Create table header
				markdown.WriteString("|")
				for i := 0; i < 50; i++ {
					markdown.WriteString(fmt.Sprintf(" Col%d |", i))
				}
				markdown.WriteString("\n|")
				for i := 0; i < 50; i++ {
					markdown.WriteString("------|")
				}
				markdown.WriteString("\n")
				// Create 100 rows
				for row := 0; row < 100; row++ {
					markdown.WriteString("|")
					for col := 0; col < 50; col++ {
						markdown.WriteString(fmt.Sprintf(" R%dC%d |", row, col))
					}
					markdown.WriteString("\n")
				}
				return []byte(markdown.String()), DefaultConfig()
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) != 1 {
					t.Errorf("Expected 1 table chunk, got %d", len(chunks))
				}
				if len(chunks) > 0 {
					if chunks[0].Metadata["columns"] != "50" {
						t.Errorf("Expected 50 columns, got %s", chunks[0].Metadata["columns"])
					}
					if chunks[0].Metadata["rows"] != "101" { // 100 data rows + 1 header
						t.Errorf("Expected 101 rows, got %s", chunks[0].Metadata["rows"])
					}
				}
			},
		},
		{
			name: "Extremely long code block",
			setup: func() ([]byte, *ChunkerConfig) {
				var markdown strings.Builder
				markdown.WriteString("```go\n")
				for i := 0; i < 10000; i++ {
					markdown.WriteString(fmt.Sprintf("// Line %d\nfmt.Printf(\"Line %%d\\n\", %d)\n", i, i))
				}
				markdown.WriteString("```")
				return []byte(markdown.String()), DefaultConfig()
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) != 1 {
					t.Errorf("Expected 1 code chunk, got %d", len(chunks))
				}
				if len(chunks) > 0 {
					if chunks[0].Type != "code" {
						t.Errorf("Expected code chunk, got %s", chunks[0].Type)
					}
					if chunks[0].Metadata["line_count"] != "20000" {
						t.Errorf("Expected 20000 lines, got %s", chunks[0].Metadata["line_count"])
					}
				}
			},
		},
		{
			name: "Memory limit exceeded",
			setup: func() ([]byte, *ChunkerConfig) {
				// Create a document that should exceed memory limit
				var markdown strings.Builder
				for i := 0; i < 1000; i++ {
					markdown.WriteString(fmt.Sprintf("# Section %d\n\n", i))
					markdown.WriteString(strings.Repeat("This is a very long paragraph with lots of content. ", 100))
					markdown.WriteString("\n\n")
				}
				config := DefaultConfig()
				config.MemoryLimit = 1024 * 1024 // 1MB limit
				config.ErrorHandling = ErrorModePermissive
				return []byte(markdown.String()), config
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				// Should handle gracefully in permissive mode
				if err != nil {
					t.Errorf("Unexpected error in permissive mode: %v", err)
				}
				// Should still produce chunks
				if len(chunks) == 0 {
					t.Error("Expected some chunks even with memory limit")
				}
			},
		},
		{
			name: "Zero-width characters",
			setup: func() ([]byte, *ChunkerConfig) {
				// Include zero-width space, zero-width non-joiner, etc.
				markdown := "# Title\u200B with\u200C zero\u200D width\u2060 chars\n\nParagraph\uFEFF with\u00AD more\u034F special\u061C chars."
				return []byte(markdown), DefaultConfig()
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(chunks) < 2 {
					t.Errorf("Expected at least 2 chunks, got %d", len(chunks))
				}
				// Verify content is valid UTF-8
				for _, chunk := range chunks {
					if !utf8.ValidString(chunk.Content) {
						t.Error("Chunk content is not valid UTF-8")
					}
					if !utf8.ValidString(chunk.Text) {
						t.Error("Chunk text is not valid UTF-8")
					}
				}
			},
		},
		{
			name: "Maximum nesting depth",
			setup: func() ([]byte, *ChunkerConfig) {
				var markdown strings.Builder
				// Create deeply nested blockquotes
				for i := 0; i < 50; i++ {
					markdown.WriteString(strings.Repeat("> ", i+1))
					markdown.WriteString(fmt.Sprintf("Level %d content\n", i+1))
				}
				return []byte(markdown.String()), DefaultConfig()
			},
			verify: func(t *testing.T, chunks []Chunk, err error, chunker *MarkdownChunker) {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Should handle deep nesting gracefully
				if len(chunks) == 0 {
					t.Error("Expected some chunks from nested content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, config := tt.setup()
			chunker := NewMarkdownChunkerWithConfig(config)
			chunks, err := chunker.ChunkDocument(content)
			tt.verify(t, chunks, err, chunker)
		})
	}
}

// TestSpecialCharacterHandling tests handling of various special characters
func TestSpecialCharacterHandling(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "HTML entities",
			markdown: "This text has &lt;HTML&gt; entities &amp; more.",
			expected: "This text has &lt;HTML&gt; entities &amp; more.",
		},
		{
			name:     "Escaped markdown",
			markdown: "This text has \\*escaped\\* \\[brackets\\] and \\`backticks\\`.",
			expected: "This text has \\*escaped\\* \\[brackets\\] and \\`backticks\\`.",
		},
		{
			name:     "Mathematical symbols",
			markdown: "Formula: α + β = γ, ∑(x²) ≥ 0, π ≈ 3.14159",
			expected: "Formula: α + β = γ, ∑(x²) ≥ 0, π ≈ 3.14159",
		},
		{
			name:     "Currency symbols",
			markdown: "Prices: $100, €85, ¥1000, £75, ₹5000",
			expected: "Prices: $100, €85, ¥1000, £75, ₹5000",
		},
		{
			name:     "Punctuation marks",
			markdown: "Various punctuation: \"quotes\", 'apostrophes', —dashes—, …ellipsis…",
			expected: "Various punctuation: \"quotes\", 'apostrophes', —dashes—, …ellipsis…",
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
				t.Errorf("Expected text %q, got %q", tt.expected, chunks[0].Text)
			}
		})
	}
}

// TestPerformanceEdgeCases tests performance with edge case inputs
func TestPerformanceEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		generateDoc func() []byte
		maxTime     time.Duration
	}{
		{
			name: "Many empty lines",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 0; i < 10000; i++ {
					if i%100 == 0 {
						doc.WriteString(fmt.Sprintf("# Heading %d\n", i/100))
					}
					doc.WriteString("\n")
				}
				return []byte(doc.String())
			},
			maxTime: 2 * time.Second,
		},
		{
			name: "Alternating content types",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 0; i < 1000; i++ {
					switch i % 6 {
					case 0:
						doc.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
					case 1:
						doc.WriteString(fmt.Sprintf("Paragraph %d with some content.\n\n", i))
					case 2:
						doc.WriteString(fmt.Sprintf("```\ncode block %d\n```\n\n", i))
					case 3:
						doc.WriteString(fmt.Sprintf("- List item %d\n\n", i))
					case 4:
						doc.WriteString(fmt.Sprintf("> Quote %d\n\n", i))
					case 5:
						doc.WriteString("---\n\n")
					}
				}
				return []byte(doc.String())
			},
			maxTime: 3 * time.Second,
		},
		{
			name: "Repeated identical content",
			generateDoc: func() []byte {
				content := "# Same Heading\n\nSame paragraph content.\n\n```go\nfunc same() {}\n```\n\n"
				return []byte(strings.Repeat(content, 1000))
			},
			maxTime: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.generateDoc()
			chunker := NewMarkdownChunker()

			start := time.Now()
			chunks, err := chunker.ChunkDocument(doc)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}

			if elapsed > tt.maxTime {
				t.Errorf("Processing took too long: %v (max: %v)", elapsed, tt.maxTime)
			}

			if len(chunks) == 0 {
				t.Error("Expected some chunks")
			}

			t.Logf("Edge case performance test '%s':", tt.name)
			t.Logf("  Document size: %d bytes", len(doc))
			t.Logf("  Chunks produced: %d", len(chunks))
			t.Logf("  Processing time: %v", elapsed)
		})
	}
}
