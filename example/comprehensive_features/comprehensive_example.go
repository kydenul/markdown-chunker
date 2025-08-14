package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 综合功能演示示例 ===")

	// 创建一个复杂的测试文档
	complexMarkdown := createComplexTestDocument()

	// 演示各种功能
	demonstrateBasicUsage(complexMarkdown)
	demonstrateChunkingStrategies(complexMarkdown) // NEW: Strategy demonstration
	demonstrateAdvancedFeatures(complexMarkdown)
	demonstrateErrorHandlingAndRecovery(complexMarkdown)
	demonstratePerformanceMonitoring(complexMarkdown)
	demonstrateLoggingFeatures(complexMarkdown)
	demonstrateMetadataExtraction(complexMarkdown)
	demonstrateContentAnalysis(complexMarkdown)
}

// createComplexTestDocument 创建复杂的测试文档
func createComplexTestDocument() string {
	return `# 综合功能测试文档

这是一个用于测试 Markdown Chunker 所有功能的综合文档。

## 链接和图片测试

### 各种类型的链接

- 外部链接: [Google](https://www.google.com)
- 内部链接: [关于页面](/about)
- 锚点链接: [跳转到结论](#conclusion)
- 邮件链接: [联系我们](mailto:contact@example.com)
- 自动链接: https://github.com/example/repo

### 图片测试

![主图片](https://example.com/main.jpg "主图片标题")

![本地图片](./images/local.png)

![无标题图片](image-without-title.gif)

## 代码复杂度测试

### 简单代码

` + "```python" + `
def simple_function():
    return "Hello, World!"
` + "```" + `

### 复杂代码

` + "```javascript" + `
function complexAlgorithm(data) {
    let result = [];
    
    for (let i = 0; i < data.length; i++) {
        if (data[i] > 0) {
            for (let j = 0; j < data[i]; j++) {
                if (j % 2 === 0) {
                    result.push(j * 2);
                } else {
                    try {
                        result.push(processOddNumber(j));
                    } catch (error) {
                        console.error("Error processing:", error);
                        continue;
                    }
                }
            }
        } else if (data[i] < 0) {
            while (Math.abs(data[i]) > 0) {
                result.push(Math.abs(data[i]));
                data[i]++;
            }
        } else {
            switch (data[i]) {
                case 0:
                    result.push(0);
                    break;
                default:
                    result.push(null);
            }
        }
    }
    
    return result.filter(item => item !== null);
}

function processOddNumber(num) {
    if (num < 0) throw new Error("Negative number");
    return num * 3;
}
` + "```" + `

### Go 代码示例

` + "```go" + `
package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    ch := make(chan int, 10)
    
    // 启动工作协程
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for num := range ch {
                if num%2 == 0 {
                    fmt.Printf("Worker %d: Even %d\n", id, num)
                } else {
                    fmt.Printf("Worker %d: Odd %d\n", id, num)
                }
            }
        }(i)
    }
    
    // 发送数据
    for i := 0; i < 20; i++ {
        ch <- i
    }
    close(ch)
    
    wg.Wait()
}
` + "```" + `

## 表格测试

### 标准表格

| 产品名称 | 价格 | 库存状态 | 链接 |
|----------|------|----------|------|
| 笔记本电脑 | ¥5999 | 有库存 | [查看详情](https://shop.com/laptop) |
| 智能手机 | ¥2999 | 缺货 | [查看详情](https://shop.com/phone) |
| 平板电脑 | ¥1999 | 有库存 | [查看详情](https://shop.com/tablet) |

### 带对齐的表格

| 左对齐 | 居中对齐 | 右对齐 | 数值 |
|:-------|:--------:|-------:|-----:|
| 文本A | 文本B | 文本C | 123.45 |
| 长文本内容 | 短文本 | 文本 | 67.89 |
| A | 中等长度文本 | 很长的文本内容 | 0.12 |

### 格式不规范的表格

| 姓名 | 年龄 | 城市 |
|------|------|------|
| 张三 | 25 | 北京 | 多余列 |
| 李四 | 30 |  |
| 王五 |  | 上海 |

## 列表测试

### 有序列表

1. 第一项
2. 第二项
   1. 子项 2.1
   2. 子项 2.2
      1. 深层子项 2.2.1
3. 第三项

### 无序列表

- 项目 A
- 项目 B
  - 子项目 B.1
  - 子项目 B.2
    - 深层子项目 B.2.1
- 项目 C

### 混合列表

1. 有序项目 1
   - 无序子项目
   - 另一个无序子项目
2. 有序项目 2
   1. 有序子项目
   2. 另一个有序子项目

## 引用块测试

> 这是一个简单的引用块。

> 这是一个多行引用块。
> 它包含多行内容，
> 用于测试引用块的处理能力。

> 这是一个包含链接的引用块：[链接](https://example.com)
> 
> 还包含**粗体**和*斜体*文本。

### 嵌套引用

> 这是外层引用。
> 
> > 这是嵌套引用。
> > 
> > > 这是深层嵌套引用。
> 
> 回到外层引用。

## 特殊内容测试

### 包含特殊字符的内容

这个段落包含特殊字符：` + "`代码`" + `、**粗体**、*斜体*、~~删除线~~。

还有一些 Unicode 字符：🚀 📝 💻 🎯

### 长段落测试

` + strings.Repeat("这是一个用于测试长段落处理能力的重复内容。", 20) + `

---

## 结论 {#conclusion}

这个文档包含了各种 Markdown 元素，用于全面测试分块器的功能：

- ✅ 标题层次结构
- ✅ 段落和文本格式
- ✅ 代码块（多种语言）
- ✅ 表格（标准和不规范）
- ✅ 列表（有序、无序、嵌套）
- ✅ 引用块（简单和嵌套）
- ✅ 链接（各种类型）
- ✅ 图片
- ✅ 分隔线
- ✅ 特殊字符和 Unicode

*文档结束*`
}

// demonstrateBasicUsage 演示基本用法
func demonstrateBasicUsage(markdown string) {
	fmt.Println("\n=== 1. 基本用法演示 ===")

	chunker := mc.NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("文档统计:\n")
	fmt.Printf("  原始文档大小: %d 字节\n", len(markdown))
	fmt.Printf("  生成块数量: %d\n", len(chunks))

	// 统计各类型块
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.Type]++
	}

	fmt.Printf("  块类型分布:\n")
	for chunkType, count := range typeCount {
		fmt.Printf("    %s: %d\n", chunkType, count)
	}
}

// demonstrateChunkingStrategies 演示分块策略
func demonstrateChunkingStrategies(markdown string) {
	fmt.Println("\n=== 2. 分块策略演示 ===")

	strategies := []struct {
		name   string
		config *mc.StrategyConfig
		desc   string
	}{
		{"元素级策略", mc.ElementLevelConfig(), "逐个元素分块（默认行为）"},
		{"层级策略(深度2)", mc.HierarchicalConfig(2), "按标题层级分组内容"},
		{"层级策略(深度3)", mc.HierarchicalConfig(3), "更深层级的内容分组"},
		{"文档级策略", mc.DocumentLevelConfig(), "整个文档作为单个块"},
	}

	fmt.Printf("测试文档大小: %d 字节\n\n", len(markdown))

	for i, strategy := range strategies {
		fmt.Printf("%d. %s - %s\n", i+1, strategy.name, strategy.desc)

		config := mc.DefaultConfig()
		config.ChunkingStrategy = strategy.config

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		start := time.Now()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   错误: %v\n\n", err)
			continue
		}

		stats := chunker.GetPerformanceStats()
		fmt.Printf("   块数量: %d\n", len(chunks))
		fmt.Printf("   处理时间: %v\n", duration)
		fmt.Printf("   内存使用: %d KB\n", stats.MemoryUsed/1024)

		// 显示前几个块的信息
		fmt.Printf("   前3个块:\n")
		for j, chunk := range chunks[:min(3, len(chunks))] {
			preview := strings.ReplaceAll(chunk.Text, "\n", " ")
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("     %d. %s (Level %d): %s\n", j+1, chunk.Type, chunk.Level, preview)
		}
		if len(chunks) > 3 {
			fmt.Printf("     ... 还有 %d 个块\n", len(chunks)-3)
		}
		fmt.Println()
	}

	// 演示动态策略切换
	fmt.Println("动态策略切换演示:")
	chunker := mc.NewMarkdownChunker()

	// 开始使用元素级策略
	chunks1, _ := chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("  元素级策略: %d 块\n", len(chunks1))

	// 切换到层级策略
	chunker.SetStrategy("hierarchical", mc.HierarchicalConfig(2))
	chunks2, _ := chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("  层级策略: %d 块\n", len(chunks2))

	// 切换到文档级策略
	chunker.SetStrategy("document-level", mc.DocumentLevelConfig())
	chunks3, _ := chunker.ChunkDocument([]byte(markdown))
	fmt.Printf("  文档级策略: %d 块\n", len(chunks3))
}

// demonstrateAdvancedFeatures 演示高级功能
func demonstrateAdvancedFeatures(markdown string) {
	fmt.Println("\n=== 2. 高级功能演示 ===")

	config := mc.DefaultConfig()
	config.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}
	config.FilterEmptyChunks = true
	config.MaxChunkSize = 2000

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("高级功能结果:\n")
	fmt.Printf("  处理块数: %d\n", len(chunks))

	// 统计链接和图片
	totalLinks := 0
	totalImages := 0
	for _, chunk := range chunks {
		totalLinks += len(chunk.Links)
		totalImages += len(chunk.Images)
	}

	fmt.Printf("  提取的链接总数: %d\n", totalLinks)
	fmt.Printf("  提取的图片总数: %d\n", totalImages)

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
		fmt.Printf("      内容长度: %d\n", len(chunk.Content))
		fmt.Printf("      哈希: %s...\n", chunk.Hash[:16])

		if len(chunk.Links) > 0 {
			fmt.Printf("      链接:\n")
			for _, link := range chunk.Links {
				fmt.Printf("        - %s (%s): %s\n", link.Text, link.Type, link.URL)
			}
		}

		if len(chunk.Images) > 0 {
			fmt.Printf("      图片:\n")
			for _, img := range chunk.Images {
				fmt.Printf("        - %s: %s\n", img.Alt, img.URL)
			}
		}
	}
}

// demonstrateErrorHandlingAndRecovery 演示错误处理和恢复
func demonstrateErrorHandlingAndRecovery(markdown string) {
	fmt.Println("\n=== 3. 错误处理和恢复演示 ===")

	// 创建会产生错误的配置
	config := mc.DefaultConfig()
	config.MaxChunkSize = 200 // 较小的限制
	config.ErrorHandling = mc.ErrorModePermissive

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))

	fmt.Printf("错误处理结果:\n")
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  成功处理块数: %d\n", len(chunks))
	fmt.Printf("  记录的错误数: %d\n", len(chunker.GetErrors()))

	if chunker.HasErrors() {
		// 按类型统计错误
		errorTypeCount := make(map[mc.ErrorType]int)
		for _, err := range chunker.GetErrors() {
			errorTypeCount[err.Type]++
		}

		fmt.Printf("  错误类型分布:\n")
		for errorType, count := range errorTypeCount {
			fmt.Printf("    %s: %d\n", errorType.String(), count)
		}

		// 显示前几个错误的详细信息
		fmt.Printf("  前3个错误详情:\n")
		for i, err := range chunker.GetErrors() {
			if i >= 3 {
				break
			}
			fmt.Printf("    错误 %d: %s - %s\n", i+1, err.Type.String(), err.Message)
			if len(err.Context) > 0 {
				fmt.Printf("      上下文: %+v\n", err.Context)
			}
		}
	}
}

// demonstratePerformanceMonitoring 演示性能监控
func demonstratePerformanceMonitoring(markdown string) {
	fmt.Println("\n=== 4. 性能监控演示 ===")

	config := mc.DefaultConfig()
	config.PerformanceMode = mc.PerformanceModeSpeedOptimized
	config.EnableObjectPooling = true

	chunker := mc.NewMarkdownChunkerWithConfig(config)

	// 多次处理以获得更准确的性能数据
	var totalChunks int
	iterations := 3

	start := time.Now()
	for i := 0; i < iterations; i++ {
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("处理错误: %v", err)
			continue
		}
		totalChunks += len(chunks)
		chunker.ResetPerformanceMonitor()
	}
	totalTime := time.Since(start)

	// 最后一次处理获取详细统计
	chunks, _ := chunker.ChunkDocument([]byte(markdown))
	stats := chunker.GetPerformanceStats()

	fmt.Printf("性能监控结果 (%d次处理):\n", iterations)
	fmt.Printf("  总处理时间: %v\n", totalTime)
	fmt.Printf("  平均每次时间: %v\n", totalTime/time.Duration(iterations))
	fmt.Printf("  最后一次统计:\n")
	fmt.Printf("    处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("    内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("    峰值内存: %d KB\n", stats.PeakMemory/1024)
	fmt.Printf("    处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)
	fmt.Printf("    字节处理速度: %.2f KB/秒\n", stats.BytesPerSecond/1024)
	fmt.Printf("    总块数: %d\n", stats.TotalChunks)
	fmt.Printf("    块内容总大小: %d 字节\n", stats.ChunkBytes)

	fmt.Printf("  平均每块大小: %.2f 字节\n", float64(stats.ChunkBytes)/float64(len(chunks)))
}

// demonstrateLoggingFeatures 演示日志功能
func demonstrateLoggingFeatures(markdown string) {
	fmt.Println("\n=== 5. 日志功能演示 ===")

	// 演示不同日志级别
	logLevels := []string{"ERROR", "WARN", "INFO", "DEBUG"}

	for _, level := range logLevels {
		fmt.Printf("测试日志级别: %s\n", level)

		config := mc.DefaultConfig()
		config.EnableLog = true
		config.LogLevel = level
		config.LogFormat = "console"
		config.LogDirectory = fmt.Sprintf("./demo-logs/%s", strings.ToLower(level))

		// 为了演示错误日志，在某些级别设置小的块大小限制
		if level == "ERROR" || level == "WARN" {
			config.MaxChunkSize = 200
			config.ErrorHandling = mc.ErrorModePermissive
		}

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		start := time.Now()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		processingTime := time.Since(start)

		fmt.Printf("  处理结果: %d 个块\n", len(chunks))
		fmt.Printf("  处理时间: %v\n", processingTime)
		fmt.Printf("  返回错误: %v\n", err)

		if chunker.HasErrors() {
			fmt.Printf("  记录的错误: %d\n", len(chunker.GetErrors()))
		}

		fmt.Printf("  日志目录: %s\n", config.LogDirectory)
		fmt.Println()
	}

	// 演示JSON格式日志
	fmt.Println("JSON格式日志演示:")
	config := mc.DefaultConfig()
	config.EnableLog = true
	config.LogLevel = "INFO"
	config.LogFormat = "json"
	config.LogDirectory = "./demo-logs/json"
	config.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))

	fmt.Printf("  JSON格式处理结果: %d 个块\n", len(chunks))
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  日志目录: %s\n", config.LogDirectory)
	fmt.Println("  JSON格式便于日志聚合和分析")

	// 显示性能统计（也会被记录到日志中）
	stats := chunker.GetPerformanceStats()
	fmt.Printf("  性能统计（已记录到日志）:\n")
	fmt.Printf("    处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("    内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("    处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)
}

// demonstrateMetadataExtraction 演示元数据提取
func demonstrateMetadataExtraction(markdown string) {
	fmt.Println("\n=== 5. 元数据提取演示 ===")

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

	fmt.Printf("元数据提取结果:\n")

	// 分析代码块的复杂度
	fmt.Printf("  代码复杂度分析:\n")
	for _, chunk := range chunks {
		if chunk.Type == "code" {
			if complexity, exists := chunk.Metadata["code_complexity"]; exists {
				language := chunk.Metadata["language"]
				fmt.Printf("    %s 代码块复杂度: %s\n", language, complexity)

				if funcCount, exists := chunk.Metadata["function_count"]; exists {
					fmt.Printf("      函数数量: %s\n", funcCount)
				}
				if loopCount, exists := chunk.Metadata["loop_count"]; exists {
					fmt.Printf("      循环数量: %s\n", loopCount)
				}
				if condCount, exists := chunk.Metadata["conditional_count"]; exists {
					fmt.Printf("      条件语句数量: %s\n", condCount)
				}
			}
		}
	}

	// 分析链接分布
	fmt.Printf("  链接分布分析:\n")
	linkTypes := make(map[string]int)
	for _, chunk := range chunks {
		for _, link := range chunk.Links {
			linkTypes[link.Type]++
		}
	}
	for linkType, count := range linkTypes {
		fmt.Printf("    %s 链接: %d 个\n", linkType, count)
	}

	// 分析图片类型
	fmt.Printf("  图片类型分析:\n")
	imageExts := make(map[string]int)
	for _, chunk := range chunks {
		for _, img := range chunk.Images {
			// 从URL中提取扩展名
			parts := strings.Split(img.URL, ".")
			if len(parts) > 1 {
				ext := strings.ToLower(parts[len(parts)-1])
				imageExts[ext]++
			}
		}
	}
	for ext, count := range imageExts {
		fmt.Printf("    .%s 图片: %d 个\n", ext, count)
	}
}

// demonstrateContentAnalysis 演示内容分析
func demonstrateContentAnalysis(markdown string) {
	fmt.Println("\n=== 7. 内容分析演示 ===")

	chunker := mc.NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	// 分析文档结构
	fmt.Printf("文档结构分析:\n")

	// 标题层次分析
	headingLevels := make(map[int]int)
	for _, chunk := range chunks {
		if chunk.Type == "heading" {
			headingLevels[chunk.Level]++
		}
	}

	fmt.Printf("  标题层次分布:\n")
	for level := 1; level <= 6; level++ {
		if count, exists := headingLevels[level]; exists {
			fmt.Printf("    H%d: %d 个\n", level, count)
		}
	}

	// 内容长度分析
	fmt.Printf("  内容长度分析:\n")
	var totalContentLength, totalTextLength int
	var minLength, maxLength int = 999999, 0

	for _, chunk := range chunks {
		contentLen := len(chunk.Content)
		textLen := len(chunk.Text)

		totalContentLength += contentLen
		totalTextLength += textLen

		if contentLen < minLength {
			minLength = contentLen
		}
		if contentLen > maxLength {
			maxLength = contentLen
		}
	}

	fmt.Printf("    平均内容长度: %.2f 字符\n", float64(totalContentLength)/float64(len(chunks)))
	fmt.Printf("    平均文本长度: %.2f 字符\n", float64(totalTextLength)/float64(len(chunks)))
	fmt.Printf("    最短块: %d 字符\n", minLength)
	fmt.Printf("    最长块: %d 字符\n", maxLength)

	// 生成内容摘要JSON（前5个块）
	fmt.Printf("  内容摘要 (前5个块):\n")
	summary := make([]map[string]interface{}, 0)
	for i, chunk := range chunks {
		if i >= 5 {
			break
		}

		chunkSummary := map[string]interface{}{
			"id":           chunk.ID,
			"type":         chunk.Type,
			"level":        chunk.Level,
			"content_size": len(chunk.Content),
			"text_size":    len(chunk.Text),
			"position": map[string]int{
				"start_line": chunk.Position.StartLine,
				"end_line":   chunk.Position.EndLine,
			},
			"hash":         chunk.Hash[:16],
			"links_count":  len(chunk.Links),
			"images_count": len(chunk.Images),
		}

		// 添加关键元数据
		if chunk.Type == "heading" {
			chunkSummary["heading_level"] = chunk.Metadata["heading_level"]
		} else if chunk.Type == "code" {
			chunkSummary["language"] = chunk.Metadata["language"]
			chunkSummary["line_count"] = chunk.Metadata["line_count"]
		} else if chunk.Type == "table" {
			chunkSummary["rows"] = chunk.Metadata["rows"]
			chunkSummary["columns"] = chunk.Metadata["columns"]
		}

		summary = append(summary, chunkSummary)
	}

	jsonData, _ := json.MarshalIndent(summary, "    ", "  ")
	fmt.Printf("    %s\n", string(jsonData))
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
