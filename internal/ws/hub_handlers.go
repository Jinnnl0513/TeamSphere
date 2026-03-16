package ws

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/security"
	"github.com/jackc/pgx/v5/pgconn"
)

// mentionRegex matches @username patterns in message content.
var mentionRegex = regexp.MustCompile(`@(\w{3,32})`)

func (h *Hub) handleRegister(c *Client) {
	h.clients[c] = true

	if h.userIndex[c.UserID] == nil {
		h.userIndex[c.UserID] = make(map[*Client]bool)
	}
	h.userIndex[c.UserID][c] = true
	firstLocal := len(h.userIndex[c.UserID]) == 1

	if len(h.userIndex[c.UserID]) == 1 {
		userID := c.UserID
		go func() {
			roomList, err := h.roomRepo.ListByUser(h.ctx, userID)
			if err != nil {
				slog.Error("failed to load room list for cache", "user_id", userID, "error", err)
				return
			}
			roomIDs := make([]int64, 0, len(roomList))
			for _, r := range roomList {
				roomIDs = append(roomIDs, r.ID)
			}
			h.enqueueAction(&Action{
				Type:   "_member_cache_init",
				UserID: userID,
				Data:   roomIDs,
			})
		}()
	}

	userID := c.UserID
	notifyFriends := firstLocal
	if h.presence != nil {
		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		first, err := h.presence.MarkOnline(ctx, userID)
		cancel()
		if err != nil {
			slog.Error("presence mark online failed", "user_id", userID, "error", err)
		} else {
			notifyFriends = first
		}
	}

	go func(client *Client, uid int64, notify bool) {
		friendIDs, _ := h.friendshipRepo.ListFriendIDs(h.ctx, uid)
		h.enqueueAction(&Action{
			Type: "_friend_status_init",
			Data: &friendStatusInitAction{client: client, isRegister: true, friendIDs: friendIDs, userID: uid, notifyFriends: notify},
		})
	}(c, userID, notifyFriends)

	h.pushOnlineForMemberRooms(c.UserID)

	slog.Debug("client registered", "user_id", c.UserID, "username", c.Username, "connections", len(h.userIndex[c.UserID]))
}

func (h *Hub) handleUnregister(c *Client) {
	if _, ok := h.clients[c]; !ok {
		return
	}
	delete(h.clients, c)
	close(c.send)

	localLast := false
	if conns, ok := h.userIndex[c.UserID]; ok {
		delete(conns, c)
		if len(conns) == 0 {
			localLast = true
			delete(h.userIndex, c.UserID)
		}
	}
	userID := c.UserID
	notifyFriends := localLast
	if h.presence != nil {
		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		last, err := h.presence.MarkOffline(ctx, userID)
		cancel()
		if err != nil {
			slog.Error("presence mark offline failed", "user_id", userID, "error", err)
		} else {
			notifyFriends = last
		}
	}
	if notifyFriends {
		go func(uid int64) {
			friendIDs, _ := h.friendshipRepo.ListFriendIDs(h.ctx, uid)
			h.enqueueAction(&Action{
				Type: "_friend_status_init",
				Data: &friendStatusInitAction{client: nil, isRegister: false, friendIDs: friendIDs, userID: uid, notifyFriends: true},
			})
		}(userID)
	}

	for _, roomID := range c.viewingRooms() {
		h.removeClientFromRoom(c, roomID)
	}

	h.pushOnlineForMemberRooms(c.UserID)

	slog.Debug("client unregistered", "user_id", c.UserID, "username", c.Username)
}

func (h *Hub) handleBroadcast(ctx context.Context, bm *broadcastMsg) {
	msg := bm.message
	sender := bm.sender

	msgType := msg.MsgType
	if msgType == "" {
		msgType = "text"
	}

	senderSnap := model.UserInfo{
		ID:        sender.UserID,
		Username:  sender.Username,
		AvatarURL: sender.AvatarURL,
	}
	go func() {
		var mentions []int64
		matches := mentionRegex.FindAllStringSubmatch(msg.Content, -1)
		if len(matches) > 0 {
			usernameMap, err := h.roomRepo.ListMemberUsernames(ctx, msg.RoomID)
			if err != nil {
				slog.Error("failed to list member usernames for mentions", "error", err)
			} else {
				seen := make(map[int64]bool)
				for _, match := range matches {
					username := match[1]
					if username == "all" {
						member, err := h.roomRepo.GetMember(ctx, msg.RoomID, senderSnap.ID)
						if err != nil || member == nil || (member.Role != "admin" && member.Role != "owner") {
							allowed := false
							if sender.Role == "admin" || sender.Role == "owner" {
								allowed = true
							}
							if !allowed && h.settingsRepo != nil && member != nil {
								if perm, _ := h.settingsRepo.GetPermission(ctx, msg.RoomID, member.Role); perm != nil && perm.CanMentionAll {
									allowed = true
								}
							}
							if !allowed {
								h.enqueueAction(&Action{
									Type: "_send_error",
									Data: &sendErrorAction{client: sender, code: ErrCodeMentionAllForbid, message: "@all requires permission"},
								})
								return
							}
						}
						if !seen[0] {
							mentions = append(mentions, 0)
							seen[0] = true
						}
						continue
					}
					if uid, ok := usernameMap[username]; ok && uid != senderSnap.ID && !seen[uid] {
						mentions = append(mentions, uid)
						seen[uid] = true
					}
				}
			}
		}

		safeContent := security.SanitizeMessageContent(msg.Content)
		dbMsg, err := h.messageRepo.Create(ctx, safeContent, senderSnap.ID, msg.RoomID, msgType, mentions, msg.FileSize, msg.MimeType, ptrStr(msg.ClientMsgID), msg.ReplyToID, nil)
		if err != nil {
			if isUniqueViolation(err) {
				if existing, dupErr := h.messageRepo.GetByClientMsgID(ctx, msg.ClientMsgID); dupErr == nil && existing != nil {
					h.enqueueAction(&Action{
						Type: "_chat_ack",
						Data: &chatAckAction{
							sender: sender,
							ack:    &ChatAck{ClientMsgID: msg.ClientMsgID, ID: existing.ID, CreatedAt: existing.CreatedAt},
						},
					})
					return
				}
			}
			slog.Error("failed to persist message", "error", err, "room_id", msg.RoomID)
			h.enqueueAction(&Action{
				Type: "_send_error",
				Data: &sendErrorAction{client: sender, code: "INTERNAL_ERROR", message: "failed to save message"},
			})
			return
		}

		var replyInfo *model.ReplyInfo
		if dbMsg.ReplyToID != nil {
			replyInfo = h.buildReplyInfo(ctx, *dbMsg.ReplyToID)
		}

		chatMsg := &ChatBroadcast{
			ID:          dbMsg.ID,
			ClientMsgID: msg.ClientMsgID,
			Content:     dbMsg.Content,
			MsgType:     dbMsg.MsgType,
			FileSize:    dbMsg.FileSize,
			MimeType:    dbMsg.MimeType,
			ForwardMeta: dbMsg.ForwardMeta,
			User:        senderSnap,
			RoomID:      msg.RoomID,
			Mentions:    dbMsg.Mentions,
			ReplyTo:     replyInfo,
			CreatedAt:   dbMsg.CreatedAt,
		}
		ack := &ChatAck{ClientMsgID: msg.ClientMsgID, ID: dbMsg.ID, CreatedAt: dbMsg.CreatedAt}

		if h.readRepo != nil {
			memberIDs, err := h.roomRepo.ListMemberIDs(ctx, msg.RoomID)
			if err == nil {
				for _, uid := range memberIDs {
					if uid == sender.UserID {
						continue
					}
					lastRead, _ := h.readRepo.GetLastReadID(ctx, uid, msg.RoomID)
					unreadCount, err := h.readRepo.GetUnreadCount(ctx, uid, msg.RoomID, lastRead)
					if err == nil {
						h.enqueueAction(&Action{
							Type:   TypeUnreadUpdate,
							UserID: uid,
							Data: map[string]any{
								"room_id":       msg.RoomID,
								"unread_count":  unreadCount,
							},
						})
					}
				}
			}
		}

		if dbMsg.ReplyToID != nil {
			h.enqueueAction(&Action{
				Type:   TypeThreadReply,
				RoomID: msg.RoomID,
				Data: map[string]any{
					"room_id": msg.RoomID,
					"root_msg_id": *dbMsg.ReplyToID,
					"message": chatMsg,
				},
			})
		}

		h.enqueueAction(&Action{
			Type: "_broadcast_result",
			Data: &broadcastResultAction{
				sender:   sender,
				chatMsg:  chatMsg,
				ack:      ack,
				msg:      dbMsg,
				mentions: dbMsg.Mentions,
			},
		})
	}()
}

func (h *Hub) handleDirect(ctx context.Context, dm *directMsg) {
	msg := dm.message
	sender := dm.sender

	msgType := msg.MsgType
	if msgType == "" {
		msgType = "text"
	}

	senderInfo := model.UserInfo{
		ID:        sender.UserID,
		Username:  sender.Username,
		AvatarURL: sender.AvatarURL,
	}
	targetUserID := msg.TargetUserID
	clientMsgID := msg.ClientMsgID

	go func() {
		safeContent := security.SanitizeMessageContent(msg.Content)
		dbDM, err := h.messageRepo.CreateDM(ctx, safeContent, sender.UserID, targetUserID, msgType, msg.FileSize, msg.MimeType, ptrStr(clientMsgID), msg.ReplyToID, nil)
		if err != nil {
			if isUniqueViolation(err) {
				if existing, dupErr := h.messageRepo.GetDMByClientMsgID(ctx, clientMsgID); dupErr == nil && existing != nil {
					h.enqueueAction(&Action{
						Type: "_chat_ack",
						Data: &chatAckAction{
							sender: sender,
							ack:    &ChatAck{ClientMsgID: clientMsgID, ID: existing.ID, CreatedAt: existing.CreatedAt},
						},
					})
					return
				}
			}
			slog.Error("failed to persist dm", "error", err, "target", targetUserID)
			h.enqueueAction(&Action{
				Type: "_send_error",
				Data: &sendErrorAction{client: sender, code: "INTERNAL_ERROR", message: "failed to save message"},
			})
			return
		}

		var dmReplyInfo *model.ReplyInfo
		if dbDM.ReplyToID != nil {
			dmReplyInfo = h.buildDMReplyInfo(ctx, *dbDM.ReplyToID)
		}

		dmBroadcast := &DMBroadcast{
			ID:          dbDM.ID,
			ClientMsgID: clientMsgID,
			Content:     dbDM.Content,
			MsgType:     dbDM.MsgType,
			FileSize:    dbDM.FileSize,
			MimeType:    dbDM.MimeType,
			ForwardMeta: dbDM.ForwardMeta,
			User:        senderInfo,
			ReplyTo:     dmReplyInfo,
			CreatedAt:   dbDM.CreatedAt,
		}
		ack := &DMSent{
			ClientMsgID: clientMsgID,
			ID:          dbDM.ID,
			CreatedAt:   dbDM.CreatedAt,
			Delivered:   false,
		}

		h.enqueueAction(&Action{
			Type: "_direct_result",
			Data: &directResultAction{
				sender:       sender,
				targetUserID: targetUserID,
				dmBroadcast:  dmBroadcast,
				ack:          ack,
			},
		})
	}()
}

func (h *Hub) enqueueAction(act *Action) bool {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case h.action <- act:
		return true
	case <-h.done:
		return false
	case <-timer.C:
		slog.Warn("hub action queue timeout, dropping action", "type", act.Type)
		return false
	}
}

func (h *Hub) isUserOnline(userID int64) bool {
	if h.presence != nil {
		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()
		ok, err := h.presence.IsOnline(ctx, userID)
		if err == nil {
			return ok
		}
		slog.Error("presence isOnline failed, falling back to local", "error", err, "user_id", userID)
	}
	conns, ok := h.userIndex[userID]
	return ok && len(conns) > 0
}

// isUniqueViolation checks if an error is a PostgreSQL unique constraint violation (code 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
