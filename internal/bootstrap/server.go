package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/database"
	"github.com/teamsphere/server/internal/repository"
	"github.com/gin-gonic/gin"
)

// TryStart attempts to start the server. Returns true if it ran in normal mode.
func TryStart() bool {
	configPath := resolveConfigPath()

	if !config.Exists(configPath) {
		slog.Warn("config file not found, starting in setup mode", "path", configPath)
		runSetupMode(configPath, nil, nil, defaultSetupPort)
		return false
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load config, entering setup recovery mode", "error", err, "path", configPath)
		runSetupMode(configPath, nil, nil, defaultSetupPort)
		return false
	}

	port := cfg.Server.Port
	if port <= 0 {
		port = defaultSetupPort
	}

	gin.SetMode(cfg.Server.Mode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, &cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database, entering setup recovery mode", "error", err, "path", configPath)
		runSetupMode(configPath, nil, nil, port)
		return false
	}
	defer pool.Close()
	slog.Info("database connected")

	if err := database.Migrate(ctx, pool); err != nil {
		slog.Error("failed to run migrations, entering setup recovery mode", "error", err, "path", configPath)
		runSetupMode(configPath, nil, nil, port)
		return false
	}

	userRepo := repository.NewUserRepo(pool)
	hasUsers, err := userRepo.ExistsAny(ctx)
	if err != nil {
		slog.Error("failed to check users, entering setup recovery mode", "error", err, "path", configPath)
		runSetupMode(configPath, nil, nil, port)
		return false
	}
	if !hasUsers {
		slog.Warn("no users found, starting in setup mode (admin creation)")
		runSetupMode(configPath, pool, &cfg.JWT, port)
		return false
	}

	deps, err := initDeps(ctx, pool, cfg)
	if err != nil {
		slog.Error("failed to init dependencies", "error", err)
		return true
	}
	defer deps.Close()
	deps.authService.StartBlacklistCleanup(ctx)
	deps.wsHandler.StartTicketCleanup(ctx)
	go deps.hub.Run(ctx)

	router := setupRouter(cfg, deps)
	serve(router, port, cancel, deps.hub.Done())
	return true
}

func serve(router *gin.Engine, port int, cancel context.CancelFunc, hubDone <-chan struct{}) {
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down server", "signal", sig.String())

	cancel()

	select {
	case <-hubDone:
		slog.Info("hub stopped cleanly")
	case <-time.After(5 * time.Second):
		slog.Warn("hub did not stop within 5 s, proceeding with shutdown")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}
	slog.Info("server exited")
}
