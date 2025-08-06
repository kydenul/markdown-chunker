package main

import (
	"fmt"

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
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
