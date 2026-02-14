package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chadmayfield/tempest-cli/internal/config"
	tempest "github.com/chadmayfield/tempest-go"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive configuration wizard",
	RunE:  runConfigInit,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if viper.GetBool("json") {
		redacted := redactedConfig(cfg)
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(redacted)
	}

	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(w, "Config file: %s\n\n", viper.ConfigFileUsed())
	_, _ = fmt.Fprintf(w, "Default station: %s\n", cfg.DefaultStation)
	_, _ = fmt.Fprintf(w, "Units:           %s\n", displayUnits(cfg.Units))
	if server := cfg.EffectiveServerURL(); server != "" {
		_, _ = fmt.Fprintf(w, "Server:          %s\n", server)
	}
	_, _ = fmt.Fprintln(w)

	for name, sc := range cfg.Stations {
		def := ""
		if name == cfg.DefaultStation {
			def = " (default)"
		}
		_, _ = fmt.Fprintf(w, "Station: %s%s\n", name, def)
		_, _ = fmt.Fprintf(w, "  Name:       %s\n", sc.Name)
		_, _ = fmt.Fprintf(w, "  Token:      %s\n", config.RedactToken(sc.Token))
		_, _ = fmt.Fprintf(w, "  Station ID: %d\n", sc.StationID)
		_, _ = fmt.Fprintf(w, "  Device ID:  %d\n", sc.DeviceID)
		_, _ = fmt.Fprintln(w)
	}

	return nil
}

func displayUnits(u string) string {
	if u == "" {
		return "metric (default)"
	}
	return u
}

type redactedStationConfig struct {
	Token     string `json:"token"`
	StationID int    `json:"station_id"`
	DeviceID  int    `json:"device_id"`
	Name      string `json:"name"`
}

func redactedConfig(cfg *config.Config) map[string]any {
	stations := make(map[string]redactedStationConfig)
	for name, sc := range cfg.Stations {
		stations[name] = redactedStationConfig{
			Token:     config.RedactToken(sc.Token),
			StationID: sc.StationID,
			DeviceID:  sc.DeviceID,
			Name:      sc.Name,
		}
	}
	return map[string]any{
		"default_station": cfg.DefaultStation,
		"units":           cfg.Units,
		"stations":        stations,
		"server":          cfg.EffectiveServerURL(),
		"config_file":     viper.ConfigFileUsed(),
	}
}

// -- Config init: bubbletea wizard --
// Note: The spec calls for "fetching available stations" but the tempest-go
// library (and the WeatherFlow REST API) only supports GetStation(id), not
// listing all stations for a token. The wizard therefore asks the user to
// enter their station ID (visible at tempestwx.com) and verifies it via API.

type initStep int

const (
	stepToken initStep = iota
	stepStationID
	stepFetchStation
	stepNameStation
	stepUnits
	stepWriting
	stepDone
)

type initModel struct {
	step     initStep
	token    string
	input    string
	station  *tempest.Station
	name     string
	units    string
	err      error
	written  string
	selected int // for units selection (0=imperial, 1=metric)
}

type stationFetchedMsg struct {
	station *tempest.Station
	err     error
}

type configWrittenMsg struct {
	path string
	err  error
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "up":
			if m.step == stepUnits && m.selected > 0 {
				m.selected--
			}
		case "down":
			if m.step == stepUnits && m.selected < 1 {
				m.selected++
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}

	case stationFetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepStationID
			m.input = ""
			return m, nil
		}
		m.station = msg.station
		m.step = stepNameStation
		// Pre-fill name suggestion
		m.input = strings.ToLower(strings.ReplaceAll(m.station.Name, " ", "-"))

	case configWrittenMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.written = msg.path
		m.step = stepDone
		return m, tea.Quit
	}

	return m, nil
}

func (m initModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepToken:
		m.token = strings.TrimSpace(m.input)
		if m.token == "" {
			return m, nil
		}
		m.input = ""
		m.err = nil
		m.step = stepStationID

	case stepStationID:
		idStr := strings.TrimSpace(m.input)
		if idStr == "" {
			return m, nil
		}
		if _, err := strconv.Atoi(idStr); err != nil {
			m.err = fmt.Errorf("invalid station ID: must be a number")
			m.input = ""
			return m, nil
		}
		m.err = nil
		m.step = stepFetchStation
		return m, m.fetchStation

	case stepNameStation:
		m.name = strings.TrimSpace(m.input)
		if m.name == "" {
			return m, nil
		}
		m.input = ""
		m.step = stepUnits
		m.selected = 0

	case stepUnits:
		if m.selected == 0 {
			m.units = "imperial"
		} else {
			m.units = "metric"
		}
		m.step = stepWriting
		return m, m.writeConfig
	}

	return m, nil
}

func (m initModel) fetchStation() tea.Msg {
	client, err := tempest.NewClient(m.token)
	if err != nil {
		return stationFetchedMsg{err: err}
	}

	idStr := strings.TrimSpace(m.input)
	stationID, _ := strconv.Atoi(idStr)

	station, err := client.GetStation(context.TODO(), stationID)
	if err != nil {
		return stationFetchedMsg{err: fmt.Errorf("could not fetch station %d: %w (check your token and station ID)", stationID, err)}
	}

	return stationFetchedMsg{station: station}
}

func (m initModel) writeConfig() tea.Msg {
	home, err := os.UserHomeDir()
	if err != nil {
		return configWrittenMsg{err: err}
	}

	configDir := filepath.Join(home, ".config", "tempest")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return configWrittenMsg{err: err}
	}

	deviceID := 0
	if len(m.station.Devices) > 0 {
		deviceID = m.station.Devices[0].DeviceID
	}

	cfg := config.Config{
		DefaultStation: m.name,
		Units:          m.units,
		Stations: map[string]config.StationConfig{
			m.name: {
				Token:     m.token,
				StationID: m.station.StationID,
				DeviceID:  deviceID,
				Name:      m.station.Name,
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return configWrittenMsg{err: err}
	}

	path := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return configWrittenMsg{err: err}
	}

	return configWrittenMsg{path: path}
}

func (m initModel) View() string {
	var b strings.Builder

	b.WriteString("Tempest CLI Configuration\n\n")

	switch m.step {
	case stepToken:
		b.WriteString("Enter your WeatherFlow API token:\n")
		b.WriteString("> " + strings.Repeat("*", len(m.input)) + "█\n")

	case stepStationID:
		b.WriteString(fmt.Sprintf("Token: %s****\n\n", m.token[:min(4, len(m.token))]))
		b.WriteString("Enter your station ID:\n")
		b.WriteString("> " + m.input + "█\n")

	case stepFetchStation:
		b.WriteString("Verifying station...\n")

	case stepNameStation:
		b.WriteString(fmt.Sprintf("Found station: %s\n\n", m.station.Name))
		b.WriteString("Config name for this station (e.g., home, office):\n")
		b.WriteString("> " + m.input + "█\n")

	case stepUnits:
		units := []string{"Imperial (°F, mph, inHg)", "Metric (°C, m/s, hPa)"}
		b.WriteString("Select units:\n\n")
		for i, u := range units {
			cursor := "  "
			if i == m.selected {
				cursor = "> "
			}
			b.WriteString(cursor + u + "\n")
		}
		b.WriteString("\nUse arrow keys to select, Enter to confirm\n")

	case stepWriting:
		b.WriteString("Writing config...\n")

	case stepDone:
		b.WriteString(fmt.Sprintf("Config written to %s\n\n", m.written))
		b.WriteString("Run 'tempest current' to see current conditions.\n")
	}

	if m.err != nil {
		b.WriteString(fmt.Sprintf("\nError: %v\n", m.err))
	}

	return b.String()
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	m := initModel{step: stepToken}
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("config wizard error: %w", err)
	}

	final := finalModel.(initModel)
	if final.err != nil {
		return final.err
	}

	if final.step != stepDone {
		return fmt.Errorf("configuration cancelled")
	}

	return nil
}
