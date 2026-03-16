package database

import (
	"context"
	"fmt"

	"github.com/teamsphere/server/internal/contract/tx"
	"github.com/teamsphere/server/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a transaction manager backed by pgx.
func NewTxManager(pool *pgxpool.Pool) tx.Manager {
	return &txManager{pool: pool}
}

func (m *txManager) WithTx(ctx context.Context, fn func(t tx.Tx) error) error {
	pgxTx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	t := &txImpl{
		tx:             pgxTx,
		userRepo:       repository.NewUserRepo(pgxTx),
		roomRepo:       repository.NewRoomRepo(pgxTx),
		messageRepo:    repository.NewMessageRepo(pgxTx),
		friendshipRepo: repository.NewFriendshipRepo(pgxTx),
		settingsRepo:   repository.NewSettingsRepo(pgxTx),
		inviteLinkRepo: repository.NewInviteLinkRepo(pgxTx),
	}
	if err := fn(t); err != nil {
		_ = pgxTx.Rollback(ctx)
		return err
	}
	if err := pgxTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

type txImpl struct {
	tx             pgx.Tx
	userRepo       repository.UserRepository
	roomRepo       repository.RoomRepository
	messageRepo    repository.MessageRepository
	friendshipRepo repository.FriendshipRepository
	settingsRepo   repository.SettingsRepository
	inviteLinkRepo repository.InviteLinkRepository
}

func (t *txImpl) UserRepo() repository.UserRepository       { return t.userRepo }
func (t *txImpl) RoomRepo() repository.RoomRepository       { return t.roomRepo }
func (t *txImpl) MessageRepo() repository.MessageRepository { return t.messageRepo }
func (t *txImpl) FriendshipRepo() repository.FriendshipRepository {
	return t.friendshipRepo
}
func (t *txImpl) SettingsRepo() repository.SettingsRepository { return t.settingsRepo }
func (t *txImpl) InviteLinkRepo() repository.InviteLinkRepository {
	return t.inviteLinkRepo
}
