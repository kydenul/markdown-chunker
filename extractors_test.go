package markdownchunker

import (
	"slices"
	"strings"
	"testing"
)

func TestLinkExtractor(t *testing.T) {
	markdown := `This is a paragraph with [a link](https://example.com) and [another link](https://test.com).`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	// 测试链接提取器
	extractor := &LinkExtractor{}

	// 需要获取AST节点，这里我们模拟一个简单的测试
	supportedTypes := extractor.SupportedTypes()
	expectedTypes := []string{"paragraph", "heading", "blockquote", "list"}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}

	for i, expectedType := range expectedTypes {
		if supportedTypes[i] != expectedType {
			t.Errorf("Expected type %s, got %s", expectedType, supportedTypes[i])
		}
	}
}

func TestImageExtractor(t *testing.T) {
	extractor := &ImageExtractor{}

	supportedTypes := extractor.SupportedTypes()
	expectedTypes := []string{"paragraph", "heading", "blockquote", "list"}

	if len(supportedTypes) != len(expectedTypes) {
		t.Errorf("Expected %d supported types, got %d", len(expectedTypes), len(supportedTypes))
	}
}

func TestCodeComplexityExtractor(t *testing.T) {
	extractor := &CodeComplexityExtractor{}

	supportedTypes := extractor.SupportedTypes()
	if len(supportedTypes) != 1 || supportedTypes[0] != "code" {
		t.Errorf("Expected ['code'], got %v", supportedTypes)
	}
}

func TestChunkDocumentWithCustomExtractors(t *testing.T) {
	markdown := `# Heading with [link](https://example.com)

This is a paragraph with [a link](https://example.com) and ![image](image.jpg).

` + "```go" + `
func main() {
    if true {
        for i := 0; i < 10; i++ {
            fmt.Println("Hello")
        }
    }
}
` + "```"

	config := DefaultConfig()
	config.CustomExtractors = []MetadataExtractor{
		&LinkExtractor{},
		&ImageExtractor{},
		&CodeComplexityExtractor{},
	}

	chunker := NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// 检查是否有链接元数据
	foundLinkMetadata := false
	foundImageMetadata := false
	foundCodeMetadata := false

	for _, chunk := range chunks {
		if chunk.Type == "heading" || chunk.Type == "paragraph" {
			if _, exists := chunk.Metadata["link_count"]; exists {
				foundLinkMetadata = true
			}
			if _, exists := chunk.Metadata["image_count"]; exists {
				foundImageMetadata = true
			}
		}
		if chunk.Type == "code" {
			if _, exists := chunk.Metadata["code_complexity"]; exists {
				foundCodeMetadata = true
			}
		}
	}

	if !foundLinkMetadata {
		t.Error("Expected to find link metadata")
	}
	if !foundImageMetadata {
		t.Error("Expected to find image metadata")
	}
	if !foundCodeMetadata {
		t.Error("Expected to find code complexity metadata")
	}
}

func TestFilterEmptyChunks(t *testing.T) {
	markdown := `# Heading

   

Another paragraph.`

	t.Run("filter empty chunks enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.FilterEmptyChunks = true

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// 检查是否过滤了空块
		for _, chunk := range chunks {
			if strings.TrimSpace(chunk.Text) == "" {
				t.Errorf("Found empty chunk when FilterEmptyChunks is enabled: %+v", chunk)
			}
		}
	})

	t.Run("filter empty chunks disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.FilterEmptyChunks = false

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// 应该包含所有块，包括可能的空块
		// 这个测试主要确保不会因为空块而崩溃
		if len(chunks) == 0 {
			t.Error("Expected some chunks even with FilterEmptyChunks disabled")
		}
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("min function", func(t *testing.T) {
		tests := []struct {
			a, b, expected int
		}{
			{1, 2, 1},
			{5, 3, 3},
			{0, 0, 0},
			{-1, 1, -1},
		}

		for _, test := range tests {
			result := min(test.a, test.b)
			if result != test.expected {
				t.Errorf("min(%d, %d) = %d, expected %d", test.a, test.b, result, test.expected)
			}
		}
	})

	t.Run("contains function", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}

		if !slices.Contains(slice, "banana") {
			t.Error("Expected contains to return true for 'banana'")
		}

		if slices.Contains(slice, "grape") {
			t.Error("Expected contains to return false for 'grape'")
		}

		if slices.Contains([]string{}, "anything") {
			t.Error("Expected contains to return false for empty slice")
		}
	})
}
