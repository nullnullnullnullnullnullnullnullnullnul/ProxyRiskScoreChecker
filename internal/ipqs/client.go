// Package ipqs is a thin client for the IPQualityScore reputation API.
//
// See https://www.ipqualityscore.com/documentation/proxy-detection/overview
// for the API surface.
package ipqs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://ipqualityscore.com"
	pathTemplate   = "/api/json/ip/%s/%s"
)

// Response is the subset of the IPQS payload this client consumes.
// The full payload contains many more fields; add them here as needed.
type Response struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	FraudScore int    `json:"fraud_score"`
}

// Client queries the IPQualityScore reputation API.
type Client struct {
	APIKey  string
	Timeout time.Duration
	// BaseURL overrides the production host. Empty means production.
	BaseURL string
}

// CheckIP returns the IPQS response for ip, with the given strictness level
// (typically 0-3). The caller is responsible for interpreting FraudScore.
//
// Returns an error if the HTTP call fails, the response cannot be decoded,
// or IPQS reports `success: false`.
func (c Client) CheckIP(ctx context.Context, ip string, strictness int) (Response, error) {
	base := c.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	endpoint, err := url.Parse(base + fmt.Sprintf(pathTemplate, c.APIKey, ip))
	if err != nil {
		return Response{}, fmt.Errorf("build url: %w", err)
	}
	q := endpoint.Query()
	q.Set("strictness", strconv.Itoa(strictness))
	endpoint.RawQuery = q.Encode()

	client := &http.Client{Timeout: c.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return Response{}, fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read body: %w", err)
	}
	var out Response
	if err := json.Unmarshal(body, &out); err != nil {
		return Response{}, fmt.Errorf("unmarshal: %w", err)
	}
	if !out.Success {
		return out, fmt.Errorf("ipqs reported failure: %s", out.Message)
	}
	return out, nil
}
