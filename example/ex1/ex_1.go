package main

import (
	"fmt"
	"strings"

	mc "github.com/kydenul/markdown-chunker"
)

// 示例用法
func main() {
	markdown := `# 数据库设计文档

这是一个关于数据库设计的文档。

## 表结构

以下是用户表的结构：

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | int | 主键 |
| name | varchar(100) | 用户名 |
| email | varchar(255) | 邮箱 |

## 示例代码

以下是创建表的 SQL 语句：

` + "```sql" + `
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE
);
` + "```" + `

## 注意事项

> **重要提示**: 请确保数据的安全性和完整性。

以下是一些重要的注意事项：

- 确保邮箱字段的唯一性
- 定期备份数据  
- 监控数据库性能

### 性能优化建议

1. 为常用查询字段创建索引
2. 定期分析查询执行计划
3. 合理设置缓存策略

---

*文档最后更新时间：2024年*`

	chunker := mc.NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("总共分块数量: %d\n\n", len(chunks))

	for _, chunk := range chunks {
		fmt.Printf("=== Chunk %d ===\n", chunk.ID)
		fmt.Printf("类型: %s\n", chunk.Type)
		if chunk.Level > 0 {
			fmt.Printf("层级: %d\n", chunk.Level)
		}
		fmt.Printf("原始内容:\n%s\n", chunk.Content)
		fmt.Printf("纯文本:\n%s\n", chunk.Text)
		fmt.Printf("元数据: %+v\n\n", chunk.Metadata)
		fmt.Println(strings.Repeat("-", 50))
	}
}
