package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/taipm/go-llm-agent/pkg/tools"
	"github.com/taipm/go-llm-agent/pkg/types"
)

// SSLConfig contains configuration for SSL/TLS certificate checks
type SSLConfig struct {
	// Timeout for connection
	Timeout time.Duration

	// SkipVerify disables certificate verification (useful for self-signed certs)
	SkipVerify bool

	// CheckExpiry enables expiration warnings
	CheckExpiry bool

	// ExpiryWarningDays warns if certificate expires within this many days
	ExpiryWarningDays int
}

// DefaultSSLConfig provides sensible defaults for SSL operations
var DefaultSSLConfig = SSLConfig{
	Timeout:           30 * time.Second,
	SkipVerify:        false,
	CheckExpiry:       true,
	ExpiryWarningDays: 30,
}

// SSLCertTool performs SSL/TLS certificate checks
type SSLCertTool struct {
	tools.BaseTool
	config SSLConfig
}

// NewSSLCertTool creates a new SSL certificate check tool
func NewSSLCertTool(config SSLConfig) *SSLCertTool {
	return &SSLCertTool{
		BaseTool: tools.NewBaseTool(
			"network_ssl_cert_check",
			"Check SSL/TLS certificates for a domain. Returns certificate details including issuer, expiration date, subject alternative names, and validity status.",
			tools.CategoryNetwork,
			false, // no auth required
			true,  // safe operation (read-only)
		),
		config: config,
	}
}

// Parameters returns the JSON schema for the tool's parameters
func (t *SSLCertTool) Parameters() *types.JSONSchema {
	return &types.JSONSchema{
		Type: "object",
		Properties: map[string]*types.JSONSchema{
			"host": {
				Type:        "string",
				Description: "The hostname to check SSL certificate for (e.g., 'google.com')",
			},
			"port": {
				Type:        "integer",
				Description: "The port to connect to (default: 443)",
			},
		},
		Required: []string{"host"},
	}
}

// Execute performs the SSL certificate check
func (t *SSLCertTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters
	host, port, err := t.extractParams(params)
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%d", host, port)

	// Configure TLS
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: t.config.SkipVerify,
	}

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: t.config.Timeout,
	}

	// Connect with TLS
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"host":    host,
			"port":    port,
			"error":   err.Error(),
		}, nil
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()

	// Build result
	result := map[string]interface{}{
		"success":             true,
		"host":                host,
		"port":                port,
		"tls_version":         tlsVersionString(state.Version),
		"cipher_suite":        tls.CipherSuiteName(state.CipherSuite),
		"server_name":         state.ServerName,
		"negotiated_protocol": state.NegotiatedProtocol,
	}

	// Certificate chain info
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0] // Leaf certificate

		certInfo := t.extractCertificateInfo(cert)
		result["certificate"] = certInfo

		// Check expiration
		if t.config.CheckExpiry {
			now := time.Now()
			daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

			certInfo["days_until_expiry"] = daysUntilExpiry
			certInfo["is_valid"] = now.After(cert.NotBefore) && now.Before(cert.NotAfter)

			if daysUntilExpiry < 0 {
				certInfo["expiry_status"] = "expired"
			} else if daysUntilExpiry <= t.config.ExpiryWarningDays {
				certInfo["expiry_status"] = "expiring_soon"
			} else {
				certInfo["expiry_status"] = "valid"
			}
		}

		// Certificate chain
		if len(state.PeerCertificates) > 1 {
			chain := make([]map[string]interface{}, 0, len(state.PeerCertificates)-1)
			for i := 1; i < len(state.PeerCertificates); i++ {
				chain = append(chain, t.extractCertificateInfo(state.PeerCertificates[i]))
			}
			result["certificate_chain"] = chain
		}

		// Verify certificate chain
		if !t.config.SkipVerify {
			opts := x509.VerifyOptions{
				DNSName:       host,
				Intermediates: x509.NewCertPool(),
			}

			// Add intermediate certificates to pool
			for i := 1; i < len(state.PeerCertificates); i++ {
				opts.Intermediates.AddCert(state.PeerCertificates[i])
			}

			// Verify
			_, err := cert.Verify(opts)
			if err != nil {
				result["verification_error"] = err.Error()
				result["verified"] = false
			} else {
				result["verified"] = true
			}
		}
	}

	return result, nil
}

// extractCertificateInfo extracts information from an X.509 certificate
func (t *SSLCertTool) extractCertificateInfo(cert *x509.Certificate) map[string]interface{} {
	info := map[string]interface{}{
		"subject": map[string]interface{}{
			"common_name":  cert.Subject.CommonName,
			"organization": cert.Subject.Organization,
			"country":      cert.Subject.Country,
		},
		"issuer": map[string]interface{}{
			"common_name":  cert.Issuer.CommonName,
			"organization": cert.Issuer.Organization,
			"country":      cert.Issuer.Country,
		},
		"serial_number":        cert.SerialNumber.String(),
		"not_before":           cert.NotBefore.Format(time.RFC3339),
		"not_after":            cert.NotAfter.Format(time.RFC3339),
		"signature_algorithm":  cert.SignatureAlgorithm.String(),
		"public_key_algorithm": cert.PublicKeyAlgorithm.String(),
		"version":              cert.Version,
	}

	// Subject Alternative Names (SANs)
	if len(cert.DNSNames) > 0 {
		info["dns_names"] = cert.DNSNames
	}
	if len(cert.IPAddresses) > 0 {
		ips := make([]string, len(cert.IPAddresses))
		for i, ip := range cert.IPAddresses {
			ips[i] = ip.String()
		}
		info["ip_addresses"] = ips
	}
	if len(cert.EmailAddresses) > 0 {
		info["email_addresses"] = cert.EmailAddresses
	}

	// Key usage
	var keyUsage []string
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		keyUsage = append(keyUsage, "DigitalSignature")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		keyUsage = append(keyUsage, "KeyEncipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		keyUsage = append(keyUsage, "KeyAgreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		keyUsage = append(keyUsage, "CertSign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		keyUsage = append(keyUsage, "CRLSign")
	}
	if len(keyUsage) > 0 {
		info["key_usage"] = keyUsage
	}

	// Extended key usage
	if len(cert.ExtKeyUsage) > 0 {
		extKeyUsage := make([]string, 0, len(cert.ExtKeyUsage))
		for _, usage := range cert.ExtKeyUsage {
			extKeyUsage = append(extKeyUsage, extKeyUsageString(usage))
		}
		info["ext_key_usage"] = extKeyUsage
	}

	// Is CA
	info["is_ca"] = cert.IsCA

	return info
}

// extractParams extracts and validates parameters
func (t *SSLCertTool) extractParams(params map[string]interface{}) (string, int, error) {
	host, ok := params["host"].(string)
	if !ok || host == "" {
		return "", 0, fmt.Errorf("host parameter is required and must be a non-empty string")
	}

	// Remove protocol if present
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	// Remove path if present
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}

	port := 443 // default HTTPS port
	if portParam, ok := params["port"].(float64); ok {
		port = int(portParam)
	} else if portParam, ok := params["port"].(int); ok {
		port = portParam
	}

	if port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port: %d (must be 1-65535)", port)
	}

	return host, port, nil
}

// tlsVersionString converts TLS version to string
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}

// extKeyUsageString converts extended key usage to string
func extKeyUsageString(usage x509.ExtKeyUsage) string {
	switch usage {
	case x509.ExtKeyUsageAny:
		return "Any"
	case x509.ExtKeyUsageServerAuth:
		return "ServerAuth"
	case x509.ExtKeyUsageClientAuth:
		return "ClientAuth"
	case x509.ExtKeyUsageCodeSigning:
		return "CodeSigning"
	case x509.ExtKeyUsageEmailProtection:
		return "EmailProtection"
	case x509.ExtKeyUsageTimeStamping:
		return "TimeStamping"
	case x509.ExtKeyUsageOCSPSigning:
		return "OCSPSigning"
	default:
		return fmt.Sprintf("Unknown(%d)", usage)
	}
}
