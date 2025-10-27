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

// DeleteConfig extends Config with delete-specific options
type DeleteConfig struct {
	Config

	// ProtectedPaths are paths that cannot be deleted (e.g., system directories)
	ProtectedPaths []string

	// AllowRecursive allows recursive deletion of directories
	AllowRecursive bool

	// RequireConfirmation would require explicit confirmation (for future use)
	RequireConfirmation bool
}

// DefaultDeleteConfig provides sensible defaults for delete operations
var DefaultDeleteConfig = DeleteConfig{
	Config: DefaultConfig,
	ProtectedPaths: []string{
		"/",
		"/bin",
		"/boot",
		"/dev",
		"/etc",
		"/lib",
		"/proc",
		"/root",
		"/sbin",
		"/sys",
		"/usr",
		"/var",
		"C:\\Windows",
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	},
	AllowRecursive:      false,
	RequireConfirmation: true,
}

// DeleteTool deletes files or directories
type DeleteTool struct {
	tools.BaseTool
	config DeleteConfig
}

// NewDeleteTool creates a new file delete tool with the given configuration
func NewDeleteTool(config DeleteConfig) *DeleteTool {
	return &DeleteTool{
		BaseTool: tools.NewBaseTool(
			"file_delete",
			"Delete a file or directory from the filesystem. WARNING: This operation is irreversible. Use with caution.",
			tools.CategoryFile,
			false, // no auth required
			false, // NOT safe (can delete files)
		),
		config: config,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *DeleteTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path": {
				Type:        "string",
				Description: "Absolute or relative path to the file or directory to delete",
			},
			"recursive": {
				Type:        "boolean",
				Description: "If true, recursively delete directories and their contents (default: false)",
			},
		},
		Required: []string{"path"},
	}
}

// Execute deletes a file or directory
func (t *DeleteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract and validate parameters
	pathStr, recursive, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Validate path
	absPath, err := t.validatePath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Security checks
	if err := t.securityChecks(absPath, recursive); err != nil {
		return nil, err
	}

	// Get file/directory info before deletion
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", absPath)
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	// Perform deletion
	deletedItems, err := t.performDelete(absPath, info.IsDir(), recursive)
	if err != nil {
		return nil, err
	}

	// Build result
	return t.buildResult(absPath, info, deletedItems), nil
}

// extractParams extracts and validates parameters
func (t *DeleteTool) extractParams(params map[string]interface{}) (string, bool, error) {
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return "", false, fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	recursive := false
	if recursiveParam, ok := params["recursive"].(bool); ok {
		recursive = recursiveParam
	}

	return pathStr, recursive, nil
}

// securityChecks performs security validation
func (t *DeleteTool) securityChecks(absPath string, recursive bool) error {
	// Check if path is protected
	if t.isProtectedPath(absPath) {
		return fmt.Errorf("cannot delete protected path: %s", absPath)
	}

	// Check if path is allowed
	if !t.isPathAllowed(absPath) {
		return fmt.Errorf("path %s is not in allowed directories", absPath)
	}

	// Check if recursive deletion is allowed
	info, err := os.Stat(absPath)
	if err == nil && info.IsDir() && recursive && !t.config.AllowRecursive {
		return fmt.Errorf("recursive deletion is not allowed by configuration")
	}

	return nil
}

// performDelete performs the actual deletion
func (t *DeleteTool) performDelete(absPath string, isDir, recursive bool) (int, error) {
	if !isDir {
		// Delete file
		if err := os.Remove(absPath); err != nil {
			return 0, fmt.Errorf("failed to delete file: %w", err)
		}
		return 1, nil
	}

	// Delete directory
	if recursive {
		// Count items before deletion
		count, err := t.countItems(absPath)
		if err != nil {
			return 0, fmt.Errorf("failed to count items: %w", err)
		}

		if err := os.RemoveAll(absPath); err != nil {
			return 0, fmt.Errorf("failed to delete directory recursively: %w", err)
		}
		return count, nil
	}

	// Non-recursive directory deletion (must be empty)
	if err := os.Remove(absPath); err != nil {
		if strings.Contains(err.Error(), "directory not empty") {
			return 0, fmt.Errorf("directory is not empty; use recursive=true to delete non-empty directories")
		}
		return 0, fmt.Errorf("failed to delete directory: %w", err)
	}
	return 1, nil
}

// countItems counts files and directories recursively
func (t *DeleteTool) countItems(root string) (int, error) {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

// buildResult creates the result map
func (t *DeleteTool) buildResult(absPath string, info os.FileInfo, deletedItems int) map[string]interface{} {
	result := map[string]interface{}{
		"success":       true,
		"path":          absPath,
		"deleted_items": deletedItems,
	}

	if info.IsDir() {
		result["type"] = "directory"
	} else {
		result["type"] = "file"
		result["size"] = info.Size()
	}

	return result
}

// validatePath validates and resolves the file path
func (t *DeleteTool) validatePath(path string) (string, error) {
	// Check for empty path
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path contains directory traversal (..), which is not allowed")
	}

	// Check for null bytes (security)
	if strings.Contains(path, "\x00") {
		return "", fmt.Errorf("path contains null bytes")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Clean the path
	absPath = filepath.Clean(absPath)

	return absPath, nil
}

// isPathAllowed checks if the path is within allowed directories
func (t *DeleteTool) isPathAllowed(absPath string) bool {
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

// isProtectedPath checks if the path is in the protected list
func (t *DeleteTool) isProtectedPath(absPath string) bool {
	absPath = filepath.Clean(absPath)

	for _, protected := range t.config.ProtectedPaths {
		protectedAbs, err := filepath.Abs(protected)
		if err != nil {
			// If we can't resolve, treat as exact match
			if absPath == protected {
				return true
			}
			continue
		}

		protectedAbs = filepath.Clean(protectedAbs)

		// Exact match
		if absPath == protectedAbs {
			return true
		}

		// Check if absPath is under protected path
		rel, err := filepath.Rel(protectedAbs, absPath)
		if err != nil {
			continue
		}

		// If path is under protected directory, block it
		if !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}
