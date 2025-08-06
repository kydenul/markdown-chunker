package main

import (
	"fmt"
	"log"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	markdown := `# 配置系统示例

这是一个展示配置系统功能的示例文档。

## 代码示例

` + "```go" + `
func main() {
    if true {
        for i := 0; i < 10; i++ {
            fmt.Println("Hello World")
        }
    }
}
` + "```" + `

## 链接和图片

这里有一个[链接](https://example.com)和一张![图片](image.jpg)。

## 列表

- 项目 1
- 项目 2
- 项目 3

> 这是一个引用块，包含一些重要信息。

---

*文档结束*`

	fmt.Println("=== 默认配置示例 ===")
	defaultChunker := mc.NewMarkdownChunker()
	chunks, err := defaultChunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("默认配置处理结果：%d 个块\n", len(chunks))
	for _, chunk := range chunks {
		fmt.Printf("- %s: %s\n", chunk.Type, truncateString(chunk.Text, 50))
	}

	fmt.Println("\n=== 自定义配置示例 1：只处理标题和段落 ===")
	config1 := mc.DefaultConfig()
	config1.EnabledTypes = map[string]bool{
		"heading":        true,
		"paragraph":      true,
		"code":           false,
		"list":           false,
		"blockquote":     false,
		"thematic_break": false,
	}

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
	chunks1, err := chunker1.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("只处理标题和段落：%d 个块\n", len(chunks1))
	for _, chunk := range chunks1 {
		fmt.Printf("- %s: %s\n", chunk.Type, truncateString(chunk.Text, 50))
	}

	fmt.Println("\n=== 自定义配置示例 2：使用自定义元数据提取器 ===")
	config2 := mc.DefaultConfig()
	config2.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
	chunks2, err := chunker2.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("使用自定义元数据提取器：%d 个块\n", len(chunks2))
	for _, chunk := range chunks2 {
		fmt.Printf("- %s: %s\n", chunk.Type, truncateString(chunk.Text, 50))
		// 显示自定义元数据
		for key, value := range chunk.Metadata {
			if key == "link_count" || key == "image_count" || key == "code_complexity" {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	fmt.Println("\n=== 自定义配置示例 3：限制块大小 ===")
	config3 := mc.DefaultConfig()
	config3.MaxChunkSize = 100                     // 限制每个块最大100字符
	config3.ErrorHandling = mc.ErrorModePermissive // 宽松模式，截断而不是报错

	chunker3 := mc.NewMarkdownChunkerWithConfig(config3)
	chunks3, err := chunker3.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("限制块大小（100字符）：%d 个块\n", len(chunks3))
	for _, chunk := range chunks3 {
		fmt.Printf("- %s (%d chars): %s\n", chunk.Type, len(chunk.Content), truncateString(chunk.Text, 50))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
