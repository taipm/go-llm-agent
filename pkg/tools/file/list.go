package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ListTool lists files and directories in a path
type ListTool struct {
	tools.BaseTool
	config Config
}

// NewListTool creates a new file list tool
func NewListTool(config Config) *ListTool {
	return &ListTool{
		BaseTool: tools.NewBaseTool(
			"file_list",
			"List all files and directories in a specified path",
			tools.CategoryFile,
			false, // no auth required
			true,  // safe operation
		),
		config: config,
	}
}

// Parameters implements Tool.Parameters
func (t *ListTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path": {
				Type:        "string",
				Description: "Path to directory to list",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List files recursively (default: false)",
			},
			"pattern": {
				Type:        "string",
				Description: "Glob pattern to filter files (e.g., '*.txt', '*.go')",
			},
		},
		Required: []string{"path"},
	}
}

// Execute implements Tool.Execute
func (t *ListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return nil, fmt.Errorf("path is required")
	}

	absPath, err := filepath.Abs(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	recursive := false
	if r, ok := params["recursive"].(bool); ok {
		recursive = r
	}

	pattern := "*"
	if p, ok := params["pattern"].(string); ok && p != "" {
		pattern = p
	}

	var files []map[string]interface{}

	if recursive {
		err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Match pattern
			matched, _ := filepath.Match(pattern, filepath.Base(path))
			if !matched && pattern != "*" {
				return nil
			}

			relPath, _ := filepath.Rel(absPath, path)
			files = append(files, map[string]interface{}{
				"name":     filepath.Base(path),
				"path":     path,
				"rel_path": relPath,
				"is_dir":   info.IsDir(),
				"size":     info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			})
			return nil
		})
	} else {
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			matched, _ := filepath.Match(pattern, entry.Name())
			if !matched && pattern != "*" {
				continue
			}

			info, _ := entry.Info()
			fullPath := filepath.Join(absPath, entry.Name())

			files = append(files, map[string]interface{}{
				"name":     entry.Name(),
				"path":     fullPath,
				"is_dir":   entry.IsDir(),
				"size":     info.Size(),
				"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	return map[string]interface{}{
		"directory": absPath,
		"files":     files,
		"count":     len(files),
		"recursive": recursive,
		"pattern":   pattern,
	}, err
}
