package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchFromTempestd_Success(t *testing.T) {
	type testPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/test" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testPayload{Name: "hello", Value: 42})
	}))
	defer srv.Close()

	result, err := fetchFromTempestd[testPayload](context.Background(), srv.URL, "/api/v1/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "hello" {
		t.Errorf("Name = %q, want %q", result.Name, "hello")
	}
	if result.Value != 42 {
		t.Errorf("Value = %d, want 42", result.Value)
	}
}

func TestFetchFromTempestd_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	type dummy struct{}
	_, err := fetchFromTempestd[dummy](context.Background(), srv.URL, "/api/v1/missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestFetchFromTempestd_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	type dummy struct{ X int }
	_, err := fetchFromTempestd[dummy](context.Background(), srv.URL, "/api/v1/bad")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchFromTempestd_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	type dummy struct{}
	_, err := fetchFromTempestd[dummy](ctx, srv.URL, "/api/v1/test")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
