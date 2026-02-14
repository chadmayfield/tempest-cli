package display

import (
	"fmt"
	"strings"

	tempest "github.com/chadmayfield/tempest-go"
	"github.com/charmbracelet/lipgloss"
)

// RenderForecast renders forecast day cards.
func RenderForecast(theme *Theme, forecast *tempest.Forecast, days int, imperial bool, termWidth int) string {
	var b strings.Builder

	b.WriteString(theme.Title.Render("Forecast"))
	b.WriteString("\n\n")

	daily := forecast.Daily
	if len(daily) > days {
		daily = daily[:days]
	}

	if len(daily) == 0 {
		b.WriteString(theme.Muted.Render("No forecast data available"))
		return b.String()
	}

	// Build card contents (without borders) and find the tallest.
	contents := make([]string, len(daily))
	maxHeight := 0
	for i, day := range daily {
		contents[i] = forecastCardContent(theme, day, imperial)
		h := lipgloss.Height(contents[i])
		if h > maxHeight {
			maxHeight = h
		}
	}

	// Render all cards with uniform height via the card style.
	cardStyle := theme.Card.Width(20).Height(maxHeight)
	cards := make([]string, len(contents))
	for i, c := range contents {
		cards[i] = cardStyle.Render(c)
	}

	// Determine layout: side-by-side if enough width, else vertical
	cardWidth := lipgloss.Width(cards[0])
	cardsPerRow := termWidth / (cardWidth + 1)
	if cardsPerRow < 1 {
		cardsPerRow = 1
	}

	for i := 0; i < len(cards); i += cardsPerRow {
		end := i + cardsPerRow
		if end > len(cards) {
			end = len(cards)
		}
		row := cards[i:end]
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
		b.WriteString("\n")
	}

	return b.String()
}

func forecastCardContent(theme *Theme, day tempest.DailyForecast, imperial bool) string {
	var b strings.Builder

	// Date
	dateStr := day.Date.Format("Mon Jan 2")
	b.WriteString(theme.Title.Render(dateStr))
	b.WriteString("\n")

	// Icon + conditions
	var icon string
	if theme.NoEmoji {
		icon = ConditionLabel(day.Icon)
	} else {
		icon = ConditionIcon(day.Icon)
	}
	b.WriteString(icon + " " + day.Conditions)
	b.WriteString("\n")

	// High / Low
	high := FormatTemp(day.HighTemp, imperial)
	low := FormatTemp(day.LowTemp, imperial)
	b.WriteString(theme.TempColor(day.HighTemp, high) + " / " + theme.TempColor(day.LowTemp, low))
	b.WriteString("\n")

	// Precip chance
	if day.PrecipChance > 0 {
		b.WriteString(fmt.Sprintf("Precip: %d%%", day.PrecipChance))
		b.WriteString("\n")
	}

	// Sunrise / Sunset
	if !day.Sunrise.IsZero() {
		b.WriteString(fmt.Sprintf("↑%s ↓%s",
			day.Sunrise.Format("3:04pm"),
			day.Sunset.Format("3:04pm")))
		b.WriteString("\n")
	}

	return b.String()
}
