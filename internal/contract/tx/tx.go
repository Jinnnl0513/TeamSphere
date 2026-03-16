package tx

import (
	"context"

	"github.com/teamsphere/server/internal/repository"
)

// Tx exposes repositories bound to a single database transaction.
type Tx interface {
	UserRepo() repository.UserRepository
	RoomRepo() repository.RoomRepository
	MessageRepo() repository.MessageRepository
	FriendshipRepo() repository.FriendshipRepository
	SettingsRepo() repository.SettingsRepository
	InviteLinkRepo() repository.InviteLinkRepository
}

// Manager executes a function within a database transaction.
type Manager interface {
	WithTx(ctx context.Context, fn func(tx Tx) error) error
}
