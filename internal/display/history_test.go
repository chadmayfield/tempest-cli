package display

import (
	"strings"
	"testing"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
)

func TestRenderHistory(t *testing.T) {
	theme := NewTheme(true)

	obs := []tempest.Observation{
		{
			Timestamp:        time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			AirTemperature:   22.5,
			RelativeHumidity: 65.0,
			WindAvg:          3.5,
			WindDirection:    180,
			StationPressure:  1013.25,
			RainAccumulation: 0.0,
			UVIndex:          5.0,
			FeelsLike:        23.0,
		},
		{
			Timestamp:        time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			AirTemperature:   23.0,
			RelativeHumidity: 60.0,
			WindAvg:          4.0,
			WindDirection:    200,
			StationPressure:  1012.0,
			RainAccumulation: 1.5,
			UVIndex:          6.0,
			FeelsLike:        24.0,
		},
	}

	output := RenderHistory(theme, obs, false, 120)

	if !strings.Contains(output, "History") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "2 observations") {
		t.Error("missing observation count")
	}
	if !strings.Contains(output, "01-15") {
		t.Error("missing date")
	}
	if !strings.Contains(output, "°C") {
		t.Error("missing temperature unit")
	}
}

func TestRenderHistoryEmpty(t *testing.T) {
	theme := NewTheme(true)

	output := RenderHistory(theme, nil, false, 80)
	if !strings.Contains(output, "No observations") {
		t.Error("missing empty message")
	}
}

func TestRenderHistoryImperial(t *testing.T) {
	theme := NewTheme(true)

	obs := []tempest.Observation{
		{
			Timestamp:        time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			AirTemperature:   22.5,
			RelativeHumidity: 65.0,
			WindAvg:          3.5,
			StationPressure:  1013.25,
			FeelsLike:        23.0,
		},
	}

	output := RenderHistory(theme, obs, true, 120)
	if !strings.Contains(output, "°F") {
		t.Error("missing fahrenheit")
	}
}

func TestRenderHistoryNarrowTerminal(t *testing.T) {
	theme := NewTheme(true)

	obs := []tempest.Observation{
		{
			Timestamp:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			AirTemperature: 22.5,
			FeelsLike:      23.0,
		},
	}

	output := RenderHistory(theme, obs, false, 40)
	if output == "" {
		t.Error("expected non-empty output for narrow terminal")
	}
}
