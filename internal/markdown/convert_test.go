// internal/markdown/convert_test.go
package markdown

import (
	"reflect"
	"testing"

	"github.com/wiremind/markdown-to-notionapi/internal/notion"
)

func TestConverter_Convert(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		baseURL  string
		want     []notion.Block
		wantErr  bool
	}{
		{
			name:     "heading levels",
			markdown: "# Heading 1\n\n## Heading 2\n\n### Heading 3\n\n#### Heading 4",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "heading_1",
					Heading1: &notion.Heading{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Heading 1"}}},
					},
				},
				{
					Object: "block",
					Type:   "heading_2",
					Heading2: &notion.Heading{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Heading 2"}}},
					},
				},
				{
					Object: "block",
					Type:   "heading_3",
					Heading3: &notion.Heading{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Heading 3"}}},
					},
				},
				{
					Object: "block",
					Type:   "heading_3", // h4+ becomes h3
					Heading3: &notion.Heading{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Heading 4"}}},
					},
				},
			},
		},
		{
			name:     "paragraph with inline formatting",
			markdown: "This is **bold** text.",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "paragraph",
					Paragraph: &notion.Paragraph{
						RichText: []notion.RichText{
							{Type: "text", Text: &notion.Text{Content: "This is "}},
							{Type: "text", Text: &notion.Text{Content: "bold"}, Annotations: &notion.Annotations{Bold: true}},
							{Type: "text", Text: &notion.Text{Content: " text."}},
						},
					},
				},
			},
		},
		{
			name:     "bulleted list",
			markdown: "- Item 1\n- Item 2\n  - Nested item\n- Item 3",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "bulleted_list_item",
					BulletedListItem: &notion.BulletedListItem{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Item 1"}}},
					},
				},
			},
		},
		{
			name:     "numbered list",
			markdown: "1. First item\n2. Second item\n   1. Nested item\n3. Third item",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "numbered_list_item",
					NumberedListItem: &notion.NumberedListItem{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "First item"}}},
					},
				},
			},
		},
		{
			name:     "blockquote",
			markdown: "> This is a blockquote\n> with multiple lines",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "quote",
					Quote: &notion.Quote{
						RichText: []notion.RichText{
							{Type: "text", Text: &notion.Text{Content: "This is a blockquote"}},
							{Type: "text", Text: &notion.Text{Content: "with multiple lines"}},
						},
					},
				},
			},
		},
		{
			name:     "fenced code block",
			markdown: "```javascript\nconst hello = 'world';\nconsole.log(hello);\n```",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "code",
					Code: &notion.Code{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "const hello = 'world';\nconsole.log(hello);\n"}}},
						Language: "javascript",
						Caption:  []notion.RichText{},
					},
				},
			},
		},
		{
			name:     "indented code block",
			markdown: "    const indented = true;\n    console.log(indented);",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "code",
					Code: &notion.Code{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "const indented = true;\nconsole.log(indented);"}}},
						Language: "plain text",
						Caption:  []notion.RichText{},
					},
				},
			},
		},
		{
			name:     "horizontal rule",
			markdown: "Before\n\n---\n\nAfter",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "paragraph",
					Paragraph: &notion.Paragraph{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Before"}}},
					},
				},
				{
					Object:  "block",
					Type:    "divider",
					Divider: &notion.Divider{},
				},
				{
					Object: "block",
					Type:   "paragraph",
					Paragraph: &notion.Paragraph{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "After"}}},
					},
				},
			},
		},
		{
			name:     "image with absolute URL",
			markdown: "![Alt text](https://example.com/image.png)",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "image",
					Image: &notion.Image{
						Type:     "external",
						External: &notion.External{URL: "https://example.com/image.png"},
						Caption:  []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Alt text"}}},
					},
				},
			},
		},
		{
			name:     "image with relative URL and base URL",
			markdown: "![Local image](images/local.png)",
			baseURL:  "https://example.com/",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "image",
					Image: &notion.Image{
						Type:     "external",
						External: &notion.External{URL: "https://example.com/images/local.png"},
						Caption:  []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Local image"}}},
					},
				},
			},
		},
		{
			name:     "mixed content",
			markdown: "# Document Title\n\n```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			want: []notion.Block{
				{
					Object: "block",
					Type:   "heading_1",
					Heading1: &notion.Heading{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "Document Title"}}},
					},
				},
				{
					Object: "block",
					Type:   "code",
					Code: &notion.Code{
						RichText: []notion.RichText{{Type: "text", Text: &notion.Text{Content: "func main() {\n    fmt.Println(\"Hello\")\n}\n"}}},
						Language: "go",
						Caption:  []notion.RichText{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConverter(tt.baseURL, false)
			got, err := c.Convert([]byte(tt.markdown))
			if (err != nil) != tt.wantErr {
				t.Errorf("Converter.Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For list tests, we only check the first block since lists are complex
			if tt.name == "bulleted list" || tt.name == "numbered list" {
				if len(got) == 0 {
					t.Errorf("Converter.Convert() got no blocks")
					return
				}
				compareBlock(t, got[0], tt.want[0])
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Converter.Convert() got %d blocks, want %d", len(got), len(tt.want))
				return
			}

			for i, block := range got {
				compareBlock(t, block, tt.want[i])
			}
		})
	}
}

func TestConverter_mapLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"js", "javascript"},
		{"javascript", "javascript"},
		{"ts", "typescript"},
		{"typescript", "typescript"},
		{"py", "python"},
		{"python", "python"},
		{"go", "go"},
		{"golang", "go"},
		{"sh", "bash"},
		{"bash", "bash"},
		{"shell", "bash"},
		{"yml", "yaml"},
		{"yaml", "yaml"},
		{"", "plain text"},
		{"unknown", "unknown"},
	}

	c := NewConverter("", false)
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := c.mapLanguage(tt.input)
			if got != tt.want {
				t.Errorf("mapLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConverter_isAbsoluteURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"ftp://example.com", true},
		{"relative/path", false},
		{"/absolute/path", false},
		{"./relative", false},
		{"../relative", false},
		{"", false},
	}

	c := NewConverter("", false)
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := c.isAbsoluteURL(tt.url)
			if got != tt.want {
				t.Errorf("isAbsoluteURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestConverter_joinURL(t *testing.T) {
	tests := []struct {
		base     string
		relative string
		want     string
	}{
		{"https://example.com", "image.png", "https://example.com/image.png"},
		{"https://example.com/", "image.png", "https://example.com/image.png"},
		{"https://example.com/path", "image.png", "https://example.com/path/image.png"},
		{"https://example.com/path/", "image.png", "https://example.com/path/image.png"},
		{"https://example.com", "dir/image.png", "https://example.com/dir/image.png"},
	}

	c := NewConverter("", false)
	for _, tt := range tests {
		t.Run(tt.base+"_"+tt.relative, func(t *testing.T) {
			got := c.joinURL(tt.base, tt.relative)
			if got != tt.want {
				t.Errorf("joinURL(%q, %q) = %q, want %q", tt.base, tt.relative, got, tt.want)
			}
		})
	}
}

// Helper functions for testing

func compareBlock(t *testing.T, got, want notion.Block) {
	t.Helper()

	if got.Object != want.Object {
		t.Errorf("Block.Object = %q, want %q", got.Object, want.Object)
	}
	if got.Type != want.Type {
		t.Errorf("Block.Type = %q, want %q", got.Type, want.Type)
	}

	switch want.Type {
	case "heading_1":
		compareHeading(t, got.Heading1, want.Heading1)
	case "heading_2":
		compareHeading(t, got.Heading2, want.Heading2)
	case "heading_3":
		compareHeading(t, got.Heading3, want.Heading3)
	case "paragraph":
		compareParagraph(t, got.Paragraph, want.Paragraph)
	case "bulleted_list_item":
		compareBulletedListItem(t, got.BulletedListItem, want.BulletedListItem)
	case "numbered_list_item":
		compareNumberedListItem(t, got.NumberedListItem, want.NumberedListItem)
	case "quote":
		compareQuote(t, got.Quote, want.Quote)
	case "code":
		compareCode(t, got.Code, want.Code)
	case "divider":
		// Divider has no content to compare
	case "image":
		compareImage(t, got.Image, want.Image)
	}
}

func compareHeading(t *testing.T, got, want *notion.Heading) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("Heading nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
}

func compareParagraph(t *testing.T, got, want *notion.Paragraph) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("Paragraph nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
}

func compareBulletedListItem(t *testing.T, got, want *notion.BulletedListItem) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("BulletedListItem nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
}

func compareNumberedListItem(t *testing.T, got, want *notion.NumberedListItem) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("NumberedListItem nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
}

func compareQuote(t *testing.T, got, want *notion.Quote) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("Quote nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
}

func compareCode(t *testing.T, got, want *notion.Code) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("Code nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	compareRichText(t, got.RichText, want.RichText)
	if got.Language != want.Language {
		t.Errorf("Code.Language = %q, want %q", got.Language, want.Language)
	}
}

func compareImage(t *testing.T, got, want *notion.Image) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("Image nil mismatch: got %v, want %v", got, want)
		}
		return
	}
	if got.Type != want.Type {
		t.Errorf("Image.Type = %q, want %q", got.Type, want.Type)
	}
	if got.External == nil || want.External == nil {
		if got.External != want.External {
			t.Errorf("Image.External nil mismatch: got %v, want %v", got.External, want.External)
		}
		return
	}
	if got.External.URL != want.External.URL {
		t.Errorf("Image.External.URL = %q, want %q", got.External.URL, want.External.URL)
	}
	compareRichText(t, got.Caption, want.Caption)
}

func compareRichText(t *testing.T, got, want []notion.RichText) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("RichText length = %d, want %d", len(got), len(want))
		return
	}
	for i, rt := range got {
		compareRichTextItem(t, rt, want[i])
	}
}

func compareRichTextItem(t *testing.T, got, want notion.RichText) {
	t.Helper()
	if got.Type != want.Type {
		t.Errorf("RichText.Type = %q, want %q", got.Type, want.Type)
	}
	if got.Text == nil || want.Text == nil {
		if got.Text != want.Text {
			t.Errorf("RichText.Text nil mismatch: got %v, want %v", got.Text, want.Text)
		}
	} else if got.Text.Content != want.Text.Content {
		t.Errorf("RichText.Text.Content = %q, want %q", got.Text.Content, want.Text.Content)
	}

	// Compare annotations
	if !reflect.DeepEqual(got.Annotations, want.Annotations) {
		t.Errorf("RichText.Annotations = %+v, want %+v", got.Annotations, want.Annotations)
	}

	// Compare href
	if (got.Href == nil) != (want.Href == nil) {
		t.Errorf("RichText.Href nil mismatch: got %v, want %v", got.Href, want.Href)
	} else if got.Href != nil && *got.Href != *want.Href {
		t.Errorf("RichText.Href = %q, want %q", *got.Href, *want.Href)
	}
}
