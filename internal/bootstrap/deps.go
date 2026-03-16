package bootstrap

import (
	"context"
	"time"

	"github.com/teamsphere/server/internal/config"
	contractratelimit "github.com/teamsphere/server/internal/contract/ratelimit"
	"github.com/teamsphere/server/internal/contract/tx"
	"github.com/teamsphere/server/internal/database"
	"github.com/teamsphere/server/internal/handler"
	redisinfra "github.com/teamsphere/server/internal/infra/redis"
	"github.com/teamsphere/server/internal/presence"
	"github.com/teamsphere/server/internal/ratelimit"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/service"
	"github.com/teamsphere/server/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type deps struct {
	dbPool              *pgxpool.Pool
	userRepo            *repository.UserRepo
	roomRepo            *repository.RoomRepo
	messageRepo         *repository.MessageRepo
	friendshipRepo      *repository.FriendshipRepo
	settingsRepo        *repository.SettingsRepo
	oauthRepo           *repository.OAuthIdentityRepo
	totpRepo            *repository.UserTOTPSecretRepo
	recoveryRepo        *repository.RecoveryCodeRepo
	inviteLinkRepo      *repository.InviteLinkRepo
	reactionRepo        *repository.ReactionRepo
	messageReadRepo     *repository.MessageReadRepo
	notificationRepo    *repository.NotificationRepo
	auditLogRepo        *repository.AuditLogRepo
	loginAttemptRepo    *repository.LoginAttemptRepo
	roomSettingsRepo    *repository.RoomSettingsRepo
	redisClient         *redis.Client
	rateLimiter         contractratelimit.Limiter
	txManager           tx.Manager
	authService         *service.AuthService
	userService         *service.UserService
	uploadService       *service.UploadService
	roomService         *service.RoomService
	friendService       *service.FriendService
	messageService      *service.MessageService
	messageReadService  *service.MessageReadService
	notificationService *service.NotificationService
	auditLogService     *service.AuditLogService
	emailService        *service.EmailService
	adminService        *service.AdminService
	inviteLinkService   *service.InviteLinkService
	roomSettingsService *service.RoomSettingsService
	hub                 *ws.Hub
	authHandler         *handler.AuthHandler
	userHandler         *handler.UserHandler
	uploadHandler       *handler.UploadHandler
	roomHandler         *handler.RoomHandler
	friendHandler       *handler.FriendHandler
	dmHandler           *handler.DMHandler
	wsHandler           *handler.WSHandler
	adminHandler        *handler.AdminHandler
	inviteLinkHandler   *handler.InviteLinkHandler
	reactionHandler     *handler.ReactionHandler
	searchHandler       *handler.SearchHandler
	notificationHandler *handler.NotificationHandler
	messageHandler      *handler.MessageHandler
}

func (d *deps) Close() {
	if d.redisClient != nil {
		_ = d.redisClient.Close()
	}
}

func initDeps(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) (*deps, error) {
	userRepo := repository.NewUserRepo(pool)
	roomRepo := repository.NewRoomRepo(pool)
	messageRepo := repository.NewMessageRepo(pool)
	friendshipRepo := repository.NewFriendshipRepo(pool)
	settingsRepo := repository.NewSettingsRepo(pool)
	oauthRepo := repository.NewOAuthIdentityRepo(pool)
	totpRepo := repository.NewUserTOTPSecretRepo(pool)
	recoveryRepo := repository.NewRecoveryCodeRepo(pool)
	inviteLinkRepo := repository.NewInviteLinkRepo(pool)
	reactionRepo := repository.NewReactionRepo(pool)
	messageReadRepo := repository.NewMessageReadRepo(pool)
	notificationRepo := repository.NewNotificationRepo(pool)
	auditLogRepo := repository.NewAuditLogRepo(pool)
	loginAttemptRepo := repository.NewLoginAttemptRepo(pool)
	roomSettingsRepo := repository.NewRoomSettingsRepo(pool)

	authService := service.NewAuthService(userRepo, oauthRepo, totpRepo, recoveryRepo, settingsRepo, &cfg.JWT, loginAttemptRepo, cfg.Security.EncryptionKey)

	var redisClient *redis.Client
	var presenceStore presence.Store
	var limiter contractratelimit.Limiter
	if cfg.Redis.Enabled {
		client, err := redisinfra.NewClient(ctx, &cfg.Redis)
		if err != nil {
			return nil, err
		}
		redisClient = client
		presenceStore = redisinfra.NewPresenceStore(client)
		limiter = ratelimit.NewRedisLimiter(client, "rl:")
	} else {
		memLimiter := ratelimit.NewMemoryLimiter()
		memLimiter.StartCleanup(ctx, 5*time.Minute)
		limiter = memLimiter
	}

	hub := ws.NewHub(messageRepo, roomRepo, roomSettingsRepo, userRepo, friendshipRepo, messageReadRepo, notificationRepo, presenceStore)

	txManager := database.NewTxManager(pool)

	userService := service.NewUserService(userRepo, roomRepo, authService, hub)
	fileSecret := cfg.Security.FileTokenSecret
	if fileSecret == "" {
		fileSecret = cfg.Security.EncryptionKey
	}
	uploadService := service.NewUploadService(&cfg.Storage, fileSecret)

	roomService := service.NewRoomService(roomRepo, userRepo, roomSettingsRepo, hub, txManager)
	messageService := service.NewMessageService(messageRepo, roomRepo, userRepo, hub)
	messageReadService := service.NewMessageReadService(messageReadRepo, roomRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	friendService := service.NewFriendService(friendshipRepo, userRepo, hub, notificationService)
	auditLogService := service.NewAuditLogService(auditLogRepo)
	emailService := service.NewEmailService(settingsRepo, cfg.Security.EncryptionKey)
	adminService := service.NewAdminService(userRepo, roomRepo, messageRepo, settingsRepo, emailService, hub)
	inviteLinkService := service.NewInviteLinkService(inviteLinkRepo, roomRepo, userRepo, hub)
	roomSettingsService := service.NewRoomSettingsService(roomSettingsRepo, roomRepo, userRepo, hub)

	authHandler := handler.NewAuthHandler(authService)

	authService.SetEmailDeps(settingsRepo, emailService)
	userHandler := handler.NewUserHandler(userService, uploadService)
	uploadHandler := handler.NewUploadHandler(uploadService)
	roomHandler := handler.NewRoomHandler(roomService, messageService, messageReadService, auditLogService, roomSettingsService)
	friendHandler := handler.NewFriendHandler(friendService)
	dmHandler := handler.NewDMHandler(messageRepo, friendshipRepo, messageService)
	wsHandler := handler.NewWSHandler(hub, authService, roomRepo, messageRepo,
		cfg.WebSocket.MaxMessageSize, cfg.WebSocket.RateLimit, cfg.WebSocket.TicketExpireSeconds,
		cfg.CORS.AllowedOrigins)
	adminHandler := handler.NewAdminHandler(adminService)
	inviteLinkHandler := handler.NewInviteLinkHandler(inviteLinkService)
	reactionHandler := handler.NewReactionHandler(reactionRepo, messageRepo, hub)
	searchHandler := handler.NewSearchHandler(messageService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	messageHandler := handler.NewMessageHandler(messageRepo, roomRepo, friendshipRepo, messageService)

	return &deps{
		dbPool:              pool,
		userRepo:            userRepo,
		roomRepo:            roomRepo,
		messageRepo:         messageRepo,
		friendshipRepo:      friendshipRepo,
		settingsRepo:        settingsRepo,
		oauthRepo:           oauthRepo,
		totpRepo:            totpRepo,
		recoveryRepo:        recoveryRepo,
		inviteLinkRepo:      inviteLinkRepo,
		reactionRepo:        reactionRepo,
		messageReadRepo:     messageReadRepo,
		notificationRepo:    notificationRepo,
		auditLogRepo:        auditLogRepo,
		loginAttemptRepo:    loginAttemptRepo,
		roomSettingsRepo:    roomSettingsRepo,
		redisClient:         redisClient,
		rateLimiter:         limiter,
		txManager:           txManager,
		authService:         authService,
		userService:         userService,
		uploadService:       uploadService,
		roomService:         roomService,
		friendService:       friendService,
		messageService:      messageService,
		messageReadService:  messageReadService,
		notificationService: notificationService,
		auditLogService:     auditLogService,
		emailService:        emailService,
		adminService:        adminService,
		inviteLinkService:   inviteLinkService,
		roomSettingsService: roomSettingsService,
		hub:                 hub,
		authHandler:         authHandler,
		userHandler:         userHandler,
		uploadHandler:       uploadHandler,
		roomHandler:         roomHandler,
		friendHandler:       friendHandler,
		dmHandler:           dmHandler,
		wsHandler:           wsHandler,
		adminHandler:        adminHandler,
		inviteLinkHandler:   inviteLinkHandler,
		reactionHandler:     reactionHandler,
		searchHandler:       searchHandler,
		notificationHandler: notificationHandler,
		messageHandler:      messageHandler,
	}, nil
}
