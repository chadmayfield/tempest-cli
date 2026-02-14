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
	"golang.org/x/term"
)

var forecastCmd = &cobra.Command{
	Use:   "forecast",
	Short: "Show weather forecast",
	Long:  "Display multi-day weather forecast from your Tempest station.",
	RunE:  runForecast,
}

func init() {
	forecastCmd.Flags().IntP("days", "d", 5, "number of forecast days (max 10)")
	rootCmd.AddCommand(forecastCmd)
}

func runForecast(cmd *cobra.Command, args []string) error {
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

	days, _ := cmd.Flags().GetInt("days")
	if days < 1 {
		days = 1
	}
	if days > 10 {
		days = 10
	}

	serverURL := resolveServerURL(cfg)

	var forecast *tempest.Forecast
	if serverURL != "" {
		// tempestd may not cache forecasts — fall back to cloud API on failure
		forecast, err = fetchForecastFromServer(ctx, serverURL, sc.StationID)
		if err != nil {
			slog.Debug("tempestd forecast failed, falling back to cloud API", "error", err)
			forecast, err = fetchForecastFromAPI(ctx, sc)
		}
	} else {
		forecast, err = fetchForecastFromAPI(ctx, sc)
	}
	if err != nil {
		return wrapAPIError(err)
	}

	if viper.GetBool("json") {
		imperial := cfg.IsImperial()
		units := "metric"
		if imperial {
			units = "imperial"
		}
		return jsonout.Write(cmd.OutOrStdout(), forecastJSON(forecast, sc, units, imperial, days))
	}

	imperial := cfg.IsImperial()
	noColor := viper.GetBool("no-color")
	theme := display.NewTheme(noColor, display.WithNoEmoji(viper.GetBool("no-emoji")))

	termWidth := 80
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		termWidth = w
	}

	output := display.RenderForecast(theme, forecast, days, imperial, termWidth)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)

	return nil
}

type forecastJSONOutput struct {
	Station stationMeta       `json:"station"`
	Units   string            `json:"units"`
	Days    []forecastDayJSON `json:"daily"`
}

type forecastDayJSON struct {
	Date         string  `json:"date"`
	HighTemp     float64 `json:"high_temp"`
	LowTemp      float64 `json:"low_temp"`
	Conditions   string  `json:"conditions"`
	Icon         string  `json:"icon"`
	PrecipChance int     `json:"precip_chance"`
	PrecipType   string  `json:"precip_type,omitempty"`
	Sunrise      string  `json:"sunrise,omitempty"`
	Sunset       string  `json:"sunset,omitempty"`
}

func forecastJSON(f *tempest.Forecast, sc *config.StationConfig, units string, imperial bool, days int) forecastJSONOutput {
	n := len(f.Daily)
	if days < n {
		n = days
	}
	fdays := make([]forecastDayJSON, n)
	for i := 0; i < n; i++ {
		d := f.Daily[i]
		high := d.HighTemp
		low := d.LowTemp
		if imperial {
			high = tempest.CelsiusToFahrenheit(high)
			low = tempest.CelsiusToFahrenheit(low)
		}
		fdays[i] = forecastDayJSON{
			Date:         d.Date.Format(time.DateOnly),
			HighTemp:     high,
			LowTemp:      low,
			Conditions:   d.Conditions,
			Icon:         d.Icon,
			PrecipChance: d.PrecipChance,
			PrecipType:   d.PrecipType,
		}
		if !d.Sunrise.IsZero() {
			fdays[i].Sunrise = d.Sunrise.Format(time.RFC3339)
		}
		if !d.Sunset.IsZero() {
			fdays[i].Sunset = d.Sunset.Format(time.RFC3339)
		}
	}
	return forecastJSONOutput{
		Station: stationMeta{
			Name:      sc.Name,
			StationID: sc.StationID,
			DeviceID:  sc.DeviceID,
		},
		Units: units,
		Days:  fdays,
	}
}

func fetchForecastFromAPI(ctx context.Context, sc *config.StationConfig) (*tempest.Forecast, error) {
	client, err := tempest.NewClient(sc.Token)
	if err != nil {
		return nil, fmt.Errorf("creating API client: %w", err)
	}

	forecast, err := client.GetForecast(ctx, sc.StationID)
	if err != nil {
		return nil, fmt.Errorf("fetching forecast: %w", err)
	}

	return forecast, nil
}

func fetchForecastFromServer(ctx context.Context, serverURL string, stationID int) (*tempest.Forecast, error) {
	return fetchFromTempestd[tempest.Forecast](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d/forecast", stationID))
}
