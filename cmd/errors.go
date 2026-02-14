package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// wrapAPIError enhances API errors with user-friendly messages.
func wrapAPIError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	// HTTP 401 — invalid token
	if strings.Contains(msg, "401") || strings.Contains(msg, "Unauthorized") {
		return fmt.Errorf("authentication failed: your API token may be invalid or expired.\nCheck your token at https://tempestwx.com/settings/tokens")
	}

	// HTTP 404 — unknown station
	if strings.Contains(msg, "404") || strings.Contains(msg, "Not Found") {
		return fmt.Errorf("station not found: check that your station ID is correct. %w", err)
	}

	// HTTP 429 — rate limited
	if strings.Contains(msg, "429") || strings.Contains(msg, "Too Many Requests") || strings.Contains(msg, "rate limit") {
		return fmt.Errorf("API rate limit exceeded. Try again in a moment")
	}

	// Circuit breaker open
	if strings.Contains(msg, "circuit breaker") || strings.Contains(msg, "breaker is open") {
		return fmt.Errorf("API temporarily unavailable — the circuit breaker is open. Wait a moment and try again")
	}

	// Network errors
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return fmt.Errorf("network error: cannot reach the WeatherFlow API. Check your internet connection: %w", err)
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return fmt.Errorf("request timed out: the WeatherFlow API did not respond in time. Try again later")
		}
		return fmt.Errorf("network error: %w", err)
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Errorf("DNS lookup failed: cannot resolve the WeatherFlow API hostname. Check your internet connection")
	}

	return err
}

// wrapConfigError wraps a config loading error with a helpful suggestion.
func wrapConfigError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "no stations configured") || strings.Contains(msg, "not set in config") {
		return fmt.Errorf("%w\nRun 'tempest config init' to set up your configuration", err)
	}
	return err
}
