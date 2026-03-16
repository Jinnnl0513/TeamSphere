package service

import (
	"context"
	"testing"
	"time"

	"github.com/teamsphere/server/internal/contract/tx"
	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

type mockTxManager struct {
	tx tx.Tx
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(tx.Tx) error) error {
	return fn(m.tx)
}

type mockTx struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
}

func (t *mockTx) UserRepo() repository.UserRepository       { return t.userRepo }
func (t *mockTx) RoomRepo() repository.RoomRepository       { return t.roomRepo }
func (t *mockTx) MessageRepo() repository.MessageRepository { return nil }
func (t *mockTx) FriendshipRepo() repository.FriendshipRepository {
	return nil
}
func (t *mockTx) SettingsRepo() repository.SettingsRepository { return nil }
func (t *mockTx) InviteLinkRepo() repository.InviteLinkRepository {
	return nil
}

type mockRoomRepo struct {
	invite            *model.RoomInvite
	getInviteCalls    int
	updateStatusCalls []string
	addMemberCalls    int
}

func (m *mockRoomRepo) Create(ctx context.Context, name, description string, creatorID int64) (*model.Room, error) {
	return nil, nil
}
func (m *mockRoomRepo) GetByID(ctx context.Context, id int64) (*model.Room, error) { return nil, nil }
func (m *mockRoomRepo) Update(ctx context.Context, id int64, name, description string) (*model.Room, error) {
	return nil, nil
}
func (m *mockRoomRepo) Delete(ctx context.Context, id int64) error { return nil }
func (m *mockRoomRepo) ListByUser(ctx context.Context, userID int64) ([]*model.Room, error) {
	return nil, nil
}
func (m *mockRoomRepo) DiscoverAll(ctx context.Context) ([]*model.Room, error) { return nil, nil }
func (m *mockRoomRepo) GetMember(ctx context.Context, roomID, userID int64) (*model.RoomMember, error) {
	return nil, nil
}
func (m *mockRoomRepo) ListMemberIDs(ctx context.Context, roomID int64) ([]int64, error) {
	return nil, nil
}
func (m *mockRoomRepo) IsMember(ctx context.Context, roomID, userID int64) (bool, error) {
	return false, nil
}
func (m *mockRoomRepo) HasOwnedRooms(ctx context.Context, userID int64) (bool, error) {
	return false, nil
}
func (m *mockRoomRepo) ListMembers(ctx context.Context, roomID int64) ([]*repository.MemberInfo, error) {
	return nil, nil
}
func (m *mockRoomRepo) AddMember(ctx context.Context, roomID, userID int64, role string) error {
	m.addMemberCalls++
	return nil
}
func (m *mockRoomRepo) RemoveMember(ctx context.Context, roomID, userID int64) error { return nil }
func (m *mockRoomRepo) UpdateMemberRole(ctx context.Context, roomID, userID int64, role string) error {
	return nil
}
func (m *mockRoomRepo) SetMuted(ctx context.Context, roomID, userID int64, mutedUntil *time.Time) error {
	return nil
}
func (m *mockRoomRepo) TransferOwner(ctx context.Context, roomID, oldOwnerID, newOwnerID int64) error {
	return nil
}
func (m *mockRoomRepo) CreateInvite(ctx context.Context, roomID, inviterID, inviteeID int64) (*model.RoomInvite, error) {
	return nil, nil
}
func (m *mockRoomRepo) GetInviteByID(ctx context.Context, id int64) (*model.RoomInvite, error) {
	m.getInviteCalls++
	return m.invite, nil
}
func (m *mockRoomRepo) HasPendingInvite(ctx context.Context, roomID, inviteeID int64) (bool, error) {
	return false, nil
}
func (m *mockRoomRepo) UpdateInviteStatus(ctx context.Context, id int64, status string) error {
	m.updateStatusCalls = append(m.updateStatusCalls, status)
	return nil
}
func (m *mockRoomRepo) ListPendingInvitesByUser(ctx context.Context, userID int64) ([]*repository.InviteInfo, error) {
	return nil, nil
}
func (m *mockRoomRepo) AreFriends(ctx context.Context, userA, userB int64) (bool, error) {
	return true, nil
}
func (m *mockRoomRepo) CountMembers(ctx context.Context, roomID int64) (int, error) { return 0, nil }
func (m *mockRoomRepo) ListMemberUsernames(ctx context.Context, roomID int64) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockRoomRepo) CountAll(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockRoomRepo) ListAllRooms(ctx context.Context, offset, limit int) ([]*repository.RoomWithMemberCount, error) {
	return nil, nil
}

type mockUserRepo struct {
	user *model.User
}

func (m *mockUserRepo) Create(ctx context.Context, username, password, role, email string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return m.user, nil
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListUserInfosByIDs(ctx context.Context, ids []int64) ([]model.UserInfo, error) {
	return nil, nil
}
func (m *mockUserRepo) ExistsAny(ctx context.Context) (bool, error) { return false, nil }
func (m *mockUserRepo) BlacklistToken(ctx context.Context, entry *model.TokenBlacklist) error {
	return nil
}
func (m *mockUserRepo) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) CleanExpiredTokens(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) UpdateBioAndColor(ctx context.Context, id int64, bio, profileColor string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) UpdateAvatar(ctx context.Context, id int64, avatarURL string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) SoftDelete(ctx context.Context, id int64) error { return nil }
func (m *mockUserRepo) ListAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) CountAll(ctx context.Context) (int64, error)                 { return 0, nil }
func (m *mockUserRepo) CountActive(ctx context.Context) (int64, error)              { return 0, nil }
func (m *mockUserRepo) UpdateRole(ctx context.Context, id int64, role string) error { return nil }
func (m *mockUserRepo) HardDelete(ctx context.Context, id int64) error              { return nil }
func (m *mockUserRepo) SetEmailVerified(ctx context.Context, id int64, email string) error {
	return nil
}
func (m *mockUserRepo) CreateEmailVerification(ctx context.Context, email, code string, expiresAt time.Time) (*model.EmailVerification, error) {
	return nil, nil
}
func (m *mockUserRepo) GetLatestVerification(ctx context.Context, email string) (*model.EmailVerification, error) {
	return nil, nil
}
func (m *mockUserRepo) MarkVerificationUsed(ctx context.Context, id int64) error { return nil }
func (m *mockUserRepo) IncrementVerificationAttempts(ctx context.Context, id int64) (int, error) {
	return 0, nil
}
func (m *mockUserRepo) CleanExpiredVerifications(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ipAddress, userAgent, deviceName *string) (*model.RefreshToken, error) {
	return nil, nil
}
func (m *mockUserRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	return nil, nil
}
func (m *mockUserRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error { return nil }
func (m *mockUserRepo) RevokeAllRefreshTokensForUser(ctx context.Context, userID int64) error {
	return nil
}
func (m *mockUserRepo) RevokeRefreshTokenByID(ctx context.Context, userID, tokenID int64) error {
	return nil
}
func (m *mockUserRepo) ListRefreshTokens(ctx context.Context, userID int64) ([]*model.RefreshToken, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdateRefreshTokenLastUsed(ctx context.Context, tokenHash string) error {
	return nil
}
func (m *mockUserRepo) RevokeOtherRefreshTokens(ctx context.Context, userID int64, tokenHash string) error {
	return nil
}
func (m *mockUserRepo) CleanExpiredRefreshTokens(ctx context.Context) (int64, error) { return 0, nil }

func TestRoomService_RespondInvite_Accept_WithTx(t *testing.T) {
	invite := &model.RoomInvite{ID: 1, RoomID: 10, InviterID: 2, InviteeID: 3, Status: "pending"}
	roomRepo := &mockRoomRepo{invite: invite}
	userRepo := &mockUserRepo{user: &model.User{ID: 3, Username: "u"}}
	txMgr := &mockTxManager{tx: &mockTx{roomRepo: roomRepo, userRepo: userRepo}}

	hub := ws.NewHub(
		repository.NewMessageRepo(nil),
		repository.NewRoomRepo(nil),
		repository.NewRoomSettingsRepo(nil),
		repository.NewUserRepo(nil),
		repository.NewFriendshipRepo(nil),
		repository.NewMessageReadRepo(nil),
		repository.NewNotificationRepo(nil),
		nil,
	)

	svc := NewRoomService(roomRepo, userRepo, nil, hub, txMgr)
	if err := svc.RespondInvite(context.Background(), 3, 1, true); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if roomRepo.addMemberCalls != 1 {
		t.Fatalf("expected AddMember to be called once, got %d", roomRepo.addMemberCalls)
	}
	if len(roomRepo.updateStatusCalls) != 1 || roomRepo.updateStatusCalls[0] != "accepted" {
		t.Fatalf("expected UpdateInviteStatus accepted, got %v", roomRepo.updateStatusCalls)
	}
}

func TestRoomService_RespondInvite_Decline_WithTx(t *testing.T) {
	invite := &model.RoomInvite{ID: 1, RoomID: 10, InviterID: 2, InviteeID: 3, Status: "pending"}
	roomRepo := &mockRoomRepo{invite: invite}
	userRepo := &mockUserRepo{user: &model.User{ID: 3, Username: "u"}}
	txMgr := &mockTxManager{tx: &mockTx{roomRepo: roomRepo, userRepo: userRepo}}

	hub := ws.NewHub(
		repository.NewMessageRepo(nil),
		repository.NewRoomRepo(nil),
		repository.NewRoomSettingsRepo(nil),
		repository.NewUserRepo(nil),
		repository.NewFriendshipRepo(nil),
		repository.NewMessageReadRepo(nil),
		repository.NewNotificationRepo(nil),
		nil,
	)

	svc := NewRoomService(roomRepo, userRepo, nil, hub, txMgr)
	if err := svc.RespondInvite(context.Background(), 3, 1, false); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if roomRepo.addMemberCalls != 0 {
		t.Fatalf("expected AddMember not to be called, got %d", roomRepo.addMemberCalls)
	}
	if len(roomRepo.updateStatusCalls) != 1 || roomRepo.updateStatusCalls[0] != "declined" {
		t.Fatalf("expected UpdateInviteStatus declined, got %v", roomRepo.updateStatusCalls)
	}
}
