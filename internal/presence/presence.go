package presence

import "context"

// Store manages online presence across instances.
// Implementations should be safe for concurrent use.
type Store interface {
	// MarkOnline increments presence for the user and returns true
	// if this transition makes the user globally online.
	MarkOnline(ctx context.Context, userID int64) (bool, error)

	// MarkOffline decrements presence for the user and returns true
	// if this transition makes the user globally offline.
	MarkOffline(ctx context.Context, userID int64) (bool, error)

	// IsOnline checks whether the user is currently online.
	IsOnline(ctx context.Context, userID int64) (bool, error)

	// OnlineCount returns the number of unique online users.
	OnlineCount(ctx context.Context) (int, error)
}

// NoopStore is a fallback presence store (single-instance only).
type NoopStore struct{}

func (NoopStore) MarkOnline(context.Context, int64) (bool, error)  { return true, nil }
func (NoopStore) MarkOffline(context.Context, int64) (bool, error) { return true, nil }
func (NoopStore) IsOnline(context.Context, int64) (bool, error)    { return false, nil }
func (NoopStore) OnlineCount(context.Context) (int, error)         { return 0, nil }
