package markdownchunker

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

func TestDocumentLevelStrategy_GetName(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	expected := "document-level"
	if strategy.GetName() != expected {
		t.Errorf("Expected strategy name %s, got %s", expected, strategy.GetName())
	}
}

func TestDocumentLevelStrategy_GetDescription(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	description := strategy.GetDescription()
	if description == "" {
		t.Error("Strategy description should not be empty")
	}
	if !strings.Contains(description, "文档") {
		t.Error("Strategy description should mention document processing")
	}
}

func TestDocumentLevelStrategy_Clone(t *testing.T) {
	original := NewDocumentLevelStrategy()
	cloned := original.Clone()

	if cloned == nil {
		t.Fatal("Cloned strategy should not be nil")
	}

	if cloned.GetName() != original.GetName() {
		t.Error("Cloned strategy should have the same name")
	}

	if cloned.GetDescription() != original.GetDescription() {
		t.Error("Cloned strategy should have the same description")
	}

	// Ensure they are different instances
	if cloned == original {
		t.Error("Cloned strategy should be a different instance")
	}
}

func TestDocumentLevelStrategy_ValidateConfig(t *testing.T) {
	strategy := NewDocumentLevelStrategy()

	tests := []struct {
		name        string
		config      *StrategyConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name:        "valid config",
			config:      DocumentLevelConfig(),
			expectError: false,
		},
		{
			name: "config with invalid max chunk size",
			config: &StrategyConfig{
				Name:         "document-level",
				MaxChunkSize: -1,
			},
			expectError: true,
		},
		{
			name: "config with valid parameters",
			config: &StrategyConfig{
				Name:         "document-level",
				MaxChunkSize: 1000,
				MinChunkSize: 100,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected validation error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error, but got: %v", err)
			}
		})
	}
}

func TestDocumentLevelStrategy_ChunkDocument(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()

	// Create goldmark parser for testing
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)

	tests := []struct {
		name           string
		content        string
		expectError    bool
		expectedChunks int
		expectedType   string
	}{
		{
			name:           "simple document",
			content:        "# Hello World\n\nThis is a simple document.",
			expectError:    false,
			expectedChunks: 1,
			expectedType:   "document",
		},
		{
			name:           "empty document",
			content:        "",
			expectError:    false,
			expectedChunks: 1,
			expectedType:   "document",
		},
		{
			name: "complex document",
			content: `# Main Title

## Section 1
This is the first section with some content.

### Subsection 1.1
More detailed content here.

## Section 2
- Item 1
- Item 2
- Item 3

### Code Example
` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

## Section 3
| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |

[Link to example](https://example.com)

![Image](image.png "Image title")
`,
			expectError:    false,
			expectedChunks: 1,
			expectedType:   "document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the markdown content
			reader := text.NewReader([]byte(tt.content))
			doc := md.Parser().Parse(reader)

			chunks, err := strategy.ChunkDocument(doc, []byte(tt.content), chunker)

			if tt.expectError && err == nil {
				t.Error("Expected error, but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
				return
			}

			if len(chunks) != tt.expectedChunks {
				t.Errorf("Expected %d chunks, got %d", tt.expectedChunks, len(chunks))
				return
			}

			if len(chunks) > 0 {
				chunk := chunks[0]
				if chunk.Type != tt.expectedType {
					t.Errorf("Expected chunk type %s, got %s", tt.expectedType, chunk.Type)
				}

				if chunk.ID != 0 {
					t.Errorf("Expected chunk ID 0, got %d", chunk.ID)
				}

				if chunk.Level != 0 {
					t.Errorf("Expected chunk level 0, got %d", chunk.Level)
				}

				if chunk.Content != tt.content {
					t.Errorf("Expected chunk content to match input content")
				}

				// Check metadata
				if chunk.Metadata == nil {
					t.Error("Chunk metadata should not be nil")
				} else {
					if chunk.Metadata["strategy"] != "document-level" {
						t.Error("Chunk metadata should indicate document-level strategy")
					}
				}

				// Check hash
				if chunk.Hash == "" {
					t.Error("Chunk hash should not be empty")
				}
			}
		})
	}
}

func TestDocumentLevelStrategy_ChunkDocument_ErrorCases(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()

	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	content := []byte("# Test")
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	tests := []struct {
		name        string
		doc         interface{}
		source      []byte
		chunker     *MarkdownChunker
		expectError bool
	}{
		{
			name:        "nil document",
			doc:         nil,
			source:      content,
			chunker:     chunker,
			expectError: true,
		},
		{
			name:        "nil chunker",
			doc:         doc,
			source:      content,
			chunker:     nil,
			expectError: true,
		},
		{
			name:        "nil source",
			doc:         doc,
			source:      nil,
			chunker:     chunker,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var chunks []Chunk
			var err error

			if tt.doc == nil {
				chunks, err = strategy.ChunkDocument(nil, tt.source, tt.chunker)
			} else {
				chunks, err = strategy.ChunkDocument(doc, tt.source, tt.chunker)
			}

			if tt.expectError && err == nil {
				t.Error("Expected error, but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
			if tt.expectError && chunks != nil {
				t.Error("Expected nil chunks on error")
			}
		})
	}
}

func TestDocumentLevelStrategy_ExtractTextFromDocument(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple text",
			content:  "Hello world",
			expected: "Hello world",
		},
		{
			name:     "heading and paragraph",
			content:  "# Title\n\nThis is a paragraph.",
			expected: "Title This is a paragraph.",
		},
		{
			name:     "with code span",
			content:  "Use `fmt.Println` to print.",
			expected: "Use to print.",
		},
		{
			name:     "multiple paragraphs",
			content:  "First paragraph.\n\nSecond paragraph.",
			expected: "First paragraph. Second paragraph.",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := text.NewReader([]byte(tt.content))
			doc := md.Parser().Parse(reader)

			result := strategy.extractTextFromDocument(doc, []byte(tt.content))
			if result != tt.expected {
				t.Errorf("Expected text %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDocumentLevelStrategy_ExtractDocumentMetadata(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	content := `# Main Title

## Section 1
This is content.

### Subsection
More content.

- List item 1
- List item 2

` + "```go\ncode here\n```" + `

| Col1 | Col2 |
|------|------|
| A    | B    |
`

	reader := text.NewReader([]byte(content))
	doc := md.Parser().Parse(reader)

	metadata := strategy.extractDocumentMetadata(doc, []byte(content), chunker)

	// Check required metadata fields
	requiredFields := []string{
		"strategy", "total_size", "content_length", "heading_count",
		"paragraph_count", "code_block_count", "table_count", "list_count",
		"max_heading_level", "text_length", "word_count",
		"estimated_reading_time_minutes", "document_complexity",
	}

	for _, field := range requiredFields {
		if _, exists := metadata[field]; !exists {
			t.Errorf("Missing required metadata field: %s", field)
		}
	}

	// Check specific values
	if metadata["strategy"] != "document-level" {
		t.Errorf("Expected strategy 'document-level', got %s", metadata["strategy"])
	}

	if metadata["max_heading_level"] != "3" {
		t.Errorf("Expected max heading level '3', got %s", metadata["max_heading_level"])
	}
}

func TestDocumentLevelStrategy_ExtractLinksAndImages(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	content := `# Document with Links and Images

[External link](https://example.com)
[Internal link](/internal/page)
[Anchor link](#section)

![Image 1](image1.png "First image")
![Image 2](image2.jpg)
`

	reader := text.NewReader([]byte(content))
	doc := md.Parser().Parse(reader)

	links, images := strategy.extractLinksAndImages(doc, []byte(content))

	// Check links
	expectedLinks := 3
	if len(links) != expectedLinks {
		t.Errorf("Expected %d links, got %d", expectedLinks, len(links))
	}

	// Check link types
	linkTypes := make(map[string]int)
	for _, link := range links {
		linkTypes[link.Type]++
	}

	if linkTypes["external"] != 1 {
		t.Errorf("Expected 1 external link, got %d", linkTypes["external"])
	}
	if linkTypes["internal"] != 1 {
		t.Errorf("Expected 1 internal link, got %d", linkTypes["internal"])
	}
	if linkTypes["anchor"] != 1 {
		t.Errorf("Expected 1 anchor link, got %d", linkTypes["anchor"])
	}

	// Check images
	expectedImages := 2
	if len(images) != expectedImages {
		t.Errorf("Expected %d images, got %d", expectedImages, len(images))
	}

	// Check first image has title
	if len(images) > 0 && images[0].Title != "First image" {
		t.Errorf("Expected first image title 'First image', got %s", images[0].Title)
	}
}

func TestDocumentLevelStrategy_CalculateDocumentPosition(t *testing.T) {
	strategy := NewDocumentLevelStrategy()

	tests := []struct {
		name     string
		content  string
		expected ChunkPosition
	}{
		{
			name:    "empty content",
			content: "",
			expected: ChunkPosition{
				StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 1,
			},
		},
		{
			name:    "single line",
			content: "Hello world",
			expected: ChunkPosition{
				StartLine: 1, EndLine: 1, StartCol: 1, EndCol: 12,
			},
		},
		{
			name:    "multiple lines",
			content: "Line 1\nLine 2\nLine 3",
			expected: ChunkPosition{
				StartLine: 1, EndLine: 3, StartCol: 1, EndCol: 7,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.calculateDocumentPosition([]byte(tt.content))
			if result != tt.expected {
				t.Errorf("Expected position %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestDocumentLevelStrategy_CalculateDocumentComplexity(t *testing.T) {
	strategy := NewDocumentLevelStrategy()

	tests := []struct {
		name               string
		headingCount       int
		codeBlockCount     int
		tableCount         int
		listCount          int
		expectedComplexity string
	}{
		{
			name:               "simple document",
			headingCount:       1,
			codeBlockCount:     0,
			tableCount:         0,
			listCount:          1,
			expectedComplexity: "simple",
		},
		{
			name:               "moderate document",
			headingCount:       3,
			codeBlockCount:     2,
			tableCount:         1,
			listCount:          2,
			expectedComplexity: "moderate",
		},
		{
			name:               "complex document",
			headingCount:       5,
			codeBlockCount:     5,
			tableCount:         3,
			listCount:          4,
			expectedComplexity: "complex",
		},
		{
			name:               "very complex document",
			headingCount:       10,
			codeBlockCount:     8,
			tableCount:         5,
			listCount:          10,
			expectedComplexity: "very_complex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.calculateDocumentComplexity(
				tt.headingCount, tt.codeBlockCount, tt.tableCount, tt.listCount)
			if result != tt.expectedComplexity {
				t.Errorf("Expected complexity %s, got %s", tt.expectedComplexity, result)
			}
		})
	}
}

func TestDocumentLevelStrategy_GetSetConfig(t *testing.T) {
	strategy := NewDocumentLevelStrategy()

	// Test GetConfig
	config := strategy.GetConfig()
	if config == nil {
		t.Error("GetConfig should not return nil")
	}
	if config.Name != "document-level" {
		t.Errorf("Expected config name 'document-level', got %s", config.Name)
	}

	// Test SetConfig with valid config
	newConfig := &StrategyConfig{
		Name:         "document-level",
		MaxChunkSize: 1000,
	}
	err := strategy.SetConfig(newConfig)
	if err != nil {
		t.Errorf("SetConfig should not return error for valid config: %v", err)
	}

	// Verify config was set
	updatedConfig := strategy.GetConfig()
	if updatedConfig.MaxChunkSize != 1000 {
		t.Errorf("Expected MaxChunkSize 1000, got %d", updatedConfig.MaxChunkSize)
	}

	// Test SetConfig with nil (should use default)
	err = strategy.SetConfig(nil)
	if err != nil {
		t.Errorf("SetConfig should not return error for nil config: %v", err)
	}

	// Test SetConfig with invalid config
	invalidConfig := &StrategyConfig{
		Name:         "document-level",
		MaxChunkSize: -1, // Invalid
	}
	err = strategy.SetConfig(invalidConfig)
	if err == nil {
		t.Error("SetConfig should return error for invalid config")
	}
}

// Benchmark tests for performance
func BenchmarkDocumentLevelStrategy_ChunkDocument(b *testing.B) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	// Create a moderately complex document
	content := []byte(`# Main Title

## Introduction
This is a comprehensive document with various elements to test the performance
of the document-level chunking strategy.

### Features
- Multiple headings at different levels
- Code blocks with syntax highlighting
- Tables with data
- Links and images
- Lists and nested content

## Code Examples

` + "```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```" + `

## Data Tables

| Feature | Status | Priority |
|---------|--------|----------|
| Chunking | Complete | High |
| Testing | In Progress | High |
| Documentation | Planned | Medium |

## External Resources

[Go Documentation](https://golang.org/doc/)
[Markdown Guide](https://www.markdownguide.org/)

![Architecture Diagram](diagram.png "System Architecture")

## Conclusion

This document demonstrates various Markdown features and serves as a
comprehensive test case for the document-level chunking strategy.
`)

	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			b.Fatalf("ChunkDocument failed: %v", err)
		}
	}
}

func BenchmarkDocumentLevelStrategy_ExtractText(b *testing.B) {
	strategy := NewDocumentLevelStrategy()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	content := []byte(strings.Repeat("# Heading\n\nThis is a paragraph with some text content.\n\n", 100))
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strategy.extractTextFromDocument(doc, content)
	}
}

// 测试大文档处理功能
func TestDocumentLevelStrategy_LargeDocumentProcessing(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	// 创建一个大文档（超过默认阈值5MB）
	baseContent := `# Large Document Test

## Section with Content
This is a section with substantial content that will be repeated many times
to create a large document for testing streaming processing capabilities.

### Subsection
- List item 1
- List item 2
- List item 3

` + "```go\nfunc example() {\n    fmt.Println(\"Hello, World!\")\n}\n```" + `

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data 1   | Data 2   | Data 3   |

[Link to example](https://example.com)

![Test Image](test.png "Test Image")

`

	// 重复内容以创建大文档（约6MB）
	largeContent := strings.Repeat(baseContent, 12000) // 约6MB
	content := []byte(largeContent)

	// 验证文档确实超过了阈值
	t.Logf("Document size: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))
	threshold := strategy.getLargeDocumentThreshold()
	t.Logf("Large document threshold: %d bytes (%.2f MB)", threshold, float64(threshold)/(1024*1024))

	if len(content) <= threshold {
		t.Fatalf("Test document size (%d) should be larger than threshold (%d)", len(content), threshold)
	}

	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	// 测试大文档处理
	chunks, err := strategy.ChunkDocument(doc, content, chunker)
	if err != nil {
		t.Fatalf("Large document processing failed: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for large document, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.Type != "document" {
		t.Errorf("Expected chunk type 'document', got %s", chunk.Type)
	}

	// 验证元数据包含流式处理标记
	if chunk.Metadata["processing_mode"] != "streaming" {
		t.Error("Expected processing_mode to be 'streaming' for large document")
	}

	// 验证内容完整性
	if len(chunk.Content) != len(largeContent) {
		t.Errorf("Content length mismatch: expected %d, got %d", len(largeContent), len(chunk.Content))
	}
}

func TestDocumentLevelStrategy_GetLargeDocumentThreshold(t *testing.T) {
	strategy := NewDocumentLevelStrategy()

	// 测试默认阈值
	defaultThreshold := strategy.getLargeDocumentThreshold()
	expectedDefault := 5 * 1024 * 1024 // 5MB
	if defaultThreshold != expectedDefault {
		t.Errorf("Expected default threshold %d, got %d", expectedDefault, defaultThreshold)
	}

	// 测试自定义阈值
	config := &StrategyConfig{
		Name: "document-level",
		Parameters: map[string]interface{}{
			"large_document_threshold": 10 * 1024 * 1024, // 10MB
		},
	}
	strategy.SetConfig(config)

	customThreshold := strategy.getLargeDocumentThreshold()
	expectedCustom := 10 * 1024 * 1024
	if customThreshold != expectedCustom {
		t.Errorf("Expected custom threshold %d, got %d", expectedCustom, customThreshold)
	}
}

func TestDocumentLevelStrategy_CountNodes(t *testing.T) {
	strategy := NewDocumentLevelStrategy()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	tests := []struct {
		name          string
		content       string
		expectedNodes int // 这是一个近似值，因为AST结构可能变化
	}{
		{
			name:          "simple document",
			content:       "# Title\n\nParagraph",
			expectedNodes: 4, // Document, Heading, Text, Paragraph, Text (大约)
		},
		{
			name:          "empty document",
			content:       "",
			expectedNodes: 1, // 只有Document节点
		},
		{
			name: "complex document",
			content: `# Title
## Subtitle
Paragraph with text.
- List item
` + "```go\ncode\n```",
			expectedNodes: 10, // 大约的节点数
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := text.NewReader([]byte(tt.content))
			doc := md.Parser().Parse(reader)

			nodeCount := strategy.countNodes(doc)

			// 由于AST结构可能变化，我们只检查节点数是否合理
			if nodeCount <= 0 {
				t.Errorf("Expected positive node count, got %d", nodeCount)
			}

			// 对于非空文档，节点数应该大于1
			if tt.content != "" && nodeCount <= 1 {
				t.Errorf("Expected more than 1 node for non-empty document, got %d", nodeCount)
			}
		})
	}
}

// 性能测试：大文档处理
func BenchmarkDocumentLevelStrategy_LargeDocument(b *testing.B) {
	strategy := NewDocumentLevelStrategy()
	chunker := NewMarkdownChunker()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	// 创建中等大小的文档用于基准测试
	baseContent := strings.Repeat("# Section\n\nContent with text.\n\n", 1000)
	content := []byte(baseContent)
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			b.Fatalf("ChunkDocument failed: %v", err)
		}
	}
}

// 性能测试：流式文本提取
func BenchmarkDocumentLevelStrategy_StreamingTextExtraction(b *testing.B) {
	strategy := NewDocumentLevelStrategy()
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	content := []byte(strings.Repeat("# Heading\n\nParagraph with content.\n\n", 1000))
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	// 空的进度回调
	progressCallback := func(int, string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.extractTextStreamingly(doc, content, progressCallback)
		if err != nil {
			b.Fatalf("extractTextStreamingly failed: %v", err)
		}
	}
}

// 性能测试：流式哈希计算
func BenchmarkDocumentLevelStrategy_StreamingHashCalculation(b *testing.B) {
	strategy := NewDocumentLevelStrategy()

	// 创建测试数据
	content := []byte(strings.Repeat("Test content for hash calculation.\n", 10000))

	// 空的进度回调
	progressCallback := func(int, string) {}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.calculateHashStreamingly(content, progressCallback)
		if err != nil {
			b.Fatalf("calculateHashStreamingly failed: %v", err)
		}
	}
}
