package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/teamsphere/server/internal/presence"
	"github.com/teamsphere/server/internal/repository"
	"github.com/coder/websocket"
)

// Hub maintains the set of active clients and routes messages.
// All map access happens in the single Run goroutine.
type Hub struct {
	clients     map[*Client]bool
	userIndex   map[int64]map[*Client]bool
	rooms       map[int64]map[*Client]bool
	roomMembers map[int64]map[int64]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan *broadcastMsg
	direct     chan *directMsg
	action     chan *Action
	done       chan struct{}

	ctx            context.Context
	messageRepo    repository.MessageRepository
	roomRepo       repository.RoomRepository
	settingsRepo   repository.RoomSettingsRepository
	userRepo       repository.UserRepository
	friendshipRepo repository.FriendshipRepository
	readRepo       repository.MessageReadRepository
	notificationRepo repository.NotificationRepository
	presence       presence.Store
}

type broadcastMsg struct {
	sender  *Client
	message *ChatMessage
}

type directMsg struct {
	sender  *Client
	message *DMMessage
}

func NewHub(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, settingsRepo repository.RoomSettingsRepository, userRepo repository.UserRepository, friendshipRepo repository.FriendshipRepository, readRepo repository.MessageReadRepository, notificationRepo repository.NotificationRepository, presenceStore presence.Store) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		userIndex:      make(map[int64]map[*Client]bool),
		rooms:          make(map[int64]map[*Client]bool),
		roomMembers:    make(map[int64]map[int64]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan *broadcastMsg, 256),
		direct:         make(chan *directMsg, 256),
		action:         make(chan *Action, 256),
		done:           make(chan struct{}),
		messageRepo:    messageRepo,
		roomRepo:       roomRepo,
		settingsRepo:   settingsRepo,
		userRepo:       userRepo,
		friendshipRepo: friendshipRepo,
		readRepo:       readRepo,
		notificationRepo: notificationRepo,
		presence:       presenceStore,
	}
}

// Run starts the Hub event loop. Blocks until ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	h.ctx = ctx
	slog.Info("hub started")
	defer func() {
		close(h.done)
		slog.Info("hub stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			for c := range h.clients {
				_ = c.conn.Close(websocket.StatusGoingAway, "server shutting down")
				close(c.send)
			}
			return
		case c := <-h.register:
			h.handleRegister(c)
		case c := <-h.unregister:
			h.handleUnregister(c)
		case msg := <-h.broadcast:
			h.handleBroadcast(ctx, msg)
		case dm := <-h.direct:
			h.handleDirect(ctx, dm)
		case act := <-h.action:
			h.handleAction(act)
		}
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	select {
	case h.unregister <- c:
	case <-h.done:
	}
}

func (h *Hub) Done() <-chan struct{} {
	return h.done
}

func (h *Hub) Broadcast(sender *Client, msg *ChatMessage) {
	h.broadcast <- &broadcastMsg{sender: sender, message: msg}
}

func (h *Hub) Direct(sender *Client, msg *DMMessage) {
	h.direct <- &directMsg{sender: sender, message: msg}
}

func (h *Hub) SendAction(act *Action) {
	h.enqueueAction(act)
}

// OnlineCount returns the number of unique online users.
func (h *Hub) OnlineCount() int {
	if h.presence != nil {
		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		count, err := h.presence.OnlineCount(ctx)
		if err == nil {
			return count
		}
		slog.Error("presence online count failed, falling back to local", "error", err)
	}
	ch := make(chan int, 1)
	select {
	case h.action <- &Action{Type: "_online_count", Data: ch}:
	default:
		return 0
	}
	select {
	case count := <-ch:
		return count
	case <-time.After(2 * time.Second):
		return 0
	}
}
