package bootstrap

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/handler"
	"github.com/teamsphere/server/internal/middleware"
	"github.com/teamsphere/server/internal/ratelimit"
	"github.com/teamsphere/server/internal/service"
	"github.com/teamsphere/server/web"
)

// runSetupMode starts the server in setup mode.
// pool is nil when no config exists; non-nil when config exists but no users.
func runSetupMode(configPath string, pool *pgxpool.Pool, jwtCfg *config.JWTConfig, port int) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(setupAllowedOrigins(port)))
	r.Use(middleware.SetupGuard())

	limiter := ratelimit.NewMemoryLimiter()
	limiter.StartCleanup(context.Background(), 5*time.Minute)

	frontendFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		slog.Error("failed to get frontend FS", "error", err)
	} else {
		r.NoRoute(handler.SPA(frontendFS))
	}

	setupDone := make(chan struct{})
	setupService := service.NewSetupService(configPath)
	setupHandler := handler.NewSetupHandler(setupService, configPath, pool, jwtCfg, setupDone)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.Health(pool, nil, nil))

		setup := v1.Group("/setup")
		{
			setup.GET("/status", func(c *gin.Context) {
				status, err := service.GetStatus(configPath, pool)
				if err != nil {
					handler.Error(c, http.StatusInternalServerError, 50001, "internal server error")
					return
				}

				handler.Success(c, status)
			})
			setup.POST("/test-db", middleware.RequireSetupAccess(), middleware.RateLimit(limiter, 10, time.Minute), setupHandler.TestDB)
			setup.POST("/test-connection", middleware.RequireSetupAccess(), middleware.RateLimit(limiter, 10, time.Minute), setupHandler.TestConnection)
			setup.POST("/test-email", middleware.RequireSetupAccess(), middleware.RateLimit(limiter, 5, time.Minute), setupHandler.TestEmail)
			setup.POST("", middleware.RequireSetupAccess(), middleware.RateLimit(limiter, 5, time.Minute), setupHandler.Complete)
		}
	}

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		slog.Info("server starting in setup mode", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-setupDone:
		slog.Info("setup completed, shutting down setup server...")
	case <-quit:
		slog.Info("received signal, shutting down setup server...")
	}

	signal.Stop(quit)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	slog.Info("setup mode server exited")
}
