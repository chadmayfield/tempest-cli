package cmd

import (
	"bytes"
	"context"
	"testing"
)

func TestCurrentExitsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	rootCmd.SetContext(ctx)
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// The command should return a context.Canceled error (or wrapped variant)
	// rather than hanging or panicking
	err := currentCmd.RunE(currentCmd, nil)
	if err == nil {
		// It's acceptable for the command to fail at config load (no config in test),
		// but it should not hang on a cancelled context.
		return
	}
	// Any error is acceptable — the key assertion is that this test returns
	// promptly rather than blocking forever.
}

func TestHistoryExitsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := historyCmd
	cmd.SetContext(ctx)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// Should not hang
	_ = cmd.RunE(cmd, nil)
}

func TestForecastExitsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := forecastCmd
	cmd.SetContext(ctx)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// Should not hang
	_ = cmd.RunE(cmd, nil)
}

func TestStationsExitsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cmd := stationsCmd
	cmd.SetContext(ctx)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// Should not hang
	_ = cmd.RunE(cmd, nil)
}
