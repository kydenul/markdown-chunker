package main

import (
	"fmt"
	"log"
	"strings"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 高级配置使用示例 ===")

	// 测试文档
	testMarkdown := `# 高级配置测试文档

这是一个用于测试高级配置功能的示例文档。

## 链接和图片测试

这里有一个[外部链接](https://example.com)，一个[内部链接](/page)，和一个[锚点链接](#section)。

还有一些图片：![图片1](image1.jpg "图片标题") 和 ![图片2](image2.png)。

## 代码复杂度测试

` + "```python" + `
def complex_algorithm(data):
    result = []
    for item in data:
        if item > 0:
            for i in range(item):
                if i % 2 == 0:
                    result.append(i * 2)
                else:
                    result.append(i * 3)
        elif item < 0:
            try:
                result.append(abs(item))
            except Exception as e:
                print(f"Error: {e}")
        else:
            result.append(0)
    return result
` + "```" + `

## 表格测试

| 用户名 | 邮箱 | 状态 | 积分 |
|--------|------|------|------|
| 张三 | zhang@example.com | 活跃 | 1250 |
| 李四 | li@test.org | 非活跃 | 890.5 |
| 王五 | wang@demo.net | 活跃 | 2100 |

## 列表测试

1. 第一项
2. 第二项
   - 子项 A
   - 子项 B
3. 第三项

> 这是一个引用块，包含重要信息。

---

*文档结束*`

	// 示例 1: 基础配置自定义
	fmt.Println("\n1. 基础配置自定义")
	demonstrateBasicConfiguration(testMarkdown)

	// 示例 2: 内容类型过滤
	fmt.Println("\n2. 内容类型过滤")
	demonstrateContentTypeFiltering(testMarkdown)

	// 示例 3: 自定义元数据提取器
	fmt.Println("\n3. 自定义元数据提取器")
	demonstrateCustomExtractors(testMarkdown)

	// 示例 4: 错误处理模式配置
	fmt.Println("\n4. 错误处理模式配置")
	demonstrateErrorHandlingModes(testMarkdown)

	// 示例 5: 性能模式配置
	fmt.Println("\n5. 性能模式配置")
	demonstratePerformanceModes(testMarkdown)

	// 示例 6: 块大小限制配置
	fmt.Println("\n6. 块大小限制配置")
	demonstrateChunkSizeLimits(testMarkdown)

	// 示例 7: 完整的高级配置
	fmt.Println("\n7. 完整的高级配置")
	demonstrateCompleteAdvancedConfiguration(testMarkdown)
}

// demonstrateBasicConfiguration 演示基础配置自定义
func demonstrateBasicConfiguration(markdown string) {
	config := mc.DefaultConfig()
	config.FilterEmptyChunks = true
	config.PreserveWhitespace = false

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("基础配置结果: %d 个块\n", len(chunks))
	fmt.Printf("过滤空块: %t\n", config.FilterEmptyChunks)
	fmt.Printf("保留空白: %t\n", config.PreserveWhitespace)
}

// demonstrateContentTypeFiltering 演示内容类型过滤
func demonstrateContentTypeFiltering(markdown string) {
	// 只处理标题和段落
	config1 := mc.DefaultConfig()
	config1.EnabledTypes = map[string]bool{
		"heading":        true,
		"paragraph":      true,
		"code":           false,
		"table":          false,
		"list":           false,
		"blockquote":     false,
		"thematic_break": false,
	}

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
	chunks1, _ := chunker1.ChunkDocument([]byte(markdown))

	fmt.Printf("只处理标题和段落: %d 个块\n", len(chunks1))
	for _, chunk := range chunks1 {
		fmt.Printf("  - %s: %s\n", chunk.Type, truncateString(chunk.Text, 40))
	}

	// 只处理代码和表格
	config2 := mc.DefaultConfig()
	config2.EnabledTypes = map[string]bool{
		"heading":        false,
		"paragraph":      false,
		"code":           true,
		"table":          true,
		"list":           false,
		"blockquote":     false,
		"thematic_break": false,
	}

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
	chunks2, _ := chunker2.ChunkDocument([]byte(markdown))

	fmt.Printf("只处理代码和表格: %d 个块\n", len(chunks2))
	for _, chunk := range chunks2 {
		fmt.Printf("  - %s: %s\n", chunk.Type, truncateString(chunk.Text, 40))
	}
}

// demonstrateCustomExtractors 演示自定义元数据提取器
func demonstrateCustomExtractors(markdown string) {
	config := mc.DefaultConfig()
	config.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("使用自定义元数据提取器: %d 个块\n", len(chunks))
	for _, chunk := range chunks {
		fmt.Printf("\n块类型: %s\n", chunk.Type)
		fmt.Printf("内容: %s\n", truncateString(chunk.Text, 50))

		// 显示链接信息
		if len(chunk.Links) > 0 {
			fmt.Printf("链接 (%d):\n", len(chunk.Links))
			for _, link := range chunk.Links {
				fmt.Printf("  - %s (%s): %s\n", link.Text, link.Type, link.URL)
			}
		}

		// 显示图片信息
		if len(chunk.Images) > 0 {
			fmt.Printf("图片 (%d):\n", len(chunk.Images))
			for _, img := range chunk.Images {
				fmt.Printf("  - %s: %s\n", img.Alt, img.URL)
				if img.Title != "" {
					fmt.Printf("    标题: %s\n", img.Title)
				}
			}
		}

		// 显示自定义元数据
		for key, value := range chunk.Metadata {
			if key == "link_count" || key == "image_count" || key == "code_complexity" ||
				key == "external_links" || key == "internal_links" || key == "anchor_links" ||
				key == "function_count" || key == "loop_count" || key == "conditional_count" {
				fmt.Printf("元数据 - %s: %s\n", key, value)
			}
		}
	}
}

// demonstrateErrorHandlingModes 演示错误处理模式
func demonstrateErrorHandlingModes(markdown string) {
	// 创建一个会导致错误的配置（块大小限制很小）
	problematicMarkdown := markdown + strings.Repeat("这是一个非常长的段落，用于测试错误处理。", 50)

	// 严格模式
	fmt.Println("严格模式:")
	config1 := mc.DefaultConfig()
	config1.MaxChunkSize = 100
	config1.ErrorHandling = mc.ErrorModeStrict

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
	chunks1, err1 := chunker1.ChunkDocument([]byte(problematicMarkdown))

	if err1 != nil {
		fmt.Printf("  错误: %s\n", err1.Error())
		fmt.Printf("  块数量: %d\n", len(chunks1))
	} else {
		fmt.Printf("  成功处理: %d 个块\n", len(chunks1))
	}

	// 宽松模式
	fmt.Println("宽松模式:")
	config2 := mc.DefaultConfig()
	config2.MaxChunkSize = 100
	config2.ErrorHandling = mc.ErrorModePermissive

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
	chunks2, err2 := chunker2.ChunkDocument([]byte(problematicMarkdown))

	fmt.Printf("  返回错误: %v\n", err2)
	fmt.Printf("  块数量: %d\n", len(chunks2))
	fmt.Printf("  记录的错误: %d\n", len(chunker2.GetErrors()))

	// 静默模式
	fmt.Println("静默模式:")
	config3 := mc.DefaultConfig()
	config3.MaxChunkSize = 100
	config3.ErrorHandling = mc.ErrorModeSilent

	chunker3 := mc.NewMarkdownChunkerWithConfig(config3)
	chunks3, err3 := chunker3.ChunkDocument([]byte(problematicMarkdown))

	fmt.Printf("  返回错误: %v\n", err3)
	fmt.Printf("  块数量: %d\n", len(chunks3))
	fmt.Printf("  记录的错误: %d\n", len(chunker3.GetErrors()))
}

// demonstratePerformanceModes 演示性能模式
func demonstratePerformanceModes(markdown string) {
	modes := []struct {
		name string
		mode mc.PerformanceMode
	}{
		{"默认模式", mc.PerformanceModeDefault},
		{"内存优化", mc.PerformanceModeMemoryOptimized},
		{"速度优化", mc.PerformanceModeSpeedOptimized},
	}

	for _, m := range modes {
		config := mc.DefaultConfig()
		config.PerformanceMode = m.mode
		config.EnableObjectPooling = true

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}

		stats := chunker.GetPerformanceStats()
		fmt.Printf("%s:\n", m.name)
		fmt.Printf("  块数量: %d\n", len(chunks))
		fmt.Printf("  处理时间: %v\n", stats.ProcessingTime)
		fmt.Printf("  内存使用: %d KB\n", stats.MemoryUsed/1024)
		fmt.Printf("  处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)
	}
}

// demonstrateChunkSizeLimits 演示块大小限制
func demonstrateChunkSizeLimits(markdown string) {
	limits := []int{50, 200, 500, 0} // 0 表示无限制

	for _, limit := range limits {
		config := mc.DefaultConfig()
		config.MaxChunkSize = limit
		config.ErrorHandling = mc.ErrorModePermissive

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}

		limitStr := fmt.Sprintf("%d", limit)
		if limit == 0 {
			limitStr = "无限制"
		}

		fmt.Printf("块大小限制 %s:\n", limitStr)
		fmt.Printf("  块数量: %d\n", len(chunks))

		// 显示块大小分布
		var oversized int
		for _, chunk := range chunks {
			if limit > 0 && len(chunk.Content) > limit {
				oversized++
			}
		}

		if oversized > 0 {
			fmt.Printf("  超大块数量: %d\n", oversized)
		}

		if chunker.HasErrors() {
			sizeErrors := chunker.GetErrorsByType(mc.ErrorTypeChunkTooLarge)
			fmt.Printf("  大小错误: %d\n", len(sizeErrors))
		}
	}
}

// demonstrateCompleteAdvancedConfiguration 演示完整的高级配置
func demonstrateCompleteAdvancedConfiguration(markdown string) {
	config := mc.DefaultConfig()

	// 内容类型配置
	config.EnabledTypes = map[string]bool{
		"heading":        true,
		"paragraph":      true,
		"code":           true,
		"table":          true,
		"list":           true,
		"blockquote":     true,
		"thematic_break": false, // 禁用分隔线
	}

	// 大小和过滤配置
	config.MaxChunkSize = 1000
	config.FilterEmptyChunks = true
	config.PreserveWhitespace = false

	// 错误处理配置
	config.ErrorHandling = mc.ErrorModePermissive

	// 性能配置
	config.PerformanceMode = mc.PerformanceModeSpeedOptimized
	config.EnableObjectPooling = true
	config.MemoryLimit = 50 * 1024 * 1024 // 50MB

	// 自定义元数据提取器
	config.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}

	// 验证配置
	if err := mc.ValidateConfig(config); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("完整高级配置结果:\n")
	fmt.Printf("  块数量: %d\n", len(chunks))

	// 统计各类型块数量
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.Type]++
	}

	fmt.Printf("  块类型分布:\n")
	for chunkType, count := range typeCount {
		fmt.Printf("    %s: %d\n", chunkType, count)
	}

	// 性能统计
	stats := chunker.GetPerformanceStats()
	fmt.Printf("  性能统计:\n")
	fmt.Printf("    处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("    内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("    处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)

	// 错误统计
	if chunker.HasErrors() {
		fmt.Printf("  错误统计:\n")
		errors := chunker.GetErrors()
		errorTypeCount := make(map[string]int)
		for _, err := range errors {
			errorTypeCount[err.Type.String()]++
		}
		for errorType, count := range errorTypeCount {
			fmt.Printf("    %s: %d\n", errorType, count)
		}
	}

	// 显示前几个块的详细信息
	fmt.Printf("  前3个块的详细信息:\n")
	for i, chunk := range chunks {
		if i >= 3 {
			break
		}

		fmt.Printf("    块 %d (%s):\n", i+1, chunk.Type)
		fmt.Printf("      位置: %d:%d-%d:%d\n",
			chunk.Position.StartLine, chunk.Position.StartCol,
			chunk.Position.EndLine, chunk.Position.EndCol)
		fmt.Printf("      内容长度: %d 字符\n", len(chunk.Content))
		fmt.Printf("      哈希: %s\n", chunk.Hash[:16])

		if len(chunk.Links) > 0 {
			fmt.Printf("      链接: %d 个\n", len(chunk.Links))
		}
		if len(chunk.Images) > 0 {
			fmt.Printf("      图片: %d 个\n", len(chunk.Images))
		}

		// 显示特殊元数据
		for key, value := range chunk.Metadata {
			if key == "code_complexity" || key == "link_count" || key == "image_count" {
				fmt.Printf("      %s: %s\n", key, value)
			}
		}
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
