// Package account 提供多平台账号管理服务
package account

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"publisher-core/database"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	EncryptionKey string // AES加密密钥（32字节）
}

// AccountService 账号管理服务
type AccountService struct {
	db              *gorm.DB
	encryptionKey   []byte
	mu              sync.RWMutex
	accountCache    map[string]*database.PlatformAccount
	healthCheckers  map[string]HealthChecker
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Check(ctx context.Context, account *database.PlatformAccount) error
}

// NewAccountService 创建账号管理服务
func NewAccountService(db *gorm.DB, config *EncryptionConfig) *AccountService {
	service := &AccountService{
		db:             db,
		encryptionKey:  deriveKey(config.EncryptionKey),
		accountCache:   make(map[string]*database.PlatformAccount),
		healthCheckers: make(map[string]HealthChecker),
	}

	// 加载活跃账号到缓存
	service.loadActiveAccounts()

	return service
}

// deriveKey 从密钥字符串派生32字节密钥
func deriveKey(key string) []byte {
	hash := sha256.Sum256([]byte(key))
	return hash[:]
}

// loadActiveAccounts 加载活跃账号到缓存
func (s *AccountService) loadActiveAccounts() {
	var accounts []database.PlatformAccount
	if err := s.db.Where("status = ?", database.AccountStatusActive).Find(&accounts).Error; err != nil {
		logrus.Warnf("加载活跃账号失败: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, account := range accounts {
		s.accountCache[account.AccountID] = &account
	}

	logrus.Infof("已加载 %d 个活跃账号到缓存", len(accounts))
}

// RegisterHealthChecker 注册健康检查器
func (s *AccountService) RegisterHealthChecker(platform string, checker HealthChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthCheckers[platform] = checker
	logrus.Infof("注册健康检查器: %s", platform)
}

// CreateAccount 创建账号
func (s *AccountService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*database.PlatformAccount, error) {
	// 生成账号ID
	accountID := uuid.New().String()

	// 加密Cookie数据
	encryptedCookie, err := s.encrypt(req.CookieData)
	if err != nil {
		return nil, fmt.Errorf("加密Cookie失败: %w", err)
	}

	// 计算Cookie哈希
	cookieHash := s.hashCookie(req.CookieData)

	account := &database.PlatformAccount{
		AccountID:   accountID,
		Platform:    req.Platform,
		AccountName: req.AccountName,
		AccountType: req.AccountType,
		CookieData:  encryptedCookie,
		CookieHash:  cookieHash,
		Status:      database.AccountStatusPending,
		Priority:    req.Priority,
		UserID:      req.UserID,
		ProjectID:   req.ProjectID,
		Tags:        marshalJSON(req.Tags),
		Metadata:    marshalJSON(req.Metadata),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.ExpiresAt != nil {
		account.ExpiresAt = req.ExpiresAt
	}

	if err := s.db.Create(account).Error; err != nil {
		return nil, fmt.Errorf("创建账号失败: %w", err)
	}

	logrus.Infof("创建账号成功: %s (%s)", accountID, req.Platform)
	return account, nil
}

// GetAccount 获取账号
func (s *AccountService) GetAccount(ctx context.Context, accountID string) (*database.PlatformAccount, error) {
	// 先从缓存获取
	s.mu.RLock()
	if account, ok := s.accountCache[accountID]; ok {
		s.mu.RUnlock()
		return account, nil
	}
	s.mu.RUnlock()

	// 从数据库获取
	var account database.PlatformAccount
	if err := s.db.Where("account_id = ?", accountID).First(&account).Error; err != nil {
		return nil, fmt.Errorf("账号不存在: %s", accountID)
	}

	return &account, nil
}

// GetDecryptedCookies 获取解密后的Cookie
func (s *AccountService) GetDecryptedCookies(ctx context.Context, accountID string) (string, error) {
	account, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return "", err
	}

	decrypted, err := s.decrypt(account.CookieData)
	if err != nil {
		return "", fmt.Errorf("解密Cookie失败: %w", err)
	}

	return decrypted, nil
}

// UpdateAccount 更新账号
func (s *AccountService) UpdateAccount(ctx context.Context, accountID string, req *UpdateAccountRequest) error {
	account, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.AccountName != "" {
		updates["account_name"] = req.AccountName
	}
	if req.Priority > 0 {
		updates["priority"] = req.Priority
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.CookieData != "" {
		encrypted, err := s.encrypt(req.CookieData)
		if err != nil {
			return fmt.Errorf("加密Cookie失败: %w", err)
		}
		updates["cookie_data"] = encrypted
		updates["cookie_hash"] = s.hashCookie(req.CookieData)
	}
	if req.Tags != nil {
		updates["tags"] = marshalJSON(req.Tags)
	}
	if req.Metadata != nil {
		updates["metadata"] = marshalJSON(req.Metadata)
	}

	if err := s.db.Model(account).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新账号失败: %w", err)
	}

	// 更新缓存
	s.mu.Lock()
	delete(s.accountCache, accountID)
	s.mu.Unlock()

	logrus.Infof("更新账号成功: %s", accountID)
	return nil
}

// DeleteAccount 删除账号
func (s *AccountService) DeleteAccount(ctx context.Context, accountID string) error {
	if err := s.db.Where("account_id = ?", accountID).Delete(&database.PlatformAccount{}).Error; err != nil {
		return fmt.Errorf("删除账号失败: %w", err)
	}

	// 从缓存移除
	s.mu.Lock()
	delete(s.accountCache, accountID)
	s.mu.Unlock()

	logrus.Infof("删除账号成功: %s", accountID)
	return nil
}

// ListAccounts 列出账号
func (s *AccountService) ListAccounts(ctx context.Context, req *ListAccountsRequest) ([]*database.PlatformAccount, int64, error) {
	query := s.db.Model(&database.PlatformAccount{})

	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.ProjectID != "" {
		query = query.Where("project_id = ?", req.ProjectID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var accounts []*database.PlatformAccount
	if err := query.Order("priority DESC, created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&accounts).Error; err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

// SelectAccount 选择账号（负载均衡）
func (s *AccountService) SelectAccount(ctx context.Context, platform string, strategy SelectionStrategy) (*database.PlatformAccount, error) {
	query := s.db.Model(&database.PlatformAccount{}).
		Where("platform = ? AND status = ?", platform, database.AccountStatusActive)

	var accounts []database.PlatformAccount
	if err := query.Order("priority DESC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("查询账号失败: %w", err)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("没有可用的账号: %s", platform)
	}

	// 根据策略选择账号
	var selected *database.PlatformAccount
	switch strategy {
	case StrategyRoundRobin:
		selected = s.selectRoundRobin(accounts)
	case StrategyRandom:
		selected = s.selectRandom(accounts)
	case StrategyPriority:
		selected = &accounts[0] // 已按优先级排序
	case StrategyLeastUsed:
		selected = s.selectLeastUsed(accounts)
	default:
		selected = &accounts[0]
	}

	// 更新使用统计
	now := time.Now()
	s.db.Model(selected).Updates(map[string]interface{}{
		"use_count":   gorm.Expr("use_count + 1"),
		"last_used_at": now,
		"updated_at":  now,
	})

	return selected, nil
}

// RecordUsage 记录使用情况
func (s *AccountService) RecordUsage(ctx context.Context, accountID string, req *UsageRecordRequest) error {
	log := &database.AccountUsageLog{
		AccountID:  accountID,
		Action:     req.Action,
		Success:    req.Success,
		Error:      req.Error,
		DurationMs: req.DurationMs,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		TaskID:     req.TaskID,
		Metadata:   marshalJSON(req.Metadata),
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(log).Error; err != nil {
		logrus.Warnf("记录使用日志失败: %v", err)
	}

	// 更新账号统计
	account, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Success {
		updates["success_count"] = gorm.Expr("success_count + 1")
	} else {
		updates["fail_count"] = gorm.Expr("fail_count + 1")
		updates["last_error"] = req.Error
	}

	// 如果连续失败次数过多，标记为不活跃
	if account.FailCount >= 5 && !req.Success {
		updates["status"] = database.AccountStatusInactive
		logrus.Warnf("账号 %s 连续失败次数过多，已标记为不活跃", accountID)
	}

	return s.db.Model(account).Updates(updates).Error
}

// HealthCheck 健康检查
func (s *AccountService) HealthCheck(ctx context.Context, accountID string) error {
	account, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	checker, ok := s.healthCheckers[account.Platform]
	if !ok {
		return fmt.Errorf("未找到平台 %s 的健康检查器", account.Platform)
	}

	err = checker.Check(ctx, account)
	now := time.Now()

	updates := map[string]interface{}{
		"last_check_at": now,
		"updated_at":    now,
	}

	if err != nil {
		updates["status"] = database.AccountStatusInactive
		updates["last_error"] = err.Error()
		logrus.Warnf("账号 %s 健康检查失败: %v", accountID, err)
	} else {
		updates["status"] = database.AccountStatusActive
		logrus.Infof("账号 %s 健康检查通过", accountID)
	}

	return s.db.Model(account).Updates(updates).Error
}

// BatchHealthCheck 批量健康检查
func (s *AccountService) BatchHealthCheck(ctx context.Context, platform string) (int, int, error) {
	query := s.db.Model(&database.PlatformAccount{})
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	var accounts []database.PlatformAccount
	if err := query.Find(&accounts).Error; err != nil {
		return 0, 0, err
	}

	successCount := 0
	failCount := 0

	for _, account := range accounts {
		if err := s.HealthCheck(ctx, account.AccountID); err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	return successCount, failCount, nil
}

// GetAccountStats 获取账号统计
func (s *AccountService) GetAccountStats(ctx context.Context, accountID string) (*AccountStats, error) {
	account, err := s.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	stats := &AccountStats{
		AccountID:    account.AccountID,
		Platform:     account.Platform,
		Status:       string(account.Status),
		UseCount:     account.UseCount,
		SuccessCount: account.SuccessCount,
		FailCount:    account.FailCount,
		SuccessRate:  0,
	}

	if account.UseCount > 0 {
		stats.SuccessRate = float64(account.SuccessCount) / float64(account.UseCount) * 100
	}

	// 获取最近使用记录
	var recentLogs []database.AccountUsageLog
	s.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(10).
		Find(&recentLogs)
	stats.RecentLogs = recentLogs

	return stats, nil
}

// 加密方法
func (s *AccountService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密方法
func (s *AccountService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文太短")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// hashCookie 计算Cookie哈希
func (s *AccountService) hashCookie(cookie string) string {
	hash := sha256.Sum256([]byte(cookie))
	return fmt.Sprintf("%x", hash)
}

// 选择策略实现
func (s *AccountService) selectRoundRobin(accounts []database.PlatformAccount) *database.PlatformAccount {
	// 简单轮询：选择使用次数最少的
	return s.selectLeastUsed(accounts)
}

func (s *AccountService) selectRandom(accounts []database.PlatformAccount) *database.PlatformAccount {
	// 随机选择（简化实现）
	return &accounts[0]
}

func (s *AccountService) selectLeastUsed(accounts []database.PlatformAccount) *database.PlatformAccount {
	var selected *database.PlatformAccount
	minUse := int(^uint(0) >> 1) // Max int

	for i := range accounts {
		if accounts[i].UseCount < minUse {
			minUse = accounts[i].UseCount
			selected = &accounts[i]
		}
	}

	return selected
}

// 辅助函数
func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	data, _ := json.Marshal(v)
	return string(data)
}

// 请求和响应类型
type CreateAccountRequest struct {
	Platform    string                 `json:"platform"`
	AccountName string                 `json:"account_name"`
	AccountType string                 `json:"account_type"`
	CookieData  string                 `json:"cookie_data"`
	Priority    int                    `json:"priority"`
	UserID      string                 `json:"user_id"`
	ProjectID   string                 `json:"project_id"`
	ExpiresAt   *time.Time             `json:"expires_at"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type UpdateAccountRequest struct {
	AccountName string                 `json:"account_name"`
	Priority    int                    `json:"priority"`
	Status      database.AccountStatus `json:"status"`
	CookieData  string                 `json:"cookie_data"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ListAccountsRequest struct {
	Platform  string `json:"platform"`
	Status    string `json:"status"`
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type UsageRecordRequest struct {
	Action     string                 `json:"action"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error"`
	DurationMs int                    `json:"duration_ms"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	TaskID     string                 `json:"task_id"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type AccountStats struct {
	AccountID    string                      `json:"account_id"`
	Platform     string                      `json:"platform"`
	Status       string                      `json:"status"`
	UseCount     int                         `json:"use_count"`
	SuccessCount int                         `json:"success_count"`
	FailCount    int                         `json:"fail_count"`
	SuccessRate  float64                     `json:"success_rate"`
	RecentLogs   []database.AccountUsageLog  `json:"recent_logs"`
}

// SelectionStrategy 选择策略
type SelectionStrategy string

const (
	StrategyRoundRobin SelectionStrategy = "round_robin"
	StrategyRandom     SelectionStrategy = "random"
	StrategyPriority   SelectionStrategy = "priority"
	StrategyLeastUsed  SelectionStrategy = "least_used"
)
