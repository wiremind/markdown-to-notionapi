# md2notion

A production-ready Go CLI tool that converts Markdown to Notion blocks and uploads them to Notion pages via the public REST API.

## Features

- **Zero dependencies**: Direct HTTP calls to Notion API, no external SDK required
- **CommonMark + GitHub flavored Markdown**: Support for headings, paragraphs, lists, code blocks, quotes, images, and more
- **Flexible operations**: Append to existing pages, replace content, or create new pages
- **Robust error handling**: Rate limiting, retry logic with exponential backoff, and clear error messages
- **Developer-friendly**: Dry-run mode, verbose logging, stdin support

## Installation

```bash
go install github.com/wiremind/markdown-to-notionapi/cmd/notion-md@latest
```

## Quick Start

1. **Create a Notion Integration**:
   - Go to [https://www.notion.so/my-integrations](https://www.notion.so/my-integrations)
   - Click "New integration" and give it a name
   - Copy the "Internal Integration Token"

2. **Share your page with the integration**:
   - Open your Notion page
   - Click "Share" → "Invite" → Select your integration

3. **Set your token**:
   ```bash
   export NOTION_TOKEN="secret_abc123..."
   ```

4. **Upload Markdown**:
   ```bash
   notion-md --page-id YOUR_PAGE_ID --md document.md
   ```

## Usage Examples

### Append to existing page
```bash
notion-md --page-id abc123def456 --md notes.md
```

### Read from stdin
```bash
cat README.md | notion-md --page-id abc123def456
```

### Replace page content
```bash
notion-md --page-id abc123def456 --md document.md --replace
```

### Create new page
```bash
notion-md --create --parent-id xyz789 --title "My Document" --md notes.md
```

### Dry run (preview JSON)
```bash
notion-md --page-id abc123def456 --md notes.md --dry-run
```

### Handle relative images
```bash
notion-md --page-id abc123def456 --md notes.md --image-base-url "https://example.com/assets/"
```

## Supported Markdown

| Markdown | Notion Block |
|----------|--------------|
| `# ## ###` | heading_1/2/3 |
| Paragraphs | paragraph |
| **bold**, *italic*, `code` | Rich text formatting |
| ~~strikethrough~~ | Rich text formatting |
| [links](url) | Rich text links |
| `- bulleted lists` | bulleted_list_item |
| `1. numbered lists` | numbered_list_item |
| Nested lists | Nested list items |
| `> blockquotes` | quote |
| ` ```code blocks``` ` | code |
| `---` horizontal rules | divider |
| `![images](url)` | image (external URLs only) |

## Command Line Options

```
Usage: notion-md [options]

Options:
  --page-id string         Notion page ID to append blocks to (required unless --create is set)
  --parent-id string       Parent page ID for creating new page (required if --create is set)
  --title string           Title for new page (required if --create is set)
  --md string              Path to markdown file (default: read from stdin)
  --append                 Append blocks to the end of the page (default true)
  --replace                Replace existing page content
  --create                 Create a new page
  --image-base-url string  Base URL for relative image paths
  --dry-run                Print JSON that would be sent, don't call API
  --notion-version string  Notion API version (default "2022-06-28")
  -v, --verbose            Verbose output
  --timeout duration       HTTP request timeout (default 15s)
  -h, --help              Show help

Environment Variables:
  NOTION_TOKEN    Notion integration token (required)
```

## Finding Page IDs

### From URL
If your page URL is `https://notion.so/My-Page-abc123def456`, the page ID is `abc123def456`.

### From API
You can also use the Notion API to list pages in a database or workspace.

## Error Handling

The tool provides clear, actionable error messages:

- **Authentication errors**: Check your `NOTION_TOKEN`
- **Permission errors**: Ensure the integration has access to the page
- **Rate limiting**: Automatically handled with exponential backoff
- **Invalid page ID**: Verify the page ID format and permissions

## Limitations

- **Images**: Only external URLs are supported (no binary upload)
- **Tables**: Not supported (would require complex Notion table structure)
- **Advanced formatting**: Some complex Markdown features may not translate perfectly

## Development

### Quick Setup
```bash
git clone https://github.com/wiremind/markdown-to-notionapi.git
cd markdown-to-notionapi
make dev-setup  # Install development tools
make build      # Build the binary
```

### Available Commands
```bash
make help           # Show all available commands
make test           # Run tests
make coverage       # Run tests with coverage report
make lint           # Run golangci-lint
make security       # Run security scans
make check          # Run all quality checks
make build          # Build binary
make build-all      # Build for all platforms
make container-build # Build container image
make clean          # Clean build artifacts
```

### Running Tests
```bash
# Unit tests
make test

# With coverage
make coverage

# All quality checks
make check
```

### Local Testing
```bash
# Set up test environment
export NOTION_TOKEN="your-integration-token"
export TEST_PAGE_ID="your-test-page-id"

# Test locally
make run-local

# Test container
make container-test
```

### GitHub Actions CI
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.25'
    - run: go test ./...
    - run: go vet ./...
    - run: go build ./cmd/notion-md
```

## Docker Usage

### Pull and Run
```bash
# Pull the latest image
docker pull ghcr.io/wiremind/markdown-to-notionapi:latest

# Run with a file
docker run --rm -v $(pwd):/data -e NOTION_TOKEN=$NOTION_TOKEN \
  ghcr.io/wiremind/markdown-to-notionapi:latest \
  -page-id=YOUR_PAGE_ID -md=/data/document.md

# Run with stdin
cat README.md | docker run --rm -i -e NOTION_TOKEN=$NOTION_TOKEN \
  ghcr.io/wiremind/markdown-to-notionapi:latest \
  -page-id=YOUR_PAGE_ID
```

### Build Locally
```bash
# Using Docker
docker build -f Containerfile -t md2notion .

# Using Podman (also works with Containerfile)
podman build -f Containerfile -t md2notion .

# Run the container
docker run --rm -e NOTION_TOKEN=$NOTION_TOKEN md2notion -page-id=YOUR_PAGE_ID -dry-run
```

## GitHub Actions Integration

### Using Docker Image Directly
```yaml
name: Upload to Notion
on: [push]

jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Upload README to Notion
      uses: docker://ghcr.io/wiremind/markdown-to-notionapi:latest
      env:
        NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}
      with:
        args: -page-id=${{ secrets.NOTION_PAGE_ID }} -verbose README.md
```

### Using Custom Action
```yaml
name: Upload to Notion
on: [push]

jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Upload documentation
      uses: wiremind/markdown-to-notionapi@main
      with:
        page-id: ${{ secrets.NOTION_PAGE_ID }}
        markdown-file: 'docs/api.md'
        dry-run: 'false'
        verbose: 'true'
      env:
        NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}
    
    - name: Upload inline content
      uses: wiremind/markdown-to-notionapi@main
      with:
        page-id: ${{ secrets.NOTION_PAGE_ID }}
        markdown-content: |
          # Build Report
          
          **Status**: ✅ Success
          **Commit**: ${{ github.sha }}
          **Branch**: ${{ github.ref_name }}
          **Time**: ${{ github.event.head_commit.timestamp }}
      env:
        NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}
```

### Action Inputs
| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `page-id` | Notion page ID to upload to | ✅ | |
| `markdown-file` | Path to markdown file | | |
| `markdown-content` | Inline markdown content | | |
| `notion-token` | Notion token (or use env var) | | |
| `dry-run` | Only show what would be uploaded | | `false` |
| `verbose` | Enable verbose output | | `false` |
| `replace` | Replace page content instead of append | | `false` |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Troubleshooting

### "NOTION_TOKEN environment variable is required"
Set your integration token:
```bash
export NOTION_TOKEN="secret_your_token_here"
```

### "Object not found" or "Unauthorized"
- Verify the page ID is correct
- Ensure the integration has access to the page (share the page with your integration)
- Check that the token is valid and not expired

### Images not appearing
- Only external URLs (http/https) are supported
- Use `--image-base-url` for relative paths
- Verify image URLs are publicly accessible

### Rate limiting
The tool automatically handles rate limiting with exponential backoff. If you hit limits frequently, consider:
- Reducing concurrency
- Adding delays between operations
- Processing smaller batches

### Large documents
For very large Markdown files:
- The tool automatically chunks requests (100 blocks max per API call)
- Use `--verbose` to monitor progress
- Consider breaking large documents into smaller sections
