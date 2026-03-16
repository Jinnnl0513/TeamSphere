package ws

import (
	"log/slog"

	"github.com/teamsphere/server/internal/model"
)

// JoinRoom adds a client to a room's viewer set.
func (h *Hub) JoinRoom(c *Client, roomID int64) {
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][c] = true
	c.addRoom(roomID)

	h.pushOnlineUsers(roomID)
}

// LeaveRoom removes a client from a room's viewer set.
func (h *Hub) LeaveRoom(c *Client, roomID int64) {
	h.removeClientFromRoom(c, roomID)
}

func (h *Hub) removeClientFromRoom(c *Client, roomID int64) {
	if roomClients, ok := h.rooms[roomID]; ok {
		delete(roomClients, c)
		if len(roomClients) == 0 {
			delete(h.rooms, roomID)
		}
	}
	c.removeRoom(roomID)
}

// pushOnlineUsers sends the online_users list for a room to all viewers.
func (h *Hub) pushOnlineUsers(roomID int64) {
	if _, ok := h.rooms[roomID]; !ok {
		return
	}

	var users []model.UserInfo
	var missingIDs []int64

	if memberSet, ok := h.roomMembers[roomID]; ok {
		for uid := range memberSet {
			if h.isUserOnline(uid) {
				if conns, ok := h.userIndex[uid]; ok {
					for c := range conns {
						users = append(users, model.UserInfo{
							ID:        c.UserID,
							Username:  c.Username,
							AvatarURL: c.AvatarURL,
						})
						break
					}
				} else {
					missingIDs = append(missingIDs, uid)
				}
			}
		}
	} else {
		go func(rid int64) {
			members, err := h.roomRepo.ListMembers(h.ctx, rid)
			if err != nil {
				slog.Error("pushOnlineUsers (async): failed to list room members", "error", err, "room_id", rid)
				return
			}
			var uids []int64
			for _, m := range members {
				uids = append(uids, m.UserID)
			}
			h.enqueueAction(&Action{
				Type:   "_room_members_cache_push",
				RoomID: rid,
				Data:   uids,
			})
		}(roomID)
		return
	}

	if len(missingIDs) > 0 {
		usersCopy := append([]model.UserInfo(nil), users...)
		go func(rid int64, uids []int64, localUsers []model.UserInfo) {
			infos, err := h.userRepo.ListUserInfosByIDs(h.ctx, uids)
			if err != nil {
				slog.Error("pushOnlineUsers: failed to fetch user info", "error", err, "room_id", rid)
				return
			}
			allUsers := append(localUsers, infos...)
			h.enqueueAction(&Action{
				Type:   "_online_users_broadcast",
				RoomID: rid,
				Data:   allUsers,
			})
		}(roomID, missingIDs, usersCopy)
		return
	}

	msg := &OnlineUsersMessage{
		RoomID: roomID,
		Users:  users,
	}
	h.broadcastToRoom(roomID, TypeOnlineUsers, msg, nil)
}

// pushOnlineForMemberRooms pushes updated online users to all rooms the user is a member of.
func (h *Hub) pushOnlineForMemberRooms(userID int64) {
	for roomID, memberSet := range h.roomMembers {
		if !memberSet[userID] {
			continue
		}
		if _, ok := h.rooms[roomID]; ok {
			h.pushOnlineUsers(roomID)
		}
	}
}
