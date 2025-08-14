# Markdown Chunker

A Go library for intelligently splitting Markdown documents into semantic chunks. This library parses Markdown content and breaks it down into meaningful segments like headings, paragraphs, code blocks, tables, lists, and more.

## Features

- **Flexible Chunking Strategies**: Multiple built-in strategies (element-level, hierarchical, document-level) with support for custom strategies
- **Semantic Chunking**: Splits Markdown documents based on content structure rather than arbitrary text length
- **Multiple Content Types**: Supports headings, paragraphs, code blocks, tables, lists, blockquotes, and thematic breaks
- **Rich Metadata**: Each chunk includes metadata like heading levels, word counts, code language, table dimensions, etc.
- **GitHub Flavored Markdown**: Full support for GFM features including tables
- **Pure Text Extraction**: Provides both original Markdown content and clean text for each chunk
- **Configurable Processing**: Flexible configuration system for customizing chunking behavior
- **Advanced Error Handling**: Comprehensive error handling with multiple modes (strict, permissive, silent)
- **Performance Monitoring**: Built-in performance monitoring and memory optimization
- **Enhanced Metadata Extraction**: Extensible metadata extraction system with link, image, and code analysis
- **Position Tracking**: Precise position information for each chunk in the original document
- **Content Deduplication**: SHA256 hash-based content deduplication
- **Memory Optimization**: Object pooling and memory-efficient processing for large documents
- **Comprehensive Logging**: Configurable logging system with multiple levels and formats
- **Easy Integration**: Simple API for processing Markdown documents

## Installation

```bash
go get github.com/kydenul/markdown-chunker
```

## Quick Start

### Basic Usage

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

### Advanced Usage with Configuration

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Create custom configuration
    config := mc.DefaultConfig()
    config.MaxChunkSize = 1000
    config.ErrorHandling = mc.ErrorModePermissive
    config.EnabledTypes = map[string]bool{
        "heading":    true,
        "paragraph":  true,
        "code":       true,
        "table":      true,
        "list":       false, // Disable list processing
    }
    
    // Configure logging
    config.EnableLog = true
    config.LogLevel = "INFO"
    config.LogFormat = "console"
    config.LogDirectory = "./logs"
    
    // Add custom metadata extractors
    config.CustomExtractors = []mc.MetadataExtractor{
        &mc.LinkExtractor{},
        &mc.ImageExtractor{},
        &mc.CodeComplexityExtractor{},
    }

    chunker := mc.NewMarkdownChunkerWithConfig(config)
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }

    // Process chunks with enhanced metadata
    for _, chunk := range chunks {
        fmt.Printf("Type: %s, Position: %d:%d-%d:%d\n", 
            chunk.Type, 
            chunk.Position.StartLine, chunk.Position.StartCol,
            chunk.Position.EndLine, chunk.Position.EndCol)
        
        // Display links and images
        if len(chunk.Links) > 0 {
            fmt.Printf("  Links: %d\n", len(chunk.Links))
        }
        if len(chunk.Images) > 0 {
            fmt.Printf("  Images: %d\n", len(chunk.Images))
        }
        
        // Display hash for deduplication
        fmt.Printf("  Hash: %s\n", chunk.Hash[:8])
    }
    
    // Check for errors
    if chunker.HasErrors() {
        fmt.Printf("Processing errors: %d\n", len(chunker.GetErrors()))
    }
    
    // Get performance statistics
    stats := chunker.GetPerformanceStats()
    fmt.Printf("Processing time: %v\n", stats.ProcessingTime)
    fmt.Printf("Memory used: %d bytes\n", stats.MemoryUsed)
}
```

## Chunking Strategies

The library supports multiple chunking strategies to handle different use cases and document structures. You can choose from built-in strategies or create custom ones.

### Built-in Strategies

#### Element-Level Strategy (Default)

The element-level strategy processes each Markdown element individually, maintaining the original behavior of the library.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)
```

**Use Cases:**
- Fine-grained content analysis
- Search indexing with precise matching
- Content that doesn't have clear hierarchical structure

#### Hierarchical Strategy

The hierarchical strategy groups content by heading levels, creating chunks that contain a heading and all its subordinate content.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3) // Max depth of 3 levels

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)
```

**Configuration Options:**
- `MaxDepth`: Maximum heading level to consider (1-6)
- `MinDepth`: Minimum heading level to start chunking
- `MergeEmpty`: Whether to merge empty sections with parent
- `MinChunkSize`: Minimum size for hierarchical chunks
- `MaxChunkSize`: Maximum size before splitting hierarchical chunks

**Use Cases:**
- Documentation with clear section structure
- Books and articles with hierarchical organization
- Content where context within sections is important

#### Document-Level Strategy

The document-level strategy treats the entire document as a single chunk.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.DocumentLevelConfig()

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)
```

**Use Cases:**
- Small documents that should be processed as a whole
- Document classification tasks
- When you need complete document context

### Strategy Configuration Examples

#### Basic Strategy Usage

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    markdown := `# User Guide

Welcome to our application.

## Getting Started

Follow these steps to begin.

### Installation

Run the following command:

` + "```bash" + `
npm install our-app
` + "```" + `

### Configuration

Edit your config file:

` + "```json" + `
{
  "theme": "dark",
  "language": "en"
}
` + "```" + `

## Advanced Features

Learn about advanced functionality.

### Custom Themes

Create your own themes.

### Plugins

Extend functionality with plugins.`

    // Element-level strategy (default)
    fmt.Println("=== Element-Level Strategy ===")
    config1 := mc.DefaultConfig()
    config1.ChunkingStrategy = mc.ElementLevelConfig()
    
    chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
    chunks1, _ := chunker1.ChunkDocument([]byte(markdown))
    
    fmt.Printf("Chunks: %d\n", len(chunks1))
    for i, chunk := range chunks1 {
        fmt.Printf("  %d. %s: %s\n", i+1, chunk.Type, 
            truncateText(chunk.Text, 50))
    }

    // Hierarchical strategy
    fmt.Println("\n=== Hierarchical Strategy (Max Depth 2) ===")
    config2 := mc.DefaultConfig()
    config2.ChunkingStrategy = mc.HierarchicalConfig(2)
    
    chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
    chunks2, _ := chunker2.ChunkDocument([]byte(markdown))
    
    fmt.Printf("Chunks: %d\n", len(chunks2))
    for i, chunk := range chunks2 {
        fmt.Printf("  %d. %s (Level %d): %s\n", i+1, chunk.Type, 
            chunk.Level, truncateText(chunk.Text, 50))
    }

    // Document-level strategy
    fmt.Println("\n=== Document-Level Strategy ===")
    config3 := mc.DefaultConfig()
    config3.ChunkingStrategy = mc.DocumentLevelConfig()
    
    chunker3 := mc.NewMarkdownChunkerWithConfig(config3)
    chunks3, _ := chunker3.ChunkDocument([]byte(markdown))
    
    fmt.Printf("Chunks: %d\n", len(chunks3))
    fmt.Printf("  1. %s: %d characters\n", chunks3[0].Type, len(chunks3[0].Content))
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen] + "..."
}
```

#### Dynamic Strategy Switching

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    chunker := mc.NewMarkdownChunker()
    content := []byte(`# Document
    
## Section 1
Content for section 1.

## Section 2  
Content for section 2.`)

    // Start with element-level strategy
    chunks1, _ := chunker.ChunkDocument(content)
    fmt.Printf("Element-level: %d chunks\n", len(chunks1))

    // Switch to hierarchical strategy
    err := chunker.SetStrategy("hierarchical", mc.HierarchicalConfig(2))
    if err != nil {
        panic(err)
    }
    
    chunks2, _ := chunker.ChunkDocument(content)
    fmt.Printf("Hierarchical: %d chunks\n", len(chunks2))

    // Switch to document-level strategy
    err = chunker.SetStrategy("document-level", mc.DocumentLevelConfig())
    if err != nil {
        panic(err)
    }
    
    chunks3, _ := chunker.ChunkDocument(content)
    fmt.Printf("Document-level: %d chunks\n", len(chunks3))
}
```

### Custom Strategies

You can create custom chunking strategies by implementing the `ChunkingStrategy` interface or using the strategy builder.

#### Using the Strategy Builder

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Create a custom strategy that:
    // 1. Creates separate chunks for H1 and H2 headings
    // 2. Merges paragraphs and lists with their parent heading
    // 3. Creates separate chunks for code blocks and tables
    
    builder := mc.NewCustomStrategyBuilder("content-focused", 
        "Groups content by importance, separating code and tables")
    
    // High priority: Separate chunks for major headings
    builder.AddRule(
        mc.HeadingLevelCondition{MinLevel: 1, MaxLevel: 2},
        mc.CreateSeparateChunkAction{},
        10,
    )
    
    // High priority: Separate chunks for code and tables
    builder.AddRule(
        mc.ContentTypeCondition{Types: []string{"code", "table"}},
        mc.CreateSeparateChunkAction{},
        9,
    )
    
    // Medium priority: Merge text content with parent
    builder.AddRule(
        mc.ContentTypeCondition{Types: []string{"paragraph", "list", "blockquote"}},
        mc.MergeWithParentAction{},
        5,
    )
    
    // Low priority: Skip minor headings (merge with parent)
    builder.AddRule(
        mc.HeadingLevelCondition{MinLevel: 3, MaxLevel: 6},
        mc.MergeWithParentAction{},
        3,
    )
    
    customStrategy := builder.Build()
    
    // Register and use the custom strategy
    chunker := mc.NewMarkdownChunker()
    err := chunker.RegisterStrategy(customStrategy)
    if err != nil {
        panic(err)
    }
    
    err = chunker.SetStrategy("content-focused", nil)
    if err != nil {
        panic(err)
    }
    
    markdown := `# Main Title

Introduction paragraph.

## Important Section

Some content here.

### Details

More details.

` + "```go" + `
func example() {
    fmt.Println("code")
}
` + "```" + `

| Feature | Status |
|---------|--------|
| Custom  | Active |`

    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Custom strategy created %d chunks:\n", len(chunks))
    for i, chunk := range chunks {
        fmt.Printf("  %d. %s: %s\n", i+1, chunk.Type, 
            truncateText(chunk.Text, 60))
    }
}
```

#### Implementing a Custom Strategy Interface

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
    "github.com/yuin/goldmark/ast"
)

// CodeFocusedStrategy creates separate chunks for all code blocks
// and merges everything else by heading level
type CodeFocusedStrategy struct {
    config *mc.StrategyConfig
}

func (s *CodeFocusedStrategy) GetName() string {
    return "code-focused"
}

func (s *CodeFocusedStrategy) GetDescription() string {
    return "Separates all code blocks and groups other content hierarchically"
}

func (s *CodeFocusedStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *mc.MarkdownChunker) ([]mc.Chunk, error) {
    var chunks []mc.Chunk
    chunkID := 0
    
    // First pass: extract all code blocks as separate chunks
    ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
        if !entering {
            return ast.WalkContinue, nil
        }
        
        if codeBlock, ok := node.(*ast.FencedCodeBlock); ok {
            if chunk := chunker.ProcessCodeBlock(codeBlock, chunkID); chunk != nil {
                chunks = append(chunks, *chunk)
                chunkID++
            }
        }
        
        return ast.WalkContinue, nil
    })
    
    // Second pass: process remaining content hierarchically
    hierarchicalStrategy := &mc.HierarchicalStrategy{}
    hierarchicalChunks, err := hierarchicalStrategy.ChunkDocument(doc, source, chunker)
    if err != nil {
        return nil, err
    }
    
    // Filter out code blocks from hierarchical chunks and add them
    for _, chunk := range hierarchicalChunks {
        if chunk.Type != "code" {
            chunk.ID = chunkID
            chunks = append(chunks, chunk)
            chunkID++
        }
    }
    
    return chunks, nil
}

func (s *CodeFocusedStrategy) ValidateConfig(config *mc.StrategyConfig) error {
    // No specific validation needed for this strategy
    return nil
}

func (s *CodeFocusedStrategy) Clone() mc.ChunkingStrategy {
    return &CodeFocusedStrategy{
        config: s.config,
    }
}

func main() {
    chunker := mc.NewMarkdownChunker()
    
    // Register custom strategy
    customStrategy := &CodeFocusedStrategy{}
    err := chunker.RegisterStrategy(customStrategy)
    if err != nil {
        panic(err)
    }
    
    // Use custom strategy
    err = chunker.SetStrategy("code-focused", nil)
    if err != nil {
        panic(err)
    }
    
    markdown := `# API Documentation

This document describes our API.

## Authentication

Use API keys for authentication.

` + "```bash" + `
curl -H "Authorization: Bearer TOKEN" https://api.example.com
` + "```" + `

## Endpoints

### Users

Get user information:

` + "```javascript" + `
fetch('/api/users/123')
  .then(response => response.json())
  .then(data => console.log(data));
` + "```" + `

### Posts

Create a new post:

` + "```python" + `
import requests

response = requests.post('/api/posts', json={
    'title': 'My Post',
    'content': 'Post content here'
})
` + "```"

    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Code-focused strategy created %d chunks:\n", len(chunks))
    for i, chunk := range chunks {
        fmt.Printf("  %d. %s", i+1, chunk.Type)
        if chunk.Type == "code" {
            if lang, exists := chunk.Metadata["language"]; exists {
                fmt.Printf(" (%s)", lang)
            }
        }
        fmt.Printf(": %s\n", truncateText(chunk.Text, 50))
    }
}
```

### Strategy Best Practices

#### Choosing the Right Strategy

1. **Element-Level Strategy**
   - Use for: Search indexing, fine-grained analysis, content without clear structure
   - Pros: Maximum granularity, consistent chunk sizes
   - Cons: May break logical content groupings

2. **Hierarchical Strategy**
   - Use for: Documentation, books, structured content
   - Pros: Preserves logical structure, maintains context
   - Cons: Variable chunk sizes, may create very large chunks

3. **Document-Level Strategy**
   - Use for: Small documents, document classification, full-context analysis
   - Pros: Complete context, simple processing
   - Cons: Not suitable for large documents, limited granularity

4. **Custom Strategies**
   - Use for: Specific business requirements, specialized content types
   - Pros: Tailored to exact needs, maximum flexibility
   - Cons: Requires development effort, needs testing

#### Performance Considerations

```go
// For high-performance scenarios, configure strategy caching
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3)
config.PerformanceMode = mc.PerformanceModeSpeedOptimized
config.EnableObjectPooling = true

// For memory-constrained environments
config.PerformanceMode = mc.PerformanceModeMemoryOptimized
config.MemoryLimit = 100 * 1024 * 1024 // 100MB limit
```

#### Error Handling with Strategies

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3)
config.ErrorHandling = mc.ErrorModePermissive // Continue on strategy errors

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

if err != nil {
    // Handle strategy-specific errors
    if chunkerErr, ok := err.(*mc.ChunkerError); ok {
        if chunkerErr.Type == mc.ErrorTypeStrategyExecutionFailed {
            fmt.Printf("Strategy error: %s\n", chunkerErr.Message)
            // Fallback to element-level strategy
            chunker.SetStrategy("element-level", mc.ElementLevelConfig())
            chunks, err = chunker.ChunkDocument(content)
        }
    }
}
```

### Logging Usage

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Configure logging
    config := mc.DefaultConfig()
    config.EnableLog = true
    config.LogLevel = "DEBUG"        // DEBUG, INFO, WARN, ERROR
    config.LogFormat = "json"        // console, json
    config.LogDirectory = "./logs"   // Log file directory

    chunker := mc.NewMarkdownChunkerWithConfig(config)
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Processed %d chunks with detailed logging\n", len(chunks))
    fmt.Printf("Check %s directory for log files\n", config.LogDirectory)
}
```

## Supported Content Types

### Headings

- **Type**: `heading`
- **Metadata**: `heading_level` (1-6), `word_count`
- **Level**: Heading level (1-6)
- **Enhanced Features**: Position tracking, link/image extraction

### Paragraphs

- **Type**: `paragraph`
- **Metadata**: `word_count`, `char_count`
- **Level**: 0
- **Enhanced Features**: Position tracking, link/image extraction, content hashing

### Code Blocks

- **Type**: `code`
- **Metadata**: `language`, `line_count`
- **Level**: 0
- **Enhanced Features**: Code complexity analysis, syntax detection, position tracking

### Tables

- **Type**: `table`
- **Metadata**: `rows`, `columns`, `has_header`, `is_well_formed`, `alignments`, `cell_types`, `errors`, `error_count`
- **Level**: 0
- **Enhanced Features**: Advanced table analysis, format validation, cell type detection

### Lists

- **Type**: `list`
- **Metadata**: `list_type` (ordered/unordered), `item_count`
- **Level**: 0
- **Enhanced Features**: Nested list support, position tracking, link/image extraction

### Blockquotes

- **Type**: `blockquote`
- **Metadata**: `word_count`
- **Level**: 0
- **Enhanced Features**: Nested blockquote support, position tracking, link/image extraction

### Thematic Breaks

- **Type**: `thematic_break`
- **Metadata**: `type` (horizontal_rule)
- **Level**: 0
- **Enhanced Features**: Position tracking, content hashing

## Configuration Options

The library provides extensive configuration options through the `ChunkerConfig` struct:

### ChunkerConfig

```go
type ChunkerConfig struct {
    MaxChunkSize        int                    // Maximum chunk size in characters (0 = unlimited)
    EnabledTypes        map[string]bool        // Enable/disable specific content types
    CustomExtractors    []MetadataExtractor    // Custom metadata extractors
    ErrorHandling       ErrorHandlingMode      // Error handling mode
    PerformanceMode     PerformanceMode        // Performance optimization mode
    FilterEmptyChunks   bool                   // Filter out empty chunks
    PreserveWhitespace  bool                   // Preserve whitespace in content
    MemoryLimit         int64                  // Memory usage limit in bytes
    EnableObjectPooling bool                   // Enable object pooling for performance
    
    // Strategy configuration
    ChunkingStrategy    *StrategyConfig        // Chunking strategy configuration
    
    // Logging configuration
    LogLevel            string                 // Log level: DEBUG, INFO, WARN, ERROR
    EnableLog           bool                   // Enable/disable logging
    LogFormat           string                 // Log format: console, json
    LogDirectory        string                 // Log file directory
}
```

### StrategyConfig

```go
type StrategyConfig struct {
    Name        string                 `json:"name"`         // Strategy name
    Parameters  map[string]interface{} `json:"parameters"`   // Strategy parameters
    
    // Hierarchical strategy specific
    MaxDepth    int  `json:"max_depth,omitempty"`     // Maximum heading level depth
    MinDepth    int  `json:"min_depth,omitempty"`     // Minimum heading level depth
    MergeEmpty  bool `json:"merge_empty,omitempty"`   // Merge empty sections
    
    // Size constraints
    MinChunkSize int `json:"min_chunk_size,omitempty"` // Minimum chunk size
    MaxChunkSize int `json:"max_chunk_size,omitempty"` // Maximum chunk size
    
    // Content filtering
    IncludeTypes []string `json:"include_types,omitempty"` // Include content types
    ExcludeTypes []string `json:"exclude_types,omitempty"` // Exclude content types
}
```

### Error Handling Modes

```go
const (
    ErrorModeStrict     ErrorHandlingMode = iota // Stop on first error
    ErrorModePermissive                          // Log errors but continue
    ErrorModeSilent                              // Ignore errors silently
)
```

### Performance Modes

```go
const (
    PerformanceModeDefault         PerformanceMode = iota
    PerformanceModeMemoryOptimized // Optimize for memory usage
    PerformanceModeSpeedOptimized  // Optimize for processing speed
)
```

## API Reference

### Types

#### Enhanced Chunk Structure

```go
type Chunk struct {
    ID       int               `json:"id"`       // Unique chunk identifier
    Type     string            `json:"type"`     // Content type (heading, paragraph, etc.)
    Content  string            `json:"content"`  // Original Markdown content
    Text     string            `json:"text"`     // Plain text content
    Level    int               `json:"level"`    // Heading level (0 for non-headings)
    Metadata map[string]string `json:"metadata"` // Additional metadata
    
    // Enhanced fields
    Position ChunkPosition     `json:"position"` // Position in document
    Links    []Link           `json:"links"`    // Extracted links
    Images   []Image          `json:"images"`   // Extracted images
    Hash     string           `json:"hash"`     // Content hash for deduplication
}
```

#### Supporting Types

```go
type ChunkPosition struct {
    StartLine int `json:"start_line"` // Starting line number
    EndLine   int `json:"end_line"`   // Ending line number
    StartCol  int `json:"start_col"`  // Starting column number
    EndCol    int `json:"end_col"`    // Ending column number
}

type Link struct {
    Text string `json:"text"` // Link text
    URL  string `json:"url"`  // Link URL
    Type string `json:"type"` // Link type: internal, external, anchor
}

type Image struct {
    Alt    string `json:"alt"`    // Alt text
    URL    string `json:"url"`    // Image URL
    Title  string `json:"title"`  // Image title
    Width  string `json:"width"`  // Image width (if available)
    Height string `json:"height"` // Image height (if available)
}
```

### Core Functions

#### NewMarkdownChunker

```go
func NewMarkdownChunker() *MarkdownChunker
```

Creates a new Markdown chunker instance with default configuration.

#### NewMarkdownChunkerWithConfig

```go
func NewMarkdownChunkerWithConfig(config *ChunkerConfig) *MarkdownChunker
```

Creates a new Markdown chunker instance with custom configuration.

#### ChunkDocument

```go
func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error)
```

Processes a Markdown document and returns an array of semantic chunks.

### Strategy Management Functions

#### SetStrategy

```go
func (c *MarkdownChunker) SetStrategy(strategyName string, config *StrategyConfig) error
```

Sets the chunking strategy for the chunker instance.

#### RegisterStrategy

```go
func (c *MarkdownChunker) RegisterStrategy(strategy ChunkingStrategy) error
```

Registers a custom chunking strategy.

#### GetAvailableStrategies

```go
func (c *MarkdownChunker) GetAvailableStrategies() []string
```

Returns a list of all available strategy names.

#### GetCurrentStrategy

```go
func (c *MarkdownChunker) GetCurrentStrategy() string
```

Returns the name of the currently active strategy.

### Strategy Configuration Functions

#### ElementLevelConfig

```go
func ElementLevelConfig() *StrategyConfig
```

Creates configuration for element-level chunking strategy.

#### HierarchicalConfig

```go
func HierarchicalConfig(maxDepth int) *StrategyConfig
```

Creates configuration for hierarchical chunking strategy with specified maximum depth.

#### DocumentLevelConfig

```go
func DocumentLevelConfig() *StrategyConfig
```

Creates configuration for document-level chunking strategy.

#### CustomStrategyBuilder

```go
func NewCustomStrategyBuilder(name, description string) *CustomStrategyBuilder
```

Creates a new builder for custom chunking strategies.

### Error Handling Functions

```go
func (c *MarkdownChunker) GetErrors() []*ChunkerError
func (c *MarkdownChunker) HasErrors() bool
func (c *MarkdownChunker) ClearErrors()
func (c *MarkdownChunker) GetErrorsByType(errorType ErrorType) []*ChunkerError
```

### Performance Monitoring Functions

```go
func (c *MarkdownChunker) GetPerformanceStats() PerformanceStats
func (c *MarkdownChunker) GetPerformanceMonitor() *PerformanceMonitor
func (c *MarkdownChunker) ResetPerformanceMonitor()
```

### Utility Functions

```go
func DefaultConfig() *ChunkerConfig
func ValidateConfig(config *ChunkerConfig) error
```

## Logging Features

The library provides comprehensive logging capabilities to help with debugging, monitoring, and performance analysis.

### Log Levels

- **DEBUG**: Detailed information for debugging, including node processing and metadata extraction
- **INFO**: General information about processing progress and results
- **WARN**: Warning messages for potential issues
- **ERROR**: Error messages for processing failures

### Log Formats

- **console**: Human-readable format suitable for development and debugging
- **json**: Structured JSON format suitable for log aggregation and analysis

### Logging Configuration

```go
config := mc.DefaultConfig()

// Enable logging
config.EnableLog = true

// Set log level (DEBUG, INFO, WARN, ERROR)
config.LogLevel = "INFO"

// Set log format (console, json)
config.LogFormat = "console"

// Set log directory
config.LogDirectory = "./logs"
```

### Logging Examples

#### Basic Logging

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "INFO"
config.LogFormat = "console"
config.LogDirectory = "./logs"

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))
```

#### Debug Logging with JSON Format

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "json"
config.LogDirectory = "./debug-logs"

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))
```

#### Error Logging

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "ERROR"
config.LogFormat = "console"
config.LogDirectory = "./error-logs"
config.MaxChunkSize = 100  // Small limit to trigger errors
config.ErrorHandling = mc.ErrorModePermissive

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))

// Errors are logged to files and can also be retrieved programmatically
if chunker.HasErrors() {
    for _, err := range chunker.GetErrors() {
        fmt.Printf("Error: %s - %s\n", err.Type.String(), err.Message)
    }
}
```

### Log Output Examples

#### Console Format

```LOG
2024-01-15 10:30:45 INFO  [chunker.go:123] Starting document processing
2024-01-15 10:30:45 DEBUG [chunker.go:145] Processing heading node: "Introduction"
2024-01-15 10:30:45 INFO  [chunker.go:234] Document processing completed: 15 chunks, 2.3ms
```

#### JSON Format

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "message": "Starting document processing",
  "function": "ChunkDocument",
  "file": "chunker.go",
  "line": 123,
  "context": {
    "document_size": 1024,
    "config": {"max_chunk_size": 1000}
  }
}
```

## Examples

### Error Handling Example

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Configure for strict error handling
    config := mc.DefaultConfig()
    config.MaxChunkSize = 100
    config.ErrorHandling = mc.ErrorModeStrict

    chunker := mc.NewMarkdownChunkerWithConfig(config)
    
    longContent := `# Very Long Title That Exceeds The Maximum Chunk Size Limit
    
This is a very long paragraph that will definitely exceed the configured maximum chunk size limit and should trigger an error in strict mode.`

    chunks, err := chunker.ChunkDocument([]byte(longContent))
    if err != nil {
        if chunkerErr, ok := err.(*mc.ChunkerError); ok {
            fmt.Printf("Error Type: %s\n", chunkerErr.Type.String())
            fmt.Printf("Error Message: %s\n", chunkerErr.Message)
            fmt.Printf("Context: %+v\n", chunkerErr.Context)
        }
        return
    }

    fmt.Printf("Processed %d chunks\n", len(chunks))
}
```

### Performance Monitoring Example

```go
package main

import (
    "fmt"
    "time"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    config := mc.DefaultConfig()
    config.PerformanceMode = mc.PerformanceModeMemoryOptimized
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    
    largeDocument := generateLargeMarkdown() // Your large document
    
    start := time.Now()
    chunks, err := chunker.ChunkDocument([]byte(largeDocument))
    if err != nil {
        panic(err)
    }
    
    // Get performance statistics
    stats := chunker.GetPerformanceStats()
    fmt.Printf("Processing Results:\n")
    fmt.Printf("  Chunks: %d\n", len(chunks))
    fmt.Printf("  Processing Time: %v\n", stats.ProcessingTime)
    fmt.Printf("  Memory Used: %d bytes\n", stats.MemoryUsed)
    fmt.Printf("  Chunks/Second: %.2f\n", stats.ChunksPerSecond)
    fmt.Printf("  Bytes/Second: %.2f\n", stats.BytesPerSecond)
    fmt.Printf("  Peak Memory: %d bytes\n", stats.PeakMemory)
}
```

### Comprehensive Logging Example

```go
package main

import (
    "fmt"
    "os"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Create configuration with comprehensive logging
    config := mc.DefaultConfig()
    
    // Enable detailed logging
    config.EnableLog = true
    config.LogLevel = "DEBUG"
    config.LogFormat = "json"
    config.LogDirectory = "./comprehensive-logs"
    
    // Configure processing options
    config.MaxChunkSize = 1000
    config.ErrorHandling = mc.ErrorModePermissive
    config.PerformanceMode = mc.PerformanceModeSpeedOptimized
    
    // Add metadata extractors for detailed logging
    config.CustomExtractors = []mc.MetadataExtractor{
        &mc.LinkExtractor{},
        &mc.ImageExtractor{},
        &mc.CodeComplexityExtractor{},
    }
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    
    markdown := `# Logging Test Document

This document tests comprehensive logging features.

## Code Analysis

` + "```python" + `
def complex_algorithm(data):
    result = []
    for item in data:
        if item > 0:
            for i in range(item):
                if i % 2 == 0:
                    result.append(i * 2)
                else:
                    result.append(i * 3)
    return result
` + "```" + `

## Links and Images

Visit [our website](https://example.com) or check the ![logo](logo.png).

| Feature | Status | Link |
|---------|--------|------|
| Logging | Active | [docs](/logging) |
| Metrics | Beta | [metrics](/metrics) |`

    // Process with detailed logging
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        fmt.Printf("Processing error: %v\n", err)
    }
    
    fmt.Printf("Processing Results:\n")
    fmt.Printf("  Chunks created: %d\n", len(chunks))
    fmt.Printf("  Log directory: %s\n", config.LogDirectory)
    
    // Display performance stats (also logged)
    stats := chunker.GetPerformanceStats()
    fmt.Printf("  Processing time: %v\n", stats.ProcessingTime)
    fmt.Printf("  Memory used: %d KB\n", stats.MemoryUsed/1024)
    
    // Show log files created
    if files, err := os.ReadDir(config.LogDirectory); err == nil {
        fmt.Printf("  Log files created:\n")
        for _, file := range files {
            if !file.IsDir() {
                fmt.Printf("    - %s\n", file.Name())
            }
        }
    }
    
    // Display any errors (also logged)
    if chunker.HasErrors() {
        fmt.Printf("  Errors encountered: %d\n", len(chunker.GetErrors()))
        for _, err := range chunker.GetErrors() {
            fmt.Printf("    - %s: %s\n", err.Type.String(), err.Message)
        }
    }
    
    fmt.Println("\nCheck the log files for detailed processing information:")
    fmt.Println("  - DEBUG logs show node processing details")
    fmt.Println("  - INFO logs show processing progress")
    fmt.Println("  - Performance metrics are logged")
    fmt.Println("  - Error details are logged with context")
}
```

### Advanced Configuration Example

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    // Create advanced configuration
    config := mc.DefaultConfig()
    
    // Only process specific content types
    config.EnabledTypes = map[string]bool{
        "heading":    true,
        "paragraph":  true,
        "code":       true,
        "table":      true,
        "list":       false,
        "blockquote": false,
    }
    
    // Configure logging
    config.EnableLog = true
    config.LogLevel = "INFO"
    config.LogFormat = "console"
    config.LogDirectory = "./processing-logs"
    
    // Add custom metadata extractors
    config.CustomExtractors = []mc.MetadataExtractor{
        &mc.LinkExtractor{},
        &mc.ImageExtractor{},
        &mc.CodeComplexityExtractor{},
    }
    
    // Configure error handling and performance
    config.ErrorHandling = mc.ErrorModePermissive
    config.PerformanceMode = mc.PerformanceModeSpeedOptimized
    config.FilterEmptyChunks = true
    config.MaxChunkSize = 2000
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    
    markdown := `# Document with Links and Images

This paragraph contains a [link](https://example.com) and an ![image](image.jpg).

` + "```python" + `
def complex_function():
    for i in range(100):
        if i % 2 == 0:
            print(f"Even: {i}")
        else:
            print(f"Odd: {i}")
` + "```" + `

| Name | URL | Type |
|------|-----|------|
| Example | https://example.com | external |
| Internal | /page | internal |`

    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }
    
    for _, chunk := range chunks {
        fmt.Printf("\n=== %s Chunk ===\n", chunk.Type)
        fmt.Printf("Position: %d:%d to %d:%d\n", 
            chunk.Position.StartLine, chunk.Position.StartCol,
            chunk.Position.EndLine, chunk.Position.EndCol)
        
        if len(chunk.Links) > 0 {
            fmt.Printf("Links found: %d\n", len(chunk.Links))
            for _, link := range chunk.Links {
                fmt.Printf("  - %s (%s): %s\n", link.Text, link.Type, link.URL)
            }
        }
        
        if len(chunk.Images) > 0 {
            fmt.Printf("Images found: %d\n", len(chunk.Images))
            for _, img := range chunk.Images {
                fmt.Printf("  - %s: %s\n", img.Alt, img.URL)
            }
        }
        
        // Display custom metadata
        for key, value := range chunk.Metadata {
            if key == "code_complexity" || key == "link_count" || key == "image_count" {
                fmt.Printf("Custom metadata - %s: %s\n", key, value)
            }
        }
        
        fmt.Printf("Hash: %s\n", chunk.Hash[:16])
    }
    
    // Check for any processing errors
    if chunker.HasErrors() {
        fmt.Printf("\nProcessing errors: %d\n", len(chunker.GetErrors()))
        for _, err := range chunker.GetErrors() {
            fmt.Printf("  - %s: %s\n", err.Type.String(), err.Message)
        }
    }
    
    fmt.Printf("\nProcessing logged to: %s\n", config.LogDirectory)
}
```

### Filtering and Analysis

```go
// Filter chunks by type
func filterChunksByType(chunks []mc.Chunk, chunkType string) []mc.Chunk {
    var filtered []mc.Chunk
    for _, chunk := range chunks {
        if chunk.Type == chunkType {
            filtered = append(filtered, chunk)
        }
    }
    return filtered
}

// Analyze table structure
func analyzeTable(chunk mc.Chunk) {
    if chunk.Type != "table" {
        return
    }
    
    fmt.Printf("Table Analysis:\n")
    fmt.Printf("  Rows: %s\n", chunk.Metadata["rows"])
    fmt.Printf("  Columns: %s\n", chunk.Metadata["columns"])
    fmt.Printf("  Well-formed: %s\n", chunk.Metadata["is_well_formed"])
    
    if alignments, exists := chunk.Metadata["alignments"]; exists {
        fmt.Printf("  Alignments: %s\n", alignments)
    }
    
    if cellTypes, exists := chunk.Metadata["cell_types"]; exists {
        fmt.Printf("  Cell Types: %s\n", cellTypes)
    }
}

// Extract all links from document
func extractAllLinks(chunks []mc.Chunk) []mc.Link {
    var allLinks []mc.Link
    for _, chunk := range chunks {
        allLinks = append(allLinks, chunk.Links...)
    }
    return allLinks
}
```

## Migration Guide

### Upgrading to Strategy System

The chunking strategy system is fully backward compatible. Existing code will continue to work without changes, using the element-level strategy by default.

#### Existing Code (Still Works)

```go
// This continues to work exactly as before
chunker := mc.NewMarkdownChunker()
chunks, err := chunker.ChunkDocument(content)
```

#### Migrating to Explicit Strategy Configuration

```go
// Explicitly specify the same behavior as before
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)
```

#### Upgrading to Hierarchical Strategy

```go
// Switch to hierarchical chunking for better structure
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3) // Max 3 heading levels

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)
```

### Configuration Migration

If you have existing configuration files, you can migrate them using the built-in migration helper:

```go
// Load your existing configuration
oldConfig := loadExistingConfig() // Your existing config loading

// Migrate to new format
newConfig := mc.MigrateConfig(oldConfig)

// The migrated config will have element-level strategy by default
chunker := mc.NewMarkdownChunkerWithConfig(newConfig)
```

### Common Migration Patterns

#### From Fixed Processing to Strategy-Based

**Before:**
```go
chunker := mc.NewMarkdownChunker()
chunks, err := chunker.ChunkDocument(content)

// Process all chunks the same way
for _, chunk := range chunks {
    processChunk(chunk)
}
```

**After:**
```go
config := mc.DefaultConfig()

// Choose strategy based on content type
if isDocumentation(content) {
    config.ChunkingStrategy = mc.HierarchicalConfig(3)
} else if isSmallDocument(content) {
    config.ChunkingStrategy = mc.DocumentLevelConfig()
} else {
    config.ChunkingStrategy = mc.ElementLevelConfig()
}

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

// Process chunks with strategy-aware logic
for _, chunk := range chunks {
    processChunkWithStrategy(chunk, config.ChunkingStrategy.Name)
}
```

#### From Manual Grouping to Hierarchical Strategy

**Before:**
```go
chunker := mc.NewMarkdownChunker()
chunks, err := chunker.ChunkDocument(content)

// Manually group chunks by headings
groupedChunks := groupChunksByHeading(chunks)
```

**After:**
```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(2) // Group by H1 and H2

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

// Chunks are already grouped hierarchically
for _, chunk := range chunks {
    if chunk.Type == "heading" {
        fmt.Printf("Section: %s (Level %d)\n", chunk.Text, chunk.Level)
    }
}
```

### Breaking Changes

There are no breaking changes in the public API. All existing functions and types remain unchanged. The strategy system is additive and optional.

### Performance Considerations During Migration

- **Element-level strategy**: Same performance as before
- **Hierarchical strategy**: Slightly higher memory usage due to structure building
- **Document-level strategy**: Lower memory usage for small documents
- **Custom strategies**: Performance depends on implementation

### Testing Your Migration

```go
func TestMigration(t *testing.T) {
    content := []byte(`# Test Document
    
## Section 1
Content here.

## Section 2  
More content.`)

    // Test that old behavior is preserved
    oldChunker := mc.NewMarkdownChunker()
    oldChunks, err := oldChunker.ChunkDocument(content)
    assert.NoError(t, err)

    // Test that explicit element-level config produces same results
    config := mc.DefaultConfig()
    config.ChunkingStrategy = mc.ElementLevelConfig()
    newChunker := mc.NewMarkdownChunkerWithConfig(config)
    newChunks, err := newChunker.ChunkDocument(content)
    assert.NoError(t, err)

    // Should produce identical results
    assert.Equal(t, len(oldChunks), len(newChunks))
    for i, oldChunk := range oldChunks {
        assert.Equal(t, oldChunk.Type, newChunks[i].Type)
        assert.Equal(t, oldChunk.Content, newChunks[i].Content)
        assert.Equal(t, oldChunk.Text, newChunks[i].Text)
    }
}
```

## Use Cases

- **Documentation Processing**: Break down large documentation into searchable chunks with precise position tracking
- **Content Analysis**: Analyze document structure, content distribution, and extract metadata
- **RAG Systems**: Prepare Markdown content for vector databases with enhanced metadata and deduplication
- **Content Migration**: Convert Markdown documents to structured data with comprehensive error handling
- **Static Site Generation**: Process Markdown files with advanced table processing and link extraction
- **Content Quality Assurance**: Validate document structure and identify formatting issues
- **Performance-Critical Applications**: Process large documents efficiently with memory optimization
- **Multi-language Documentation**: Handle complex documents with configurable processing options

## Dependencies

- [goldmark](https://github.com/yuin/goldmark): CommonMark compliant Markdown parser

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the BSD-3 License - see the [LICENSE](LICENSE) file for details.

## Metadata Extractors

The library includes several built-in metadata extractors and supports custom extractors:

### Built-in Extractors

#### LinkExtractor

Extracts link information from content:

- `link_count`: Number of links found
- `external_links`: Count of external links
- `internal_links`: Count of internal links
- `anchor_links`: Count of anchor links

#### ImageExtractor

Extracts image information from content:

- `image_count`: Number of images found
- `image_types`: Types of images (by extension)

#### CodeComplexityExtractor

Analyzes code blocks for complexity:

- `code_complexity`: Complexity score based on control structures
- `function_count`: Number of functions detected
- `loop_count`: Number of loops detected
- `conditional_count`: Number of conditional statements

### Custom Extractors

You can create custom metadata extractors by implementing the `MetadataExtractor` interface:

```go
type CustomExtractor struct{}

func (e *CustomExtractor) Extract(node ast.Node, source []byte) map[string]string {
    metadata := make(map[string]string)
    // Your custom extraction logic here
    return metadata
}

func (e *CustomExtractor) SupportedTypes() []string {
    return []string{"heading", "paragraph"} // Specify supported types
}
```

## Error Types

The library defines several error types for comprehensive error handling:

```go
const (
    ErrorTypeInvalidInput    // Invalid or nil input
    ErrorTypeParsingFailed   // Markdown parsing failed
    ErrorTypeMemoryExhausted // Memory limit exceeded
    ErrorTypeTimeout         // Processing timeout
    ErrorTypeConfigInvalid   // Invalid configuration
    ErrorTypeChunkTooLarge   // Chunk exceeds size limit
)
```

## Performance Optimization

### Memory Optimization Features

- **Object Pooling**: Reuse objects to reduce garbage collection
- **Streaming Processing**: Process large documents without loading everything into memory
- **Memory Monitoring**: Track memory usage and detect leaks
- **Configurable Limits**: Set memory limits to prevent excessive usage

### Performance Monitoring

The library provides detailed performance statistics:

```go
type PerformanceStats struct {
    ProcessingTime  time.Duration // Total processing time
    MemoryUsed      int64         // Memory used during processing
    ChunksPerSecond float64       // Processing throughput
    BytesPerSecond  float64       // Byte processing rate
    TotalChunks     int           // Total chunks processed
    TotalBytes      int64         // Total bytes processed
    ChunkBytes      int64         // Total chunk content bytes
    PeakMemory      int64         // Peak memory usage
}
```

## Changelog

### v3.0.0 (Latest)

- **Flexible Chunking Strategies**: Multiple built-in strategies (element-level, hierarchical, document-level)
- **Custom Strategy Support**: Create custom chunking strategies using builder pattern or interface implementation
- **Strategy Registry**: Register, manage, and switch between different chunking strategies
- **Hierarchical Chunking**: Group content by heading levels with configurable depth and merging options
- **Document-Level Chunking**: Process entire documents as single chunks for specific use cases
- **Dynamic Strategy Switching**: Change strategies at runtime without recreating chunker instances
- **Strategy Configuration**: Rich configuration options for each strategy type
- **Performance Optimization**: Strategy-specific caching and memory optimization
- **Backward Compatibility**: Full compatibility with existing code - no breaking changes
- **Migration Tools**: Built-in configuration migration helpers

### v2.1.0

- **Comprehensive Logging System**: Configurable logging with multiple levels (DEBUG, INFO, WARN, ERROR)
- **Multiple Log Formats**: Support for console and JSON log formats
- **Structured Logging**: Rich context information including function names, line numbers, and processing metrics
- **Performance Logging**: Detailed performance metrics and memory usage tracking
- **Error Context Logging**: Enhanced error logging with full context information
- **Configurable Log Directory**: Flexible log file location configuration
- **Integration with All Features**: Logging integrated with error handling, performance monitoring, and metadata extraction

### v2.0.0

- **Enhanced Configuration System**: Flexible configuration with validation
- **Advanced Error Handling**: Multiple error modes with detailed error information
- **Performance Monitoring**: Built-in performance tracking and optimization
- **Enhanced Metadata Extraction**: Extensible metadata system with link, image, and code analysis
- **Position Tracking**: Precise position information for each chunk
- **Content Deduplication**: SHA256-based content hashing
- **Memory Optimization**: Object pooling and memory-efficient processing
- **Advanced Table Processing**: Improved table analysis with format validation
- **Custom Extractors**: Support for custom metadata extractors

### v1.0.0

- Initial release
- Support for all major Markdown elements
- GitHub Flavored Markdown support
- Basic metadata extraction
