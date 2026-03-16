package ratelimit

import (
	"context"
	"time"
)

// Limiter enforces request rate limits.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
