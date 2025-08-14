package markdownchunker

import (
	"testing"
)

// TestStrategyIntegration 测试策略系统集成
func TestStrategyIntegration(t *testing.T) {
	t.Run("默认策略行为", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 验证默认策略
		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "element-level" {
			t.Errorf("默认策略应该是 element-level，实际是 %s", strategyName)
		}

		content := []byte(`# Heading 1

This is a paragraph.

## Heading 2

Another paragraph.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		if len(chunks) != 4 {
			t.Errorf("应该生成4个块，实际生成 %d 个", len(chunks))
		}

		expectedTypes := []string{"heading", "paragraph", "heading", "paragraph"}
		for i, chunk := range chunks {
			if i < len(expectedTypes) && chunk.Type != expectedTypes[i] {
				t.Errorf("块 %d 类型错误，期望 '%s'，实际 '%s'", i, expectedTypes[i], chunk.Type)
			}
		}
	})

	t.Run("策略切换", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 验证可用策略
		strategies := chunker.GetAvailableStrategies()
		if len(strategies) == 0 {
			t.Error("应该有可用的策略")
		}

		found := false
		for _, strategy := range strategies {
			if strategy == "element-level" {
				found = true
				break
			}
		}
		if !found {
			t.Error("应该包含 element-level 策略")
		}

		// 测试切换到相同策略
		err := chunker.SetStrategy("element-level", nil)
		if err != nil {
			t.Errorf("切换到相同策略失败: %v", err)
		}

		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "element-level" {
			t.Error("策略切换后应该仍然是 element-level")
		}
	})

	t.Run("策略配置", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 使用配置切换策略
		config := &StrategyConfig{
			Name:         "element-level",
			IncludeTypes: []string{"heading"},
		}

		err := chunker.SetStrategy("element-level", config)
		if err != nil {
			t.Errorf("使用配置切换策略失败: %v", err)
		}

		content := []byte(`# Heading 1

This is a paragraph.

## Heading 2

Another paragraph.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		// 应该只包含标题块
		for _, chunk := range chunks {
			if chunk.Type != "heading" {
				t.Errorf("不应该包含 %s 类型的块", chunk.Type)
			}
		}

		if len(chunks) != 2 {
			t.Errorf("应该生成2个标题块，实际生成 %d 个", len(chunks))
		}
	})

	t.Run("无效策略处理", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 尝试切换到不存在的策略
		err := chunker.SetStrategy("non-existent", nil)
		if err == nil {
			t.Error("切换到不存在的策略应该返回错误")
		}

		// 验证策略没有改变
		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "element-level" {
			t.Error("策略不应该改变")
		}
	})

	t.Run("无效配置处理", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 使用无效配置
		invalidConfig := &StrategyConfig{
			Name: "", // 空名称
		}

		err := chunker.SetStrategy("element-level", invalidConfig)
		if err == nil {
			t.Error("使用无效配置应该返回错误")
		}
	})

	t.Run("自定义策略注册", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 注册自定义策略
		customStrategy := &MockStrategy{
			name:        "custom-test",
			description: "自定义测试策略",
			shouldError: false,
		}

		err := chunker.RegisterStrategy(customStrategy)
		if err != nil {
			t.Errorf("注册自定义策略失败: %v", err)
		}

		// 验证策略已注册
		strategies := chunker.GetAvailableStrategies()
		found := false
		for _, strategy := range strategies {
			if strategy == "custom-test" {
				found = true
				break
			}
		}
		if !found {
			t.Error("自定义策略应该在可用策略列表中")
		}

		// 切换到自定义策略
		err = chunker.SetStrategy("custom-test", nil)
		if err != nil {
			t.Errorf("切换到自定义策略失败: %v", err)
		}

		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "custom-test" {
			t.Error("当前策略应该是 custom-test")
		}
	})
}

// TestStrategyBackwardCompatibility 测试向后兼容性
func TestStrategyBackwardCompatibility(t *testing.T) {
	t.Run("默认行为保持不变", func(t *testing.T) {
		// 使用默认构造函数
		chunker := NewMarkdownChunker()

		content := []byte(`# Test

Content here.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		// 验证结果与之前版本一致
		if len(chunks) != 2 {
			t.Errorf("应该生成2个块，实际生成 %d 个", len(chunks))
		}

		if chunks[0].Type != "heading" {
			t.Errorf("第一个块应该是标题，实际是 %s", chunks[0].Type)
		}

		if chunks[1].Type != "paragraph" {
			t.Errorf("第二个块应该是段落，实际是 %s", chunks[1].Type)
		}
	})

	t.Run("配置构造函数兼容性", func(t *testing.T) {
		// 使用配置构造函数，不指定策略
		config := &ChunkerConfig{
			MaxChunkSize: 1000,
			EnabledTypes: map[string]bool{
				"heading":   true,
				"paragraph": true,
			},
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		// 应该使用默认策略
		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "element-level" {
			t.Errorf("应该使用默认策略，实际是 %s", strategyName)
		}

		content := []byte(`# Test

Content here.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("分块失败: %v", err)
		}

		// 验证配置仍然生效
		if len(chunks) != 2 {
			t.Errorf("应该生成2个块，实际生成 %d 个", len(chunks))
		}
	})

	t.Run("现有API方法保持不变", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 验证现有方法仍然可用
		_ = chunker.GetErrors()
		_ = chunker.HasErrors()
		chunker.ClearErrors()

		// 这些方法应该不会引起panic或错误
	})
}

// TestStrategyPerformance 测试策略性能
func TestStrategyPerformance(t *testing.T) {
	t.Run("策略切换性能", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 多次切换策略不应该有明显性能问题
		for i := 0; i < 100; i++ {
			err := chunker.SetStrategy("element-level", nil)
			if err != nil {
				t.Errorf("第 %d 次策略切换失败: %v", i, err)
			}
		}
	})

	t.Run("策略执行性能", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		content := []byte(`# Heading 1

This is a paragraph.

## Heading 2

Another paragraph.

### Heading 3

More content here.`)

		// 多次执行分块不应该有明显性能问题
		for i := 0; i < 100; i++ {
			_, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Errorf("第 %d 次分块失败: %v", i, err)
			}
		}
	})
}

// TestHierarchicalStrategyIntegration 测试层级策略集成
func TestHierarchicalStrategyIntegration(t *testing.T) {
	t.Run("层级策略可用性", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		// 验证层级策略已注册
		strategies := chunker.GetAvailableStrategies()
		found := false
		for _, strategy := range strategies {
			if strategy == "hierarchical" {
				found = true
				break
			}
		}
		if !found {
			t.Error("层级策略应该在可用策略列表中")
		}
	})

	t.Run("切换到层级策略", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		config := HierarchicalConfig(2)
		err := chunker.SetStrategy("hierarchical", config)
		if err != nil {
			t.Errorf("切换到层级策略失败: %v", err)
		}

		strategyName, _ := chunker.GetCurrentStrategy()
		if strategyName != "hierarchical" {
			t.Error("当前策略应该是 hierarchical")
		}
	})

	t.Run("层级策略基本分块", func(t *testing.T) {
		chunker := NewMarkdownChunker()

		config := HierarchicalConfig(2)
		err := chunker.SetStrategy("hierarchical", config)
		if err != nil {
			t.Fatalf("切换到层级策略失败: %v", err)
		}

		content := []byte(`# Chapter 1

Introduction to the chapter.

## Section 1.1

Content of section 1.1.

### Subsection 1.1.1

Detailed content.

## Section 1.2

Content of section 1.2.

# Chapter 2

Second chapter content.`)

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("层级分块失败: %v", err)
		}

		// 验证生成的块数量合理
		if len(chunks) == 0 {
			t.Error("应该生成至少一个块")
		}

		// 验证块包含层级信息
		for i, chunk := range chunks {
			if chunk.Metadata["strategy"] != "hierarchical" {
				t.Errorf("块 %d 应该标记为层级策略生成", i)
			}

			if chunk.Metadata["hierarchy_level"] == "" {
				t.Errorf("块 %d 应该包含层级信息", i)
			}
		}
	})
}
