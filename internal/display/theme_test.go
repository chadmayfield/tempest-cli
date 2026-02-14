package display

import (
	"strings"
	"testing"
)

func TestNewTheme(t *testing.T) {
	theme := NewTheme(false)
	if theme.NoColor {
		t.Error("expected NoColor=false")
	}

	noColorTheme := NewTheme(true)
	if !noColorTheme.NoColor {
		t.Error("expected NoColor=true")
	}
}

func TestTempColorNoColor(t *testing.T) {
	theme := NewTheme(true)
	got := theme.TempColor(25.0, "25.0°C")
	if got != "25.0°C" {
		t.Errorf("TempColor with noColor = %q, want %q", got, "25.0°C")
	}
}

func TestUVLabel(t *testing.T) {
	tests := []struct {
		uv   float64
		want string
	}{
		{0, "Low"},
		{2, "Low"},
		{3, "Moderate"},
		{5, "Moderate"},
		{6, "High"},
		{7, "High"},
		{8, "Very High"},
		{10, "Very High"},
		{11, "Extreme"},
	}
	for _, tt := range tests {
		got := UVLabel(tt.uv)
		if got != tt.want {
			t.Errorf("UVLabel(%.0f) = %q, want %q", tt.uv, got, tt.want)
		}
	}
}

func TestBatteryLabel(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{2.5, "Good"},
		{2.4, "Good"},
		{2.3, "Fair"},
		{2.1, "Fair"},
		{2.0, "Low"},
	}
	for _, tt := range tests {
		got := BatteryLabel(tt.v)
		if got != tt.want {
			t.Errorf("BatteryLabel(%.1f) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestConditionIcon(t *testing.T) {
	if got := ConditionIcon("clear-day"); got != "☀" {
		t.Errorf("ConditionIcon(clear-day) = %q", got)
	}
	if got := ConditionIcon("rainy"); got != "☂" {
		t.Errorf("ConditionIcon(rainy) = %q", got)
	}
	if got := ConditionIcon("unknown"); got != "•" {
		t.Errorf("ConditionIcon(unknown) = %q, want •", got)
	}
}

func TestWindArrow(t *testing.T) {
	tests := []struct {
		deg  float64
		want string
	}{
		{0, "↓"},
		{90, "←"},
		{180, "↑"},
		{270, "→"},
	}
	for _, tt := range tests {
		got := WindArrow(tt.deg)
		if got != tt.want {
			t.Errorf("WindArrow(%.0f) = %q, want %q", tt.deg, got, tt.want)
		}
	}
}

func TestFormatTemp(t *testing.T) {
	got := FormatTemp(0, false)
	if got != "0.0°C" {
		t.Errorf("FormatTemp(0, metric) = %q", got)
	}
	got = FormatTemp(0, true)
	if got != "32.0°F" {
		t.Errorf("FormatTemp(0, imperial) = %q", got)
	}
}

func TestFormatWind(t *testing.T) {
	got := FormatWind(10.0, false)
	if got != "10.0 m/s" {
		t.Errorf("FormatWind(10, metric) = %q", got)
	}
	got = FormatWind(10.0, true)
	if !strings.Contains(got, "mph") {
		t.Errorf("FormatWind(10, imperial) = %q, want mph", got)
	}
}

func TestFormatPressure(t *testing.T) {
	got := FormatPressure(1013.25, false)
	if !strings.Contains(got, "hPa") {
		t.Errorf("FormatPressure metric = %q", got)
	}
	got = FormatPressure(1013.25, true)
	if !strings.Contains(got, "inHg") {
		t.Errorf("FormatPressure imperial = %q", got)
	}
}

func TestFormatPrecip(t *testing.T) {
	got := FormatPrecip(25.4, false)
	if !strings.Contains(got, "mm") {
		t.Errorf("FormatPrecip metric = %q", got)
	}
	got = FormatPrecip(25.4, true)
	if !strings.Contains(got, "in") {
		t.Errorf("FormatPrecip imperial = %q", got)
	}
}
