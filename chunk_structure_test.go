package markdownchunker

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChunkStructureExtensions(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试包含链接和图片的markdown内容
	content := `# Test Heading

This is a paragraph with a [link](https://example.com) and an ![image](image.jpg "Image Title").

## Another Heading

- List item with [internal link](./local.md)
- Another item with [anchor link](#section)

` + "```go" + `
func main() {
    fmt.Println("Hello World")
}
` + "```" + `

> This is a blockquote with [external link](https://github.com)

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |

---
`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证每个chunk都有新的字段
	for i, chunk := range chunks {
		t.Logf("Chunk %d: Type=%s", i, chunk.Type)

		// 验证Position字段
		if chunk.Position.StartLine <= 0 {
			t.Errorf("Chunk %d: Position.StartLine should be > 0, got %d", i, chunk.Position.StartLine)
		}
		if chunk.Position.EndLine < chunk.Position.StartLine {
			t.Errorf("Chunk %d: Position.EndLine (%d) should be >= StartLine (%d)", i, chunk.Position.EndLine, chunk.Position.StartLine)
		}

		// 验证Hash字段
		if chunk.Hash == "" {
			t.Errorf("Chunk %d: Hash should not be empty", i)
		}
		if len(chunk.Hash) != 64 { // SHA256 hex string length
			t.Errorf("Chunk %d: Hash should be 64 characters long, got %d", i, len(chunk.Hash))
		}

		// 验证Links和Images字段存在（即使为空）
		if chunk.Links == nil {
			t.Errorf("Chunk %d: Links should not be nil", i)
		}
		if chunk.Images == nil {
			t.Errorf("Chunk %d: Images should not be nil", i)
		}

		// 验证JSON序列化
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			t.Errorf("Chunk %d: Failed to marshal to JSON: %v", i, err)
		}

		// 验证JSON包含新字段
		jsonStr := string(jsonData)
		requiredFields := []string{"position", "links", "images", "hash"}
		for _, field := range requiredFields {
			if !strings.Contains(jsonStr, `"`+field+`"`) {
				t.Errorf("Chunk %d: JSON should contain field '%s'", i, field)
			}
		}
	}

	// 验证特定内容的链接和图片提取
	var paragraphChunk *Chunk
	for _, chunk := range chunks {
		if chunk.Type == "paragraph" && strings.Contains(chunk.Text, "link") {
			paragraphChunk = &chunk
			break
		}
	}

	if paragraphChunk == nil {
		t.Fatal("Could not find paragraph chunk with links")
	}

	// 验证链接提取
	if len(paragraphChunk.Links) < 1 {
		t.Errorf("Paragraph should contain at least 1 link, got %d", len(paragraphChunk.Links))
	} else {
		link := paragraphChunk.Links[0]
		if link.URL != "https://example.com" {
			t.Errorf("Expected link URL 'https://example.com', got '%s'", link.URL)
		}
		if link.Type != "external" {
			t.Errorf("Expected link type 'external', got '%s'", link.Type)
		}
	}

	// 验证图片提取
	if len(paragraphChunk.Images) < 1 {
		t.Errorf("Paragraph should contain at least 1 image, got %d", len(paragraphChunk.Images))
	} else {
		image := paragraphChunk.Images[0]
		if image.URL != "image.jpg" {
			t.Errorf("Expected image URL 'image.jpg', got '%s'", image.URL)
		}
		if image.Title != "Image Title" {
			t.Errorf("Expected image title 'Image Title', got '%s'", image.Title)
		}
	}
}

func TestLinkTypeDetection(t *testing.T) {
	chunker := NewMarkdownChunker()

	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com", "external"},
		{"http://example.com", "external"},
		{"mailto:test@example.com", "external"},
		{"#section", "anchor"},
		{"./local.md", "internal"},
		{"../parent.md", "internal"},
		{"file.txt", "internal"},
	}

	for _, test := range tests {
		result := chunker.determineLinkType(test.url)
		if result != test.expected {
			t.Errorf("determineLinkType(%s): expected %s, got %s", test.url, test.expected, result)
		}
	}
}

func TestContentHashConsistency(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := "# Test Heading"

	// 计算同一内容的哈希值多次
	hash1 := chunker.calculateContentHash(content)
	hash2 := chunker.calculateContentHash(content)

	if hash1 != hash2 {
		t.Errorf("Hash should be consistent for same content: %s != %s", hash1, hash2)
	}

	// 不同内容应该有不同的哈希值
	differentContent := "# Different Heading"
	hash3 := chunker.calculateContentHash(differentContent)

	if hash1 == hash3 {
		t.Errorf("Different content should have different hashes")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Test Heading

This is a paragraph.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证原有字段仍然存在且正确
	for _, chunk := range chunks {
		// 验证原有字段
		if chunk.ID < 0 {
			t.Errorf("ID should be non-negative, got %d", chunk.ID)
		}
		if chunk.Type == "" {
			t.Errorf("Type should not be empty")
		}
		if chunk.Content == "" {
			t.Errorf("Content should not be empty")
		}
		if chunk.Text == "" {
			t.Errorf("Text should not be empty")
		}
		if chunk.Metadata == nil {
			t.Errorf("Metadata should not be nil")
		}

		// 验证新字段有默认值
		if chunk.Links == nil {
			t.Errorf("Links should be initialized (even if empty)")
		}
		if chunk.Images == nil {
			t.Errorf("Images should be initialized (even if empty)")
		}
		if chunk.Hash == "" {
			t.Errorf("Hash should not be empty")
		}
	}
}
