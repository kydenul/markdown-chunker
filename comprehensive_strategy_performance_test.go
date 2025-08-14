package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestComprehensiveStrategyPerformance 全面的策略性能测试
func TestComprehensiveStrategyPerformance(t *testing.T) {
	t.Run("各策略性能基准测试", func(t *testing.T) {
		// 生成不同大小的测试内容
		testCases := []struct {
			name    string
			content []byte
		}{
			{
				name:    "小文档",
				content: generateTestContent(100), // ~100行
			},
			{
				name:    "中等文档",
				content: generateTestContent(1000), // ~1000行
			},
			{
				name:    "大文档",
				content: generateTestContent(10000), // ~10000行
			},
		}

		strategies := []string{"element-level", "hierarchical", "document-level"}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				results := make(map[string]time.Duration)

				for _, strategyName := range strategies {
					t.Run(fmt.Sprintf("策略_%s", strategyName), func(t *testing.T) {
						chunker := NewMarkdownChunker()

						var config *StrategyConfig
						switch strategyName {
						case "hierarchical":
							config = HierarchicalConfig(3)
						case "document-level":
							config = DocumentLevelConfig()
						}

						err := chunker.SetStrategy(strategyName, config)
						if err != nil {
							t.Fatalf("设置策略失败: %v", err)
						}

						// 预热
						for range 3 {
							_, _ = chunker.ChunkDocument(tc.content)
						}

						// 性能测试
						const iterations = 10
						start := time.Now()

						for range iterations {
							chunks, err := chunker.ChunkDocument(tc.content)
							if err != nil {
								t.Errorf("分块失败: %v", err)
							}
							if len(chunks) == 0 {
								t.Error("应该产生至少一个块")
							}
						}

						duration := time.Since(start)
						avgDuration := duration / iterations
						results[strategyName] = avgDuration

						t.Logf("策略 %s 平均处理时间: %v", strategyName, avgDuration)

						// 性能阈值检查
						var threshold time.Duration
						switch tc.name {
						case "小文档":
							threshold = time.Millisecond * 10
						case "中等文档":
							threshold = time.Millisecond * 100
						case "大文档":
							threshold = time.Second * 1
						}

						if avgDuration > threshold {
							t.Errorf("策略 %s 处理 %s 耗时过长: %v > %v",
								strategyName, tc.name, avgDuration, threshold)
						}
					})
				}

				// 比较不同策略的性能
				t.Logf("性能比较 (%s):", tc.name)
				for strategy, duration := range results {
					t.Logf("  %s: %v", strategy, duration)
				}
			})
		}
	})

	t.Run("内存使用测试", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		strategies := []string{"element-level", "hierarchical", "document-level"}

		// 生成大文档
		largeContent := generateTestContent(5000)

		for _, strategyName := range strategies {
			t.Run(fmt.Sprintf("策略_%s", strategyName), func(t *testing.T) {
				var config *StrategyConfig
				switch strategyName {
				case "hierarchical":
					config = HierarchicalConfig(3)
				case "document-level":
					config = DocumentLevelConfig()
				}

				err := chunker.SetStrategy(strategyName, config)
				if err != nil {
					t.Fatalf("设置策略失败: %v", err)
				}

				// 记录初始内存
				runtime.GC()
				var m1 runtime.MemStats
				runtime.ReadMemStats(&m1)

				// 执行分块操作
				const iterations = 100
				for range iterations {
					chunks, err := chunker.ChunkDocument(largeContent)
					if err != nil {
						t.Errorf("分块失败: %v", err)
					}
					if len(chunks) == 0 {
						t.Error("应该产生至少一个块")
					}
				}

				// 记录最终内存
				runtime.GC()
				var m2 runtime.MemStats
				runtime.ReadMemStats(&m2)

				memoryUsed := m2.Alloc - m1.Alloc
				t.Logf("策略 %s 内存使用: %d bytes", strategyName, memoryUsed)

				// 内存使用阈值检查（10MB）
				const memoryThreshold = 10 * 1024 * 1024
				if memoryUsed > memoryThreshold {
					t.Errorf("策略 %s 内存使用过多: %d bytes > %d bytes",
						strategyName, memoryUsed, memoryThreshold)
				}
			})
		}
	})

	t.Run("高并发场景测试", func(t *testing.T) {
		testContent := generateTestContent(500)
		strategies := []string{"element-level", "hierarchical", "document-level"}

		for _, strategyName := range strategies {
			t.Run(fmt.Sprintf("策略_%s", strategyName), func(t *testing.T) {
				const numGoroutines = 50
				const operationsPerGoroutine = 20

				var wg sync.WaitGroup
				errors := make(chan error, numGoroutines*operationsPerGoroutine)
				durations := make(chan time.Duration, numGoroutines*operationsPerGoroutine)

				start := time.Now()

				for i := range numGoroutines {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						// 每个goroutine使用独立的分块器实例
						chunker := NewMarkdownChunker()

						var config *StrategyConfig
						switch strategyName {
						case "hierarchical":
							config = HierarchicalConfig(2)
						case "document-level":
							config = DocumentLevelConfig()
						}

						err := chunker.SetStrategy(strategyName, config)
						if err != nil {
							errors <- fmt.Errorf("goroutine %d 设置策略失败: %v", id, err)
							return
						}

						for j := range operationsPerGoroutine {
							opStart := time.Now()
							chunks, err := chunker.ChunkDocument(testContent)
							opDuration := time.Since(opStart)

							if err != nil {
								errors <- fmt.Errorf("goroutine %d 操作 %d 失败: %v", id, j, err)
								continue
							}

							if len(chunks) == 0 {
								errors <- fmt.Errorf("goroutine %d 操作 %d 没有产生块", id, j)
								continue
							}

							durations <- opDuration
						}
					}(i)
				}

				// 等待所有goroutine完成
				go func() {
					wg.Wait()
					close(errors)
					close(durations)
				}()

				totalDuration := time.Since(start)

				// 收集错误
				var errorCount int
				for err := range errors {
					t.Errorf("并发测试错误: %v", err)
					errorCount++
				}

				// 收集性能数据
				var totalOpDuration time.Duration
				var opCount int
				var maxDuration time.Duration

				for duration := range durations {
					totalOpDuration += duration
					opCount++
					if duration > maxDuration {
						maxDuration = duration
					}
				}

				if errorCount > 0 {
					t.Errorf("并发测试中有 %d 个错误", errorCount)
				}

				expectedOps := numGoroutines * operationsPerGoroutine
				if opCount != expectedOps-errorCount {
					t.Errorf("期望 %d 个操作，实际完成 %d 个", expectedOps-errorCount, opCount)
				}

				if opCount > 0 {
					avgOpDuration := totalOpDuration / time.Duration(opCount)
					throughput := float64(opCount) / totalDuration.Seconds()

					t.Logf("策略 %s 并发性能:", strategyName)
					t.Logf("  总时间: %v", totalDuration)
					t.Logf("  平均操作时间: %v", avgOpDuration)
					t.Logf("  最大操作时间: %v", maxDuration)
					t.Logf("  吞吐量: %.2f ops/sec", throughput)

					// 性能阈值检查
					if avgOpDuration > time.Millisecond*100 {
						t.Errorf("平均操作时间过长: %v", avgOpDuration)
					}

					if throughput < 10 {
						t.Errorf("吞吐量过低: %.2f ops/sec", throughput)
					}
				}
			})
		}
	})

	t.Run("策略切换性能测试", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		testContent := generateTestContent(100)

		strategies := []string{"element-level", "hierarchical", "document-level"}
		configs := map[string]*StrategyConfig{
			"element-level":  nil,
			"hierarchical":   HierarchicalConfig(2),
			"document-level": DocumentLevelConfig(),
		}

		const switchCount = 1000
		start := time.Now()

		for i := range switchCount {
			strategy := strategies[i%len(strategies)]
			config := configs[strategy]

			switchStart := time.Now()
			err := chunker.SetStrategy(strategy, config)
			switchDuration := time.Since(switchStart)

			if err != nil {
				t.Errorf("第 %d 次策略切换失败: %v", i, err)
				continue
			}

			// 验证切换成功
			currentStrategy, _ := chunker.GetCurrentStrategy()
			if currentStrategy != strategy {
				t.Errorf("第 %d 次策略切换后验证失败，期望 %s，实际 %s",
					i, strategy, currentStrategy)
			}

			// 执行一次分块操作验证功能正常
			if i%100 == 0 { // 每100次切换验证一次
				_, err = chunker.ChunkDocument(testContent)
				if err != nil {
					t.Errorf("第 %d 次切换后分块失败: %v", i, err)
				}
			}

			// 单次切换时间不应该过长
			if switchDuration > time.Millisecond*10 {
				t.Errorf("第 %d 次策略切换耗时过长: %v", i, switchDuration)
			}
		}

		totalDuration := time.Since(start)
		avgSwitchTime := totalDuration / switchCount

		t.Logf("策略切换性能:")
		t.Logf("  总切换次数: %d", switchCount)
		t.Logf("  总时间: %v", totalDuration)
		t.Logf("  平均切换时间: %v", avgSwitchTime)

		// 平均切换时间不应该过长
		if avgSwitchTime > time.Millisecond*5 {
			t.Errorf("平均策略切换时间过长: %v", avgSwitchTime)
		}
	})

	t.Run("内存泄漏检测", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		testContent := generateTestContent(1000)

		// 记录初始内存
		runtime.GC()
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// 执行大量操作
		const iterations = 1000
		for i := range iterations {
			// 切换策略
			strategies := []string{"element-level", "hierarchical", "document-level"}
			strategy := strategies[i%len(strategies)]

			var config *StrategyConfig
			switch strategy {
			case "hierarchical":
				config = HierarchicalConfig(3)
			case "document-level":
				config = DocumentLevelConfig()
			}

			err := chunker.SetStrategy(strategy, config)
			if err != nil {
				t.Errorf("第 %d 次策略切换失败: %v", i, err)
				continue
			}

			// 执行分块
			chunks, err := chunker.ChunkDocument(testContent)
			if err != nil {
				t.Errorf("第 %d 次分块失败: %v", i, err)
				continue
			}

			if len(chunks) == 0 {
				t.Errorf("第 %d 次分块没有产生块", i)
			}

			// 定期强制GC
			if i%100 == 0 {
				runtime.GC()
			}
		}

		// 最终GC和内存检查
		runtime.GC()
		runtime.GC() // 双重GC确保清理完成
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		memoryGrowth := int64(m2.Alloc) - int64(m1.Alloc)
		t.Logf("内存增长: %d bytes", memoryGrowth)

		// 内存增长阈值检查（5MB）
		const memoryGrowthThreshold = 5 * 1024 * 1024
		if memoryGrowth > memoryGrowthThreshold {
			t.Errorf("可能存在内存泄漏，内存增长: %d bytes > %d bytes",
				memoryGrowth, memoryGrowthThreshold)
		}
	})
}

// BenchmarkStrategyPerformance 策略性能基准测试
func BenchmarkStrategyPerformance(b *testing.B) {
	testCases := []struct {
		name    string
		content []byte
	}{
		{
			name:    "小文档",
			content: generateTestContent(50),
		},
		{
			name:    "中等文档",
			content: generateTestContent(500),
		},
		{
			name:    "大文档",
			content: generateTestContent(2000),
		},
	}

	strategies := []string{"element-level", "hierarchical", "document-level"}

	for _, tc := range testCases {
		for _, strategyName := range strategies {
			b.Run(fmt.Sprintf("%s_%s", tc.name, strategyName), func(b *testing.B) {
				chunker := NewMarkdownChunker()

				var config *StrategyConfig
				switch strategyName {
				case "hierarchical":
					config = HierarchicalConfig(3)
				case "document-level":
					config = DocumentLevelConfig()
				}

				err := chunker.SetStrategy(strategyName, config)
				if err != nil {
					b.Fatalf("设置策略失败: %v", err)
				}

				b.ResetTimer()
				b.ReportAllocs()

				for b.Loop() {
					chunks, err := chunker.ChunkDocument(tc.content)
					if err != nil {
						b.Errorf("分块失败: %v", err)
					}
					if len(chunks) == 0 {
						b.Error("应该产生至少一个块")
					}
				}
			})
		}
	}
}

// BenchmarkStrategySwitching 策略切换性能基准测试
func BenchmarkStrategySwitching(b *testing.B) {
	chunker := NewMarkdownChunker()
	strategies := []string{"element-level", "hierarchical", "document-level"}
	configs := map[string]*StrategyConfig{
		"element-level":  nil,
		"hierarchical":   HierarchicalConfig(2),
		"document-level": DocumentLevelConfig(),
	}

	b.ReportAllocs()

	for i := 0; b.Loop(); i++ {
		strategy := strategies[i%len(strategies)]
		config := configs[strategy]

		err := chunker.SetStrategy(strategy, config)
		if err != nil {
			b.Errorf("策略切换失败: %v", err)
		}
	}
}

// BenchmarkConcurrentChunking 并发分块性能基准测试
func BenchmarkConcurrentChunking(b *testing.B) {
	testContent := generateTestContent(200)
	strategies := []string{"element-level", "hierarchical", "document-level"}

	for _, strategyName := range strategies {
		b.Run(fmt.Sprintf("策略_%s", strategyName), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				chunker := NewMarkdownChunker()

				var config *StrategyConfig
				switch strategyName {
				case "hierarchical":
					config = HierarchicalConfig(2)
				case "document-level":
					config = DocumentLevelConfig()
				}

				err := chunker.SetStrategy(strategyName, config)
				if err != nil {
					b.Errorf("设置策略失败: %v", err)
					return
				}

				for pb.Next() {
					chunks, err := chunker.ChunkDocument(testContent)
					if err != nil {
						b.Errorf("分块失败: %v", err)
					}
					if len(chunks) == 0 {
						b.Error("应该产生至少一个块")
					}
				}
			})
		})
	}
}

// generateTestContent 生成测试内容
func generateTestContent(lines int) []byte {
	var content strings.Builder

	content.WriteString("# Test Document\n\n")
	content.WriteString("This is a test document for performance testing.\n\n")

	// 每20行一个section
	sectionCount := max(lines/20, 1)

	for i := range sectionCount {
		content.WriteString(fmt.Sprintf("## Section %d\n\n", i+1))
		content.WriteString(fmt.Sprintf("This is the introduction for section %d.\n\n", i+1))

		// 添加子section
		subSections := 3
		for j := range subSections {
			content.WriteString(fmt.Sprintf("### Subsection %d.%d\n\n", i+1, j+1))
			content.WriteString(fmt.Sprintf("Content for subsection %d.%d with some detailed information.\n\n", i+1, j+1))

			// 添加段落
			paragraphs := 2
			for k := range paragraphs {
				content.WriteString(fmt.Sprintf("This is paragraph %d in subsection %d.%d. ", k+1, i+1, j+1))
				content.WriteString("It contains some sample text to make the document longer and more realistic. ")
				content.WriteString("We need enough content to test performance properly.\n\n")
			}

			// 添加列表
			content.WriteString("Here's a list:\n\n")
			for l := range 3 {
				content.WriteString(fmt.Sprintf("- List item %d\n", l+1))
			}
			content.WriteString("\n")

			// 添加代码块
			if j%2 == 0 {
				content.WriteString("```go\n")
				content.WriteString("func example() {\n")
				content.WriteString("    fmt.Println(\"Hello, World!\")\n")
				content.WriteString("}\n")
				content.WriteString("```\n\n")
			}
		}
	}

	return []byte(content.String())
}

// TestStrategyPerformanceRegression 性能回归测试
func TestStrategyPerformanceRegression(t *testing.T) {
	// 这个测试用于检测性能回归
	// 在实际项目中，可以将基准结果保存到文件中进行比较

	chunker := NewMarkdownChunker()
	testContent := generateTestContent(1000)

	strategies := []string{"element-level", "hierarchical", "document-level"}

	// 性能基准（这些值应该根据实际环境调整）
	performanceBaselines := map[string]time.Duration{
		"element-level":  time.Millisecond * 50,
		"hierarchical":   time.Millisecond * 100,
		"document-level": time.Millisecond * 30,
	}

	for _, strategyName := range strategies {
		t.Run(fmt.Sprintf("回归测试_%s", strategyName), func(t *testing.T) {
			var config *StrategyConfig
			switch strategyName {
			case "hierarchical":
				config = HierarchicalConfig(3)
			case "document-level":
				config = DocumentLevelConfig()
			}

			err := chunker.SetStrategy(strategyName, config)
			if err != nil {
				t.Fatalf("设置策略失败: %v", err)
			}

			// 预热
			for range 5 {
				_, _ = chunker.ChunkDocument(testContent)
			}

			// 性能测试
			const iterations = 10
			start := time.Now()

			for range iterations {
				chunks, err := chunker.ChunkDocument(testContent)
				if err != nil {
					t.Errorf("分块失败: %v", err)
				}
				if len(chunks) == 0 {
					t.Error("应该产生至少一个块")
				}
			}

			duration := time.Since(start)
			avgDuration := duration / iterations

			baseline := performanceBaselines[strategyName]

			t.Logf("策略 %s 平均处理时间: %v (基准: %v)", strategyName, avgDuration, baseline)

			// 允许20%的性能波动
			threshold := baseline + baseline/5
			if avgDuration > threshold {
				t.Errorf("策略 %s 性能回归，当前: %v，阈值: %v",
					strategyName, avgDuration, threshold)
			}
		})
	}
}
