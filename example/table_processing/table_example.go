package main

import (
	"fmt"
	"log"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 高级表格处理示例 ===")

	// 示例 1: 标准表格
	fmt.Println("\n1. 标准表格处理")
	standardTable := `| 产品 | 价格 | 库存 | 网址 |
|------|------|------|------|
| 小工具 | 19.99 | true | https://shop.com/widget |
| 大工具 | 29.50 | false | https://shop.com/gadget |
| 工具箱 | 15.00 | yes | https://shop.com/toolbox |`

	chunker := mc.NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(standardTable))
	if err != nil {
		log.Fatal(err)
	}

	if len(chunks) > 0 {
		chunk := chunks[0]
		fmt.Printf("表格类型: %s\n", chunk.Type)
		fmt.Printf("行数: %s\n", chunk.Metadata["rows"])
		fmt.Printf("列数: %s\n", chunk.Metadata["columns"])
		fmt.Printf("有表头: %s\n", chunk.Metadata["has_header"])
		fmt.Printf("格式良好: %s\n", chunk.Metadata["is_well_formed"])

		if cellTypes, exists := chunk.Metadata["cell_types"]; exists {
			fmt.Printf("单元格类型: %s\n", cellTypes)
		}

		if alignments, exists := chunk.Metadata["alignments"]; exists {
			fmt.Printf("对齐方式: %s\n", alignments)
		}
	}

	// 示例 2: 带对齐的表格
	fmt.Println("\n2. 带对齐的表格")
	alignedTable := `| 左对齐 | 居中 | 右对齐 |
|:-------|:----:|-------:|
| L1     | C1   | R1     |
| L2     | C2   | R2     |`

	chunks2, err := chunker.ChunkDocument([]byte(alignedTable))
	if err != nil {
		log.Fatal(err)
	}

	if len(chunks2) > 0 {
		chunk := chunks2[0]
		fmt.Printf("表格类型: %s\n", chunk.Type)
		fmt.Printf("行数: %s\n", chunk.Metadata["rows"])
		fmt.Printf("列数: %s\n", chunk.Metadata["columns"])

		if alignments, exists := chunk.Metadata["alignments"]; exists {
			fmt.Printf("对齐方式: %s\n", alignments)
		}
	}

	// 示例 3: 格式不规范的表格
	fmt.Println("\n3. 格式不规范的表格")
	malformedTable := `| 姓名 | 年龄 |
|------|-----|
| 张三 | 25  | 多余列 |
| 李四 |`

	chunks3, err := chunker.ChunkDocument([]byte(malformedTable))
	if err != nil {
		log.Fatal(err)
	}

	if len(chunks3) > 0 {
		chunk := chunks3[0]
		fmt.Printf("表格类型: %s\n", chunk.Type)
		fmt.Printf("行数: %s\n", chunk.Metadata["rows"])
		fmt.Printf("列数: %s\n", chunk.Metadata["columns"])
		fmt.Printf("格式良好: %s\n", chunk.Metadata["is_well_formed"])

		if errors, exists := chunk.Metadata["errors"]; exists {
			fmt.Printf("错误信息: %s\n", errors)
		}

		if errorCount, exists := chunk.Metadata["error_count"]; exists {
			fmt.Printf("错误数量: %s\n", errorCount)
		}
	}

	// 示例 4: 复杂数据类型表格
	fmt.Println("\n4. 复杂数据类型表格")
	complexTable := `| 用户 | 邮箱 | 注册日期 | 活跃 | 积分 |
|------|------|----------|------|------|
| 张三 | zhang@example.com | 2023-01-15 | true | 1250 |
| 李四 | li@test.org | 2023-02-20 | false | 890.5 |
| 王五 | wang@demo.net | 2023-03-10 | yes | 2100 |`

	chunks4, err := chunker.ChunkDocument([]byte(complexTable))
	if err != nil {
		log.Fatal(err)
	}

	if len(chunks4) > 0 {
		chunk := chunks4[0]
		fmt.Printf("表格类型: %s\n", chunk.Type)
		fmt.Printf("行数: %s\n", chunk.Metadata["rows"])
		fmt.Printf("列数: %s\n", chunk.Metadata["columns"])
		fmt.Printf("格式良好: %s\n", chunk.Metadata["is_well_formed"])

		if cellTypes, exists := chunk.Metadata["cell_types"]; exists {
			fmt.Printf("单元格类型分析: %s\n", cellTypes)
		}
	}

	// 示例 5: 空表格
	fmt.Println("\n5. 空表格处理")
	emptyTable := `| 标题 |
|------|`

	chunks5, err := chunker.ChunkDocument([]byte(emptyTable))
	if err != nil {
		log.Fatal(err)
	}

	if len(chunks5) > 0 {
		chunk := chunks5[0]
		fmt.Printf("表格类型: %s\n", chunk.Type)
		fmt.Printf("行数: %s\n", chunk.Metadata["rows"])
		fmt.Printf("列数: %s\n", chunk.Metadata["columns"])
		fmt.Printf("格式良好: %s\n", chunk.Metadata["is_well_formed"])

		if errors, exists := chunk.Metadata["errors"]; exists {
			fmt.Printf("错误信息: %s\n", errors)
		}
	}

	// 示例 6: 检查错误处理
	fmt.Println("\n6. 错误处理检查")
	if chunker.HasErrors() {
		fmt.Printf("处理过程中发现 %d 个错误:\n", len(chunker.GetErrors()))
		for i, err := range chunker.GetErrors() {
			fmt.Printf("  错误 %d: %s\n", i+1, err.Error())
		}
	} else {
		fmt.Println("处理过程中没有发现错误")
	}
}
