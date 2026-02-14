package cmd

import (
	"context"
	"fmt"
	"log/slog"
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

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current weather conditions",
	Long:  "Display current weather conditions from your Tempest station with styled output.",
	RunE:  runCurrent,
}

func init() {
	rootCmd.AddCommand(currentCmd)

	// Make current the default command when no subcommand is given
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCurrent(cmd, args)
	}
}

func runCurrent(cmd *cobra.Command, args []string) error {
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

	serverURL := resolveServerURL(cfg)

	var obs *tempest.StationObservation
	var station *tempest.Station

	units := "metric"
	if cfg.IsImperial() {
		units = "imperial"
	}

	slog.Debug("fetching current conditions", "station_id", sc.StationID, "server", serverURL)
	if serverURL != "" {
		obs, station, err = fetchCurrentFromServer(ctx, serverURL, sc.StationID, units)
	} else {
		obs, station, err = fetchCurrentFromAPI(ctx, sc)
	}
	if err != nil {
		return wrapAPIError(err)
	}

	imperial := cfg.IsImperial()

	if viper.GetBool("json") {
		return jsonout.Write(cmd.OutOrStdout(), currentJSON(obs, station, sc, units, imperial))
	}
	noColor := viper.GetBool("no-color")
	theme := display.NewTheme(noColor, display.WithNoEmoji(viper.GetBool("no-emoji")))

	displayName := station.Name
	if displayName == "" {
		displayName = sc.Name
	}

	termWidth := 80
	if w, _, err := term.GetSize(0); err == nil && w > 0 {
		termWidth = w
	}

	output := display.RenderCurrent(theme, obs, displayName, imperial, termWidth)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)

	return nil
}

func fetchCurrentFromAPI(ctx context.Context, sc *config.StationConfig) (*tempest.StationObservation, *tempest.Station, error) {
	client, err := tempest.NewClient(sc.Token)
	if err != nil {
		return nil, nil, fmt.Errorf("creating API client: %w", err)
	}

	obs, err := client.GetStationObservation(ctx, sc.StationID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching observation: %w", err)
	}

	station, err := client.GetStation(ctx, sc.StationID)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching station: %w", err)
	}

	return obs, station, nil
}

type currentJSONOutput struct {
	Station               stationMeta `json:"station"`
	Units                 string      `json:"units"`
	Timestamp             time.Time   `json:"timestamp"`
	Temperature           float64     `json:"temperature"`
	FeelsLike             float64     `json:"feels_like"`
	DewPoint              float64     `json:"dew_point"`
	Humidity              float64     `json:"humidity"`
	WindSpeed             float64     `json:"wind_speed"`
	WindGust              float64     `json:"wind_gust"`
	WindLull              float64     `json:"wind_lull"`
	WindDirection         float64     `json:"wind_direction"`
	WindDirectionCardinal string      `json:"wind_direction_cardinal"`
	Pressure              float64     `json:"pressure"`
	PressureTrend         string      `json:"pressure_trend,omitempty"`
	UVIndex               float64     `json:"uv_index"`
	SolarRadiation        float64     `json:"solar_radiation"`
	RainToday             float64     `json:"rain_today"`
	LightningCount        int         `json:"lightning_count"`
	LightningDistance     float64     `json:"lightning_distance"`
	Battery               float64     `json:"battery"`
}

type stationMeta struct {
	Name      string `json:"name"`
	StationID int    `json:"station_id"`
	DeviceID  int    `json:"device_id"`
}

func currentJSON(obs *tempest.StationObservation, station *tempest.Station, sc *config.StationConfig, units string, imperial bool) currentJSONOutput {
	temp := obs.AirTemperature
	feels := obs.FeelsLike
	dew := obs.DewPoint
	wind := obs.WindAvg
	gust := obs.WindGust
	lull := obs.WindLull
	pressure := obs.SeaLevelPressure
	rain := obs.PrecipAccumDay
	lightningDist := obs.LightningStrikeLastDistance
	if imperial {
		temp = tempest.CelsiusToFahrenheit(temp)
		feels = tempest.CelsiusToFahrenheit(feels)
		dew = tempest.CelsiusToFahrenheit(dew)
		wind = tempest.MpsToMph(wind)
		gust = tempest.MpsToMph(gust)
		lull = tempest.MpsToMph(lull)
		pressure = tempest.HpaToInhg(pressure)
		rain = tempest.MmToInches(rain)
		lightningDist = tempest.KmToMiles(lightningDist)
	}

	deviceID := sc.DeviceID
	stationName := station.Name
	if stationName == "" {
		stationName = sc.Name
	}

	return currentJSONOutput{
		Station: stationMeta{
			Name:      stationName,
			StationID: sc.StationID,
			DeviceID:  deviceID,
		},
		Units:                 units,
		Timestamp:             obs.Timestamp,
		Temperature:           temp,
		FeelsLike:             feels,
		DewPoint:              dew,
		Humidity:              obs.RelativeHumidity,
		WindSpeed:             wind,
		WindGust:              gust,
		WindLull:              lull,
		WindDirection:         obs.WindDirection,
		WindDirectionCardinal: tempest.WindDirectionToCompass(obs.WindDirection),
		Pressure:              pressure,
		PressureTrend:         obs.PressureTrend,
		UVIndex:               obs.UV,
		SolarRadiation:        obs.SolarRadiation,
		RainToday:             rain,
		LightningCount:        obs.LightningCount3hr,
		LightningDistance:     lightningDist,
	}
}

func fetchCurrentFromServer(ctx context.Context, serverURL string, stationID int, units string) (*tempest.StationObservation, *tempest.Station, error) {
	params := url.Values{}
	params.Set("units", units)
	obs, err := fetchFromTempestd[tempest.StationObservation](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d/current?%s", stationID, params.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("fetching observation from tempestd: %w", err)
	}

	station, err := fetchFromTempestd[tempest.Station](ctx, serverURL, fmt.Sprintf("/api/v1/stations/%d", stationID))
	if err != nil {
		return nil, nil, fmt.Errorf("fetching station from tempestd: %w", err)
	}

	return obs, station, nil
}
