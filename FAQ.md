# Frequently Asked Questions (FAQ)

## General Questions

### Q: What are chunking strategies and why do I need them?

**A:** Chunking strategies determine how your Markdown documents are divided into semantic pieces. Different strategies are optimal for different use cases:

- **Search systems** benefit from element-level chunking for precise matching
- **Documentation** works better with hierarchical chunking to preserve context
- **Classification tasks** may need document-level chunking for complete context
- **Specialized applications** can use custom strategies for specific requirements

### Q: Is the strategy system backward compatible?

**A:** Yes, completely. All existing code continues to work without changes. The library uses element-level strategy by default, which maintains the original behavior.

```go
// This still works exactly as before
chunker := mc.NewMarkdownChunker()
chunks, err := chunker.ChunkDocument(content)
```

### Q: Which strategy should I use for my use case?

**A:** Here's a quick decision guide:

| Use Case | Recommended Strategy | Reason |
|----------|---------------------|---------|
| Search indexing | Element-Level | Fine-grained matching |
| Documentation | Hierarchical | Preserves section context |
| RAG systems | Hierarchical | Maintains logical structure |
| Document classification | Document-Level | Complete context |
| Code analysis | Custom | Specialized requirements |

## Strategy-Specific Questions

### Q: How does hierarchical chunking work?

**A:** Hierarchical chunking groups content by heading levels:

```markdown
# Main Title          ← Level 1 heading
Introduction text     ← Grouped with Main Title

## Section A          ← Level 2 heading  
Section A content     ← Grouped with Section A
### Subsection       ← Level 3 heading (grouped with Section A if MaxDepth=2)
Subsection content    ← Grouped with parent section

## Section B          ← Level 2 heading
Section B content     ← Grouped with Section B
```

With `HierarchicalConfig(2)`, you get:
- Chunk 1: "Main Title + Introduction text"
- Chunk 2: "Section A + content + Subsection + content" 
- Chunk 3: "Section B + content"

### Q: Can I limit the size of hierarchical chunks?

**A:** Yes, use size constraints:

```go
config := mc.HierarchicalConfig(3)
config.MaxChunkSize = 2000  // Split large sections
config.MinChunkSize = 100   // Merge small sections
```

### Q: What happens to content without headings?

**A:** Content without headings is handled differently by each strategy:

- **Element-Level**: Each element becomes a separate chunk
- **Hierarchical**: Content is grouped with the nearest preceding heading
- **Document-Level**: All content is in one chunk
- **Custom**: Depends on your implementation

## Performance Questions

### Q: Which strategy is fastest?

**A:** Performance ranking (fastest to slowest):

1. **Document-Level**: Minimal processing, one chunk
2. **Element-Level**: Simple node-by-node processing  
3. **Hierarchical**: Requires structure analysis and grouping
4. **Custom**: Depends on implementation complexity

### Q: How much memory do different strategies use?

**A:** Memory usage comparison:

- **Document-Level**: Lowest (single chunk)
- **Element-Level**: Low (processes nodes individually)
- **Hierarchical**: Medium (builds structure tree)
- **Custom**: Varies by implementation

### Q: Can I process multiple documents concurrently?

**A:** Yes, but create separate chunker instances for each goroutine:

```go
func processDocuments(docs [][]byte) [][]mc.Chunk {
    var wg sync.WaitGroup
    results := make([][]mc.Chunk, len(docs))
    
    for i, doc := range docs {
        wg.Add(1)
        go func(index int, content []byte) {
            defer wg.Done()
            
            // Create separate chunker for each goroutine
            chunker := mc.NewMarkdownChunker()
            chunks, _ := chunker.ChunkDocument(content)
            results[index] = chunks
        }(i, doc)
    }
    
    wg.Wait()
    return results
}
```

## Configuration Questions

### Q: How do I migrate my existing configuration?

**A:** Use the built-in migration helper:

```go
// Your existing config
oldConfig := &mc.ChunkerConfig{
    MaxChunkSize: 1000,
    EnabledTypes: map[string]bool{"heading": true, "paragraph": true},
}

// Migrate to include strategy
newConfig := mc.MigrateConfig(oldConfig)

// newConfig now includes element-level strategy by default
chunker := mc.NewMarkdownChunkerWithConfig(newConfig)
```

### Q: Can I change strategies at runtime?

**A:** Yes, use `SetStrategy()`:

```go
chunker := mc.NewMarkdownChunker()

// Process with element-level
chunks1, _ := chunker.ChunkDocument(content)

// Switch to hierarchical
chunker.SetStrategy("hierarchical", mc.HierarchicalConfig(3))
chunks2, _ := chunker.ChunkDocument(content)
```

### Q: How do I validate my strategy configuration?

**A:** The library validates configurations automatically, but you can also validate manually:

```go
config := mc.HierarchicalConfig(3)
config.MaxDepth = 10 // Invalid: exceeds maximum

chunker := mc.NewMarkdownChunker()
err := chunker.SetStrategy("hierarchical", config)
if err != nil {
    fmt.Printf("Configuration error: %v\n", err)
}
```

## Custom Strategy Questions

### Q: When should I create a custom strategy?

**A:** Create custom strategies when:

- Built-in strategies don't meet your specific requirements
- You need specialized content handling (e.g., code-focused chunking)
- You have domain-specific document structures
- You need to integrate with existing systems that expect specific chunk formats

### Q: What's easier: Strategy Builder or implementing the interface?

**A:** Choose based on complexity:

**Strategy Builder** (easier):
- Good for rule-based logic
- Declarative approach
- Built-in condition and action types
- Less code to write

**Interface Implementation** (more control):
- Complete control over chunking logic
- Can implement complex algorithms
- Better performance for specialized cases
- More code but maximum flexibility

### Q: Can I combine multiple strategies?

**A:** Not directly, but you can create a custom strategy that uses multiple approaches:

```go
type HybridStrategy struct{}

func (s *HybridStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *mc.MarkdownChunker) ([]mc.Chunk, error) {
    // Use hierarchical for structured content
    if hasHeadings(doc) {
        hierarchical := &mc.HierarchicalStrategy{}
        return hierarchical.ChunkDocument(doc, source, chunker)
    }
    
    // Use element-level for unstructured content
    elementLevel := &mc.ElementLevelStrategy{}
    return elementLevel.ChunkDocument(doc, source, chunker)
}
```

## Error Handling Questions

### Q: What happens if a strategy fails?

**A:** Depends on your error handling mode:

```go
config := mc.DefaultConfig()
config.ErrorHandling = mc.ErrorModePermissive // Continue on errors
config.ChunkingStrategy = mc.HierarchicalConfig(3)

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

if err != nil {
    // Strategy failed completely
    fmt.Printf("Strategy error: %v\n", err)
} else if chunker.HasErrors() {
    // Partial success with some errors
    fmt.Printf("Processed with %d errors\n", len(chunker.GetErrors()))
}
```

### Q: How do I debug strategy issues?

**A:** Enable detailed logging:

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "console"
config.LogDirectory = "./debug-logs"

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument(content)

// Check debug logs for detailed execution trace
```

### Q: Can I recover from strategy failures?

**A:** Yes, implement fallback logic:

```go
func robustChunking(content []byte) ([]mc.Chunk, error) {
    strategies := []string{"hierarchical", "element-level", "document-level"}
    
    for _, strategyName := range strategies {
        chunker := mc.NewMarkdownChunker()
        
        var config *mc.StrategyConfig
        switch strategyName {
        case "hierarchical":
            config = mc.HierarchicalConfig(3)
        case "element-level":
            config = mc.ElementLevelConfig()
        case "document-level":
            config = mc.DocumentLevelConfig()
        }
        
        chunker.SetStrategy(strategyName, config)
        chunks, err := chunker.ChunkDocument(content)
        
        if err == nil {
            return chunks, nil
        }
        
        fmt.Printf("Strategy %s failed: %v\n", strategyName, err)
    }
    
    return nil, fmt.Errorf("all strategies failed")
}
```

## Integration Questions

### Q: How do I integrate with existing RAG systems?

**A:** Use hierarchical chunking for better context:

```go
func prepareForRAG(documents []string) []RAGChunk {
    config := mc.DefaultConfig()
    config.ChunkingStrategy = mc.HierarchicalConfig(3)
    config.MaxChunkSize = 1500 // Optimal for most embedding models
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    var ragChunks []RAGChunk
    
    for docID, content := range documents {
        chunks, err := chunker.ChunkDocument([]byte(content))
        if err != nil {
            continue
        }
        
        for _, chunk := range chunks {
            ragChunk := RAGChunk{
                ID:       fmt.Sprintf("%s_%d", docID, chunk.ID),
                Content:  chunk.Text,
                Metadata: chunk.Metadata,
                Source:   docID,
            }
            ragChunks = append(ragChunks, ragChunk)
        }
    }
    
    return ragChunks
}
```

### Q: How do I use this with search engines?

**A:** Element-level chunking works well for search indexing:

```go
func indexForSearch(documents map[string]string) []SearchDocument {
    config := mc.DefaultConfig()
    config.ChunkingStrategy = mc.ElementLevelConfig()
    
    chunker := mc.NewMarkdownChunkerWithConfig(config)
    var searchDocs []SearchDocument
    
    for docID, content := range documents {
        chunks, err := chunker.ChunkDocument([]byte(content))
        if err != nil {
            continue
        }
        
        for _, chunk := range chunks {
            searchDoc := SearchDocument{
                ID:      fmt.Sprintf("%s_%d", docID, chunk.ID),
                Title:   extractTitle(chunk),
                Content: chunk.Text,
                Type:    chunk.Type,
                Source:  docID,
            }
            searchDocs = append(searchDocs, searchDoc)
        }
    }
    
    return searchDocs
}
```

### Q: Can I export chunks to different formats?

**A:** Yes, chunks are regular Go structs that can be serialized:

```go
// JSON export
func exportToJSON(chunks []mc.Chunk) ([]byte, error) {
    return json.MarshalIndent(chunks, "", "  ")
}

// CSV export
func exportToCSV(chunks []mc.Chunk) ([]byte, error) {
    var buffer bytes.Buffer
    writer := csv.NewWriter(&buffer)
    
    // Header
    writer.Write([]string{"ID", "Type", "Content", "Text", "Level"})
    
    // Data
    for _, chunk := range chunks {
        record := []string{
            fmt.Sprintf("%d", chunk.ID),
            chunk.Type,
            chunk.Content,
            chunk.Text,
            fmt.Sprintf("%d", chunk.Level),
        }
        writer.Write(record)
    }
    
    writer.Flush()
    return buffer.Bytes(), writer.Error()
}

// Database export
func exportToDatabase(chunks []mc.Chunk, db *sql.DB) error {
    stmt, err := db.Prepare(`
        INSERT INTO chunks (id, type, content, text, level, metadata) 
        VALUES (?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, chunk := range chunks {
        metadataJSON, _ := json.Marshal(chunk.Metadata)
        _, err := stmt.Exec(
            chunk.ID, chunk.Type, chunk.Content, 
            chunk.Text, chunk.Level, string(metadataJSON),
        )
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## Troubleshooting Common Issues

### Q: My hierarchical chunks are too large

**A:** Set size limits:

```go
config := mc.HierarchicalConfig(2) // Reduce depth
config.MaxChunkSize = 1000         // Set size limit
config.MinChunkSize = 100          // Prevent tiny chunks
```

### Q: Element-level chunking creates too many small chunks

**A:** Use filtering and merging:

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()
config.FilterEmptyChunks = true
config.MinChunkSize = 50 // Filter very small chunks
```

### Q: Custom strategy is not working

**A:** Check common issues:

1. **Registration**: Make sure you registered the strategy
2. **Interface**: Verify all interface methods are implemented
3. **Validation**: Check if `ValidateConfig()` passes
4. **Thread Safety**: Ensure `Clone()` creates independent instances

```go
// Debug custom strategy
func debugCustomStrategy(strategy mc.ChunkingStrategy) {
    fmt.Printf("Strategy name: %s\n", strategy.GetName())
    fmt.Printf("Description: %s\n", strategy.GetDescription())
    
    // Test validation
    config := &mc.StrategyConfig{Name: strategy.GetName()}
    if err := strategy.ValidateConfig(config); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    }
    
    // Test cloning
    clone := strategy.Clone()
    fmt.Printf("Clone created: %v\n", clone != nil)
}
```

### Q: Performance is slower than expected

**A:** Optimize configuration:

```go
config := mc.DefaultConfig()
config.PerformanceMode = mc.PerformanceModeSpeedOptimized
config.EnableObjectPooling = true
config.ChunkingStrategy = mc.ElementLevelConfig() // Fastest strategy

// For memory optimization instead
config.PerformanceMode = mc.PerformanceModeMemoryOptimized
config.MemoryLimit = 50 * 1024 * 1024 // 50MB limit
```

## Getting Help

### Q: Where can I find more examples?

**A:** Check these resources:

1. **Example directory**: `example/` in the repository
2. **Strategy guide**: `STRATEGY_GUIDE.md`
3. **API documentation**: `README.md`
4. **Test files**: `*_test.go` files show usage patterns

### Q: How do I report bugs or request features?

**A:** 

1. **Check existing issues** in the repository
2. **Provide minimal reproduction case**
3. **Include configuration and input data**
4. **Specify expected vs actual behavior**

### Q: Can I contribute custom strategies?

**A:** Yes! Contributions are welcome:

1. **Implement the strategy** with comprehensive tests
2. **Add documentation** and usage examples  
3. **Follow the existing code style**
4. **Submit a pull request** with description

For more specific questions or advanced use cases, please refer to the detailed documentation or open an issue in the repository.