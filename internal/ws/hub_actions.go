package ws

import (
	"encoding/json"
	"log/slog"

	"github.com/teamsphere/server/internal/model"
)

type friendStatusInitAction struct {
	client        *Client
	isRegister    bool
	friendIDs     []int64
	userID        int64
	notifyFriends bool
}

type chatAckAction struct {
	sender *Client
	ack    *ChatAck
}

type sendErrorAction struct {
	client  *Client
	code    string
	message string
}

type broadcastResultAction struct {
	sender   *Client
	chatMsg  *ChatBroadcast
	ack      *ChatAck
	msg      interface{}
	mentions []int64
}

type directResultAction struct {
	sender       *Client
	targetUserID int64
	dmBroadcast  *DMBroadcast
	ack          *DMSent
}

type memberCacheUpdate struct {
	roomID int64
	userID int64
	joined bool
}

func (h *Hub) handleAction(act *Action) {
	switch act.Type {
	case "_join_room":
		if c, ok := act.Data.(*Client); ok {
			h.JoinRoom(c, act.RoomID)
		}
		return
	case "_leave_room":
		if c, ok := act.Data.(*Client); ok {
			h.LeaveRoom(c, act.RoomID)
		}
		return
	case "_typing":
		if ta, ok := act.Data.(*typingAction); ok {
			h.broadcastToRoom(ta.roomID, TypeTyping, &TypingBroadcast{
				RoomID:   ta.roomID,
				UserID:   ta.client.UserID,
				Username: ta.client.Username,
			}, ta.client)
		}
		return
	case "_dm_typing":
		if dta, ok := act.Data.(*dmTypingAction); ok {
			h.sendToUser(act.UserID, TypeDMTyping, &DMTypingBroadcast{
				UserID:   dta.senderID,
				Username: dta.username,
			})
		}
		return
	case "_force_disconnect":
		if userID, ok := act.Data.(int64); ok {
			conns := h.userIndex[userID]
			for c := range conns {
				if _, exists := h.clients[c]; exists {
					delete(h.clients, c)
					for _, roomID := range c.viewingRooms() {
						h.removeClientFromRoom(c, roomID)
					}
					close(c.send)
				}
			}
			delete(h.userIndex, userID)
			go func(uid int64) {
				friendIDs, _ := h.friendshipRepo.ListFriendIDs(h.ctx, uid)
				h.enqueueAction(&Action{
					Type: "_friend_status_init",
					Data: &friendStatusInitAction{client: nil, isRegister: false, friendIDs: friendIDs, userID: uid},
				})
			}(userID)
			h.pushOnlineForMemberRooms(userID)
		}
		return
	case "_online_count":
		if ch, ok := act.Data.(chan int); ok {
			ch <- len(h.userIndex)
		}
		return
	case "_member_cache_init":
		if roomIDs, ok := act.Data.([]int64); ok {
			for _, rid := range roomIDs {
				if h.roomMembers[rid] == nil {
					h.roomMembers[rid] = make(map[int64]bool)
				}
				h.roomMembers[rid][act.UserID] = true
			}
		}
		return
	case "_member_cache_update":
		if upd, ok := act.Data.(*memberCacheUpdate); ok {
			if upd.joined {
				if h.roomMembers[upd.roomID] == nil {
					h.roomMembers[upd.roomID] = make(map[int64]bool)
				}
				h.roomMembers[upd.roomID][upd.userID] = true
			} else {
				if set := h.roomMembers[upd.roomID]; set != nil {
					delete(set, upd.userID)
					if len(set) == 0 {
						delete(h.roomMembers, upd.roomID)
					}
				}
			}
		}
		return
	case "_friend_status_init":
		if f, ok := act.Data.(*friendStatusInitAction); ok {
			var evType string
			var uid int64
			if f.isRegister {
				var onlineIDs []int64
				for _, fid := range f.friendIDs {
					if h.isUserOnline(fid) {
						onlineIDs = append(onlineIDs, fid)
					}
				}
				if len(onlineIDs) > 0 {
					env, _ := NewEnvelope(TypeFriendsOnlineList, map[string]any{"user_ids": onlineIDs})
					raw, _ := json.Marshal(env)
					select {
					case f.client.send <- raw:
					default:
					}
				}
				if !f.notifyFriends {
					return
				}
				evType = TypeFriendOnline
				uid = f.client.UserID
			} else {
				if !f.notifyFriends {
					return
				}
				evType = TypeFriendOffline
				uid = f.userID
			}

			env, _ := NewEnvelope(evType, map[string]any{"user_id": uid})
			raw, _ := json.Marshal(env)
			for _, fid := range f.friendIDs {
				if conns, ok := h.userIndex[fid]; ok {
					for c := range conns {
						select {
						case c.send <- raw:
						default:
						}
					}
				}
			}
		}
		return
	case "_room_members_cache_push":
		if uids, ok := act.Data.([]int64); ok {
			roomID := act.RoomID
			if h.roomMembers[roomID] == nil {
				h.roomMembers[roomID] = make(map[int64]bool)
			}
			for _, uid := range uids {
				h.roomMembers[roomID][uid] = true
			}
			h.pushOnlineUsers(roomID)
		}
		return
	case "_online_users_broadcast":
		if users, ok := act.Data.([]model.UserInfo); ok {
			msg := &OnlineUsersMessage{
				RoomID: act.RoomID,
				Users:  users,
			}
			h.broadcastToRoom(act.RoomID, TypeOnlineUsers, msg, nil)
		}
		return
	case "_broadcast_result":
		if r, ok := act.Data.(*broadcastResultAction); ok {
			h.broadcastToRoom(r.chatMsg.RoomID, TypeChat, r.chatMsg, nil)
			h.sendToUser(r.sender.UserID, TypeChatAck, r.ack)
			if len(r.mentions) > 0 {
				if dbMsg, ok := r.msg.(*model.Message); ok {
					h.sendMentionNotifications(h.ctx, r.sender, r.chatMsg.RoomID, dbMsg)
				}
			}
		}
		return
	case "_direct_result":
		if r, ok := act.Data.(*directResultAction); ok {
			h.sendToUser(r.targetUserID, TypeDM, r.dmBroadcast)
			if senderConns, ok := h.userIndex[r.sender.UserID]; ok {
				env, err := NewEnvelope(TypeDM, r.dmBroadcast)
				if err == nil {
					raw, _ := json.Marshal(env)
					for c := range senderConns {
						if c == r.sender {
							continue
						}
						select {
						case c.send <- raw:
						default:
						}
					}
				}
			}
			h.sendToClient(r.sender, TypeDMSent, r.ack)
		}
		return
	case "_chat_ack":
		if r, ok := act.Data.(*chatAckAction); ok {
			h.sendToClient(r.sender, TypeChatAck, r.ack)
		}
		return
	case "_send_error":
		if r, ok := act.Data.(*sendErrorAction); ok {
			h.sendToClient(r.client, TypeError, &ErrorMessage{Code: r.code, Message: r.message})
		}
		return
	}

	env, err := NewEnvelope(act.Type, act.Data)
	if err != nil {
		slog.Error("failed to marshal action", "type", act.Type, "error", err)
		return
	}
	raw, err := json.Marshal(env)
	if err != nil {
		slog.Error("failed to marshal envelope", "type", act.Type, "error", err)
		return
	}

	if act.UserID != 0 {
		if conns, ok := h.userIndex[act.UserID]; ok {
			for c := range conns {
				select {
				case c.send <- raw:
				default:
				}
			}
		}
		return
	}

	if act.RoomID != 0 {
		if roomClients, ok := h.rooms[act.RoomID]; ok {
			for c := range roomClients {
				select {
				case c.send <- raw:
				default:
				}
			}
		}
	}
}
