package config

import (
	"fmt"
	"os"
	"strings"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// CORS配置
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials  bool
	MaxAge             int

	// JWT配置
	JWTSecret         string
	JWTExpiration     int // 小时

	// 其他安全配置
	EnableCSRF        bool
	EnableRateLimit   bool
	MaxRequestSize    int64 // 字节
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedOrigins:    []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:    []string{"Content-Type", "Authorization"},
		ExposedHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:            3600,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiration:     24,
		EnableCSRF:        false,
		EnableRateLimit:   true,
		MaxRequestSize:    100 * 1024 * 1024, // 100MB
	}
}

// LoadSecurityConfig 从环境变量加载安全配置
func LoadSecurityConfig() *SecurityConfig {
	config := DefaultSecurityConfig()

	// 从环境变量读取允许的来源
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		config.AllowedOrigins = strings.Split(origins, ",")
	}

	// 从环境变量读取JWT密钥
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.JWTSecret = secret
	} else {
		// 如果没有配置,使用默认的警告日志
		// 注意: 生产环境必须配置JWT_SECRET
	}

	// 从环境变量读取其他配置
	if jwtExp := os.Getenv("JWT_EXPIRATION"); jwtExp != "" {
		if exp := parseDuration(jwtExp); exp > 0 {
			config.JWTExpiration = exp
		}
	}

	if maxRequestSize := os.Getenv("MAX_REQUEST_SIZE"); maxRequestSize != "" {
		if size := parseSize(maxRequestSize); size > 0 {
			config.MaxRequestSize = size
		}
	}

	return config
}

// parseDuration 解析持续时间字符串(小时)
func parseDuration(s string) int {
	// 简单实现,支持数字
	var hours int
	_, err := fmt.Sscanf(s, "%d", &hours)
	if err != nil {
		return 0
	}
	return hours
}

// parseSize 解析大小字符串
func parseSize(s string) int64 {
	// 简单实现,支持数字
	var size int64
	_, err := fmt.Sscanf(s, "%d", &size)
	if err != nil {
		return 0
	}
	return size
}
