package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server" yaml:"server"`
	Database  DatabaseConfig  `mapstructure:"database" yaml:"database"`
	Redis     RedisConfig     `mapstructure:"redis" yaml:"redis"`
	Security  SecurityConfig  `mapstructure:"security" yaml:"security"`
	JWT       JWTConfig       `mapstructure:"jwt" yaml:"jwt"`
	WebSocket WebSocketConfig `mapstructure:"websocket" yaml:"websocket"`
	Storage   StorageConfig   `mapstructure:"storage" yaml:"storage"`
	CORS      CORSConfig      `mapstructure:"cors" yaml:"cors"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Mode string `mapstructure:"mode" yaml:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	DBName   string `mapstructure:"dbname" yaml:"dbname"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
	MaxConns int32  `mapstructure:"max_conns" yaml:"max_conns"`
	MinConns int32  `mapstructure:"min_conns" yaml:"min_conns"`
}

type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Host     string `mapstructure:"host" yaml:"host"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"db" yaml:"db"`
}

func (d *DatabaseConfig) DSN() string {
	sslmode := d.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:     "/" + d.DBName,
		RawQuery: "sslmode=" + url.QueryEscape(sslmode),
	}
	return u.String()
}

type SecurityConfig struct {
	EncryptionKey   string `mapstructure:"encryption_key" yaml:"encryption_key"`
	FileTokenSecret string `mapstructure:"file_token_secret" yaml:"file_token_secret"`
}

type JWTConfig struct {
	Secret              string `mapstructure:"secret" yaml:"secret"`
	ExpireHours         int    `mapstructure:"expire_hours" yaml:"expire_hours"` // deprecated, use access_expire_minutes
	AccessExpireMinutes int    `mapstructure:"access_expire_minutes" yaml:"access_expire_minutes"`
	RefreshExpireDays   int    `mapstructure:"refresh_expire_days" yaml:"refresh_expire_days"`
}

type WebSocketConfig struct {
	MaxMessageSize      int `mapstructure:"max_message_size" yaml:"max_message_size"`
	RateLimit           int `mapstructure:"rate_limit" yaml:"rate_limit"`
	TicketExpireSeconds int `mapstructure:"ticket_expire_seconds" yaml:"ticket_expire_seconds"`
}

type StorageConfig struct {
	UploadDir   string `mapstructure:"upload_dir" yaml:"upload_dir"`
	MaxFileSize int64  `mapstructure:"max_file_size" yaml:"max_file_size"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`
}

var defaultAllowedOrigins = []string{
	"http://localhost:5173",
	"http://localhost:4173",
}

func DefaultAllowedOrigins() []string {
	out := make([]string, len(defaultAllowedOrigins))
	copy(out, defaultAllowedOrigins)
	return out
}

// Write serializes the config to YAML and writes it to the given path.
func Write(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

// GenerateRandomHex generates a cryptographically random hex string of n bytes.
func GenerateRandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Exists checks whether the config file exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Load reads config from the YAML file at path and applies env var overrides.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("jwt.expire_hours", 720)         // deprecated
	v.SetDefault("jwt.access_expire_minutes", 15) // 15 minutes for access token
	v.SetDefault("jwt.refresh_expire_days", 7)    // 7 days for refresh token
	v.SetDefault("websocket.max_message_size", 2048)
	v.SetDefault("websocket.rate_limit", 10)
	v.SetDefault("websocket.ticket_expire_seconds", 30)
	v.SetDefault("storage.upload_dir", "./uploads")
	v.SetDefault("storage.max_file_size", 5242880)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("cors.allowed_origins", DefaultAllowedOrigins())

	// Read YAML
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Env var overrides: TEAMSPHERE_DATABASE_HOST => database.host
	v.SetEnvPrefix("TEAMSPHERE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
