package system

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// AppsTool lists installed applications on the system
type AppsTool struct {
	tools.BaseTool
}

// NewAppsTool creates a new tool for listing installed applications
func NewAppsTool() *AppsTool {
	return &AppsTool{
		BaseTool: tools.NewBaseTool(
			"system_apps",
			"List installed applications on the system. Returns application names and installation paths. Platform-specific: macOS (.app bundles), Linux (package managers), Windows (Program Files).",
			tools.CategorySystem,
			false, // no auth required
			true,  // safe operation (read-only)
		),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *AppsTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"limit": {
				Type:        "integer",
				Description: "Maximum number of applications to return. Default is 100. Set to 0 for all.",
			},
			"name_filter": {
				Type:        "string",
				Description: "Filter applications by name (case-insensitive substring match). Optional.",
			},
			"source": {
				Type:        "string",
				Description: "Source to query: auto (default), applications (app directories), homebrew (macOS), apt (Linux), all",
				Enum:        []interface{}{"auto", "applications", "homebrew", "apt", "all"},
			},
		},
		Required: []string{},
	}
}

// Execute retrieves the list of installed applications
func (t *AppsTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	config, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	var apps []map[string]interface{}
	var sources []string

	// Determine which sources to query based on platform and config
	if config.Source == "all" {
		apps, sources = t.getAllApps(ctx, config)
	} else if config.Source == "auto" {
		apps, sources = t.getAutoApps(ctx, config)
	} else {
		apps, sources = t.getSourceApps(ctx, config.Source, config)
	}

	// Apply name filter
	if config.NameFilter != "" {
		apps = t.filterApps(apps, config.NameFilter)
	}

	// Sort by name
	sort.Slice(apps, func(i, j int) bool {
		name1, _ := apps[i]["name"].(string)
		name2, _ := apps[j]["name"].(string)
		return name1 < name2
	})

	// Apply limit
	if config.Limit > 0 && len(apps) > config.Limit {
		apps = apps[:config.Limit]
	}

	return map[string]interface{}{
		"type":         "applications",
		"count":        len(apps),
		"platform":     runtime.GOOS,
		"sources":      sources,
		"applications": apps,
	}, nil
}

// appsConfig holds the configuration for app listing
type appsConfig struct {
	Limit      int
	NameFilter string
	Source     string
}

// extractParams extracts and validates parameters
func (t *AppsTool) extractParams(params map[string]interface{}) (*appsConfig, error) {
	config := &appsConfig{
		Limit:  100, // default
		Source: "auto",
	}

	if limit, ok := params["limit"]; ok {
		if limitFloat, ok := limit.(float64); ok {
			config.Limit = int(limitFloat)
		} else if limitInt, ok := limit.(int); ok {
			config.Limit = limitInt
		}
	}

	if nameFilter, ok := params["name_filter"].(string); ok {
		config.NameFilter = nameFilter
	}

	if source, ok := params["source"].(string); ok {
		config.Source = source
	}

	return config, nil
}

// getAutoApps gets apps from the most appropriate source for the platform
func (t *AppsTool) getAutoApps(ctx context.Context, config *appsConfig) ([]map[string]interface{}, []string) {
	switch runtime.GOOS {
	case "darwin":
		return t.getMacOSApps(ctx)
	case "linux":
		return t.getLinuxApps(ctx)
	case "windows":
		return t.getWindowsApps(ctx)
	default:
		return []map[string]interface{}{}, []string{}
	}
}

// getAllApps gets apps from all available sources
func (t *AppsTool) getAllApps(ctx context.Context, config *appsConfig) ([]map[string]interface{}, []string) {
	var allApps []map[string]interface{}
	var allSources []string
	seen := make(map[string]bool)

	switch runtime.GOOS {
	case "darwin":
		apps, sources := t.getMacOSApps(ctx)
		for _, app := range apps {
			name, _ := app["name"].(string)
			if !seen[name] {
				allApps = append(allApps, app)
				seen[name] = true
			}
		}
		allSources = append(allSources, sources...)

		// Also try Homebrew
		apps, sources = t.getHomebrewApps(ctx)
		for _, app := range apps {
			name, _ := app["name"].(string)
			if !seen[name] {
				allApps = append(allApps, app)
				seen[name] = true
			}
		}
		allSources = append(allSources, sources...)

	case "linux":
		apps, sources := t.getLinuxApps(ctx)
		allApps = apps
		allSources = sources
	}

	return allApps, allSources
}

// getSourceApps gets apps from a specific source
func (t *AppsTool) getSourceApps(ctx context.Context, source string, config *appsConfig) ([]map[string]interface{}, []string) {
	switch source {
	case "applications":
		if runtime.GOOS == "darwin" {
			return t.getMacOSApps(ctx)
		} else if runtime.GOOS == "linux" {
			return t.getLinuxApps(ctx)
		}
	case "homebrew":
		return t.getHomebrewApps(ctx)
	case "apt":
		return t.getAPTApps(ctx)
	}
	return []map[string]interface{}{}, []string{}
}

// getMacOSApps lists applications from /Applications and ~/Applications
func (t *AppsTool) getMacOSApps(ctx context.Context) ([]map[string]interface{}, []string) {
	var apps []map[string]interface{}
	sources := []string{"/Applications"}

	// Check /Applications
	apps = append(apps, t.scanDirectory("/Applications", ".app")...)

	// Check ~/Applications
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userApps := filepath.Join(homeDir, "Applications")
		if _, err := os.Stat(userApps); err == nil {
			apps = append(apps, t.scanDirectory(userApps, ".app")...)
			sources = append(sources, userApps)
		}
	}

	return apps, sources
}

// getLinuxApps lists applications from common directories
func (t *AppsTool) getLinuxApps(ctx context.Context) ([]map[string]interface{}, []string) {
	var apps []map[string]interface{}
	var sources []string

	// Try apt first (Debian/Ubuntu)
	aptApps, aptSources := t.getAPTApps(ctx)
	if len(aptApps) > 0 {
		return aptApps, aptSources
	}

	// Fallback to desktop files
	desktopDirs := []string{
		"/usr/share/applications",
		"/usr/local/share/applications",
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		desktopDirs = append(desktopDirs, filepath.Join(homeDir, ".local/share/applications"))
	}

	for _, dir := range desktopDirs {
		if _, err := os.Stat(dir); err == nil {
			apps = append(apps, t.scanDirectory(dir, ".desktop")...)
			sources = append(sources, dir)
		}
	}

	return apps, sources
}

// getWindowsApps lists applications from Program Files
func (t *AppsTool) getWindowsApps(ctx context.Context) ([]map[string]interface{}, []string) {
	var apps []map[string]interface{}
	sources := []string{
		"C:\\Program Files",
		"C:\\Program Files (x86)",
	}

	for _, dir := range sources {
		if _, err := os.Stat(dir); err == nil {
			apps = append(apps, t.scanDirectory(dir, ".exe")...)
		}
	}

	return apps, sources
}

// getHomebrewApps lists Homebrew casks (macOS)
func (t *AppsTool) getHomebrewApps(ctx context.Context) ([]map[string]interface{}, []string) {
	var apps []map[string]interface{}

	cmd := exec.CommandContext(ctx, "brew", "list", "--cask")
	output, err := cmd.Output()
	if err != nil {
		return apps, []string{}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			apps = append(apps, map[string]interface{}{
				"name":   line,
				"source": "homebrew",
			})
		}
	}

	return apps, []string{"homebrew"}
}

// getAPTApps lists installed packages via apt (Linux)
func (t *AppsTool) getAPTApps(ctx context.Context) ([]map[string]interface{}, []string) {
	var apps []map[string]interface{}

	cmd := exec.CommandContext(ctx, "dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return apps, []string{}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ii ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				apps = append(apps, map[string]interface{}{
					"name":   fields[1],
					"source": "apt",
				})
			}
		}
	}

	return apps, []string{"apt"}
}

// scanDirectory scans a directory for files with a specific extension
func (t *AppsTool) scanDirectory(dir, ext string) []map[string]interface{} {
	var apps []map[string]interface{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return apps
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ext) {
			name := entry.Name()
			// Remove extension for cleaner display
			if ext != "" {
				name = strings.TrimSuffix(name, ext)
			}

			apps = append(apps, map[string]interface{}{
				"name":   name,
				"path":   filepath.Join(dir, entry.Name()),
				"source": dir,
			})
		}
	}

	return apps
}

// filterApps filters applications by name
func (t *AppsTool) filterApps(apps []map[string]interface{}, filter string) []map[string]interface{} {
	var filtered []map[string]interface{}
	filterLower := toLower(filter)

	for _, app := range apps {
		name, _ := app["name"].(string)
		if contains(name, filterLower) {
			filtered = append(filtered, app)
		}
	}

	return filtered
}
