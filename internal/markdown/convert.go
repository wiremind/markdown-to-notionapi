// internal/markdown/convert.go
package markdown

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/wiremind/markdown-to-notionapi/internal/notion"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Converter handles Markdown to Notion block conversion
type Converter struct {
	imageBaseURL string
	verbose      bool
}

// NewConverter creates a new Markdown converter
func NewConverter(imageBaseURL string, verbose bool) *Converter {
	return &Converter{
		imageBaseURL: imageBaseURL,
		verbose:      verbose,
	}
}

// Convert parses Markdown content and returns Notion blocks
func (c *Converter) Convert(markdown []byte) ([]notion.Block, error) {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(markdown))

	var blocks []notion.Block
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		block, err := c.convertNode(child, markdown)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node: %w", err)
		}
		if block != nil {
			blocks = append(blocks, *block)
		}
	}

	return blocks, nil
}

// convertNode converts a single AST node to a Notion block
func (c *Converter) convertNode(node ast.Node, source []byte) (*notion.Block, error) {
	switch n := node.(type) {
	case *ast.Heading:
		return c.convertHeading(n, source)
	case *ast.Paragraph:
		return c.convertParagraph(n, source)
	case *ast.List:
		return c.convertList(n, source)
	case *ast.Blockquote:
		return c.convertBlockquote(n, source)
	case *ast.CodeBlock:
		return c.convertCodeBlock(n, source)
	case *ast.FencedCodeBlock:
		return c.convertFencedCodeBlock(n, source)
	case *ast.ThematicBreak:
		return c.convertThematicBreak()
	case *ast.HTMLBlock:
		// Skip HTML blocks for simplicity
		return nil, nil
	default:
		// Skip unknown node types
		return nil, nil
	}
}

// convertHeading converts heading nodes
func (c *Converter) convertHeading(node *ast.Heading, source []byte) (*notion.Block, error) {
	richText, err := c.convertInlineNodes(node, source)
	if err != nil {
		return nil, err
	}

	heading := &notion.Heading{RichText: richText}
	block := &notion.Block{Object: "block"}

	switch node.Level {
	case 1:
		block.Type = "heading_1"
		block.Heading1 = heading
	case 2:
		block.Type = "heading_2"
		block.Heading2 = heading
	case 3:
		block.Type = "heading_3"
		block.Heading3 = heading
	default:
		// Default to heading_3 for h4+ levels
		block.Type = "heading_3"
		block.Heading3 = heading
	}

	return block, nil
}

// convertParagraph converts paragraph nodes
func (c *Converter) convertParagraph(node *ast.Paragraph, source []byte) (*notion.Block, error) {
	// Check if this paragraph contains only an image
	if node.ChildCount() == 1 {
		if img, ok := node.FirstChild().(*ast.Image); ok {
			return c.convertImage(img, source)
		}
	}

	richText, err := c.convertInlineNodes(node, source)
	if err != nil {
		return nil, err
	}

	// Skip empty paragraphs
	if len(richText) == 0 {
		return nil, nil
	}

	return &notion.Block{
		Object:    "block",
		Type:      "paragraph",
		Paragraph: &notion.Paragraph{RichText: richText},
	}, nil
}

// convertList converts list nodes
func (c *Converter) convertList(node *ast.List, source []byte) (*notion.Block, error) {
	// Lists in Notion are handled as individual list items, not as containers
	// We'll return the first item and let the caller handle siblings
	if node.FirstChild() == nil {
		return nil, nil
	}

	firstItem := node.FirstChild().(*ast.ListItem)
	return c.convertListItem(firstItem, node.IsOrdered(), source)
}

// convertListItem converts a single list item
func (c *Converter) convertListItem(node *ast.ListItem, isOrdered bool, source []byte) (*notion.Block, error) {
	richText, err := c.convertInlineNodes(node, source)
	if err != nil {
		return nil, err
	}

	// Handle nested lists
	var children []notion.Block
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			for item := list.FirstChild(); item != nil; item = item.NextSibling() {
				if listItem, ok := item.(*ast.ListItem); ok {
					childBlock, err := c.convertListItem(listItem, list.IsOrdered(), source)
					if err != nil {
						return nil, err
					}
					if childBlock != nil {
						children = append(children, *childBlock)
					}
				}
			}
		}
	}

	block := &notion.Block{Object: "block"}

	if isOrdered {
		block.Type = "numbered_list_item"
		block.NumberedListItem = &notion.NumberedListItem{
			RichText: richText,
			Children: children,
		}
	} else {
		block.Type = "bulleted_list_item"
		block.BulletedListItem = &notion.BulletedListItem{
			RichText: richText,
			Children: children,
		}
	}

	return block, nil
}

// convertBlockquote converts blockquote nodes
func (c *Converter) convertBlockquote(node *ast.Blockquote, source []byte) (*notion.Block, error) {
	richText, err := c.convertInlineNodes(node, source)
	if err != nil {
		return nil, err
	}

	return &notion.Block{
		Object: "block",
		Type:   "quote",
		Quote:  &notion.Quote{RichText: richText},
	}, nil
}

// convertCodeBlock converts indented code blocks
func (c *Converter) convertCodeBlock(node *ast.CodeBlock, source []byte) (*notion.Block, error) {
	var content strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		content.Write(line.Value(source))
	}

	richText := []notion.RichText{{
		Type: "text",
		Text: &notion.Text{Content: content.String()},
	}}

	return &notion.Block{
		Object: "block",
		Type:   "code",
		Code: &notion.Code{
			RichText: richText,
			Language: "plain text",
			Caption:  []notion.RichText{},
		},
	}, nil
}

// convertFencedCodeBlock converts fenced code blocks
func (c *Converter) convertFencedCodeBlock(node *ast.FencedCodeBlock, source []byte) (*notion.Block, error) {
	var content strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		content.Write(line.Value(source))
	}

	var language string
	if node.Info != nil {
		language = string(node.Info.Text(source))
	}

	// Map common language names to Notion's expected values
	language = c.mapLanguage(language)

	richText := []notion.RichText{{
		Type: "text",
		Text: &notion.Text{Content: content.String()},
	}}

	return &notion.Block{
		Object: "block",
		Type:   "code",
		Code: &notion.Code{
			RichText: richText,
			Language: language,
			Caption:  []notion.RichText{},
		},
	}, nil
}

// convertThematicBreak converts horizontal rules
func (c *Converter) convertThematicBreak() (*notion.Block, error) {
	return &notion.Block{
		Object:  "block",
		Type:    "divider",
		Divider: &notion.Divider{},
	}, nil
}

// convertImage converts image nodes
func (c *Converter) convertImage(node *ast.Image, source []byte) (*notion.Block, error) {
	src := string(node.Destination)

	// Handle relative URLs
	if !c.isAbsoluteURL(src) {
		if c.imageBaseURL == "" {
			// Skip relative images without base URL
			return nil, nil
		}
		src = c.joinURL(c.imageBaseURL, src)
	}

	// Get alt text as caption
	var caption []notion.RichText
	if node.FirstChild() != nil {
		if text, ok := node.FirstChild().(*ast.Text); ok {
			caption = []notion.RichText{{
				Type: "text",
				Text: &notion.Text{Content: string(text.Segment.Value(source))},
			}}
		} else if str, ok := node.FirstChild().(*ast.String); ok {
			caption = []notion.RichText{{
				Type: "text",
				Text: &notion.Text{Content: string(str.Value)},
			}}
		}
	}

	return &notion.Block{
		Object: "block",
		Type:   "image",
		Image: &notion.Image{
			Type:     "external",
			External: &notion.External{URL: src},
			Caption:  caption,
		},
	}, nil
}

// convertInlineNodes converts child nodes to rich text
func (c *Converter) convertInlineNodes(parent ast.Node, source []byte) ([]notion.RichText, error) {
	var richText []notion.RichText

	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		texts, err := c.convertInlineNode(child, source)
		if err != nil {
			return nil, err
		}
		richText = append(richText, texts...)
	}

	return richText, nil
}

// convertInlineNode converts a single inline node
func (c *Converter) convertInlineNode(node ast.Node, source []byte) ([]notion.RichText, error) {
	switch n := node.(type) {
	case *ast.String:
		return []notion.RichText{{
			Type: "text",
			Text: &notion.Text{Content: string(n.Value)},
		}}, nil

	case *ast.Text:
		return []notion.RichText{{
			Type: "text",
			Text: &notion.Text{Content: string(n.Segment.Value(source))},
		}}, nil

	case *ast.CodeSpan:
		return []notion.RichText{{
			Type:        "text",
			Text:        &notion.Text{Content: string(n.Text(source))},
			Annotations: &notion.Annotations{Code: true},
		}}, nil

	case *ast.Emphasis:
		texts, err := c.convertInlineNodes(n, source)
		if err != nil {
			return nil, err
		}
		// Apply formatting based on emphasis level (1=italic, 2=bold)
		for i := range texts {
			if texts[i].Annotations == nil {
				texts[i].Annotations = &notion.Annotations{}
			}
			if n.Level == 2 {
				texts[i].Annotations.Bold = true
			} else {
				texts[i].Annotations.Italic = true
			}
		}
		return texts, nil

	case *ast.Link:
		texts, err := c.convertInlineNodes(n, source)
		if err != nil {
			return nil, err
		}
		// Apply link
		href := string(n.Destination)
		for i := range texts {
			texts[i].Href = &href
		}
		return texts, nil

	case *ast.AutoLink:
		href := string(n.URL(source))
		return []notion.RichText{{
			Type: "text",
			Text: &notion.Text{Content: href},
			Href: &href,
		}}, nil

	case *ast.Image:
		// Images in inline context are skipped (handled at paragraph level)
		return nil, nil

	default:
		// For other inline elements, try to extract text content
		if n.HasChildren() {
			return c.convertInlineNodes(n, source)
		}
		return nil, nil
	}
}

// mapLanguage maps common language identifiers to Notion's expected values
func (c *Converter) mapLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Common mappings
	switch lang {
	case "js", "javascript":
		return "javascript"
	case "ts", "typescript":
		return "typescript"
	case "py", "python":
		return "python"
	case "go", "golang":
		return "go"
	case "sh", "bash", "shell":
		return "bash"
	case "yml", "yaml":
		return "yaml"
	case "json":
		return "json"
	case "xml":
		return "xml"
	case "html":
		return "html"
	case "css":
		return "css"
	case "sql":
		return "sql"
	case "md", "markdown":
		return "markdown"
	case "":
		return "plain text"
	default:
		return lang
	}
}

// isAbsoluteURL checks if a URL is absolute
func (c *Converter) isAbsoluteURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.IsAbs()
}

// joinURL joins a base URL with a relative path
func (c *Converter) joinURL(baseURL, relativePath string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return relativePath
	}

	u.Path = path.Join(u.Path, relativePath)
	return u.String()
}
