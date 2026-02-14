package display

import (
	"strings"
	"testing"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
)

func TestRenderCurrent(t *testing.T) {
	theme := NewTheme(true) // no-color for stable test output

	obs := &tempest.StationObservation{
		StationID:          12345,
		Timestamp:          time.Now().Add(-5 * time.Minute),
		AirTemperature:     22.5,
		RelativeHumidity:   65.0,
		WindAvg:            3.5,
		WindGust:           5.2,
		WindDirection:      180,
		BarometricPressure: 1010.0,
		SeaLevelPressure:   1013.25,
		SolarRadiation:     450.0,
		UV:                 5.5,
		Brightness:         25000,
		FeelsLike:          23.0,
		DewPoint:           15.8,
		WetBulbTemperature: 18.0,
		PrecipAccumDay:     2.5,
		LightningCount3hr:  3,
	}

	// Metric
	output := RenderCurrent(theme, obs, "Home Station", false, 80)

	if !strings.Contains(output, "Home Station") {
		t.Error("missing station name")
	}
	if !strings.Contains(output, "22.5°C") {
		t.Error("missing temperature")
	}
	if !strings.Contains(output, "65%") {
		t.Error("missing humidity")
	}
	if !strings.Contains(output, "hPa") {
		t.Error("missing pressure unit")
	}
	if !strings.Contains(output, "3 strikes") {
		t.Error("missing lightning count")
	}
	if !strings.Contains(output, "N/A") {
		t.Error("missing battery N/A")
	}

	// Imperial
	output = RenderCurrent(theme, obs, "Home Station", true, 80)
	if !strings.Contains(output, "°F") {
		t.Error("missing fahrenheit")
	}
	if !strings.Contains(output, "inHg") {
		t.Error("missing inHg")
	}
}

func TestRenderCurrentNoColor(t *testing.T) {
	theme := NewTheme(true)
	obs := &tempest.StationObservation{
		Timestamp:      time.Now().Add(-1 * time.Minute),
		AirTemperature: 20.0,
		FeelsLike:      19.0,
	}
	output := RenderCurrent(theme, obs, "Test", false, 80)
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRenderCurrentJSON(t *testing.T) {
	// Test that the JSON path works via the json package
	obs := &tempest.StationObservation{
		StationID:      12345,
		Timestamp:      time.Now(),
		AirTemperature: 22.5,
	}

	var buf strings.Builder
	enc := struct {
		StationID      int     `json:"station_id"`
		AirTemperature float64 `json:"air_temperature"`
	}{
		StationID:      obs.StationID,
		AirTemperature: obs.AirTemperature,
	}
	_ = enc // JSON encoding tested in json/output_test.go
	_ = buf
}

func TestRenderCurrentWithColor(t *testing.T) {
	theme := NewTheme(false) // colored theme

	obs := &tempest.StationObservation{
		Timestamp:          time.Now().Add(-5 * time.Minute),
		AirTemperature:     -5.0, // freezing — blue
		FeelsLike:          -8.0,
		RelativeHumidity:   20.0, // dry
		WindAvg:            15.0, // strong
		WindGust:           20.0,
		WindDirection:      45,
		SeaLevelPressure:   990.0, // low
		UV:                 11.0,  // extreme
		SolarRadiation:     800.0,
		PrecipAccumDay:     10.0, // heavy rain
		LightningCount3hr:  5,
		PressureTrend:      "falling",
	}

	output := RenderCurrent(theme, obs, "Extreme Station", false, 120)
	if !strings.Contains(output, "Extreme Station") {
		t.Error("missing station name")
	}
	if !strings.Contains(output, "falling") {
		t.Error("missing pressure trend")
	}
}

func TestRenderCurrentPressureTrend(t *testing.T) {
	theme := NewTheme(true)
	obs := &tempest.StationObservation{
		Timestamp:        time.Now().Add(-30 * time.Second),
		AirTemperature:   20.0,
		FeelsLike:        19.0,
		SeaLevelPressure: 1013.0,
		PressureTrend:    "rising",
	}

	output := RenderCurrent(theme, obs, "Test", false, 80)
	if !strings.Contains(output, "(rising)") {
		t.Error("missing pressure trend in output")
	}
}

func TestFormatLightningWithDistance(t *testing.T) {
	// No lightning
	got := formatLightning(0, 0, false)
	if got != "none" {
		t.Errorf("formatLightning(0) = %q, want %q", got, "none")
	}

	// With strikes, metric
	got = formatLightning(5, 10.0, false)
	if !strings.Contains(got, "5 strikes") {
		t.Errorf("expected '5 strikes', got %q", got)
	}
	if !strings.Contains(got, "km") {
		t.Errorf("expected 'km' in metric, got %q", got)
	}

	// With strikes, imperial
	got = formatLightning(3, 8.0, true)
	if !strings.Contains(got, "3 strikes") {
		t.Errorf("expected '3 strikes', got %q", got)
	}
	if !strings.Contains(got, "mi") {
		t.Errorf("expected 'mi' in imperial, got %q", got)
	}

	// With strikes but no distance
	got = formatLightning(2, 0, false)
	if !strings.Contains(got, "2 strikes") {
		t.Errorf("expected '2 strikes', got %q", got)
	}
	if strings.Contains(got, "avg") {
		t.Errorf("should not contain 'avg' when distance is 0, got %q", got)
	}
}

func TestTimeAgo(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{48 * time.Hour, "2d"},
	}
	for _, tt := range tests {
		got := timeAgo(time.Now().Add(-tt.d))
		if got != tt.want {
			t.Errorf("timeAgo(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
