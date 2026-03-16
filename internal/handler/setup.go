package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teamsphere/server/internal/config"
	"github.com/teamsphere/server/internal/service"
)

var setupUsernameRegex = regexp.MustCompile(`^\w{3,32}$`)

type SetupHandler struct {
	setupService *service.SetupService
	configPath   string
	pool         *pgxpool.Pool // nil when no config exists
	jwtCfg       *config.JWTConfig
	done         chan struct{} // closed when setup completes, signals server to restart
}

func NewSetupHandler(setupService *service.SetupService, configPath string, pool *pgxpool.Pool, jwtCfg *config.JWTConfig, done chan struct{}) *SetupHandler {
	return &SetupHandler{
		setupService: setupService,
		configPath:   configPath,
		pool:         pool,
		jwtCfg:       jwtCfg,
		done:         done,
	}
}

// Status handles GET /setup/status.
func (h *SetupHandler) Status(c *gin.Context) {
	status, err := service.GetStatus(h.configPath, h.pool)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	Success(c, status)
}

// TestDB handles POST /setup/test-db.
func (h *SetupHandler) TestDB(c *gin.Context) {
	var req service.TestDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "无效的请求体")
		return
	}
	if req.Host == "" || req.User == "" || req.DBName == "" {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	if err := h.setupService.TestDB(c.Request.Context(), &req); err != nil {
		Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	Success(c, gin.H{"connected": true})
}

type testConnectionRequest struct {
	DB           service.TestDBRequest     `json:"db"`
	RedisEnabled bool                      `json:"redis_enabled"`
	Redis        *service.RedisSetupConfig `json:"redis,omitempty"`
}

// TestConnection handles POST /setup/test-connection.
// It always tests database connectivity, and tests Redis when enabled.
func (h *SetupHandler) TestConnection(c *gin.Context) {
	var req testConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "无效的请求体")
		return
	}
	if req.DB.Host == "" || req.DB.User == "" || req.DB.DBName == "" {
		Error(c, http.StatusBadRequest, 40001, "数据库地址、用户名和数据库名为必填项")
		return
	}
	if err := h.setupService.TestDB(c.Request.Context(), &req.DB); err != nil {
		Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	if req.RedisEnabled {
		if req.Redis == nil || strings.TrimSpace(req.Redis.Host) == "" {
			Error(c, http.StatusBadRequest, 40001, "Redis 地址不能为空")
			return
		}
		if err := h.setupService.TestRedis(c.Request.Context(), &service.TestRedisRequest{
			Host:     strings.TrimSpace(req.Redis.Host),
			Port:     req.Redis.Port,
			Password: req.Redis.Password,
			DB:       req.Redis.DB,
		}); err != nil {
			Error(c, http.StatusBadRequest, 40001, err.Error())
			return
		}
	}
	Success(c, gin.H{"db": true, "redis": req.RedisEnabled})
}

// TestEmail handles POST /setup/test-email.
func (h *SetupHandler) TestEmail(c *gin.Context) {
	var req service.TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "无效的请求体")
		return
	}
	if req.SMTPHost == "" || req.From == "" || req.To == "" {
		Error(c, http.StatusBadRequest, 40001, "smtp_host、from_address 和 to 为必填项")
		return
	}

	if err := h.setupService.TestEmail(c.Request.Context(), &req); err != nil {
		Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	Success(c, gin.H{"sent": true})
}

// Complete handles POST /setup.
func (h *SetupHandler) Complete(c *gin.Context) {
	// Check if setup is still needed
	status, err := service.GetStatus(h.configPath, h.pool)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if !status.Needed {
		Error(c, http.StatusForbidden, 40301, "初始化已完成")
		return
	}

	// If DB already configured, only need admin credentials
	if status.DBConfigured {
		var req struct {
			AdminUsername string `json:"admin_username"`
			AdminPassword string `json:"admin_password"`
			AdminEmail    string `json:"admin_email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, 40001, "无效的请求体")
			return
		}

		req.AdminUsername = strings.TrimSpace(req.AdminUsername)
		req.AdminEmail = strings.TrimSpace(req.AdminEmail)
		if req.AdminEmail == "" {
			Error(c, http.StatusBadRequest, 40001, "管理员邮箱不能为空")
			return
		}
		if !emailRegex.MatchString(req.AdminEmail) {
			Error(c, http.StatusBadRequest, 40001, "邮箱格式错误")
			return
		}
		if !service.ValidatePassword(req.AdminPassword) {
			Error(c, http.StatusBadRequest, 40001, "管理员密码需包含大小写字母和数字，长度 8-128 位")
			return
		}
		if err := validateSetupAdminUsername(req.AdminUsername); err != nil {
			Error(c, http.StatusBadRequest, 40001, err.Error())
			return
		}

		result, err := h.setupService.CompleteSetupAdminOnly(c.Request.Context(), h.pool, h.jwtCfg, req.AdminUsername, req.AdminPassword, req.AdminEmail)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, err.Error())
			return
		}
		Success(c, result)
		h.signalDone()
		return
	}

	// Full setup
	var req service.CompleteSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "无效的请求体")
		return
	}
	req.AdminUsername = strings.TrimSpace(req.AdminUsername)
	if req.DB.Host == "" || req.DB.User == "" || req.DB.DBName == "" {
		Error(c, http.StatusBadRequest, 40001, "数据库地址、用户名和数据库名为必填项")
		return
	}

	if req.RedisEnabled {
		if req.Redis == nil {
			Error(c, http.StatusBadRequest, 40001, "请填写 Redis 配置")
			return
		}
		req.Redis.Host = strings.TrimSpace(req.Redis.Host)
		if req.Redis.Host == "" {
			Error(c, http.StatusBadRequest, 40001, "Redis 地址不能为空")
			return
		}
		if req.Redis.Port <= 0 {
			req.Redis.Port = 6379
		}
	}

	if req.AdminEmail == "" {
		Error(c, http.StatusBadRequest, 40001, "管理员邮箱不能为空")
		return
	}
	if !emailRegex.MatchString(req.AdminEmail) {
		Error(c, http.StatusBadRequest, 40001, "邮箱格式错误")
		return
	}
	if !service.ValidatePassword(req.AdminPassword) {
		Error(c, http.StatusBadRequest, 40001, "管理员密码需包含大小写字母和数字，长度 8-128 位")
		return
	}
	if err := validateSetupAdminUsername(req.AdminUsername); err != nil {
		Error(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	result, err := h.setupService.CompleteSetup(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	Success(c, result)
	h.signalDone()
}

// signalDone closes the done channel to tell the server to restart in normal mode.
func (h *SetupHandler) signalDone() {
	select {
	case <-h.done:
		// already closed
	default:
		close(h.done)
	}
}

func validateSetupAdminUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("管理员用户名不能为空")
	}
	if !setupUsernameRegex.MatchString(username) {
		return errors.New("管理员用户名必须为 3-32 个字符，仅支持字母、数字和下划线")
	}
	return nil
}
