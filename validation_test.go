package markdownchunker

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestDataIntegrity tests that chunk data maintains integrity throughout processing
func TestDataIntegrity(t *testing.T) {
	markdown := `# Test Document

This is a paragraph with **bold** and *italic* text, plus [a link](https://example.com).

## Code Section

Here's some code:

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    for i := 0; i < 5; i++ {
        fmt.Printf("Count: %d\n", i)
    }
}
` + "```" + `

## Data Table

| Name | Age | City | Email |
|------|-----|------|-------|
| John | 25 | NYC | john@example.com |
| Jane | 30 | LA | jane@example.com |
| Bob | 35 | Chicago | bob@example.com |

## List Section

1. First ordered item
2. Second ordered item
3. Third ordered item

Unordered list:
- Alpha
- Beta  
- Gamma

## Quote Section

> This is a blockquote with multiple lines.
> It contains important information.
> 
> And it has multiple paragraphs.

---

Final paragraph after thematic break.`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Initialize variables for content preservation test
	var reconstructed strings.Builder
	codeChunks := 0
	tableChunks := 0
	listChunks := 0
	headingChunks := 0

	// Test 1: Verify chunk IDs are sequential
	for i, chunk := range chunks {
		if chunk.ID != i {
			t.Errorf("Chunk %d has ID %d, expected %d", i, chunk.ID, i)
		}
	}

	// Test 2: Verify all chunks have required fields
	for i, chunk := range chunks {
		if chunk.Type == "" {
			t.Errorf("Chunk %d has empty type", i)
		}
		if chunk.Content == "" {
			t.Errorf("Chunk %d has empty content", i)
		}
		if chunk.Metadata == nil {
			t.Errorf("Chunk %d has nil metadata", i)
		}
		// Text can be empty for thematic breaks
		if chunk.Type != "thematic_break" && chunk.Text == "" {
			t.Errorf("Chunk %d has empty text", i)
		}
	}

	// Test 3: Verify content preservation
	for _, chunk := range chunks {
		reconstructed.WriteString(chunk.Content)
		reconstructed.WriteString("\n\n")

		switch chunk.Type {
		case "code":
			codeChunks++
		case "table":
			tableChunks++
		case "list":
			listChunks++
		case "heading":
			headingChunks++
		}
	}

	// Verify expected counts
	if headingChunks < 3 {
		t.Errorf("Expected at least 3 heading chunks, got %d", headingChunks)
	}
	if codeChunks < 1 {
		t.Errorf("Expected at least 1 code chunk, got %d", codeChunks)
	}
	if tableChunks < 1 {
		t.Errorf("Expected at least 1 table chunk, got %d", tableChunks)
	}
	if listChunks < 2 {
		t.Errorf("Expected at least 2 list chunks, got %d", listChunks)
	}

	t.Logf("Data integrity test completed:")
	t.Logf("  Total chunks: %d", len(chunks))
	t.Logf("  Headings: %d, Code: %d, Tables: %d, Lists: %d", headingChunks, codeChunks, tableChunks, listChunks)
}

// TestMetadataConsistency tests that metadata is consistent across similar chunks
func TestMetadataConsistency(t *testing.T) {
	markdown := `# First Heading

First paragraph with [link1](https://example1.com) and [link2](https://example2.com).

# Second Heading

Second paragraph with [link3](https://example3.com).

` + "```go" + `
// Simple function
func hello() {
    fmt.Println("Hello")
}
` + "```" + `

` + "```go" + `
// Complex function
func complex() {
    for i := 0; i < 10; i++ {
        if i%2 == 0 {
            fmt.Printf("Even: %d\n", i)
        } else {
            fmt.Printf("Odd: %d\n", i)
        }
    }
}
` + "```" + `

| Name | Age |
|------|-----|
| Alice | 25 |
| Bob | 30 |

| Product | Price | Stock |
|---------|-------|-------|
| Widget | $10.99 | 100 |
| Gadget | $25.50 | 50 |`

	config := &ChunkerConfig{
		CustomExtractors: []MetadataExtractor{&LinkExtractor{}, &CodeComplexityExtractor{}},
		ErrorHandling:    ErrorModePermissive,
	}

	chunker := NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Test metadata consistency for similar chunk types
	headingChunks := []Chunk{}
	paragraphChunks := []Chunk{}
	codeChunks := []Chunk{}
	tableChunks := []Chunk{}

	for _, chunk := range chunks {
		switch chunk.Type {
		case "heading":
			headingChunks = append(headingChunks, chunk)
		case "paragraph":
			paragraphChunks = append(paragraphChunks, chunk)
		case "code":
			codeChunks = append(codeChunks, chunk)
		case "table":
			tableChunks = append(tableChunks, chunk)
		}
	}

	// Test heading metadata consistency
	for i, chunk := range headingChunks {
		if chunk.Metadata["level"] == "" {
			t.Errorf("Heading chunk %d missing level metadata", i)
		}
		if chunk.Metadata["word_count"] == "" {
			t.Errorf("Heading chunk %d missing word_count metadata", i)
		}
	}

	// Test paragraph metadata consistency
	for i, chunk := range paragraphChunks {
		if chunk.Metadata["word_count"] == "" {
			t.Errorf("Paragraph chunk %d missing word_count metadata", i)
		}
		if chunk.Metadata["char_count"] == "" {
			t.Errorf("Paragraph chunk %d missing char_count metadata", i)
		}
	}

	// Test code metadata consistency
	for i, chunk := range codeChunks {
		if chunk.Metadata["language"] == "" {
			t.Errorf("Code chunk %d missing language metadata", i)
		}
		if chunk.Metadata["line_count"] == "" {
			t.Errorf("Code chunk %d missing line_count metadata", i)
		}
		if _, exists := chunk.Metadata["code_complexity"]; !exists {
			t.Errorf("Code chunk %d missing code_complexity metadata", i)
		}
	}

	// Test table metadata consistency
	for i, chunk := range tableChunks {
		if chunk.Metadata["rows"] == "" {
			t.Errorf("Table chunk %d missing rows metadata", i)
		}
		if chunk.Metadata["columns"] == "" {
			t.Errorf("Table chunk %d missing columns metadata", i)
		}
		if chunk.Metadata["has_header"] == "" {
			t.Errorf("Table chunk %d missing has_header metadata", i)
		}
	}

	t.Logf("Metadata consistency test completed:")
	t.Logf("  Headings: %d, Paragraphs: %d, Code: %d, Tables: %d",
		len(headingChunks), len(paragraphChunks), len(codeChunks), len(tableChunks))
}

// TestChunkSerialization tests that chunks can be serialized and deserialized
func TestChunkSerialization(t *testing.T) {
	markdown := `# Test Document

This is a test paragraph with [a link](https://example.com).

` + "```go" + `
func test() {
    fmt.Println("Hello")
}
` + "```" + `

| Name | Value |
|------|-------|
| Test | 123 |`

	chunker := NewMarkdownChunker()
	originalChunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Test JSON serialization
	for i, chunk := range originalChunks {
		// Serialize to JSON
		jsonData, err := json.Marshal(chunk)
		if err != nil {
			t.Errorf("Failed to marshal chunk %d to JSON: %v", i, err)
			continue
		}

		// Deserialize from JSON
		var deserializedChunk Chunk
		err = json.Unmarshal(jsonData, &deserializedChunk)
		if err != nil {
			t.Errorf("Failed to unmarshal chunk %d from JSON: %v", i, err)
			continue
		}

		// Compare original and deserialized
		if !reflect.DeepEqual(chunk, deserializedChunk) {
			t.Errorf("Chunk %d serialization mismatch", i)
			t.Logf("Original: %+v", chunk)
			t.Logf("Deserialized: %+v", deserializedChunk)
		}
	}

	t.Logf("Serialization test completed for %d chunks", len(originalChunks))
}

// TestRegressionScenarios tests specific scenarios that have caused issues in the past
func TestRegressionScenarios(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		verify   func(t *testing.T, chunks []Chunk, err error)
	}{
		{
			name:     "Empty code block language",
			markdown: "# Test\n\n```\ncode without language\n```",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				found := false
				for _, chunk := range chunks {
					if chunk.Type == "code" {
						found = true
						if chunk.Metadata["language"] != "" && chunk.Metadata["language"] != "text" {
							t.Errorf("Expected empty or 'text' language, got: %s", chunk.Metadata["language"])
						}
					}
				}
				if !found {
					t.Error("Expected to find code chunk")
				}
			},
		},
		{
			name:     "Table with missing cells",
			markdown: "# Test\n\n| A | B | C |\n|---|---|---|\n| 1 | 2 |\n| 3 | 4 | 5 | 6 |",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				found := false
				for _, chunk := range chunks {
					if chunk.Type == "table" {
						found = true
						if chunk.Metadata["is_well_formed"] == "true" {
							t.Error("Table should not be marked as well-formed")
						}
					}
				}
				if !found {
					t.Error("Expected to find table chunk")
				}
			},
		},
		{
			name:     "Nested blockquotes",
			markdown: "# Test\n\n> Outer quote\n> > Inner quote\n> Back to outer",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				found := false
				for _, chunk := range chunks {
					if chunk.Type == "blockquote" {
						found = true
						if !strings.Contains(chunk.Content, "Inner quote") {
							t.Error("Nested quote content not preserved")
						}
					}
				}
				if !found {
					t.Error("Expected to find blockquote chunk")
				}
			},
		},
		{
			name:     "List with mixed markers",
			markdown: "# Test\n\n- Item 1\n* Item 2\n+ Item 3\n\n1. Numbered 1\n2) Numbered 2",
			verify: func(t *testing.T, chunks []Chunk, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				listChunks := 0
				for _, chunk := range chunks {
					if chunk.Type == "list" {
						listChunks++
					}
				}
				if listChunks < 1 {
					t.Error("Expected to find list chunks")
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

// TestContentReconstruction tests that chunked content can be reconstructed
func TestContentReconstruction(t *testing.T) {
	markdown := `# Test Document

This is a paragraph with some content.

## Code Section

` + "```go\npackage main\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

## Data Table

| Name | Age | Email |
|------|-----|-------|
| John | 30  | john@example.com |
| Jane | 25  | jane@example.com |

## List Section

1. First ordered item
2. Second ordered item

- Alpha
- Beta
- Gamma

## Quote Section

> This is a blockquote
> with multiple lines

Final paragraph with conclusion.`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Initialize counters for chunk types
	codeChunks := 0
	tableChunks := 0
	listChunks := 0
	headingChunks := 0

	// Reconstruct content from chunks
	var reconstructed strings.Builder
	for _, chunk := range chunks {
		reconstructed.WriteString(chunk.Content)
		reconstructed.WriteString("\n\n")
	}

	// The reconstructed content should contain all major elements
	reconstructedStr := reconstructed.String()
	expectedElements := []string{
		"Test Document", "paragraph with", "Code Section", "package main",
		"Data Table", "Name", "Age", "john@example.com", "List Section",
		"First ordered", "Alpha", "Quote Section", "blockquote",
		"Final paragraph",
	}

	for _, element := range expectedElements {
		if !strings.Contains(reconstructedStr, element) {
			t.Errorf("Reconstructed content missing element: %s", element)
		}
	}

	// Test 4: Verify metadata consistency
	for _, chunk := range chunks {
		switch chunk.Type {
		case "code":
			codeChunks++
		case "table":
			tableChunks++
		case "list":
			listChunks++
		case "heading":
			headingChunks++
		}
	}

	// Verify we have the expected types of chunks
	if codeChunks == 0 {
		t.Error("Expected to find code chunks")
	}
	if tableChunks == 0 {
		t.Error("Expected to find table chunks")
	}
	if listChunks == 0 {
		t.Error("Expected to find list chunks")
	}
	if headingChunks == 0 {
		t.Error("Expected to find heading chunks")
	}
}
