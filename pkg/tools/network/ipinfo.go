package network

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// IPInfoConfig contains configuration for IP information lookups
type IPInfoConfig struct {
	// GeoIPDatabasePath is the path to MaxMind GeoIP2/GeoLite2 database file
	// Download from: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
	GeoIPDatabasePath string

	// EnableGeolocation enables GeoIP lookups (requires database file)
	EnableGeolocation bool
}

// DefaultIPInfoConfig provides sensible defaults for IP info operations
var DefaultIPInfoConfig = IPInfoConfig{
	GeoIPDatabasePath: "", // No default path, user must provide
	EnableGeolocation: false,
}

// IPInfoTool performs IP address information lookups
type IPInfoTool struct {
	tools.BaseTool
	config  IPInfoConfig
	geoipDB *geoip2.Reader
}

// NewIPInfoTool creates a new IP information lookup tool
func NewIPInfoTool(config IPInfoConfig) (*IPInfoTool, error) {
	tool := &IPInfoTool{
		BaseTool: tools.NewBaseTool(
			"network_ip_info",
			"Get detailed information about an IP address including geolocation (city, country, coordinates), ISP, ASN, and reverse DNS. Requires MaxMind GeoIP2 database for geolocation features.",
			tools.CategoryNetwork,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
	}

	// Load GeoIP database if path provided and geolocation enabled
	if config.EnableGeolocation && config.GeoIPDatabasePath != "" {
		// Check if file exists
		if _, err := os.Stat(config.GeoIPDatabasePath); err != nil {
			return nil, fmt.Errorf("GeoIP database file not found: %s (download from https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)", config.GeoIPDatabasePath)
		}

		db, err := geoip2.Open(config.GeoIPDatabasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open GeoIP database: %w", err)
		}
		tool.geoipDB = db
	}

	return tool, nil
}

// Close closes the GeoIP database
func (t *IPInfoTool) Close() error {
	if t.geoipDB != nil {
		return t.geoipDB.Close()
	}
	return nil
}

// Parameters returns the JSON schema for the tool's parameters
func (t *IPInfoTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"ip": {
				Type:        "string",
				Description: "The IP address to lookup (IPv4 or IPv6, e.g., '8.8.8.8' or '2001:4860:4860::8888')",
			},
		},
		Required: []string{"ip"},
	}
}

// Execute performs the IP information lookup
func (t *IPInfoTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	ipStr, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Parse IP address
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Build result
	result := map[string]interface{}{
		"success": true,
		"ip":      ipStr,
	}

	// Determine IP version
	if ip.To4() != nil {
		result["version"] = "IPv4"
	} else {
		result["version"] = "IPv6"
	}

	// Check if private IP
	result["is_private"] = isPrivateIPAddress(ip)
	result["is_loopback"] = ip.IsLoopback()
	result["is_multicast"] = ip.IsMulticast()

	// Reverse DNS lookup
	if names, err := net.LookupAddr(ipStr); err == nil && len(names) > 0 {
		result["hostnames"] = names
	}

	// GeoIP lookup if available
	if t.geoipDB != nil {
		geoInfo, err := t.lookupGeoIP(ip)
		if err == nil {
			result["geolocation"] = geoInfo
		} else {
			result["geolocation_error"] = err.Error()
		}
	} else if t.config.EnableGeolocation {
		result["geolocation_error"] = "GeoIP database not loaded (check configuration)"
	}

	return result, nil
}

// lookupGeoIP performs GeoIP lookup
func (t *IPInfoTool) lookupGeoIP(ip net.IP) (map[string]interface{}, error) {
	// Try City lookup first (includes ASN data)
	city, err := t.geoipDB.City(ip)
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})

	// Country information
	if city.Country.IsoCode != "" {
		info["country"] = map[string]interface{}{
			"name":     city.Country.Names["en"],
			"iso_code": city.Country.IsoCode,
		}
	}

	// City information
	if city.City.Names["en"] != "" {
		info["city"] = city.City.Names["en"]
	}

	// Continent
	if city.Continent.Code != "" {
		info["continent"] = map[string]interface{}{
			"name": city.Continent.Names["en"],
			"code": city.Continent.Code,
		}
	}

	// Subdivisions (states/provinces)
	if len(city.Subdivisions) > 0 {
		subdivisions := make([]map[string]interface{}, len(city.Subdivisions))
		for i, sub := range city.Subdivisions {
			subdivisions[i] = map[string]interface{}{
				"name":     sub.Names["en"],
				"iso_code": sub.IsoCode,
			}
		}
		info["subdivisions"] = subdivisions
	}

	// Coordinates
	if city.Location.Latitude != 0 || city.Location.Longitude != 0 {
		info["location"] = map[string]interface{}{
			"latitude":        city.Location.Latitude,
			"longitude":       city.Location.Longitude,
			"accuracy_radius": city.Location.AccuracyRadius,
			"time_zone":       city.Location.TimeZone,
		}
	}

	// Postal code
	if city.Postal.Code != "" {
		info["postal_code"] = city.Postal.Code
	}

	// Registered country (may differ from country for proxies/VPNs)
	if city.RegisteredCountry.IsoCode != "" && city.RegisteredCountry.IsoCode != city.Country.IsoCode {
		info["registered_country"] = map[string]interface{}{
			"name":     city.RegisteredCountry.Names["en"],
			"iso_code": city.RegisteredCountry.IsoCode,
		}
	}

	// Try ASN lookup
	if asn, err := t.geoipDB.ASN(ip); err == nil {
		info["asn"] = map[string]interface{}{
			"number":       asn.AutonomousSystemNumber,
			"organization": asn.AutonomousSystemOrganization,
		}
	}

	return info, nil
}

// extractParams extracts and validates parameters
func (t *IPInfoTool) extractParams(params map[string]interface{}) (string, error) {
	ip, ok := params["ip"].(string)
	if !ok || ip == "" {
		return "", fmt.Errorf("ip parameter is required and must be a non-empty string")
	}

	return ip, nil
}

// isPrivateIPAddress checks if an IP is private
func isPrivateIPAddress(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"fc00::/7",
		"fe80::/10",
	}

	for _, cidr := range privateRanges {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}

	return false
}
