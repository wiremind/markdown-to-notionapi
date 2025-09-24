// internal/notion/client.go
package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	NotionAPIBase = "https://api.notion.com/v1"
	MaxRetries    = 3
	BaseBackoff   = 1 * time.Second
	MaxBackoff    = 16 * time.Second
)

// Client handles Notion API interactions
type Client struct {
	httpClient *http.Client
	token      string
	version    string
	verbose    bool
}

// NewClient creates a new Notion API client
func NewClient(token, version string, timeout time.Duration, verbose bool) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		token:      token,
		version:    version,
		verbose:    verbose,
	}
}

// formatPageID formats a page ID to the proper UUID format with dashes
func (c *Client) formatPageID(pageID string) string {
	// Remove any existing dashes and lowercase
	cleaned := strings.ToLower(strings.ReplaceAll(pageID, "-", ""))

	// If it's not 32 hex characters, return as-is (might be a valid format we don't recognize)
	if len(cleaned) != 32 {
		return pageID
	}

	// Validate that all characters are hex
	for _, char := range cleaned {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return pageID
		}
	}

	// Format as UUID: 8-4-4-4-12
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		cleaned[0:8], cleaned[8:12], cleaned[12:16], cleaned[16:20], cleaned[20:32])
}

// AppendBlockChildren appends blocks to a page or block
func (c *Client) AppendBlockChildren(ctx context.Context, blockID string, blocks []Block) error {
	formattedID := c.formatPageID(blockID)

	// Split blocks into chunks of 25 for better reliability with large documents
	chunkSize := 25
	for i := 0; i < len(blocks); i += chunkSize {
		end := i + chunkSize
		if end > len(blocks) {
			end = len(blocks)
		}

		chunk := blocks[i:end]
		req := AppendBlockChildrenRequest{Children: chunk}

		if err := c.makeRequest(ctx, "PATCH", fmt.Sprintf("/blocks/%s/children", formattedID), req, nil); err != nil {
			return fmt.Errorf("failed to append blocks (chunk %d-%d): %w", i+1, end, err)
		}

		if c.verbose {
			fmt.Fprintf(os.Stderr, "Uploaded %d blocks (chunk %d-%d)\n", len(chunk), i+1, end)
		}

		// Small pause between chunks to be nice to the API
		if end < len(blocks) {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return nil
}

// CreatePage creates a new page under a parent page
func (c *Client) CreatePage(ctx context.Context, parentID, title string, blocks []Block) (*PageResponse, error) {
	formattedParentID := c.formatPageID(parentID)
	titleText := []RichText{{
		Type: "text",
		Text: &Text{Content: title},
	}}

	req := CreatePageRequest{
		Parent: Parent{
			Type:   "page_id",
			PageID: formattedParentID,
		},
		Properties: PageProperties{
			Title: TitleProperty{Title: titleText},
		},
		Children: blocks,
	}

	var resp PageResponse
	if err := c.makeRequest(ctx, "POST", "/pages", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return &resp, nil
}

// ListBlockChildren retrieves all child blocks of a block
func (c *Client) ListBlockChildren(ctx context.Context, blockID string) ([]Block, error) {
	formattedID := c.formatPageID(blockID)
	var allBlocks []Block
	var cursor *string

	for {
		url := fmt.Sprintf("/blocks/%s/children?page_size=100", formattedID)
		if cursor != nil {
			url += "&start_cursor=" + *cursor
		}

		var resp ListBlockChildrenResponse
		if err := c.makeRequest(ctx, "GET", url, nil, &resp); err != nil {
			return nil, fmt.Errorf("failed to list block children: %w", err)
		}

		allBlocks = append(allBlocks, resp.Results...)

		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}

	return allBlocks, nil
}

// DeleteBlock archives a block (soft delete)
func (c *Client) DeleteBlock(ctx context.Context, blockID string) error {
	formattedID := c.formatPageID(blockID)
	req := map[string]bool{"archived": true}
	return c.makeRequest(ctx, "PATCH", fmt.Sprintf("/blocks/%s", formattedID), req, nil)
}

// makeRequest performs an HTTP request with retry logic
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := NotionAPIBase + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)

		if c.verbose {
			fmt.Fprintf(os.Stderr, "Request: %s %s\n", method, path)
			fmt.Fprintf(os.Stderr, "Body: %s\n", string(jsonData))
		}
	}

	var lastErr error
	backoff := BaseBackoff

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			// Reset body reader for retry
			if body != nil {
				jsonData, _ := json.Marshal(body)
				reqBody = bytes.NewReader(jsonData)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Notion-Version", c.version)
		req.Header.Set("User-Agent", "md2notion/1.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			backoff = c.nextBackoff(backoff)
			continue
		}

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				// Log close error but don't override the main error
				fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
			}
		}()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			backoff = c.nextBackoff(backoff)
			continue
		}

		if c.verbose {
			fmt.Fprintf(os.Stderr, "Response: %d %s\n", resp.StatusCode, resp.Status)
			if len(respBody) > 0 {
				fmt.Fprintf(os.Stderr, "Body: %s\n", c.truncateBody(string(respBody)))
			}
		}

		// Handle rate limiting
		if resp.StatusCode == 429 {
			retryAfter := c.getRetryAfter(resp)
			if c.verbose {
				fmt.Fprintf(os.Stderr, "Rate limited, waiting %v\n", retryAfter)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryAfter):
			}
			continue
		}

		// Handle server errors with retry
		if resp.StatusCode >= 500 {
			lastErr = c.handleErrorResponse(resp.StatusCode, respBody)
			backoff = c.nextBackoff(backoff)
			continue
		}

		// Handle client errors (no retry)
		if resp.StatusCode >= 400 {
			return c.handleErrorResponse(resp.StatusCode, respBody)
		}

		// Success - parse response if needed
		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", MaxRetries+1, lastErr)
}

// getRetryAfter extracts retry delay from response headers
func (c *Client) getRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return BaseBackoff
	}

	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return BaseBackoff
}

// nextBackoff calculates the next backoff duration
func (c *Client) nextBackoff(current time.Duration) time.Duration {
	next := current * 2
	if next > MaxBackoff {
		return MaxBackoff
	}
	return next
}

// handleErrorResponse creates a user-friendly error from API response
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	var notionErr ErrorResponse
	if json.Unmarshal(body, &notionErr) == nil && notionErr.Message != "" {
		return fmt.Errorf("notion API error (%d): %s (code: %s)", statusCode, notionErr.Message, notionErr.Code)
	}

	// Fallback to raw response
	bodyPreview := c.truncateBody(string(body))
	return fmt.Errorf("notion API error (%d): %s", statusCode, bodyPreview)
}

// truncateBody truncates response body for error messages
func (c *Client) truncateBody(body string) string {
	if len(body) <= 300 {
		return body
	}
	return body[:300] + "..."
}
