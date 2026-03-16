package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter is a fixed-window rate limiter backed by Redis.
type RedisLimiter struct {
	client *redis.Client
	prefix string
}

func NewRedisLimiter(client *redis.Client, prefix string) *RedisLimiter {
	if prefix == "" {
		prefix = "rl:"
	}
	return &RedisLimiter{client: client, prefix: prefix}
}

var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 || window <= 0 {
		return true, nil
	}
	ttlMs := window.Milliseconds()
	if ttlMs <= 0 {
		return true, nil
	}
	redisKey := fmt.Sprintf("%s%s:%d:%d", l.prefix, key, limit, ttlMs)
	val, err := rateLimitScript.Run(ctx, l.client, []string{redisKey}, ttlMs).Int64()
	if err != nil {
		return false, err
	}
	return val <= int64(limit), nil
}
