package markdownchunker

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewChunkFields_Position(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading 1

This is paragraph 1.

## Heading 2

This is paragraph 2.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证位置信息
	expectedPositions := []struct {
		chunkType string
		startLine int
		endLine   int
	}{
		{"heading", 1, 1},   // # Heading 1
		{"paragraph", 3, 3}, // This is paragraph 1.
		{"heading", 5, 5},   // ## Heading 2
		{"paragraph", 7, 7}, // This is paragraph 2.
	}

	if len(chunks) != len(expectedPositions) {
		t.Fatalf("Expected %d chunks, got %d", len(expectedPositions), len(chunks))
	}

	for i, chunk := range chunks {
		expected := expectedPositions[i]

		if chunk.Type != expected.chunkType {
			t.Errorf("Chunk %d: expected type %s, got %s", i, expected.chunkType, chunk.Type)
		}

		if chunk.Position.StartLine != expected.startLine {
			t.Errorf("Chunk %d: expected start line %d, got %d", i, expected.startLine, chunk.Position.StartLine)
		}

		if chunk.Position.EndLine != expected.endLine {
			t.Errorf("Chunk %d: expected end line %d, got %d", i, expected.endLine, chunk.Position.EndLine)
		}

		// 验证列位置合理
		if chunk.Position.StartCol <= 0 {
			t.Errorf("Chunk %d: start column should be > 0, got %d", i, chunk.Position.StartCol)
		}

		if chunk.Position.EndCol < chunk.Position.StartCol {
			t.Errorf("Chunk %d: end column (%d) should be >= start column (%d)", i, chunk.Position.EndCol, chunk.Position.StartCol)
		}
	}
}

func TestNewChunkFields_Links(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading with [external link](https://example.com)

This paragraph has [internal link](./file.md) and [anchor link](#section).

Visit [GitHub](https://github.com) or email <mailto:test@example.com>.

## Another heading

No links here.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 查找包含链接的块
	var headingChunk, paragraphChunk1, paragraphChunk2 *Chunk
	for i := range chunks {
		chunk := &chunks[i]
		switch {
		case chunk.Type == "heading" && strings.Contains(chunk.Text, "external link"):
			headingChunk = chunk
		case chunk.Type == "paragraph" && strings.Contains(chunk.Text, "internal link"):
			paragraphChunk1 = chunk
		case chunk.Type == "paragraph" && strings.Contains(chunk.Text, "GitHub"):
			paragraphChunk2 = chunk
		}
	}

	// 验证标题中的链接
	if headingChunk == nil {
		t.Fatal("Could not find heading chunk with external link")
	}
	if len(headingChunk.Links) != 1 {
		t.Errorf("Heading should have 1 link, got %d", len(headingChunk.Links))
	} else {
		link := headingChunk.Links[0]
		if link.URL != "https://example.com" {
			t.Errorf("Expected link URL 'https://example.com', got '%s'", link.URL)
		}
		if link.Type != "external" {
			t.Errorf("Expected link type 'external', got '%s'", link.Type)
		}
		if link.Text != "external link" {
			t.Errorf("Expected link text 'external link', got '%s'", link.Text)
		}
	}

	// 验证第一个段落中的链接
	if paragraphChunk1 == nil {
		t.Fatal("Could not find paragraph chunk with internal and anchor links")
	}
	if len(paragraphChunk1.Links) != 2 {
		t.Errorf("First paragraph should have 2 links, got %d", len(paragraphChunk1.Links))
	} else {
		// 内部链接
		internalLink := paragraphChunk1.Links[0]
		if internalLink.Type != "internal" {
			t.Errorf("Expected first link type 'internal', got '%s'", internalLink.Type)
		}
		if internalLink.URL != "./file.md" {
			t.Errorf("Expected first link URL './file.md', got '%s'", internalLink.URL)
		}

		// 锚点链接
		anchorLink := paragraphChunk1.Links[1]
		if anchorLink.Type != "anchor" {
			t.Errorf("Expected second link type 'anchor', got '%s'", anchorLink.Type)
		}
		if anchorLink.URL != "#section" {
			t.Errorf("Expected second link URL '#section', got '%s'", anchorLink.URL)
		}
	}

	// 验证第二个段落中的链接
	if paragraphChunk2 == nil {
		t.Fatal("Could not find paragraph chunk with GitHub and email links")
	}
	if len(paragraphChunk2.Links) != 2 {
		t.Errorf("Second paragraph should have 2 links, got %d", len(paragraphChunk2.Links))
	} else {
		// GitHub链接
		githubLink := paragraphChunk2.Links[0]
		if githubLink.Type != "external" {
			t.Errorf("Expected GitHub link type 'external', got '%s'", githubLink.Type)
		}
		if githubLink.URL != "https://github.com" {
			t.Errorf("Expected GitHub link URL 'https://github.com', got '%s'", githubLink.URL)
		}

		// 邮件链接
		emailLink := paragraphChunk2.Links[1]
		if emailLink.Type != "external" {
			t.Errorf("Expected email link type 'external', got '%s'", emailLink.Type)
		}
		if emailLink.URL != "mailto:test@example.com" {
			t.Errorf("Expected email link URL 'mailto:test@example.com', got '%s'", emailLink.URL)
		}
	}
}

func TestNewChunkFields_Images(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading with ![icon](icon.png "Icon Title")

This paragraph has ![image1](./images/pic1.jpg) and ![image2](https://example.com/pic2.png "Remote Image").

## Another heading

No images here.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 查找包含图片的块
	var headingChunk, paragraphChunk *Chunk
	for i := range chunks {
		chunk := &chunks[i]
		switch {
		case chunk.Type == "heading" && strings.Contains(chunk.Text, "icon"):
			headingChunk = chunk
		case chunk.Type == "paragraph" && strings.Contains(chunk.Text, "image1"):
			paragraphChunk = chunk
		}
	}

	// 验证标题中的图片
	if headingChunk == nil {
		t.Fatal("Could not find heading chunk with icon")
	}
	if len(headingChunk.Images) != 1 {
		t.Errorf("Heading should have 1 image, got %d", len(headingChunk.Images))
	} else {
		image := headingChunk.Images[0]
		if image.URL != "icon.png" {
			t.Errorf("Expected image URL 'icon.png', got '%s'", image.URL)
		}
		if image.Alt != "icon" {
			t.Errorf("Expected image alt 'icon', got '%s'", image.Alt)
		}
		if image.Title != "Icon Title" {
			t.Errorf("Expected image title 'Icon Title', got '%s'", image.Title)
		}
	}

	// 验证段落中的图片
	if paragraphChunk == nil {
		t.Fatal("Could not find paragraph chunk with images")
	}
	if len(paragraphChunk.Images) != 2 {
		t.Errorf("Paragraph should have 2 images, got %d", len(paragraphChunk.Images))
	} else {
		// 第一个图片
		image1 := paragraphChunk.Images[0]
		if image1.URL != "./images/pic1.jpg" {
			t.Errorf("Expected first image URL './images/pic1.jpg', got '%s'", image1.URL)
		}
		if image1.Alt != "image1" {
			t.Errorf("Expected first image alt 'image1', got '%s'", image1.Alt)
		}

		// 第二个图片
		image2 := paragraphChunk.Images[1]
		if image2.URL != "https://example.com/pic2.png" {
			t.Errorf("Expected second image URL 'https://example.com/pic2.png', got '%s'", image2.URL)
		}
		if image2.Alt != "image2" {
			t.Errorf("Expected second image alt 'image2', got '%s'", image2.Alt)
		}
		if image2.Title != "Remote Image" {
			t.Errorf("Expected second image title 'Remote Image', got '%s'", image2.Title)
		}
	}
}

func TestNewChunkFields_Hash(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Same Heading

Different paragraph.

# Same Heading

Same paragraph.

# Same Heading

Different paragraph.`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证哈希值
	var headingHashes []string
	var paragraphHashes []string

	for _, chunk := range chunks {
		if chunk.Hash == "" {
			t.Errorf("Chunk hash should not be empty for type %s", chunk.Type)
		}
		if len(chunk.Hash) != 64 {
			t.Errorf("Chunk hash should be 64 characters long, got %d for type %s", len(chunk.Hash), chunk.Type)
		}

		switch chunk.Type {
		case "heading":
			headingHashes = append(headingHashes, chunk.Hash)
		case "paragraph":
			paragraphHashes = append(paragraphHashes, chunk.Hash)
		}
	}

	// 相同内容的标题应该有相同的哈希值
	if len(headingHashes) >= 2 {
		if headingHashes[0] != headingHashes[1] || headingHashes[1] != headingHashes[2] {
			t.Errorf("Same heading content should have same hash values")
		}
	}

	// 不同内容的段落应该有不同的哈希值
	if len(paragraphHashes) >= 2 {
		if paragraphHashes[0] == paragraphHashes[1] {
			t.Errorf("Different paragraph content should have different hash values")
		}
	}
}

func TestNewChunkFields_JSONSerialization(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Test Heading

This is a paragraph with [link](https://example.com) and ![image](test.jpg "Test Image").`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	for i, chunk := range chunks {
		// 序列化为JSON
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			t.Errorf("Chunk %d: Failed to marshal to JSON: %v", i, err)
			continue
		}

		// 反序列化
		var deserializedChunk Chunk
		err = json.Unmarshal(jsonData, &deserializedChunk)
		if err != nil {
			t.Errorf("Chunk %d: Failed to unmarshal from JSON: %v", i, err)
			continue
		}

		// 验证所有字段都正确序列化和反序列化
		if deserializedChunk.ID != chunk.ID {
			t.Errorf("Chunk %d: ID mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Type != chunk.Type {
			t.Errorf("Chunk %d: Type mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Content != chunk.Content {
			t.Errorf("Chunk %d: Content mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Text != chunk.Text {
			t.Errorf("Chunk %d: Text mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Level != chunk.Level {
			t.Errorf("Chunk %d: Level mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Hash != chunk.Hash {
			t.Errorf("Chunk %d: Hash mismatch after JSON round-trip", i)
		}

		// 验证Position
		if deserializedChunk.Position.StartLine != chunk.Position.StartLine {
			t.Errorf("Chunk %d: Position.StartLine mismatch after JSON round-trip", i)
		}
		if deserializedChunk.Position.EndLine != chunk.Position.EndLine {
			t.Errorf("Chunk %d: Position.EndLine mismatch after JSON round-trip", i)
		}

		// 验证Links
		if len(deserializedChunk.Links) != len(chunk.Links) {
			t.Errorf("Chunk %d: Links length mismatch after JSON round-trip", i)
		}

		// 验证Images
		if len(deserializedChunk.Images) != len(chunk.Images) {
			t.Errorf("Chunk %d: Images length mismatch after JSON round-trip", i)
		}

		// 验证Metadata
		if len(deserializedChunk.Metadata) != len(chunk.Metadata) {
			t.Errorf("Chunk %d: Metadata length mismatch after JSON round-trip", i)
		}
	}
}

func TestNewChunkFields_BackwardCompatibility(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Test Heading

This is a test paragraph.

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```"

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证所有原有字段仍然正常工作
	for i, chunk := range chunks {
		// 原有字段验证
		if chunk.ID < 0 {
			t.Errorf("Chunk %d: ID should be non-negative", i)
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

		// 新字段应该有合理的默认值
		if chunk.Links == nil {
			t.Errorf("Chunk %d: Links should not be nil", i)
		}
		if chunk.Images == nil {
			t.Errorf("Chunk %d: Images should not be nil", i)
		}
		if chunk.Hash == "" {
			t.Errorf("Chunk %d: Hash should not be empty", i)
		}
		if chunk.Position.StartLine <= 0 {
			t.Errorf("Chunk %d: Position.StartLine should be > 0", i)
		}
	}

	// 验证特定类型的元数据仍然存在
	for _, chunk := range chunks {
		switch chunk.Type {
		case "heading":
			if _, exists := chunk.Metadata["heading_level"]; !exists {
				t.Errorf("Heading chunk should have 'heading_level' metadata")
			}
			if _, exists := chunk.Metadata["word_count"]; !exists {
				t.Errorf("Heading chunk should have 'word_count' metadata")
			}
		case "paragraph":
			if _, exists := chunk.Metadata["word_count"]; !exists {
				t.Errorf("Paragraph chunk should have 'word_count' metadata")
			}
			if _, exists := chunk.Metadata["char_count"]; !exists {
				t.Errorf("Paragraph chunk should have 'char_count' metadata")
			}
		case "code":
			if _, exists := chunk.Metadata["language"]; !exists {
				t.Errorf("Code chunk should have 'language' metadata")
			}
			if _, exists := chunk.Metadata["line_count"]; !exists {
				t.Errorf("Code chunk should have 'line_count' metadata")
			}
		}
	}
}

func TestNewChunkFields_AllContentTypes(t *testing.T) {
	chunker := NewMarkdownChunker()

	content := `# Heading with [link](https://example.com)

This is a paragraph with ![image](test.jpg).

- List item 1 with [internal link](./file.md)
- List item 2

` + "```go" + `
func main() {
    fmt.Println("Hello")
}
` + "```" + `

> Blockquote with [external link](https://github.com)

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |

---`

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("Failed to chunk document: %v", err)
	}

	// 验证每种内容类型都有新字段
	contentTypes := make(map[string]bool)
	for _, chunk := range chunks {
		contentTypes[chunk.Type] = true

		// 验证所有新字段都存在
		if chunk.Links == nil {
			t.Errorf("Chunk type %s: Links should not be nil", chunk.Type)
		}
		if chunk.Images == nil {
			t.Errorf("Chunk type %s: Images should not be nil", chunk.Type)
		}
		if chunk.Hash == "" {
			t.Errorf("Chunk type %s: Hash should not be empty", chunk.Type)
		}
		if chunk.Position.StartLine <= 0 {
			t.Errorf("Chunk type %s: Position.StartLine should be > 0", chunk.Type)
		}
	}

	// 验证所有预期的内容类型都存在
	expectedTypes := []string{"heading", "paragraph", "list", "code", "blockquote", "table", "thematic_break"}
	for _, expectedType := range expectedTypes {
		if !contentTypes[expectedType] {
			t.Errorf("Expected content type '%s' not found in chunks", expectedType)
		}
	}
}
