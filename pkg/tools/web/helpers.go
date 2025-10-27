package web

import (
	"context"
	"net"
	"fmt"
	"time"
)

// createSecureDialContext creates a DialContext function with security checks
func createSecureDialContext(allowPrivateIPs bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Extract host from address
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		// Resolve IP address
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, err
		}

		// Check for private IPs if not allowed
		if !allowPrivateIPs {
			for _, ip := range ips {
				if isPrivateIP(ip) {
					return nil, fmt.Errorf("requests to private IP addresses are not allowed: %s", ip.String())
				}
			}
		}

		// Use default dialer
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		return dialer.DialContext(ctx, network, addr)
	}
}
