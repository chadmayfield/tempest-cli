package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	DefaultStation string                   `mapstructure:"default_station" yaml:"default_station"`
	Units          string                   `mapstructure:"units" yaml:"units"`
	Stations       map[string]StationConfig `mapstructure:"stations" yaml:"stations"`
	Tempestd       TempestdConfig           `mapstructure:"tempestd" yaml:"tempestd,omitempty"`
	ServerURL      string                   `mapstructure:"server" yaml:"server,omitempty"` // flat alias for backward compat
}

// TempestdConfig holds tempestd daemon settings.
type TempestdConfig struct {
	Server string `mapstructure:"server" yaml:"server,omitempty"`
}

// EffectiveServerURL returns the server URL from tempestd.server or the flat server key.
func (c *Config) EffectiveServerURL() string {
	if c.Tempestd.Server != "" {
		return c.Tempestd.Server
	}
	return c.ServerURL
}

// StationConfig holds per-station settings.
type StationConfig struct {
	Token     string `mapstructure:"token" yaml:"token"`
	StationID int    `mapstructure:"station_id" yaml:"station_id"`
	DeviceID  int    `mapstructure:"device_id" yaml:"device_id"`
	Name      string `mapstructure:"name" yaml:"name"`
}

// Load reads the merged config from viper into a Config struct.
// It also applies TEMPEST_TOKEN, TEMPEST_STATION_ID, and TEMPEST_DEVICE_ID env vars
// as overrides for the default station.
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	applyEnvOverrides(&cfg)
	return &cfg, nil
}

// applyEnvOverrides applies TEMPEST_TOKEN, TEMPEST_STATION_ID, and TEMPEST_DEVICE_ID
// as overrides for the default station's configuration.
func applyEnvOverrides(cfg *Config) {
	token := os.Getenv("TEMPEST_TOKEN")
	stationIDStr := os.Getenv("TEMPEST_STATION_ID")
	deviceIDStr := os.Getenv("TEMPEST_DEVICE_ID")

	if token == "" && stationIDStr == "" && deviceIDStr == "" {
		return
	}

	name := cfg.DefaultStation
	if name == "" {
		name = "default"
		cfg.DefaultStation = name
	}

	if cfg.Stations == nil {
		cfg.Stations = make(map[string]StationConfig)
	}

	sc := cfg.Stations[name]
	if token != "" {
		sc.Token = token
	}
	if stationIDStr != "" {
		if id, err := strconv.Atoi(stationIDStr); err == nil {
			sc.StationID = id
		}
	}
	if deviceIDStr != "" {
		if id, err := strconv.Atoi(deviceIDStr); err == nil {
			sc.DeviceID = id
		}
	}
	cfg.Stations[name] = sc
}

// Validate checks for required fields.
func (c *Config) Validate() error {
	if len(c.Stations) == 0 {
		return fmt.Errorf("no stations configured; run 'tempest config init' to set up")
	}
	if c.DefaultStation != "" {
		if _, ok := c.Stations[c.DefaultStation]; !ok {
			return fmt.Errorf("default_station %q not found in stations", c.DefaultStation)
		}
	}
	if c.Units != "" && c.Units != "metric" && c.Units != "imperial" {
		return fmt.Errorf("units must be 'metric' or 'imperial', got %q", c.Units)
	}
	return nil
}

// ResolveStation returns the StationConfig for the given name, or the default.
// If no name is given and no default_station is set, falls back to the first configured station.
func (c *Config) ResolveStation(name string) (*StationConfig, error) {
	if name == "" {
		name = c.DefaultStation
	}
	if name == "" {
		// Fall back to first configured station
		names := c.StationNames()
		if len(names) == 0 {
			return nil, fmt.Errorf("no stations configured; run 'tempest config init' to set up")
		}
		name = names[0]
	}

	sc, ok := c.Stations[name]
	if !ok {
		available := c.StationNames()
		return nil, fmt.Errorf("station %q not found; available: %s", name, strings.Join(available, ", "))
	}
	return &sc, nil
}

// StationNames returns sorted station names.
func (c *Config) StationNames() []string {
	names := make([]string, 0, len(c.Stations))
	for name := range c.Stations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsImperial returns true if the configured units are imperial.
func (c *Config) IsImperial() bool {
	return strings.EqualFold(c.Units, "imperial")
}

// RedactToken returns a token with the first 4 characters visible and 4 stars.
func RedactToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return token[:4] + "****"
}
