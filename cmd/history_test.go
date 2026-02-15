package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/spf13/cobra"
)

func TestFetchHistoryFromServer_QueryParams(t *testing.T) {
	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"observations": []any{}})
	}))
	defer srv.Close()

	start := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 13, 0, 0, 0, time.UTC)

	_, err := fetchHistoryFromServer(context.Background(), srv.URL, 1001, start, end, "metric", "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the captured URL and verify start/end are RFC3339, not epoch integers.
	parsed, err := http.NewRequest("GET", capturedURL, nil)
	if err != nil {
		t.Fatalf("failed to parse captured URL: %v", err)
	}
	q := parsed.URL.Query()

	startParam := q.Get("start")
	if _, err := time.Parse(time.RFC3339, startParam); err != nil {
		t.Errorf("start param %q is not RFC3339: %v", startParam, err)
	}
	if startParam != "2024-06-15T12:00:00Z" {
		t.Errorf("start = %q, want %q", startParam, "2024-06-15T12:00:00Z")
	}

	endParam := q.Get("end")
	if _, err := time.Parse(time.RFC3339, endParam); err != nil {
		t.Errorf("end param %q is not RFC3339: %v", endParam, err)
	}
	if endParam != "2024-06-15T13:00:00Z" {
		t.Errorf("end = %q, want %q", endParam, "2024-06-15T13:00:00Z")
	}
}

func TestFetchHistoryFromServer_Deserialization(t *testing.T) {
	// Return a response matching the exact shape from tempestd's API.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"station_id": 1001,
			"start": "2025-12-25T00:00:00Z",
			"end": "2025-12-25T00:06:00Z",
			"resolution": "1m",
			"units": "imperial",
			"total": 1,
			"limit": 1000,
			"offset": 0,
			"observations": [
				{
					"air_temperature": 55.22,
					"battery": 2.51,
					"dew_point": 42.85,
					"feels_like": 55.22,
					"lightning_avg_distance": 0,
					"lightning_strike_count": 0,
					"precipitation_type": 1,
					"rain_accumulation": 0.003,
					"relative_humidity": 63,
					"solar_radiation": 1,
					"station_id": 1001,
					"station_pressure": 24.85,
					"timestamp": "2025-12-25T00:00:00Z",
					"uv_index": 0.01,
					"wind_avg": 4.25,
					"wind_direction": 182,
					"wind_gust": 8.95,
					"wind_lull": 0.22
				}
			]
		}`))
	}))
	defer srv.Close()

	start := time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 25, 0, 6, 0, 0, time.UTC)

	obs, err := fetchHistoryFromServer(context.Background(), srv.URL, 1001, start, end, "imperial", "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("got %d observations, want 1", len(obs))
	}

	o := obs[0]
	if o.AirTemperature != 55.22 {
		t.Errorf("AirTemperature = %f, want 55.22", o.AirTemperature)
	}
	if o.StationID != 1001 {
		t.Errorf("StationID = %d, want 1001", o.StationID)
	}
	if o.WindAvg != 4.25 {
		t.Errorf("WindAvg = %f, want 4.25", o.WindAvg)
	}
	if o.Battery != 2.51 {
		t.Errorf("Battery = %f, want 2.51", o.Battery)
	}
	if o.WindDirection != 182 {
		t.Errorf("WindDirection = %f, want 182", o.WindDirection)
	}
	if o.RelativeHumidity != 63 {
		t.Errorf("RelativeHumidity = %f, want 63", o.RelativeHumidity)
	}
	if o.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestResolveResolution(t *testing.T) {
	tests := []struct {
		flag string
		span time.Duration
		want time.Duration
	}{
		{"1m", 24 * time.Hour, time.Minute},
		{"5m", 24 * time.Hour, 5 * time.Minute},
		{"30m", 24 * time.Hour, 30 * time.Minute},
		{"3h", 24 * time.Hour, 3 * time.Hour},
		{"", 12 * time.Hour, time.Minute},              // ≤1 day → 1m
		{"", 3 * 24 * time.Hour, 5 * time.Minute},      // 1-7 days → 5m
		{"", 14 * 24 * time.Hour, 30 * time.Minute},    // 7-30 days → 30m
		{"", 60 * 24 * time.Hour, 3 * time.Hour},       // 30+ days → 3h
	}

	for _, tt := range tests {
		got := resolveResolution(tt.flag, tt.span)
		if got != tt.want {
			t.Errorf("resolveResolution(%q, %v) = %v, want %v", tt.flag, tt.span, got, tt.want)
		}
	}
}

func TestDownsample(t *testing.T) {
	base := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	obs := make([]tempest.Observation, 60)
	for i := range obs {
		obs[i] = tempest.Observation{
			Timestamp:      base.Add(time.Duration(i) * time.Minute),
			AirTemperature: float64(i),
		}
	}

	// Downsample to 5-minute intervals: should keep ~12 observations
	result := downsample(obs, 5*time.Minute)
	if len(result) < 10 || len(result) > 14 {
		t.Errorf("downsample got %d observations, want ~12", len(result))
	}

	// First observation should be kept
	if result[0].AirTemperature != 0 {
		t.Error("first observation should be the original first")
	}
}

func TestDownsampleEmpty(t *testing.T) {
	result := downsample(nil, 5*time.Minute)
	if len(result) != 0 {
		t.Errorf("downsample(nil) = %d, want 0", len(result))
	}
}

func TestParseHistoryDates(t *testing.T) {
	tests := []struct {
		name      string
		flags     map[string]string
		wantStart string // "2006-01-02"
		wantEnd   string // "2006-01-02"
		wantErr   bool
	}{
		{
			name:      "single date",
			flags:     map[string]string{"date": "2024-01-15"},
			wantStart: "2024-01-15",
			wantEnd:   "2024-01-16",
		},
		{
			name:      "date range",
			flags:     map[string]string{"from": "2024-01-01", "to": "2024-01-31"},
			wantStart: "2024-01-01",
			wantEnd:   "2024-02-01",
		},
		{
			name:    "only from",
			flags:   map[string]string{"from": "2024-01-01"},
			wantErr: true,
		},
		{
			name:    "only to",
			flags:   map[string]string{"to": "2024-01-31"},
			wantErr: true,
		},
		{
			name:    "invalid date format",
			flags:   map[string]string{"date": "01-15-2024"},
			wantErr: true,
		},
		{
			name:    "invalid from format",
			flags:   map[string]string{"from": "bad", "to": "2024-01-31"},
			wantErr: true,
		},
		{
			name:    "invalid to format",
			flags:   map[string]string{"from": "2024-01-01", "to": "bad"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("date", "", "")
			cmd.Flags().String("from", "", "")
			cmd.Flags().String("to", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			start, end, err := parseHistoryDates(cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHistoryDates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got := start.Format("2006-01-02"); got != tt.wantStart {
				t.Errorf("start = %s, want %s", got, tt.wantStart)
			}
			if got := end.Format("2006-01-02"); got != tt.wantEnd {
				t.Errorf("end = %s, want %s", got, tt.wantEnd)
			}
		})
	}

	// Default (no flags) should be last 24 hours
	t.Run("default last 24h", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("date", "", "")
		cmd.Flags().String("from", "", "")
		cmd.Flags().String("to", "", "")

		start, end, err := parseHistoryDates(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		span := end.Sub(start)
		if span < 23*time.Hour || span > 25*time.Hour {
			t.Errorf("default span = %v, want ~24h", span)
		}
	})
}
