package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/spf13/viper"
)

func TestResolveServerURL(t *testing.T) {
	tests := []struct {
		name     string
		viperVal string
		cfg      *config.Config
		want     string
	}{
		{
			name:     "viper flag takes priority",
			viperVal: "http://flag:8080",
			cfg:      &config.Config{ServerURL: "http://config:9090"},
			want:     "http://flag:8080",
		},
		{
			name: "falls back to config",
			cfg:  &config.Config{ServerURL: "http://config:9090"},
			want: "http://config:9090",
		},
		{
			name: "nil config",
			cfg:  nil,
			want: "",
		},
		{
			name: "empty everywhere",
			cfg:  &config.Config{},
			want: "",
		},
		{
			name:     "rejects non-http scheme",
			viperVal: "ftp://evil.com",
			cfg:      &config.Config{},
			want:     "",
		},
		{
			name:     "rejects missing host",
			viperVal: "http://",
			cfg:      &config.Config{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			defer viper.Reset()

			if tt.viperVal != "" {
				viper.Set("server", tt.viperVal)
			}

			got := resolveServerURL(tt.cfg)
			if got != tt.want {
				t.Errorf("resolveServerURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateServerURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"http://localhost:8080", false},
		{"https://tempestd.local:443", false},
		{"http://192.168.1.100:8080", false},
		{"ftp://evil.com", true},
		{"javascript:alert(1)", true},
		{"file:///etc/passwd", true},
		{"not-a-url", true},
		{"http://", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateServerURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServerURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestCheckConfigPermissions(t *testing.T) {
	// Create a temp file with world-readable permissions
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should not panic — just prints a warning to stderr
	checkConfigPermissions(path)

	// Non-existent file — should not panic
	checkConfigPermissions(filepath.Join(tmpDir, "nonexistent"))

	// Restricted permissions — no warning expected
	restrictedPath := filepath.Join(tmpDir, "restricted.yaml")
	if err := os.WriteFile(restrictedPath, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	checkConfigPermissions(restrictedPath)
}
