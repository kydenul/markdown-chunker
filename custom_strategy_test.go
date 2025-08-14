package markdownchunker

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// TestCustomStrategyBuilder 测试自定义策略构建器
func TestCustomStrategyBuilder(t *testing.T) {
	t.Run("创建基本构建器", func(t *testing.T) {
		builder := NewCustomStrategyBuilder("test-strategy", "测试策略")

		if builder.Name != "test-strategy" {
			t.Errorf("期望策略名称为 'test-strategy'，实际为 '%s'", builder.Name)
		}

		if builder.Description != "测试策略" {
			t.Errorf("期望策略描述为 '测试策略'，实际为 '%s'", builder.Description)
		}

		if builder.GetRuleCount() != 0 {
			t.Errorf("期望规则数量为 0，实际为 %d", builder.GetRuleCount())
		}
	})

	t.Run("添加和管理规则", func(t *testing.T) {
		builder := NewCustomStrategyBuilder("test-strategy", "测试策略")

		// 添加规则
		condition := NewHeadingLevelCondition(1, 3)
		action := NewCreateSeparateChunkAction("", nil)
		builder.AddRule("heading-rule", "标题规则", condition, action, 100)

		if builder.GetRuleCount() != 1 {
			t.Errorf("期望规则数量为 1，实际为 %d", builder.GetRuleCount())
		}

		// 检查规则是否存在
		if !builder.HasRule("heading-rule") {
			t.Error("期望规则 'heading-rule' 存在")
		}

		// 获取规则
		rule := builder.GetRule("heading-rule")
		if rule == nil {
			t.Error("期望能够获取规则 'heading-rule'")
		} else {
			if rule.Name != "heading-rule" {
				t.Errorf("期望规则名称为 'heading-rule'，实际为 '%s'", rule.Name)
			}
			if rule.Priority != 100 {
				t.Errorf("期望规则优先级为 100，实际为 %d", rule.Priority)
			}
		}

		// 移除规则
		builder.RemoveRule("heading-rule")
		if builder.GetRuleCount() != 0 {
			t.Errorf("期望规则数量为 0，实际为 %d", builder.GetRuleCount())
		}
	})

	t.Run("规则优先级排序", func(t *testing.T) {
		builder := NewCustomStrategyBuilder("test-strategy", "测试策略")

		// 添加不同优先级的规则
		builder.AddRule("low-priority", "低优先级", NewHeadingLevelCondition(1, 1), NewSkipNodeAction(""), 10)
		builder.AddRule("high-priority", "高优先级", NewHeadingLevelCondition(2, 2), NewSkipNodeAction(""), 100)
		builder.AddRule("medium-priority", "中优先级", NewHeadingLevelCondition(3, 3), NewSkipNodeAction(""), 50)

		rules := builder.GetRules()
		if len(rules) != 3 {
			t.Errorf("期望规则数量为 3，实际为 %d", len(rules))
		}

		// 检查排序是否正确（优先级高的在前）
		if rules[0].Name != "high-priority" {
			t.Errorf("期望第一个规则为 'high-priority'，实际为 '%s'", rules[0].Name)
		}
		if rules[1].Name != "medium-priority" {
			t.Errorf("期望第二个规则为 'medium-priority'，实际为 '%s'", rules[1].Name)
		}
		if rules[2].Name != "low-priority" {
			t.Errorf("期望第三个规则为 'low-priority'，实际为 '%s'", rules[2].Name)
		}
	})

	t.Run("构建器验证", func(t *testing.T) {
		// 测试空名称验证
		builder := NewCustomStrategyBuilder("", "测试策略")
		if err := builder.Validate(); err == nil {
			t.Error("期望空名称验证失败")
		}

		// 测试无规则验证
		builder = NewCustomStrategyBuilder("test-strategy", "测试策略")
		if err := builder.Validate(); err == nil {
			t.Error("期望无规则验证失败")
		}

		// 测试有效配置
		builder.AddRule("test-rule", "测试规则",
			NewHeadingLevelCondition(1, 3),
			NewCreateSeparateChunkAction("", nil),
			100)
		if err := builder.Validate(); err != nil {
			t.Errorf("期望有效配置验证通过，实际错误：%v", err)
		}
	})

	t.Run("构建策略", func(t *testing.T) {
		builder := NewCustomStrategyBuilder("test-strategy", "测试策略")
		builder.AddRule("heading-rule", "标题规则",
			NewHeadingLevelCondition(1, 3),
			NewCreateSeparateChunkAction("", nil),
			100)

		strategy, err := builder.Build()
		if err != nil {
			t.Errorf("期望构建成功，实际错误：%v", err)
		}

		if strategy == nil {
			t.Error("期望策略不为空")
		}

		if strategy.GetName() != "test-strategy" {
			t.Errorf("期望策略名称为 'test-strategy'，实际为 '%s'", strategy.GetName())
		}

		customStrategy, ok := strategy.(*CustomStrategy)
		if !ok {
			t.Error("期望策略类型为 CustomStrategy")
		} else {
			if customStrategy.GetRuleCount() != 1 {
				t.Errorf("期望规则数量为 1，实际为 %d", customStrategy.GetRuleCount())
			}
		}
	})
}

// TestRuleConditions 测试规则条件
func TestRuleConditions(t *testing.T) {
	// 创建测试用的 Markdown 解析器
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	t.Run("HeadingLevelCondition", func(t *testing.T) {
		condition := NewHeadingLevelCondition(1, 3)

		// 测试验证
		if err := condition.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试无效配置
		invalidCondition := NewHeadingLevelCondition(0, 7)
		if err := invalidCondition.Validate(); err == nil {
			t.Error("期望无效配置验证失败")
		}

		// 测试匹配
		content := []byte("# 标题1\n## 标题2\n#### 标题4\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(nil, content)

		// 遍历节点测试匹配
		matchCount := 0
		for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
			if condition.Match(child, context) {
				matchCount++
			}
		}

		// 应该匹配 H1 和 H2，不匹配 H4
		if matchCount != 2 {
			t.Errorf("期望匹配 2 个标题，实际匹配 %d 个", matchCount)
		}
	})

	t.Run("ContentTypeCondition", func(t *testing.T) {
		condition := NewContentTypeCondition("heading", "paragraph")

		// 测试验证
		if err := condition.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试无效类型
		invalidCondition := NewContentTypeCondition("invalid-type")
		if err := invalidCondition.Validate(); err == nil {
			t.Error("期望无效类型验证失败")
		}

		// 测试匹配
		content := []byte("# 标题\n\n这是段落\n\n```\n代码块\n```\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(nil, content)

		matchCount := 0
		for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
			if condition.Match(child, context) {
				matchCount++
			}
		}

		// 应该匹配标题和段落，不匹配代码块
		if matchCount != 2 {
			t.Errorf("期望匹配 2 个节点，实际匹配 %d 个", matchCount)
		}
	})

	t.Run("ContentSizeCondition", func(t *testing.T) {
		condition := NewContentSizeCondition(10, 100)

		// 测试验证
		if err := condition.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试无效配置
		invalidCondition := NewContentSizeCondition(-1, 10)
		if err := invalidCondition.Validate(); err == nil {
			t.Error("期望无效配置验证失败")
		}

		// 测试匹配
		content := []byte("# 短标题\n\n这是一个中等长度的段落，包含足够的文字来测试大小条件。\n\n很短\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(nil, content)

		matchCount := 0
		for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
			if condition.Match(child, context) {
				matchCount++
			}
		}

		// 应该匹配中等长度的段落
		if matchCount < 1 {
			// 调试信息
			for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
				var buf bytes.Buffer
				extractNodeText(child, content, &buf)
				t.Logf("节点类型: %s, 内容长度: %d, 内容: %s", getNodeType(child), buf.Len(), buf.String())
			}
			t.Errorf("期望至少匹配 1 个节点，实际匹配 %d 个", matchCount)
		}
	})

	t.Run("DepthCondition", func(t *testing.T) {
		condition := NewDepthCondition(1, 3)

		// 测试验证
		if err := condition.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试匹配
		content := []byte("# Test")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)
		node := doc.FirstChild()
		context := NewChunkingContext(nil, content)

		// 测试不同深度
		context.Depth = 0
		if condition.Match(node, context) {
			t.Error("期望深度 0 不匹配")
		}

		context.Depth = 2
		if !condition.Match(node, context) {
			t.Errorf("期望深度 2 匹配，条件: min=%d, max=%d", condition.MinDepth, condition.MaxDepth)
		}

		context.Depth = 4
		if condition.Match(node, context) {
			t.Error("期望深度 4 不匹配")
		}
	})
}

// TestRuleActions 测试规则动作
func TestRuleActions(t *testing.T) {
	// 创建测试用的分块器
	chunker := NewMarkdownChunker()

	// 创建测试用的 Markdown 解析器
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	t.Run("CreateSeparateChunkAction", func(t *testing.T) {
		action := NewCreateSeparateChunkAction("custom-type", map[string]string{
			"custom": "true",
		})

		// 测试验证
		if err := action.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试执行
		content := []byte("# 测试标题\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(chunker, content)

		node := doc.FirstChild()
		chunk, err := action.Execute(node, context)
		if err != nil {
			t.Errorf("期望执行成功，实际错误：%v", err)
		}

		if chunk == nil {
			t.Error("期望创建块不为空")
		} else {
			if chunk.Type != "custom-type" {
				t.Errorf("期望块类型为 'custom-type'，实际为 '%s'", chunk.Type)
			}

			if chunk.Metadata["custom"] != "true" {
				t.Error("期望自定义元数据存在")
			}

			if chunk.Metadata["action"] != "create-separate-chunk" {
				t.Error("期望动作标识存在")
			}
		}
	})

	t.Run("MergeWithParentAction", func(t *testing.T) {
		action := NewMergeWithParentAction("\n\n")

		// 测试验证
		if err := action.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试执行（无父块）
		content := []byte("这是段落内容\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(chunker, content)

		node := doc.FirstChild()
		chunk, err := action.Execute(node, context)
		if err != nil {
			t.Errorf("期望执行成功，实际错误：%v", err)
		}

		if chunk == nil {
			t.Error("期望创建新块不为空")
		} else {
			if chunk.Metadata["fallback_to_new_chunk"] != "true" {
				t.Error("期望回退标识存在")
			}
		}

		// 测试执行（有父块）
		parentChunk := &Chunk{
			ID:       0,
			Type:     "heading",
			Content:  "# 父标题",
			Text:     "父标题",
			Metadata: make(map[string]string),
		}
		context.PreviousChunk = parentChunk

		mergedChunk, err := action.Execute(node, context)
		if err != nil {
			t.Errorf("期望合并执行成功，实际错误：%v", err)
		}

		if mergedChunk == nil {
			t.Error("期望合并块不为空")
		} else {
			if !strings.Contains(mergedChunk.Content, "父标题") {
				t.Error("期望合并块包含父内容")
			}

			if mergedChunk.Metadata["merged"] != "true" {
				t.Error("期望合并标识存在")
			}
		}
	})

	t.Run("SkipNodeAction", func(t *testing.T) {
		action := NewSkipNodeAction("测试跳过")

		// 测试验证
		if err := action.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		// 测试执行
		content := []byte("# 测试标题\n")
		reader := text.NewReader(content)
		doc := md.Parser().Parse(reader)

		context := NewChunkingContext(chunker, content)

		node := doc.FirstChild()
		chunk, err := action.Execute(node, context)
		if err != nil {
			t.Errorf("期望执行成功，实际错误：%v", err)
		}

		if chunk != nil {
			t.Error("期望跳过动作返回空块")
		}
	})
}

// TestCustomStrategy 测试自定义策略
func TestCustomStrategy(t *testing.T) {
	t.Run("基本功能测试", func(t *testing.T) {
		// 创建自定义策略
		builder := NewCustomStrategyBuilder("test-strategy", "测试策略")

		// 添加标题处理规则
		builder.AddRule("heading-rule", "标题规则",
			NewHeadingLevelCondition(1, 2),
			NewCreateSeparateChunkAction("", map[string]string{
				"rule_type": "heading",
			}),
			100)

		// 添加段落合并规则
		builder.AddRule("paragraph-rule", "段落规则",
			NewContentTypeCondition("paragraph"),
			NewMergeWithParentAction("\n\n"),
			50)

		strategy, err := builder.Build()
		if err != nil {
			t.Fatalf("构建策略失败：%v", err)
		}

		// 创建分块器
		chunker := NewMarkdownChunker()

		// 测试内容
		content := []byte(`# 主标题

这是第一段内容。

## 子标题

这是第二段内容。

### 三级标题

这是第三段内容。`)

		// 解析文档
		reader := text.NewReader(content)
		doc := chunker.md.Parser().Parse(reader)

		// 使用自定义策略分块
		chunks, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			t.Fatalf("分块失败：%v", err)
		}

		if len(chunks) == 0 {
			t.Error("期望生成块，实际为空")
		}

		// 检查块的元数据
		hasHeadingRule := false
		for _, chunk := range chunks {
			if chunk.Metadata["rule_type"] == "heading" {
				hasHeadingRule = true
				break
			}
		}

		if !hasHeadingRule {
			t.Error("期望找到标题规则处理的块")
		}
	})

	t.Run("规则优先级测试", func(t *testing.T) {
		// 创建有冲突规则的策略
		builder := NewCustomStrategyBuilder("priority-test", "优先级测试策略")

		// 低优先级：跳过所有标题
		builder.AddRule("skip-all-headings", "跳过所有标题",
			NewHeadingLevelCondition(1, 6),
			NewSkipNodeAction("低优先级跳过"),
			10)

		// 高优先级：处理 H1 标题
		builder.AddRule("process-h1", "处理H1标题",
			NewHeadingLevelCondition(1, 1),
			NewCreateSeparateChunkAction("h1-special", nil),
			100)

		strategy, err := builder.Build()
		if err != nil {
			t.Fatalf("构建策略失败：%v", err)
		}

		// 创建分块器
		chunker := NewMarkdownChunker()

		// 测试内容
		content := []byte(`# H1标题
## H2标题`)

		// 解析文档
		reader := text.NewReader(content)
		doc := chunker.md.Parser().Parse(reader)

		// 使用自定义策略分块
		chunks, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			t.Fatalf("分块失败：%v", err)
		}

		// 应该只有 H1 被处理，H2 被跳过
		h1Found := false
		h2Found := false

		for _, chunk := range chunks {
			if chunk.Type == "h1-special" {
				h1Found = true
			}
			if strings.Contains(chunk.Content, "H2标题") {
				h2Found = true
			}
		}

		if !h1Found {
			t.Error("期望找到 H1 标题块")
		}

		if h2Found {
			t.Error("期望 H2 标题被跳过")
		}
	})
}

// TestPrebuiltStrategyBuilders 测试预构建的策略构建器
func TestPrebuiltStrategyBuilders(t *testing.T) {
	t.Run("NewHeadingBasedStrategyBuilder", func(t *testing.T) {
		builder := NewHeadingBasedStrategyBuilder("heading-strategy", 3)

		if builder.GetRuleCount() != 2 {
			t.Errorf("期望规则数量为 2，实际为 %d", builder.GetRuleCount())
		}

		if err := builder.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		strategy, err := builder.Build()
		if err != nil {
			t.Errorf("期望构建成功，实际错误：%v", err)
		}

		if strategy.GetName() != "heading-strategy" {
			t.Errorf("期望策略名称为 'heading-strategy'，实际为 '%s'", strategy.GetName())
		}
	})

	t.Run("NewContentTypeBasedStrategyBuilder", func(t *testing.T) {
		separateTypes := []string{"heading", "code"}
		mergeTypes := []string{"paragraph", "list"}

		builder := NewContentTypeBasedStrategyBuilder("content-type-strategy", separateTypes, mergeTypes)

		if builder.GetRuleCount() != 2 {
			t.Errorf("期望规则数量为 2，实际为 %d", builder.GetRuleCount())
		}

		if err := builder.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		strategy, err := builder.Build()
		if err != nil {
			t.Errorf("期望构建成功，实际错误：%v", err)
		}

		if strategy.GetName() != "content-type-strategy" {
			t.Errorf("期望策略名称为 'content-type-strategy'，实际为 '%s'", strategy.GetName())
		}
	})

	t.Run("NewSizeBasedStrategyBuilder", func(t *testing.T) {
		builder := NewSizeBasedStrategyBuilder("size-strategy", 50, 200)

		if builder.GetRuleCount() != 3 {
			t.Errorf("期望规则数量为 3，实际为 %d", builder.GetRuleCount())
		}

		if err := builder.Validate(); err != nil {
			t.Errorf("期望验证通过，实际错误：%v", err)
		}

		strategy, err := builder.Build()
		if err != nil {
			t.Errorf("期望构建成功，实际错误：%v", err)
		}

		if strategy.GetName() != "size-strategy" {
			t.Errorf("期望策略名称为 'size-strategy'，实际为 '%s'", strategy.GetName())
		}
	})
}

// TestComplexCustomStrategy 测试复杂的自定义策略
func TestComplexCustomStrategy(t *testing.T) {
	// 创建复杂的自定义策略
	builder := NewCustomStrategyBuilder("complex-strategy", "复杂的自定义策略")

	// 规则1：H1标题独立处理
	builder.AddRule("h1-separate", "H1标题独立处理",
		NewHeadingLevelCondition(1, 1),
		NewCreateSeparateChunkAction("main-heading", map[string]string{
			"importance": "high",
		}),
		100)

	// 规则2：H2-H3标题独立处理
	builder.AddRule("h2-h3-separate", "H2-H3标题独立处理",
		NewHeadingLevelCondition(2, 3),
		NewCreateSeparateChunkAction("sub-heading", map[string]string{
			"importance": "medium",
		}),
		90)

	// 规则3：大段落独立处理
	builder.AddRule("large-paragraph", "大段落独立处理",
		NewContentSizeCondition(100, 0),
		NewCreateSeparateChunkAction("large-content", map[string]string{
			"size": "large",
		}),
		80)

	// 规则4：小段落合并
	builder.AddRule("small-paragraph-merge", "小段落合并",
		NewContentSizeCondition(0, 50),
		NewMergeWithParentAction(" "),
		70)

	// 规则5：代码块独立处理
	builder.AddRule("code-separate", "代码块独立处理",
		NewContentTypeCondition("code"),
		NewCreateSeparateChunkAction("code-block", map[string]string{
			"type": "code",
		}),
		60)

	strategy, err := builder.Build()
	if err != nil {
		t.Fatalf("构建复杂策略失败：%v", err)
	}

	// 创建分块器
	chunker := NewMarkdownChunker()

	// 复杂测试内容
	content := []byte(`# 主标题

这是一个很长的段落，包含足够多的文字来触发大段落规则。这个段落应该被独立处理，因为它的长度超过了100个字符的阈值。

## 子标题

短段落。

### 三级标题

另一个短段落。

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

这是另一个很长的段落，同样包含足够多的文字来触发大段落规则。这个段落也应该被独立处理。`)

	// 解析文档
	reader := text.NewReader(content)
	doc := chunker.md.Parser().Parse(reader)

	// 使用复杂策略分块
	chunks, err := strategy.ChunkDocument(doc, content, chunker)
	if err != nil {
		t.Fatalf("复杂策略分块失败：%v", err)
	}

	if len(chunks) == 0 {
		t.Error("期望生成块，实际为空")
	}

	// 验证不同类型的块
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.Type]++
	}

	// 应该有主标题
	if typeCount["main-heading"] == 0 {
		t.Error("期望找到主标题块")
	}

	// 应该有子标题
	if typeCount["sub-heading"] == 0 {
		t.Error("期望找到子标题块")
	}

	// 应该有大内容块
	if typeCount["large-content"] == 0 {
		t.Error("期望找到大内容块")
	}

	// 打印调试信息
	t.Logf("生成了 %d 个块", len(chunks))
	for i, chunk := range chunks {
		t.Logf("块 %d: 类型=%s, 长度=%d, 元数据=%v",
			i, chunk.Type, len(chunk.Content), chunk.Metadata)
	}
}

// BenchmarkCustomStrategy 自定义策略性能基准测试
func BenchmarkCustomStrategy(b *testing.B) {
	// 创建测试策略
	builder := NewCustomStrategyBuilder("benchmark-strategy", "基准测试策略")
	builder.AddRule("heading-rule", "标题规则",
		NewHeadingLevelCondition(1, 3),
		NewCreateSeparateChunkAction("", nil),
		100)
	builder.AddRule("paragraph-rule", "段落规则",
		NewContentTypeCondition("paragraph"),
		NewMergeWithParentAction("\n\n"),
		50)

	strategy, err := builder.Build()
	if err != nil {
		b.Fatalf("构建策略失败：%v", err)
	}

	// 创建分块器
	chunker := NewMarkdownChunker()

	// 测试内容
	content := []byte(`# 标题1

这是第一段内容。

## 标题2

这是第二段内容。

### 标题3

这是第三段内容。`)

	// 解析文档
	reader := text.NewReader(content)
	doc := chunker.md.Parser().Parse(reader)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			b.Fatalf("分块失败：%v", err)
		}
	}
}
