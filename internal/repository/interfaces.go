package repository

import (
	"context"
	"time"

	"github.com/teamsphere/server/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, username, password, role, email string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	ListUserInfosByIDs(ctx context.Context, ids []int64) ([]model.UserInfo, error)
	ExistsAny(ctx context.Context) (bool, error)
	BlacklistToken(ctx context.Context, entry *model.TokenBlacklist) error
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
	CleanExpiredTokens(ctx context.Context) (int64, error)
	UpdateBioAndColor(ctx context.Context, id int64, bio, profileColor string) (*model.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	UpdateAvatar(ctx context.Context, id int64, avatarURL string) (*model.User, error)
	SoftDelete(ctx context.Context, id int64) error
	ListAll(ctx context.Context, offset, limit int) ([]*model.User, error)
	CountAll(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	UpdateRole(ctx context.Context, id int64, role string) error
	HardDelete(ctx context.Context, id int64) error
	SetEmailVerified(ctx context.Context, id int64, email string) error
	CreateEmailVerification(ctx context.Context, email, code string, expiresAt time.Time) (*model.EmailVerification, error)
	GetLatestVerification(ctx context.Context, email string) (*model.EmailVerification, error)
	MarkVerificationUsed(ctx context.Context, id int64) error
	IncrementVerificationAttempts(ctx context.Context, id int64) (int, error)
	CleanExpiredVerifications(ctx context.Context) (int64, error)
	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ipAddress, userAgent, deviceName *string) (*model.RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID int64) error
	RevokeRefreshTokenByID(ctx context.Context, userID, tokenID int64) error
	ListRefreshTokens(ctx context.Context, userID int64) ([]*model.RefreshToken, error)
	UpdateRefreshTokenLastUsed(ctx context.Context, tokenHash string) error
	RevokeOtherRefreshTokens(ctx context.Context, userID int64, tokenHash string) error
	CleanExpiredRefreshTokens(ctx context.Context) (int64, error)
}

type RoomRepository interface {
	Create(ctx context.Context, name, description string, creatorID int64) (*model.Room, error)
	GetByID(ctx context.Context, id int64) (*model.Room, error)
	Update(ctx context.Context, id int64, name, description string) (*model.Room, error)
	Delete(ctx context.Context, id int64) error
	ListByUser(ctx context.Context, userID int64) ([]*model.Room, error)
	DiscoverAll(ctx context.Context) ([]*model.Room, error)
	GetMember(ctx context.Context, roomID, userID int64) (*model.RoomMember, error)
	IsMember(ctx context.Context, roomID, userID int64) (bool, error)
	HasOwnedRooms(ctx context.Context, userID int64) (bool, error)
	ListMembers(ctx context.Context, roomID int64) ([]*MemberInfo, error)
	ListMemberIDs(ctx context.Context, roomID int64) ([]int64, error)
	AddMember(ctx context.Context, roomID, userID int64, role string) error
	RemoveMember(ctx context.Context, roomID, userID int64) error
	UpdateMemberRole(ctx context.Context, roomID, userID int64, role string) error
	SetMuted(ctx context.Context, roomID, userID int64, mutedUntil *time.Time) error
	TransferOwner(ctx context.Context, roomID, oldOwnerID, newOwnerID int64) error
	CreateInvite(ctx context.Context, roomID, inviterID, inviteeID int64) (*model.RoomInvite, error)
	GetInviteByID(ctx context.Context, id int64) (*model.RoomInvite, error)
	HasPendingInvite(ctx context.Context, roomID, inviteeID int64) (bool, error)
	UpdateInviteStatus(ctx context.Context, id int64, status string) error
	ListPendingInvitesByUser(ctx context.Context, userID int64) ([]*InviteInfo, error)
	AreFriends(ctx context.Context, userA, userB int64) (bool, error)
	CountMembers(ctx context.Context, roomID int64) (int, error)
	ListMemberUsernames(ctx context.Context, roomID int64) (map[string]int64, error)
	CountAll(ctx context.Context) (int64, error)
	ListAllRooms(ctx context.Context, offset, limit int) ([]*RoomWithMemberCount, error)
}

type RoomSettingsRepository interface {
	GetByRoomID(ctx context.Context, roomID int64) (*model.RoomSettings, error)
	CreateDefault(ctx context.Context, roomID int64) error
	Update(ctx context.Context, roomID int64, s *model.RoomSettings) (*model.RoomSettings, error)
	ListPermissions(ctx context.Context, roomID int64) ([]*model.RoomRolePermission, error)
	GetPermission(ctx context.Context, roomID int64, role string) (*model.RoomRolePermission, error)
	UpsertPermissions(ctx context.Context, roomID int64, perms []*model.RoomRolePermission) error
	CreateJoinRequest(ctx context.Context, roomID, userID int64, reason *string) (*model.RoomJoinRequest, error)
	GetJoinRequest(ctx context.Context, id int64) (*model.RoomJoinRequest, error)
	ListJoinRequests(ctx context.Context, roomID int64) ([]*model.RoomJoinRequest, error)
	UpdateJoinRequestStatus(ctx context.Context, id int64, status string, reviewerID int64) error
	CreateMessageEvent(ctx context.Context, roomID int64, userID *int64, eventType string, meta any) error
	CreateAuditLog(ctx context.Context, roomID int64, actorID *int64, action string, before, after any) error
	GetStatsSummary(ctx context.Context, roomID int64, since time.Time) (*RoomStatsSummary, error)
}

type MessageRepository interface {
	CountMessages(ctx context.Context) (int64, error)
	CountDMs(ctx context.Context) (int64, error)
	Create(ctx context.Context, content string, userID, roomID int64, msgType string, mentions []int64, fileSize *int64, mimeType *string, clientMsgID *string, replyToID *int64, forwardMeta *model.ForwardInfo) (*model.Message, error)
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByClientMsgID(ctx context.Context, clientMsgID string) (*model.Message, error)
	ListByRoom(ctx context.Context, roomID int64, beforeID, afterID int64, limit int) ([]*MessageWithUser, error)
	ListThreadByRoom(ctx context.Context, roomID, rootMsgID int64, beforeID, afterID int64, limit int) ([]*MessageWithUser, error)
	SoftDelete(ctx context.Context, msgID int64) error
	GetRoomIDByMsgID(ctx context.Context, msgID int64) (int64, error)
	CreateDM(ctx context.Context, content string, senderID, receiverID int64, msgType string, fileSize *int64, mimeType *string, clientMsgID *string, replyToID *int64, forwardMeta *model.ForwardInfo) (*model.DirectMessage, error)
	GetDMByClientMsgID(ctx context.Context, clientMsgID string) (*model.DirectMessage, error)
	GetDMByID(ctx context.Context, id int64) (*model.DirectMessage, error)
	SoftDeleteDM(ctx context.Context, msgID int64) error
	UpdateContent(ctx context.Context, msgID int64, content string) (time.Time, error)
	UpdateDMContent(ctx context.Context, msgID int64, content string) (time.Time, error)
	ListDMs(ctx context.Context, userA, userB int64, beforeID, afterID int64, limit int) ([]*DMWithUser, error)
	ListConversations(ctx context.Context, userID int64) ([]*Conversation, error)
	PinMessage(ctx context.Context, roomID, msgID, userID int64) error
	UnpinMessage(ctx context.Context, roomID, msgID int64) error
	ListPinnedMessages(ctx context.Context, roomID int64) ([]*MessageWithUser, error)
	SearchMessages(ctx context.Context, userID int64, query string, roomID, senderID int64, from, to *time.Time, limit int) ([]*MessageWithUser, error)
	SoftDeleteBatch(ctx context.Context, roomID int64, msgIDs []int64) ([]int64, error)
	MarkDMRead(ctx context.Context, receiverID, peerID int64, lastReadMsgID int64) (int64, *time.Time, error)
}

type FriendshipRepository interface {
	Create(ctx context.Context, userID, friendID int64) (*model.Friendship, error)
	GetByID(ctx context.Context, id int64) (*model.Friendship, error)
	CheckExisting(ctx context.Context, userA, userB int64) (*model.Friendship, error)
	Accept(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
	ListFriends(ctx context.Context, userID int64) ([]*FriendInfo, error)
	ListPendingRequests(ctx context.Context, userID int64) ([]*FriendRequestInfo, error)
	AreFriends(ctx context.Context, userA, userB int64) (bool, error)
	ListFriendIDs(ctx context.Context, userID int64) ([]int64, error)
	SearchUsers(ctx context.Context, query string, excludeUserID int64, limit int) ([]*model.UserInfo, error)
}

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	GetAll(ctx context.Context) (map[string]string, error)
	GetByPrefix(ctx context.Context, prefix string) (map[string]string, error)
}

type OAuthIdentityRepository interface {
	GetByProviderSubject(ctx context.Context, provider, subject string) (*model.OAuthIdentity, error)
	GetByUserProvider(ctx context.Context, userID int64, provider string) (*model.OAuthIdentity, error)
	Create(ctx context.Context, identity *model.OAuthIdentity) (*model.OAuthIdentity, error)
}

type UserTOTPSecretRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*model.UserTOTPSecret, error)
	Upsert(ctx context.Context, userID int64, secretEnc string, enabled bool) (*model.UserTOTPSecret, error)
	SetEnabled(ctx context.Context, userID int64, enabled bool) error
	UpdateLastUsed(ctx context.Context, userID int64) error
	Delete(ctx context.Context, userID int64) error
}

type RecoveryCodeRepository interface {
	ReplaceCodes(ctx context.Context, userID int64, codeHashes []string) error
	ConsumeCode(ctx context.Context, userID int64, codeHash string) (bool, error)
	CountAvailable(ctx context.Context, userID int64) (int, error)
}

type InviteLinkRepository interface {
	Create(ctx context.Context, roomID, creatorID int64, maxUses int, expiresAt *time.Time) (*model.InviteLink, error)
	GetByCode(ctx context.Context, code string) (*model.InviteLink, error)
	Use(ctx context.Context, code string, userID int64) (*UseInviteLinkResult, error)
	IncrementUses(ctx context.Context, id int64) error
	ListByRoom(ctx context.Context, roomID int64) ([]*model.InviteLink, error)
	Delete(ctx context.Context, id int64) error
}

type MessageReadRepository interface {
	Upsert(ctx context.Context, userID, roomID, lastReadMsgID int64) (*model.MessageRead, error)
	Get(ctx context.Context, userID, roomID int64) (*model.MessageRead, error)
	GetUnreadCount(ctx context.Context, userID, roomID int64, lastReadMsgID int64) (int64, error)
	GetLastReadID(ctx context.Context, userID, roomID int64) (int64, error)
	TouchIfMissing(ctx context.Context, userID, roomID, fallbackMsgID int64) error
	MarkReadAtLatest(ctx context.Context, userID, roomID int64) (*model.MessageRead, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, n *model.Notification) (*model.Notification, error)
	List(ctx context.Context, userID int64, unreadOnly bool, limit int) ([]*model.Notification, error)
	MarkRead(ctx context.Context, userID, id int64) error
	Get(ctx context.Context, userID, id int64) (*model.Notification, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, userID int64, action, entityType string, entityID int64, meta any, ip, userAgent string) error
	ListByUser(ctx context.Context, userID int64, limit int) ([]*model.AuditLog, error)
}

type LoginAttemptRepository interface {
	Get(ctx context.Context, key string) (*model.LoginAttempt, error)
	Upsert(ctx context.Context, key string, attempts int, lockedUntil *time.Time) error
	Reset(ctx context.Context, key string) error
}
