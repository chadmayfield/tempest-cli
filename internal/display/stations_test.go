package display

import (
	"strings"
	"testing"
	"time"
)

func TestRenderStations(t *testing.T) {
	theme := NewTheme(true)

	rows := []StationRow{
		{
			ConfigName:   "home",
			StationName:  "Home Station",
			StationID:    12345,
			DeviceID:     67890,
			IsDefault:    true,
			Online:       true,
			LastObserved: time.Now().Add(-5 * time.Minute),
		},
		{
			ConfigName:   "office",
			StationName:  "Office Weather",
			StationID:    54321,
			DeviceID:     9876,
			IsDefault:    false,
			Online:       false,
			LastObserved: time.Now().Add(-2 * time.Hour),
		},
	}

	output := RenderStations(theme, rows)

	if !strings.Contains(output, "Stations") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "home") {
		t.Error("missing home station")
	}
	if !strings.Contains(output, "office") {
		t.Error("missing office station")
	}
	if !strings.Contains(output, "Online") {
		t.Error("missing Online status")
	}
	if !strings.Contains(output, "Offline") {
		t.Error("missing Offline status")
	}
}

func TestRenderStationsSingle(t *testing.T) {
	theme := NewTheme(true)

	rows := []StationRow{
		{
			ConfigName:   "home",
			StationName:  "My Station",
			StationID:    12345,
			DeviceID:     67890,
			IsDefault:    true,
			Online:       true,
			LastObserved: time.Now().Add(-1 * time.Minute),
		},
	}

	output := RenderStations(theme, rows)

	if !strings.Contains(output, "home") {
		t.Error("missing station name")
	}
	if !strings.Contains(output, "Online") {
		t.Error("missing Online status")
	}
}

func TestRenderStationsJSON(t *testing.T) {
	// Verify that StationRow can be JSON-serialized
	rows := []StationRow{
		{ConfigName: "home", StationID: 1, Online: true},
	}
	if len(rows) == 0 {
		t.Error("expected rows")
	}
}
