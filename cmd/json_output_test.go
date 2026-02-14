package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/chadmayfield/tempest-cli/internal/display"
	tempest "github.com/chadmayfield/tempest-go"
)

func TestCurrentJSON(t *testing.T) {
	obs := &tempest.StationObservation{
		StationID:                  12345,
		Timestamp:                  time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		AirTemperature:             22.5,
		FeelsLike:                  21.0,
		DewPoint:                   15.0,
		RelativeHumidity:           65.0,
		WindAvg:                    5.0,
		WindGust:                   8.0,
		WindLull:                   2.0,
		WindDirection:              180.0,
		SeaLevelPressure:           1013.25,
		PressureTrend:              "rising",
		UV:                         5.5,
		SolarRadiation:             450.0,
		PrecipAccumDay:             2.5,
		LightningCount3hr:          3,
		LightningStrikeLastDistance: 10.0,
	}

	station := &tempest.Station{
		StationID: 12345,
		Name:      "Home Station",
	}

	sc := &config.StationConfig{
		Token:     "test-token",
		StationID: 12345,
		DeviceID:  67890,
		Name:      "Home",
	}

	// Test metric
	result := currentJSON(obs, station, sc, "metric", false)
	if result.Station.Name != "Home Station" {
		t.Errorf("Station.Name = %q, want %q", result.Station.Name, "Home Station")
	}
	if result.Temperature != 22.5 {
		t.Errorf("Temperature = %f, want 22.5", result.Temperature)
	}
	if result.PressureTrend != "rising" {
		t.Errorf("PressureTrend = %q, want %q", result.PressureTrend, "rising")
	}
	if result.WindLull != 2.0 {
		t.Errorf("WindLull = %f, want 2.0", result.WindLull)
	}
	if result.LightningDistance != 10.0 {
		t.Errorf("LightningDistance = %f, want 10.0", result.LightningDistance)
	}
	if result.WindDirectionCardinal != "S" {
		t.Errorf("WindDirectionCardinal = %q, want %q", result.WindDirectionCardinal, "S")
	}

	// Test imperial conversion
	resultImp := currentJSON(obs, station, sc, "imperial", true)
	if resultImp.Temperature == 22.5 {
		t.Error("imperial temperature should be converted from celsius")
	}
	if resultImp.Units != "imperial" {
		t.Errorf("Units = %q, want %q", resultImp.Units, "imperial")
	}

	// Verify JSON serialization
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if _, ok := decoded["pressure_trend"]; !ok {
		t.Error("missing pressure_trend in JSON output")
	}
}

func TestCurrentJSONFallbackStationName(t *testing.T) {
	obs := &tempest.StationObservation{
		Timestamp: time.Now(),
	}
	station := &tempest.Station{Name: ""} // empty station name
	sc := &config.StationConfig{Name: "Config Name"}

	result := currentJSON(obs, station, sc, "metric", false)
	if result.Station.Name != "Config Name" {
		t.Errorf("expected fallback to config name, got %q", result.Station.Name)
	}
}

func TestForecastJSON(t *testing.T) {
	forecast := &tempest.Forecast{
		StationID: 12345,
		Daily: []tempest.DailyForecast{
			{
				Date:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				HighTemp:     25.0,
				LowTemp:      12.0,
				Conditions:   "Clear",
				Icon:         "clear-day",
				PrecipChance: 10,
				PrecipType:   "rain",
				Sunrise:      time.Date(2024, 1, 15, 7, 30, 0, 0, time.UTC),
				Sunset:       time.Date(2024, 1, 15, 17, 45, 0, 0, time.UTC),
			},
			{
				Date:         time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
				HighTemp:     20.0,
				LowTemp:      8.0,
				Conditions:   "Cloudy",
				Icon:         "cloudy",
				PrecipChance: 60,
			},
			{
				Date:     time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC),
				HighTemp: 18.0,
				LowTemp:  6.0,
			},
		},
	}
	sc := &config.StationConfig{Name: "Test", StationID: 12345, DeviceID: 67890}

	// Metric, 2 days
	result := forecastJSON(forecast, sc, "metric", false, 2)
	if len(result.Days) != 2 {
		t.Errorf("len(Days) = %d, want 2", len(result.Days))
	}
	if result.Days[0].HighTemp != 25.0 {
		t.Errorf("Day0 HighTemp = %f, want 25.0", result.Days[0].HighTemp)
	}
	if result.Days[0].Sunrise == "" {
		t.Error("Day0 Sunrise should be set")
	}
	if result.Days[1].Sunrise != "" {
		t.Error("Day1 Sunrise should be empty (zero time)")
	}

	// Imperial
	resultImp := forecastJSON(forecast, sc, "imperial", true, 1)
	if resultImp.Days[0].HighTemp == 25.0 {
		t.Error("imperial high temp should be converted")
	}

	// Requesting more days than available
	resultAll := forecastJSON(forecast, sc, "metric", false, 10)
	if len(resultAll.Days) != 3 {
		t.Errorf("len(Days) = %d, want 3 (capped by available data)", len(resultAll.Days))
	}

	// JSON round-trip
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
}

func TestHistoryJSON(t *testing.T) {
	obs := []tempest.Observation{
		{
			Timestamp:        time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			AirTemperature:   22.5,
			FeelsLike:        23.0,
			RelativeHumidity: 65.0,
			WindAvg:          3.5,
			WindDirection:    180.0,
			StationPressure:  1013.25,
			RainAccumulation: 0.5,
			UVIndex:          5.0,
		},
		{
			Timestamp:        time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			AirTemperature:   23.0,
			FeelsLike:        24.0,
			RelativeHumidity: 60.0,
			WindAvg:          4.0,
			WindDirection:    200.0,
			StationPressure:  1012.0,
			RainAccumulation: 1.5,
			UVIndex:          6.0,
		},
	}
	sc := &config.StationConfig{Name: "Test", StationID: 12345, DeviceID: 67890}
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)

	result := historyJSON(obs, sc, "metric", start, end, "5m")
	if result.Units != "metric" {
		t.Errorf("Units = %q, want %q", result.Units, "metric")
	}
	if result.Resolution != "5m" {
		t.Errorf("Resolution = %q, want %q", result.Resolution, "5m")
	}
	if len(result.Observations) != 2 {
		t.Fatalf("len(Observations) = %d, want 2", len(result.Observations))
	}
	if result.Observations[0].Temperature != 22.5 {
		t.Errorf("Obs[0].Temperature = %f, want 22.5", result.Observations[0].Temperature)
	}
	if result.Observations[0].WindDirectionCardinal != "S" {
		t.Errorf("Obs[0].WindDirectionCardinal = %q, want %q", result.Observations[0].WindDirectionCardinal, "S")
	}
	if result.Station.StationID != 12345 {
		t.Errorf("Station.StationID = %d, want 12345", result.Station.StationID)
	}

	// Empty observations
	empty := historyJSON(nil, sc, "metric", start, end, "1m")
	if len(empty.Observations) != 0 {
		t.Errorf("expected 0 observations, got %d", len(empty.Observations))
	}

	// JSON round-trip
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
}

func TestStationsJSON(t *testing.T) {
	rows := []display.StationRow{
		{
			ConfigName:   "home",
			StationName:  "Home Station",
			StationID:    12345,
			DeviceID:     67890,
			IsDefault:    true,
			Online:       true,
			LastObserved: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			ConfigName:  "office",
			StationName: "Office",
			StationID:   54321,
			DeviceID:    9876,
			Online:      false,
		},
	}

	result := stationsJSON(rows)
	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}
	if result[0].Status != "online" {
		t.Errorf("result[0].Status = %q, want %q", result[0].Status, "online")
	}
	if result[0].LastObservation == nil {
		t.Error("result[0].LastObservation should not be nil")
	}
	if result[1].Status != "offline" {
		t.Errorf("result[1].Status = %q, want %q", result[1].Status, "offline")
	}
	if result[1].LastObservation != nil {
		t.Error("result[1].LastObservation should be nil (zero time)")
	}

	// JSON round-trip
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}
	if decoded[0]["status"] != "online" {
		t.Error("JSON status mismatch")
	}
}

func TestResolutionLabel(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{time.Minute, "1m"},
		{5 * time.Minute, "5m"},
		{30 * time.Minute, "30m"},
		{3 * time.Hour, "3h"},
		{10 * time.Minute, "10m0s"}, // non-standard falls through to default
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := resolutionLabel(tt.d)
			if got != tt.want {
				t.Errorf("resolutionLabel(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
