package markdownchunker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// LinkExtractor 链接提取器
type LinkExtractor struct{}

// Extract 提取链接信息
func (e *LinkExtractor) Extract(node ast.Node, source []byte) map[string]string {
	metadata := make(map[string]string)
	linkCount := 0
	var links []string

	// 遍历节点查找链接
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindLink {
			link := n.(*ast.Link)
			linkURL := string(link.Destination)
			links = append(links, linkURL)
			linkCount++
		}
		return ast.WalkContinue, nil
	})

	if err == nil && linkCount > 0 {
		metadata["link_count"] = fmt.Sprintf("%d", linkCount)
		metadata["links"] = strings.Join(links, ",")
	}

	return metadata
}

// SupportedTypes 返回支持的内容类型
func (e *LinkExtractor) SupportedTypes() []string {
	return []string{"paragraph", "heading", "blockquote", "list", "code", "table"}
}

// ImageExtractor 图片提取器
type ImageExtractor struct{}

// Extract 提取图片信息
func (e *ImageExtractor) Extract(node ast.Node, source []byte) map[string]string {
	metadata := make(map[string]string)
	imageCount := 0
	var images []string

	// 遍历节点查找图片
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindImage {
			image := n.(*ast.Image)
			imageURL := string(image.Destination)
			images = append(images, imageURL)
			imageCount++
		}
		return ast.WalkContinue, nil
	})

	if err == nil && imageCount > 0 {
		metadata["image_count"] = fmt.Sprintf("%d", imageCount)
		metadata["images"] = strings.Join(images, ",")
	}

	return metadata
}

// SupportedTypes 返回支持的内容类型
func (e *ImageExtractor) SupportedTypes() []string {
	return []string{"paragraph", "heading", "blockquote", "list"}
}

// CodeComplexityExtractor 代码复杂度分析提取器
type CodeComplexityExtractor struct{}

// Extract 提取代码复杂度信息
func (e *CodeComplexityExtractor) Extract(node ast.Node, source []byte) map[string]string {
	metadata := make(map[string]string)

	// 只处理代码块
	if node.Kind() != ast.KindFencedCodeBlock && node.Kind() != ast.KindCodeBlock {
		return metadata
	}

	// 获取代码内容
	var codeContent strings.Builder
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		codeContent.Write(line.Value(source))
	}

	code := codeContent.String()

	// 基础统计
	lines := strings.Split(strings.TrimSpace(code), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	metadata["code_lines"] = fmt.Sprintf("%d", len(lines))
	metadata["code_non_empty_lines"] = fmt.Sprintf("%d", nonEmptyLines)

	// 简单的复杂度分析（基于关键字）
	complexityKeywords := []string{
		"if", "else", "for", "while", "switch", "case", "try", "catch",
		"function", "def", "class", "struct", "interface",
	}

	complexity := 0
	codeUpper := strings.ToUpper(code)
	for _, keyword := range complexityKeywords {
		// 使用正则表达式匹配单词边界
		pattern := `\b` + strings.ToUpper(keyword) + `\b`
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(codeUpper, -1)
		complexity += len(matches)
	}

	metadata["code_complexity"] = fmt.Sprintf("%d", complexity)

	return metadata
}

// SupportedTypes 返回支持的内容类型
func (e *CodeComplexityExtractor) SupportedTypes() []string {
	return []string{"code"}
}
