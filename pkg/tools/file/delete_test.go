package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/taipm/go-llm-agent/pkg/tools"
)

func TestDeleteTool_Execute(t *testing.T) {
	// Create temp directory for tests
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		config    DeleteConfig
		params    map[string]interface{}
		setupFunc func() error
		wantErr   bool
		validate  func(t *testing.T, result interface{})
	}{
		{
			name: "delete file",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "delete_me.txt"),
			},
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "delete_me.txt"), []byte("test"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if !r["success"].(bool) {
					t.Error("expected success to be true")
				}
				if r["type"].(string) != "file" {
					t.Errorf("expected type file, got %s", r["type"].(string))
				}
				if r["deleted_items"].(int) != 1 {
					t.Errorf("expected 1 deleted item, got %d", r["deleted_items"].(int))
				}

				// Verify file is deleted
				if _, err := os.Stat(filepath.Join(tmpDir, "delete_me.txt")); !os.IsNotExist(err) {
					t.Error("file should be deleted")
				}
			},
		},
		{
			name: "delete empty directory",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "empty_dir"),
			},
			setupFunc: func() error {
				return os.Mkdir(filepath.Join(tmpDir, "empty_dir"), 0755)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if r["type"].(string) != "directory" {
					t.Errorf("expected type directory, got %s", r["type"].(string))
				}

				// Verify directory is deleted
				if _, err := os.Stat(filepath.Join(tmpDir, "empty_dir")); !os.IsNotExist(err) {
					t.Error("directory should be deleted")
				}
			},
		},
		{
			name: "delete directory recursively",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: true,
			},
			params: map[string]interface{}{
				"path":      filepath.Join(tmpDir, "recursive_dir"),
				"recursive": true,
			},
			setupFunc: func() error {
				dir := filepath.Join(tmpDir, "recursive_dir")
				if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("test"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "subdir", "file2.txt"), []byte("test"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				// Should delete: recursive_dir/ + subdir/ + file1.txt + file2.txt = 4 items
				if r["deleted_items"].(int) < 3 {
					t.Errorf("expected at least 3 deleted items, got %d", r["deleted_items"].(int))
				}

				// Verify directory is deleted
				if _, err := os.Stat(filepath.Join(tmpDir, "recursive_dir")); !os.IsNotExist(err) {
					t.Error("directory should be deleted recursively")
				}
			},
		},
		{
			name: "reject non-empty directory without recursive flag",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path":      filepath.Join(tmpDir, "non_empty"),
				"recursive": false,
			},
			setupFunc: func() error {
				dir := filepath.Join(tmpDir, "non_empty")
				if err := os.Mkdir(dir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644)
			},
			wantErr: true,
		},
		{
			name: "reject recursive deletion when not allowed",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false, // Recursive not allowed
			},
			params: map[string]interface{}{
				"path":      filepath.Join(tmpDir, "no_recursive"),
				"recursive": true,
			},
			setupFunc: func() error {
				dir := filepath.Join(tmpDir, "no_recursive")
				if err := os.Mkdir(dir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644)
			},
			wantErr: true,
		},
		{
			name: "reject protected path",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{filepath.Join(tmpDir, "protected")},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "protected", "file.txt"),
			},
			setupFunc: func() error {
				dir := filepath.Join(tmpDir, "protected")
				if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644)
			},
			wantErr: true,
		},
		{
			name: "reject path outside allowed directories",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": "/tmp/forbidden.txt",
			},
			wantErr: true,
		},
		{
			name: "reject directory traversal",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "..", "traversal.txt"),
			},
			wantErr: true,
		},
		{
			name: "reject empty path",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": "",
			},
			wantErr: true,
		},
		{
			name: "reject non-existent path",
			config: DeleteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
				},
				ProtectedPaths: []string{},
				AllowRecursive: false,
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "does_not_exist.txt"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			tool := NewDeleteTool(tt.config)
			result, err := tool.Execute(context.Background(), tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestDeleteTool_ProtectedPaths(t *testing.T) {
	tool := NewDeleteTool(DefaultDeleteConfig)

	// Test system paths are protected
	systemPaths := []string{"/", "/etc", "/usr", "/bin"}

	for _, path := range systemPaths {
		if !tool.isProtectedPath(path) {
			t.Errorf("path %s should be protected", path)
		}
	}
}

func TestDeleteTool_Parameters(t *testing.T) {
	tool := NewDeleteTool(DefaultDeleteConfig)
	params := tool.Parameters()

	if params == nil {
		t.Fatal("expected parameters to be non-nil")
	}

	if params.Type != "object" {
		t.Error("expected type to be object")
	}

	if params.Properties["path"] == nil {
		t.Error("expected path property")
	}
	if params.Properties["recursive"] == nil {
		t.Error("expected recursive property")
	}

	if len(params.Required) != 1 || params.Required[0] != "path" {
		t.Errorf("expected only path to be required, got %v", params.Required)
	}
}

func TestDeleteTool_Metadata(t *testing.T) {
	tool := NewDeleteTool(DefaultDeleteConfig)

	if tool.Name() != "file_delete" {
		t.Errorf("expected name file_delete, got %s", tool.Name())
	}

	if tool.Category() != tools.CategoryFile {
		t.Errorf("expected category File, got %s", tool.Category())
	}

	if tool.IsSafe() {
		t.Error("expected file_delete to be unsafe")
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	// Description should contain warning
	if !contains(tool.Description(), "WARNING") {
		t.Error("expected description to contain WARNING")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
