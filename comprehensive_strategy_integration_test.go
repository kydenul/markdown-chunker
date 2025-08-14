package markdownchunker

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestComprehensiveStrategyIntegration 全面的策略系统集成测试
func TestComprehensiveStrategyIntegration(t *testing.T) {
	t.Run("多策略组合使用", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		testContent := []byte(`# Document Title

Introduction paragraph.

## Section 1

Content of section 1.

### Subsection 1.1

Detailed content here.

## Section 2

Content of section 2.

# Chapter 2

Second chapter content.`)

		// 测试所有可用策略
		strategies := chunker.GetAvailableStrategies()
		results := make(map[string][]Chunk)

		for _, strategyName := range strategies {
			var config *StrategyConfig
			switch strategyName {
			case "hierarchical":
				config = HierarchicalConfig(3)
			case "document-level":
				config = DocumentLevelConfig()
			case "element-level":
				config = nil // 使用默认配置
			}

			err := chunker.SetStrategy(strategyName, config)
			if err != nil {
				t.Errorf("切换到策略 %s 失败: %v", strategyName, err)
				continue
			}

			chunks, err := chunker.ChunkDocument(testContent)
			if err != nil {
				t.Errorf("使用策略 %s 分块失败: %v", strategyName, err)
				continue
			}

			results[strategyName] = chunks

			// 验证每个策略都产生了有效结果
			if len(chunks) == 0 {
				t.Errorf("策略 %s 应该产生至少一个块", strategyName)
			}

			// 验证块的基本属性
			for i, chunk := range chunks {
				if chunk.Content == "" {
					t.Errorf("策略 %s 的块 %d 内容为空", strategyName, i)
				}
				if chunk.Type == "" {
					t.Errorf("策略 %s 的块 %d 类型为空", strategyName, i)
				}
				if chunk.Metadata["strategy"] != strategyName {
					t.Errorf("策略 %s 的块 %d 策略标记错误", strategyName, i)
				}
			}
		}

		// 验证不同策略产生不同结果
		if len(results) < 2 {
			t.Skip("需要至少2个策略来比较结果")
		}

		elementChunks := results["element-level"]
		hierarchicalChunks := results["hierarchical"]
		documentChunks := results["document-level"]

		if len(elementChunks) > 0 && len(hierarchicalChunks) > 0 {
			// 层级策略通常产生更少的块
			if len(hierarchicalChunks) >= len(elementChunks) {
				t.Logf("警告: 层级策略产生的块数 (%d) 不少于元素级策略 (%d)",
					len(hierarchicalChunks), len(elementChunks))
			}
		}

		if len(documentChunks) > 0 {
			// 文档级策略应该只产生一个块
			if len(documentChunks) != 1 {
				t.Errorf("文档级策略应该产生1个块，实际产生 %d 个", len(documentChunks))
			}
		}
	})

	t.Run("策略切换正确性验证", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		testCases := []struct {
			name         string
			fromStrategy string
			toStrategy   string
			fromConfig   *StrategyConfig
			toConfig     *StrategyConfig
			expectError  bool
		}{
			{
				name:         "元素级到层级",
				fromStrategy: "element-level",
				toStrategy:   "hierarchical",
				fromConfig:   nil,
				toConfig:     HierarchicalConfig(2),
				expectError:  false,
			},
			{
				name:         "层级到文档级",
				fromStrategy: "hierarchical",
				toStrategy:   "document-level",
				fromConfig:   HierarchicalConfig(3),
				toConfig:     DocumentLevelConfig(),
				expectError:  false,
			},
			{
				name:         "文档级到元素级",
				fromStrategy: "document-level",
				toStrategy:   "element-level",
				fromConfig:   DocumentLevelConfig(),
				toConfig:     nil,
				expectError:  false,
			},
			{
				name:         "切换到不存在的策略",
				fromStrategy: "element-level",
				toStrategy:   "non-existent",
				fromConfig:   nil,
				toConfig:     nil,
				expectError:  true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 设置初始策略
				err := chunker.SetStrategy(tc.fromStrategy, tc.fromConfig)
				if err != nil {
					t.Fatalf("设置初始策略失败: %v", err)
				}

				// 验证初始策略
				currentStrategy, _ := chunker.GetCurrentStrategy()
				if currentStrategy != tc.fromStrategy {
					t.Errorf("初始策略设置错误，期望 %s，实际 %s", tc.fromStrategy, currentStrategy)
				}

				// 切换策略
				err = chunker.SetStrategy(tc.toStrategy, tc.toConfig)

				if tc.expectError {
					if err == nil {
						t.Error("期望切换失败，但成功了")
					}
					// 验证策略没有改变
					currentStrategy, _ = chunker.GetCurrentStrategy()
					if currentStrategy != tc.fromStrategy {
						t.Errorf("策略不应该改变，期望 %s，实际 %s", tc.fromStrategy, currentStrategy)
					}
				} else {
					if err != nil {
						t.Errorf("策略切换失败: %v", err)
					}
					// 验证策略已改变
					currentStrategy, _ = chunker.GetCurrentStrategy()
					if currentStrategy != tc.toStrategy {
						t.Errorf("策略切换后错误，期望 %s，实际 %s", tc.toStrategy, currentStrategy)
					}
				}
			})
		}
	})

	t.Run("并发环境下的策略安全性", func(t *testing.T) {
		testContent := []byte(`# Test Document

This is test content for concurrent processing.

## Section 1

Content here.`)

		var wg sync.WaitGroup
		errors := make(chan error, 100)
		results := make(chan []Chunk, 100)

		// 并发执行分块操作
		for i := range 50 {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// 随机选择策略
				strategies := []string{"element-level", "hierarchical", "document-level"}
				strategy := strategies[id%len(strategies)]

				var config *StrategyConfig
				switch strategy {
				case "hierarchical":
					config = HierarchicalConfig(2)
				case "document-level":
					config = DocumentLevelConfig()
				}

				// 创建独立的分块器实例以避免竞争
				localChunker := NewMarkdownChunker()
				err := localChunker.SetStrategy(strategy, config)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d 设置策略失败: %v", id, err)
					return
				}

				chunks, err := localChunker.ChunkDocument(testContent)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d 分块失败: %v", id, err)
					return
				}

				results <- chunks
			}(i)
		}

		// 等待所有goroutine完成
		go func() {
			wg.Wait()
			close(errors)
			close(results)
		}()

		// 检查错误
		var errorCount int
		for err := range errors {
			t.Errorf("并发测试错误: %v", err)
			errorCount++
		}

		// 检查结果
		var resultCount int
		for range results {
			resultCount++
		}

		if errorCount > 0 {
			t.Errorf("并发测试中有 %d 个错误", errorCount)
		}

		if resultCount != 50-errorCount {
			t.Errorf("期望 %d 个结果，实际得到 %d 个", 50-errorCount, resultCount)
		}
	})

	t.Run("策略注册和注销的线程安全性", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// 并发注册和注销策略
		for i := range 20 {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				strategyName := fmt.Sprintf("test-strategy-%d", id)
				strategy := &MockStrategy{
					name:        strategyName,
					description: fmt.Sprintf("测试策略 %d", id),
					shouldError: false,
				}

				// 注册策略
				err := chunker.RegisterStrategy(strategy)
				if err != nil {
					errors <- fmt.Errorf("注册策略 %s 失败: %v", strategyName, err)
					return
				}

				// 验证策略存在
				if !chunker.HasStrategy(strategyName) {
					errors <- fmt.Errorf("策略 %s 注册后不存在", strategyName)
					return
				}

				// 短暂等待
				time.Sleep(time.Millisecond * 10)

				// 注销策略
				err = chunker.UnregisterStrategy(strategyName)
				if err != nil {
					errors <- fmt.Errorf("注销策略 %s 失败: %v", strategyName, err)
					return
				}

				// 验证策略不存在
				if chunker.HasStrategy(strategyName) {
					errors <- fmt.Errorf("策略 %s 注销后仍然存在", strategyName)
					return
				}
			}(i)
		}

		// 等待所有goroutine完成
		go func() {
			wg.Wait()
			close(errors)
		}()

		// 检查错误
		var errorCount int
		for err := range errors {
			t.Errorf("并发注册/注销测试错误: %v", err)
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("并发注册/注销测试中有 %d 个错误", errorCount)
		}
	})

	t.Run("策略配置更新的原子性", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 设置初始策略
		initialConfig := HierarchicalConfig(2)
		err := chunker.SetStrategy("hierarchical", initialConfig)
		if err != nil {
			t.Fatalf("设置初始策略失败: %v", err)
		}

		var wg sync.WaitGroup
		errors := make(chan error, 50)

		// 并发更新策略配置
		for i := range 25 {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				config := HierarchicalConfig(id%5 + 1) // 深度1-5
				err := chunker.UpdateStrategyConfig(config)
				if err != nil {
					errors <- fmt.Errorf("更新配置失败 (goroutine %d): %v", id, err)
				}
			}(i)
		}

		// 并发执行分块操作
		testContent := []byte(`# Test

Content here.`)

		for i := range 25 {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				_, err := chunker.ChunkDocument(testContent)
				if err != nil {
					errors <- fmt.Errorf("分块失败 (goroutine %d): %v", id, err)
				}
			}(i)
		}

		// 等待所有goroutine完成
		go func() {
			wg.Wait()
			close(errors)
		}()

		// 检查错误
		var errorCount int
		for err := range errors {
			t.Errorf("并发配置更新测试错误: %v", err)
			errorCount++
		}

		if errorCount > 0 {
			t.Errorf("并发配置更新测试中有 %d 个错误", errorCount)
		}

		// 验证最终状态一致性
		currentStrategy, currentDescription := chunker.GetCurrentStrategy()
		if currentStrategy != "hierarchical" {
			t.Errorf("最终策略应该是 hierarchical，实际是 %s", currentStrategy)
		}

		if currentDescription == "" {
			t.Error("最终策略描述不应该为空")
		}
	})

	t.Run("策略执行错误恢复", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 注册一个会出错的策略
		errorStrategy := &MockStrategy{
			name:        "error-strategy",
			description: "会出错的测试策略",
			shouldError: true,
		}

		err := chunker.RegisterStrategy(errorStrategy)
		if err != nil {
			t.Fatalf("注册错误策略失败: %v", err)
		}

		// 切换到错误策略
		err = chunker.SetStrategy("error-strategy", nil)
		if err != nil {
			t.Fatalf("切换到错误策略失败: %v", err)
		}

		testContent := []byte(`# Test

Content here.`)

		// 执行分块，应该失败
		_, err = chunker.ChunkDocument(testContent)
		if err == nil {
			t.Error("使用错误策略分块应该失败")
		}

		// 切换回正常策略
		err = chunker.SetStrategy("element-level", nil)
		if err != nil {
			t.Fatalf("切换回正常策略失败: %v", err)
		}

		// 验证可以正常工作
		chunks, err := chunker.ChunkDocument(testContent)
		if err != nil {
			t.Errorf("切换回正常策略后分块失败: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("切换回正常策略后应该产生块")
		}
	})
}

// TestStrategySystemRobustness 测试策略系统的健壮性
func TestStrategySystemRobustness(t *testing.T) {
	t.Run("空内容处理", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		strategies := chunker.GetAvailableStrategies()

		for _, strategyName := range strategies {
			t.Run(fmt.Sprintf("策略_%s", strategyName), func(t *testing.T) {
				var config *StrategyConfig
				switch strategyName {
				case "hierarchical":
					config = HierarchicalConfig(2)
				case "document-level":
					config = DocumentLevelConfig()
				}

				err := chunker.SetStrategy(strategyName, config)
				if err != nil {
					t.Fatalf("设置策略失败: %v", err)
				}

				// 测试空内容
				chunks, err := chunker.ChunkDocument([]byte(""))
				if err != nil {
					t.Errorf("处理空内容失败: %v", err)
				}

				// 空内容可能产生0个或1个空块，都是合理的
				if len(chunks) > 1 {
					t.Errorf("空内容不应该产生多个块，实际产生 %d 个", len(chunks))
				}
			})
		}
	})

	t.Run("大文档处理", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 生成大文档
		var content []byte
		content = append(content, []byte("# Large Document\n\n")...)
		for i := range 1000 {
			content = append(content, fmt.Appendf(nil, "## Section %d\n\nThis is content for section %d.\n\n", i, i)...)
		}

		strategies := chunker.GetAvailableStrategies()
		for _, strategyName := range strategies {
			t.Run(fmt.Sprintf("策略_%s", strategyName), func(t *testing.T) {
				var config *StrategyConfig
				switch strategyName {
				case "hierarchical":
					config = HierarchicalConfig(2)
				case "document-level":
					config = DocumentLevelConfig()
				}

				err := chunker.SetStrategy(strategyName, config)
				if err != nil {
					t.Fatalf("设置策略失败: %v", err)
				}

				start := time.Now()
				chunks, err := chunker.ChunkDocument(content)
				duration := time.Since(start)

				if err != nil {
					t.Errorf("处理大文档失败: %v", err)
				}

				if len(chunks) == 0 {
					t.Error("大文档应该产生至少一个块")
				}

				// 性能检查：处理时间不应该过长
				if duration > time.Second*10 {
					t.Errorf("处理大文档耗时过长: %v", duration)
				}

				t.Logf("策略 %s 处理大文档: %d 块, 耗时 %v", strategyName, len(chunks), duration)
			})
		}
	})

	t.Run("异常Markdown内容处理", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		testCases := []struct {
			name    string
			content string
		}{
			{
				name:    "不完整的标题",
				content: "# Incomplete heading\n\n## Another heading without content",
			},
			{
				name:    "嵌套很深的标题",
				content: "# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6\n####### H7 (invalid)\n",
			},
			{
				name:    "特殊字符",
				content: "# 标题 with émojis 🚀\n\nContent with **bold** and *italic* and `code`.\n",
			},
			{
				name:    "混合内容",
				content: "# Title\n\n- List item 1\n- List item 2\n\n```code\ncode block\n```\n\n> Quote\n",
			},
		}

		strategies := chunker.GetAvailableStrategies()
		for _, strategyName := range strategies {
			for _, tc := range testCases {
				t.Run(fmt.Sprintf("策略_%s_%s", strategyName, tc.name), func(t *testing.T) {
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

					chunks, err := chunker.ChunkDocument([]byte(tc.content))
					if err != nil {
						t.Errorf("处理异常内容失败: %v", err)
					}

					// 验证基本属性
					for i, chunk := range chunks {
						if chunk.Content == "" && chunk.Text == "" {
							t.Errorf("块 %d 内容和文本都为空", i)
						}
					}
				})
			}
		}
	})
}
