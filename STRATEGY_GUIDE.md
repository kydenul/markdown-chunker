# Chunking Strategy Guide

This guide provides comprehensive information about using and creating chunking strategies in the Markdown Chunker library.

## Table of Contents

1. [Understanding Chunking Strategies](#understanding-chunking-strategies)
2. [Built-in Strategies](#built-in-strategies)
3. [Choosing the Right Strategy](#choosing-the-right-strategy)
4. [Custom Strategy Development](#custom-strategy-development)
5. [Performance Optimization](#performance-optimization)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Understanding Chunking Strategies

Chunking strategies determine how a Markdown document is divided into semantic chunks. Different strategies are suitable for different use cases:

- **Element-Level**: Processes each Markdown element individually
- **Hierarchical**: Groups content by heading structure
- **Document-Level**: Treats the entire document as one chunk
- **Custom**: User-defined logic for specific requirements

### Strategy Selection Flowchart

```
Document Type?
├── Small document (< 1KB) → Document-Level Strategy
├── Structured documentation → Hierarchical Strategy  
├── Search indexing → Element-Level Strategy
└── Special requirements → Custom Strategy
```

## Built-in Strategies

### Element-Level Strategy (Default)

**When to use:**
- Search indexing and retrieval systems
- Fine-grained content analysis
- Documents without clear hierarchical structure
- When you need consistent, small chunk sizes

**Configuration:**
```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()
```

**Example Output:**
```
Input: # Title\nParagraph 1\n## Subtitle\nParagraph 2
Output: 
- Chunk 1: "# Title" (heading)
- Chunk 2: "Paragraph 1" (paragraph)  
- Chunk 3: "## Subtitle" (heading)
- Chunk 4: "Paragraph 2" (paragraph)
```

### Hierarchical Strategy

**When to use:**
- Documentation with clear section structure
- Books, articles, and tutorials
- When context within sections is important
- Content that benefits from logical grouping

**Configuration:**
```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3) // Max depth 3
```

**Advanced Configuration:**
```go
strategyConfig := &mc.StrategyConfig{
    Name: "hierarchical",
    MaxDepth: 3,           // Process up to H3 headings
    MinDepth: 1,           // Start from H1 headings
    MergeEmpty: true,      // Merge empty sections
    MinChunkSize: 100,     // Minimum chunk size
    MaxChunkSize: 5000,    // Maximum chunk size
}
config.ChunkingStrategy = strategyConfig
```**Exa
mple Output:**
```
Input: # Title\nIntro\n## Section A\nContent A\n### Subsection\nSub content\n## Section B\nContent B
Output:
- Chunk 1: "# Title\nIntro" (heading + content)
- Chunk 2: "## Section A\nContent A\n### Subsection\nSub content" (section with subsection)
- Chunk 3: "## Section B\nContent B" (section)
```

### Document-Level Strategy

**When to use:**
- Small documents that should be processed as a whole
- Document classification tasks
- When you need complete document context
- Simple documents without complex structure

**Configuration:**
```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.DocumentLevelConfig()
```

**Example Output:**
```
Input: # Title\nParagraph 1\n## Subtitle\nParagraph 2
Output:
- Chunk 1: "# Title\nParagraph 1\n## Subtitle\nParagraph 2" (document)
```

## Choosing the Right Strategy

### Decision Matrix

| Use Case | Document Size | Structure | Recommended Strategy |
|----------|---------------|-----------|---------------------|
| Search indexing | Any | Any | Element-Level |
| Documentation | Large | Hierarchical | Hierarchical |
| Classification | Small | Any | Document-Level |
| Code analysis | Any | Mixed | Custom |
| RAG systems | Medium-Large | Hierarchical | Hierarchical |
| Content migration | Any | Structured | Hierarchical |

### Performance Comparison

| Strategy | Memory Usage | Processing Speed | Chunk Count |
|----------|--------------|------------------|-------------|
| Element-Level | Low | Fast | High |
| Hierarchical | Medium | Medium | Medium |
| Document-Level | Very Low | Very Fast | 1 |
| Custom | Varies | Varies | Varies |

## Custom Strategy Development

### Using the Strategy Builder

The strategy builder provides a declarative way to create custom strategies:

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func createContentFocusedStrategy() mc.ChunkingStrategy {
    builder := mc.NewCustomStrategyBuilder(
        "content-focused", 
        "Separates important content types")
    
    // Rule 1: Major headings get separate chunks (high priority)
    builder.AddRule(
        mc.HeadingLevelCondition{MinLevel: 1, MaxLevel: 2},
        mc.CreateSeparateChunkAction{},
        10, // High priority
    )
    
    // Rule 2: Code blocks always separate (high priority)
    builder.AddRule(
        mc.ContentTypeCondition{Types: []string{"code"}},
        mc.CreateSeparateChunkAction{},
        9,
    )
    
    // Rule 3: Tables always separate (high priority)
    builder.AddRule(
        mc.ContentTypeCondition{Types: []string{"table"}},
        mc.CreateSeparateChunkAction{},
        8,
    )
    
    // Rule 4: Text content merges with parent (medium priority)
    builder.AddRule(
        mc.ContentTypeCondition{Types: []string{"paragraph", "list"}},
        mc.MergeWithParentAction{},
        5,
    )
    
    // Rule 5: Minor headings merge with parent (low priority)
    builder.AddRule(
        mc.HeadingLevelCondition{MinLevel: 3, MaxLevel: 6},
        mc.MergeWithParentAction{},
        3,
    )
    
    return builder.Build()
}

func main() {
    chunker := mc.NewMarkdownChunker()
    
    // Register and use custom strategy
    strategy := createContentFocusedStrategy()
    chunker.RegisterStrategy(strategy)
    chunker.SetStrategy("content-focused", nil)
    
    // Test with sample content
    markdown := `# API Guide

Welcome to our API documentation.

## Authentication

Use these methods for authentication:

` + "```bash" + `
curl -H "Authorization: Bearer TOKEN" /api/endpoint
` + "```" + `

### API Keys

Generate API keys in the dashboard.

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| /users   | GET    | List users  |
| /posts   | POST   | Create post |

### Rate Limits

API calls are limited to 1000 per hour.`

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

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen] + "..."
}
```

### Implementing the Strategy Interface

For maximum control, implement the `ChunkingStrategy` interface directly:

```go
package main

import (
    "fmt"
    "strings"
    mc "github.com/kydenul/markdown-chunker"
    "github.com/yuin/goldmark/ast"
)

// SizeBasedStrategy creates chunks based on content size
type SizeBasedStrategy struct {
    targetSize int
    config     *mc.StrategyConfig
}

func NewSizeBasedStrategy(targetSize int) *SizeBasedStrategy {
    return &SizeBasedStrategy{
        targetSize: targetSize,
        config: &mc.StrategyConfig{
            Name: "size-based",
            Parameters: map[string]interface{}{
                "target_size": targetSize,
            },
        },
    }
}

func (s *SizeBasedStrategy) GetName() string {
    return "size-based"
}

func (s *SizeBasedStrategy) GetDescription() string {
    return fmt.Sprintf("Creates chunks targeting %d characters each", s.targetSize)
}

func (s *SizeBasedStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *mc.MarkdownChunker) ([]mc.Chunk, error) {
    var chunks []mc.Chunk
    var currentChunk strings.Builder
    var currentText strings.Builder
    chunkID := 0
    
    // Walk through all nodes and accumulate content
    ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
        if !entering {
            return ast.WalkContinue, nil
        }
        
        // Extract content from node
        nodeContent := extractNodeContent(node, source)
        nodeText := extractNodeText(node, source)
        
        // Check if adding this node would exceed target size
        if currentChunk.Len() > 0 && 
           currentChunk.Len()+len(nodeContent) > s.targetSize {
            
            // Create chunk from accumulated content
            chunk := mc.Chunk{
                ID:      chunkID,
                Type:    "mixed",
                Content: currentChunk.String(),
                Text:    currentText.String(),
                Level:   0,
                Metadata: map[string]string{
                    "strategy": "size-based",
                    "size":     fmt.Sprintf("%d", currentChunk.Len()),
                },
            }
            chunks = append(chunks, chunk)
            chunkID++
            
            // Start new chunk
            currentChunk.Reset()
            currentText.Reset()
        }
        
        // Add current node to chunk
        currentChunk.WriteString(nodeContent)
        currentText.WriteString(nodeText)
        
        return ast.WalkContinue, nil
    })
    
    // Add final chunk if there's remaining content
    if currentChunk.Len() > 0 {
        chunk := mc.Chunk{
            ID:      chunkID,
            Type:    "mixed",
            Content: currentChunk.String(),
            Text:    currentText.String(),
            Level:   0,
            Metadata: map[string]string{
                "strategy": "size-based",
                "size":     fmt.Sprintf("%d", currentChunk.Len()),
            },
        }
        chunks = append(chunks, chunk)
    }
    
    return chunks, nil
}

func (s *SizeBasedStrategy) ValidateConfig(config *mc.StrategyConfig) error {
    if targetSize, exists := config.Parameters["target_size"]; exists {
        if size, ok := targetSize.(int); ok && size <= 0 {
            return fmt.Errorf("target_size must be positive, got %d", size)
        }
    }
    return nil
}

func (s *SizeBasedStrategy) Clone() mc.ChunkingStrategy {
    return &SizeBasedStrategy{
        targetSize: s.targetSize,
        config:     s.config,
    }
}

// Helper functions (implementation depends on your needs)
func extractNodeContent(node ast.Node, source []byte) string {
    // Implementation to extract raw content from node
    return ""
}

func extractNodeText(node ast.Node, source []byte) string {
    // Implementation to extract plain text from node
    return ""
}

func main() {
    chunker := mc.NewMarkdownChunker()
    
    // Create and register size-based strategy
    strategy := NewSizeBasedStrategy(500) // 500 character target
    chunker.RegisterStrategy(strategy)
    chunker.SetStrategy("size-based", strategy.config)
    
    // Test with content
    content := generateLargeContent() // Your content generation
    chunks, err := chunker.ChunkDocument([]byte(content))
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Size-based strategy created %d chunks:\n", len(chunks))
    for i, chunk := range chunks {
        fmt.Printf("  %d. Size: %s chars, Content: %s\n", 
            i+1, chunk.Metadata["size"], truncateText(chunk.Text, 50))
    }
}
```

## Performance Optimization

### Strategy Caching

Enable strategy caching for better performance with repeated operations:

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3)
config.EnableObjectPooling = true
config.PerformanceMode = mc.PerformanceModeSpeedOptimized

chunker := mc.NewMarkdownChunkerWithConfig(config)
```

### Memory Optimization

For memory-constrained environments:

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()
config.PerformanceMode = mc.PerformanceModeMemoryOptimized
config.MemoryLimit = 50 * 1024 * 1024 // 50MB limit

chunker := mc.NewMarkdownChunkerWithConfig(config)
```

### Concurrent Processing

Process multiple documents concurrently:

```go
package main

import (
    "sync"
    mc "github.com/kydenul/markdown-chunker"
)

func processDocumentsConcurrently(documents [][]byte) [][]mc.Chunk {
    var wg sync.WaitGroup
    results := make([][]mc.Chunk, len(documents))
    
    // Create chunker for each goroutine (strategies are not thread-safe)
    for i, doc := range documents {
        wg.Add(1)
        go func(index int, content []byte) {
            defer wg.Done()
            
            config := mc.DefaultConfig()
            config.ChunkingStrategy = mc.HierarchicalConfig(3)
            chunker := mc.NewMarkdownChunkerWithConfig(config)
            
            chunks, err := chunker.ChunkDocument(content)
            if err != nil {
                // Handle error
                return
            }
            
            results[index] = chunks
        }(i, doc)
    }
    
    wg.Wait()
    return results
}
```

## Best Practices

### 1. Strategy Selection Guidelines

```go
func selectOptimalStrategy(content []byte, useCase string) *mc.StrategyConfig {
    contentSize := len(content)
    hasHeadings := detectHeadings(content)
    
    switch useCase {
    case "search":
        return mc.ElementLevelConfig()
    case "documentation":
        if hasHeadings && contentSize > 1000 {
            return mc.HierarchicalConfig(3)
        }
        return mc.ElementLevelConfig()
    case "classification":
        if contentSize < 500 {
            return mc.DocumentLevelConfig()
        }
        return mc.HierarchicalConfig(2)
    default:
        return mc.ElementLevelConfig()
    }
}
```

### 2. Error Handling

```go
func robustChunking(content []byte) ([]mc.Chunk, error) {
    config := mc.DefaultConfig()
    config.ChunkingStrategy = mc.HierarchicalConfig(3)
    config.ErrorHandling = mc.ErrorModePermissive
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    chunks, err := chunker.ChunkDocument(content)
    
    if err != nil {
        // Fallback to element-level strategy
        config.ChunkingStrategy = mc.ElementLevelConfig()
        chunker = mc.NewMarkdownChunkerWithConfig(config)
        return chunker.ChunkDocument(content)
    }
    
    return chunks, nil
}
```

### 3. Configuration Validation

```go
func validateStrategyConfig(config *mc.StrategyConfig) error {
    switch config.Name {
    case "hierarchical":
        if config.MaxDepth < 1 || config.MaxDepth > 6 {
            return fmt.Errorf("hierarchical strategy MaxDepth must be 1-6")
        }
        if config.MinDepth > config.MaxDepth {
            return fmt.Errorf("MinDepth cannot be greater than MaxDepth")
        }
    case "element-level", "document-level":
        // No specific validation needed
    default:
        return fmt.Errorf("unknown strategy: %s", config.Name)
    }
    return nil
}
```

### 4. Testing Strategies

```go
func TestStrategyBehavior(t *testing.T) {
    testCases := []struct {
        name     string
        strategy *mc.StrategyConfig
        input    string
        expected int // expected chunk count
    }{
        {
            name:     "Element-level with headings",
            strategy: mc.ElementLevelConfig(),
            input:    "# Title\nContent\n## Subtitle\nMore content",
            expected: 4,
        },
        {
            name:     "Hierarchical with depth 2",
            strategy: mc.HierarchicalConfig(2),
            input:    "# Title\nContent\n## Subtitle\nMore content",
            expected: 2,
        },
        {
            name:     "Document-level",
            strategy: mc.DocumentLevelConfig(),
            input:    "# Title\nContent\n## Subtitle\nMore content",
            expected: 1,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            config := mc.DefaultConfig()
            config.ChunkingStrategy = tc.strategy
            
            chunker := mc.NewMarkdownChunkerWithConfig(config)
            chunks, err := chunker.ChunkDocument([]byte(tc.input))
            
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, len(chunks))
        })
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Strategy Not Found

```
Error: strategy "my-strategy" not found
```

**Solution:**
```go
// Make sure to register custom strategies
chunker := mc.NewMarkdownChunker()
chunker.RegisterStrategy(myCustomStrategy)
chunker.SetStrategy("my-strategy", config)
```

#### 2. Large Chunks with Hierarchical Strategy

```
Warning: Chunk size exceeds limit
```

**Solution:**
```go
config := mc.HierarchicalConfig(3)
config.MaxChunkSize = 2000 // Set reasonable limit
config.MinChunkSize = 100  // Prevent tiny chunks
```

#### 3. Memory Issues with Large Documents

```
Error: memory limit exceeded
```

**Solution:**
```go
config := mc.DefaultConfig()
config.PerformanceMode = mc.PerformanceModeMemoryOptimized
config.MemoryLimit = 100 * 1024 * 1024 // 100MB
config.ChunkingStrategy = mc.ElementLevelConfig() // Use less memory
```

#### 4. Strategy Execution Errors

```
Error: strategy execution failed
```

**Solution:**
```go
config := mc.DefaultConfig()
config.ErrorHandling = mc.ErrorModePermissive // Continue on errors
config.ChunkingStrategy = mc.HierarchicalConfig(3)

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

// Check for partial results
if chunker.HasErrors() {
    fmt.Printf("Processed with %d errors\n", len(chunker.GetErrors()))
}
```

### Debug Mode

Enable detailed logging for troubleshooting:

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "console"
config.LogDirectory = "./debug-logs"
config.ChunkingStrategy = mc.HierarchicalConfig(3)

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

// Check debug logs for detailed execution information
```

### Performance Profiling

Profile strategy performance:

```go
import (
    "runtime"
    "time"
)

func profileStrategy(strategy *mc.StrategyConfig, content []byte) {
    config := mc.DefaultConfig()
    config.ChunkingStrategy = strategy
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    
    // Memory before
    var m1 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Time execution
    start := time.Now()
    chunks, err := chunker.ChunkDocument(content)
    duration := time.Since(start)
    
    // Memory after
    var m2 runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Strategy: %s\n", strategy.Name)
    fmt.Printf("  Chunks: %d\n", len(chunks))
    fmt.Printf("  Time: %v\n", duration)
    fmt.Printf("  Memory: %d KB\n", (m2.Alloc-m1.Alloc)/1024)
    
    if err != nil {
        fmt.Printf("  Error: %v\n", err)
    }
}
```

This guide provides comprehensive coverage of the chunking strategy system. For more specific use cases or advanced customization, refer to the API documentation and example implementations.