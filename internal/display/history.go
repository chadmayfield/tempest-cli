package display

import (
	"fmt"
	"strings"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// RenderHistory renders a table of historical observations using bubbles/table.
func RenderHistory(theme *Theme, observations []tempest.Observation, imperial bool, termWidth int) string {
	var b strings.Builder

	b.WriteString(theme.Title.Render("History"))
	b.WriteString(fmt.Sprintf("  %s\n\n", theme.Muted.Render(fmt.Sprintf("%d observations", len(observations)))))

	if len(observations) == 0 {
		b.WriteString(theme.Muted.Render("No observations in this time range"))
		return b.String()
	}

	columns := []table.Column{
		{Title: "Time", Width: 16},
		{Title: "Temp", Width: 8},
		{Title: "Feels Like", Width: 10},
		{Title: "Hum%", Width: 6},
		{Title: "Wind", Width: 14},
		{Title: "Pressure", Width: 10},
		{Title: "Rain", Width: 8},
		{Title: "UV", Width: 5},
	}

	// Adjust column widths to fit terminal
	totalWidth := 0
	for _, c := range columns {
		totalWidth += c.Width + 1 // +1 for separator
	}
	if termWidth > 0 && totalWidth > termWidth {
		scale := float64(termWidth) / float64(totalWidth)
		for i := range columns {
			columns[i].Width = max(int(float64(columns[i].Width)*scale), 4)
		}
	}

	var rows []table.Row
	for _, obs := range observations {
		ts := obs.Timestamp.Format("01-02 15:04")
		temp := FormatTemp(obs.AirTemperature, imperial)
		feels := FormatTemp(obs.FeelsLike, imperial)
		hum := fmt.Sprintf("%.0f%%", obs.RelativeHumidity)
		compass := tempest.WindDirectionToCompass(obs.WindDirection)
		wind := fmt.Sprintf("%s %s", FormatWind(obs.WindAvg, imperial), compass)
		pressure := FormatPressure(obs.StationPressure, imperial)
		rain := FormatPrecip(obs.RainAccumulation, imperial)
		uv := fmt.Sprintf("%.1f", obs.UVIndex)

		rows = append(rows, table.Row{ts, temp, feels, hum, wind, pressure, rain, uv})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	if theme.NoColor {
		s.Header = lipgloss.NewStyle().Bold(true)
		s.Cell = lipgloss.NewStyle()
		s.Selected = lipgloss.NewStyle()
	} else {
		s.Header = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#cccccc", Dark: "#444444"})
		s.Selected = lipgloss.NewStyle()
	}
	t.SetStyles(s)

	b.WriteString(t.View())

	return b.String()
}
