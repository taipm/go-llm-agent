package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// DNSConfig contains configuration for DNS lookups
type DNSConfig struct {
	// Timeout for DNS queries
	Timeout time.Duration

	// DNSServers is a list of DNS servers to use (e.g., "8.8.8.8:53", "1.1.1.1:53")
	// If empty, uses system default resolver (via /etc/resolv.conf)
	DNSServers []string

	// UseTCP forces TCP for DNS queries instead of UDP
	UseTCP bool
}

// DefaultDNSConfig provides sensible defaults for DNS operations
var DefaultDNSConfig = DNSConfig{
	Timeout: 10 * time.Second,
	DNSServers: []string{
		"8.8.8.8:53",        // Google Public DNS
		"1.1.1.1:53",        // Cloudflare DNS
		"208.67.222.222:53", // OpenDNS
	},
	UseTCP: false,
}

// DNSLookupTool performs DNS lookups using miekg/dns library
type DNSLookupTool struct {
	tools.BaseTool
	config DNSConfig
	client *dns.Client
}

// NewDNSLookupTool creates a new DNS lookup tool with professional DNS library
func NewDNSLookupTool(config DNSConfig) *DNSLookupTool {
	network := "udp"
	if config.UseTCP {
		network = "tcp"
	}

	client := &dns.Client{
		Net:     network,
		Timeout: config.Timeout,
	}

	return &DNSLookupTool{
		BaseTool: tools.NewBaseTool(
			"network_dns_lookup",
			"Perform DNS lookups to resolve domain names and query various DNS record types (A, AAAA, MX, TXT, NS, CNAME, SOA, PTR). Uses professional DNS library for accurate results.",
			tools.CategoryNetwork,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
		client: client,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *DNSLookupTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"domain": {
				Type:        "string",
				Description: "The domain name or IP address to lookup (e.g., 'google.com' or '8.8.8.8' for reverse DNS)",
			},
			"type": {
				Type:        "string",
				Description: "DNS record type: A, AAAA, MX, TXT, NS, CNAME, SOA, PTR (for reverse DNS), or ALL (default: A)",
			},
			"server": {
				Type:        "string",
				Description: "Optional: Specific DNS server to query (e.g., '8.8.8.8:53'). If not specified, uses configured defaults.",
			},
		},
		Required: []string{"domain"},
	}
}

// Execute performs the DNS lookup
func (t *DNSLookupTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	domain, recordType, server, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	// Ensure domain ends with dot for FQDN
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	// Perform lookup based on type
	result := make(map[string]interface{})
	result["domain"] = strings.TrimSuffix(domain, ".")
	result["success"] = true
	result["server"] = server

	switch recordType {
	case "A":
		records, ttl, err := t.query(domain, dns.TypeA, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "A"
		result["records"] = t.extractARecords(records)
		result["ttl"] = ttl

	case "AAAA":
		records, ttl, err := t.query(domain, dns.TypeAAAA, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "AAAA"
		result["records"] = t.extractAAAARecords(records)
		result["ttl"] = ttl

	case "MX":
		records, ttl, err := t.query(domain, dns.TypeMX, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "MX"
		result["records"] = t.extractMXRecords(records)
		result["ttl"] = ttl

	case "TXT":
		records, ttl, err := t.query(domain, dns.TypeTXT, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "TXT"
		result["records"] = t.extractTXTRecords(records)
		result["ttl"] = ttl

	case "NS":
		records, ttl, err := t.query(domain, dns.TypeNS, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "NS"
		result["records"] = t.extractNSRecords(records)
		result["ttl"] = ttl

	case "CNAME":
		records, ttl, err := t.query(domain, dns.TypeCNAME, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "CNAME"
		cnames := t.extractCNAMERecords(records)
		if len(cnames) > 0 {
			result["cname"] = cnames[0]
		}
		result["ttl"] = ttl

	case "SOA":
		records, ttl, err := t.query(domain, dns.TypeSOA, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "SOA"
		result["record"] = t.extractSOARecord(records)
		result["ttl"] = ttl

	case "PTR":
		// Convert IP to reverse DNS format
		arpa, err := dns.ReverseAddr(strings.TrimSuffix(domain, "."))
		if err != nil {
			return t.errorResult(domain, recordType, server, fmt.Errorf("invalid IP address: %w", err)), nil
		}
		records, ttl, err := t.query(arpa, dns.TypePTR, server)
		if err != nil {
			return t.errorResult(domain, recordType, server, err), nil
		}
		result["type"] = "PTR"
		result["records"] = t.extractPTRRecords(records)
		result["ttl"] = ttl

	case "ALL":
		allRecords := make(map[string]interface{})

		// Query all common record types
		if records, ttl, err := t.query(domain, dns.TypeA, server); err == nil {
			if ips := t.extractARecords(records); len(ips) > 0 {
				allRecords["A"] = map[string]interface{}{"records": ips, "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeAAAA, server); err == nil {
			if ips := t.extractAAAARecords(records); len(ips) > 0 {
				allRecords["AAAA"] = map[string]interface{}{"records": ips, "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeMX, server); err == nil {
			if mxs := t.extractMXRecords(records); len(mxs) > 0 {
				allRecords["MX"] = map[string]interface{}{"records": mxs, "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeTXT, server); err == nil {
			if txts := t.extractTXTRecords(records); len(txts) > 0 {
				allRecords["TXT"] = map[string]interface{}{"records": txts, "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeNS, server); err == nil {
			if nss := t.extractNSRecords(records); len(nss) > 0 {
				allRecords["NS"] = map[string]interface{}{"records": nss, "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeCNAME, server); err == nil {
			if cnames := t.extractCNAMERecords(records); len(cnames) > 0 {
				allRecords["CNAME"] = map[string]interface{}{"cname": cnames[0], "ttl": ttl}
			}
		}
		if records, ttl, err := t.query(domain, dns.TypeSOA, server); err == nil {
			if soa := t.extractSOARecord(records); soa != nil {
				allRecords["SOA"] = map[string]interface{}{"record": soa, "ttl": ttl}
			}
		}

		result["type"] = "ALL"
		result["records"] = allRecords
	}

	return result, nil
}

// query performs the actual DNS query
func (t *DNSLookupTool) query(domain string, qtype uint16, server string) ([]dns.RR, uint32, error) {
	m := new(dns.Msg)
	m.SetQuestion(domain, qtype)
	m.RecursionDesired = true

	// Try configured servers or specified server
	servers := []string{server}
	if server == "" {
		servers = t.config.DNSServers
	}

	var lastErr error
	for _, srv := range servers {
		r, _, err := t.client.Exchange(m, srv)
		if err != nil {
			lastErr = err
			continue
		}

		if r.Rcode != dns.RcodeSuccess {
			lastErr = fmt.Errorf("DNS query failed with code: %s", dns.RcodeToString[r.Rcode])
			continue
		}

		// Get TTL from first answer
		var ttl uint32
		if len(r.Answer) > 0 {
			ttl = r.Answer[0].Header().Ttl
		}

		return r.Answer, ttl, nil
	}

	if lastErr != nil {
		return nil, 0, lastErr
	}
	return nil, 0, fmt.Errorf("no DNS servers responded")
}

// extractParams extracts and validates parameters
func (t *DNSLookupTool) extractParams(params map[string]interface{}) (string, string, string, error) {
	domain, ok := params["domain"].(string)
	if !ok || domain == "" {
		return "", "", "", fmt.Errorf("domain parameter is required and must be a non-empty string")
	}

	recordType := "A" // default
	if typeParam, ok := params["type"].(string); ok && typeParam != "" {
		recordType = strings.ToUpper(typeParam)
	}

	server := ""
	if serverParam, ok := params["server"].(string); ok && serverParam != "" {
		server = serverParam
		// Ensure server has port
		if !strings.Contains(server, ":") {
			server = server + ":53"
		}
	}

	return domain, recordType, server, nil
}

// Extract methods for different record types
func (t *DNSLookupTool) extractARecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if a, ok := rr.(*dns.A); ok {
			results = append(results, a.A.String())
		}
	}
	return results
}

func (t *DNSLookupTool) extractAAAARecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if aaaa, ok := rr.(*dns.AAAA); ok {
			results = append(results, aaaa.AAAA.String())
		}
	}
	return results
}

func (t *DNSLookupTool) extractMXRecords(rrs []dns.RR) []map[string]interface{} {
	var results []map[string]interface{}
	for _, rr := range rrs {
		if mx, ok := rr.(*dns.MX); ok {
			results = append(results, map[string]interface{}{
				"host":     strings.TrimSuffix(mx.Mx, "."),
				"priority": mx.Preference,
			})
		}
	}
	return results
}

func (t *DNSLookupTool) extractTXTRecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if txt, ok := rr.(*dns.TXT); ok {
			results = append(results, strings.Join(txt.Txt, ""))
		}
	}
	return results
}

func (t *DNSLookupTool) extractNSRecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if ns, ok := rr.(*dns.NS); ok {
			results = append(results, strings.TrimSuffix(ns.Ns, "."))
		}
	}
	return results
}

func (t *DNSLookupTool) extractCNAMERecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if cname, ok := rr.(*dns.CNAME); ok {
			results = append(results, strings.TrimSuffix(cname.Target, "."))
		}
	}
	return results
}

func (t *DNSLookupTool) extractPTRRecords(rrs []dns.RR) []string {
	var results []string
	for _, rr := range rrs {
		if ptr, ok := rr.(*dns.PTR); ok {
			results = append(results, strings.TrimSuffix(ptr.Ptr, "."))
		}
	}
	return results
}

func (t *DNSLookupTool) extractSOARecord(rrs []dns.RR) map[string]interface{} {
	for _, rr := range rrs {
		if soa, ok := rr.(*dns.SOA); ok {
			return map[string]interface{}{
				"mname":   strings.TrimSuffix(soa.Ns, "."),
				"rname":   strings.TrimSuffix(soa.Mbox, "."),
				"serial":  soa.Serial,
				"refresh": soa.Refresh,
				"retry":   soa.Retry,
				"expire":  soa.Expire,
				"minimum": soa.Minttl,
			}
		}
	}
	return nil
}

// errorResult creates an error result
func (t *DNSLookupTool) errorResult(domain, recordType, server string, err error) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"domain":  strings.TrimSuffix(domain, "."),
		"type":    recordType,
		"server":  server,
		"error":   err.Error(),
	}
}
