package auth

import (
	"context"
	"errors"
	"time"

	"publisher-core/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户管理服务
type UserService struct {
	db          *gorm.DB
	authService *AuthService
}

// NewUserService 创建用户管理服务
func NewUserService(db *gorm.DB, authService *AuthService) *UserService {
	return &UserService{
		db:          db,
		authService: authService,
	}
}

// RegisterUser 注册新用户
func (s *UserService) RegisterUser(ctx context.Context, req *RegisterRequest) (*database.User, error) {
	// 检查用户名是否已存在
	var count int64
	s.db.Model(&database.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		s.db.Model(&database.User{}).Where("email = ?", req.Email).Count(&count)
		if count > 0 {
			return nil, errors.New("email already exists")
		}
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 生成 API Key
	apiKey, err := s.authService.GenerateAPIKey()
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &database.User{
		UserID:       uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user", // 默认角色
		APIKey:       apiKey,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	logrus.Infof("User registered: %s (%s)", user.UserID, user.Username)
	return user, nil
}

// LoginUser 用户登录
func (s *UserService) LoginUser(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 查找用户
	var user database.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 生成 JWT Token
	token, err := s.authService.GenerateToken(user.UserID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)

	logrus.Infof("User logged in: %s (%s)", user.UserID, user.Username)

	return &LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.authService.config.JWTExpiration),
		User:      &user,
	}, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, userID string) (*database.User, error) {
	var user database.User
	if err := s.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) error {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password_hash"] = string(hashedPassword)
	}

	return s.db.Model(&database.User{}).Where("user_id = ?", userID).Updates(updates).Error
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	result := s.db.Where("user_id = ?", userID).Delete(&database.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	logrus.Infof("User deleted: %s", userID)
	return nil
}

// ListUsers 列出用户（管理员功能）
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*database.User, int64, error) {
	var users []*database.User
	var total int64

	s.db.Model(&database.User{}).Count(&total)

	if err := s.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ChangeUserRole 修改用户角色（管理员功能）
func (s *UserService) ChangeUserRole(ctx context.Context, userID, role string) error {
	// 验证角色
	validRoles := map[string]bool{
		"user":  true,
		"admin": true,
	}
	if !validRoles[role] {
		return errors.New("invalid role")
	}

	return s.db.Model(&database.User{}).Where("user_id = ?", userID).Update("role", role).Error
}

// RegenerateAPIKey 重新生成 API Key
func (s *UserService) RegenerateAPIKey(ctx context.Context, userID string) (string, error) {
	apiKey, err := s.authService.GenerateAPIKey()
	if err != nil {
		return "", err
	}

	if err := s.db.Model(&database.User{}).Where("user_id = ?", userID).Update("api_key", apiKey).Error; err != nil {
		return "", err
	}

	logrus.Infof("API Key regenerated for user: %s", userID)
	return apiKey, nil
}

// DisableUser 禁用用户
func (s *UserService) DisableUser(ctx context.Context, userID string) error {
	return s.db.Model(&database.User{}).Where("user_id = ?", userID).Update("is_active", false).Error
}

// EnableUser 启用用户
func (s *UserService) EnableUser(ctx context.Context, userID string) error {
	return s.db.Model(&database.User{}).Where("user_id = ?", userID).Update("is_active", true).Error
}

// =====================================================
// 请求和响应类型
// =====================================================

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"` // 可以是用户名或邮箱
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	User      *database.User `json:"user"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
