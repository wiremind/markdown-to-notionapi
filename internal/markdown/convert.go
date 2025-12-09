// internal/markdown/convert.go
package markdown

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/wiremind/markdown-to-notionapi/internal/notion"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
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
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)
	doc := md.Parser().Parse(text.NewReader(markdown))

	var blocks []notion.Block
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		nodeBlocks, err := c.convertNode(child, markdown)
		if err != nil {
			return nil, fmt.Errorf("failed to convert node: %w", err)
		}
		blocks = append(blocks, nodeBlocks...)
	}

	return blocks, nil
}

// convertNode converts a single AST node to one or more Notion blocks
func (c *Converter) convertNode(node ast.Node, source []byte) ([]notion.Block, error) {
	switch n := node.(type) {
	case *ast.Heading:
		block, err := c.convertHeading(n, source)
		if err != nil || block == nil {
			return nil, err
		}
		return []notion.Block{*block}, nil
	case *ast.Paragraph:
		block, err := c.convertParagraph(n, source)
		if err != nil || block == nil {
			return nil, err
		}
		return []notion.Block{*block}, nil
	case *ast.List:
		block, err := c.convertList(n, source)
		if err != nil || block == nil {
			return nil, err
		}
		return []notion.Block{*block}, nil
	case *ast.Blockquote:
		block, err := c.convertBlockquote(n, source)
		if err != nil || block == nil {
			return nil, err
		}
		return []notion.Block{*block}, nil
	case *ast.CodeBlock:
		blocks, err := c.convertCodeBlock(n, source)
		if err != nil {
			return nil, err
		}
		return blocks, nil
	case *ast.FencedCodeBlock:
		blocks, err := c.convertFencedCodeBlock(n, source)
		if err != nil {
			return nil, err
		}
		return blocks, nil
	case *ast.ThematicBreak:
		block, err := c.convertThematicBreak()
		if err != nil || block == nil {
			return nil, err
		}
		return []notion.Block{*block}, nil
	case *ast.HTMLBlock:
		// Skip HTML blocks for simplicity
		return []notion.Block{}, nil
	case *extast.Table:
		return c.convertTable(n, source)
	default:
		// Skip unknown node types
		return []notion.Block{}, nil
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
func (c *Converter) convertCodeBlock(node *ast.CodeBlock, source []byte) ([]notion.Block, error) {
	var content strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		content.Write(line.Value(source))
	}

	return c.createCodeBlocks(content.String(), "plain text"), nil
}

// convertFencedCodeBlock converts fenced code blocks
func (c *Converter) convertFencedCodeBlock(node *ast.FencedCodeBlock, source []byte) ([]notion.Block, error) {
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

	return c.createCodeBlocks(content.String(), language), nil
}

// createCodeBlocks splits large code content into multiple blocks if needed
// Notion API has a 2000 character limit for code block content
func (c *Converter) createCodeBlocks(content, language string) []notion.Block {
	const (
		maxChars = 2000
		// minDistanceFromLimit defines the minimum acceptable distance from the character limit
		// when breaking at a newline. This ensures we don't create unnecessarily small chunks
		// just to break at a newline. A value of 200 means we'll accept breaking up to 200 chars
		// before the limit to preserve line integrity.
		minDistanceFromLimit = 200
	)

	// If content fits in one block, return it
	if len(content) <= maxChars {
		return []notion.Block{{
			Object: "block",
			Type:   "code",
			Code: &notion.Code{
				RichText: []notion.RichText{{
					Type: "text",
					Text: &notion.Text{Content: content},
				}},
				Language: language,
				Caption:  []notion.RichText{},
			},
		}}
	}

	// Need to split into multiple blocks
	if c.verbose {
		fmt.Fprintf(os.Stderr, "Splitting code block of %d characters into multiple blocks\n", len(content))
	}

	var blocks []notion.Block
	for len(content) > 0 {
		chunkSize := maxChars
		if len(content) < maxChars {
			chunkSize = len(content)
		}

		// Try to break at a newline to avoid splitting lines
		if chunkSize == maxChars && chunkSize < len(content) {
			// Look for the last newline before maxChars
			lastNewline := strings.LastIndex(content[:chunkSize], "\n")
			if lastNewline > 0 && lastNewline > maxChars-minDistanceFromLimit {
				// Only break at newline if it's reasonably close to maxChars
				chunkSize = lastNewline + 1
			}
		}

		chunk := content[:chunkSize]
		content = content[chunkSize:]

		blocks = append(blocks, notion.Block{
			Object: "block",
			Type:   "code",
			Code: &notion.Code{
				RichText: []notion.RichText{{
					Type: "text",
					Text: &notion.Text{Content: chunk},
				}},
				Language: language,
				Caption:  []notion.RichText{},
			},
		})
	}

	return blocks
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

	// Skip invalid or unsupported image URLs
	if !c.isValidImageURL(src) {
		if c.verbose {
			fmt.Fprintf(os.Stderr, "Skipping invalid/unsupported image URL: %s\n", src)
		}
		return nil, nil
	}

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

// convertTable converts table nodes to native Notion table blocks
func (c *Converter) convertTable(node *extast.Table, source []byte) ([]notion.Block, error) {
	var tableWidth int
	var tableRows []notion.Block
	hasColumnHeader := false

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *extast.TableHeader:
			// Handle table header row
			hasColumnHeader = true

			// Count columns from header
			if tableWidth == 0 {
				for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
					if _, ok := cell.(*extast.TableCell); ok {
						tableWidth++
					}
				}
			}

			// Convert header to table_row block
			rowBlock, err := c.convertTableHeaderToBlock(n, source)
			if err != nil {
				return nil, err
			}
			tableRows = append(tableRows, *rowBlock)

		case *extast.TableRow:
			// Handle regular table rows

			// Count columns from first row if not already done
			if tableWidth == 0 {
				for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
					if _, ok := cell.(*extast.TableCell); ok {
						tableWidth++
					}
				}
			}

			// Convert row to table_row block
			rowBlock, err := c.convertTableRowToBlock(n, source)
			if err != nil {
				return nil, err
			}
			tableRows = append(tableRows, *rowBlock)
		}
	}

	// Create the table block with table_row children inside the Table struct
	tableBlock := notion.Block{
		Object: "block",
		Type:   "table",
		Table: &notion.Table{
			TableWidth:      tableWidth,
			HasColumnHeader: hasColumnHeader,
			HasRowHeader:    false,
			Children:        tableRows, // Children go inside the Table struct
		},
	}

	// Return just the table block with rows as children
	return []notion.Block{tableBlock}, nil
}

// convertTableHeaderToBlock converts a table header to a table_row block
func (c *Converter) convertTableHeaderToBlock(header *extast.TableHeader, source []byte) (*notion.Block, error) {
	var cells [][]notion.RichText

	for cell := header.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tableCell, ok := cell.(*extast.TableCell); ok {
			cellRichText, err := c.convertTableCellToRichText(tableCell, source)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cellRichText)
		}
	}

	return &notion.Block{
		Object: "block",
		Type:   "table_row",
		TableRow: &notion.TableRow{
			Cells: cells,
		},
	}, nil
}

// convertTableRowToBlock converts a table row to a table_row block
func (c *Converter) convertTableRowToBlock(row *extast.TableRow, source []byte) (*notion.Block, error) {
	var cells [][]notion.RichText

	for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tableCell, ok := cell.(*extast.TableCell); ok {
			cellRichText, err := c.convertTableCellToRichText(tableCell, source)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cellRichText)
		}
	}

	return &notion.Block{
		Object: "block",
		Type:   "table_row",
		TableRow: &notion.TableRow{
			Cells: cells,
		},
	}, nil
}

// convertTableCellToRichText converts a table cell to rich text
func (c *Converter) convertTableCellToRichText(cell *extast.TableCell, source []byte) ([]notion.RichText, error) {
	// Use the existing inline conversion logic
	richTexts, err := c.convertInlineNodes(cell, source)
	if err != nil {
		return nil, err
	}

	// If no content, return a single empty text element
	if len(richTexts) == 0 {
		return []notion.RichText{{
			Type: "text",
			Text: &notion.Text{Content: ""},
		}}, nil
	}

	return richTexts, nil
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

// isValidImageURL checks if an image URL is valid for Notion API
func (c *Converter) isValidImageURL(urlStr string) bool {
	// Skip empty URLs
	if urlStr == "" {
		return false
	}

	// Skip Notion attachment URLs (they start with "attachment:")
	if strings.HasPrefix(urlStr, "attachment:") {
		return false
	}

	// Skip data URLs (base64 images)
	if strings.HasPrefix(urlStr, "data:") {
		return false
	}

	// For relative URLs, we'll validate them later after joining with base URL
	if !c.isAbsoluteURL(urlStr) {
		return true
	}

	// Parse the URL to ensure it's valid
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// Must be an absolute URL with http or https scheme
	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}
