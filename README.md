# Markdown Chunker

A Go library for intelligently splitting Markdown documents into semantic chunks. This library parses Markdown content and breaks it down into meaningful segments like headings, paragraphs, code blocks, tables, lists, and more.

## Features

- **Semantic Chunking**: Splits Markdown documents based on content structure rather than arbitrary text length
- **Multiple Content Types**: Supports headings, paragraphs, code blocks, tables, lists, blockquotes, and thematic breaks
- **Rich Metadata**: Each chunk includes metadata like heading levels, word counts, code language, table dimensions, etc.
- **GitHub Flavored Markdown**: Full support for GFM features including tables
- **Pure Text Extraction**: Provides both original Markdown content and clean text for each chunk
- **Easy Integration**: Simple API for processing Markdown documents

## Installation

```bash
go get github.com/kydenul/markdown-chunker
```

## Quick Start

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    markdown := `# My Document

This is a paragraph with some content.

## Code Example

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |
`

    chunker := mc.NewMarkdownChunker()
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }

    for _, chunk := range chunks {
        fmt.Printf("Type: %s, Content: %s\n", chunk.Type, chunk.Text)
    }
}
```

## Supported Content Types

### Headings

- **Type**: `heading`
- **Metadata**: `heading_level` (1-6)
- **Level**: Heading level (1-6)

### Paragraphs

- **Type**: `paragraph`
- **Metadata**: `word_count`
- **Level**: 0

### Code Blocks

- **Type**: `code`
- **Metadata**: `language`, `line_count`
- **Level**: 0

### Tables

- **Type**: `table`
- **Metadata**: `rows`, `columns`
- **Level**: 0

### Lists

- **Type**: `list`
- **Metadata**: `list_type` (ordered/unordered), `item_count`
- **Level**: 0

### Blockquotes

- **Type**: `blockquote`
- **Metadata**: `word_count`
- **Level**: 0

### Thematic Breaks

- **Type**: `thematic_break`
- **Metadata**: `type` (horizontal_rule)
- **Level**: 0

## API Reference

### Types

#### Chunk

```go
type Chunk struct {
    ID       int               `json:"id"`       // Unique chunk identifier
    Type     string            `json:"type"`     // Content type (heading, paragraph, etc.)
    Content  string            `json:"content"`  // Original Markdown content
    Text     string            `json:"text"`     // Plain text content
    Level    int               `json:"level"`    // Heading level (0 for non-headings)
    Metadata map[string]string `json:"metadata"` // Additional metadata
}
```

#### MarkdownChunker

```go
type MarkdownChunker struct {
    // Internal fields
}
```

### Functions

#### NewMarkdownChunker

```go
func NewMarkdownChunker() *MarkdownChunker
```

Creates a new Markdown chunker instance with GitHub Flavored Markdown support.

#### ChunkDocument

```go
func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error)
```

Processes a Markdown document and returns an array of semantic chunks.

**Parameters:**

- `content`: Raw Markdown content as bytes

**Returns:**

- `[]Chunk`: Array of processed chunks
- `error`: Error if parsing fails

## Examples

### Processing a Complex Document

```go
package main

import (
    "encoding/json"
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    markdown := `# Database Design Document

This document describes the database schema.

## User Table

The users table stores user information:

| Field | Type | Description |
|-------|------|-------------|
| id | int | Primary key |
| name | varchar(100) | Username |
| email | varchar(255) | Email address |

### SQL Example

` + "```sql" + `
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE
);
` + "```" + `

## Important Notes

> **Warning**: Always backup your data before making schema changes.

Key considerations:
- Ensure email uniqueness
- Regular backups
- Monitor performance
`

    chunker := mc.NewMarkdownChunker()
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }

    // Convert to JSON for easy inspection
    jsonData, _ := json.MarshalIndent(chunks, "", "  ")
    fmt.Println(string(jsonData))
}
```

### Filtering Chunks by Type

```go
func filterChunksByType(chunks []mc.Chunk, chunkType string) []mc.Chunk {
    var filtered []mc.Chunk
    for _, chunk := range chunks {
        if chunk.Type == chunkType {
            filtered = append(filtered, chunk)
        }
    }
    return filtered
}

// Usage
headings := filterChunksByType(chunks, "heading")
codeBlocks := filterChunksByType(chunks, "code")
```

### Extracting Table Data

```go
func analyzeTable(chunk mc.Chunk) {
    if chunk.Type != "table" {
        return
    }
    
    rows := chunk.Metadata["rows"]
    columns := chunk.Metadata["columns"]
    
    fmt.Printf("Table: %s rows × %s columns\n", rows, columns)
    fmt.Printf("Content:\n%s\n", chunk.Content)
}
```

## Use Cases

- **Documentation Processing**: Break down large documentation into searchable chunks
- **Content Analysis**: Analyze document structure and content distribution
- **RAG Systems**: Prepare Markdown content for vector databases and retrieval systems
- **Content Migration**: Convert Markdown documents to structured data
- **Static Site Generation**: Process Markdown files for static site generators

## Dependencies

- [goldmark](https://github.com/yuin/goldmark): CommonMark compliant Markdown parser

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.0.0

- Initial release
- Support for all major Markdown elements
- GitHub Flavored Markdown support
- Rich metadata extraction
