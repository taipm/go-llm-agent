# Network Tools

Professional network diagnostic and information gathering tools for Go LLM Agent.

## Overview

The network tools package provides 5 powerful tools for network operations using industry-standard libraries:

1. **DNS Lookup** - Query DNS records using `miekg/dns` (used by Cloudflare, CoreDNS)
2. **Ping/Connectivity Check** - ICMP ping and TCP port testing using `go-ping/ping`
3. **Whois Lookup** - Domain/IP ownership information using `likexian/whois-parser`
4. **SSL Certificate Check** - TLS/SSL certificate validation using Go's `crypto/tls`
5. **IP Geolocation** - IP address geolocation using `oschwald/geoip2-golang` (MaxMind)

## Tools

### 1. DNS Lookup (`network_dns_lookup`)

Query various DNS record types with professional DNS library.

**Supported Record Types:**
- A (IPv4 addresses)
- AAAA (IPv6 addresses)
- MX (Mail exchange servers)
- TXT (Text records, SPF, DKIM, etc.)
- NS (Name servers)
- CNAME (Canonical names)
- SOA (Start of authority)
- PTR (Reverse DNS)
- ALL (Query all types)

**Example:**
```json
{
  "domain": "google.com",
  "type": "A"
}
```

**Features:**
- Custom DNS servers (Google DNS, Cloudflare DNS, OpenDNS by default)
- TCP/UDP support
- TTL information
- SOA record parsing
- Reverse DNS (PTR) lookups

### 2. Ping/Connectivity Check (`network_ping`)

Test network connectivity and measure latency.

**Modes:**
- `icmp` - ICMP ping (default)
- `tcp` - TCP port connectivity check

**Example ICMP:**
```json
{
  "host": "google.com",
  "mode": "icmp",
  "count": 4
}
```

**Example TCP:**
```json
{
  "host": "google.com",
  "mode": "tcp",
  "port": 443
}
```

**Features:**
- Packet loss calculation
- RTT (Round Trip Time) statistics (min, max, avg, stddev)
- TCP port availability testing
- Connection latency measurement

### 3. Whois Lookup (`network_whois_lookup`)

Get domain registration and IP ownership information.

**Example:**
```json
{
  "domain": "google.com"
}
```

**Returns:**
- Domain registration dates (created, updated, expires)
- Registrar information
- Registrant (owner) details
- Administrative contact
- Technical contact
- Name servers
- Domain status

**Features:**
- Automatic parsing of WHOIS data
- Raw text output option
- Works with domains and IP addresses

### 4. SSL Certificate Check (`network_ssl_cert_check`)

Validate and inspect SSL/TLS certificates.

**Example:**
```json
{
  "host": "google.com",
  "port": 443
}
```

**Returns:**
- Certificate subject and issuer
- Validity period (not before, not after)
- Days until expiration
- Subject Alternative Names (SANs)
- Certificate chain
- TLS version and cipher suite
- Key usage information

**Features:**
- Certificate chain validation
- Expiration warnings (30 days default)
- Support for custom ports
- TLS version detection (TLS 1.0-1.3)

### 5. IP Geolocation (`network_ip_info`)

Get detailed information about IP addresses.

**Example:**
```json
{
  "ip": "8.8.8.8"
}
```

**Returns (Basic):**
- IP version (IPv4/IPv6)
- Private/Public status
- Loopback/Multicast status
- Reverse DNS (PTR records)

**Returns (with GeoIP database):**
- Country, city, continent
- Coordinates (latitude, longitude)
- Time zone
- ISP and ASN information
- Postal code
- Subdivisions (states/provinces)

**GeoIP Database Setup:**

To enable geolocation features, download MaxMind GeoLite2 database:

1. Sign up at https://www.maxmind.com/en/geolite2/signup
2. Download GeoLite2-City.mmdb
3. Configure the tool:

```go
config := builtin.DefaultConfig()
config.Network.IPInfo.GeoIPDatabasePath = "/path/to/GeoLite2-City.mmdb"
config.Network.IPInfo.EnableGeolocation = true

registry := builtin.GetRegistryWithConfig(config)
```

**Note:** Without the GeoIP database, the tool will still work but only provide basic IP information (version, private/public status, reverse DNS).

## Configuration

### Default Configuration

All network tools come with sensible defaults:

```go
config := builtin.DefaultConfig()

// DNS: Google DNS, Cloudflare DNS, OpenDNS
// Ping: 4 packets, 1s interval, 10s timeout
// Whois: 30s timeout, auto-parse results
// SSL: 30s timeout, check expiry, warn if < 30 days
// IPInfo: GeoIP disabled (requires manual setup)
```

### Custom Configuration

```go
config := builtin.DefaultConfig()

// Use custom DNS servers
config.Network.DNS.DNSServers = []string{"8.8.8.8:53"}
config.Network.DNS.UseTCP = true

// More aggressive ping
config.Network.Ping.Count = 10
config.Network.Ping.Interval = 500 * time.Millisecond

// Skip certificate verification (testing only!)
config.Network.SSL.SkipVerify = true

// Enable GeoIP
config.Network.IPInfo.EnableGeolocation = true
config.Network.IPInfo.GeoIPDatabasePath = "/path/to/GeoLite2-City.mmdb"

registry := builtin.GetRegistryWithConfig(config)
```

### Disable Specific Tools

```go
config := builtin.DefaultConfig()
config.NoNetwork = true // Disable all network tools

// Or disable all tools except network
config.NoFile = true
config.NoWeb = true
config.NoTime = true
config.NoSystem = true
config.NoMath = true
config.NoMongoDB = true
```

## Dependencies

This package uses professional, battle-tested libraries:

- **miekg/dns** - Complete DNS library used by CoreDNS and Cloudflare
- **go-ping/ping** - ICMP ping implementation
- **likexian/whois** + **likexian/whois-parser** - WHOIS client and parser
- **oschwald/geoip2-golang** - MaxMind GeoIP2 reader (official)
- **crypto/tls** - Go standard library for TLS/SSL

## Usage in Agent

All network tools are automatically loaded by default:

```go
import (
    "github.com/taipm/go-llm-agent/pkg/agent"
    "github.com/taipm/go-llm-agent/pkg/provider"
)

// Create agent - network tools automatically available!
llm := provider.FromEnv()
a := agent.New(llm)

// Agent can now:
// - Lookup DNS records
// - Ping hosts and check TCP ports
// - Query WHOIS information
// - Validate SSL certificates
// - Get IP geolocation (if GeoIP DB configured)
```

## Security Considerations

1. **DNS**: Uses public DNS servers by default (Google, Cloudflare, OpenDNS)
2. **Ping**: Uses unprivileged mode (works without root)
3. **WHOIS**: Queries public WHOIS servers
4. **SSL**: Validates certificates by default (can disable for testing)
5. **IP Info**: GeoIP database is local (no external API calls)

All tools are **read-only** and **safe** to use - they don't modify any resources.

## Troubleshooting

### Ping requires root privileges

The ping tool uses unprivileged mode by default. If you see permission errors:

```go
// This is already the default, but you can explicitly set:
pinger.SetPrivileged(false)
```

### GeoIP database not found

If you see "GeoIP database file not found":

1. Download from https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
2. Extract GeoLite2-City.mmdb
3. Update config:

```go
config.Network.IPInfo.GeoIPDatabasePath = "/path/to/GeoLite2-City.mmdb"
config.Network.IPInfo.EnableGeolocation = true
```

### DNS queries timeout

If DNS queries fail:

```go
// Try different DNS servers
config.Network.DNS.DNSServers = []string{"1.1.1.1:53"} // Cloudflare only

// Increase timeout
config.Network.DNS.Timeout = 30 * time.Second

// Use TCP instead of UDP
config.Network.DNS.UseTCP = true
```

## Examples

See `examples/` directory for complete working examples.

## License

Same as go-llm-agent project.
