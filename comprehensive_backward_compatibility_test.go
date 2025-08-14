package markdownchunker

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestComprehensiveBackwardCompatibility 全面的向后兼容性测试
func TestComprehensiveBackwardCompatibility(t *testing.T) {
	t.Run("现有API完全兼容性", func(t *testing.T) {
		t.Run("默认构造函数", func(t *testing.T) {
			// 使用默认构造函数应该与之前版本行为一致
			chunker := NewMarkdownChunker()

			// 验证默认策略
			strategyName, _ := chunker.GetCurrentStrategy()
			if strategyName != "element-level" {
				t.Errorf("默认策略应该是 element-level，实际是 %s", strategyName)
			}

			// 验证基本分块功能
			content := []byte(`# Heading 1

This is a paragraph.

## Heading 2

Another paragraph.

- List item 1
- List item 2

> This is a quote.`)

			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			// 验证结果与之前版本一致
			expectedTypes := []string{"heading", "paragraph", "heading", "paragraph", "list", "blockquote"}
			if len(chunks) != len(expectedTypes) {
				t.Errorf("块数量不匹配，期望 %d，实际 %d", len(expectedTypes), len(chunks))
			}

			for i, chunk := range chunks {
				if i < len(expectedTypes) && chunk.Type != expectedTypes[i] {
					t.Errorf("块 %d 类型错误，期望 '%s'，实际 '%s'", i, expectedTypes[i], chunk.Type)
				}
			}

			// 验证块的基本属性
			for i, chunk := range chunks {
				if chunk.ID != i {
					t.Errorf("块 %d ID错误，期望 %d，实际 %d", i, i, chunk.ID)
				}
				if chunk.Content == "" {
					t.Errorf("块 %d 内容为空", i)
				}
				if chunk.Text == "" {
					t.Errorf("块 %d 文本为空", i)
				}
			}
		})

		t.Run("配置构造函数", func(t *testing.T) {
			// 使用配置构造函数，不指定策略
			config := &ChunkerConfig{
				MaxChunkSize: 1000,
				EnabledTypes: map[string]bool{
					"heading":   true,
					"paragraph": true,
					"list":      true,
				},
			}

			chunker := NewMarkdownChunkerWithConfig(config)

			// 应该使用默认策略
			strategyName, _ := chunker.GetCurrentStrategy()
			if strategyName != "element-level" {
				t.Errorf("应该使用默认策略，实际是 %s", strategyName)
			}

			content := []byte(`# Test Heading

This is a test paragraph.

- Item 1
- Item 2

> This quote should be filtered out.`)

			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			// 验证配置仍然生效（过滤掉blockquote）
			for _, chunk := range chunks {
				if chunk.Type == "blockquote" {
					t.Error("不应该包含被过滤的blockquote类型")
				}
			}

			// 验证包含的类型
			foundTypes := make(map[string]bool)
			for _, chunk := range chunks {
				foundTypes[chunk.Type] = true
			}

			expectedTypes := []string{"heading", "paragraph", "list"}
			for _, expectedType := range expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("应该包含类型 %s", expectedType)
				}
			}
		})

		t.Run("现有方法保持不变", func(t *testing.T) {
			chunker := NewMarkdownChunker()

			// 验证所有现有方法仍然可用且行为一致
			methods := []string{
				"ChunkDocument",
				"GetErrors",
				"HasErrors",
				"ClearErrors",
			}

			chunkerType := reflect.TypeOf(chunker)
			for _, methodName := range methods {
				method, exists := chunkerType.MethodByName(methodName)
				if !exists {
					t.Errorf("方法 %s 不存在", methodName)
					continue
				}

				// 验证方法签名没有改变
				switch methodName {
				case "ChunkDocument":
					// func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error)
					if method.Type.NumIn() != 2 || method.Type.NumOut() != 2 {
						t.Errorf("方法 %s 签名改变", methodName)
					}
				case "GetErrors":
					// func (c *MarkdownChunker) GetErrors() []error
					if method.Type.NumIn() != 1 || method.Type.NumOut() != 1 {
						t.Errorf("方法 %s 签名改变", methodName)
					}
				case "HasErrors":
					// func (c *MarkdownChunker) HasErrors() bool
					if method.Type.NumIn() != 1 || method.Type.NumOut() != 1 {
						t.Errorf("方法 %s 签名改变", methodName)
					}
				case "ClearErrors":
					// func (c *MarkdownChunker) ClearErrors()
					if method.Type.NumIn() != 1 || method.Type.NumOut() != 0 {
						t.Errorf("方法 %s 签名改变", methodName)
					}
				}
			}

			// 测试方法功能
			content := []byte(`# Test

Content here.`)

			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Errorf("ChunkDocument 方法失败: %v", err)
			}
			if len(chunks) == 0 {
				t.Error("ChunkDocument 应该返回块")
			}

			// 测试错误处理方法
			errors := chunker.GetErrors()
			if errors == nil {
				t.Error("GetErrors 应该返回非nil切片")
			}

			hasErrors := chunker.HasErrors()
			if hasErrors != (len(errors) > 0) {
				t.Error("HasErrors 结果与 GetErrors 不一致")
			}

			chunker.ClearErrors()
			if chunker.HasErrors() {
				t.Error("ClearErrors 后不应该有错误")
			}
		})
	})

	t.Run("现有配置自动迁移", func(t *testing.T) {
		t.Run("空配置迁移", func(t *testing.T) {
			// 空配置应该自动设置默认策略
			config := &ChunkerConfig{}
			chunker := NewMarkdownChunkerWithConfig(config)

			strategyName, _ := chunker.GetCurrentStrategy()
			if strategyName != "element-level" {
				t.Errorf("空配置应该使用默认策略，实际是 %s", strategyName)
			}

			// 验证配置被正确迁移
			if config.ChunkingStrategy == nil {
				t.Error("配置迁移后应该包含策略配置")
			}

			if config.ChunkingStrategy.Name != "element-level" {
				t.Errorf("迁移后策略名称错误，期望 element-level，实际 %s",
					config.ChunkingStrategy.Name)
			}
		})

		t.Run("现有配置字段保持有效", func(t *testing.T) {
			// 测试现有配置字段仍然有效
			config := &ChunkerConfig{
				MaxChunkSize:       2000,
				FilterEmptyChunks:  true,
				PreserveWhitespace: true,
				EnabledTypes: map[string]bool{
					"heading":   true,
					"paragraph": true,
				},
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeDefault,
				EnableObjectPooling: true,
				MemoryLimit:         1024 * 1024, // 1MB
			}

			chunker := NewMarkdownChunkerWithConfig(config)

			// 验证配置值被正确保留
			if chunker.config.MaxChunkSize != 2000 {
				t.Errorf("MaxChunkSize 配置丢失，期望 2000，实际 %d", chunker.config.MaxChunkSize)
			}

			if !chunker.config.FilterEmptyChunks {
				t.Error("FilterEmptyChunks 配置丢失")
			}

			if !chunker.config.PreserveWhitespace {
				t.Error("PreserveWhitespace 配置丢失")
			}

			if chunker.config.ErrorHandling != ErrorModePermissive {
				t.Errorf("ErrorHandling 配置丢失，期望 %v，实际 %v",
					ErrorModePermissive, chunker.config.ErrorHandling)
			}

			if chunker.config.PerformanceMode != PerformanceModeDefault {
				t.Errorf("PerformanceMode 配置丢失，期望 %v，实际 %v",
					PerformanceModeDefault, chunker.config.PerformanceMode)
			}

			if !chunker.config.EnableObjectPooling {
				t.Error("EnableObjectPooling 配置丢失")
			}

			if chunker.config.MemoryLimit != 1024*1024 {
				t.Errorf("MemoryLimit 配置丢失，期望 %d，实际 %d",
					1024*1024, chunker.config.MemoryLimit)
			}

			// 验证类型过滤仍然有效
			content := []byte(`# Heading

Paragraph content.

` + "```go\ncode here\n```")

			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Fatalf("分块失败: %v", err)
			}

			// 验证允许的类型存在
			foundHeading := false
			foundParagraph := false
			for _, chunk := range chunks {
				if chunk.Type == "heading" {
					foundHeading = true
				}
				if chunk.Type == "paragraph" {
					foundParagraph = true
				}
			}

			if !foundHeading {
				t.Error("应该包含 heading 类型")
			}
			if !foundParagraph {
				t.Error("应该包含 paragraph 类型")
			}
		})

		t.Run("配置迁移函数", func(t *testing.T) {
			// 测试显式配置迁移函数
			oldConfig := &ChunkerConfig{
				MaxChunkSize: 1500,
				EnabledTypes: map[string]bool{
					"heading": true,
					"list":    true,
				},
			}

			migratedResult, err := MigrateConfig(oldConfig)
			if err != nil {
				t.Fatalf("配置迁移失败: %v", err)
			}
			migratedConfig := migratedResult.Config

			// 验证旧配置被保留
			if migratedConfig.MaxChunkSize != 1500 {
				t.Errorf("迁移后 MaxChunkSize 错误，期望 1500，实际 %d",
					migratedConfig.MaxChunkSize)
			}

			if !reflect.DeepEqual(migratedConfig.EnabledTypes, oldConfig.EnabledTypes) {
				t.Error("迁移后 EnabledTypes 不匹配")
			}

			// 验证策略配置被添加
			if migratedConfig.ChunkingStrategy == nil {
				t.Error("迁移后应该包含策略配置")
			}

			if migratedConfig.ChunkingStrategy.Name != "element-level" {
				t.Errorf("迁移后默认策略错误，期望 element-level，实际 %s",
					migratedConfig.ChunkingStrategy.Name)
			}

			// 验证原配置没有被修改
			if oldConfig.ChunkingStrategy != nil {
				t.Error("原配置不应该被修改")
			}
		})
	})

	t.Run("现有测试用例全部通过", func(t *testing.T) {
		// 这个测试确保现有的测试用例在新系统中仍然通过
		// 我们重新运行一些关键的现有测试逻辑

		t.Run("基本分块功能", func(t *testing.T) {
			chunker := NewMarkdownChunker()

			testCases := []struct {
				name     string
				content  string
				expected int
			}{
				{
					name:     "简单标题和段落",
					content:  "# Title\n\nParagraph content.",
					expected: 2,
				},
				{
					name:     "多级标题",
					content:  "# H1\n\n## H2\n\n### H3\n\nContent.",
					expected: 4,
				},
				{
					name:     "列表处理",
					content:  "# Title\n\n- Item 1\n- Item 2\n- Item 3",
					expected: 2,
				},
				{
					name:     "代码块处理",
					content:  "# Title\n\n```go\nfunc main() {}\n```",
					expected: 2,
				},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					chunks, err := chunker.ChunkDocument([]byte(tc.content))
					if err != nil {
						t.Errorf("分块失败: %v", err)
					}

					if len(chunks) != tc.expected {
						t.Errorf("块数量不匹配，期望 %d，实际 %d", tc.expected, len(chunks))
					}

					// 验证块的基本属性
					for i, chunk := range chunks {
						if chunk.ID != i {
							t.Errorf("块 %d ID错误", i)
						}
						if chunk.Content == "" {
							t.Errorf("块 %d 内容为空", i)
						}
						if chunk.Type == "" {
							t.Errorf("块 %d 类型为空", i)
						}
					}
				})
			}
		})

		t.Run("错误处理兼容性", func(t *testing.T) {
			chunker := NewMarkdownChunker()

			// 测试无效输入处理
			chunks, err := chunker.ChunkDocument(nil)
			if err == nil {
				t.Error("nil输入应该返回错误")
			}

			// 验证错误处理方法
			if !chunker.HasErrors() {
				t.Error("应该有错误记录")
			}

			errors := chunker.GetErrors()
			if len(errors) == 0 {
				t.Error("应该有错误记录")
			}

			chunker.ClearErrors()
			if chunker.HasErrors() {
				t.Error("清除错误后不应该有错误")
			}

			// 测试空内容处理
			chunks, err = chunker.ChunkDocument([]byte(""))
			if err != nil {
				t.Errorf("空内容处理失败: %v", err)
			}

			// 空内容可能产生0个或1个空块
			if len(chunks) > 1 {
				t.Errorf("空内容不应该产生多个块，实际 %d 个", len(chunks))
			}
		})

		t.Run("配置处理兼容性", func(t *testing.T) {
			// 测试各种配置组合
			configs := []*ChunkerConfig{
				{
					MaxChunkSize: 500,
				},
				{
					EnabledTypes: map[string]bool{
						"heading": true,
					},
				},
				{
					FilterEmptyChunks: true,
				},
				{
					PreserveWhitespace: true,
				},
			}

			content := []byte(`# Test

Paragraph content.

` + "```go\ncode\n```")

			for i, config := range configs {
				t.Run(fmt.Sprintf("配置_%d", i), func(t *testing.T) {
					chunker := NewMarkdownChunkerWithConfig(config)

					chunks, err := chunker.ChunkDocument(content)
					if err != nil {
						t.Errorf("配置 %d 分块失败: %v", i, err)
					}

					// 基本验证
					if len(chunks) == 0 {
						t.Errorf("配置 %d 应该产生至少一个块", i)
					}

					for j, chunk := range chunks {
						if chunk.Content == "" {
							t.Errorf("配置 %d 块 %d 内容为空", i, j)
						}
					}
				})
			}
		})
	})

	t.Run("性能兼容性", func(t *testing.T) {
		// 确保新系统的性能不低于旧系统
		chunker := NewMarkdownChunker()

		// 生成测试内容
		content := []byte(`# Performance Test

This is a performance test document.

## Section 1

Content for section 1.

### Subsection 1.1

More detailed content here.

## Section 2

Content for section 2.

- List item 1
- List item 2
- List item 3

` + "```go\nfunc test() {\n    fmt.Println(\"test\")\n}\n```")

		// 预热
		for i := 0; i < 5; i++ {
			_, _ = chunker.ChunkDocument(content)
		}

		// 性能测试
		const iterations = 100
		start := time.Now()

		for i := 0; i < iterations; i++ {
			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Errorf("第 %d 次分块失败: %v", i, err)
			}
			if len(chunks) == 0 {
				t.Errorf("第 %d 次分块没有产生块", i)
			}
		}

		duration := time.Since(start)
		avgDuration := duration / iterations

		t.Logf("平均分块时间: %v", avgDuration)

		// 性能阈值（应该根据实际基准调整）
		if avgDuration > time.Millisecond*10 {
			t.Errorf("性能可能有回归，平均时间: %v", avgDuration)
		}
	})

	t.Run("数据结构兼容性", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		content := []byte(`# Test

Content here.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		if len(chunks) == 0 {
			t.Fatal("应该产生至少一个块")
		}

		// 验证Chunk结构体的所有字段仍然存在且类型正确
		chunk := chunks[0]

		// 验证基本字段
		if chunk.ID < 0 {
			t.Error("ID 字段类型或值错误")
		}

		if chunk.Type == "" {
			t.Error("Type 字段为空")
		}

		if chunk.Content == "" {
			t.Error("Content 字段为空")
		}

		if chunk.Text == "" {
			t.Error("Text 字段为空")
		}

		if chunk.Level < 0 {
			t.Error("Level 字段值错误")
		}

		// 验证Metadata字段
		if chunk.Metadata == nil {
			t.Error("Metadata 字段不应该为nil")
		}

		// 验证关键的元数据字段存在
		expectedMetadataKeys := []string{"line_start", "line_end", "char_start", "char_end"}
		for _, key := range expectedMetadataKeys {
			if _, exists := chunk.Metadata[key]; !exists {
				t.Errorf("元数据缺少关键字段: %s", key)
			}
		}

		// 验证新增的策略元数据不影响现有功能
		if strategy, exists := chunk.Metadata["strategy"]; exists {
			if strategy != "element-level" {
				t.Errorf("默认策略元数据错误，期望 element-level，实际 %s", strategy)
			}
		}
	})
}
