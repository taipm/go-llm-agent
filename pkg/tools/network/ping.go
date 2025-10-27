package network

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// PingConfig contains configuration for ping operations
type PingConfig struct {
	// Timeout for ping operations
	Timeout time.Duration

	// Count is the number of ping packets to send
	Count int

	// Interval between ping packets
	Interval time.Duration

	// Size of the ping packet in bytes
	PacketSize int
}

// DefaultPingConfig provides sensible defaults for ping operations
var DefaultPingConfig = PingConfig{
	Timeout:    10 * time.Second,
	Count:      4,
	Interval:   time.Second,
	PacketSize: 64,
}

// PingTool performs network connectivity checks
type PingTool struct {
	tools.BaseTool
	config PingConfig
}

// NewPingTool creates a new ping/connectivity check tool
func NewPingTool(config PingConfig) *PingTool {
	return &PingTool{
		BaseTool: tools.NewBaseTool(
			"network_ping",
			"Perform ICMP ping to check network connectivity and measure latency to a host. Also supports TCP port checking.",
			tools.CategoryNetwork,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *PingTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"host": {
				Type:        "string",
				Description: "The hostname or IP address to ping (e.g., 'google.com' or '8.8.8.8')",
			},
			"mode": {
				Type:        "string",
				Description: "Ping mode: 'icmp' for ICMP ping (default), 'tcp' for TCP port check",
			},
			"port": {
				Type:        "integer",
				Description: "TCP port to check (only used when mode is 'tcp', e.g., 80, 443)",
			},
			"count": {
				Type:        "integer",
				Description: fmt.Sprintf("Number of ping packets to send (default: %d, only for ICMP mode)", t.config.Count),
			},
		},
		Required: []string{"host"},
	}
}

// Execute performs the ping/connectivity check
func (t *PingTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	host, mode, port, count, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	switch mode {
	case "icmp":
		return t.icmpPing(host, count)
	case "tcp":
		return t.tcpCheck(host, port)
	default:
		return nil, fmt.Errorf("invalid mode: %s (must be 'icmp' or 'tcp')", mode)
	}
}

// icmpPing performs ICMP ping
func (t *PingTool) icmpPing(host string, count int) (map[string]interface{}, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pinger: %w", err)
	}

	// Configure pinger
	pinger.Count = count
	pinger.Timeout = t.config.Timeout
	pinger.Interval = t.config.Interval
	pinger.Size = t.config.PacketSize
	pinger.SetPrivileged(false) // Use unprivileged mode (works on most systems)

	// Run ping
	err = pinger.Run()
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	stats := pinger.Statistics()

	result := map[string]interface{}{
		"success":       true,
		"host":          host,
		"mode":          "icmp",
		"ip_addr":       stats.IPAddr.String(),
		"packets_sent":  stats.PacketsSent,
		"packets_recv":  stats.PacketsRecv,
		"packet_loss":   stats.PacketLoss,
		"min_rtt_ms":    stats.MinRtt.Milliseconds(),
		"max_rtt_ms":    stats.MaxRtt.Milliseconds(),
		"avg_rtt_ms":    stats.AvgRtt.Milliseconds(),
		"stddev_rtt_ms": stats.StdDevRtt.Milliseconds(),
	}

	// Add individual packet results
	if len(stats.Rtts) > 0 {
		rtts := make([]float64, len(stats.Rtts))
		for i, rtt := range stats.Rtts {
			rtts[i] = float64(rtt.Microseconds()) / 1000.0 // Convert to ms
		}
		result["rtts_ms"] = rtts
	}

	return result, nil
}

// tcpCheck performs TCP port connectivity check
func (t *PingTool) tcpCheck(host string, port int) (map[string]interface{}, error) {
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port: %d (must be 1-65535)", port)
	}

	address := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	// Attempt TCP connection
	conn, err := net.DialTimeout("tcp", address, t.config.Timeout)
	latency := time.Since(start)

	result := map[string]interface{}{
		"host":       host,
		"port":       port,
		"mode":       "tcp",
		"latency_ms": latency.Milliseconds(),
	}

	if err != nil {
		result["success"] = false
		result["reachable"] = false
		result["error"] = err.Error()

		// Determine error type
		if strings.Contains(err.Error(), "timeout") {
			result["error_type"] = "timeout"
		} else if strings.Contains(err.Error(), "connection refused") {
			result["error_type"] = "connection_refused"
		} else {
			result["error_type"] = "unknown"
		}

		return result, nil // Return result with error details, not error
	}

	defer conn.Close()

	// Get remote address
	result["success"] = true
	result["reachable"] = true
	result["remote_addr"] = conn.RemoteAddr().String()

	return result, nil
}

// extractParams extracts and validates parameters
func (t *PingTool) extractParams(params map[string]interface{}) (string, string, int, int, error) {
	host, ok := params["host"].(string)
	if !ok || host == "" {
		return "", "", 0, 0, fmt.Errorf("host parameter is required and must be a non-empty string")
	}

	mode := "icmp" // default
	if modeParam, ok := params["mode"].(string); ok && modeParam != "" {
		mode = strings.ToLower(modeParam)
	}

	port := 0
	if portParam, ok := params["port"].(float64); ok {
		port = int(portParam)
	} else if portParam, ok := params["port"].(int); ok {
		port = portParam
	}

	count := t.config.Count // default
	if countParam, ok := params["count"].(float64); ok {
		count = int(countParam)
	} else if countParam, ok := params["count"].(int); ok {
		count = countParam
	}

	// Validate count
	if count < 1 {
		count = 1
	} else if count > 100 {
		count = 100 // Max 100 pings
	}

	return host, mode, port, count, nil
}
