package bootstrap

import (
	"io/fs"
	"log/slog"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/handler"
	"github.com/teamsphere/server/internal/middleware"
	"github.com/teamsphere/server/web"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupRouter(cfg *config.Config, d *deps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Metrics())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.SecurityHeaders())

	r.Use(middleware.LimitBody(8 << 20))
	r.MaxMultipartMemory = int64(cfg.Storage.MaxFileSize) + (1 << 20)

	fileSecret := cfg.Security.FileTokenSecret
	if fileSecret == "" {
		fileSecret = cfg.Security.EncryptionKey
	}
	r.GET("/uploads/*filepath", handler.StaticUploadsWithAuth(cfg.Storage.UploadDir, fileSecret, d.authService))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	frontendFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		slog.Error("failed to get frontend FS", "error", err)
	} else {
		r.NoRoute(handler.SPA(frontendFS))
	}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.Health(d.dbPool, d.redisClient, d.hub))
		v1.GET("/ready", handler.Ready(d.dbPool, d.redisClient))

		auth := v1.Group("/auth")
		{
			auth.POST("/register", middleware.RateLimit(d.rateLimiter, 3, time.Minute), d.authHandler.Register)
			auth.POST("/login", middleware.RateLimit(d.rateLimiter, 5, time.Minute), d.authHandler.Login)
			auth.POST("/logout", middleware.Auth(d.authService), d.authHandler.Logout)
			auth.POST("/refresh", middleware.RateLimit(d.rateLimiter, 10, time.Minute), d.authHandler.RefreshToken)
			auth.POST("/send-code", middleware.RateLimit(d.rateLimiter, 5, time.Minute), d.authHandler.SendCode)
			auth.POST("/verify-email", middleware.RateLimit(d.rateLimiter, 10, time.Minute), d.authHandler.VerifyEmail)
			auth.POST("/password/reset-code", middleware.RateLimit(d.rateLimiter, 5, time.Minute), d.authHandler.SendPasswordResetCode)
			auth.POST("/password/reset", middleware.RateLimit(d.rateLimiter, 10, time.Minute), d.authHandler.ResetPassword)
			auth.GET("/email-required", d.authHandler.EmailRequired)
			auth.GET("/oauth/providers", d.authHandler.OAuthProviders)
			auth.GET("/oauth/:provider/start", d.authHandler.OAuthStart)
			auth.GET("/oauth/:provider/callback", d.authHandler.OAuthCallback)
			auth.POST("/2fa/verify-login", d.authHandler.TOTPVerifyLogin)
			auth.POST("/2fa/setup-required", middleware.RateLimit(d.rateLimiter, 5, time.Minute), d.authHandler.TOTPSetupRequired)
			auth.POST("/2fa/enable-required", middleware.RateLimit(d.rateLimiter, 5, time.Minute), d.authHandler.TOTPEnableRequired)
		}

		authed := v1.Group("")
		authed.Use(middleware.Auth(d.authService))
		{
			authAuthed := authed.Group("/auth")
			{
				authAuthed.GET("/2fa/status", d.authHandler.TOTPStatus)
				authAuthed.POST("/2fa/setup", d.authHandler.TOTPSetup)
				authAuthed.POST("/2fa/enable", d.authHandler.TOTPEnable)
				authAuthed.POST("/2fa/disable", d.authHandler.TOTPDisable)
				authAuthed.GET("/2fa/recovery-codes/status", d.authHandler.TOTPRecoveryStatus)
				authAuthed.POST("/2fa/recovery-codes/regen", d.authHandler.TOTPRecoveryRegen)
				authAuthed.GET("/sessions", d.authHandler.ListSessions)
				authAuthed.POST("/sessions/revoke", d.authHandler.RevokeSession)
				authAuthed.POST("/sessions/revoke-others", d.authHandler.RevokeOtherSessions)
			}

			users := authed.Group("/users")
			{
				users.GET("/me", d.userHandler.GetMe)
				users.GET("/profile/:id", d.userHandler.GetByID)
				users.PUT("/me/profile", d.userHandler.UpdateProfile)
				users.PUT("/me/password", d.userHandler.ChangePassword)
				users.POST("/me/avatar", d.userHandler.UploadAvatar)
				users.DELETE("/me", d.userHandler.DeleteAccount)
			}

			authed.POST("/upload", middleware.RateLimit(d.rateLimiter, 20, time.Minute), d.uploadHandler.Upload)

			authed.GET("/users/search", d.friendHandler.SearchUsers)
			authed.GET("/search/messages", middleware.RateLimit(d.rateLimiter, 10, time.Minute), d.searchHandler.SearchMessages)

			friends := authed.Group("/friends")
			{
				friends.GET("", d.friendHandler.ListFriends)
				friends.POST("/request", d.friendHandler.SendRequest)
				friends.GET("/requests", d.friendHandler.ListPendingRequests)
				friends.PUT("/requests/:id", d.friendHandler.RespondRequest)
				friends.DELETE("/:id", d.friendHandler.DeleteFriend)
			}

			dm := authed.Group("/dm")
			{
				dm.GET("/conversations", d.dmHandler.ListConversations)
				dm.GET("/:user_id/messages", d.dmHandler.ListMessages)
				dm.POST("/:user_id/read", d.dmHandler.MarkRead)
				dm.DELETE("/messages/:msg_id", d.dmHandler.RecallDM)
				dm.PUT("/messages/:msg_id", d.dmHandler.EditDM)
			}

			rooms := authed.Group("/rooms")
			{
				rooms.GET("", d.roomHandler.List)
				rooms.POST("", d.roomHandler.Create)
				rooms.GET("/discover", d.roomHandler.DiscoverAll)
				rooms.POST("/:id/join", d.roomHandler.JoinRoom)
				rooms.GET("/:id/settings", d.roomHandler.GetRoomSettings)
				rooms.PUT("/:id/settings", d.roomHandler.UpdateRoomSettings)
				rooms.GET("/:id/permissions", d.roomHandler.GetRoomPermissions)
				rooms.PUT("/:id/permissions", d.roomHandler.UpdateRoomPermissions)
				rooms.GET("/:id/join-requests", d.roomHandler.ListJoinRequests)
				rooms.POST("/:id/join-requests/:req_id/approve", d.roomHandler.ApproveJoinRequest)
				rooms.POST("/:id/join-requests/:req_id/reject", d.roomHandler.RejectJoinRequest)
				rooms.GET("/:id/stats/summary", d.roomHandler.GetRoomStatsSummary)
				rooms.GET("/invites", d.roomHandler.ListPendingInvites)
				rooms.PUT("/invites/:id", d.roomHandler.RespondInvite)
				rooms.GET("/:id", d.roomHandler.GetByID)
				rooms.PUT("/:id", d.roomHandler.Update)
				rooms.DELETE("/:id", d.roomHandler.Delete)
				rooms.GET("/:id/members", d.roomHandler.ListMembers)
				rooms.GET("/:id/messages", middleware.RateLimit(d.rateLimiter, 60, time.Minute), d.wsHandler.ListRoomMessages)
				rooms.POST("/:id/read", d.roomHandler.MarkRead)
				rooms.GET("/:id/unread-count", d.roomHandler.UnreadCount)
				rooms.GET("/:id/messages/:msg_id/thread", d.roomHandler.ListThreadMessages)
				rooms.DELETE("/:id/messages/batch", d.roomHandler.BatchDeleteRoomMessages)
				rooms.GET("/:id/pinned-messages", d.roomHandler.ListPinnedMessages)
				rooms.POST("/:id/messages/:msg_id/pin", d.roomHandler.PinRoomMessage)
				rooms.DELETE("/:id/messages/:msg_id/pin", d.roomHandler.UnpinRoomMessage)
				rooms.DELETE("/:id/messages/:msg_id", d.roomHandler.RecallRoomMessage)
				rooms.PUT("/:id/messages/:msg_id", d.roomHandler.EditRoomMessage)
				rooms.POST("/:id/invite", d.roomHandler.InviteFriend)
				rooms.PUT("/:id/members/:user_id", d.roomHandler.UpdateMemberRole)
				rooms.DELETE("/:id/members/:user_id", d.roomHandler.KickMember)
				rooms.POST("/:id/members/:user_id/mute", d.roomHandler.MuteMember)
				rooms.DELETE("/:id/members/:user_id/mute", d.roomHandler.UnmuteMember)
				rooms.POST("/:id/leave", d.roomHandler.LeaveRoom)
				rooms.PUT("/:id/transfer", d.roomHandler.TransferOwner)
				rooms.POST("/:id/invite-links", d.inviteLinkHandler.CreateLink)
				rooms.GET("/:id/invite-links", d.inviteLinkHandler.ListLinks)
				rooms.DELETE("/:id/invite-links/:link_id", d.inviteLinkHandler.DeleteLink)
			}

			inviteLinks := authed.Group("/invite-links")
			{
				inviteLinks.GET("/:code", d.inviteLinkHandler.GetLinkInfo)
				inviteLinks.POST("/:code/use", d.inviteLinkHandler.UseLink)
			}

			// Reactions (Phase 3)
			messages := authed.Group("/messages")
			{
				messages.POST("/:msg_id/reactions", d.reactionHandler.AddReaction)
				messages.DELETE("/:msg_id/reactions/*emoji", d.reactionHandler.RemoveReaction)
				messages.GET("/:msg_id/reactions", d.reactionHandler.GetReactions)
				messages.POST("/:msg_id/forward", d.messageHandler.ForwardMessage)
			}

			authed.POST("/ws/ticket", middleware.RateLimit(d.rateLimiter, 30, time.Minute), d.wsHandler.CreateTicket)

			authed.GET("/announcement", d.adminHandler.GetAnnouncement)

			notifications := authed.Group("/notifications")
			{
				notifications.GET("", d.notificationHandler.List)
				notifications.PUT("/:id/read", d.notificationHandler.MarkRead)
			}

			admin := authed.Group("/admin")
			admin.Use(middleware.RequireAdmin())
			admin.Use(middleware.RequireAdmin2FA(d.authService))
			{
				admin.GET("/stats", d.adminHandler.GetStats)
				admin.GET("/users", d.adminHandler.ListUsers)
				admin.PUT("/users/:id/role", d.adminHandler.UpdateUserRole)
				admin.DELETE("/users/:id", d.adminHandler.DeleteUser)
				admin.GET("/rooms", d.adminHandler.ListRooms)
				admin.DELETE("/rooms/:id", d.adminHandler.DeleteRoom)
				admin.GET("/settings", d.adminHandler.GetSettings)
				admin.PUT("/settings", d.adminHandler.UpdateSettings)
				admin.GET("/email", d.adminHandler.GetEmailSettings)
				admin.PUT("/email", d.adminHandler.UpdateEmailSettings)
				admin.GET("/announcement", d.adminHandler.GetAnnouncement)
				admin.POST("/announcement", d.adminHandler.SetAnnouncement)
			}
		}

		v1.GET("/ws", d.wsHandler.Upgrade)
	}

	return r
}
