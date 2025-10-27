# Built-in Tools Example

This example demonstrates how to use go-llm-agent's built-in tools.

## Features Demonstrated

- Tool registry for managing multiple tools
- File operations (read, list)
- Date/time operations
- Integration with LLM providers
- Tool execution and result handling

## Available Tools

### File Tools

- **file_read**: Read content from a file
- **file_list**: List files and directories

### DateTime Tools

- **datetime_now**: Get current date and time with formatting

## Usage

1. Set up your environment variables in `.env`:

```bash
LLM_PROVIDER=ollama
LLM_MODEL=qwen3:1.7b
OLLAMA_BASE_URL=http://localhost:11434
```

2. Run the example:

```bash
cd examples/builtin_tools
go run .
```

## Example Output

```
=== go-llm-agent Built-in Tools Demo ===

Registered 3 tools:
  - file_read (file): Read the complete content of a text file from the filesystem
  - file_list (file): List all files and directories in a specified path
  - datetime_now (datetime): Get the current date and time in a specified format and timezone

Example 1: List current directory
{
  "count": 2,
  "directory": "/path/to/examples/builtin_tools",
  "files": [
    {
      "is_dir": false,
      "modified": "2025-10-27 12:34:56",
      "name": "main.go",
      "path": "/path/to/examples/builtin_tools/main.go",
      "size": 4567
    }
  ],
  "pattern": "*.go",
  "recursive": false
}

Example 2: Get current time
{
  "datetime": "2025-10-27T12:34:56+09:00",
  "format": "RFC3339",
  "timezone": "Asia/Tokyo",
  "unix": 1729995296,
  "unix_nano": 1729995296123456789
}

Example 3: Use with LLM provider
LLM requested tool calls:
  Tool: datetime_now
  Arguments: map[format:RFC3339 timezone:Asia/Tokyo]
  Result:
    {
      "datetime": "2025-10-27T12:34:56+09:00",
      "format": "RFC3339",
      "timezone": "Asia/Tokyo",
      "unix": 1729995296,
      "unix_nano": 1729995296123456789
    }

Final answer:
The current time in Tokyo is 12:34:56 PM on October 27, 2025 (JST).
```

## Security Notes

- File operations are restricted to allowed paths (current directory and `/tmp`)
- Maximum file size is 10MB
- Symbolic links are not followed
- All paths are validated for directory traversal attempts

## Next Steps

- Explore adding more tools (web fetch, system info)
- Build custom tools for your use case
- Integrate with agent workflows
- Add error handling and logging
