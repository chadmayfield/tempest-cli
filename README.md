# tempest-cli

Beautiful command-line tool for querying WeatherFlow Tempest weather station data. Displays current conditions, historical observations, and forecasts with styled terminal output using the [Charm](https://charm.sh) ecosystem. Supports multiple stations, JSON output for scripting, and optional integration with [tempestd](https://github.com/chadmayfield/tempestd) for local data.

## Installation

### From source

```bash
go install github.com/chadmayfield/tempest-cli@latest
```

### From releases

Download the latest binary for your platform from the [releases page](https://github.com/chadmayfield/tempest-cli/releases).

### Build from source

```bash
git clone https://github.com/chadmayfield/tempest-cli.git
cd tempest-cli
go build -o tempest .
```

## Quick Start

1. Get a WeatherFlow API token from [tempestwx.com](https://tempestwx.com/settings/tokens)
2. Run the configuration wizard:
   ```bash
   tempest config init
   ```
3. View current conditions:
   ```bash
   tempest
   ```

## Commands

### `tempest current` (default)

Display current weather conditions in a styled box. This is the default command when running `tempest` with no arguments.

```bash
tempest                          # current conditions (default)
tempest current                  # explicit
tempest current --json           # JSON output
tempest current --station office # specific station
```

### `tempest forecast`

Show multi-day weather forecast as side-by-side cards.

```bash
tempest forecast                 # 5-day forecast (default)
tempest forecast --days 10       # 10-day forecast
tempest forecast --json          # JSON output
```

### `tempest history`

Display historical observations as a table.

```bash
tempest history                              # last 24 hours
tempest history --date 2024-01-15            # single day
tempest history --from 2024-01-01 --to 2024-01-31  # date range
tempest history --resolution 5m              # specific resolution
```

Resolution options: `1m`, `5m`, `30m`, `3h`. Auto-selected by range if omitted.

### `tempest stations`

List all configured stations with online/offline status.

```bash
tempest stations
tempest stations --json
```

### `tempest config`

Manage configuration.

```bash
tempest config init    # interactive setup wizard
tempest config show    # show current config (tokens redacted)
```

### `tempest version`

Print version, commit, and build information.

### `tempest completion`

Generate shell completion scripts.

```bash
source <(tempest completion bash)
source <(tempest completion zsh)
tempest completion fish | source
```

## Configuration

Configuration is stored in `~/.config/tempest/config.yaml`:

```yaml
default_station: home
units: imperial
stations:
  home:
    token: your-api-token-here
    station_id: 12345
    device_id: 67890
    name: Home Station
  office:
    token: another-token
    station_id: 54321
    device_id: 9876
    name: Office Station

# Optional: configure tempestd for local data
tempestd:
  server: "http://localhost:8080"
```

### Precedence

Configuration values are resolved in order (highest priority first):

1. Command-line flags (`--station`, `--units`, etc.)
2. Environment variables (`TEMPEST_STATION`, `TEMPEST_UNITS`, etc.)
3. Config file

### Environment Variables

All config keys can be set via environment variables with the `TEMPEST_` prefix:

| Variable | Description |
|----------|-------------|
| `TEMPEST_TOKEN` | API token (overrides default station's token) |
| `TEMPEST_STATION_ID` | Station ID (overrides default station's station_id) |
| `TEMPEST_DEVICE_ID` | Device ID (overrides default station's device_id) |
| `TEMPEST_STATION` | Station name to use |
| `TEMPEST_UNITS` | Unit system (`metric` or `imperial`) |
| `TEMPEST_SERVER` | tempestd server URL |
| `NO_COLOR` | Disable colored output (any value) |

## Global Flags

| Flag | Description |
|------|-------------|
| `--station` | Station name from config |
| `--units` | Unit system: `metric` or `imperial` |
| `--json` | Output as JSON for scripting |
| `--no-color` | Disable colored output |
| `--no-emoji` | Use text labels instead of Unicode symbols for condition icons |
| `--server` | tempestd server URL for local data |
| `--config` | Config file path |

## tempestd Integration

When `--server` is specified, tempest-cli queries a local [tempestd](https://github.com/chadmayfield/tempestd) instance instead of the WeatherFlow cloud API. This is useful for local-only setups or reducing API calls.

```bash
tempest --server http://localhost:8080 current
```

## Development

```bash
# Build
go build -o tempest .

# Test
go test ./... -race -count=1

# Lint
golangci-lint run

# Build with version info
go build -ldflags "-X github.com/chadmayfield/tempest-cli/cmd.version=1.0.0 \
  -X github.com/chadmayfield/tempest-cli/cmd.commit=$(git rev-parse --short HEAD) \
  -X github.com/chadmayfield/tempest-cli/cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o tempest .
```

## License

MIT License - see [LICENSE](LICENSE) for details.
