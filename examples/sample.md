# Sample Document

This is a **sample Markdown document** to demonstrate the `md2notion` tool's capabilities.

## Features Demonstration

### Text Formatting

You can use *italic text*, **bold text**, `inline code`, and even ~~strikethrough~~ text. 

Here's a paragraph with a [link to Notion](https://notion.so) embedded within it.

### Lists

Bulleted lists work great:
- First item
- Second item with **bold text**
- Third item
  - Nested item
  - Another nested item
- Back to top level

Numbered lists are also supported:
1. First numbered item
2. Second item with *italic text*
3. Third item
   1. Nested numbered item
   2. Another nested item
4. Final top-level item

### Code Blocks

Here's a fenced code block with syntax highlighting:

```javascript
function greetUser(name) {
    console.log(`Hello, ${name}!`);
    return `Welcome to md2notion, ${name}`;
}

greetUser("Developer");
```

And here's a Python example:

```python
def calculate_fibonacci(n):
    if n <= 1:
        return n
    return calculate_fibonacci(n-1) + calculate_fibonacci(n-2)

# Calculate the 10th Fibonacci number
result = calculate_fibonacci(10)
print(f"The 10th Fibonacci number is: {result}")
```

### Blockquotes

> This is a blockquote. It can be used for highlighting important information,
> quotes from other sources, or just to break up the visual flow of your document.
>
> Blockquotes can span multiple paragraphs and can contain **formatted text** too.

### Images

Here's an example of an external image:

![Notion Logo](https://www.notion.so/images/logo-ios.png)

### Horizontal Rules

You can use horizontal rules to separate sections:

---

## Advanced Features

### Mixed Content

This section demonstrates how different Markdown elements work together:

1. **Configuration Setup**
   - Set your `NOTION_TOKEN` environment variable
   - Find your page ID from the Notion URL
   - Run the command: `notion-md --page-id YOUR_ID --md sample.md`

2. **Code Example**
   ```bash
   # Export your token
   export NOTION_TOKEN="secret_abc123..."
   
   # Upload this sample file
   notion-md --page-id abc123def456 --md examples/sample.md --verbose
   ```

3. **Results**
   - Your Markdown content appears in Notion
   - Formatting is preserved
   - Images and links work correctly

### Tips and Tricks

> **Pro Tip**: Use the `--dry-run` flag to preview the JSON that will be sent to Notion before actually uploading.

- Use `--verbose` for detailed logging
- Set `--image-base-url` for relative image paths
- Try `--replace` to completely replace page content

---

## Conclusion

This sample demonstrates the core features of `md2notion`. The tool handles:

- ✅ All heading levels (H1, H2, H3+)
- ✅ Text formatting (bold, italic, code, strikethrough)
- ✅ Lists (bulleted, numbered, nested)
- ✅ Code blocks with syntax highlighting
- ✅ Blockquotes
- ✅ Links and images
- ✅ Horizontal rules

Happy documenting! 🚀
