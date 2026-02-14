package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func clearTempestEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"TEMPEST_TOKEN", "TEMPEST_STATION_ID", "TEMPEST_DEVICE_ID"} {
		if orig, ok := os.LookupEnv(key); ok {
			t.Cleanup(func() { _ = os.Setenv(key, orig) })
		} else {
			t.Cleanup(func() { _ = os.Unsetenv(key) })
		}
		_ = os.Unsetenv(key)
	}
}

func TestLoadFromViper(t *testing.T) {
	clearTempestEnv(t)

	v := viper.New()
	v.SetConfigType("yaml")

	yamlContent := `
default_station: home
units: metric
stations:
  home:
    token: abc123def456
    station_id: 12345
    device_id: 67890
    name: Home Station
  office:
    token: xyz789
    station_id: 54321
    device_id: 9876
    name: Office Station
`
	if err := v.ReadConfig(strings.NewReader(yamlContent)); err != nil {
		t.Fatalf("reading config: %v", err)
	}

	// Merge into global viper for Load()
	for _, key := range v.AllKeys() {
		viper.Set(key, v.Get(key))
	}
	defer viper.Reset()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.DefaultStation != "home" {
		t.Errorf("DefaultStation = %q, want %q", cfg.DefaultStation, "home")
	}
	if cfg.Units != "metric" {
		t.Errorf("Units = %q, want %q", cfg.Units, "metric")
	}
	if len(cfg.Stations) != 2 {
		t.Errorf("len(Stations) = %d, want 2", len(cfg.Stations))
	}

	home, ok := cfg.Stations["home"]
	if !ok {
		t.Fatal("missing 'home' station")
	}
	if home.StationID != 12345 {
		t.Errorf("home.StationID = %d, want 12345", home.StationID)
	}
	if home.Token != "abc123def456" {
		t.Errorf("home.Token = %q, want %q", home.Token, "abc123def456")
	}
}

func TestPrecedence(t *testing.T) {
	clearTempestEnv(t)
	viper.Reset()
	defer viper.Reset()

	// 1. Config file values (simulated via ReadConfig)
	v := viper.GetViper()
	v.SetConfigType("yaml")
	yamlContent := `
units: metric
default_station: home
stations:
  home:
    token: tok
    station_id: 1
`
	if err := v.ReadConfig(strings.NewReader(yamlContent)); err != nil {
		t.Fatalf("ReadConfig error: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Units != "metric" {
		t.Errorf("expected metric from config file, got %q", cfg.Units)
	}

	// 2. Env var overrides config file
	viper.SetEnvPrefix("TEMPEST")
	viper.AutomaticEnv()
	_ = viper.BindEnv("units")
	_ = os.Setenv("TEMPEST_UNITS", "imperial")
	defer func() { _ = os.Unsetenv("TEMPEST_UNITS") }()

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Units != "imperial" {
		t.Errorf("expected imperial from env override, got %q", cfg.Units)
	}

	// 3. viper.Set (flag equivalent) overrides env
	viper.Set("units", "metric")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Units != "metric" {
		t.Errorf("expected metric from flag override, got %q", cfg.Units)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				DefaultStation: "home",
				Units:          "metric",
				Stations: map[string]StationConfig{
					"home": {Token: "tok", StationID: 1},
				},
			},
		},
		{
			name: "no stations",
			cfg: Config{
				DefaultStation: "home",
				Units:          "metric",
			},
			wantErr: true,
		},
		{
			name: "missing default_station is valid (falls back to first)",
			cfg: Config{
				Units: "metric",
				Stations: map[string]StationConfig{
					"home": {Token: "tok", StationID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "default not in stations",
			cfg: Config{
				DefaultStation: "missing",
				Units:          "metric",
				Stations: map[string]StationConfig{
					"home": {Token: "tok", StationID: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "bad units",
			cfg: Config{
				DefaultStation: "home",
				Units:          "kelvin",
				Stations: map[string]StationConfig{
					"home": {Token: "tok", StationID: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "empty units is valid",
			cfg: Config{
				DefaultStation: "home",
				Stations: map[string]StationConfig{
					"home": {Token: "tok", StationID: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveStation(t *testing.T) {
	cfg := &Config{
		DefaultStation: "home",
		Stations: map[string]StationConfig{
			"home":   {Token: "tok1", StationID: 1},
			"office": {Token: "tok2", StationID: 2},
		},
	}

	// Explicit name
	sc, err := cfg.ResolveStation("office")
	if err != nil {
		t.Fatalf("ResolveStation(office) error: %v", err)
	}
	if sc.StationID != 2 {
		t.Errorf("StationID = %d, want 2", sc.StationID)
	}

	// Empty falls back to default
	sc, err = cfg.ResolveStation("")
	if err != nil {
		t.Fatalf("ResolveStation('') error: %v", err)
	}
	if sc.StationID != 1 {
		t.Errorf("StationID = %d, want 1", sc.StationID)
	}

	// Unknown station
	_, err = cfg.ResolveStation("unknown")
	if err == nil {
		t.Error("expected error for unknown station")
	}

	// No default_station — falls back to first configured (alphabetically)
	cfgNoDefault := &Config{
		Stations: map[string]StationConfig{
			"bravo": {Token: "tok3", StationID: 3},
			"alpha": {Token: "tok4", StationID: 4},
		},
	}
	sc, err = cfgNoDefault.ResolveStation("")
	if err != nil {
		t.Fatalf("ResolveStation('') with no default error: %v", err)
	}
	if sc.StationID != 4 {
		t.Errorf("expected fallback to 'alpha' (StationID=4), got StationID=%d", sc.StationID)
	}
}

func TestRedactToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc123def456", "abc1****"},
		{"abcde", "abcd****"},
		{"abcd", "****"},
		{"ab", "****"},
		{"", "****"},
	}
	for _, tt := range tests {
		got := RedactToken(tt.input)
		if got != tt.want {
			t.Errorf("RedactToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStationNamesSorted(t *testing.T) {
	cfg := &Config{
		Stations: map[string]StationConfig{
			"zebra":  {},
			"alpha":  {},
			"middle": {},
		},
	}

	names := cfg.StationNames()
	if len(names) != 3 {
		t.Fatalf("len(StationNames) = %d, want 3", len(names))
	}
	if names[0] != "alpha" || names[1] != "middle" || names[2] != "zebra" {
		t.Errorf("StationNames not sorted: %v", names)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Setenv("TEMPEST_TOKEN", "env-token")
	t.Setenv("TEMPEST_STATION_ID", "99999")
	t.Setenv("TEMPEST_DEVICE_ID", "88888")

	cfg := &Config{
		DefaultStation: "home",
		Stations: map[string]StationConfig{
			"home": {Token: "file-token", StationID: 1, DeviceID: 2},
		},
	}

	applyEnvOverrides(cfg)

	home := cfg.Stations["home"]
	if home.Token != "env-token" {
		t.Errorf("Token = %q, want %q", home.Token, "env-token")
	}
	if home.StationID != 99999 {
		t.Errorf("StationID = %d, want 99999", home.StationID)
	}
	if home.DeviceID != 88888 {
		t.Errorf("DeviceID = %d, want 88888", home.DeviceID)
	}
}

func TestApplyEnvOverrides_CreatesDefault(t *testing.T) {
	t.Setenv("TEMPEST_TOKEN", "new-token")
	t.Setenv("TEMPEST_STATION_ID", "111")

	cfg := &Config{}
	applyEnvOverrides(cfg)

	if cfg.DefaultStation != "default" {
		t.Errorf("DefaultStation = %q, want 'default'", cfg.DefaultStation)
	}
	sc, ok := cfg.Stations["default"]
	if !ok {
		t.Fatal("expected 'default' station to be created")
	}
	if sc.Token != "new-token" {
		t.Errorf("Token = %q, want 'new-token'", sc.Token)
	}
	if sc.StationID != 111 {
		t.Errorf("StationID = %d, want 111", sc.StationID)
	}
}

func TestEffectiveServerURL(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		want     string
	}{
		{
			name: "tempestd.server takes priority",
			cfg:  Config{Tempestd: TempestdConfig{Server: "http://nested:8080"}, ServerURL: "http://flat:9090"},
			want: "http://nested:8080",
		},
		{
			name: "falls back to flat server",
			cfg:  Config{ServerURL: "http://flat:9090"},
			want: "http://flat:9090",
		},
		{
			name: "both empty",
			cfg:  Config{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.EffectiveServerURL()
			if got != tt.want {
				t.Errorf("EffectiveServerURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsImperial(t *testing.T) {
	if (&Config{Units: "imperial"}).IsImperial() != true {
		t.Error("expected imperial")
	}
	if (&Config{Units: "Imperial"}).IsImperial() != true {
		t.Error("expected Imperial to be imperial")
	}
	if (&Config{Units: "metric"}).IsImperial() != false {
		t.Error("expected metric to not be imperial")
	}
}
