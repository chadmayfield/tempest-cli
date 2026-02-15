package cmd

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/chadmayfield/tempest-cli/internal/display"
	jsonout "github.com/chadmayfield/tempest-cli/internal/json"
	tempest "github.com/chadmayfield/tempest-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show historical weather data",
	Long:  "Display historical weather observations as a table.",
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().String("date", "", "single day (YYYY-MM-DD)")
	historyCmd.Flags().String("from", "", "range start (YYYY-MM-DD)")
	historyCmd.Flags().String("to", "", "range end (YYYY-MM-DD)")
	historyCmd.Flags().String("resolution", "", "data resolution: 1m, 5m, 30m, 3h (auto if omitted)")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return wrapConfigError(err)
	}

	stationName := viper.GetString("station")
	sc, err := cfg.ResolveStation(stationName)
	if err != nil {
		return wrapConfigError(err)
	}

	start, end, err := parseHistoryDates(cmd)
	if err != nil {
		return err
	}

	serverURL := resolveServerURL(cfg)
	imperial := cfg.IsImperial()
	resFlag, _ := cmd.Flags().GetString("resolution")
	resolution := resolveResolution(resFlag, end.Sub(start))

	units := "metric"
	if imperial {
		units = "imperial"
	}

	var observations []tempest.Observation
	if serverURL != "" {
		resLabel := resFlag
		if resLabel == "" {
			resLabel = resolutionLabel(resolution)
		}
		observations, err = fetchHistoryFromServer(ctx, serverURL, sc.StationID, start, end, units, resLabel)
	} else {
		observations, err = fetchHistoryFromAPI(ctx, sc, start, end)
	}
	if err != nil {
		return wrapAPIError(err)
	}

	// Apply unit conversion (only for cloud API; tempestd handles units server-side)
	if serverURL == "" && imperial {
		for i := range observations {
			converted := tempest.ConvertObservation(&observations[i], tempest.Imperial)
			observations[i] = *converted
		}
	}

	// Apply client-side downsampling
	if resolution > 0 {
		observations = downsample(observations, resolution)
	}

	if viper.GetBool("json") {
		jsonResLabel := resFlag
		if jsonResLabel == "" {
			jsonResLabel = resolutionLabel(resolution)
		}
		return jsonout.Write(cmd.OutOrStdout(), historyJSON(observations, sc, units, start, end, jsonResLabel))
	}

	noColor := viper.GetBool("no-color")
	theme := display.NewTheme(noColor, display.WithNoEmoji(viper.GetBool("no-emoji")))

	termWidth := 80
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		termWidth = w
	}

	// When imperial, data is already converted, so pass imperial=false to avoid double conversion in display
	output := display.RenderHistory(theme, observations, false, termWidth)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)

	return nil
}

type historyJSONOutput struct {
	Station      stationMeta          `json:"station"`
	Units        string               `json:"units"`
	From         time.Time            `json:"from"`
	To           time.Time            `json:"to"`
	Resolution   string               `json:"resolution"`
	Observations []historyObsJSON     `json:"observations"`
}

type historyObsJSON struct {
	Timestamp             time.Time `json:"timestamp"`
	Temperature           float64   `json:"temperature"`
	FeelsLike             float64   `json:"feels_like"`
	Humidity              float64   `json:"humidity"`
	WindSpeed             float64   `json:"wind_speed"`
	WindDirection         float64   `json:"wind_direction"`
	WindDirectionCardinal string    `json:"wind_direction_cardinal"`
	Pressure              float64   `json:"pressure"`
	Rain                  float64   `json:"rain"`
	UVIndex               float64   `json:"uv_index"`
}

func historyJSON(obs []tempest.Observation, sc *config.StationConfig, units string, start, end time.Time, resolution string) historyJSONOutput {
	items := make([]historyObsJSON, len(obs))
	for i, o := range obs {
		items[i] = historyObsJSON{
			Timestamp:             o.Timestamp,
			Temperature:           o.AirTemperature,
			FeelsLike:             o.FeelsLike,
			Humidity:              o.RelativeHumidity,
			WindSpeed:             o.WindAvg,
			WindDirection:         o.WindDirection,
			WindDirectionCardinal: tempest.WindDirectionToCompass(o.WindDirection),
			Pressure:              o.StationPressure,
			Rain:                  o.RainAccumulation,
			UVIndex:               o.UVIndex,
		}
	}
	return historyJSONOutput{
		Station: stationMeta{
			Name:      sc.Name,
			StationID: sc.StationID,
			DeviceID:  sc.DeviceID,
		},
		Units:        units,
		From:         start,
		To:           end,
		Resolution:   resolution,
		Observations: items,
	}
}

func resolutionLabel(d time.Duration) string {
	switch d {
	case time.Minute:
		return "1m"
	case 5 * time.Minute:
		return "5m"
	case 30 * time.Minute:
		return "30m"
	case 3 * time.Hour:
		return "3h"
	default:
		return d.String()
	}
}

func fetchHistoryFromAPI(ctx context.Context, sc *config.StationConfig, start, end time.Time) ([]tempest.Observation, error) {
	client, err := tempest.NewClient(sc.Token)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	if sc.DeviceID <= 0 {
		return nil, fmt.Errorf("device_id is required for historical data; add it to your station config or re-run 'tempest config init'")
	}

	observations, err := client.GetDeviceObservations(ctx, sc.DeviceID, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetching observations: %w", err)
	}

	return observations, nil
}

func fetchHistoryFromServer(ctx context.Context, serverURL string, stationID int, start, end time.Time, units, resolution string) ([]tempest.Observation, error) {
	params := url.Values{}
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("units", units)
	params.Set("resolution", resolution)
	path := fmt.Sprintf("/api/v1/stations/%d/observations?%s", stationID, params.Encode())
	result, err := fetchFromTempestd[[]tempest.Observation](ctx, serverURL, path)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

func parseHistoryDates(cmd *cobra.Command) (time.Time, time.Time, error) {
	dateStr, _ := cmd.Flags().GetString("date")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")

	if dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --date format (use YYYY-MM-DD): %w", err)
		}
		return d, d.Add(24 * time.Hour), nil
	}

	if fromStr != "" && toStr != "" {
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --from format (use YYYY-MM-DD): %w", err)
		}
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --to format (use YYYY-MM-DD): %w", err)
		}
		// Include the end date fully
		return from, to.Add(24 * time.Hour), nil
	}

	if fromStr != "" || toStr != "" {
		return time.Time{}, time.Time{}, fmt.Errorf("both --from and --to are required for a date range")
	}

	// Default: last 24 hours
	now := time.Now()
	return now.Add(-24 * time.Hour), now, nil
}

// resolveResolution returns the downsampling interval.
func resolveResolution(flag string, span time.Duration) time.Duration {
	if flag != "" {
		switch flag {
		case "1m":
			return time.Minute
		case "5m":
			return 5 * time.Minute
		case "30m":
			return 30 * time.Minute
		case "3h":
			return 3 * time.Hour
		}
	}

	// Auto-select based on time range
	switch {
	case span <= 24*time.Hour:
		return time.Minute
	case span <= 7*24*time.Hour:
		return 5 * time.Minute
	case span <= 30*24*time.Hour:
		return 30 * time.Minute
	default:
		return 3 * time.Hour
	}
}

// downsample picks the nearest observation to each interval boundary.
func downsample(obs []tempest.Observation, interval time.Duration) []tempest.Observation {
	if len(obs) == 0 {
		return obs
	}

	var result []tempest.Observation
	start := obs[0].Timestamp
	nextBoundary := start

	for _, o := range obs {
		if o.Timestamp.Before(nextBoundary) {
			continue
		}
		result = append(result, o)
		nextBoundary = o.Timestamp.Add(interval)
	}

	return result
}
