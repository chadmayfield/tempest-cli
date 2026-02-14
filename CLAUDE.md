# CLAUDE.md — tempest-cli

## Project Overview
CLI tool for querying WeatherFlow Tempest weather station data. Uses the `tempest-go` library for all API interactions, types, and unit conversions. The CLI handles commands, configuration, and display only.

## Build & Test Commands
- **Build:** `go build -o tempest .`
- **Test:** `go test ./... -race -count=1`
- **Lint:** `golangci-lint run`
- **Vet:** `go vet ./...`
- **Single test:** `go test ./internal/config -run TestConfigLoad -v`

## Architecture
- `main.go` — entrypoint with signal handling (SIGINT → "Interrupted" on stderr, exit 130) and slog setup
- `cmd/` — cobra commands (root, version, current, forecast, history, stations, config, completion)
- `cmd/tempestd.go` — generic HTTP helper for tempestd with 30s connect + 60s total timeout and 10MB body limit
- `cmd/errors.go` — user-friendly error wrapping (401, 404, 429, circuit breaker, network, DNS, missing config)
- `internal/config/` — viper-based config loading, validation, station resolution, env var overrides
- `internal/display/` — lipgloss/bubbles terminal rendering (theme, current, forecast, history, stations)
- `internal/json/` — JSON output helpers

## Key Conventions
- All commands support `--json` for machine-readable output with spec-compliant flat structures (station metadata, units, snake_case keys)
- Unit conversion: `StationObservation` and `Forecast` types need manual per-field conversion; `Observation` type uses `tempest.ConvertObservation()`
- Errors go to stderr; `SilenceUsage` and `SilenceErrors` are true on root command
- Config precedence: flags > env (TEMPEST_ prefix) > config file
- Station resolution: `--station` flag → lookup by name in config map → default_station
- `--no-color` flag and `NO_COLOR` env var disable all styling
- `--server` flag or `tempestd.server` config key routes requests to tempestd instead of WeatherFlow API
- Environment variables: TEMPEST_TOKEN, TEMPEST_STATION_ID, TEMPEST_DEVICE_ID override the default station's values
- Structured logging via `log/slog` (debug level enabled with TEMPEST_DEBUG env var)

## Dependencies
- `github.com/chadmayfield/tempest-go` — API client, types, unit conversion, derived metrics
- `github.com/spf13/cobra` + `github.com/spf13/viper` — CLI framework and config
- `github.com/charmbracelet/lipgloss` — styled terminal output
- `github.com/charmbracelet/bubbletea` + `github.com/charmbracelet/bubbles` — interactive TUI (config init wizard, table rendering)
- `golang.org/x/term` — terminal width detection for responsive layout
- `log/slog` (stdlib) — structured logging

## Version Info
Set via ldflags: `-X github.com/chadmayfield/tempest-cli/cmd.version=... -X ...commit=... -X ...date=...`
