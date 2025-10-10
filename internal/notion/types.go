// internal/notion/types.go
package notion

import "time"

// Block represents a Notion block structure
type Block struct {
	Object           string            `json:"object"`
	ID               string            `json:"id,omitempty"`
	Type             string            `json:"type"`
	Paragraph        *Paragraph        `json:"paragraph,omitempty"`
	Heading1         *Heading          `json:"heading_1,omitempty"`
	Heading2         *Heading          `json:"heading_2,omitempty"`
	Heading3         *Heading          `json:"heading_3,omitempty"`
	Code             *Code             `json:"code,omitempty"`
	Quote            *Quote            `json:"quote,omitempty"`
	Divider          *Divider          `json:"divider,omitempty"`
	Image            *Image            `json:"image,omitempty"`
	BulletedListItem *BulletedListItem `json:"bulleted_list_item,omitempty"`
	NumberedListItem *NumberedListItem `json:"numbered_list_item,omitempty"`
	Table            *Table            `json:"table,omitempty"`
	TableRow         *TableRow         `json:"table_row,omitempty"`
	Children         []Block           `json:"children,omitempty"`
}

// RichText represents formatted text content
type RichText struct {
	Type        string       `json:"type"`
	Text        *Text        `json:"text,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Href        *string      `json:"href,omitempty"`
}

// Text contains the actual text content
type Text struct {
	Content string `json:"content"`
	Link    *Link  `json:"link,omitempty"`
}

// Link represents a hyperlink
type Link struct {
	URL string `json:"url"`
}

// Annotations define text formatting
type Annotations struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color,omitempty"`
}

// Paragraph block type
type Paragraph struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
}

// Heading block type (shared for h1, h2, h3)
type Heading struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
}

// Code block type
type Code struct {
	Caption  []RichText `json:"caption"`
	RichText []RichText `json:"rich_text"`
	Language string     `json:"language"`
}

// Quote block type
type Quote struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
}

// Divider block type
type Divider struct{}

// Image block type
type Image struct {
	Type     string     `json:"type"`
	External *External  `json:"external,omitempty"`
	Caption  []RichText `json:"caption"`
}

// External represents an external image URL
type External struct {
	URL string `json:"url"`
}

// BulletedListItem represents a bulleted list item
type BulletedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
	Children []Block    `json:"children,omitempty"`
}

// NumberedListItem represents a numbered list item
type NumberedListItem struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
	Children []Block    `json:"children,omitempty"`
}

// AppendBlockChildrenRequest is the request body for appending blocks
type AppendBlockChildrenRequest struct {
	Children []Block `json:"children"`
}

// CreatePageRequest is the request body for creating a new page
type CreatePageRequest struct {
	Parent     Parent         `json:"parent"`
	Properties PageProperties `json:"properties"`
	Children   []Block        `json:"children,omitempty"`
}

// Parent specifies the parent of a page
type Parent struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
}

// PageProperties contains page metadata
type PageProperties struct {
	Title TitleProperty `json:"title"`
}

// TitleProperty represents a page title
type TitleProperty struct {
	Title []RichText `json:"title"`
}

// ListBlockChildrenResponse represents the response from listing block children
type ListBlockChildrenResponse struct {
	Object     string  `json:"object"`
	Results    []Block `json:"results"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

// ErrorResponse represents a Notion API error
type ErrorResponse struct {
	Object  string `json:"object"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Table represents a table block
type Table struct {
	TableWidth      int     `json:"table_width"`
	HasColumnHeader bool    `json:"has_column_header"`
	HasRowHeader    bool    `json:"has_row_header"`
	Children        []Block `json:"children"`
}

// TableRow represents a table row block
type TableRow struct {
	Cells [][]RichText `json:"cells"`
}

// PageResponse represents a created or retrieved page
type PageResponse struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	CreatedTime    time.Time      `json:"created_time"`
	LastEditedTime time.Time      `json:"last_edited_time"`
	URL            string         `json:"url"`
	Parent         Parent         `json:"parent"`
	Properties     PageProperties `json:"properties"`
}
