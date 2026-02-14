package cmd

import (
	"testing"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/spf13/cobra"
)

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
