package main

import (
	"fmt"
	"log"
	"strings"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== Markdown Chunker Strategy Examples ===\n")

	// Sample markdown content for testing
	markdown := `# User Guide

Welcome to our comprehensive user guide. This document will help you get started with our application.

## Getting Started

Follow these steps to begin using the application effectively.

### Installation

To install the application, run the following command:

` + "```bash" + `
npm install our-app
cd our-app
npm start
` + "```" + `

### Configuration

Edit your configuration file to customize the application:

` + "```json" + `
{
  "theme": "dark",
  "language": "en",
  "features": {
    "notifications": true,
    "analytics": false
  }
}
` + "```" + `

## Advanced Features

Learn about the advanced functionality available in the application.

### Custom Themes

You can create your own themes by following these guidelines:

1. Create a new CSS file
2. Define color variables
3. Apply the theme in settings

### Plugin System

Extend the application functionality with plugins:

| Plugin Type | Description | Status |
|-------------|-------------|--------|
| UI Themes | Custom interface themes | Active |
| Data Export | Export data in various formats | Beta |
| Integrations | Third-party service connections | Planned |

### API Integration

Connect with external services using our REST API:

` + "```javascript" + `
fetch('/api/data', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer ' + token,
    'Content-Type': 'application/json'
  }
})
.then(response => response.json())
.then(data => console.log(data));
` + "```" + `

## Troubleshooting

Common issues and their solutions.

### Performance Issues

If you experience slow performance:

- Check system requirements
- Clear application cache
- Restart the application

### Connection Problems

For network connectivity issues:

- Verify internet connection
- Check firewall settings
- Contact support if problems persist`

	// Example 1: Element-Level Strategy (Default)
	fmt.Println("1. Element-Level Strategy (Default)")
	fmt.Println("=====================================")

	config1 := mc.DefaultConfig()
	config1.ChunkingStrategy = mc.ElementLevelConfig()

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
	chunks1, err := chunker1.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total chunks: %d\n", len(chunks1))
	fmt.Println("First 5 chunks:")
	for i, chunk := range chunks1[:min(5, len(chunks1))] {
		fmt.Printf("  %d. %s: %s\n", i+1, chunk.Type, truncateText(chunk.Text, 60))
	}
	fmt.Printf("  ... and %d more chunks\n\n", max(0, len(chunks1)-5))

	// Example 2: Hierarchical Strategy
	fmt.Println("2. Hierarchical Strategy (Max Depth 2)")
	fmt.Println("======================================")

	config2 := mc.DefaultConfig()
	config2.ChunkingStrategy = mc.HierarchicalConfig(2)

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
	chunks2, err := chunker2.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total chunks: %d\n", len(chunks2))
	fmt.Println("All chunks:")
	for i, chunk := range chunks2 {
		fmt.Printf("  %d. %s (Level %d): %s\n", i+1, chunk.Type, chunk.Level, truncateText(chunk.Text, 80))
	}
	fmt.Println()

	// Example 3: Document-Level Strategy
	fmt.Println("3. Document-Level Strategy")
	fmt.Println("==========================")

	config3 := mc.DefaultConfig()
	config3.ChunkingStrategy = mc.DocumentLevelConfig()

	chunker3 := mc.NewMarkdownChunkerWithConfig(config3)
	chunks3, err := chunker3.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total chunks: %d\n", len(chunks3))
	fmt.Printf("Document chunk: %d characters\n", len(chunks3[0].Content))
	fmt.Printf("Preview: %s\n\n", truncateText(chunks3[0].Text, 100))

	// Example 4: Dynamic Strategy Switching
	fmt.Println("4. Dynamic Strategy Switching")
	fmt.Println("=============================")

	chunker := mc.NewMarkdownChunker()

	// Start with element-level
	chunks, _ := chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("Element-level: %d chunks\n", len(chunks))

	// Switch to hierarchical
	err = chunker.SetStrategy("hierarchical", mc.HierarchicalConfig(3))
	if err != nil {
		log.Fatal(err)
	}
	chunks, _ = chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("Hierarchical (depth 3): %d chunks\n", len(chunks))

	// Switch to document-level
	err = chunker.SetStrategy("document-level", mc.DocumentLevelConfig())
	if err != nil {
		log.Fatal(err)
	}
	chunks, _ = chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("Document-level: %d chunks\n\n", len(chunks))

	// Example 5: Strategy Comparison
	fmt.Println("5. Strategy Performance Comparison")
	fmt.Println("==================================")

	strategies := []struct {
		name   string
		config *mc.StrategyConfig
	}{
		{"Element-Level", mc.ElementLevelConfig()},
		{"Hierarchical (depth 2)", mc.HierarchicalConfig(2)},
		{"Hierarchical (depth 3)", mc.HierarchicalConfig(3)},
		{"Document-Level", mc.DocumentLevelConfig()},
	}

	for _, strategy := range strategies {
		config := mc.DefaultConfig()
		config.ChunkingStrategy = strategy.config

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", strategy.name, err)
			continue
		}

		stats := chunker.GetPerformanceStats()
		fmt.Printf("  %s:\n", strategy.name)
		fmt.Printf("    Chunks: %d\n", len(chunks))
		fmt.Printf("    Processing time: %v\n", stats.ProcessingTime)
		fmt.Printf("    Memory used: %d KB\n", stats.MemoryUsed/1024)
		fmt.Println()
	}

	// Example 6: Strategy with Size Constraints
	fmt.Println("6. Hierarchical Strategy with Size Constraints")
	fmt.Println("==============================================")

	constrainedConfig := mc.HierarchicalConfig(3)
	constrainedConfig.MaxChunkSize = 800 // Limit chunk size
	constrainedConfig.MinChunkSize = 100 // Minimum chunk size
	constrainedConfig.MergeEmpty = true  // Merge empty sections

	config6 := mc.DefaultConfig()
	config6.ChunkingStrategy = constrainedConfig

	chunker6 := mc.NewMarkdownChunkerWithConfig(config6)
	chunks6, err := chunker6.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total chunks: %d\n", len(chunks6))
	fmt.Println("Chunk sizes:")
	for i, chunk := range chunks6 {
		fmt.Printf("  %d. %s: %d chars\n", i+1, chunk.Type, len(chunk.Content))
	}
	fmt.Println()

	// Example 7: Content Type Filtering with Strategies
	fmt.Println("7. Strategy with Content Type Filtering")
	fmt.Println("=======================================")

	config7 := mc.DefaultConfig()
	config7.ChunkingStrategy = mc.HierarchicalConfig(2)
	config7.EnabledTypes = map[string]bool{
		"heading":   true,
		"paragraph": true,
		"code":      true,
		"table":     false, // Exclude tables
		"list":      false, // Exclude lists
	}

	chunker7 := mc.NewMarkdownChunkerWithConfig(config7)
	chunks7, err := chunker7.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total chunks (filtered): %d\n", len(chunks7))
	fmt.Println("Content types found:")
	typeCount := make(map[string]int)
	for _, chunk := range chunks7 {
		typeCount[chunk.Type]++
	}
	for contentType, count := range typeCount {
		fmt.Printf("  %s: %d\n", contentType, count)
	}
	fmt.Println()

	// Example 8: Error Handling with Strategies
	fmt.Println("8. Error Handling with Strategies")
	fmt.Println("=================================")

	// Create problematic content
	problematicMarkdown := `# Very Long Title That Might Cause Issues With Size Constraints And Processing

This is content that might cause issues with certain configurations.

## Another Section

More content here.`

	config8 := mc.DefaultConfig()
	config8.ChunkingStrategy = mc.HierarchicalConfig(2)
	config8.MaxChunkSize = 50                      // Very small limit to trigger errors
	config8.ErrorHandling = mc.ErrorModePermissive // Continue on errors

	chunker8 := mc.NewMarkdownChunkerWithConfig(config8)
	chunks8, err := chunker8.ChunkDocument([]byte(problematicMarkdown))

	if err != nil {
		fmt.Printf("Processing error: %v\n", err)
	} else {
		fmt.Printf("Processed %d chunks\n", len(chunks8))
	}

	if chunker8.HasErrors() {
		fmt.Printf("Encountered %d errors during processing:\n", len(chunker8.GetErrors()))
		for i, chunkErr := range chunker8.GetErrors() {
			fmt.Printf("  %d. %s: %s\n", i+1, chunkErr.Type.String(), chunkErr.Message)
		}
	}
	fmt.Println()

	fmt.Println("=== Strategy Examples Complete ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("- Element-level: Most chunks, fine-grained control")
	fmt.Println("- Hierarchical: Preserves document structure, fewer chunks")
	fmt.Println("- Document-level: Single chunk, complete context")
	fmt.Println("- Choose strategy based on your specific use case")
	fmt.Println("- Use size constraints to control chunk sizes")
	fmt.Println("- Enable error handling for robust processing")
}

// Helper functions
func truncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(text)
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
