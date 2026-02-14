package display

import (
	"fmt"
	"strings"
	"time"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/charmbracelet/lipgloss"
)

// RenderCurrent renders a styled current conditions display.
func RenderCurrent(theme *Theme, obs *tempest.StationObservation, stationName string, imperial bool, termWidth int) string {
	var b strings.Builder

	// Header
	header := theme.Title.Render(stationName)
	ago := timeAgo(obs.Timestamp)
	updated := theme.Subtitle.Render(fmt.Sprintf("Updated %s ago", ago))

	b.WriteString(header + "  " + updated + "\n\n")

	// Temperature block
	tempStr := FormatTemp(obs.AirTemperature, imperial)
	feelsStr := FormatTemp(obs.FeelsLike, imperial)

	bigTemp := theme.TempColor(obs.AirTemperature, tempStr)
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(bigTemp))
	b.WriteString("  " + theme.Muted.Render("Feels like ") + theme.TempColor(obs.FeelsLike, feelsStr))
	b.WriteString("\n\n")

	// Key-value grid — all values color-coded
	humStr := fmt.Sprintf("%.0f%%", obs.RelativeHumidity)
	windStr := formatWindFull(obs.WindAvg, obs.WindDirection, imperial)
	gustStr := FormatWind(obs.WindGust, imperial)
	pressureStr := FormatPressure(obs.SeaLevelPressure, imperial)
	if obs.PressureTrend != "" {
		pressureStr += " (" + obs.PressureTrend + ")"
	}
	rainStr := FormatPrecip(obs.PrecipAccumDay, imperial)
	lightningStr := formatLightning(obs.LightningCount3hr, obs.LightningStrikeLastDistance, imperial)

	rows := []struct {
		label string
		value string
	}{
		{"Humidity", theme.HumidityColor(obs.RelativeHumidity, humStr)},
		{"Dew Point", theme.TempColor(obs.DewPoint, FormatTemp(obs.DewPoint, imperial))},
		{"Wind", theme.WindColor(obs.WindAvg, windStr)},
		{"Wind Gust", theme.WindColor(obs.WindGust, gustStr)},
		{"Wind Lull", theme.WindColor(obs.WindLull, FormatWind(obs.WindLull, imperial))},
		{"Pressure", theme.PressureColor(obs.SeaLevelPressure, pressureStr)},
		{"UV Index", formatUV(theme, obs.UV)},
		{"Solar Radiation", theme.Value.Render(fmt.Sprintf("%.0f W/m²", obs.SolarRadiation))},
		{"Rain Today", theme.RainColor(obs.PrecipAccumDay, rainStr)},
		{"Lightning (3hr)", theme.LightningColor(obs.LightningCount3hr, lightningStr)},
		// Battery: not available in StationObservation (REST endpoint doesn't return it).
		{"Battery", theme.Muted.Render("N/A")},
	}

	// Two-column layout
	mid := (len(rows) + 1) / 2
	leftCol := rows[:mid]
	rightCol := rows[mid:]

	// Calculate per-column label widths for tight alignment.
	leftLabelW := 0
	for _, r := range leftCol {
		if len(r.label) > leftLabelW {
			leftLabelW = len(r.label)
		}
	}
	leftLabelW += 2 // padding after label

	rightLabelW := 0
	for _, r := range rightCol {
		if len(r.label) > rightLabelW {
			rightLabelW = len(r.label)
		}
	}
	rightLabelW += 2 // padding after label

	// Fixed-width cell for the left column so the right column aligns.
	leftCell := lipgloss.NewStyle().Width(leftLabelW + 24)

	for i := range leftCol {
		leftLabel := theme.Label.Width(leftLabelW).Render(leftCol[i].label)
		left := leftCell.Render(leftLabel + leftCol[i].value)
		if i < len(rightCol) {
			rightLabel := theme.Label.Width(rightLabelW).Render(rightCol[i].label)
			right := rightLabel + rightCol[i].value
			b.WriteString(left + right + "\n")
		} else {
			b.WriteString(left + "\n")
		}
	}

	// Wrap in border, respect terminal width
	content := b.String()
	if !theme.NoColor {
		border := theme.Border
		if termWidth > 0 {
			border = border.MaxWidth(termWidth - 2)
		}
		content = border.Render(content)
	}

	return content
}

func formatWindFull(mps, degrees float64, imperial bool) string {
	compass := tempest.WindDirectionToCompass(degrees)
	arrow := WindArrow(degrees)
	speed := FormatWind(mps, imperial)
	return fmt.Sprintf("%s %s %s", speed, compass, arrow)
}

func formatUV(theme *Theme, uv float64) string {
	label := UVLabel(uv)
	formatted := fmt.Sprintf("%.1f (%s)", uv, label)
	return theme.UVColor(uv, formatted)
}

func formatLightning(count int, distKm float64, imperial bool) string {
	if count == 0 {
		return "none"
	}
	s := fmt.Sprintf("%d strikes", count)
	if distKm > 0 {
		s += " " + FormatDistance(distKm, imperial) + " avg"
	}
	return s
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int(d.Minutes()) % 60
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
