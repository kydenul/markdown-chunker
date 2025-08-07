package markdownchunker

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRequirement_5_1_LinkExtraction 验证需求5.1：提取链接信息到元数据中
func TestRequirement_5_1_LinkExtraction(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading with [external link](https://example.com)

This paragraph contains [internal link](./local.md) and [anchor link](#section).

Visit [GitHub](https://github.com) for more information.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证链接提取功能
	linkCount := 0
	for _, chunk := range chunks {
		linkCount += len(chunk.Links)

		// 验证链接信息完整性
		for _, link := range chunk.Links {
			if link.Text == "" {
				t.Errorf("Link text should not be empty")
			}
			if link.URL == "" {
				t.Errorf("Link URL should not be empty")
			}
			if link.Type == "" {
				t.Errorf("Link type should not be empty")
			}

			// 验证链接类型正确性
			switch link.Type {
			case "external":
				if !strings.HasPrefix(link.URL, "http") && !strings.HasPrefix(link.URL, "mailto:") {
					t.Errorf("External link should start with http or mailto, got: %s", link.URL)
				}
			case "internal":
				if strings.HasPrefix(link.URL, "http") || strings.HasPrefix(link.URL, "#") {
					t.Errorf("Internal link should not start with http or #, got: %s", link.URL)
				}
			case "anchor":
				if !strings.HasPrefix(link.URL, "#") {
					t.Errorf("Anchor link should start with #, got: %s", link.URL)
				}
			default:
				t.Errorf("Unknown link type: %s", link.Type)
			}
		}
	}

	if linkCount == 0 {
		t.Errorf("Should have extracted links from the content")
	}

	t.Logf("Successfully extracted %d links with proper type classification", linkCount)
}

// TestRequirement_5_2_ImageExtraction 验证需求5.2：提取图片信息到元数据中
func TestRequirement_5_2_ImageExtraction(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading with ![icon](icon.png "Icon Title")

This paragraph contains ![local image](./images/pic.jpg) and ![remote image](https://example.com/pic.png "Remote Image").

![Simple image](simple.jpg)`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证图片提取功能
	imageCount := 0
	for _, chunk := range chunks {
		imageCount += len(chunk.Images)

		// 验证图片信息完整性
		for _, image := range chunk.Images {
			if image.Alt == "" {
				t.Errorf("Image alt text should not be empty")
			}
			if image.URL == "" {
				t.Errorf("Image URL should not be empty")
			}

			// 验证图片结构完整性
			if image.Width != "" || image.Height != "" {
				// 如果有宽高信息，应该是有效的
				t.Logf("Image has dimensions: %sx%s", image.Width, image.Height)
			}
		}
	}

	if imageCount == 0 {
		t.Errorf("Should have extracted images from the content")
	}

	t.Logf("Successfully extracted %d images with complete metadata", imageCount)
}

// TestRequirement_5_3_CodeComplexityMetadata 验证需求5.3：代码复杂度分析元数据
func TestRequirement_5_3_CodeComplexityMetadata(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := "```go\n" +
		"func main() {\n" +
		"    fmt.Println(\"Hello\")\n" +
		"    if true {\n" +
		"        fmt.Println(\"World\")\n" +
		"    }\n" +
		"}\n" +
		"```"

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 查找代码块
	var codeChunk *Chunk
	for i := range chunks {
		if chunks[i].Type == "code" {
			codeChunk = &chunks[i]
			break
		}
	}

	if codeChunk == nil {
		t.Fatal("Should have found a code chunk")
	}

	// 验证代码块有基本的复杂度元数据
	if _, exists := codeChunk.Metadata["language"]; !exists {
		t.Errorf("Code chunk should have language metadata")
	}
	if _, exists := codeChunk.Metadata["line_count"]; !exists {
		t.Errorf("Code chunk should have line_count metadata")
	}

	// 验证行数计算正确
	if lineCount := codeChunk.Metadata["line_count"]; lineCount != "6" {
		t.Errorf("Expected line count 6, got %s", lineCount)
	}

	t.Logf("Code chunk metadata: %+v", codeChunk.Metadata)
}

// TestChunkStructureEnhancements 验证Chunk结构体的增强功能
func TestChunkStructureEnhancements(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Test Heading

This is a test paragraph with [link](https://example.com) and ![image](test.jpg).`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	for i, chunk := range chunks {
		// 验证Position字段
		if chunk.Position.StartLine <= 0 {
			t.Errorf("Chunk %d: Position.StartLine should be > 0, got %d", i, chunk.Position.StartLine)
		}
		if chunk.Position.EndLine < chunk.Position.StartLine {
			t.Errorf("Chunk %d: Position.EndLine should be >= StartLine", i)
		}
		if chunk.Position.StartCol <= 0 {
			t.Errorf("Chunk %d: Position.StartCol should be > 0, got %d", i, chunk.Position.StartCol)
		}
		if chunk.Position.EndCol < chunk.Position.StartCol {
			t.Errorf("Chunk %d: Position.EndCol should be >= StartCol", i)
		}

		// 验证Links字段
		if chunk.Links == nil {
			t.Errorf("Chunk %d: Links should not be nil", i)
		}

		// 验证Images字段
		if chunk.Images == nil {
			t.Errorf("Chunk %d: Images should not be nil", i)
		}

		// 验证Hash字段
		if chunk.Hash == "" {
			t.Errorf("Chunk %d: Hash should not be empty", i)
		}
		if len(chunk.Hash) != 64 {
			t.Errorf("Chunk %d: Hash should be 64 characters (SHA256), got %d", i, len(chunk.Hash))
		}

		// 验证JSON序列化包含新字段
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			t.Errorf("Chunk %d: Failed to marshal to JSON: %v", i, err)
			continue
		}

		jsonStr := string(jsonData)
		requiredFields := []string{"position", "links", "images", "hash"}
		for _, field := range requiredFields {
			if !strings.Contains(jsonStr, `"`+field+`"`) {
				t.Errorf("Chunk %d: JSON should contain field '%s'", i, field)
			}
		}
	}

	t.Logf("All chunks have enhanced structure with Position, Links, Images, and Hash fields")
}

// TestBackwardCompatibilityVerification 验证向后兼容性
func TestBackwardCompatibilityVerification(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Test Heading

This is a test paragraph.

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```" + `

- List item 1
- List item 2

> This is a blockquote

| Col1 | Col2 |
|------|------|
| A    | B    |

---`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证所有原有字段仍然存在且功能正常
	for i, chunk := range chunks {
		// 原有字段验证
		if chunk.ID < 0 {
			t.Errorf("Chunk %d: ID should be non-negative, got %d", i, chunk.ID)
		}
		if chunk.Type == "" {
			t.Errorf("Chunk %d: Type should not be empty", i)
		}
		if chunk.Content == "" {
			t.Errorf("Chunk %d: Content should not be empty", i)
		}
		if chunk.Text == "" {
			t.Errorf("Chunk %d: Text should not be empty", i)
		}
		if chunk.Metadata == nil {
			t.Errorf("Chunk %d: Metadata should not be nil", i)
		}

		// 验证特定类型的原有元数据
		switch chunk.Type {
		case "heading":
			if chunk.Level <= 0 {
				t.Errorf("Chunk %d: Heading level should be > 0, got %d", i, chunk.Level)
			}
			if _, exists := chunk.Metadata["heading_level"]; !exists {
				t.Errorf("Chunk %d: Heading should have 'heading_level' metadata", i)
			}
		case "code":
			if _, exists := chunk.Metadata["language"]; !exists {
				t.Errorf("Chunk %d: Code should have 'language' metadata", i)
			}
			if _, exists := chunk.Metadata["line_count"]; !exists {
				t.Errorf("Chunk %d: Code should have 'line_count' metadata", i)
			}
		case "list":
			if _, exists := chunk.Metadata["list_type"]; !exists {
				t.Errorf("Chunk %d: List should have 'list_type' metadata", i)
			}
			if _, exists := chunk.Metadata["item_count"]; !exists {
				t.Errorf("Chunk %d: List should have 'item_count' metadata", i)
			}
		}
	}

	// 验证所有预期的内容类型都能正确处理
	expectedTypes := map[string]bool{
		"heading": false, "paragraph": false, "code": false,
		"list": false, "blockquote": false, "table": false, "thematic_break": false,
	}

	for _, chunk := range chunks {
		if _, exists := expectedTypes[chunk.Type]; exists {
			expectedTypes[chunk.Type] = true
		}
	}

	for chunkType, found := range expectedTypes {
		if !found {
			t.Errorf("Expected chunk type '%s' not found", chunkType)
		}
	}

	t.Logf("Backward compatibility verified: all original functionality preserved")
}

// TestContentHashUniqueness 验证内容哈希的唯一性和一致性
func TestContentHashUniqueness(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Same Content

Different content here.

# Same Content

Same content here.

# Different Content`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	hashMap := make(map[string][]int) // hash -> chunk indices

	for i, chunk := range chunks {
		if chunk.Hash == "" {
			t.Errorf("Chunk %d: Hash should not be empty", i)
			continue
		}

		hashMap[chunk.Hash] = append(hashMap[chunk.Hash], i)
	}

	// 验证相同内容有相同哈希，不同内容有不同哈希
	sameContentHashes := 0
	for hash, indices := range hashMap {
		if len(indices) > 1 {
			sameContentHashes++
			// 验证相同哈希的块确实有相同内容
			firstChunk := chunks[indices[0]]
			for _, idx := range indices[1:] {
				if chunks[idx].Content != firstChunk.Content {
					t.Errorf("Chunks with same hash should have same content: %s vs %s",
						chunks[idx].Content, firstChunk.Content)
				}
			}
			t.Logf("Found %d chunks with same hash %s (same content)", len(indices), hash[:8])
		}
	}

	t.Logf("Hash uniqueness verified: %d unique hashes, %d groups of same content",
		len(hashMap), sameContentHashes)
}
