package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/teamsphere/server/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Health returns a deep health check handler.
func Health(pool *pgxpool.Pool, redisClient *redis.Client, hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbOK := pool != nil && pool.Ping(ctx) == nil
		redisOK := true
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				redisOK = false
			}
		}
		hubOK := true
		if hub != nil {
			select {
			case <-hub.Done():
				hubOK = false
			default:
			}
		}

		if dbOK && redisOK && hubOK {
			Success(c, gin.H{"status": "ok"})
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    50301,
			"message": "service unhealthy",
			"data": gin.H{
				"db":    dbOK,
				"redis": redisOK,
				"hub":   hubOK,
			},
		})
	}
}

// Ready returns readiness probe handler (DB/Redis only).
func Ready(pool *pgxpool.Pool, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		dbOK := pool != nil && pool.Ping(ctx) == nil
		redisOK := true
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				redisOK = false
			}
		}
		if dbOK && redisOK {
			Success(c, gin.H{"status": "ready"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    50302,
			"message": "not ready",
			"data": gin.H{
				"db":    dbOK,
				"redis": redisOK,
			},
		})
	}
}
