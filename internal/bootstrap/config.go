package bootstrap

import (
	"os"
	"strings"
)

const (
	defaultConfigPath = "config.yaml"
	defaultSetupPort  = 8080
)

func resolveConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("TEAMSPHERE_CONFIG_PATH")); path != "" {
		return path
	}
	return defaultConfigPath
}

func setupAllowedOrigins(port int) []string {
	_ = port
	return []string{"*"}
}
