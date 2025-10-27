package builtin

import (
	"testing"

	"github.com/taipm/go-llm-agent/pkg/tools"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// File config checks
	if config.File.Base.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected MaxFileSize 10MB, got %d", config.File.Base.MaxFileSize)
	}
	if config.File.Write.Backup != true {
		t.Error("Expected Write.Backup to be true")
	}
	if config.File.Delete.RequireConfirmation != true {
		t.Error("Expected Delete.RequireConfirmation to be true")
	}

	// Web config checks
	if config.Web.Fetch.UserAgent != defaultUserAgent {
		t.Errorf("Expected UserAgent %s, got %s", defaultUserAgent, config.Web.Fetch.UserAgent)
	}
	if config.Web.Post.MaxResponseSize != 1024*1024 {
		t.Errorf("Expected MaxResponseSize 1MB, got %d", config.Web.Post.MaxResponseSize)
	}
	if config.Web.Scrape.AllowPrivateIPs != false {
		t.Error("Expected Scrape.AllowPrivateIPs to be false")
	}

	// Category flags
	if config.NoFile || config.NoWeb || config.NoTime {
		t.Error("Expected all category flags to be false by default")
	}
}

func TestGetRegistry(t *testing.T) {
	registry := GetRegistry()

	if registry == nil {
		t.Fatal("GetRegistry returned nil")
	}

	// Should have all 11 tools (was 10, now +1 system)
	if registry.Count() != 11 {
		t.Errorf("Expected 11 tools, got %d", registry.Count())
	}

	// Check specific tools exist
	expectedTools := []string{
		"file_read",
		"file_list",
		"file_write",
		"file_delete",
		"web_fetch",
		"web_post",
		"web_scrape",
		"datetime_now",
		"datetime_format",
		"datetime_calc",
		"system_info",
	}

	for _, name := range expectedTools {
		if !registry.Has(name) {
			t.Errorf("Expected tool %s to be registered", name)
		}
	}
}

func TestGetRegistryWithConfig_NoFile(t *testing.T) {
	config := DefaultConfig()
	config.NoFile = true

	registry := GetRegistryWithConfig(config)

	// Should have 7 tools (3 web + 3 datetime + 1 system)
	if registry.Count() != 7 {
		t.Errorf("Expected 7 tools, got %d", registry.Count())
	}

	// File tools should not exist
	if registry.Has("file_read") {
		t.Error("Expected file_read to not be registered")
	}

	// Web and datetime tools should exist
	if !registry.Has("web_fetch") {
		t.Error("Expected web_fetch to be registered")
	}
	if !registry.Has("datetime_now") {
		t.Error("Expected datetime_now to be registered")
	}
}

func TestGetRegistryWithConfig_NoWeb(t *testing.T) {
	config := DefaultConfig()
	config.NoWeb = true

	registry := GetRegistryWithConfig(config)

	// Should have 8 tools (4 file + 3 datetime + 1 system)
	if registry.Count() != 8 {
		t.Errorf("Expected 8 tools, got %d", registry.Count())
	}

	// Web tools should not exist
	if registry.Has("web_fetch") {
		t.Error("Expected web_fetch to not be registered")
	}

	// File and datetime tools should exist
	if !registry.Has("file_read") {
		t.Error("Expected file_read to be registered")
	}
	if !registry.Has("datetime_now") {
		t.Error("Expected datetime_now to be registered")
	}
}

func TestGetRegistryWithConfig_NoTime(t *testing.T) {
	config := DefaultConfig()
	config.NoTime = true

	registry := GetRegistryWithConfig(config)

	// Should have 8 tools (4 file + 3 web + 1 system)
	if registry.Count() != 8 {
		t.Errorf("Expected 8 tools, got %d", registry.Count())
	}

	// DateTime tools should not exist
	if registry.Has("datetime_now") {
		t.Error("Expected datetime_now to not be registered")
	}

	// File and web tools should exist
	if !registry.Has("file_read") {
		t.Error("Expected file_read to be registered")
	}
	if !registry.Has("web_fetch") {
		t.Error("Expected web_fetch to be registered")
	}
}

func TestGetRegistryWithConfig_OnlyFile(t *testing.T) {
	config := DefaultConfig()
	config.NoWeb = true
	config.NoTime = true
	config.NoSystem = true

	registry := GetRegistryWithConfig(config)

	// Should have only 4 file tools
	if registry.Count() != 4 {
		t.Errorf("Expected 4 tools, got %d", registry.Count())
	}

	fileTools := []string{"file_read", "file_list", "file_write", "file_delete"}
	for _, name := range fileTools {
		if !registry.Has(name) {
			t.Errorf("Expected %s to be registered", name)
		}
	}
}

func TestGetRegistryWithConfig_NoSystem(t *testing.T) {
	config := DefaultConfig()
	config.NoSystem = true

	registry := GetRegistryWithConfig(config)

	// Should have 10 tools (4 file + 3 web + 3 datetime)
	if registry.Count() != 10 {
		t.Errorf("Expected 10 tools, got %d", registry.Count())
	}

	// System tools should not exist
	if registry.Has("system_info") {
		t.Error("Expected system_info to not be registered")
	}

	// Other tools should exist
	if !registry.Has("file_read") {
		t.Error("Expected file_read to be registered")
	}
	if !registry.Has("web_fetch") {
		t.Error("Expected web_fetch to be registered")
	}
	if !registry.Has("datetime_now") {
		t.Error("Expected datetime_now to be registered")
	}
}

func TestGetAllTools(t *testing.T) {
	tools := GetAllTools()

	if len(tools) != 11 {
		t.Errorf("Expected 11 tools, got %d", len(tools))
	}

	// Check each tool has required methods
	for _, tool := range tools {
		if tool.Name() == "" {
			t.Error("Tool has empty name")
		}
		if tool.Description() == "" {
			t.Error("Tool has empty description")
		}
		if tool.Category() == "" {
			t.Errorf("Tool %s has empty category", tool.Name())
		}
	}
}

func TestGetToolsByCategory(t *testing.T) {
	fileTools := GetToolsByCategory(tools.CategoryFile)
	webTools := GetToolsByCategory(tools.CategoryWeb)
	datetimeTools := GetToolsByCategory(tools.CategoryDateTime)
	systemTools := GetToolsByCategory(tools.CategorySystem)

	if len(fileTools) != 4 {
		t.Errorf("Expected 4 file tools, got %d", len(fileTools))
	}
	if len(webTools) != 3 {
		t.Errorf("Expected 3 web tools, got %d", len(webTools))
	}
	if len(datetimeTools) != 3 {
		t.Errorf("Expected 3 datetime tools, got %d", len(datetimeTools))
	}
	if len(systemTools) != 1 {
		t.Errorf("Expected 1 system tool, got %d", len(systemTools))
	}

	// Verify categories
	for _, tool := range fileTools {
		if tool.Category() != tools.CategoryFile {
			t.Errorf("Tool %s in file category has wrong category: %s", tool.Name(), tool.Category())
		}
	}
	for _, tool := range webTools {
		if tool.Category() != tools.CategoryWeb {
			t.Errorf("Tool %s in web category has wrong category: %s", tool.Name(), tool.Category())
		}
	}
	for _, tool := range datetimeTools {
		if tool.Category() != tools.CategoryDateTime {
			t.Errorf("Tool %s in datetime category has wrong category: %s", tool.Name(), tool.Category())
		}
	}
}

func TestGetFileTools(t *testing.T) {
	fileTools := GetFileTools(nil)

	if len(fileTools) != 4 {
		t.Errorf("Expected 4 file tools, got %d", len(fileTools))
	}

	expectedNames := map[string]bool{
		"file_read":   false,
		"file_list":   false,
		"file_write":  false,
		"file_delete": false,
	}

	for _, tool := range fileTools {
		if _, ok := expectedNames[tool.Name()]; !ok {
			t.Errorf("Unexpected tool %s", tool.Name())
		}
		expectedNames[tool.Name()] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestGetWebTools(t *testing.T) {
	webTools := GetWebTools(nil)

	if len(webTools) != 3 {
		t.Errorf("Expected 3 web tools, got %d", len(webTools))
	}

	expectedNames := map[string]bool{
		"web_fetch":  false,
		"web_post":   false,
		"web_scrape": false,
	}

	for _, tool := range webTools {
		if _, ok := expectedNames[tool.Name()]; !ok {
			t.Errorf("Unexpected tool %s", tool.Name())
		}
		expectedNames[tool.Name()] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestGetDateTimeTools(t *testing.T) {
	datetimeTools := GetDateTimeTools()

	if len(datetimeTools) != 3 {
		t.Errorf("Expected 3 datetime tools, got %d", len(datetimeTools))
	}

	expectedNames := map[string]bool{
		"datetime_now":    false,
		"datetime_format": false,
		"datetime_calc":   false,
	}

	for _, tool := range datetimeTools {
		if _, ok := expectedNames[tool.Name()]; !ok {
			t.Errorf("Unexpected tool %s", tool.Name())
		}
		expectedNames[tool.Name()] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestGetSystemTools(t *testing.T) {
	systemTools := GetSystemTools()

	if len(systemTools) != 1 {
		t.Errorf("Expected 1 system tool, got %d", len(systemTools))
	}

	expectedNames := map[string]bool{
		"system_info": false,
	}

	for _, tool := range systemTools {
		if _, ok := expectedNames[tool.Name()]; !ok {
			t.Errorf("Unexpected tool %s", tool.Name())
		}
		expectedNames[tool.Name()] = true
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected tool %s not found", name)
		}
	}
}

func TestToolCount(t *testing.T) {
	count := ToolCount()
	if count != 11 {
		t.Errorf("Expected ToolCount to return 11, got %d", count)
	}
}

func TestCustomConfig(t *testing.T) {
	config := DefaultConfig()
	config.File.Base.MaxFileSize = 5 * 1024 * 1024 // 5MB
	config.Web.Fetch.UserAgent = "CustomAgent/2.0"

	registry := GetRegistryWithConfig(config)

	// Registry should still have all 11 tools
	if registry.Count() != 11 {
		t.Errorf("Expected 11 tools, got %d", registry.Count())
	}

	// Note: We can't directly verify the config was applied since tools don't expose it
	// But we can verify all tools are registered
	if !registry.Has("file_read") || !registry.Has("web_fetch") {
		t.Error("Expected all tools to be registered with custom config")
	}
}

func TestSafeTools(t *testing.T) {
	registry := GetRegistry()
	safeTools := registry.SafeTools()

	// All tools should be safe except file_write and file_delete
	expectedSafeCount := 9 // 11 total - 2 unsafe (write, delete)

	if len(safeTools) != expectedSafeCount {
		t.Errorf("Expected %d safe tools, got %d", expectedSafeCount, len(safeTools))
	}

	// Verify unsafe tools are not in safe list
	for _, tool := range safeTools {
		if tool.Name() == "file_write" || tool.Name() == "file_delete" {
			t.Errorf("Tool %s should not be marked as safe", tool.Name())
		}
	}
}

func TestToToolDefinitions(t *testing.T) {
	registry := GetRegistry()
	defs := registry.ToToolDefinitions()

	if len(defs) != 11 {
		t.Errorf("Expected 11 tool definitions, got %d", len(defs))
	}

	for _, def := range defs {
		if def.Type != "function" {
			t.Errorf("Expected type 'function', got %s", def.Type)
		}
		if def.Function.Name == "" {
			t.Error("Tool definition has empty name")
		}
		if def.Function.Description == "" {
			t.Error("Tool definition has empty description")
		}
	}
}
