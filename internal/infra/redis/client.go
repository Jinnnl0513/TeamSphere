package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config is nil")
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return rdb, nil
}
