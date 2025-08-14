package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	kylog "github.com/kydenul/log"
	chunker "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 配置迁移示例 ===\n")

	// 示例1: 迁移旧版本配置结构
	fmt.Println("1. 迁移旧版本配置结构:")
	demonstrateLegacyConfigMigration()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例2: 迁移JSON配置
	fmt.Println("2. 迁移JSON配置:")
	demonstrateJSONConfigMigration()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例3: 检查配置版本
	fmt.Println("3. 检查配置版本:")
	demonstrateConfigVersionDetection()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例4: 带日志的迁移
	fmt.Println("4. 带日志的迁移:")
	demonstrateLoggedMigration()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 示例5: 完整的迁移工作流
	fmt.Println("5. 完整的迁移工作流:")
	demonstrateFullMigrationWorkflow()

	fmt.Println("\n=== 示例完成 ===")
}

func demonstrateLegacyConfigMigration() {
	// 创建一个旧版本配置
	legacyConfig := &chunker.LegacyChunkerConfig{
		MaxChunkSize: 1000,
		EnabledTypes: map[string]bool{
			"heading":   true,
			"paragraph": true,
			"list":      true,
			"code":      false,
			"table":     false,
		},
		LogLevel:            "INFO",
		FilterEmptyChunks:   true,
		PreserveWhitespace:  false,
		EnableObjectPooling: true,
		Version:             chunker.ConfigVersionV1,
	}

	fmt.Printf("   旧版本配置:\n")
	fmt.Printf("   - MaxChunkSize: %d\n", legacyConfig.MaxChunkSize)
	fmt.Printf("   - EnabledTypes: %v\n", legacyConfig.EnabledTypes)
	fmt.Printf("   - LogLevel: %s\n", legacyConfig.LogLevel)
	fmt.Printf("   - Version: %s\n", legacyConfig.Version)

	// 执行迁移
	result, err := chunker.MigrateConfig(legacyConfig)
	if err != nil {
		log.Printf("迁移失败: %v", err)
		return
	}

	fmt.Printf("\n   迁移结果:\n")
	fmt.Printf("   - 是否迁移: %t\n", result.Migrated)
	fmt.Printf("   - 原始版本: %s\n", result.OriginalVersion)
	fmt.Printf("   - 目标版本: %s\n", result.TargetVersion)
	fmt.Printf("   - 警告数量: %d\n", len(result.Warnings))
	fmt.Printf("   - 说明数量: %d\n", len(result.Notes))

	// 显示迁移后的配置
	fmt.Printf("\n   迁移后的配置:\n")
	fmt.Printf("   - MaxChunkSize: %d\n", result.Config.MaxChunkSize)
	fmt.Printf("   - LogLevel: %s\n", result.Config.LogLevel)
	fmt.Printf("   - 策略名称: %s\n", result.Config.ChunkingStrategy.Name)
	fmt.Printf("   - 策略包含类型: %v\n", result.Config.ChunkingStrategy.IncludeTypes)
	fmt.Printf("   - 策略最大块大小: %d\n", result.Config.ChunkingStrategy.MaxChunkSize)

	// 显示迁移说明
	if len(result.Notes) > 0 {
		fmt.Printf("\n   迁移说明:\n")
		for i, note := range result.Notes {
			fmt.Printf("   %d. %s\n", i+1, note)
		}
	}
}

func demonstrateJSONConfigMigration() {
	// 模拟从配置文件读取的JSON（旧版本格式）
	jsonConfigStr := `{
		"max_chunk_size": 1500,
		"enabled_types": {
			"heading": true,
			"paragraph": true,
			"code": true,
			"table": false,
			"list": true
		},
		"log_level": "DEBUG",
		"filter_empty_chunks": true,
		"preserve_whitespace": false,
		"memory_limit": 1048576
	}`

	fmt.Printf("   原始JSON配置:\n%s\n", jsonConfigStr)

	// 解析JSON
	var jsonConfig map[string]interface{}
	err := json.Unmarshal([]byte(jsonConfigStr), &jsonConfig)
	if err != nil {
		log.Printf("JSON解析失败: %v", err)
		return
	}

	// 检查配置版本
	version := chunker.GetConfigVersion(jsonConfig)
	fmt.Printf("   检测到的配置版本: %s\n", version)

	// 执行迁移
	result, err := chunker.MigrateConfig(jsonConfig)
	if err != nil {
		log.Printf("JSON配置迁移失败: %v", err)
		return
	}

	fmt.Printf("\n   迁移结果:\n")
	fmt.Printf("   - 是否迁移: %t\n", result.Migrated)
	fmt.Printf("   - 原始版本: %s → 目标版本: %s\n", result.OriginalVersion, result.TargetVersion)

	// 显示迁移后的策略配置
	if result.Config.ChunkingStrategy != nil {
		fmt.Printf("   - 策略名称: %s\n", result.Config.ChunkingStrategy.Name)
		fmt.Printf("   - 包含类型: %v\n", result.Config.ChunkingStrategy.IncludeTypes)
		fmt.Printf("   - 最大块大小: %d\n", result.Config.ChunkingStrategy.MaxChunkSize)
	}

	// 测试迁移后的配置
	fmt.Printf("\n   测试迁移后的配置:\n")
	testMigratedConfig(result.Config)
}

func demonstrateConfigVersionDetection() {
	configs := []interface{}{
		// V1配置（旧版本）
		&chunker.LegacyChunkerConfig{
			MaxChunkSize: 800,
			Version:      chunker.ConfigVersionV1,
		},
		// V2配置（新版本）
		&chunker.ChunkerConfig{
			MaxChunkSize:     1000,
			ChunkingStrategy: chunker.ElementLevelConfig(),
		},
		// JSON配置（V1格式）
		map[string]interface{}{
			"max_chunk_size": 1200,
			"log_level":      "INFO",
		},
		// JSON配置（V2格式）
		map[string]interface{}{
			"max_chunk_size": 1400,
			"chunking_strategy": map[string]interface{}{
				"name": "hierarchical",
			},
		},
	}

	for i, config := range configs {
		version := chunker.GetConfigVersion(config)
		isLegacy := chunker.IsLegacyConfig(config)

		fmt.Printf("   配置 %d:\n", i+1)
		fmt.Printf("   - 类型: %T\n", config)
		fmt.Printf("   - 版本: %s\n", version)
		fmt.Printf("   - 是否为旧版本: %t\n", isLegacy)
		fmt.Println()
	}
}

func demonstrateLoggedMigration() {
	// 创建日志器
	logger := kylog.NewLogger(&kylog.Options{
		Level:      "info",
		Format:     "console",
		Directory:  "./logs",
		TimeLayout: "2006-01-02 15:04:05.000",
	})

	// 创建需要迁移的配置
	legacyConfig := &chunker.LegacyChunkerConfig{
		MaxChunkSize: 2000,
		EnabledTypes: map[string]bool{
			"heading":   true,
			"paragraph": true,
		},
		LogLevel: "WARN",
	}

	fmt.Printf("   执行带日志的配置迁移...\n")

	// 执行带日志的迁移
	result, err := chunker.MigrateConfigWithLogger(legacyConfig, logger)
	if err != nil {
		log.Printf("带日志的迁移失败: %v", err)
		return
	}

	fmt.Printf("   迁移完成:\n")
	fmt.Printf("   - 迁移状态: %t\n", result.Migrated)
	fmt.Printf("   - 版本变化: %s → %s\n", result.OriginalVersion, result.TargetVersion)
	fmt.Printf("   - 警告数量: %d\n", len(result.Warnings))
	fmt.Printf("   - 说明数量: %d\n", len(result.Notes))
}

func demonstrateFullMigrationWorkflow() {
	fmt.Printf("   完整的迁移工作流演示:\n\n")

	// 步骤1: 模拟从文件读取旧配置
	fmt.Printf("   步骤1: 读取旧配置文件\n")
	oldConfigJSON := `{
		"max_chunk_size": 1800,
		"enabled_types": {
			"heading": true,
			"paragraph": true,
			"code": true
		},
		"log_level": "INFO",
		"filter_empty_chunks": true
	}`

	var oldConfig map[string]interface{}
	json.Unmarshal([]byte(oldConfigJSON), &oldConfig)
	fmt.Printf("   - 配置类型: JSON\n")
	fmt.Printf("   - 配置版本: %s\n", chunker.GetConfigVersion(oldConfig))

	// 步骤2: 执行迁移
	fmt.Printf("\n   步骤2: 执行配置迁移\n")
	result, err := chunker.MigrateConfig(oldConfig)
	if err != nil {
		log.Printf("迁移失败: %v", err)
		return
	}
	fmt.Printf("   - 迁移成功: %t\n", result.Migrated)

	// 步骤3: 验证迁移结果
	fmt.Printf("\n   步骤3: 验证迁移结果\n")
	if err := chunker.ValidateConfig(result.Config); err != nil {
		log.Printf("配置验证失败: %v", err)
		return
	}
	fmt.Printf("   - 配置验证: 通过\n")

	// 步骤4: 使用迁移后的配置创建分块器
	fmt.Printf("\n   步骤4: 创建分块器\n")
	markdownChunker := chunker.NewMarkdownChunkerWithConfig(result.Config)
	fmt.Printf("   - 分块器创建: 成功\n")

	// 步骤5: 测试分块功能
	fmt.Printf("\n   步骤5: 测试分块功能\n")
	testMarkdown := `# 测试文档

这是一个测试段落。

## 子标题

这是另一个段落。

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

- 列表项1
- 列表项2`

	chunks, err := markdownChunker.ChunkDocument([]byte(testMarkdown))
	if err != nil {
		log.Printf("分块失败: %v", err)
		return
	}

	fmt.Printf("   - 分块结果: %d 个块\n", len(chunks))

	// 验证类型过滤是否生效
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.Type]++
	}

	fmt.Printf("   - 块类型分布: %v\n", typeCount)

	// 验证是否只包含配置中启用的类型
	expectedTypes := result.Config.ChunkingStrategy.IncludeTypes
	fmt.Printf("   - 预期类型: %v\n", expectedTypes)

	// 步骤6: 显示迁移摘要
	fmt.Printf("\n   步骤6: 迁移摘要\n")
	fmt.Printf("   - 原始配置: V1 JSON格式\n")
	fmt.Printf("   - 迁移后配置: V2结构体格式\n")
	fmt.Printf("   - 策略: %s\n", result.Config.ChunkingStrategy.Name)
	fmt.Printf("   - 类型过滤: %v\n", result.Config.ChunkingStrategy.IncludeTypes)
	fmt.Printf("   - 大小限制: %d\n", result.Config.ChunkingStrategy.MaxChunkSize)
	fmt.Printf("   - 向后兼容: 是\n")

	fmt.Printf("\n   ✅ 完整迁移工作流执行成功！\n")
}

func testMigratedConfig(config *chunker.ChunkerConfig) {
	// 创建分块器
	markdownChunker := chunker.NewMarkdownChunkerWithConfig(config)

	// 测试文档
	testDoc := `# 标题
段落内容
` + "```" + `
代码块
` + "```" + `
- 列表项`

	chunks, err := markdownChunker.ChunkDocument([]byte(testDoc))
	if err != nil {
		log.Printf("测试分块失败: %v", err)
		return
	}

	fmt.Printf("   - 测试分块: 成功，生成 %d 个块\n", len(chunks))

	// 显示块类型
	types := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		types = append(types, chunk.Type)
	}
	fmt.Printf("   - 块类型: %v\n", types)
}
