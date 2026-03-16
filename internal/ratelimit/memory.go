package ratelimit

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	count       int
	windowStart time.Time
	window      time.Duration
}

// MemoryLimiter is a process-local rate limiter.
type MemoryLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{entries: make(map[string]*entry)}
}

// StartCleanup removes expired entries periodically. Call once on startup.
func (l *MemoryLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				l.cleanup()
			}
		}
	}()
}

func (l *MemoryLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	_ = ctx
	if limit <= 0 || window <= 0 {
		return true, nil
	}
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok || now.Sub(e.windowStart) >= window {
		l.entries[key] = &entry{count: 1, windowStart: now, window: window}
		return true, nil
	}
	e.count++
	return e.count <= limit, nil
}

func (l *MemoryLimiter) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	for key, e := range l.entries {
		if now.Sub(e.windowStart) >= e.window {
			delete(l.entries, key)
		}
	}
}
