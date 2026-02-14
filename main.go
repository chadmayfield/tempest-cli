package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/chadmayfield/tempest-cli/cmd"
)

func main() {
	// Set up structured logging; default to warn level, debug if TEMPEST_DEBUG is set
	logLevel := slog.LevelWarn
	if os.Getenv("TEMPEST_DEBUG") != "" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := cmd.Execute(ctx); err != nil {
		// Check if the error was caused by context cancellation (SIGINT)
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "Interrupted")
			os.Exit(130)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
