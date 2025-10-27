package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/taipm/go-llm-agent/pkg/tools"
)

func TestWriteTool_Execute(t *testing.T) {
	// Create temp directory for tests
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		config    WriteConfig
		params    map[string]interface{}
		setupFunc func() error
		wantErr   bool
		validate  func(t *testing.T, result interface{})
	}{
		{
			name: "write new file",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
				CreateDirs: true,
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "test.txt"),
				"content": "Hello, World!",
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if !r["success"].(bool) {
					t.Error("expected success to be true")
				}
				if r["created"].(bool) != true {
					t.Error("expected created to be true")
				}
				if r["bytes_written"].(int) != 13 {
					t.Errorf("expected 13 bytes written, got %d", r["bytes_written"].(int))
				}

				// Verify file content
				content, err := os.ReadFile(filepath.Join(tmpDir, "test.txt"))
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "Hello, World!" {
					t.Errorf("unexpected content: %s", string(content))
				}
			},
		},
		{
			name: "append to existing file",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
				CreateDirs: true,
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "append.txt"),
				"content": " World!",
				"append":  true,
			},
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "append.txt"), []byte("Hello,"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if r["mode"].(string) != "append" {
					t.Errorf("expected mode append, got %s", r["mode"].(string))
				}

				// Verify file content
				content, err := os.ReadFile(filepath.Join(tmpDir, "append.txt"))
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "Hello, World!" {
					t.Errorf("unexpected content: %s", string(content))
				}
			},
		},
		{
			name: "overwrite existing file",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
				CreateDirs: true,
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "overwrite.txt"),
				"content": "New Content",
				"append":  false,
			},
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "overwrite.txt"), []byte("Old Content"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if r["existed"].(bool) != true {
					t.Error("expected existed to be true")
				}
				if r["previous_size"].(int64) != 11 {
					t.Errorf("expected previous_size 11, got %d", r["previous_size"].(int64))
				}

				// Verify file content
				content, err := os.ReadFile(filepath.Join(tmpDir, "overwrite.txt"))
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "New Content" {
					t.Errorf("unexpected content: %s", string(content))
				}
			},
		},
		{
			name: "create parent directories",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
				CreateDirs: true,
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "sub", "dir", "file.txt"),
				"content": "test",
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				// Verify directories were created
				info, err := os.Stat(filepath.Join(tmpDir, "sub", "dir"))
				if err != nil {
					t.Fatalf("directories not created: %v", err)
				}
				if !info.IsDir() {
					t.Error("expected directory")
				}
			},
		},
		{
			name: "backup existing file",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
				CreateDirs:   true,
				Backup:       true,
				BackupSuffix: ".bak",
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "backup.txt"),
				"content": "New Content",
			},
			setupFunc: func() error {
				return os.WriteFile(filepath.Join(tmpDir, "backup.txt"), []byte("Old Content"), 0644)
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				r := result.(map[string]interface{})
				if r["backup_created"].(bool) != true {
					t.Error("expected backup_created to be true")
				}

				// Verify backup file exists
				backupContent, err := os.ReadFile(filepath.Join(tmpDir, "backup.txt.bak"))
				if err != nil {
					t.Fatalf("backup file not found: %v", err)
				}
				if string(backupContent) != "Old Content" {
					t.Errorf("unexpected backup content: %s", string(backupContent))
				}

				// Verify new content
				content, err := os.ReadFile(filepath.Join(tmpDir, "backup.txt"))
				if err != nil {
					t.Fatalf("failed to read file: %v", err)
				}
				if string(content) != "New Content" {
					t.Errorf("unexpected content: %s", string(content))
				}
			},
		},
		{
			name: "reject path outside allowed directories",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
			},
			params: map[string]interface{}{
				"path":    "/tmp/forbidden.txt",
				"content": "test",
			},
			wantErr: true,
		},
		{
			name: "reject content exceeding max size",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  10,
				},
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "large.txt"),
				"content": "This content is too large for the limit",
			},
			wantErr: true,
		},
		{
			name: "reject directory traversal",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
			},
			params: map[string]interface{}{
				"path":    filepath.Join(tmpDir, "..", "traversal.txt"),
				"content": "test",
			},
			wantErr: true,
		},
		{
			name: "reject empty path",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
			},
			params: map[string]interface{}{
				"path":    "",
				"content": "test",
			},
			wantErr: true,
		},
		{
			name: "reject missing content",
			config: WriteConfig{
				Config: Config{
					AllowedPaths: []string{tmpDir},
					MaxFileSize:  1024,
				},
			},
			params: map[string]interface{}{
				"path": filepath.Join(tmpDir, "test.txt"),
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

			tool := NewWriteTool(tt.config)
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

func TestWriteTool_Parameters(t *testing.T) {
	tool := NewWriteTool(DefaultWriteConfig)
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
	if params.Properties["content"] == nil {
		t.Error("expected content property")
	}
	if params.Properties["append"] == nil {
		t.Error("expected append property")
	}

	if len(params.Required) != 2 {
		t.Errorf("expected 2 required parameters, got %d", len(params.Required))
	}
}

func TestWriteTool_Metadata(t *testing.T) {
	tool := NewWriteTool(DefaultWriteConfig)

	if tool.Name() != "file_write" {
		t.Errorf("expected name file_write, got %s", tool.Name())
	}

	if tool.Category() != tools.CategoryFile {
		t.Errorf("expected category File, got %s", tool.Category())
	}

	if tool.IsSafe() {
		t.Error("expected file_write to be unsafe")
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}
}
