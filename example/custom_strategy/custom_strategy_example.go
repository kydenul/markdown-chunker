package main

import (
	"fmt"
	"log"
	"strings"

	markdownchunker "github.com/pzierahn/markdown-chunker"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func main() {
	fmt.Println("=== 自定义分块策略示例 ===\n")

	// 示例1：基于标题层级的自定义策略
	fmt.Println("1. 基于标题层级的自定义策略")
	headingBasedExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例2：基于内容类型的自定义策略
	fmt.Println("2. 基于内容类型的自定义策略")
	contentTypeBasedExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例3：基于内容大小的自定义策略
	fmt.Println("3. 基于内容大小的自定义策略")
	sizeBasedExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例4：复杂的混合策略
	fmt.Println("4. 复杂的混合策略")
	complexStrategyExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例5：使用预构建的策略构建器
	fmt.Println("5. 使用预构建的策略构建器")
	prebuiltStrategyExample()
}

// headingBasedExample 基于标题层级的自定义策略示例
func headingBasedExample() {
	// 创建自定义策略构建器
	builder := markdownchunker.NewCustomStrategyBuilder(
		"heading-based-strategy",
		"基于标题层级的分块策略",
	)

	// 添加规则：H1-H2标题独立处理
	builder.AddRule(
		"main-headings",
		"主要标题独立处理",
		markdownchunker.NewHeadingLevelCondition(1, 2),
		markdownchunker.NewCreateSeparateChunkAction("main-heading", map[string]string{
			"importance": "high",
			"level":      "main",
		}),
		100, // 高优先级
	)

	// 添加规则：H3-H6标题合并到父块
	builder.AddRule(
		"sub-headings",
		"子标题合并处理",
		markdownchunker.NewHeadingLevelCondition(3, 6),
		markdownchunker.NewMergeWithParentAction("\n\n"),
		80, // 较高优先级
	)

	// 添加规则：段落内容合并到父块
	builder.AddRule(
		"paragraph-merge",
		"段落内容合并",
		markdownchunker.NewContentTypeCondition("paragraph"),
		markdownchunker.NewMergeWithParentAction("\n\n"),
		50, // 中等优先级
	)

	// 构建策略
	strategy, err := builder.Build()
	if err != nil {
		log.Fatalf("构建策略失败: %v", err)
	}

	// 创建分块器并设置策略
	chunker := markdownchunker.NewMarkdownChunker()

	// 测试内容
	content := []byte(`# 主标题

这是主标题下的内容。

## 二级标题

这是二级标题下的内容。

### 三级标题

这是三级标题下的内容。

#### 四级标题

这是四级标题下的内容。`)

	// 使用自定义策略分块
	chunks, err := processWithCustomStrategy(chunker, strategy, content)
	if err != nil {
		log.Fatalf("分块失败: %v", err)
	}

	// 显示结果
	fmt.Printf("生成了 %d 个块:\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("块 %d:\n", i+1)
		fmt.Printf("  类型: %s\n", chunk.Type)
		fmt.Printf("  内容长度: %d 字符\n", len(chunk.Content))
		fmt.Printf("  元数据: %v\n", chunk.Metadata)
		fmt.Printf("  内容预览: %s...\n", truncateString(chunk.Content, 50))
		fmt.Println()
	}
}

// contentTypeBasedExample 基于内容类型的自定义策略示例
func contentTypeBasedExample() {
	// 创建自定义策略构建器
	builder := markdownchunker.NewCustomStrategyBuilder(
		"content-type-strategy",
		"基于内容类型的分块策略",
	)

	// 添加规则：代码块独立处理
	builder.AddRule(
		"code-blocks",
		"代码块独立处理",
		markdownchunker.NewContentTypeCondition("code"),
		markdownchunker.NewCreateSeparateChunkAction("code-block", map[string]string{
			"type":     "code",
			"language": "auto-detect",
		}),
		100, // 最高优先级
	)

	// 添加规则：标题独立处理
	builder.AddRule(
		"headings",
		"标题独立处理",
		markdownchunker.NewContentTypeCondition("heading"),
		markdownchunker.NewCreateSeparateChunkAction("heading-block", map[string]string{
			"type": "heading",
		}),
		90, // 高优先级
	)

	// 添加规则：段落和列表合并
	builder.AddRule(
		"text-content",
		"文本内容合并",
		markdownchunker.NewContentTypeCondition("paragraph", "list", "blockquote"),
		markdownchunker.NewMergeWithParentAction("\n\n"),
		50, // 中等优先级
	)

	// 构建策略
	strategy, err := builder.Build()
	if err != nil {
		log.Fatalf("构建策略失败: %v", err)
	}

	// 创建分块器
	chunker := markdownchunker.NewMarkdownChunker()

	// 测试内容
	content := []byte(`# API 文档

这是一个 API 文档示例。

## 安装

使用以下命令安装：

` + "```bash\nnpm install example-api\n```" + `

## 使用方法

基本使用示例：

` + "```javascript\nconst api = require('example-api');\napi.connect();\n```" + `

### 配置选项

- 选项1：描述
- 选项2：描述
- 选项3：描述

> 注意：这是一个重要的提示信息。`)

	// 使用自定义策略分块
	chunks, err := processWithCustomStrategy(chunker, strategy, content)
	if err != nil {
		log.Fatalf("分块失败: %v", err)
	}

	// 显示结果
	fmt.Printf("生成了 %d 个块:\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("块 %d:\n", i+1)
		fmt.Printf("  类型: %s\n", chunk.Type)
		fmt.Printf("  内容长度: %d 字符\n", len(chunk.Content))
		if chunk.Metadata["type"] != "" {
			fmt.Printf("  内容类型: %s\n", chunk.Metadata["type"])
		}
		fmt.Printf("  内容预览: %s...\n", truncateString(chunk.Content, 50))
		fmt.Println()
	}
}

// sizeBasedExample 基于内容大小的自定义策略示例
func sizeBasedExample() {
	// 创建自定义策略构建器
	builder := markdownchunker.NewCustomStrategyBuilder(
		"size-based-strategy",
		"基于内容大小的分块策略",
	)

	// 添加规则：大内容独立处理
	builder.AddRule(
		"large-content",
		"大内容独立处理",
		markdownchunker.NewContentSizeCondition(200, 0), // 大于200字符
		markdownchunker.NewCreateSeparateChunkAction("large-chunk", map[string]string{
			"size": "large",
		}),
		100, // 最高优先级
	)

	// 添加规则：中等内容独立处理
	builder.AddRule(
		"medium-content",
		"中等内容独立处理",
		markdownchunker.NewContentSizeCondition(50, 200), // 50-200字符
		markdownchunker.NewCreateSeparateChunkAction("medium-chunk", map[string]string{
			"size": "medium",
		}),
		80, // 高优先级
	)

	// 添加规则：小内容合并
	builder.AddRule(
		"small-content",
		"小内容合并",
		markdownchunker.NewContentSizeCondition(0, 49), // 小于50字符
		markdownchunker.NewMergeWithParentAction(" "),
		60, // 中等优先级
	)

	// 构建策略
	strategy, err := builder.Build()
	if err != nil {
		log.Fatalf("构建策略失败: %v", err)
	}

	// 创建分块器
	chunker := markdownchunker.NewMarkdownChunker()

	// 测试内容
	content := []byte(`# 标题

短段落。

## 中等长度的标题

这是一个中等长度的段落，包含足够的文字来达到中等大小的阈值。这个段落应该被独立处理。

### 长内容标题

这是一个非常长的段落，包含大量的文字内容。这个段落的长度应该超过200个字符的阈值，因此会被标记为大内容并独立处理。这样的处理方式可以确保大块内容不会与其他内容混合，保持内容的完整性和可读性。这个段落继续添加更多内容以确保达到大内容的阈值。

很短。

另一个短段落。`)

	// 使用自定义策略分块
	chunks, err := processWithCustomStrategy(chunker, strategy, content)
	if err != nil {
		log.Fatalf("分块失败: %v", err)
	}

	// 显示结果
	fmt.Printf("生成了 %d 个块:\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("块 %d:\n", i+1)
		fmt.Printf("  类型: %s\n", chunk.Type)
		fmt.Printf("  内容长度: %d 字符\n", len(chunk.Content))
		if chunk.Metadata["size"] != "" {
			fmt.Printf("  大小类别: %s\n", chunk.Metadata["size"])
		}
		fmt.Printf("  内容预览: %s...\n", truncateString(chunk.Content, 50))
		fmt.Println()
	}
}

// complexStrategyExample 复杂的混合策略示例
func complexStrategyExample() {
	// 创建复杂的自定义策略构建器
	builder := markdownchunker.NewCustomStrategyBuilder(
		"complex-mixed-strategy",
		"复杂的混合分块策略",
	)

	// 规则1：重要标题（H1-H2）独立处理
	builder.AddRule(
		"important-headings",
		"重要标题独立处理",
		markdownchunker.NewHeadingLevelCondition(1, 2),
		markdownchunker.NewCreateSeparateChunkAction("important-heading", map[string]string{
			"importance": "high",
			"type":       "heading",
		}),
		100, // 最高优先级
	)

	// 规则2：代码块独立处理
	builder.AddRule(
		"code-blocks",
		"代码块独立处理",
		markdownchunker.NewContentTypeCondition("code"),
		markdownchunker.NewCreateSeparateChunkAction("code-block", map[string]string{
			"type":     "code",
			"language": "auto-detect",
		}),
		95, // 很高优先级
	)

	// 规则3：大段落独立处理
	builder.AddRule(
		"large-paragraphs",
		"大段落独立处理",
		markdownchunker.NewContentSizeCondition(150, 0),
		markdownchunker.NewCreateSeparateChunkAction("large-paragraph", map[string]string{
			"size": "large",
			"type": "paragraph",
		}),
		90, // 高优先级
	)

	// 规则4：次要标题（H3-H6）合并到父块
	builder.AddRule(
		"minor-headings",
		"次要标题合并处理",
		markdownchunker.NewHeadingLevelCondition(3, 6),
		markdownchunker.NewMergeWithParentAction("\n\n"),
		70, // 较高优先级
	)

	// 规则5：小段落合并
	builder.AddRule(
		"small-paragraphs",
		"小段落合并",
		markdownchunker.NewContentSizeCondition(0, 50),
		markdownchunker.NewMergeWithParentAction(" "),
		60, // 中等优先级
	)

	// 规则6：列表和引用合并
	builder.AddRule(
		"lists-and-quotes",
		"列表和引用合并",
		markdownchunker.NewContentTypeCondition("list", "blockquote"),
		markdownchunker.NewMergeWithParentAction("\n\n"),
		50, // 中等优先级
	)

	// 构建策略
	strategy, err := builder.Build()
	if err != nil {
		log.Fatalf("构建策略失败: %v", err)
	}

	// 创建分块器
	chunker := markdownchunker.NewMarkdownChunker()

	// 复杂测试内容
	content := []byte(`# 项目文档

这是项目的主要文档。

## 快速开始

这是一个详细的快速开始指南，包含了所有必要的步骤和说明。这个段落足够长，应该被识别为大段落并独立处理。用户可以按照这些步骤快速上手项目。

### 安装步骤

1. 下载项目
2. 安装依赖
3. 配置环境

` + "```bash\nnpm install\nnpm start\n```" + `

### 配置说明

短配置说明。

## API 参考

这是另一个很长的段落，详细描述了API的使用方法和注意事项。这个段落包含了大量的技术细节和示例代码说明，应该被独立处理以保持内容的完整性和可读性。

#### 端点列表

- GET /api/users
- POST /api/users
- PUT /api/users/:id

> 重要提示：请确保在使用API时包含正确的认证头。

` + "```javascript\nfetch('/api/users', {\n  headers: {\n    'Authorization': 'Bearer token'\n  }\n});\n```" + `

简短总结。`)

	// 使用自定义策略分块
	chunks, err := processWithCustomStrategy(chunker, strategy, content)
	if err != nil {
		log.Fatalf("分块失败: %v", err)
	}

	// 显示结果
	fmt.Printf("生成了 %d 个块:\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("块 %d:\n", i+1)
		fmt.Printf("  类型: %s\n", chunk.Type)
		fmt.Printf("  内容长度: %d 字符\n", len(chunk.Content))

		// 显示相关元数据
		if chunk.Metadata["importance"] != "" {
			fmt.Printf("  重要性: %s\n", chunk.Metadata["importance"])
		}
		if chunk.Metadata["size"] != "" {
			fmt.Printf("  大小: %s\n", chunk.Metadata["size"])
		}
		if chunk.Metadata["type"] != "" {
			fmt.Printf("  内容类型: %s\n", chunk.Metadata["type"])
		}

		fmt.Printf("  内容预览: %s...\n", truncateString(chunk.Content, 60))
		fmt.Println()
	}
}

// prebuiltStrategyExample 使用预构建的策略构建器示例
func prebuiltStrategyExample() {
	fmt.Println("使用预构建的策略构建器:")

	// 1. 基于标题的策略构建器
	fmt.Println("\n1. 基于标题的策略构建器:")
	headingBuilder := markdownchunker.NewHeadingBasedStrategyBuilder("heading-strategy", 3)
	headingStrategy, err := headingBuilder.Build()
	if err != nil {
		log.Fatalf("构建标题策略失败: %v", err)
	}
	fmt.Printf("   策略名称: %s\n", headingStrategy.GetName())
	fmt.Printf("   策略描述: %s\n", headingStrategy.GetDescription())

	// 2. 基于内容类型的策略构建器
	fmt.Println("\n2. 基于内容类型的策略构建器:")
	separateTypes := []string{"heading", "code"}
	mergeTypes := []string{"paragraph", "list"}
	contentTypeBuilder := markdownchunker.NewContentTypeBasedStrategyBuilder(
		"content-type-strategy", separateTypes, mergeTypes)
	contentTypeStrategy, err := contentTypeBuilder.Build()
	if err != nil {
		log.Fatalf("构建内容类型策略失败: %v", err)
	}
	fmt.Printf("   策略名称: %s\n", contentTypeStrategy.GetName())
	fmt.Printf("   策略描述: %s\n", contentTypeStrategy.GetDescription())

	// 3. 基于大小的策略构建器
	fmt.Println("\n3. 基于大小的策略构建器:")
	sizeBuilder := markdownchunker.NewSizeBasedStrategyBuilder("size-strategy", 50, 200)
	sizeStrategy, err := sizeBuilder.Build()
	if err != nil {
		log.Fatalf("构建大小策略失败: %v", err)
	}
	fmt.Printf("   策略名称: %s\n", sizeStrategy.GetName())
	fmt.Printf("   策略描述: %s\n", sizeStrategy.GetDescription())

	// 测试其中一个策略
	fmt.Println("\n测试基于标题的策略:")
	chunker := markdownchunker.NewMarkdownChunker()

	content := []byte(`# 主标题

主标题内容。

## 子标题

子标题内容。

### 三级标题

三级标题内容。

#### 四级标题

四级标题内容。`)

	chunks, err := processWithCustomStrategy(chunker, headingStrategy, content)
	if err != nil {
		log.Fatalf("分块失败: %v", err)
	}

	fmt.Printf("生成了 %d 个块\n", len(chunks))
	for i, chunk := range chunks {
		fmt.Printf("  块 %d: %s (长度: %d)\n", i+1, chunk.Type, len(chunk.Content))
	}
}

// processWithCustomStrategy 使用自定义策略处理内容的辅助函数
func processWithCustomStrategy(
	chunker *markdownchunker.MarkdownChunker, strategy markdownchunker.ChunkingStrategy, content []byte,
) ([]markdownchunker.Chunk, error) {
	// 由于 md 字段不是导出的，我们需要使用分块器的公共方法
	// 这里我们创建一个临时的 goldmark 实例来解析文档
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	// 使用自定义策略分块
	return strategy.ChunkDocument(doc, content, chunker)
}

// truncateString 截断字符串的辅助函数
func truncateString(s string, maxLen int) string {
	// 移除换行符和多余空格
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
