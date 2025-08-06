package main

import (
	"fmt"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	markdown := `# My Document

This is a paragraph with some content.

## Code Example

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

| Column 1 | Column 2 |
|----------|----------|
| Value 1  | Value 2  |
`

	chunker := mc.NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		panic(err)
	}

	for _, chunk := range chunks {
		fmt.Printf("Type: %s, Content: %s\n", chunk.Type, chunk.Text)
	}
}
