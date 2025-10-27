package builtin

import (
	"os"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/tools/datetime"
	"github.com/taipm/go-llm-agent/pkg/tools/file"
	"github.com/taipm/go-llm-agent/pkg/tools/system"
	"github.com/taipm/go-llm-agent/pkg/tools/web"
)

const defaultUserAgent = "GoLLMAgent/1.0"

// Config contains configuration for built-in tools.
// Use DefaultConfig() for sensible defaults or customize as needed.
type Config struct {
	File    FileConfig
	Web     WebConfig
	NoFile  bool // Skip registering file tools
	NoWeb   bool // Skip registering web tools
	NoTime  bool // Skip registering datetime tools
	NoSystem bool // Skip registering system tools
}

// FileConfig contains file tool configurations
type FileConfig struct {
	Base   file.Config
	Write  file.WriteConfig
	Delete file.DeleteConfig
}

// WebConfig contains web tool configurations
type WebConfig struct {
	Fetch  web.Config
	Post   web.PostConfig
	Scrape web.ScrapeConfig
}

// DefaultConfig returns sensible default configuration for all built-in tools.
// File tools: Allow current directory and temp, 10MB limit, no symlinks
// Web tools: 30s timeout, 1MB response limit, no private IPs
func DefaultConfig() Config {
	fileBaseConfig := file.Config{
		AllowedPaths:  []string{".", "/tmp", os.TempDir()},
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		AllowSymlinks: false,
	}

	return Config{
		File: FileConfig{
			Base: fileBaseConfig,
			Write: file.WriteConfig{
				Config:       fileBaseConfig,
				CreateDirs:   true,
				Backup:       true,
				BackupSuffix: ".bak",
			},
			Delete: file.DeleteConfig{
				Config:              fileBaseConfig,
				ProtectedPaths:      file.DefaultDeleteConfig.ProtectedPaths,
				AllowRecursive:      true,
				RequireConfirmation: true,
			},
		},
		Web: WebConfig{
			Fetch: web.Config{
				Timeout:         30 * time.Second,
				MaxResponseSize: 1024 * 1024, // 1MB
				UserAgent:       defaultUserAgent,
				AllowPrivateIPs: false,
			},
			Post: web.PostConfig{
				Timeout:         30 * time.Second,
				MaxResponseSize: 1024 * 1024,
				UserAgent:       defaultUserAgent,
				AllowPrivateIPs: false,
			},
			Scrape: web.ScrapeConfig{
				Timeout:         30 * time.Second,
				MaxResponseSize: 5 * 1024 * 1024, // 5MB for HTML
				UserAgent:       defaultUserAgent,
				AllowPrivateIPs: false,
				RateLimit:       1 * time.Second,
			},
		},
		NoFile:   false,
		NoWeb:    false,
		NoTime:   false,
		NoSystem: false,
	}
}

// GetRegistry returns a new Registry pre-populated with all built-in tools
// using default configurations.
//
// This is the simplest way to get started:
//
//	registry := builtin.GetRegistry()
//	result, err := registry.Execute(ctx, "file_read", params)
func GetRegistry() *tools.Registry {
	return GetRegistryWithConfig(DefaultConfig())
}

// GetRegistryWithConfig returns a new Registry pre-populated with built-in tools
// using custom configurations.
//
// Example:
//
//	config := builtin.DefaultConfig()
//	config.File.Base.AllowedPaths = []string{"/custom/path"}
//	config.NoWeb = true // Skip web tools
//	registry := builtin.GetRegistryWithConfig(config)
func GetRegistryWithConfig(config Config) *tools.Registry {
	registry := tools.NewRegistry()

	// Register File tools
	if !config.NoFile {
		registry.Register(file.NewReadTool(config.File.Base))
		registry.Register(file.NewListTool(config.File.Base))
		registry.Register(file.NewWriteTool(config.File.Write))
		registry.Register(file.NewDeleteTool(config.File.Delete))
	}

	// Register Web tools
	if !config.NoWeb {
		registry.Register(web.NewFetchTool(config.Web.Fetch))
		registry.Register(web.NewPostTool(config.Web.Post))
		registry.Register(web.NewScrapeTool(config.Web.Scrape))
	}

	// Register DateTime tools
	if !config.NoTime {
		registry.Register(datetime.NewNowTool())
		registry.Register(datetime.NewFormatTool())
		registry.Register(datetime.NewCalcTool())
	}

	// Register System tools
	if !config.NoSystem {
		registry.Register(system.NewInfoTool())
	}

	return registry
}

// GetAllTools returns all built-in tools as a slice using default configurations.
// This is useful if you want to inspect all available tools.
func GetAllTools() []tools.Tool {
	registry := GetRegistry()
	return registry.All()
}

// GetToolsByCategory returns all built-in tools in the specified category
// using default configurations.
// Valid categories: tools.CategoryFile, tools.CategoryWeb, tools.CategoryDateTime
//
// Example:
//
//	fileTools := builtin.GetToolsByCategory(tools.CategoryFile)
//	webTools := builtin.GetToolsByCategory(tools.CategoryWeb)
func GetToolsByCategory(category tools.ToolCategory) []tools.Tool {
	registry := GetRegistry()
	return registry.ByCategory(category)
}

// GetFileTools returns all file-related built-in tools with custom config.
// If config is nil, uses DefaultConfig().
func GetFileTools(config *FileConfig) []tools.Tool {
	if config == nil {
		cfg := DefaultConfig()
		config = &cfg.File
	}

	return []tools.Tool{
		file.NewReadTool(config.Base),
		file.NewListTool(config.Base),
		file.NewWriteTool(config.Write),
		file.NewDeleteTool(config.Delete),
	}
}

// GetWebTools returns all web-related built-in tools with custom config.
// If config is nil, uses DefaultConfig().
func GetWebTools(config *WebConfig) []tools.Tool {
	if config == nil {
		cfg := DefaultConfig()
		config = &cfg.Web
	}

	return []tools.Tool{
		web.NewFetchTool(config.Fetch),
		web.NewPostTool(config.Post),
		web.NewScrapeTool(config.Scrape),
	}
}

// GetDateTimeTools returns all datetime-related built-in tools.
func GetDateTimeTools() []tools.Tool {
	return []tools.Tool{
		datetime.NewNowTool(),
		datetime.NewFormatTool(),
		datetime.NewCalcTool(),
	}
}

// GetSystemTools returns all system-related built-in tools.
func GetSystemTools() []tools.Tool {
	return []tools.Tool{
		system.NewInfoTool(),
	}
}

// ToolCount returns the total number of built-in tools available.
func ToolCount() int {
	return 11 // 4 file + 3 web + 3 datetime + 1 system
}
