// Package auth 提供 API 认证和授权功能
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"publisher-core/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// =====================================================
// 认证配置
// =====================================================

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret          string        // JWT 密钥
	JWTExpiration      time.Duration // JWT 过期时间
	APIKeyHeader       string        // API Key 请求头名称
	EnableJWT          bool          // 是否启用 JWT 认证
	EnableAPIKey       bool          // 是否启用 API Key 认证
	SkipAuthPaths      []string      // 跳过认证的路径
}

// DefaultAuthConfig 返回默认认证配置
func DefaultAuthConfig() *AuthConfig {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// 如果没有配置,生成一个警告并使用默认值
		// 注意: 生产环境必须通过环境变量配置JWT_SECRET
		jwtSecret = "default-jwt-secret-please-change-in-production"
		logrus.Warn("JWT_SECRET not configured, using default value. Please set JWT_SECRET environment variable in production!")
	} else {
		// 验证密钥长度
		if len(jwtSecret) < 32 {
			logrus.Warn("JWT_SECRET is too short (minimum 32 characters recommended)")
		}
	}

	return &AuthConfig{
		JWTSecret:     jwtSecret,
		JWTExpiration: 24 * time.Hour,
		APIKeyHeader:  "X-API-Key",
		EnableJWT:     true,
		EnableAPIKey:  true,
		SkipAuthPaths: []string{
			"/api/v1/health",
			"/api/v1/login",
			"/api/v1/register",
		},
	}
}

// =====================================================
// JWT 认证
// =====================================================

// Claims JWT 声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	db     *gorm.DB
	config *AuthConfig
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, config *AuthConfig) *AuthService {
	if config == nil {
		config = DefaultAuthConfig()
	}
	return &AuthService{
		db:     db,
		config: config,
	}
}

// GenerateToken 生成 JWT Token
func (s *AuthService) GenerateToken(userID, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "publisher-core",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken 验证 JWT Token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// =====================================================
// API Key 认证
// =====================================================

// GenerateAPIKey 生成 API Key
func (s *AuthService) GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "pk_" + hex.EncodeToString(bytes), nil
}

// ValidateAPIKey 验证 API Key
func (s *AuthService) ValidateAPIKey(apiKey string) (*database.User, error) {
	if !strings.HasPrefix(apiKey, "pk_") {
		return nil, errors.New("invalid API key format")
	}

	var user database.User
	if err := s.db.Where("api_key = ? AND is_active = ?", apiKey, true).First(&user).Error; err != nil {
		return nil, errors.New("invalid or inactive API key")
	}

	// 更新最后使用时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)

	return &user, nil
}

// =====================================================
// 认证中间件
// =====================================================

// AuthMiddleware 认证中间件
func (s *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查是否跳过认证
		if s.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		var user *database.User
		var err error

		// 尝试 JWT 认证
		if s.config.EnableJWT {
			if token := s.extractJWTToken(r); token != "" {
				claims, err := s.ValidateToken(token)
				if err == nil {
					// 从数据库获取用户信息
					user, err = s.getUserByID(claims.UserID)
					if err == nil {
						// 将用户信息添加到上下文
						ctx := context.WithValue(r.Context(), "user", user)
						ctx = context.WithValue(ctx, "claims", claims)
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
						return
					}
				}
			}
		}

		// 尝试 API Key 认证
		if s.config.EnableAPIKey {
			apiKey := r.Header.Get(s.config.APIKeyHeader)
			if apiKey != "" {
				user, err = s.ValidateAPIKey(apiKey)
				if err == nil {
					// 将用户信息添加到上下文
					ctx := context.WithValue(r.Context(), "user", user)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		// 认证失败（不记录详细错误信息以避免泄露）
		logrus.Warnf("Authentication failed for %s %s", r.Method, r.URL.Path)
		http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`, http.StatusUnauthorized)
	})
}

// shouldSkipAuth 检查是否跳过认证
func (s *AuthService) shouldSkipAuth(path string) bool {
	for _, skipPath := range s.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractJWTToken 从请求中提取 JWT Token
func (s *AuthService) extractJWTToken(r *http.Request) string {
	// 从 Authorization 头提取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 从查询参数提取
	return r.URL.Query().Get("token")
}

// getUserByID 根据 ID 获取用户
func (s *AuthService) getUserByID(userID string) (*database.User, error) {
	var user database.User
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// =====================================================
// 授权检查
// =====================================================

// RequireRole 要求特定角色
func (s *AuthService) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*database.User)
			if !ok {
				http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`, http.StatusUnauthorized)
				return
			}

			// 检查角色
			for _, role := range roles {
				if user.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			logrus.Warnf("Authorization failed for user %s: required roles %v, actual role %s", user.UserID, roles, user.Role)
			http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"Insufficient permissions"}}`, http.StatusForbidden)
		})
	}
}

// RequireOwner 要求资源所有者或管理员
func (s *AuthService) RequireOwner(resourceUserID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*database.User)
			if !ok {
				http.Error(w, `{"success":false,"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`, http.StatusUnauthorized)
				return
			}

			// 管理员可以访问所有资源
			if user.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// 检查是否是资源所有者
			if user.UserID == resourceUserID {
				next.ServeHTTP(w, r)
				return
			}

			logrus.Warnf("Authorization failed for user %s: not owner of resource %s", user.UserID, resourceUserID)
			http.Error(w, `{"success":false,"error":{"code":"FORBIDDEN","message":"Access denied"}}`, http.StatusForbidden)
		})
	}
}

// =====================================================
// 辅助函数
// =====================================================

// generateRandomSecret 生成随机密钥
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) (*database.User, error) {
	user, ok := ctx.Value("user").(*database.User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// GetClaimsFromContext 从上下文获取 JWT 声明
func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value("claims").(*Claims)
	if !ok {
		return nil, errors.New("claims not found in context")
	}
	return claims, nil
}
