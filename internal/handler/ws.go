package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/service"
	"github.com/teamsphere/server/internal/ws"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WSHandler handles WebSocket ticket creation and upgrade.
type WSHandler struct {
	hub            *ws.Hub
	authService    *service.AuthService
	roomRepo       repository.RoomRepository
	messageRepo    repository.MessageRepository
	wsCfg          wsConfig
	allowedOrigins []string // CORS origin whitelist for WS upgrade

	tickets sync.Map // ticket string 鈫?*ticketEntry
}

type wsConfig struct {
	MaxMessageSize      int
	RateLimit           int
	TicketExpireSeconds int
}

type ticketEntry struct {
	UserID    int64
	Username  string
	AvatarURL string
	Role      string
	ExpiresAt time.Time
}

func NewWSHandler(hub *ws.Hub, authService *service.AuthService, roomRepo repository.RoomRepository, messageRepo repository.MessageRepository, maxMsgSize, rateLimit, ticketExpireSec int, allowedOrigins []string) *WSHandler {
	h := &WSHandler{
		hub:         hub,
		authService: authService,
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		wsCfg: wsConfig{
			MaxMessageSize:      maxMsgSize,
			RateLimit:           rateLimit,
			TicketExpireSeconds: ticketExpireSec,
		},
		allowedOrigins: allowedOrigins,
	}
	return h
}

// StartTicketCleanup starts a background goroutine to clean expired tickets every 30 seconds.
func (h *WSHandler) StartTicketCleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("ticket cleanup stopped")
				return
			case <-ticker.C:
				now := time.Now()
				h.tickets.Range(func(key, value any) bool {
					if entry, ok := value.(*ticketEntry); ok {
						if now.After(entry.ExpiresAt) {
							h.tickets.Delete(key)
						}
					}
					return true
				})
			}
		}
	}()
}

// CreateTicket handles POST /ws/ticket 鈥?generates a one-time WS connection ticket.
func (h *WSHandler) CreateTicket(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	avatarURL := ""

	if user, err := h.authService.GetUserByID(c.Request.Context(), userID); err == nil && user != nil {
		avatarURL = user.AvatarURL
	}

	ticket := uuid.NewString()
	expiry := time.Duration(h.wsCfg.TicketExpireSeconds) * time.Second

	h.tickets.Store(ticket, &ticketEntry{
		UserID:    userID,
		Username:  username,
		AvatarURL: avatarURL,
		Role:      role,
		ExpiresAt: time.Now().Add(expiry),
	})

	Success(c, gin.H{"ticket": ticket})
}

// Upgrade handles GET /ws?ticket=xxx 鈥?validates ticket and upgrades to WebSocket.
func (h *WSHandler) Upgrade(c *gin.Context) {
	ticket := c.Query("ticket")
	if ticket == "" {
		Error(c, http.StatusUnauthorized, 40101, "ticket涓嶈兘涓虹┖")
		return
	}

	// Look up and consume ticket (one-time use)
	val, ok := h.tickets.LoadAndDelete(ticket)
	if !ok {
		Error(c, http.StatusUnauthorized, 40101, "鏃犳晥鎴栧凡杩囨湡鐨則icket")
		return
	}
	entry := val.(*ticketEntry)

	// Check expiry
	if time.Now().After(entry.ExpiresAt) {
		Error(c, http.StatusUnauthorized, 40101, "未授权")
		return
	}

	// Determine accepted origins: if "*" is in the list, allow all via pattern
	originPatterns := h.allowedOrigins
	allowAll := false
	for _, o := range h.allowedOrigins {
		if o == "*" {
			allowAll = true
			break
		}
	}
	if allowAll {
		originPatterns = []string{"*"}
	}

	// Upgrade to WebSocket
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: originPatterns,
	})
	if err != nil {
		slog.Error("ws upgrade failed", "error", err)
		return
	}

	client := ws.NewClient(
		h.hub, conn,
		entry.UserID, entry.Username, entry.AvatarURL, entry.Role,
		h.wsCfg.RateLimit, h.wsCfg.MaxMessageSize,
	)

	h.hub.Register(client)

	// Start WritePump in a separate goroutine
	ctx := c.Request.Context()
	go client.WritePump(ctx)

	// Block Upgrade handler with ReadPump to keep context alive
	client.ReadPump(ctx)
}

// 鈹€鈹€鈹€ Room Message History 鈹€鈹€鈹€

// ListRoomMessages handles GET /rooms/:id/messages with cursor pagination.
func (h *WSHandler) ListRoomMessages(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")

	// Verify room membership
	member, err := h.roomRepo.GetMember(c.Request.Context(), roomID, userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if member == nil {
		Error(c, http.StatusForbidden, 40301, "涓嶆槸鎴块棿鎴愬憳")
		return
	}

	// Parse pagination params
	beforeID := parseOptionalInt64(c.Query("before_id"))
	afterID := parseOptionalInt64(c.Query("after_id"))
	limit := parseOptionalInt(c.Query("limit"), 50)

	// Mutual exclusion: before_id and after_id cannot both be set
	if beforeID > 0 && afterID > 0 {
		Error(c, http.StatusBadRequest, 40001, "before_id鍜宎fter_id涓嶈兘鍚屾椂浣跨敤")
		return
	}

	messages, err := h.messageRepo.ListByRoom(c.Request.Context(), roomID, beforeID, afterID, limit)
	if err != nil {
		slog.Error("failed to list messages", "error", err, "room_id", roomID)
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	if messages == nil {
		messages = []*repository.MessageWithUser{}
	}

	Success(c, messages)
}

func parseOptionalInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseOptionalInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
