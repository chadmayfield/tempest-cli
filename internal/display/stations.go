package display

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// StationRow holds display data for a single station.
type StationRow struct {
	ConfigName   string
	StationName  string
	StationID    int
	DeviceID     int
	IsDefault    bool
	Online       bool
	LastObserved time.Time
}

// RenderStations renders a table of stations using bubbles/table.
func RenderStations(theme *Theme, rows []StationRow) string {
	var b strings.Builder

	b.WriteString(theme.Title.Render("Stations"))
	b.WriteString("\n\n")

	columns := []table.Column{
		{Title: "", Width: 1},
		{Title: "Name", Width: 12},
		{Title: "Station", Width: 20},
		{Title: "SID", Width: 8},
		{Title: "DID", Width: 8},
		{Title: "Status", Width: 8},
		{Title: "Last Seen", Width: 14},
	}

	var tableRows []table.Row
	for _, r := range rows {
		def := " "
		if r.IsDefault {
			def = "*"
		}

		status := "Online"
		if !r.Online {
			status = "Offline"
		}

		lastSeen := "never"
		if !r.LastObserved.IsZero() {
			lastSeen = timeAgo(r.LastObserved) + " ago"
		}

		runes := []rune(r.StationName)
		name := r.StationName
		if len(runes) > 20 {
			name = string(runes[:19]) + "…"
		}

		tableRows = append(tableRows, table.Row{
			def, r.ConfigName, name,
			fmt.Sprintf("%d", r.StationID), fmt.Sprintf("%d", r.DeviceID),
			status, lastSeen,
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithHeight(len(tableRows)+1),
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
