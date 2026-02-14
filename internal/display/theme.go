package display

import (
	"fmt"
	"math"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/charmbracelet/lipgloss"
)

// Theme holds all lipgloss styles used for terminal rendering.
type Theme struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
	Muted    lipgloss.Style
	Border   lipgloss.Style
	Card     lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	NoColor  bool
	NoEmoji  bool
}

// ThemeOption configures optional theme settings.
type ThemeOption func(*Theme)

// WithNoEmoji disables Unicode symbols in condition icons.
func WithNoEmoji(noEmoji bool) ThemeOption {
	return func(t *Theme) {
		t.NoEmoji = noEmoji
	}
}

// NewTheme creates a theme appropriate for the terminal.
func NewTheme(noColor bool, opts ...ThemeOption) *Theme {
	t := &Theme{}
	for _, opt := range opts {
		opt(t)
	}

	if noColor {
		t.Title = lipgloss.NewStyle().Bold(true)
		t.Subtitle = lipgloss.NewStyle()
		t.Label = lipgloss.NewStyle()
		t.Value = lipgloss.NewStyle()
		t.Muted = lipgloss.NewStyle()
		t.Border = lipgloss.NewStyle()
		t.Card = lipgloss.NewStyle()
		t.Error = lipgloss.NewStyle()
		t.Success = lipgloss.NewStyle()
		t.Warning = lipgloss.NewStyle()
		t.NoColor = true
		return t
	}

	t2 := &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a2e", Dark: "#e0e0ff"}),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#555555", Dark: "#aaaaaa"}),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a2e", Dark: "#ffffff"}),
		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#444444"}).
			Padding(1, 2),
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#dddddd", Dark: "#333333"}).
			Padding(0, 1),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"}),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"}),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffaa00"}),
		NoColor: false,
		NoEmoji: t.NoEmoji,
	}
	return t2
}

// TempColor returns a styled string for a temperature value.
// Thresholds: blue (<32F/0C), white (cool), yellow (warm), red (>90F/32C).
func (t *Theme) TempColor(tempC float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case tempC <= 0:
		color = lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#66aaff"}
	case tempC <= 15:
		color = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#dddddd"}
	case tempC <= 32:
		color = lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"}
	default:
		color = lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"}
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// UVColor returns a styled string for a UV index value.
func (t *Theme) UVColor(uv float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case uv <= 2:
		color = lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"}
	case uv <= 5:
		color = lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"}
	case uv <= 7:
		color = lipgloss.AdaptiveColor{Light: "#cc4400", Dark: "#ff8800"}
	case uv <= 10:
		color = lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"}
	default:
		color = lipgloss.AdaptiveColor{Light: "#8800cc", Dark: "#cc66ff"}
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// BatteryColor returns a styled string for a battery voltage.
func (t *Theme) BatteryColor(volts float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case volts >= 2.4:
		color = lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"}
	case volts >= 2.1:
		color = lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"}
	default:
		color = lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"}
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// HumidityColor returns a styled string for a humidity percentage.
func (t *Theme) HumidityColor(hum float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case hum < 30:
		color = lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"} // dry
	case hum <= 60:
		color = lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"} // comfortable
	default:
		color = lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#66aaff"} // humid
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// WindColor returns a styled string for a wind speed in m/s.
func (t *Theme) WindColor(mps float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case mps < 5:
		color = lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"} // calm
	case mps < 10:
		color = lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"} // moderate
	default:
		color = lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"} // strong
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// PressureColor returns a styled string for a pressure value in hPa.
func (t *Theme) PressureColor(hpa float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	var color lipgloss.AdaptiveColor
	switch {
	case hpa < 1000:
		color = lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"} // low
	case hpa <= 1020:
		color = lipgloss.AdaptiveColor{Light: "#00aa00", Dark: "#44ff44"} // normal
	default:
		color = lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#66aaff"} // high
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// RainColor returns a styled string for precipitation.
func (t *Theme) RainColor(mm float64, formatted string) string {
	if t.NoColor {
		return formatted
	}
	if mm == 0 {
		return t.Muted.Render(formatted)
	}
	var color lipgloss.AdaptiveColor
	switch {
	case mm < 5:
		color = lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#66aaff"} // light
	default:
		color = lipgloss.AdaptiveColor{Light: "#0000cc", Dark: "#4444ff"} // heavy
	}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// LightningColor returns a styled string for lightning count.
func (t *Theme) LightningColor(count int, formatted string) string {
	if t.NoColor {
		return formatted
	}
	if count == 0 {
		return t.Muted.Render(formatted)
	}
	color := lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"}
	return lipgloss.NewStyle().Foreground(color).Render(formatted)
}

// UVLabel returns a human-readable UV severity label.
func UVLabel(uv float64) string {
	switch {
	case uv <= 2:
		return "Low"
	case uv <= 5:
		return "Moderate"
	case uv <= 7:
		return "High"
	case uv <= 10:
		return "Very High"
	default:
		return "Extreme"
	}
}

// BatteryLabel returns a human-readable battery status.
func BatteryLabel(volts float64) string {
	switch {
	case volts >= 2.4:
		return "Good"
	case volts >= 2.1:
		return "Fair"
	default:
		return "Low"
	}
}

// ConditionIcon maps a forecast icon string to a Unicode text symbol.
// Uses simple Unicode characters instead of emoji for maximum terminal compatibility.
func ConditionIcon(icon string) string {
	icons := map[string]string{
		"clear-day":            "☀",
		"clear-night":          "☽",
		"cloudy":               "☁",
		"foggy":                "≡",
		"partly-cloudy-day":    "⛅",
		"partly-cloudy-night":  "☁",
		"possibly-rainy-day":   "☂",
		"possibly-rainy-night": "☂",
		"rainy":                "☂",
		"sleet":                "❄",
		"snow":                 "❄",
		"thunderstorm":         "⚡",
		"windy":                "~",
	}
	if s, ok := icons[icon]; ok {
		return s
	}
	return "•"
}

// ConditionLabel returns a text-only label for --no-emoji mode.
func ConditionLabel(icon string) string {
	labels := map[string]string{
		"clear-day":            "[clear]",
		"clear-night":          "[clear]",
		"cloudy":               "[cloudy]",
		"foggy":                "[fog]",
		"partly-cloudy-day":    "[partly cloudy]",
		"partly-cloudy-night":  "[partly cloudy]",
		"possibly-rainy-day":   "[chance rain]",
		"possibly-rainy-night": "[chance rain]",
		"rainy":                "[rain]",
		"sleet":                "[sleet]",
		"snow":                 "[snow]",
		"thunderstorm":         "[storm]",
		"windy":                "[windy]",
	}
	if s, ok := labels[icon]; ok {
		return s
	}
	return "[--]"
}

// WindArrow returns a Unicode arrow character for a wind direction in degrees.
func WindArrow(degrees float64) string {
	// Wind direction is where wind comes FROM, arrow shows where it goes TO
	arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
	idx := int(math.Round(degrees/45.0)) % 8
	return arrows[idx]
}

// FormatTemp formats a temperature with units.
func FormatTemp(tempC float64, imperial bool) string {
	if imperial {
		return fmt.Sprintf("%.1f°F", tempest.CelsiusToFahrenheit(tempC))
	}
	return fmt.Sprintf("%.1f°C", tempC)
}

// FormatWind formats a wind speed with units.
func FormatWind(mps float64, imperial bool) string {
	if imperial {
		return fmt.Sprintf("%.1f mph", tempest.MpsToMph(mps))
	}
	return fmt.Sprintf("%.1f m/s", mps)
}

// FormatPressure formats a pressure with units.
func FormatPressure(hpa float64, imperial bool) string {
	if imperial {
		return fmt.Sprintf("%.2f inHg", tempest.HpaToInhg(hpa))
	}
	return fmt.Sprintf("%.1f hPa", hpa)
}

// FormatPrecip formats a precipitation amount with units.
func FormatPrecip(mm float64, imperial bool) string {
	if imperial {
		return fmt.Sprintf("%.2f in", tempest.MmToInches(mm))
	}
	return fmt.Sprintf("%.1f mm", mm)
}

// FormatDistance formats a distance with units.
func FormatDistance(km float64, imperial bool) string {
	if imperial {
		return fmt.Sprintf("%.1f mi", tempest.KmToMiles(km))
	}
	return fmt.Sprintf("%.1f km", km)
}
