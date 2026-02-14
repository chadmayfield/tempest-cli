package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const maxResponseBody = 10 * 1024 * 1024 // 10 MB

var tempestdClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
}

// validateServerURL checks that a server URL is a valid HTTP/HTTPS URL.
func validateServerURL(serverURL string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL %q: %w", serverURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid server URL %q: scheme must be http or https", serverURL)
	}
	if u.Host == "" {
		return fmt.Errorf("invalid server URL %q: missing host", serverURL)
	}
	return nil
}

// fetchFromTempestd makes an HTTP GET to a tempestd endpoint and decodes the JSON response.
func fetchFromTempestd[T any](ctx context.Context, serverURL, path string) (*T, error) {
	// Parse and validate the base URL, then resolve the path safely.
	base, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	resolved := base.ResolveReference(ref)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolved.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := tempestdClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("contacting tempestd at %s: %w", serverURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tempestd returned status %d for %s", resp.StatusCode, path)
	}

	limited := io.LimitReader(resp.Body, maxResponseBody)
	var result T
	if err := json.NewDecoder(limited).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response from %s: %w", path, err)
	}

	return &result, nil
}
