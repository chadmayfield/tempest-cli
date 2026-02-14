package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/chadmayfield/tempest-cli/internal/display"
	jsonout "github.com/chadmayfield/tempest-cli/internal/json"
	tempest "github.com/chadmayfield/tempest-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var stationsCmd = &cobra.Command{
	Use:   "stations",
	Short: "List configured stations with status",
	RunE:  runStations,
}

func init() {
	rootCmd.AddCommand(stationsCmd)
}

func runStations(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return wrapConfigError(err)
	}

	if len(cfg.Stations) == 0 {
		return fmt.Errorf("no stations configured; run 'tempest config init' to set up")
	}

	serverURL := resolveServerURL(cfg)
	var rows []display.StationRow

	// When using tempestd, try the list endpoint first
	var serverStations map[int]*tempest.Station
	if serverURL != "" {
		serverStations = fetchStationListFromServer(ctx, serverURL)
	}

	for name, sc := range cfg.Stations {
		row := display.StationRow{
			ConfigName: name,
			StationID:  sc.StationID,
			DeviceID:   sc.DeviceID,
			IsDefault:  name == cfg.DefaultStation,
		}

		var station *tempest.Station
		var obs *tempest.StationObservation

		if serverURL != "" {
			// Use cached list data if available, otherwise fall back to per-station query
			if s, ok := serverStations[sc.StationID]; ok {
				station = s
			} else {
				station, _, err = fetchStationStatusFromServer(ctx, serverURL, sc.StationID)
			}
			if station != nil {
				obs, _ = fetchCurrentObsFromServer(ctx, serverURL, sc.StationID)
			}
		} else {
			station, obs, err = fetchStationStatus(ctx, &sc)
		}

		if err != nil || station == nil {
			row.StationName = sc.Name
			row.Online = false
		} else {
			row.StationName = station.Name
			if obs != nil {
				row.LastObserved = obs.Timestamp
				row.Online = time.Since(obs.Timestamp) < 30*time.Minute
			}
		}

		rows = append(rows, row)
	}

	if viper.GetBool("json") {
		return jsonout.Write(cmd.OutOrStdout(), stationsJSON(rows))
	}

	noColor := viper.GetBool("no-color")
	theme := display.NewTheme(noColor, display.WithNoEmoji(viper.GetBool("no-emoji")))

	output := display.RenderStations(theme, rows)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)

	return nil
}

type stationJSONOutput struct {
	Name            string  `json:"name"`
	StationID       int     `json:"station_id"`
	DeviceID        int     `json:"device_id"`
	Status          string  `json:"status"`
	LastObservation *string `json:"last_observation"`
}

func stationsJSON(rows []display.StationRow) []stationJSONOutput {
	out := make([]stationJSONOutput, len(rows))
	for i, r := range rows {
		status := "online"
		if !r.Online {
			status = "offline"
		}
		var lastObs *string
		if !r.LastObserved.IsZero() {
			s := r.LastObserved.Format(time.RFC3339)
			lastObs = &s
		}
		out[i] = stationJSONOutput{
			Name:            r.StationName,
			StationID:       r.StationID,
			DeviceID:        r.DeviceID,
			Status:          status,
			LastObservation: lastObs,
		}
	}
	return out
}

func fetchStationStatus(ctx context.Context, sc *config.StationConfig) (*tempest.Station, *tempest.StationObservation, error) {
	client, err := tempest.NewClient(sc.Token)
	if err != nil {
		return nil, nil, err
	}

	station, err := client.GetStation(ctx, sc.StationID)
	if err != nil {
		return nil, nil, err
	}

	obs, err := client.GetStationObservation(ctx, sc.StationID)
	if err != nil {
		return station, nil, nil // station found but no observation
	}

	return station, obs, nil
}

// fetchStationListFromServer tries the /api/v1/stations list endpoint and returns
// a map of station ID → Station. Returns an empty map if the endpoint is unavailable.
func fetchStationListFromServer(ctx context.Context, serverURL string) map[int]*tempest.Station {
	result, err := fetchFromTempestd[[]tempest.Station](ctx, serverURL, "/api/v1/stations")
	if err != nil {
		slog.Debug("tempestd stations list endpoint unavailable, falling back to per-station queries", "error", err)
		return nil
	}
	m := make(map[int]*tempest.Station, len(*result))
	for i := range *result {
		s := &(*result)[i]
		m[s.StationID] = s
	}
	return m
}

// fetchCurrentObsFromServer fetches the current observation for a single station from tempestd.
func fetchCurrentObsFromServer(ctx context.Context, serverURL string, stationID int) (*tempest.StationObservation, error) {
	return fetchFromTempestd[tempest.StationObservation](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d/current", stationID))
}

func fetchStationStatusFromServer(ctx context.Context, serverURL string, stationID int) (*tempest.Station, *tempest.StationObservation, error) {
	station, err := fetchFromTempestd[tempest.Station](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d", stationID))
	if err != nil {
		return nil, nil, err
	}

	obs, err := fetchFromTempestd[tempest.StationObservation](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d/current", stationID))
	if err != nil {
		return station, nil, nil
	}

	return station, obs, nil
}
