// Package proxy parses proxy specifications and probes their reachability.
package proxy

import (
	"fmt"
	"regexp"
	"strings"
)

// Proxy is a parsed proxy specification.
type Proxy struct {
	Protocol string // http, https, socks5
	Host     string
	Port     string
	User     string // optional
	Password string // optional
}

// URL returns the canonical URL representation of the proxy.
// For example: "http://user:pass@10.0.0.1:8080" or "socks5://10.0.0.1:1080".
func (p Proxy) URL() string {
	if p.User != "" && p.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%s", p.Protocol, p.User, p.Password, p.Host, p.Port)
	}
	return fmt.Sprintf("%s://%s:%s", p.Protocol, p.Host, p.Port)
}

const defaultProtocol = "http"

var (
	// protocol://user:pass@host:port
	protocolAuthRe = regexp.MustCompile(`^(http|https|socks5)://([^:@/]+):([^@/]+)@([^:/]+):(\d+)$`)
	// protocol://host:port
	protocolNoAuthRe = regexp.MustCompile(`^(http|https|socks5)://([^:/]+):(\d+)$`)
	// user:pass@host:port (no protocol; defaults to http)
	userAuthHostRe = regexp.MustCompile(`^([^:@/]+):([^@/]+)@([^:/]+):(\d+)$`)
	portRe         = regexp.MustCompile(`^\d+$`)
)

// Parse extracts a Proxy from a raw input line.
//
// Supported input formats:
//   - protocol://user:pass@host:port
//   - protocol://host:port
//   - user:pass@host:port         (protocol defaults to http)
//   - host:port:user:pass         (proxy-provider list format; protocol defaults to http)
//   - host:port                   (protocol defaults to http)
//
// where protocol is one of http, https, socks5.
//
// Returns a non-nil error if the input doesn't match any supported format
// or if the resulting host/port pair is invalid.
func Parse(raw string) (Proxy, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Proxy{}, fmt.Errorf("empty input")
	}

	p, err := parseFormats(raw)
	if err != nil {
		return Proxy{}, err
	}
	if err := validate(p); err != nil {
		return Proxy{}, fmt.Errorf("%w: %q", err, raw)
	}
	return p, nil
}

func parseFormats(raw string) (Proxy, error) {
	if m := protocolAuthRe.FindStringSubmatch(raw); m != nil {
		return Proxy{Protocol: m[1], User: m[2], Password: m[3], Host: m[4], Port: m[5]}, nil
	}
	if m := protocolNoAuthRe.FindStringSubmatch(raw); m != nil {
		return Proxy{Protocol: m[1], Host: m[2], Port: m[3]}, nil
	}
	if m := userAuthHostRe.FindStringSubmatch(raw); m != nil {
		return Proxy{Protocol: defaultProtocol, User: m[1], Password: m[2], Host: m[3], Port: m[4]}, nil
	}

	parts := strings.Split(raw, ":")
	switch len(parts) {
	case 2:
		return Proxy{Protocol: defaultProtocol, Host: parts[0], Port: parts[1]}, nil
	case 4:
		return Proxy{
			Protocol: defaultProtocol,
			Host:     parts[0],
			Port:     parts[1],
			User:     parts[2],
			Password: parts[3],
		}, nil
	}
	return Proxy{}, fmt.Errorf("unrecognized proxy format: %q", raw)
}

func validate(p Proxy) error {
	if p.Host == "" {
		return fmt.Errorf("missing host")
	}
	if !portRe.MatchString(p.Port) {
		return fmt.Errorf("invalid port %q", p.Port)
	}
	return nil
}
