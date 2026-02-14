package cmd

import (
	"bytes"
	"testing"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/spf13/viper"
)

func TestDisplayUnits(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "metric (default)"},
		{"metric", "metric"},
		{"imperial", "imperial"},
	}
	for _, tt := range tests {
		got := displayUnits(tt.input)
		if got != tt.want {
			t.Errorf("displayUnits(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRedactedConfig(t *testing.T) {
	cfg := &config.Config{
		DefaultStation: "home",
		Units:          "imperial",
		Stations: map[string]config.StationConfig{
			"home": {
				Token:     "abcdefgh12345",
				StationID: 12345,
				DeviceID:  67890,
				Name:      "Home Station",
			},
			"office": {
				Token:     "xyz",
				StationID: 54321,
				DeviceID:  9876,
				Name:      "Office",
			},
		},
	}

	result := redactedConfig(cfg)

	if result["default_station"] != "home" {
		t.Errorf("default_station = %v, want home", result["default_station"])
	}
	if result["units"] != "imperial" {
		t.Errorf("units = %v, want imperial", result["units"])
	}

	stations, ok := result["stations"].(map[string]redactedStationConfig)
	if !ok {
		t.Fatal("stations not found or wrong type")
	}
	if stations["home"].Token != "abcd****" {
		t.Errorf("home token = %q, want %q", stations["home"].Token, "abcd****")
	}
	if stations["office"].Token != "****" {
		t.Errorf("office token = %q, want %q (short token redaction)", stations["office"].Token, "****")
	}
	if stations["home"].StationID != 12345 {
		t.Errorf("home StationID = %d, want 12345", stations["home"].StationID)
	}
}

func TestRunConfigShow(t *testing.T) {
	// Set up viper with a valid config
	viper.Reset()
	defer viper.Reset()

	viper.SetConfigType("yaml")
	_ = viper.ReadConfig(bytes.NewBufferString(`
default_station: home
units: imperial
stations:
  home:
    token: testtoken1234
    station_id: 12345
    device_id: 67890
    name: Home Station
`))

	var out bytes.Buffer
	configShowCmd.SetOut(&out)
	configShowCmd.SetErr(&bytes.Buffer{})

	err := runConfigShow(configShowCmd, nil)
	if err != nil {
		t.Fatalf("runConfigShow error: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}
