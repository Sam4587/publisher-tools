package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置
type Config struct {
	// 数据库配置
	Database DatabaseConfig `json:"database"`

	// AI 服务配置
	AI AIConfig `json:"ai"`

	// 服务器配置
	Server ServerConfig `json:"server"`

	// 通知配置
	Notify NotifyConfig `json:"notify"`

	// 视频处理配置
	Video VideoConfig `json:"video"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `json:"path"`
}

// AIConfig AI 服务配置
type AIConfig struct {
	DefaultProvider string `json:"default_provider"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	Debug        bool   `json:"debug"`
	CORSOrigin   string `json:"cors_origin"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	DefaultChannel string `json:"default_channel"`
}

// VideoConfig 视频处理配置
type VideoConfig struct {
	OutputDir   string `json:"output_dir"`
	Workers     int    `json:"workers"`
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Path: getEnv("DATABASE_PATH", "./data/publisher.db"),
		},
		AI: AIConfig{
			DefaultProvider: getEnv("AI_DEFAULT_PROVIDER", "openrouter"),
		},
		Server: ServerConfig{
			Port:       getEnvInt("SERVER_PORT", 8080),
			Host:       getEnv("SERVER_HOST", "0.0.0.0"),
			Debug:      getEnvBool("DEBUG_MODE", false),
			CORSOrigin: getEnv("CORS_ORIGIN", "*"),
		},
		Notify: NotifyConfig{
			DefaultChannel: getEnv("NOTIFY_DEFAULT_CHANNEL", "feishu"),
		},
		Video: VideoConfig{
			OutputDir: getEnv("VIDEO_OUTPUT_DIR", "./data/videos"),
			Workers:   getEnvInt("VIDEO_WORKERS", 2),
		},
	}
}

// getEnv 获取环境变量，支持默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lower := strings.ToLower(value)
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return defaultValue
}

// GetAIProviderConfig 获取 AI 提供商配置
func GetAIProviderConfig(provider string) (apiKey, baseURL, model string, err error) {
	keyEnv := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
	baseURLEnv := fmt.Sprintf("%s_BASE_URL", strings.ToUpper(provider))
	modelEnv := fmt.Sprintf("%s_MODEL", strings.ToUpper(provider))

	apiKey = os.Getenv(keyEnv)
	if apiKey == "" {
		return "", "", "", fmt.Errorf("API key not found for provider %s (env: %s)", provider, keyEnv)
	}

	baseURL = os.Getenv(baseURLEnv)
	model = os.Getenv(modelEnv)

	return apiKey, baseURL, model, nil
}

// GetNotifyChannelConfig 获取通知渠道配置
func GetNotifyChannelConfig(channel string) (webhook string, config map[string]string, err error) {
	webhookEnv := fmt.Sprintf("%s_WEBHOOK", strings.ToUpper(channel))
	webhook = os.Getenv(webhookEnv)

	// 特殊处理 Telegram
	if channel == "telegram" {
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID := os.Getenv("TELEGRAM_CHAT_ID")
		if botToken == "" || chatID == "" {
			return "", nil, fmt.Errorf("Telegram configuration incomplete")
		}
		config = map[string]string{
			"bot_token": botToken,
			"chat_id":   chatID,
		}
	}

	return webhook, config, nil
}

// Validate 验证必要配置
func (c *Config) Validate() error {
	// 检查是否有至少一个 AI 提供商配置
	providers := []string{"OPENROUTER", "GROQ", "GOOGLE", "DEEPSEEK", "OLLAMA"}
	hasProvider := false
	for _, p := range providers {
		if os.Getenv(p+"_API_KEY") != "" {
			hasProvider = true
			break
		}
	}
	if !hasProvider {
		return fmt.Errorf("at least one AI provider API key must be configured")
	}

	return nil
}
