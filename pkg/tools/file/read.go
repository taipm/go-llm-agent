package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// Config contains configuration for file operations
type Config struct {
	// AllowedPaths restricts file operations to these directories
	// Empty means all paths allowed (not recommended for production)
	AllowedPaths []string

	// MaxFileSize limits the maximum file size for read/write operations (in bytes)
	MaxFileSize int64

	// AllowSymlinks allows following symbolic links
	AllowSymlinks bool
}

// DefaultConfig provides sensible defaults for file operations
var DefaultConfig = Config{
	AllowedPaths:  []string{},
	MaxFileSize:   10 * 1024 * 1024, // 10MB
	AllowSymlinks: false,
}

// ReadTool reads content from a file
type ReadTool struct {
	tools.BaseTool
	config Config
}

// NewReadTool creates a new file read tool with the given configuration
func NewReadTool(config Config) *ReadTool {
	return &ReadTool{
		BaseTool: tools.NewBaseTool(
			"file_read",
			"Read the complete content of a text file from the filesystem",
			tools.CategoryFile,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
	}
}

// Parameters implements Tool.Parameters
func (t *ReadTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path": {
				Type:        "string",
				Description: "Absolute or relative path to the file to read",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8). Supported: utf-8, ascii",
				Enum:        []interface{}{"utf-8", "ascii"},
			},
		},
		Required: []string{"path"},
	}
}

// Execute implements Tool.Execute
func (t *ReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Get path parameter
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return nil, fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate path
	if err := t.validatePath(pathStr); err != nil {
		return nil, err
	}

	// Get absolute path
	absPath, err := filepath.Abs(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Check if allowed
	if !t.isPathAllowed(absPath) {
		return nil, fmt.Errorf("access denied: path %s is not in allowed paths", absPath)
	}

	// Check file exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", absPath)
		}
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return nil, fmt.Errorf("path %s is a directory, not a file", absPath)
	}

	// Check file size
	if info.Size() > t.config.MaxFileSize {
		return nil, fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes",
			info.Size(), t.config.MaxFileSize)
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return map[string]interface{}{
		"content":  string(content),
		"path":     absPath,
		"size":     info.Size(),
		"modified": info.ModTime().Format("2006-01-02 15:04:05"),
	}, nil
}

// validatePath performs basic path validation
func (t *ReadTool) validatePath(path string) error {
	// Check for empty path
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path contains directory traversal (..), which is not allowed")
	}

	// Check for null bytes (security)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	return nil
}

// isPathAllowed checks if a path is within the allowed directories
func (t *ReadTool) isPathAllowed(absPath string) bool {
	// If no restrictions, allow all paths
	if len(t.config.AllowedPaths) == 0 {
		return true
	}

	// Check if path is under any allowed directory
	for _, allowedPath := range t.config.AllowedPaths {
		absAllowed, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}

		// Check if absPath starts with absAllowed
		rel, err := filepath.Rel(absAllowed, absPath)
		if err != nil {
			continue
		}

		// If relative path doesn't start with "..", it's under the allowed path
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}
