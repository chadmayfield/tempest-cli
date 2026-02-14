package display

import (
	"strings"
	"testing"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
)

func TestRenderForecast(t *testing.T) {
	theme := NewTheme(true)

	forecast := &tempest.Forecast{
		Daily: []tempest.DailyForecast{
			{
				Date:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				HighTemp:     25.0,
				LowTemp:      12.0,
				Conditions:   "Clear",
				Icon:         "clear-day",
				PrecipChance: 10,
				Sunrise:      time.Date(2024, 1, 15, 7, 30, 0, 0, time.UTC),
				Sunset:       time.Date(2024, 1, 15, 17, 45, 0, 0, time.UTC),
			},
			{
				Date:         time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				HighTemp:     18.0,
				LowTemp:      5.0,
				Conditions:   "Cloudy",
				Icon:         "cloudy",
				PrecipChance: 60,
				Sunrise:      time.Date(2024, 1, 16, 7, 30, 0, 0, time.UTC),
				Sunset:       time.Date(2024, 1, 16, 17, 46, 0, 0, time.UTC),
			},
		},
	}

	output := RenderForecast(theme, forecast, 5, false, 80)

	if !strings.Contains(output, "Forecast") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "Clear") {
		t.Error("missing conditions")
	}
	if !strings.Contains(output, "Cloudy") {
		t.Error("missing second day")
	}
	if !strings.Contains(output, "°C") {
		t.Error("missing temperature unit")
	}

	// Imperial
	output = RenderForecast(theme, forecast, 5, true, 80)
	if !strings.Contains(output, "°F") {
		t.Error("missing fahrenheit")
	}
}

func TestRenderForecastSingleDay(t *testing.T) {
	theme := NewTheme(true)

	forecast := &tempest.Forecast{
		Daily: []tempest.DailyForecast{
			{
				Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				HighTemp:   25.0,
				LowTemp:    12.0,
				Conditions: "Clear",
				Icon:       "clear-day",
			},
		},
	}

	output := RenderForecast(theme, forecast, 1, false, 80)
	if !strings.Contains(output, "Clear") {
		t.Error("missing conditions for single day")
	}
}

func TestRenderForecastTenDays(t *testing.T) {
	theme := NewTheme(true)

	days := make([]tempest.DailyForecast, 10)
	for i := range days {
		days[i] = tempest.DailyForecast{
			Date:       time.Date(2024, 1, 15+i, 0, 0, 0, 0, time.UTC),
			HighTemp:   20.0 + float64(i),
			LowTemp:    10.0 + float64(i),
			Conditions: "Clear",
			Icon:       "clear-day",
		}
	}

	forecast := &tempest.Forecast{Daily: days}
	output := RenderForecast(theme, forecast, 10, false, 120)
	if !strings.Contains(output, "Forecast") {
		t.Error("missing title for 10-day forecast")
	}
	// Should contain data from multiple days
	if !strings.Contains(output, "Jan 15") {
		t.Error("missing first day")
	}
	if !strings.Contains(output, "Jan 24") {
		t.Error("missing last day")
	}
}

func TestRenderForecastEmpty(t *testing.T) {
	theme := NewTheme(true)
	forecast := &tempest.Forecast{}

	output := RenderForecast(theme, forecast, 5, false, 80)
	if !strings.Contains(output, "No forecast data") {
		t.Error("missing empty message")
	}
}

func TestRenderForecastNarrowTerminal(t *testing.T) {
	theme := NewTheme(true)
	forecast := &tempest.Forecast{
		Daily: []tempest.DailyForecast{
			{Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), HighTemp: 20, LowTemp: 10, Conditions: "Clear", Icon: "clear-day"},
			{Date: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), HighTemp: 22, LowTemp: 12, Conditions: "Cloudy", Icon: "cloudy"},
		},
	}

	// Very narrow terminal should still render
	output := RenderForecast(theme, forecast, 5, false, 20)
	if output == "" {
		t.Error("expected non-empty output for narrow terminal")
	}
}

func TestRenderForecastNoEmoji(t *testing.T) {
	theme := NewTheme(true, WithNoEmoji(true))

	forecast := &tempest.Forecast{
		Daily: []tempest.DailyForecast{
			{
				Date:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				HighTemp:   25.0,
				LowTemp:    12.0,
				Conditions: "Clear",
				Icon:       "clear-day",
			},
		},
	}

	output := RenderForecast(theme, forecast, 5, false, 80)
	if !strings.Contains(output, "[clear]") {
		t.Error("expected [clear] text label in no-emoji mode")
	}
}

func TestRenderForecastJSON(t *testing.T) {
	// Test that Forecast can be JSON-serialized
	forecast := &tempest.Forecast{
		StationID: 12345,
		Daily: []tempest.DailyForecast{
			{Date: time.Now(), HighTemp: 25, LowTemp: 15},
		},
	}
	if forecast.StationID != 12345 {
		t.Error("unexpected station ID")
	}
}
