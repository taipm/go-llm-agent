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

// WriteConfig extends Config with write-specific options
type WriteConfig struct {
	Config

	// CreateDirs automatically creates parent directories if they don't exist
	CreateDirs bool

	// Backup creates a backup of existing files before overwriting
	Backup bool

	// BackupSuffix is the suffix added to backup files (default: ".bak")
	BackupSuffix string
}

// DefaultWriteConfig provides sensible defaults for write operations
var DefaultWriteConfig = WriteConfig{
	Config:       DefaultConfig,
	CreateDirs:   true,
	Backup:       false,
	BackupSuffix: ".bak",
}

// WriteTool writes content to a file
type WriteTool struct {
	tools.BaseTool
	config WriteConfig
}

// NewWriteTool creates a new file write tool with the given configuration
func NewWriteTool(config WriteConfig) *WriteTool {
	return &WriteTool{
		BaseTool: tools.NewBaseTool(
			"file_write",
			"Write or append content to a file on the filesystem. Can create parent directories and backup existing files.",
			tools.CategoryFile,
			false, // no auth required
			false, // NOT safe (can modify filesystem)
		),
		config: config,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *WriteTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"path": {
				Type:        "string",
				Description: "Absolute or relative path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
			},
			"append": {
				Type:        "boolean",
				Description: "If true, append to existing file. If false, overwrite (default: false)",
			},
		},
		Required: []string{"path", "content"},
	}
}

// Execute writes content to a file
func (t *WriteTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract and validate parameters
	pathStr, content, appendMode, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Validate path
	absPath, err := t.validatePath(pathStr)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Security checks
	if err := t.securityChecks(absPath, content); err != nil {
		return nil, err
	}

	// Get file info before write
	fileInfo, fileExists := t.checkFileExists(absPath)

	// Prepare filesystem (create dirs, backup)
	if err := t.prepareFilesystem(absPath, fileExists, appendMode); err != nil {
		return nil, err
	}

	// Perform write operation
	bytesWritten, err := t.writeFile(absPath, content, appendMode)
	if err != nil {
		return nil, err
	}

	// Build and return result
	return t.buildResult(absPath, bytesWritten, fileExists, fileInfo, appendMode)
}

// extractParams extracts and validates parameters
func (t *WriteTool) extractParams(params map[string]interface{}) (string, string, bool, error) {
	pathStr, ok := params["path"].(string)
	if !ok || pathStr == "" {
		return "", "", false, fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return "", "", false, fmt.Errorf("content parameter is required and must be a string")
	}

	appendMode := false
	if appendParam, ok := params["append"].(bool); ok {
		appendMode = appendParam
	}

	return pathStr, content, appendMode, nil
}

// securityChecks performs security validation
func (t *WriteTool) securityChecks(absPath, content string) error {
	// Check content size
	contentSize := int64(len(content))
	if t.config.MaxFileSize > 0 && contentSize > t.config.MaxFileSize {
		return fmt.Errorf("content size (%d bytes) exceeds maximum allowed size (%d bytes)",
			contentSize, t.config.MaxFileSize)
	}

	// Check if path is allowed
	if !t.isPathAllowed(absPath) {
		return fmt.Errorf("path %s is not in allowed directories", absPath)
	}

	return nil
}

// checkFileExists checks if file exists and returns its info
func (t *WriteTool) checkFileExists(absPath string) (os.FileInfo, bool) {
	fileInfo, err := os.Stat(absPath)
	return fileInfo, err == nil
}

// prepareFilesystem creates directories and backups as needed
func (t *WriteTool) prepareFilesystem(absPath string, fileExists, appendMode bool) error {
	// Create parent directories if needed
	if t.config.CreateDirs {
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	// Backup existing file if configured and file exists
	if fileExists && t.config.Backup && !appendMode {
		backupPath := absPath + t.config.BackupSuffix
		if err := t.backupFile(absPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	return nil
}

// writeFile performs the actual write operation
func (t *WriteTool) writeFile(absPath, content string, appendMode bool) (int, error) {
	var flags int
	if appendMode {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}

	file, err := os.OpenFile(absPath, flags, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	bytesWritten, err := file.WriteString(content)
	if err != nil {
		return 0, fmt.Errorf("failed to write content: %w", err)
	}

	return bytesWritten, nil
}

// buildResult creates the result map
func (t *WriteTool) buildResult(absPath string, bytesWritten int, fileExists bool, fileInfo os.FileInfo, appendMode bool) (map[string]interface{}, error) {
	// Get file info after write
	newFileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info after write: %w", err)
	}

	mode := "write"
	if appendMode {
		mode = "append"
	}

	result := map[string]interface{}{
		"success":       true,
		"path":          absPath,
		"bytes_written": bytesWritten,
		"total_size":    newFileInfo.Size(),
		"mode":          mode,
	}

	if fileExists {
		result["existed"] = true
		result["previous_size"] = fileInfo.Size()
	} else {
		result["existed"] = false
		result["created"] = true
	}

	if t.config.Backup && fileExists && !appendMode {
		result["backup_created"] = true
		result["backup_path"] = absPath + t.config.BackupSuffix
	}

	return result, nil
}

// validatePath validates and resolves the file path
func (t *WriteTool) validatePath(path string) (string, error) {
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
func (t *WriteTool) isPathAllowed(absPath string) bool {
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

// backupFile creates a backup copy of the file
func (t *WriteTool) backupFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}
