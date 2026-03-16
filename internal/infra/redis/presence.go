package redis

import (
	"context"
	"fmt"

	"github.com/teamsphere/server/internal/presence"
	"github.com/redis/go-redis/v9"
)

const (
	onlineSetKey   = "presence:online"
	countKeyPrefix = "presence:count:"
)

type PresenceStore struct {
	client *redis.Client
}

func NewPresenceStore(client *redis.Client) presence.Store {
	return &PresenceStore{client: client}
}

func (s *PresenceStore) MarkOnline(ctx context.Context, userID int64) (bool, error) {
	key := countKey(userID)
	cnt, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("presence incr: %w", err)
	}
	if cnt == 1 {
		if err := s.client.SAdd(ctx, onlineSetKey, userID).Err(); err != nil {
			return false, fmt.Errorf("presence sadd: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func (s *PresenceStore) MarkOffline(ctx context.Context, userID int64) (bool, error) {
	key := countKey(userID)
	cnt, err := s.client.Decr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("presence decr: %w", err)
	}
	if cnt <= 0 {
		pipe := s.client.TxPipeline()
		pipe.Del(ctx, key)
		pipe.SRem(ctx, onlineSetKey, userID)
		if _, err := pipe.Exec(ctx); err != nil {
			return false, fmt.Errorf("presence cleanup: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func (s *PresenceStore) IsOnline(ctx context.Context, userID int64) (bool, error) {
	ok, err := s.client.SIsMember(ctx, onlineSetKey, userID).Result()
	if err != nil {
		return false, fmt.Errorf("presence sismember: %w", err)
	}
	return ok, nil
}

func (s *PresenceStore) OnlineCount(ctx context.Context) (int, error) {
	n, err := s.client.SCard(ctx, onlineSetKey).Result()
	if err != nil {
		return 0, fmt.Errorf("presence scard: %w", err)
	}
	return int(n), nil
}

func countKey(userID int64) string {
	return countKeyPrefix + fmt.Sprint(userID)
}
