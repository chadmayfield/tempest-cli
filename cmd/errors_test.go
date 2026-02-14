package cmd

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestWrapAPIError_Nil(t *testing.T) {
	if err := wrapAPIError(nil); err != nil {
		t.Errorf("wrapAPIError(nil) = %v, want nil", err)
	}
}

func TestWrapAPIError_401(t *testing.T) {
	err := wrapAPIError(fmt.Errorf("HTTP 401 Unauthorized"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected auth error message, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "tempestwx.com/settings/tokens") {
		t.Errorf("expected token URL in 401 message, got: %s", err.Error())
	}
}

func TestWrapAPIError_404(t *testing.T) {
	err := wrapAPIError(fmt.Errorf("HTTP 404 Not Found"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "station not found") {
		t.Errorf("expected station not found message, got: %s", err.Error())
	}
}

func TestWrapAPIError_429(t *testing.T) {
	err := wrapAPIError(fmt.Errorf("HTTP 429 Too Many Requests"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected rate limit message, got: %s", err.Error())
	}
}

func TestWrapAPIError_CircuitBreaker(t *testing.T) {
	err := wrapAPIError(fmt.Errorf("circuit breaker is open"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("expected circuit breaker message, got: %s", err.Error())
	}
}

func TestWrapAPIError_NetworkError(t *testing.T) {
	netErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: fmt.Errorf("connection refused"),
	}
	err := wrapAPIError(netErr)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected network error message, got: %s", err.Error())
	}
}

func TestWrapAPIError_Generic(t *testing.T) {
	orig := fmt.Errorf("some random error")
	err := wrapAPIError(orig)
	if err != orig {
		t.Errorf("expected original error to pass through, got: %v", err)
	}
}

func TestWrapConfigError_NoStations(t *testing.T) {
	err := wrapConfigError(fmt.Errorf("no stations configured"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "config init") {
		t.Errorf("expected config init suggestion, got: %s", err.Error())
	}
}

func TestWrapConfigError_Nil(t *testing.T) {
	if err := wrapConfigError(nil); err != nil {
		t.Errorf("wrapConfigError(nil) = %v, want nil", err)
	}
}

func TestWrapConfigError_Passthrough(t *testing.T) {
	orig := fmt.Errorf("some other error")
	err := wrapConfigError(orig)
	if err != orig {
		t.Errorf("expected original error to pass through, got: %v", err)
	}
}
