package main

import (
	"log/slog"
	"os"

	"github.com/teamsphere/server/internal/bootstrap"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	for {
		started := bootstrap.TryStart()
		if started {
			return
		}
		slog.Info("setup completed, restarting in normal mode...")
	}
}
