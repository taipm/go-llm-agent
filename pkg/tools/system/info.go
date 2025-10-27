package system

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// InfoTool retrieves system information (CPU, memory, disk, OS, network)
type InfoTool struct {
	tools.BaseTool
}

// NewInfoTool creates a new system information tool
func NewInfoTool() *InfoTool {
	return &InfoTool{
		BaseTool: tools.NewBaseTool(
			"system_info",
			"Get system information including CPU, memory, disk usage, OS details, and network interfaces. Returns read-only system metrics.",
			tools.CategorySystem,
			false, // no auth required
			true,  // safe operation (read-only)
		),
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *InfoTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"type": {
				Type:        "string",
				Description: "Type of system information to retrieve",
				Enum:        []interface{}{"cpu", "memory", "disk", "os", "network", "all"},
			},
			"path": {
				Type:        "string",
				Description: "Path for disk usage query (only used when type=disk). Defaults to root path.",
			},
		},
		Required: []string{"type"},
	}
}

// Execute retrieves system information based on the requested type
func (t *InfoTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	infoType, err := t.extractType(params)
	if err != nil {
		return nil, err
	}

	switch infoType {
	case "cpu":
		return t.getCPUInfo(ctx)
	case "memory":
		return t.getMemoryInfo(ctx)
	case "disk":
		path, _ := params["path"].(string)
		if path == "" {
			path = "/"
		}
		return t.getDiskInfo(ctx, path)
	case "os":
		return t.getOSInfo(ctx)
	case "network":
		return t.getNetworkInfo(ctx)
	case "all":
		return t.getAllInfo(ctx)
	default:
		return nil, fmt.Errorf("invalid info type: %s (valid: cpu, memory, disk, os, network, all)", infoType)
	}
}

// extractType validates and extracts the info type parameter
func (t *InfoTool) extractType(params map[string]interface{}) (string, error) {
	typeVal, ok := params["type"]
	if !ok {
		return "", fmt.Errorf("missing required parameter: type")
	}

	typeStr, ok := typeVal.(string)
	if !ok {
		return "", fmt.Errorf("type must be a string")
	}

	if typeStr == "" {
		return "", fmt.Errorf("type cannot be empty")
	}

	return typeStr, nil
}

// getCPUInfo retrieves CPU information
func (t *InfoTool) getCPUInfo(ctx context.Context) (interface{}, error) {
	// Get CPU info
	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Get CPU usage percentages
	percentages, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get CPU counts
	logicalCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		logicalCount = runtime.NumCPU()
	}
	physicalCount, err := cpu.CountsWithContext(ctx, false)
	if err != nil {
		physicalCount = 0
	}

	result := map[string]interface{}{
		"type":           "cpu",
		"logical_cores":  logicalCount,
		"physical_cores": physicalCount,
		"usage_percent":  0.0,
	}

	if len(percentages) > 0 {
		result["usage_percent"] = percentages[0]
	}

	if len(cpuInfo) > 0 {
		result["model"] = cpuInfo[0].ModelName
		result["mhz"] = cpuInfo[0].Mhz
		result["vendor"] = cpuInfo[0].VendorID
	}

	return result, nil
}

// getMemoryInfo retrieves memory information
func (t *InfoTool) getMemoryInfo(ctx context.Context) (interface{}, error) {
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	return map[string]interface{}{
		"type":            "memory",
		"total_bytes":     vmStat.Total,
		"available_bytes": vmStat.Available,
		"used_bytes":      vmStat.Used,
		"free_bytes":      vmStat.Free,
		"usage_percent":   vmStat.UsedPercent,
		"total_gb":        float64(vmStat.Total) / (1024 * 1024 * 1024),
		"available_gb":    float64(vmStat.Available) / (1024 * 1024 * 1024),
		"used_gb":         float64(vmStat.Used) / (1024 * 1024 * 1024),
	}, nil
}

// getDiskInfo retrieves disk usage information for a given path
func (t *InfoTool) getDiskInfo(ctx context.Context, path string) (interface{}, error) {
	usage, err := disk.UsageWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info for %s: %w", path, err)
	}

	return map[string]interface{}{
		"type":          "disk",
		"path":          usage.Path,
		"total_bytes":   usage.Total,
		"free_bytes":    usage.Free,
		"used_bytes":    usage.Used,
		"usage_percent": usage.UsedPercent,
		"total_gb":      float64(usage.Total) / (1024 * 1024 * 1024),
		"free_gb":       float64(usage.Free) / (1024 * 1024 * 1024),
		"used_gb":       float64(usage.Used) / (1024 * 1024 * 1024),
		"filesystem":    usage.Fstype,
	}, nil
}

// getOSInfo retrieves operating system information
func (t *InfoTool) getOSInfo(ctx context.Context) (interface{}, error) {
	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OS info: %w", err)
	}

	return map[string]interface{}{
		"type":             "os",
		"hostname":         hostInfo.Hostname,
		"os":               hostInfo.OS,
		"platform":         hostInfo.Platform,
		"platform_family":  hostInfo.PlatformFamily,
		"platform_version": hostInfo.PlatformVersion,
		"kernel_version":   hostInfo.KernelVersion,
		"kernel_arch":      hostInfo.KernelArch,
		"uptime_seconds":   hostInfo.Uptime,
		"uptime_hours":     float64(hostInfo.Uptime) / 3600,
		"boot_time":        time.Unix(int64(hostInfo.BootTime), 0).Format(time.RFC3339),
		"go_version":       runtime.Version(),
		"go_os":            runtime.GOOS,
		"go_arch":          runtime.GOARCH,
	}, nil
}

// getNetworkInfo retrieves network interface information
func (t *InfoTool) getNetworkInfo(ctx context.Context) (interface{}, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	var ifaceList []map[string]interface{}
	for _, iface := range interfaces {
		ifaceInfo := map[string]interface{}{
			"name":  iface.Name,
			"mtu":   iface.MTU,
			"flags": iface.Flags,
		}

		// Add addresses
		if len(iface.Addrs) > 0 {
			addrs := make([]string, 0, len(iface.Addrs))
			for _, addr := range iface.Addrs {
				addrs = append(addrs, addr.Addr)
			}
			ifaceInfo["addresses"] = addrs
		}

		// Add hardware address
		if iface.HardwareAddr != "" {
			ifaceInfo["mac"] = iface.HardwareAddr
		}

		ifaceList = append(ifaceList, ifaceInfo)
	}

	return map[string]interface{}{
		"type":       "network",
		"interfaces": ifaceList,
		"count":      len(ifaceList),
	}, nil
}

// getAllInfo retrieves all system information
func (t *InfoTool) getAllInfo(ctx context.Context) (interface{}, error) {
	result := make(map[string]interface{})

	// Get all info types (ignore individual errors, return what we can)
	if cpuInfo, err := t.getCPUInfo(ctx); err == nil {
		result["cpu"] = cpuInfo
	}

	if memInfo, err := t.getMemoryInfo(ctx); err == nil {
		result["memory"] = memInfo
	}

	if diskInfo, err := t.getDiskInfo(ctx, "/"); err == nil {
		result["disk"] = diskInfo
	}

	if osInfo, err := t.getOSInfo(ctx); err == nil {
		result["os"] = osInfo
	}

	if netInfo, err := t.getNetworkInfo(ctx); err == nil {
		result["network"] = netInfo
	}

	result["type"] = "all"
	return result, nil
}
