package system

import (
	"context"
	"fmt"
	"sort"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// ProcessesTool lists running processes on the system
type ProcessesTool struct {
	tools.BaseTool
}

// NewProcessesTool creates a new tool for listing running processes
func NewProcessesTool() *ProcessesTool {
	return &ProcessesTool{
		BaseTool: tools.NewBaseTool(
			"system_processes",
			"List running processes on the system. Returns process ID, name, CPU usage, memory usage, and status. Can filter and sort results.",
			tools.CategorySystem,
			false, // no auth required
			true,  // safe operation (read-only)
		),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *ProcessesTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"limit": {
				Type:        "integer",
				Description: "Maximum number of processes to return. Default is 50. Set to 0 for all processes.",
			},
			"sort_by": {
				Type:        "string",
				Description: "Field to sort by: pid, name, cpu, memory. Default is memory (descending).",
				Enum:        []interface{}{"pid", "name", "cpu", "memory"},
			},
			"name_filter": {
				Type:        "string",
				Description: "Filter processes by name (case-insensitive substring match). Optional.",
			},
			"min_cpu": {
				Type:        "number",
				Description: "Minimum CPU usage percentage to include. Optional.",
			},
			"min_memory": {
				Type:        "number",
				Description: "Minimum memory usage in MB to include. Optional.",
			},
		},
		Required: []string{},
	}
}

// Execute retrieves the list of running processes
func (t *ProcessesTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	config, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Get all processes
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// Collect process info
	var processList []map[string]interface{}
	for _, p := range procs {
		procInfo, err := t.getProcessInfo(ctx, p)
		if err != nil {
			// Skip processes we can't access
			continue
		}

		// Apply filters
		if !t.matchesFilters(procInfo, config) {
			continue
		}

		processList = append(processList, procInfo)
	}

	// Sort processes
	t.sortProcesses(processList, config.SortBy)

	// Apply limit
	if config.Limit > 0 && len(processList) > config.Limit {
		processList = processList[:config.Limit]
	}

	return map[string]interface{}{
		"type":      "processes",
		"count":     len(processList),
		"total":     len(procs),
		"processes": processList,
		"sort_by":   config.SortBy,
		"limit":     config.Limit,
	}, nil
}

// processConfig holds the configuration for process listing
type processConfig struct {
	Limit      int
	SortBy     string
	NameFilter string
	MinCPU     float64
	MinMemory  float64
}

// extractParams extracts and validates parameters
func (t *ProcessesTool) extractParams(params map[string]interface{}) (*processConfig, error) {
	config := &processConfig{
		Limit:  50, // default
		SortBy: "memory",
	}

	if limit, ok := params["limit"]; ok {
		if limitFloat, ok := limit.(float64); ok {
			config.Limit = int(limitFloat)
		} else if limitInt, ok := limit.(int); ok {
			config.Limit = limitInt
		}
	}

	if sortBy, ok := params["sort_by"].(string); ok {
		config.SortBy = sortBy
	}

	if nameFilter, ok := params["name_filter"].(string); ok {
		config.NameFilter = nameFilter
	}

	if minCPU, ok := params["min_cpu"]; ok {
		if cpuFloat, ok := minCPU.(float64); ok {
			config.MinCPU = cpuFloat
		}
	}

	if minMemory, ok := params["min_memory"]; ok {
		if memFloat, ok := minMemory.(float64); ok {
			config.MinMemory = memFloat
		}
	}

	return config, nil
}

// getProcessInfo extracts information from a process
func (t *ProcessesTool) getProcessInfo(ctx context.Context, p *process.Process) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// PID (always available)
	info["pid"] = p.Pid

	// Name
	if name, err := p.NameWithContext(ctx); err == nil {
		info["name"] = name
	} else {
		info["name"] = "unknown"
	}

	// CPU Percent
	if cpuPercent, err := p.CPUPercentWithContext(ctx); err == nil {
		info["cpu_percent"] = cpuPercent
	} else {
		info["cpu_percent"] = 0.0
	}

	// Memory Info
	if memInfo, err := p.MemoryInfoWithContext(ctx); err == nil {
		info["memory_mb"] = float64(memInfo.RSS) / (1024 * 1024)
		info["memory_bytes"] = memInfo.RSS
	} else {
		info["memory_mb"] = 0.0
		info["memory_bytes"] = uint64(0)
	}

	// Status
	if status, err := p.StatusWithContext(ctx); err == nil {
		info["status"] = status
	} else {
		info["status"] = "unknown"
	}

	// Username (optional, may fail for system processes)
	if username, err := p.UsernameWithContext(ctx); err == nil {
		info["username"] = username
	}

	// Command line (optional)
	if cmdline, err := p.CmdlineWithContext(ctx); err == nil && cmdline != "" {
		// Truncate long command lines
		if len(cmdline) > 200 {
			cmdline = cmdline[:200] + "..."
		}
		info["cmdline"] = cmdline
	}

	// Create time (optional)
	if createTime, err := p.CreateTimeWithContext(ctx); err == nil {
		info["create_time"] = createTime
	}

	return info, nil
}

// matchesFilters checks if a process matches the filter criteria
func (t *ProcessesTool) matchesFilters(procInfo map[string]interface{}, config *processConfig) bool {
	// Name filter
	if config.NameFilter != "" {
		name, _ := procInfo["name"].(string)
		if !contains(name, config.NameFilter) {
			return false
		}
	}

	// CPU filter
	if config.MinCPU > 0 {
		cpu, _ := procInfo["cpu_percent"].(float64)
		if cpu < config.MinCPU {
			return false
		}
	}

	// Memory filter
	if config.MinMemory > 0 {
		memory, _ := procInfo["memory_mb"].(float64)
		if memory < config.MinMemory {
			return false
		}
	}

	return true
}

// sortProcesses sorts the process list by the specified field
func (t *ProcessesTool) sortProcesses(processes []map[string]interface{}, sortBy string) {
	sort.Slice(processes, func(i, j int) bool {
		switch sortBy {
		case "pid":
			pid1, _ := processes[i]["pid"].(int32)
			pid2, _ := processes[j]["pid"].(int32)
			return pid1 < pid2
		case "name":
			name1, _ := processes[i]["name"].(string)
			name2, _ := processes[j]["name"].(string)
			return name1 < name2
		case "cpu":
			cpu1, _ := processes[i]["cpu_percent"].(float64)
			cpu2, _ := processes[j]["cpu_percent"].(float64)
			return cpu1 > cpu2 // descending
		case "memory":
			mem1, _ := processes[i]["memory_mb"].(float64)
			mem2, _ := processes[j]["memory_mb"].(float64)
			return mem1 > mem2 // descending
		default:
			return false
		}
	})
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return indexOf(sLower, substrLower) >= 0
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
