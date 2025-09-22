// internal/run/run.go
package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/wiremind/markdown-to-notionapi/internal/markdown"
	"github.com/wiremind/markdown-to-notionapi/internal/notion"
)

// Config holds configuration for the run
type Config struct {
	PageID        string
	ParentID      string
	Title         string
	MarkdownFile  string
	Append        bool
	Replace       bool
	ImageBaseURL  string
	DryRun        bool
	NotionVersion string
	Verbose       bool
	Timeout       time.Duration
	Create        bool
}

// Runner orchestrates the conversion and upload process
type Runner struct {
	config    *Config
	client    *notion.Client
	converter *markdown.Converter
}

// NewRunner creates a new runner instance
func NewRunner(config *Config, notionToken string) (*Runner, error) {
	if notionToken == "" {
		return nil, fmt.Errorf("NOTION_TOKEN environment variable is required")
	}

	client := notion.NewClient(notionToken, config.NotionVersion, config.Timeout, config.Verbose)
	converter := markdown.NewConverter(config.ImageBaseURL, config.Verbose)

	return &Runner{
		config:    config,
		client:    client,
		converter: converter,
	}, nil
}

// Run executes the conversion and upload process
func (r *Runner) Run(ctx context.Context) error {
	// Validate configuration
	if err := r.validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Read markdown content
	content, err := r.readMarkdownContent()
	if err != nil {
		return fmt.Errorf("failed to read markdown content: %w", err)
	}

	// Convert markdown to Notion blocks
	blocks, err := r.converter.Convert(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown: %w", err)
	}

	if len(blocks) == 0 {
		fmt.Fprintf(os.Stderr, "No content to upload\n")
		return nil
	}

	if r.config.Verbose {
		fmt.Fprintf(os.Stderr, "Converted %d blocks\n", len(blocks))
	}

	// Handle dry run
	if r.config.DryRun {
		return r.printDryRun(blocks)
	}

	// Execute the appropriate operation
	if r.config.Create {
		return r.createPage(ctx, blocks)
	} else if r.config.Replace {
		return r.replacePage(ctx, blocks)
	} else {
		return r.appendToPage(ctx, blocks)
	}
}

// validateConfig validates the runner configuration
func (r *Runner) validateConfig() error {
	if r.config.Create {
		if r.config.ParentID == "" {
			return fmt.Errorf("--parent-id is required when --create is set")
		}
		if r.config.Title == "" {
			return fmt.Errorf("--title is required when --create is set")
		}
	} else {
		if r.config.PageID == "" {
			return fmt.Errorf("--page-id is required unless --create is set")
		}
	}

	if r.config.Replace && r.config.Create {
		return fmt.Errorf("--replace and --create cannot be used together")
	}

	return nil
}

// readMarkdownContent reads the markdown content from file or stdin
func (r *Runner) readMarkdownContent() ([]byte, error) {
	if r.config.MarkdownFile == "" || r.config.MarkdownFile == "-" {
		// Read from stdin
		return io.ReadAll(os.Stdin)
	}

	// Read from file
	return os.ReadFile(r.config.MarkdownFile)
}

// printDryRun prints the blocks that would be uploaded
func (r *Runner) printDryRun(blocks []notion.Block) error {
	req := notion.AppendBlockChildrenRequest{Children: blocks}
	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal blocks for dry run: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

// createPage creates a new page with the converted blocks
func (r *Runner) createPage(ctx context.Context, blocks []notion.Block) error {
	if r.config.Verbose {
		fmt.Fprintf(os.Stderr, "Creating new page '%s' under parent %s\n", r.config.Title, r.config.ParentID)
	}

	page, err := r.client.CreatePage(ctx, r.config.ParentID, r.config.Title, blocks)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	fmt.Printf("Created page: %s\n", page.URL)
	return nil
}

// replacePage replaces the content of an existing page
func (r *Runner) replacePage(ctx context.Context, blocks []notion.Block) error {
	if r.config.Verbose {
		fmt.Fprintf(os.Stderr, "Replacing content of page %s\n", r.config.PageID)
	}

	// Get existing children
	existingBlocks, err := r.client.ListBlockChildren(ctx, r.config.PageID)
	if err != nil {
		return fmt.Errorf("failed to list existing blocks: %w", err)
	}

	// Delete existing blocks
	for _, block := range existingBlocks {
		if err := r.client.DeleteBlock(ctx, block.ID); err != nil {
			if r.config.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete block: %v\n", err)
			}
		}
	}

	if len(existingBlocks) > 0 && r.config.Verbose {
		fmt.Fprintf(os.Stderr, "Deleted %d existing blocks\n", len(existingBlocks))
	}

	// Add new blocks
	if err := r.client.AppendBlockChildren(ctx, r.config.PageID, blocks); err != nil {
		return fmt.Errorf("failed to append new blocks: %w", err)
	}

	fmt.Printf("Replaced content of page: https://notion.so/%s\n", r.config.PageID)
	return nil
}

// appendToPage appends blocks to an existing page
func (r *Runner) appendToPage(ctx context.Context, blocks []notion.Block) error {
	if r.config.Verbose {
		fmt.Fprintf(os.Stderr, "Appending %d blocks to page %s\n", len(blocks), r.config.PageID)
	}

	if err := r.client.AppendBlockChildren(ctx, r.config.PageID, blocks); err != nil {
		return fmt.Errorf("failed to append blocks: %w", err)
	}

	fmt.Printf("Updated page: https://notion.so/%s\n", r.config.PageID)
	return nil
}
