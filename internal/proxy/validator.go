package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultOutboundProbeURL = "http://ipinfo.io/json"

// Validator probes that a proxy is reachable and returns its outbound IP.
type Validator struct {
	Timeout time.Duration
	// ProbeURL overrides the default outbound-IP probe (ipinfo.io/json).
	// Useful for tests; in production leave empty.
	ProbeURL string
}

// OutboundIP routes a request to a public IP-echo service through p and
// returns the IP observed by that service. This is the IP that any third
// party will see when traffic flows through p.
func (v Validator) OutboundIP(ctx context.Context, p Proxy) (string, error) {
	probeURL := v.ProbeURL
	if probeURL == "" {
		probeURL = defaultOutboundProbeURL
	}
	proxyURL, err := url.Parse(p.URL())
	if err != nil {
		return "", fmt.Errorf("parse proxy url: %w", err)
	}
	client := &http.Client{
		Timeout:   v.Timeout,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	var payload struct {
		IP string `json:"ip"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if payload.IP == "" {
		return "", errors.New("response missing ip field")
	}
	return payload.IP, nil
}
