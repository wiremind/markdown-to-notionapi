// cmd/notion-md/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/wiremind/markdown-to-notionapi/internal/run"
)

const (
	defaultNotionVersion = "2022-06-28"
	defaultTimeout       = 15 * time.Second
)

func main() {
	var config run.Config
	var help bool

	// Define flags
	flag.StringVar(&config.PageID, "page-id", "", "Notion page ID to append blocks to (required unless --create is set)")
	flag.StringVar(&config.ParentID, "parent-id", "", "Parent page ID for creating new page (required if --create is set)")
	flag.StringVar(&config.Title, "title", "", "Title for new page (required if --create is set)")
	flag.StringVar(&config.MarkdownFile, "md", "", "Path to markdown file (default: read from stdin)")
	flag.BoolVar(&config.Append, "append", true, "Append blocks to the end of the page")
	flag.BoolVar(&config.Replace, "replace", false, "Replace existing page content")
	flag.BoolVar(&config.Create, "create", false, "Create a new page")
	flag.StringVar(&config.ImageBaseURL, "image-base-url", "", "Base URL for relative image paths")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Print JSON that would be sent, don't call API")
	flag.StringVar(&config.NotionVersion, "notion-version", defaultNotionVersion, "Notion API version")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")
	flag.DurationVar(&config.Timeout, "timeout", defaultTimeout, "HTTP request timeout")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Convert Markdown to Notion blocks and upload to a Notion page.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s --page-id abc123 --md document.md\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  cat notes.md | %s --page-id abc123\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --create --parent-id xyz789 --title \"My Document\" --md notes.md\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  NOTION_TOKEN    Notion integration token (required)\n")
		fmt.Fprintf(os.Stderr, "\nSupported Markdown:\n")
		fmt.Fprintf(os.Stderr, "  - Headings (# ## ###)\n")
		fmt.Fprintf(os.Stderr, "  - Paragraphs with **bold**, *italic*, `code`, ~~strikethrough~~, [links](url)\n")
		fmt.Fprintf(os.Stderr, "  - Bulleted and numbered lists (including nesting)\n")
		fmt.Fprintf(os.Stderr, "  - Block quotes\n")
		fmt.Fprintf(os.Stderr, "  - Fenced code blocks ```lang\n")
		fmt.Fprintf(os.Stderr, "  - Horizontal rules ---\n")
		fmt.Fprintf(os.Stderr, "  - Images (external URLs only)\n")
	}

	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	// Handle positional arguments (first non-flag argument is the markdown file)
	if flag.NArg() > 0 && config.MarkdownFile == "" {
		config.MarkdownFile = flag.Arg(0)
	}

	// Get Notion token from environment
	notionToken := os.Getenv("NOTION_TOKEN")
	if notionToken == "" {
		fmt.Fprintf(os.Stderr, "Error: NOTION_TOKEN environment variable is required\n")
		fmt.Fprintf(os.Stderr, "Get your token from: https://www.notion.so/my-integrations\n")
		os.Exit(1)
	}

	// Create runner
	runner, err := run.NewRunner(&config, notionToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout*2) // Extra buffer for retries
	defer cancel()

	// Run the conversion and upload
	if err := runner.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
