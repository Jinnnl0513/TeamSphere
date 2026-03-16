package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
	"net/url"
	"regexp"

	"github.com/coder/websocket"
)

const (
	writeWait             = 10 * time.Second
	pongWait              = 60 * time.Second
	pingPeriod            = 15 * time.Second
	defaultMaxMessageSize = 4096 // bytes, WS frame limit
	sendBufSize           = 256
)

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn

	// User info (set at connection time from ticket).
	UserID    int64
	Username  string
	AvatarURL string
	Role      string // system role

	send chan []byte

	// rooms this client is currently viewing (managed by Hub goroutine via add/removeRoom).
	mu       sync.Mutex
	roomsSet map[int64]bool

	// Rate limiting: message timestamps within the last second.
	rateMu    sync.Mutex
	rateSlots []time.Time
	rateLimit int // max messages per second

	// Typing rate limit: last typing event time per room.
	typingMu   sync.Mutex
	typingLast map[int64]time.Time
	maxMsgSize int64

	policyMu       sync.Mutex
	lastMsgAt      map[int64]time.Time
	lastMsgContent map[int64]string
	spamSlots      map[int64][]time.Time
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int64, username, avatarURL, role string, rateLimit, maxMsgSize int) *Client {
	if rateLimit <= 0 {
		rateLimit = 10
	}
	if maxMsgSize <= 0 {
		maxMsgSize = defaultMaxMessageSize
	}
	return &Client{
		hub:        hub,
		conn:       conn,
		UserID:     userID,
		Username:   username,
		AvatarURL:  avatarURL,
		Role:       role,
		send:       make(chan []byte, sendBufSize),
		roomsSet:   make(map[int64]bool),
		rateSlots:  make([]time.Time, 0, rateLimit),
		rateLimit:  rateLimit,
		typingLast: make(map[int64]time.Time),
		maxMsgSize: int64(maxMsgSize),
		lastMsgAt:      make(map[int64]time.Time),
		lastMsgContent: make(map[int64]string),
		spamSlots:      make(map[int64][]time.Time),
	}
}

// viewingRooms returns all room IDs this client is currently viewing.
// Called from Hub goroutine.
func (c *Client) viewingRooms() []int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	rooms := make([]int64, 0, len(c.roomsSet))
	for id := range c.roomsSet {
		rooms = append(rooms, id)
	}
	return rooms
}

func (c *Client) addRoom(roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomsSet[roomID] = true
}

func (c *Client) removeRoom(roomID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.roomsSet, roomID)
}

// checkRate returns true if the message is within rate limit.
func (c *Client) checkRate() bool {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Second)

	// Remove expired slots
	valid := c.rateSlots[:0]
	for _, t := range c.rateSlots {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	c.rateSlots = valid

	if len(c.rateSlots) >= c.rateLimit {
		return false
	}
	c.rateSlots = append(c.rateSlots, now)
	return true
}

// checkTypingRate returns true if typing event is allowed (max 1 per 3 seconds per room).
func (c *Client) checkTypingRate(roomID int64) bool {
	c.typingMu.Lock()
	defer c.typingMu.Unlock()

	now := time.Now()
	if last, ok := c.typingLast[roomID]; ok && now.Sub(last) < 3*time.Second {
		return false
	}
	c.typingLast[roomID] = now
	return true
}

// ReadPump reads messages from the WebSocket connection and dispatches them to the Hub.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	c.conn.SetReadLimit(c.maxMsgSize)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				slog.Debug("ws read closed", "user_id", c.UserID, "status", websocket.CloseStatus(err))
			} else {
				slog.Debug("ws read error", "user_id", c.UserID, "error", err)
			}
			return
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.sendError(ErrCodeInvalidFormat, "invalid JSON message")
			continue
		}

		c.dispatch(ctx, &env)
	}
}

// WritePump writes messages from the send channel to the WebSocket connection.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// Hub closed the channel.
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Debug("ws write error", "user_id", c.UserID, "error", err)
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				slog.Debug("ws ping error", "user_id", c.UserID, "error", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// dispatch routes a client message to the appropriate handler.
func (c *Client) dispatch(ctx context.Context, env *Envelope) {
	switch env.Type {
	case TypeChat:
		c.handleChat(ctx, env)
	case TypeJoinRoom:
		c.handleJoinRoom(ctx, env)
	case TypeLeaveRoom:
		c.handleLeaveRoom(env)
	case TypeTyping:
		c.handleTyping(ctx, env)
	case TypeDM:
		c.handleDM(ctx, env)
	case TypeDMTyping:
		c.handleDMTyping(ctx, env)
	default:
		c.sendError(ErrCodeUnknownType, "unknown message type: "+env.Type)
	}
}

func (c *Client) handleChat(ctx context.Context, env *Envelope) {
	var msg ChatMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid chat data")
		return
	}

	if msg.RoomID <= 0 {
		c.sendError(ErrCodeInvalidFormat, "room_id is required")
		return
	}
	if msg.Content == "" {
		c.sendError(ErrCodeInvalidFormat, "content is required")
		return
	}
	if len([]rune(msg.Content)) > 2000 {
		c.sendError(ErrCodeMessageTooLong, "message exceeds 2000 characters")
		return
	}
	if msg.ClientMsgID == "" {
		c.sendError(ErrCodeInvalidFormat, "client_msg_id is required")
		return
	}

	// Rate limit check
	if !c.checkRate() {
		c.sendError(ErrCodeRateLimited, "message rate limit exceeded")
		return
	}

	// Check mute status
	member, err := c.hub.roomRepo.GetMember(ctx, msg.RoomID, c.UserID)
	if err != nil {
		c.sendError(ErrCodeNotRoomMember, "could not verify membership")
		return
	}
	if member == nil {
		c.sendError(ErrCodeNotRoomMember, "not a room member")
		return
	}
	if member.MutedUntil != nil && member.MutedUntil.After(time.Now()) {
		c.sendError(ErrCodeMuted, "you are muted in this room")
		return
	}

	// Room policy checks
	if c.hub.settingsRepo != nil {
		settings, err := c.hub.settingsRepo.GetByRoomID(ctx, msg.RoomID)
		if err == nil && settings != nil {
			role := member.Role
			perm, _ := c.hub.settingsRepo.GetPermission(ctx, msg.RoomID, role)

			// Read-only / archived
			if (settings.ReadOnly || settings.Archived) && role != "owner" && role != "admin" {
				c.sendError(ErrCodeRoomReadOnly, "room is read-only")
				return
			}
			if perm != nil && !perm.CanSend {
				c.sendError(ErrCodeSendForbidden, "send forbidden by room policy")
				return
			}
			// Upload permission for non-text
			if msg.MsgType != "" && msg.MsgType != "text" {
				if perm != nil && !perm.CanUpload {
					c.sendError(ErrCodeUploadForbidden, "upload forbidden by room policy")
					return
				}
			}
			// Slow mode
			if settings.SlowModeSeconds > 0 && role != "owner" && role != "admin" {
				c.policyMu.Lock()
				last := c.lastMsgAt[msg.RoomID]
				if !last.IsZero() && time.Since(last) < time.Duration(settings.SlowModeSeconds)*time.Second {
					c.policyMu.Unlock()
					c.sendError(ErrCodeSlowMode, "slow mode enabled")
					return
				}
				c.lastMsgAt[msg.RoomID] = time.Now()
				c.policyMu.Unlock()
			}
			// Anti spam rate
			if settings.AntiSpamRate > 0 && settings.AntiSpamWindowSec > 0 {
				c.policyMu.Lock()
				window := time.Duration(settings.AntiSpamWindowSec) * time.Second
				cutoff := time.Now().Add(-window)
				slots := c.spamSlots[msg.RoomID]
				filtered := slots[:0]
				for _, t := range slots {
					if t.After(cutoff) {
						filtered = append(filtered, t)
					}
				}
				if len(filtered) >= settings.AntiSpamRate {
					c.spamSlots[msg.RoomID] = filtered
					c.policyMu.Unlock()
					c.sendError(ErrCodeSpamBlocked, "rate limit exceeded")
					_ = c.hub.settingsRepo.CreateMessageEvent(ctx, msg.RoomID, &c.UserID, "spam_blocked", map[string]any{"type": "rate"})
					return
				}
				filtered = append(filtered, time.Now())
				c.spamSlots[msg.RoomID] = filtered
				c.policyMu.Unlock()
			}
			// Anti repeat
			if settings.AntiRepeat {
				c.policyMu.Lock()
				lastContent := c.lastMsgContent[msg.RoomID]
				if lastContent != "" && lastContent == msg.Content {
					c.policyMu.Unlock()
					c.sendError(ErrCodeSpamBlocked, "repeated message")
					_ = c.hub.settingsRepo.CreateMessageEvent(ctx, msg.RoomID, &c.UserID, "spam_blocked", map[string]any{"type": "repeat"})
					return
				}
				c.lastMsgContent[msg.RoomID] = msg.Content
				c.policyMu.Unlock()
			}
			// Content filter
			if settings.ContentFilterMode == "block_log" && len(settings.BlockedKeywords) > 0 {
				lc := strings.ToLower(msg.Content)
				for _, kw := range settings.BlockedKeywords {
					k := strings.TrimSpace(strings.ToLower(kw))
					if k != "" && strings.Contains(lc, k) {
						c.sendError(ErrCodeContentBlocked, "content blocked")
						_ = c.hub.settingsRepo.CreateMessageEvent(ctx, msg.RoomID, &c.UserID, "blocked_by_filter", map[string]any{"keyword": k})
						return
					}
				}
			}
			// Link domain allow/block
			if len(settings.BlockedLinkDomains) > 0 || len(settings.AllowedLinkDomains) > 0 {
				if blocked := containsBlockedDomain(msg.Content, settings.AllowedLinkDomains, settings.BlockedLinkDomains); blocked {
					c.sendError(ErrCodeLinkBlocked, "link blocked")
					_ = c.hub.settingsRepo.CreateMessageEvent(ctx, msg.RoomID, &c.UserID, "link_blocked", nil)
					return
				}
			}
			// File rules
			if msg.FileSize != nil && settings.MaxFileSizeMB > 0 {
				if *msg.FileSize > int64(settings.MaxFileSizeMB)*1024*1024 {
					c.sendError(ErrCodeUploadForbidden, "file too large by room policy")
					return
				}
			}
			if msg.MimeType != nil && len(settings.AllowedFileTypes) > 0 {
				allowed := false
				for _, t := range settings.AllowedFileTypes {
					if strings.EqualFold(strings.TrimSpace(t), strings.TrimSpace(*msg.MimeType)) {
						allowed = true
						break
					}
				}
				if !allowed {
					c.sendError(ErrCodeUploadForbidden, "file type not allowed by room policy")
					return
				}
			}
		}
	}

	// The Hub will persist and broadcast
	c.hub.Broadcast(c, &msg)
}

func (c *Client) handleJoinRoom(ctx context.Context, env *Envelope) {
	var msg JoinRoomMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid join_room data")
		return
	}

	if msg.RoomID <= 0 {
		c.sendError(ErrCodeInvalidFormat, "room_id is required")
		return
	}

	// Verify room membership via DB
	member, err := c.hub.roomRepo.GetMember(ctx, msg.RoomID, c.UserID)
	if err != nil {
		slog.Error("failed to check room membership", "error", err)
		c.sendError(ErrCodeNotRoomMember, "could not verify membership")
		return
	}
	if member == nil {
		c.sendError(ErrCodeNotRoomMember, "not a room member")
		return
	}

	// Send join_room action to Hub (Hub manages room maps)
	c.hub.SendAction(&Action{
		Type:   "_join_room",
		RoomID: msg.RoomID,
		Data:   c,
	})
}

func (c *Client) handleLeaveRoom(env *Envelope) {
	var msg LeaveRoomMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid leave_room data")
		return
	}

	if msg.RoomID <= 0 {
		c.sendError(ErrCodeInvalidFormat, "room_id is required")
		return
	}

	c.hub.SendAction(&Action{
		Type:   "_leave_room",
		RoomID: msg.RoomID,
		Data:   c,
	})
}

func (c *Client) handleTyping(ctx context.Context, env *Envelope) {
	var msg TypingMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid typing data")
		return
	}

	if msg.RoomID <= 0 {
		return
	}

	member, err := c.hub.roomRepo.GetMember(ctx, msg.RoomID, c.UserID)
	if err != nil {
		slog.Error("failed to check room membership for typing", "error", err, "room_id", msg.RoomID, "user_id", c.UserID)
		return
	}
	if member == nil {
		return
	}
	// Rate limit typing events
	if !c.checkTypingRate(msg.RoomID) {
		return // silently drop
	}

	// Broadcast typing to room (excluding sender)
	c.hub.SendAction(&Action{
		Type:   "_typing",
		RoomID: msg.RoomID,
		Data: &typingAction{
			client: c,
			roomID: msg.RoomID,
		},
	})
}

type typingAction struct {
	client *Client
	roomID int64
}

func (c *Client) handleDM(ctx context.Context, env *Envelope) {
	var msg DMMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid dm data")
		return
	}

	if msg.TargetUserID <= 0 {
		c.sendError(ErrCodeInvalidFormat, "target_user_id is required")
		return
	}
	if msg.Content == "" {
		c.sendError(ErrCodeInvalidFormat, "content is required")
		return
	}
	if len([]rune(msg.Content)) > 2000 {
		c.sendError(ErrCodeMessageTooLong, "message exceeds 2000 characters")
		return
	}
	if msg.ClientMsgID == "" {
		c.sendError(ErrCodeInvalidFormat, "client_msg_id is required")
		return
	}

	// Rate limit
	if !c.checkRate() {
		c.sendError(ErrCodeRateLimited, "message rate limit exceeded")
		return
	}

	// Verify friendship
	friends, err := c.hub.friendshipRepo.AreFriends(ctx, c.UserID, msg.TargetUserID)
	if err != nil {
		slog.Error("failed to check friendship for dm", "error", err)
		c.sendError(ErrCodeNotFriends, "could not verify friendship")
		return
	}
	if !friends {
		c.sendError(ErrCodeNotFriends, "not friends")
		return
	}

	c.hub.Direct(c, &msg)
}

func (c *Client) handleDMTyping(ctx context.Context, env *Envelope) {
	var msg DMTypingMessage
	if err := json.Unmarshal(env.Data, &msg); err != nil {
		c.sendError(ErrCodeInvalidFormat, "invalid dm_typing data")
		return
	}

	if msg.TargetUserID <= 0 {
		return
	}

	friends, err := c.hub.friendshipRepo.AreFriends(ctx, c.UserID, msg.TargetUserID)
	if err != nil {
		slog.Error("failed to check friendship for dm typing", "error", err, "target_user_id", msg.TargetUserID, "user_id", c.UserID)
		return
	}
	if !friends {
		return
	}
	// Rate limit typing: reuse room typing limit with a special key (negative target user id)
	if !c.checkTypingRate(-msg.TargetUserID) {
		return
	}

	c.hub.SendAction(&Action{
		Type:   "_dm_typing",
		UserID: msg.TargetUserID,
		Data: &dmTypingAction{
			senderID: c.UserID,
			username: c.Username,
		},
	})
}

type dmTypingAction struct {
	senderID int64
	username string
}

// sendError sends a WS error message to this client.
func (c *Client) sendError(code, message string) {
	env, err := NewEnvelope(TypeError, &ErrorMessage{Code: code, Message: message})
	if err != nil {
		return
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return
	}
	select {
	case c.send <- raw:
	default:
	}
}

var urlRegex = regexp.MustCompile(`https?://\S+`)

func containsBlockedDomain(content string, allowList, blockList []string) bool {
	matches := urlRegex.FindAllString(content, -1)
	if len(matches) == 0 {
		return false
	}
	allow := normalizeDomains(allowList)
	block := normalizeDomains(blockList)
	for _, raw := range matches {
		u, err := url.Parse(raw)
		if err != nil || u.Hostname() == "" {
			continue
		}
		host := strings.ToLower(u.Hostname())
		if len(block) > 0 {
			for _, d := range block {
				if host == d || strings.HasSuffix(host, "."+d) {
					return true
				}
			}
		}
		if len(allow) > 0 {
			ok := false
			for _, d := range allow {
				if host == d || strings.HasSuffix(host, "."+d) {
					ok = true
					break
				}
			}
			if !ok {
				return true
			}
		}
	}
	return false
}

func normalizeDomains(in []string) []string {
	out := make([]string, 0, len(in))
	for _, d := range in {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			out = append(out, d)
		}
	}
	return out
}
