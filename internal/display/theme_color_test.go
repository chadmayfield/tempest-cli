package display

import "testing"

func TestWithNoEmoji(t *testing.T) {
	theme := NewTheme(true, WithNoEmoji(true))
	if !theme.NoEmoji {
		t.Error("expected NoEmoji=true")
	}

	theme2 := NewTheme(true, WithNoEmoji(false))
	if theme2.NoEmoji {
		t.Error("expected NoEmoji=false")
	}
}

func TestWithNoEmojiColoredTheme(t *testing.T) {
	theme := NewTheme(false, WithNoEmoji(true))
	if !theme.NoEmoji {
		t.Error("expected NoEmoji=true on colored theme")
	}
}

func TestConditionLabel(t *testing.T) {
	tests := []struct {
		icon string
		want string
	}{
		{"clear-day", "[clear]"},
		{"clear-night", "[clear]"},
		{"cloudy", "[cloudy]"},
		{"foggy", "[fog]"},
		{"partly-cloudy-day", "[partly cloudy]"},
		{"partly-cloudy-night", "[partly cloudy]"},
		{"possibly-rainy-day", "[chance rain]"},
		{"possibly-rainy-night", "[chance rain]"},
		{"rainy", "[rain]"},
		{"sleet", "[sleet]"},
		{"snow", "[snow]"},
		{"thunderstorm", "[storm]"},
		{"windy", "[windy]"},
		{"unknown-icon", "[--]"},
	}
	for _, tt := range tests {
		t.Run(tt.icon, func(t *testing.T) {
			got := ConditionLabel(tt.icon)
			if got != tt.want {
				t.Errorf("ConditionLabel(%q) = %q, want %q", tt.icon, got, tt.want)
			}
		})
	}
}

func TestFormatDistance(t *testing.T) {
	got := FormatDistance(10.0, false)
	if got != "10.0 km" {
		t.Errorf("FormatDistance(10, metric) = %q, want %q", got, "10.0 km")
	}
	got = FormatDistance(10.0, true)
	if got != "6.2 mi" {
		t.Errorf("FormatDistance(10, imperial) = %q, want %q", got, "6.2 mi")
	}
}

func TestTempColorBranches(t *testing.T) {
	theme := NewTheme(false) // colored theme to exercise color branches

	tests := []struct {
		name  string
		tempC float64
	}{
		{"freezing", -5.0},
		{"cold", 0.0},
		{"cool", 10.0},
		{"warm", 25.0},
		{"hot", 35.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.TempColor(tt.tempC, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}
}

func TestUVColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name string
		uv   float64
	}{
		{"low", 1.0},
		{"moderate", 4.0},
		{"high", 6.5},
		{"very_high", 9.0},
		{"extreme", 12.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.UVColor(tt.uv, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}
}

func TestBatteryColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name  string
		volts float64
	}{
		{"good", 2.6},
		{"fair", 2.2},
		{"low", 1.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.BatteryColor(tt.volts, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	// Also test noColor passthrough
	noColorTheme := NewTheme(true)
	got := noColorTheme.BatteryColor(2.5, "2.5V")
	if got != "2.5V" {
		t.Errorf("BatteryColor noColor = %q, want %q", got, "2.5V")
	}
}

func TestHumidityColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name string
		hum  float64
	}{
		{"dry", 20.0},
		{"comfortable", 45.0},
		{"humid", 75.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.HumidityColor(tt.hum, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	noColorTheme := NewTheme(true)
	got := noColorTheme.HumidityColor(50, "50%")
	if got != "50%" {
		t.Errorf("HumidityColor noColor = %q, want %q", got, "50%")
	}
}

func TestWindColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name string
		mps  float64
	}{
		{"calm", 2.0},
		{"moderate", 7.0},
		{"strong", 15.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.WindColor(tt.mps, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	noColorTheme := NewTheme(true)
	got := noColorTheme.WindColor(5, "5 m/s")
	if got != "5 m/s" {
		t.Errorf("WindColor noColor = %q, want %q", got, "5 m/s")
	}
}

func TestPressureColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name string
		hpa  float64
	}{
		{"low", 990.0},
		{"normal", 1010.0},
		{"high", 1030.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.PressureColor(tt.hpa, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	noColorTheme := NewTheme(true)
	got := noColorTheme.PressureColor(1013, "1013 hPa")
	if got != "1013 hPa" {
		t.Errorf("PressureColor noColor = %q, want %q", got, "1013 hPa")
	}
}

func TestRainColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name string
		mm   float64
	}{
		{"none", 0.0},
		{"light", 2.5},
		{"heavy", 10.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.RainColor(tt.mm, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	noColorTheme := NewTheme(true)
	got := noColorTheme.RainColor(0, "0 mm")
	if got != "0 mm" {
		t.Errorf("RainColor noColor = %q, want %q", got, "0 mm")
	}
}

func TestLightningColorBranches(t *testing.T) {
	theme := NewTheme(false)

	tests := []struct {
		name  string
		count int
	}{
		{"none", 0},
		{"some", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := theme.LightningColor(tt.count, "test")
			if result == "" {
				t.Error("expected non-empty styled string")
			}
		})
	}

	noColorTheme := NewTheme(true)
	got := noColorTheme.LightningColor(3, "3 strikes")
	if got != "3 strikes" {
		t.Errorf("LightningColor noColor = %q, want %q", got, "3 strikes")
	}
}

func TestUVColorNoColor(t *testing.T) {
	theme := NewTheme(true)
	got := theme.UVColor(8.0, "8.0 (Very High)")
	if got != "8.0 (Very High)" {
		t.Errorf("UVColor noColor = %q, want passthrough", got)
	}
}
