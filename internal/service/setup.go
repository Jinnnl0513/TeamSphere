package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/database"
	"github.com/teamsphere/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type SetupService struct {
	configPath string
}

func NewSetupService(configPath string) *SetupService {
	return &SetupService{configPath: configPath}
}

// SetupStatus returns whether setup is needed and whether DB is already configured.
type SetupStatus struct {
	Needed       bool `json:"needed"`
	DBConfigured bool `json:"db_configured"`
}

// GetStatus determines the current setup state.
// - No config file: needed=true, db_configured=false
// - Config exists but no users: needed=true, db_configured=true
// - Config exists and users exist: needed=false
func GetStatus(configPath string, pool *pgxpool.Pool) (*SetupStatus, error) {
	if !config.Exists(configPath) {
		return &SetupStatus{Needed: true, DBConfigured: false}, nil
	}
	if pool == nil {
		return &SetupStatus{Needed: true, DBConfigured: false}, nil
	}

	userRepo := repository.NewUserRepo(pool)
	exists, err := userRepo.ExistsAny(context.Background())
	if err != nil {
		return nil, fmt.Errorf("\u68c0\u67e5\u7528\u6237\u72b6\u6001\u5931\u8d25: %w", err)
	}
	return &SetupStatus{Needed: !exists, DBConfigured: true}, nil
}

// TestDBRequest holds the database connection parameters to test.
type TestDBRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

// TestRedisRequest holds Redis connection parameters to test.
type TestRedisRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// TestDB attempts to connect to the given database and returns nil on success.
func (s *SetupService) TestDB(ctx context.Context, req *TestDBRequest) error {
	if req.Port == 0 {
		req.Port = 5432
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		req.User, req.Password, req.Host, req.Port, req.DBName)

	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(testCtx, dsn)
	if err != nil {
		return fmt.Errorf("\u8fde\u63a5\u6570\u636e\u5e93\u5931\u8d25: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(testCtx); err != nil {
		return fmt.Errorf("\u6570\u636e\u5e93\u8fde\u901a\u6027\u6821\u9a8c\u5931\u8d25: %w", err)
	}
	return nil
}

// TestRedis attempts to connect to the given Redis instance and returns nil on success.
func (s *SetupService) TestRedis(ctx context.Context, req *TestRedisRequest) error {
	if req.Port == 0 {
		req.Port = 6379
	}
	if req.DB < 0 {
		req.DB = 0
	}

	addr := fmt.Sprintf("%s:%d", req.Host, req.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: req.Password,
		DB:       req.DB,
	})
	defer client.Close()

	testCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(testCtx).Err(); err != nil {
		return fmt.Errorf("\u8fde\u63a5 Redis \u5931\u8d25: %w", err)
	}
	return nil
}

// TestEmailRequest holds SMTP parameters for sending a test email.
type TestEmailRequest struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from_address"`
	FromName string `json:"from_name"`
	To       string `json:"to"`
}

// TestEmail sends a simple test email via the same SMTP path used at runtime.
func (s *SetupService) TestEmail(ctx context.Context, req *TestEmailRequest) error {
	if req.SMTPPort == 0 {
		req.SMTPPort = 587
	}

	msgBytes, err := buildPlainTextMessage(req.FromName, req.From, req.To, "TeamSphere \u6d4b\u8bd5\u90ae\u4ef6", "\u8fd9\u662f\u4e00\u5c01\u6765\u81ea TeamSphere \u7684\u6d4b\u8bd5\u90ae\u4ef6\u3002")
	if err != nil {
		return fmt.Errorf("\u6784\u9020\u6d4b\u8bd5\u90ae\u4ef6\u5931\u8d25: %w", err)
	}
	if err := sendSMTPMessage(req.SMTPHost, req.SMTPPort, req.Username, req.Password, req.From, []string{req.To}, msgBytes); err != nil {
		return fmt.Errorf("\u53d1\u9001\u6d4b\u8bd5\u90ae\u4ef6\u5931\u8d25: %w", err)
	}
	return nil
}

// CompleteSetupRequest holds all data needed to finalize the setup.
type CompleteSetupRequest struct {
	DB TestDBRequest `json:"db"`

	Email        *EmailConfig      `json:"email,omitempty"`
	EmailEnabled bool              `json:"email_enabled"`
	RedisEnabled bool              `json:"redis_enabled"`
	Redis        *RedisSetupConfig `json:"redis,omitempty"`

	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	AdminEmail    string `json:"admin_email"`

	ServerPort int `json:"server_port,omitempty"`
}

type EmailConfig struct {
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

type RedisSetupConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// CompleteSetupResult is returned after a successful setup.
type CompleteSetupResult struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
}

// CompleteSetup performs the full initialization: verify DB access, migrate, store
// email settings, create the initial owner account, then persist the generated config.
func (s *SetupService) CompleteSetup(ctx context.Context, req *CompleteSetupRequest) (*CompleteSetupResult, error) {
	jwtSecret, err := config.GenerateRandomHex(64)
	if err != nil {
		return nil, fmt.Errorf("\u751f\u6210 JWT \u5bc6\u94a5\u5931\u8d25: %w", err)
	}
	encKey, err := config.GenerateRandomHex(32)
	if err != nil {
		return nil, fmt.Errorf("\u751f\u6210\u52a0\u5bc6\u5bc6\u94a5\u5931\u8d25: %w", err)
	}

	port := 8080
	if req.ServerPort > 0 {
		port = req.ServerPort
	}
	dbPort := req.DB.Port
	if dbPort == 0 {
		dbPort = 5432
	}

	cfg := &config.Config{
		Server: config.ServerConfig{Port: port, Mode: "release"},
		Database: config.DatabaseConfig{
			Host:     req.DB.Host,
			Port:     dbPort,
			User:     req.DB.User,
			Password: req.DB.Password,
			DBName:   req.DB.DBName,
			SSLMode:  "disable",
			MaxConns: 20,
			MinConns: 5,
		},
		Redis: config.RedisConfig{
			Enabled:  req.RedisEnabled,
			Host:     "",
			Port:     0,
			Password: "",
			DB:       0,
		},
		Security: config.SecurityConfig{EncryptionKey: encKey},
		JWT: config.JWTConfig{
			Secret:              jwtSecret,
			ExpireHours:         720,
			AccessExpireMinutes: 15,
			RefreshExpireDays:   7,
		},
		WebSocket: config.WebSocketConfig{MaxMessageSize: 2048, RateLimit: 10, TicketExpireSeconds: 30},
		Storage:   config.StorageConfig{UploadDir: "./uploads", MaxFileSize: 5242880},
		CORS:      config.CORSConfig{AllowedOrigins: config.DefaultAllowedOrigins()},
	}

	// Redis setup
	if req.RedisEnabled {
		if req.Redis != nil {
			cfg.Redis.Host = req.Redis.Host
			cfg.Redis.Port = req.Redis.Port
			cfg.Redis.Password = req.Redis.Password
			cfg.Redis.DB = req.Redis.DB
		}
	}

	pool, err := database.NewPool(ctx, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("\u8fde\u63a5\u6570\u636e\u5e93\u5931\u8d25: %w", err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return nil, fmt.Errorf("\u6267\u884c\u6570\u636e\u5e93\u8fc1\u79fb\u5931\u8d25: %w", err)
	}

	settingsRepo := repository.NewSettingsRepo(pool)
	emailService := NewEmailService(settingsRepo, encKey)
	emailConfigured := req.Email != nil && req.Email.SMTPHost != "" && req.Email.Username != "" && req.Email.Password != "" && req.Email.FromAddress != ""
	if emailConfigured {
		if err := emailService.UpdateSettings(ctx, &EmailSettings{
			Enabled:     true,
			SMTPHost:    req.Email.SMTPHost,
			SMTPPort:    req.Email.SMTPPort,
			Username:    req.Email.Username,
			Password:    req.Email.Password,
			FromAddress: req.Email.FromAddress,
			FromName:    req.Email.FromName,
		}); err != nil {
			return nil, fmt.Errorf("\u4fdd\u5b58\u90ae\u4ef6\u8bbe\u7f6e\u5931\u8d25: %w", err)
		}
	} else {
		if err := settingsRepo.Set(ctx, "email.enabled", "false"); err != nil {
			return nil, fmt.Errorf("\u4fdd\u5b58\u90ae\u4ef6\u5f00\u5173\u5931\u8d25: %w", err)
		}
	}

	requireEmailVerification := "false"
	if emailConfigured {
		requireEmailVerification = "true"
	}
	if err := settingsRepo.Set(ctx, "registration.email_required", requireEmailVerification); err != nil {
		return nil, fmt.Errorf("\u4fdd\u5b58\u6ce8\u518c\u90ae\u7bb1\u9a8c\u8bc1\u5f00\u5173\u5931\u8d25: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("\u751f\u6210\u7ba1\u7406\u5458\u5bc6\u7801\u6458\u8981\u5931\u8d25: %w", err)
	}
	userRepo := repository.NewUserRepo(pool)
	user, err := userRepo.Create(ctx, req.AdminUsername, string(hash), "owner", req.AdminEmail)
	if err != nil {
		return nil, fmt.Errorf("\u521b\u5efa\u7ba1\u7406\u5458\u8d26\u53f7\u5931\u8d25: %w", err)
	}

	authService := NewAuthService(userRepo, nil, nil, nil, nil, &cfg.JWT, nil, cfg.Security.EncryptionKey)
	result, err := authService.Login(ctx, req.AdminUsername, req.AdminPassword, "", "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("\u767b\u5f55\u7ba1\u7406\u5458\u8d26\u53f7\u5931\u8d25: %w", err)
	}

	if err := config.Write(s.configPath, cfg); err != nil {
		if cleanupErr := userRepo.HardDelete(ctx, user.ID); cleanupErr != nil {
			slog.Error("failed to roll back owner after config write failure", "error", cleanupErr, "user_id", user.ID)
		}
		return nil, fmt.Errorf("\u5199\u5165\u914d\u7f6e\u6587\u4ef6\u5931\u8d25: %w", err)
	}
	slog.Info("config written", "path", s.configPath)

	return &CompleteSetupResult{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         user.ToInfo(),
	}, nil
}

// CompleteSetupAdminOnly creates the admin user when config already exists but no users.
func (s *SetupService) CompleteSetupAdminOnly(ctx context.Context, pool *pgxpool.Pool, jwtCfg *config.JWTConfig, username, password, email string) (*CompleteSetupResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("\u751f\u6210\u5bc6\u7801\u6458\u8981\u5931\u8d25: %w", err)
	}

	userRepo := repository.NewUserRepo(pool)
	user, err := userRepo.Create(ctx, username, string(hash), "owner", email)
	if err != nil {
		return nil, fmt.Errorf("\u521b\u5efa\u7ba1\u7406\u5458\u8d26\u53f7\u5931\u8d25: %w", err)
	}

	authService := NewAuthService(userRepo, nil, nil, nil, nil, jwtCfg, nil, "")
	result, err := authService.Login(ctx, username, password, "", "", "", nil)
	if err != nil {
		return nil, fmt.Errorf("\u767b\u5f55\u7ba1\u7406\u5458\u8d26\u53f7\u5931\u8d25: %w", err)
	}

	return &CompleteSetupResult{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		User:         user.ToInfo(),
	}, nil
}
