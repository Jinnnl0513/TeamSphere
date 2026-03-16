package ws

// NotifyMemberJoined updates the roomMembers cache when a user joins a room via REST.
func (h *Hub) NotifyMemberJoined(roomID, userID int64) {
	select {
	case h.action <- &Action{
		Type: "_member_cache_update",
		Data: &memberCacheUpdate{roomID: roomID, userID: userID, joined: true},
	}:
	case <-h.done:
	}
}

// NotifyMemberLeft updates the roomMembers cache when a user leaves or is kicked via REST.
func (h *Hub) NotifyMemberLeft(roomID, userID int64) {
	select {
	case h.action <- &Action{
		Type: "_member_cache_update",
		Data: &memberCacheUpdate{roomID: roomID, userID: userID, joined: false},
	}:
	case <-h.done:
	}
}
