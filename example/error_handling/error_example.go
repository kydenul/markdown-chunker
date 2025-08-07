package main

import (
	"fmt"
	"strings"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 错误处理系统示例 ===")

	// 示例 1: 处理 nil 输入
	fmt.Println("\n1. 处理 nil 输入（宽松模式）")
	config1 := mc.DefaultConfig()
	config1.ErrorHandling = mc.ErrorModePermissive

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)
	chunks1, err := chunker1.ChunkDocument(nil)

	fmt.Printf("返回错误: %v\n", err)
	fmt.Printf("块数量: %d\n", len(chunks1))
	fmt.Printf("是否有错误: %t\n", chunker1.HasErrors())

	if chunker1.HasErrors() {
		errors := chunker1.GetErrors()
		for _, e := range errors {
			fmt.Printf("错误: %s\n", e.Error())
			fmt.Printf("错误类型: %s\n", e.Type.String())
			fmt.Printf("时间戳: %s\n", e.Timestamp.Format("2006-01-02 15:04:05"))
		}
	}

	// 示例 2: 处理 nil 输入（严格模式）
	fmt.Println("\n2. 处理 nil 输入（严格模式）")
	config2 := mc.DefaultConfig()
	config2.ErrorHandling = mc.ErrorModeStrict

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)
	_, err2 := chunker2.ChunkDocument(nil)

	if err2 != nil {
		fmt.Printf("严格模式返回错误: %s\n", err2.Error())

		// 检查是否是 ChunkerError
		if chunkerErr, ok := err2.(*mc.ChunkerError); ok {
			fmt.Printf("错误类型: %s\n", chunkerErr.Type.String())
			fmt.Printf("错误消息: %s\n", chunkerErr.Message)
		}
	}

	// 示例 3: 块大小限制错误
	fmt.Println("\n3. 块大小限制错误（宽松模式）")
	longMarkdown := `# 这是一个非常长的标题，用来测试块大小限制功能

这是一个非常长的段落，包含了很多文字内容，用来测试当内容超过配置的最大块大小时，系统如何处理这种情况。在宽松模式下，系统应该截断内容并记录错误。

## 另一个长标题用于测试

更多的长内容用于测试块大小限制功能的工作情况。`

	config3 := mc.DefaultConfig()
	config3.MaxChunkSize = 50 // 设置很小的块大小限制
	config3.ErrorHandling = mc.ErrorModePermissive

	chunker3 := mc.NewMarkdownChunkerWithConfig(config3)
	chunks3, err3 := chunker3.ChunkDocument([]byte(longMarkdown))

	fmt.Printf("返回错误: %v\n", err3)
	fmt.Printf("块数量: %d\n", len(chunks3))
	fmt.Printf("是否有错误: %t\n", chunker3.HasErrors())

	// 显示块大小限制错误
	chunkTooLargeErrors := chunker3.GetErrorsByType(mc.ErrorTypeChunkTooLarge)
	fmt.Printf("块过大错误数量: %d\n", len(chunkTooLargeErrors))

	for i, err := range chunkTooLargeErrors {
		fmt.Printf("错误 %d: %s\n", i+1, err.Error())
		fmt.Printf("  块类型: %v\n", err.Context["chunk_type"])
		fmt.Printf("  块大小: %v\n", err.Context["chunk_size"])
		fmt.Printf("  最大大小: %v\n", err.Context["max_size"])
	}

	// 显示截断后的块
	fmt.Println("\n截断后的块:")
	for i, chunk := range chunks3 {
		fmt.Printf("块 %d (%s): %s... (长度: %d)\n",
			i+1, chunk.Type,
			chunk.Text[:min(len(chunk.Text), 30)],
			len(chunk.Content))
	}

	// 示例 4: 块大小限制错误（严格模式）
	fmt.Println("\n4. 块大小限制错误（严格模式）")
	config4 := mc.DefaultConfig()
	config4.MaxChunkSize = 20 // 设置非常小的块大小限制
	config4.ErrorHandling = mc.ErrorModeStrict

	chunker4 := mc.NewMarkdownChunkerWithConfig(config4)
	_, err4 := chunker4.ChunkDocument([]byte(longMarkdown))

	if err4 != nil {
		fmt.Printf("严格模式返回错误: %s\n", err4.Error())

		if chunkerErr, ok := err4.(*mc.ChunkerError); ok {
			fmt.Printf("错误类型: %s\n", chunkerErr.Type.String())
			fmt.Printf("错误上下文:\n")
			for key, value := range chunkerErr.Context {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	// 示例 5: 静默模式
	fmt.Println("\n5. 静默模式")
	config5 := mc.DefaultConfig()
	config5.MaxChunkSize = 30
	config5.ErrorHandling = mc.ErrorModeSilent

	chunker5 := mc.NewMarkdownChunkerWithConfig(config5)
	chunks5, err5 := chunker5.ChunkDocument([]byte(longMarkdown))

	fmt.Printf("返回错误: %v\n", err5)
	fmt.Printf("块数量: %d\n", len(chunks5))
	fmt.Printf("是否有错误: %t\n", chunker5.HasErrors())
	fmt.Printf("错误数量: %d\n", len(chunker5.GetErrors()))

	// 即使在静默模式下，错误仍然被记录
	if chunker5.HasErrors() {
		fmt.Println("静默模式下记录的错误:")
		for i, err := range chunker5.GetErrors() {
			fmt.Printf("  错误 %d: %s\n", i+1, err.Type.String())
		}
	}

	// 示例 6: 配置验证错误
	fmt.Println("\n6. 配置验证错误")
	demonstrateConfigValidation()

	// 示例 7: 解析错误处理
	fmt.Println("\n7. 解析错误处理")
	demonstrateParsingErrors()

	// 示例 8: 错误恢复机制
	fmt.Println("\n8. 错误恢复机制")
	demonstrateErrorRecovery()

	// 示例 9: 自定义错误处理器
	fmt.Println("\n9. 自定义错误处理器")
	demonstrateCustomErrorHandler()
}

// demonstrateConfigValidation 演示配置验证错误
func demonstrateConfigValidation() {
	// 创建无效配置
	invalidConfig := &mc.ChunkerConfig{
		MaxChunkSize: -100, // 无效的负数
		EnabledTypes: map[string]bool{
			"invalid_type": true, // 无效的内容类型
		},
		ErrorHandling:   mc.ErrorModeStrict,
		PerformanceMode: mc.PerformanceModeDefault,
	}

	// 验证配置
	err := mc.ValidateConfig(invalidConfig)
	if err != nil {
		fmt.Printf("配置验证失败: %s\n", err.Error())
	}

	// 使用无效配置创建分块器（会自动使用默认配置）
	chunker := mc.NewMarkdownChunkerWithConfig(invalidConfig)
	testContent := "# 测试标题\n\n这是一个测试段落。"
	chunks, err := chunker.ChunkDocument([]byte(testContent))

	fmt.Printf("使用无效配置的结果:\n")
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  块数量: %d\n", len(chunks))
	fmt.Printf("  (注意: 无效配置会被替换为默认配置)\n")
}

// demonstrateParsingErrors 演示解析错误处理
func demonstrateParsingErrors() {
	// 创建格式不规范的表格，可能导致解析问题
	malformedMarkdown := `# 测试文档

| 表头1 | 表头2 |
|-------|-------|
| 数据1 | 数据2 | 多余列 |
| 数据3 |  // 缺少列

` + "```invalid_language" + `
这是一个可能有问题的代码块
` + "```" + `

> 引用块
>> 嵌套引用
>>> 深度嵌套引用`

	config := mc.DefaultConfig()
	config.ErrorHandling = mc.ErrorModePermissive

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(malformedMarkdown))

	fmt.Printf("解析格式不规范内容的结果:\n")
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  块数量: %d\n", len(chunks))
	fmt.Printf("  处理错误数量: %d\n", len(chunker.GetErrors()))

	// 显示解析错误详情
	if chunker.HasErrors() {
		parsingErrors := chunker.GetErrorsByType(mc.ErrorTypeParsingFailed)
		fmt.Printf("  解析错误: %d\n", len(parsingErrors))
		for i, err := range parsingErrors {
			fmt.Printf("    错误 %d: %s\n", i+1, err.Message)
			if len(err.Context) > 0 {
				fmt.Printf("      上下文: %+v\n", err.Context)
			}
		}
	}
}

// demonstrateErrorRecovery 演示错误恢复机制
func demonstrateErrorRecovery() {
	// 创建一个会产生多种错误的文档
	problematicContent := `# 超长标题` + strings.Repeat("这是一个非常长的标题", 20) + `

这是一个正常段落。

` + "```go" + `
func normalFunction() {
    fmt.Println("这是正常代码")
}
` + "```" + `

另一个超长段落` + strings.Repeat("内容", 100) + `

| 正常表格 | 列2 |
|----------|-----|
| 数据1    | 数据2 |

最后一个正常段落。`

	config := mc.DefaultConfig()
	config.MaxChunkSize = 50                      // 很小的限制，会导致多个错误
	config.ErrorHandling = mc.ErrorModePermissive // 宽松模式，继续处理

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(problematicContent))

	fmt.Printf("错误恢复机制测试:\n")
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  成功处理的块: %d\n", len(chunks))
	fmt.Printf("  总错误数: %d\n", len(chunker.GetErrors()))

	// 按类型统计错误
	errorTypes := make(map[mc.ErrorType]int)
	for _, err := range chunker.GetErrors() {
		errorTypes[err.Type]++
	}

	fmt.Printf("  错误类型分布:\n")
	for errorType, count := range errorTypes {
		fmt.Printf("    %s: %d\n", errorType.String(), count)
	}

	// 显示成功处理的块类型
	blockTypes := make(map[string]int)
	for _, chunk := range chunks {
		blockTypes[chunk.Type]++
	}

	fmt.Printf("  成功处理的块类型:\n")
	for blockType, count := range blockTypes {
		fmt.Printf("    %s: %d\n", blockType, count)
	}
}

// CustomErrorHandler 自定义错误处理器示例
type CustomErrorHandler struct {
	errors    []mc.ChunkerError
	maxErrors int
	callback  func(*mc.ChunkerError)
}

// NewCustomErrorHandler 创建自定义错误处理器
func NewCustomErrorHandler(maxErrors int, callback func(*mc.ChunkerError)) *CustomErrorHandler {
	return &CustomErrorHandler{
		errors:    make([]mc.ChunkerError, 0),
		maxErrors: maxErrors,
		callback:  callback,
	}
}

// HandleError 处理错误
func (h *CustomErrorHandler) HandleError(err *mc.ChunkerError) error {
	// 记录错误
	h.errors = append(h.errors, *err)

	// 调用回调函数
	if h.callback != nil {
		h.callback(err)
	}

	// 如果错误数量超过限制，返回错误
	if len(h.errors) >= h.maxErrors {
		return fmt.Errorf("错误数量超过限制 (%d)", h.maxErrors)
	}

	return nil
}

// GetErrors 获取所有错误
func (h *CustomErrorHandler) GetErrors() []*mc.ChunkerError {
	errors := make([]*mc.ChunkerError, len(h.errors))
	for i := range h.errors {
		errors[i] = &h.errors[i]
	}
	return errors
}

// ClearErrors 清除所有错误
func (h *CustomErrorHandler) ClearErrors() {
	h.errors = h.errors[:0]
}

// HasErrors 检查是否有错误
func (h *CustomErrorHandler) HasErrors() bool {
	return len(h.errors) > 0
}

// demonstrateCustomErrorHandler 演示自定义错误处理器
func demonstrateCustomErrorHandler() {
	// 创建自定义错误处理器
	errorCallback := func(err *mc.ChunkerError) {
		fmt.Printf("    [回调] 捕获错误: %s - %s\n", err.Type.String(), err.Message)
	}

	customHandler := NewCustomErrorHandler(3, errorCallback) // 最多允许3个错误

	// 注意: 这里我们无法直接设置自定义错误处理器，因为当前API不支持
	// 这个示例展示了如何实现自定义错误处理器的概念
	fmt.Printf("自定义错误处理器示例:\n")
	fmt.Printf("  最大错误数: %d\n", customHandler.maxErrors)
	fmt.Printf("  当前错误数: %d\n", len(customHandler.GetErrors()))

	// 模拟错误处理
	testErrors := []*mc.ChunkerError{
		mc.NewChunkerError(mc.ErrorTypeChunkTooLarge, "块过大", nil),
		mc.NewChunkerError(mc.ErrorTypeParsingFailed, "解析失败", nil),
		mc.NewChunkerError(mc.ErrorTypeInvalidInput, "输入无效", nil),
		mc.NewChunkerError(mc.ErrorTypeMemoryExhausted, "内存不足", nil),
	}

	fmt.Printf("  模拟处理错误:\n")
	for i, err := range testErrors {
		handlerErr := customHandler.HandleError(err)
		if handlerErr != nil {
			fmt.Printf("    处理错误 %d 时失败: %s\n", i+1, handlerErr.Error())
			break
		}
	}

	fmt.Printf("  最终错误数: %d\n", len(customHandler.GetErrors()))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
