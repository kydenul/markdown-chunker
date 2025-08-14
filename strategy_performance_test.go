package markdownchunker

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/yuin/goldmark/text"
)

// BenchmarkStrategyExecution 测试策略执行性能
func BenchmarkStrategyExecution(b *testing.B) {
	content := []byte(`# Test Document

This is a test document with multiple sections.

## Section 1

Content for section 1 with some text.

### Subsection 1.1

More detailed content here.

## Section 2

Content for section 2.

### Subsection 2.1

Another subsection with content.

### Subsection 2.2

Final subsection content.`)

	b.Run("ElementLevel", func(b *testing.B) {
		chunker := NewMarkdownChunker()
		err := chunker.SetStrategy("element-level", nil)
		if err != nil {
			b.Fatalf("Failed to set strategy: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})

	b.Run("Hierarchical", func(b *testing.B) {
		chunker := NewMarkdownChunker()
		err := chunker.SetStrategy("hierarchical", HierarchicalConfig(3))
		if err != nil {
			b.Fatalf("Failed to set strategy: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})

	b.Run("DocumentLevel", func(b *testing.B) {
		chunker := NewMarkdownChunker()
		err := chunker.SetStrategy("document-level", nil)
		if err != nil {
			b.Fatalf("Failed to set strategy: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})
}

// BenchmarkStrategyCaching 测试策略缓存性能
func BenchmarkStrategyCaching(b *testing.B) {
	content := []byte(`# Test Document
Content here.`)

	b.Run("WithCaching", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 切换策略以测试缓存效果
			strategyName := []string{"element-level", "hierarchical", "document-level"}[i%3]
			chunker.SetStrategy(strategyName, nil)
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})

	b.Run("WithoutCaching", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 每次创建新的分块器（无缓存）
			chunker := NewMarkdownChunker()
			strategyName := []string{"element-level", "hierarchical", "document-level"}[i%3]
			chunker.SetStrategy(strategyName, nil)
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentStrategyExecution 测试并发策略执行性能
func BenchmarkConcurrentStrategyExecution(b *testing.B) {
	content := []byte(`# Test Document
This is a test document for concurrent processing.

## Section 1
Content for section 1.

## Section 2
Content for section 2.`)

	b.Run("Sequential", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatalf("ChunkDocument failed: %v", err)
			}
		}
	})

	b.Run("Concurrent", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := chunker.ChunkDocument(content)
				if err != nil {
					b.Fatalf("ChunkDocument failed: %v", err)
				}
			}
		})
	})
}

// BenchmarkLargeDocumentProcessingWithStrategies 测试大文档处理性能
func BenchmarkLargeDocumentProcessingWithStrategies(b *testing.B) {
	// 生成大文档
	var builder strings.Builder
	builder.WriteString("# Large Document\n\n")

	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("## Section %d\n\n", i+1))
		for j := 0; j < 10; j++ {
			builder.WriteString(fmt.Sprintf("### Subsection %d.%d\n\n", i+1, j+1))
			builder.WriteString("This is some content for the subsection. ")
			builder.WriteString("It contains multiple sentences to make it more realistic. ")
			builder.WriteString("We want to test how well the chunker handles large documents.\n\n")
		}
	}

	content := []byte(builder.String())

	strategies := []struct {
		name   string
		config *StrategyConfig
	}{
		{"element-level", nil},
		{"hierarchical", HierarchicalConfig(3)},
		{"document-level", nil},
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			chunker := NewMarkdownChunker()
			err := chunker.SetStrategy(strategy.name, strategy.config)
			if err != nil {
				b.Fatalf("Failed to set strategy: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := chunker.ChunkDocument(content)
				if err != nil {
					b.Fatalf("ChunkDocument failed: %v", err)
				}
			}
		})
	}
}

// TestStrategyPooling 测试策略池化功能
func TestStrategyPooling(t *testing.T) {
	chunker := NewMarkdownChunker()
	content := []byte("# Test\nContent")

	// 测试策略实例复用
	var instances []ChunkingStrategy
	for i := 0; i < 5; i++ {
		instance := chunker.getOptimizedStrategyInstance("element-level")
		instances = append(instances, instance)

		// 使用实例 - 需要先解析文档
		reader := text.NewReader(content)
		doc := chunker.md.Parser().Parse(reader)
		_, err := instance.ChunkDocument(doc, content, chunker)
		if err != nil {
			t.Errorf("Strategy execution failed: %v", err)
		}

		// 返回到池中
		chunker.returnStrategyInstance("element-level", instance)
	}

	// 验证缓存统计
	stats := chunker.GetCacheStats()
	if cacheSize, ok := stats["cache_size"]; !ok || cacheSize.(int) == 0 {
		t.Error("Expected cache to contain strategies")
	}

	if poolCount, ok := stats["pool_count"]; !ok || poolCount.(int) == 0 {
		t.Error("Expected pool to contain strategy pools")
	}
}

// TestStrategyCaching 测试策略缓存功能
func TestStrategyCaching(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 初始缓存应该为空
	stats := chunker.GetCacheStats()
	if cacheSize := stats["cache_size"].(int); cacheSize != 0 {
		t.Errorf("Expected empty cache, got size %d", cacheSize)
	}

	// 获取策略实例应该填充缓存
	instance1 := chunker.getOptimizedStrategyInstance("element-level")
	if instance1 == nil {
		t.Error("Expected strategy instance, got nil")
	}

	// 缓存应该包含策略
	stats = chunker.GetCacheStats()
	if cacheSize := stats["cache_size"].(int); cacheSize == 0 {
		t.Error("Expected cache to contain strategies after first access")
	}

	// 第二次获取应该使用缓存
	instance2 := chunker.getOptimizedStrategyInstance("element-level")
	if instance2 == nil {
		t.Error("Expected strategy instance from cache, got nil")
	}

	// 清空缓存
	chunker.ClearStrategyCache()
	stats = chunker.GetCacheStats()
	if cacheSize := stats["cache_size"].(int); cacheSize != 0 {
		t.Errorf("Expected empty cache after clear, got size %d", cacheSize)
	}
}

// TestConcurrentStrategyCaching 测试并发策略缓存
func TestConcurrentStrategyCaching(t *testing.T) {
	chunker := NewMarkdownChunker()
	content := []byte("# Test\nContent")

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// 并发访问策略缓存
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				strategyName := []string{"element-level", "hierarchical", "document-level"}[j%3]

				// 获取策略实例
				instance := chunker.getOptimizedStrategyInstance(strategyName)
				if instance == nil {
					t.Errorf("Goroutine %d: Expected strategy instance, got nil", id)
					return
				}

				// 使用策略 - 需要先解析文档
				reader := text.NewReader(content)
				doc := chunker.md.Parser().Parse(reader)
				_, err := instance.ChunkDocument(doc, content, chunker)
				if err != nil {
					t.Errorf("Goroutine %d: Strategy execution failed: %v", id, err)
					return
				}

				// 返回到池中
				chunker.returnStrategyInstance(strategyName, instance)
			}
		}(i)
	}

	wg.Wait()

	// 验证缓存状态
	stats := chunker.GetCacheStats()
	if cacheSize := stats["cache_size"].(int); cacheSize == 0 {
		t.Error("Expected cache to contain strategies after concurrent access")
	}
}

// BenchmarkStrategyInstanceCreation 测试策略实例创建性能
func BenchmarkStrategyInstanceCreation(b *testing.B) {
	b.Run("WithPooling", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			instance := chunker.getOptimizedStrategyInstance("element-level")
			chunker.returnStrategyInstance("element-level", instance)
		}
	})

	b.Run("WithoutPooling", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// 直接创建新实例（无池化）
			strategy, _ := chunker.strategyRegistry.Get("element-level")
			_ = strategy.Clone()
		}
	})
}

// TestStrategyPerformanceMonitoring 测试策略性能监控
func TestStrategyPerformanceMonitoring(t *testing.T) {
	chunker := NewMarkdownChunker()
	content := []byte(`# Test Document
This is a test document.

## Section 1
Content for section 1.`)

	// 执行分块以生成性能数据
	_, err := chunker.ChunkDocument(content)
	if err != nil {
		t.Fatalf("ChunkDocument failed: %v", err)
	}

	// 获取性能统计
	stats := chunker.performanceMonitor.GetStats()

	if stats.ProcessingTime == 0 {
		t.Error("Expected non-zero processing time")
	}

	if stats.TotalChunks == 0 {
		t.Error("Expected non-zero chunk count")
	}
}
