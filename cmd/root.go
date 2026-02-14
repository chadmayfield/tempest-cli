package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/chadmayfield/tempest-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:           "tempest",
	Short:         "Beautiful CLI for WeatherFlow Tempest weather stations",
	Long:          "Query current conditions, forecasts, and historical data from your WeatherFlow Tempest weather station with styled terminal output.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/tempest/config.yaml)")
	rootCmd.PersistentFlags().String("station", "", "station name from config")
	rootCmd.PersistentFlags().String("units", "", "unit system: metric or imperial")
	rootCmd.PersistentFlags().String("server", "", "tempestd server URL for local data")
	rootCmd.PersistentFlags().Bool("json", false, "output as JSON")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")
	rootCmd.PersistentFlags().Bool("no-emoji", false, "use text symbols instead of emoji for condition icons")

	_ = viper.BindPFlag("station", rootCmd.PersistentFlags().Lookup("station"))
	_ = viper.BindPFlag("units", rootCmd.PersistentFlags().Lookup("units"))
	_ = viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	_ = viper.BindPFlag("no-emoji", rootCmd.PersistentFlags().Lookup("no-emoji"))
}

func initConfig() {
	// TEMPEST_CONFIG env var overrides default config path (checked before viper reads config)
	if cfgFile == "" {
		if envCfg := os.Getenv("TEMPEST_CONFIG"); envCfg != "" {
			cfgFile = envCfg
		}
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot determine home directory: %v\n", err)
			return
		}

		configDir := filepath.Join(home, ".config", "tempest")
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("TEMPEST")
	viper.AutomaticEnv()

	// Set defaults per spec
	viper.SetDefault("units", "imperial")

	// Bind env vars that map to specific config keys
	_ = viper.BindEnv("units")
	_ = viper.BindEnv("station")
	_ = viper.BindEnv("server")

	// Detect NO_COLOR environment variable (https://no-color.org/)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		viper.Set("no-color", true)
	}

	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional — only warn if one was explicitly specified
		if cfgFile != "" {
			slog.Warn("cannot read config file", "error", err)
			fmt.Fprintf(os.Stderr, "Warning: cannot read config file: %v\n", err)
		}
	} else {
		slog.Debug("loaded config file", "path", viper.ConfigFileUsed())
		// Warn if config file is world-readable (may contain API tokens)
		checkConfigPermissions(viper.ConfigFileUsed())
	}
}

// resolveServerURL returns the tempestd server URL, checking --server flag first, then config.
// Returns an empty string if no server is configured.
func resolveServerURL(cfg *config.Config) string {
	var s string
	if s = viper.GetString("server"); s == "" && cfg != nil {
		s = cfg.EffectiveServerURL()
	}
	if s == "" {
		return ""
	}
	if err := validateServerURL(s); err != nil {
		slog.Warn("ignoring invalid server URL", "url", s, "error", err)
		return ""
	}
	return s
}

func checkConfigPermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	mode := info.Mode().Perm()
	if mode&0044 != 0 {
		fmt.Fprintf(os.Stderr, "Warning: config file %s is readable by others (mode %04o). Consider: chmod 600 %s\n",
			path, mode, path)
	}
}
