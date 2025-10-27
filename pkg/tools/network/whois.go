package network

import (
	"context"
	"fmt"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// WhoisConfig contains configuration for whois lookups
type WhoisConfig struct {
	// Timeout for whois queries
	Timeout time.Duration

	// ParseResult enables automatic parsing of whois results
	ParseResult bool
}

// DefaultWhoisConfig provides sensible defaults for whois operations
var DefaultWhoisConfig = WhoisConfig{
	Timeout:     30 * time.Second,
	ParseResult: true,
}

// WhoisLookupTool performs whois lookups for domains and IPs
type WhoisLookupTool struct {
	tools.BaseTool
	config WhoisConfig
}

// NewWhoisLookupTool creates a new whois lookup tool
func NewWhoisLookupTool(config WhoisConfig) *WhoisLookupTool {
	return &WhoisLookupTool{
		BaseTool: tools.NewBaseTool(
			"network_whois_lookup",
			"Perform WHOIS lookups to get domain registration information, registrar details, nameservers, and expiration dates. Also works with IP addresses to get network owner information.",
			tools.CategoryNetwork,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *WhoisLookupTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"domain": {
				Type:        "string",
				Description: "The domain name or IP address to lookup (e.g., 'google.com' or '8.8.8.8')",
			},
			"raw": {
				Type:        "boolean",
				Description: "If true, returns raw WHOIS text instead of parsed data (default: false)",
			},
		},
		Required: []string{"domain"},
	}
}

// Execute performs the whois lookup
func (t *WhoisLookupTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	domain, returnRaw, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Perform whois query (timeout is set via context or globally in whois client)
	rawResult, err := whois.Whois(domain)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"domain":  domain,
			"error":   err.Error(),
		}, nil
	}

	// Return raw result if requested
	if returnRaw || !t.config.ParseResult {
		return map[string]interface{}{
			"success": true,
			"domain":  domain,
			"raw":     rawResult,
		}, nil
	}

	// Parse whois result
	parsed, err := whoisparser.Parse(rawResult)
	if err != nil {
		// If parsing fails, return raw data
		return map[string]interface{}{
			"success":     true,
			"domain":      domain,
			"raw":         rawResult,
			"parse_error": err.Error(),
			"parsed":      false,
		}, nil
	}

	// Build structured result
	result := map[string]interface{}{
		"success": true,
		"domain":  domain,
		"parsed":  true,
	}

	// Domain information
	if parsed.Domain != nil {
		domainInfo := map[string]interface{}{
			"name":         parsed.Domain.Name,
			"punycode":     parsed.Domain.Punycode,
			"status":       parsed.Domain.Status,
			"dnssec":       parsed.Domain.DNSSec,
			"whois_server": parsed.Domain.WhoisServer,
		}

		// Add dates if available
		if parsed.Domain.CreatedDate != "" {
			domainInfo["created_date"] = parsed.Domain.CreatedDate
		}
		if parsed.Domain.UpdatedDate != "" {
			domainInfo["updated_date"] = parsed.Domain.UpdatedDate
		}
		if parsed.Domain.ExpirationDate != "" {
			domainInfo["expiration_date"] = parsed.Domain.ExpirationDate
		}

		// Add nameservers
		if len(parsed.Domain.NameServers) > 0 {
			domainInfo["nameservers"] = parsed.Domain.NameServers
		}

		result["domain_info"] = domainInfo
	}

	// Registrar information
	if parsed.Registrar != nil {
		registrarInfo := map[string]interface{}{
			"name":         parsed.Registrar.Name,
			"organization": parsed.Registrar.Organization,
			"email":        parsed.Registrar.Email,
			"phone":        parsed.Registrar.Phone,
			"referral_url": parsed.Registrar.ReferralURL,
		}
		result["registrar"] = registrarInfo
	}

	// Registrant (domain owner) information
	if parsed.Registrant != nil {
		registrantInfo := map[string]interface{}{
			"name":         parsed.Registrant.Name,
			"organization": parsed.Registrant.Organization,
			"email":        parsed.Registrant.Email,
			"phone":        parsed.Registrant.Phone,
			"country":      parsed.Registrant.Country,
		}
		result["registrant"] = registrantInfo
	}

	// Administrative contact
	if parsed.Administrative != nil {
		adminInfo := map[string]interface{}{
			"name":         parsed.Administrative.Name,
			"organization": parsed.Administrative.Organization,
			"email":        parsed.Administrative.Email,
			"phone":        parsed.Administrative.Phone,
		}
		result["admin_contact"] = adminInfo
	}

	// Technical contact
	if parsed.Technical != nil {
		techInfo := map[string]interface{}{
			"name":         parsed.Technical.Name,
			"organization": parsed.Technical.Organization,
			"email":        parsed.Technical.Email,
			"phone":        parsed.Technical.Phone,
		}
		result["tech_contact"] = techInfo
	}

	// Billing contact
	if parsed.Billing != nil {
		billingInfo := map[string]interface{}{
			"name":         parsed.Billing.Name,
			"organization": parsed.Billing.Organization,
			"email":        parsed.Billing.Email,
			"phone":        parsed.Billing.Phone,
		}
		result["billing_contact"] = billingInfo
	}

	return result, nil
}

// extractParams extracts and validates parameters
func (t *WhoisLookupTool) extractParams(params map[string]interface{}) (string, bool, error) {
	domain, ok := params["domain"].(string)
	if !ok || domain == "" {
		return "", false, fmt.Errorf("domain parameter is required and must be a non-empty string")
	}

	returnRaw := false
	if rawParam, ok := params["raw"].(bool); ok {
		returnRaw = rawParam
	}

	return domain, returnRaw, nil
}
