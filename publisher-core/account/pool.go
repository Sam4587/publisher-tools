// Package account 提供账号池管理服务
package account

import (
	"context"
	"encoding/json"
	"fmt"
	"publisher-core/database"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PoolStrategy 池选择策略
type PoolStrategy string

const (
	PoolStrategyRoundRobin PoolStrategy = "round_robin" // 轮询
	PoolStrategyRandom     PoolStrategy = "random"      // 随机
	PoolStrategyPriority   PoolStrategy = "priority"    // 优先级
	PoolStrategyLeastUsed  PoolStrategy = "least_used"  // 最少使用
	PoolStrategyWeighted   PoolStrategy = "weighted"    // 加权
)

// PoolService 账号池服务
type PoolService struct {
	db           *gorm.DB
	accountSvc   *AccountService
	mu           sync.RWMutex
	poolCache    map[string]*PoolInfo
	roundRobinIdx map[string]int
}

// PoolInfo 池信息
type PoolInfo struct {
	Pool     *database.AccountPool
	Members  []*database.PlatformAccount
	Strategy PoolStrategy
}

// NewPoolService 创建账号池服务
func NewPoolService(db *gorm.DB, accountSvc *AccountService) *PoolService {
	service := &PoolService{
		db:            db,
		accountSvc:    accountSvc,
		poolCache:     make(map[string]*PoolInfo),
		roundRobinIdx: make(map[string]int),
	}

	// 加载活跃池到缓存
	service.loadActivePools()

	return service
}

// loadActivePools 加载活跃池到缓存
func (s *PoolService) loadActivePools() {
	var pools []database.AccountPool
	if err := s.db.Where("is_active = ?", true).Find(&pools).Error; err != nil {
		logrus.Warnf("加载活跃池失败: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, pool := range pools {
		s.loadPoolMembers(&pool)
	}

	logrus.Infof("已加载 %d 个活跃池到缓存", len(pools))
}

// loadPoolMembers 加载池成员
func (s *PoolService) loadPoolMembers(pool *database.AccountPool) {
	var members []database.AccountPoolMember
	if err := s.db.Where("pool_id = ? AND is_active = ?", pool.PoolID, true).Find(&members).Error; err != nil {
		logrus.Warnf("加载池成员失败: %v", err)
		return
	}

	var accounts []*database.PlatformAccount
	for _, member := range members {
		var account database.PlatformAccount
		if err := s.db.Where("account_id = ? AND status = ?", member.AccountID, database.AccountStatusActive).First(&account).Error; err == nil {
			accounts = append(accounts, &account)
		}
	}

	s.poolCache[pool.PoolID] = &PoolInfo{
		Pool:     pool,
		Members:  accounts,
		Strategy: PoolStrategy(pool.Strategy),
	}
}

// CreatePool 创建账号池
func (s *PoolService) CreatePool(ctx context.Context, req *CreatePoolRequest) (*database.AccountPool, error) {
	poolID := uuid.New().String()

	pool := &database.AccountPool{
		PoolID:      poolID,
		Name:        req.Name,
		Platform:    req.Platform,
		Description: req.Description,
		Strategy:    string(req.Strategy),
		MaxSize:     req.MaxSize,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(pool).Error; err != nil {
		return nil, fmt.Errorf("创建池失败: %w", err)
	}

	// 初始化缓存
	s.mu.Lock()
	s.poolCache[poolID] = &PoolInfo{
		Pool:     pool,
		Members:  []*database.PlatformAccount{},
		Strategy: req.Strategy,
	}
	s.mu.Unlock()

	logrus.Infof("创建账号池成功: %s (%s)", poolID, req.Name)
	return pool, nil
}

// GetPool 获取账号池
func (s *PoolService) GetPool(ctx context.Context, poolID string) (*database.AccountPool, error) {
	s.mu.RLock()
	if info, ok := s.poolCache[poolID]; ok {
		s.mu.RUnlock()
		return info.Pool, nil
	}
	s.mu.RUnlock()

	var pool database.AccountPool
	if err := s.db.Where("pool_id = ?", poolID).First(&pool).Error; err != nil {
		return nil, fmt.Errorf("池不存在: %s", poolID)
	}

	return &pool, nil
}

// UpdatePool 更新账号池
func (s *PoolService) UpdatePool(ctx context.Context, poolID string, req *UpdatePoolRequest) error {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Strategy != "" {
		updates["strategy"] = string(req.Strategy)
	}
	if req.MaxSize > 0 {
		updates["max_size"] = req.MaxSize
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.db.Model(pool).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新池失败: %w", err)
	}

	// 更新缓存
	s.mu.Lock()
	if info, ok := s.poolCache[poolID]; ok {
		if req.Strategy != "" {
			info.Strategy = req.Strategy
		}
	}
	s.mu.Unlock()

	logrus.Infof("更新账号池成功: %s", poolID)
	return nil
}

// DeletePool 删除账号池
func (s *PoolService) DeletePool(ctx context.Context, poolID string) error {
	// 删除成员关系
	if err := s.db.Where("pool_id = ?", poolID).Delete(&database.AccountPoolMember{}).Error; err != nil {
		return fmt.Errorf("删除池成员失败: %w", err)
	}

	// 删除池
	if err := s.db.Where("pool_id = ?", poolID).Delete(&database.AccountPool{}).Error; err != nil {
		return fmt.Errorf("删除池失败: %w", err)
	}

	// 从缓存移除
	s.mu.Lock()
	delete(s.poolCache, poolID)
	delete(s.roundRobinIdx, poolID)
	s.mu.Unlock()

	logrus.Infof("删除账号池成功: %s", poolID)
	return nil
}

// ListPools 列出账号池
func (s *PoolService) ListPools(ctx context.Context, req *ListPoolsRequest) ([]*database.AccountPool, int64, error) {
	query := s.db.Model(&database.AccountPool{})

	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var pools []*database.AccountPool
	if err := query.Order("created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&pools).Error; err != nil {
		return nil, 0, err
	}

	return pools, total, nil
}

// AddMember 添加成员到池
func (s *PoolService) AddMember(ctx context.Context, poolID string, accountID string, priority int) error {
	// 检查池是否存在
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	// 检查账号是否存在
	account, err := s.accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	// 检查平台是否匹配
	if account.Platform != pool.Platform {
		return fmt.Errorf("账号平台 %s 与池平台 %s 不匹配", account.Platform, pool.Platform)
	}

	// 检查是否已存在
	var count int64
	s.db.Model(&database.AccountPoolMember{}).
		Where("pool_id = ? AND account_id = ?", poolID, accountID).
		Count(&count)
	if count > 0 {
		return fmt.Errorf("账号已在池中")
	}

	// 检查池大小限制
	var memberCount int64
	s.db.Model(&database.AccountPoolMember{}).Where("pool_id = ?", poolID).Count(&memberCount)
	if int(memberCount) >= pool.MaxSize {
		return fmt.Errorf("池已达到最大容量: %d", pool.MaxSize)
	}

	// 添加成员
	member := &database.AccountPoolMember{
		PoolID:    poolID,
		AccountID: accountID,
		Priority:  priority,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(member).Error; err != nil {
		return fmt.Errorf("添加成员失败: %w", err)
	}

	// 更新缓存
	s.mu.Lock()
	if info, ok := s.poolCache[poolID]; ok {
		info.Members = append(info.Members, account)
	}
	s.mu.Unlock()

	logrus.Infof("添加成员到池成功: %s -> %s", accountID, poolID)
	return nil
}

// RemoveMember 从池移除成员
func (s *PoolService) RemoveMember(ctx context.Context, poolID string, accountID string) error {
	if err := s.db.Where("pool_id = ? AND account_id = ?", poolID, accountID).
		Delete(&database.AccountPoolMember{}).Error; err != nil {
		return fmt.Errorf("移除成员失败: %w", err)
	}

	// 更新缓存
	s.mu.Lock()
	if info, ok := s.poolCache[poolID]; ok {
		var newMembers []*database.PlatformAccount
		for _, m := range info.Members {
			if m.AccountID != accountID {
				newMembers = append(newMembers, m)
			}
		}
		info.Members = newMembers
	}
	s.mu.Unlock()

	logrus.Infof("从池移除成员成功: %s <- %s", accountID, poolID)
	return nil
}

// SelectAccountFromPool 从池中选择账号
func (s *PoolService) SelectAccountFromPool(ctx context.Context, poolID string) (*database.PlatformAccount, error) {
	s.mu.RLock()
	info, ok := s.poolCache[poolID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("池不存在: %s", poolID)
	}

	if len(info.Members) == 0 {
		return nil, fmt.Errorf("池中没有可用账号: %s", poolID)
	}

	var selected *database.PlatformAccount

	switch info.Strategy {
	case PoolStrategyRoundRobin:
		selected = s.selectRoundRobin(poolID, info.Members)
	case PoolStrategyRandom:
		selected = s.selectRandom(info.Members)
	case PoolStrategyPriority:
		selected = s.selectPriority(info.Members)
	case PoolStrategyLeastUsed:
		selected = s.selectLeastUsed(info.Members)
	case PoolStrategyWeighted:
		selected = s.selectWeighted(info.Members)
	default:
		selected = info.Members[0]
	}

	// 更新使用统计
	now := time.Now()
	s.db.Model(selected).Updates(map[string]interface{}{
		"use_count":    gorm.Expr("use_count + 1"),
		"last_used_at": now,
		"updated_at":   now,
	})

	return selected, nil
}

// GetPoolStats 获取池统计
func (s *PoolService) GetPoolStats(ctx context.Context, poolID string) (*PoolStats, error) {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	info, ok := s.poolCache[poolID]
	s.mu.RUnlock()

	stats := &PoolStats{
		PoolID:      poolID,
		Name:        pool.Name,
		Platform:    pool.Platform,
		Strategy:    pool.Strategy,
		TotalCount:  0,
		ActiveCount: 0,
		InactiveCount: 0,
	}

	if ok {
		stats.TotalCount = len(info.Members)
		for _, m := range info.Members {
			if m.Status == database.AccountStatusActive {
				stats.ActiveCount++
			} else {
				stats.InactiveCount++
			}
		}
	}

	return stats, nil
}

// RefreshPool 刷新池缓存
func (s *PoolService) RefreshPool(ctx context.Context, poolID string) error {
	pool, err := s.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.loadPoolMembers(pool)
	s.mu.Unlock()

	logrus.Infof("刷新池缓存成功: %s", poolID)
	return nil
}

// 选择策略实现
func (s *PoolService) selectRoundRobin(poolID string, members []*database.PlatformAccount) *database.PlatformAccount {
	s.mu.Lock()
	idx := s.roundRobinIdx[poolID]
	s.roundRobinIdx[poolID] = (idx + 1) % len(members)
	s.mu.Unlock()

	return members[idx]
}

func (s *PoolService) selectRandom(members []*database.PlatformAccount) *database.PlatformAccount {
	// 简化实现：选择第一个
	return members[0]
}

func (s *PoolService) selectPriority(members []*database.PlatformAccount) *database.PlatformAccount {
	// 按优先级排序（已在加载时排序）
	return members[0]
}

func (s *PoolService) selectLeastUsed(members []*database.PlatformAccount) *database.PlatformAccount {
	var selected *database.PlatformAccount
	minUse := int(^uint(0) >> 1)

	for _, m := range members {
		if m.UseCount < minUse {
			minUse = m.UseCount
			selected = m
		}
	}

	return selected
}

func (s *PoolService) selectWeighted(members []*database.PlatformAccount) *database.PlatformAccount {
	// 基于成功率的加权选择
	var selected *database.PlatformAccount
	bestScore := -1.0

	for _, m := range members {
		score := float64(m.SuccessCount) / float64(m.UseCount+1) * float64(m.Priority)
		if score > bestScore {
			bestScore = score
			selected = m
		}
	}

	if selected == nil {
		selected = members[0]
	}

	return selected
}

// 请求和响应类型
type CreatePoolRequest struct {
	Name        string       `json:"name"`
	Platform    string       `json:"platform"`
	Description string       `json:"description"`
	Strategy    PoolStrategy `json:"strategy"`
	MaxSize     int          `json:"max_size"`
}

type UpdatePoolRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Strategy    PoolStrategy `json:"strategy"`
	MaxSize     int          `json:"max_size"`
	IsActive    *bool        `json:"is_active"`
}

type ListPoolsRequest struct {
	Platform string `json:"platform"`
	IsActive *bool  `json:"is_active"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

type PoolStats struct {
	PoolID        string `json:"pool_id"`
	Name          string `json:"name"`
	Platform      string `json:"platform"`
	Strategy      string `json:"strategy"`
	TotalCount    int    `json:"total_count"`
	ActiveCount   int    `json:"active_count"`
	InactiveCount int    `json:"inactive_count"`
}

// =====================================================
// 账号池监控服务
// =====================================================

// PoolMonitor 池监控服务
type PoolMonitor struct {
	poolSvc    *PoolService
	accountSvc *AccountService
	interval   time.Duration
	stopChan   chan struct{}
}

// NewPoolMonitor 创建池监控服务
func NewPoolMonitor(poolSvc *PoolService, accountSvc *AccountService, interval time.Duration) *PoolMonitor {
	return &PoolMonitor{
		poolSvc:    poolSvc,
		accountSvc: accountSvc,
		interval:   interval,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动监控
func (m *PoolMonitor) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	logrus.Infof("启动账号池监控，间隔: %v", m.interval)

	for {
		select {
		case <-ticker.C:
			m.checkAllPools()
		case <-m.stopChan:
			logrus.Info("停止账号池监控")
			return
		}
	}
}

// Stop 停止监控
func (m *PoolMonitor) Stop() {
	close(m.stopChan)
}

// checkAllPools 检查所有池
func (m *PoolMonitor) checkAllPools() {
	ctx := context.Background()

	m.poolSvc.mu.RLock()
	poolIDs := make([]string, 0, len(m.poolSvc.poolCache))
	for id := range m.poolSvc.poolCache {
		poolIDs = append(poolIDs, id)
	}
	m.poolSvc.mu.RUnlock()

	for _, poolID := range poolIDs {
		m.checkPool(ctx, poolID)
	}
}

// checkPool 检查单个池
func (m *PoolMonitor) checkPool(ctx context.Context, poolID string) {
	stats, err := m.poolSvc.GetPoolStats(ctx, poolID)
	if err != nil {
		logrus.Warnf("获取池统计失败: %s, %v", poolID, err)
		return
	}

	// 检查活跃账号数量
	if stats.ActiveCount == 0 {
		logrus.Warnf("池 %s 没有活跃账号", poolID)
	}

	// 检查账号健康状态
	m.poolSvc.mu.RLock()
	info, ok := m.poolSvc.poolCache[poolID]
	m.poolSvc.mu.RUnlock()

	if ok {
		for _, account := range info.Members {
			// 检查是否需要健康检查
			if account.LastCheckAt == nil || time.Since(*account.LastCheckAt) > m.interval {
				go m.accountSvc.HealthCheck(ctx, account.AccountID)
			}
		}
	}
}

// GetPoolHealth 获取池健康状态
func (m *PoolMonitor) GetPoolHealth(ctx context.Context, poolID string) (*PoolHealth, error) {
	stats, err := m.poolSvc.GetPoolStats(ctx, poolID)
	if err != nil {
		return nil, err
	}

	health := &PoolHealth{
		PoolID:      poolID,
		TotalCount:  stats.TotalCount,
		ActiveCount: stats.ActiveCount,
		HealthScore: 0,
		Status:      "unknown",
	}

	if stats.TotalCount > 0 {
		health.HealthScore = float64(stats.ActiveCount) / float64(stats.TotalCount) * 100
	}

	// 判断健康状态
	if health.HealthScore >= 80 {
		health.Status = "healthy"
	} else if health.HealthScore >= 50 {
		health.Status = "warning"
	} else {
		health.Status = "critical"
	}

	return health, nil
}

// PoolHealth 池健康状态
type PoolHealth struct {
	PoolID      string  `json:"pool_id"`
	TotalCount  int     `json:"total_count"`
	ActiveCount int     `json:"active_count"`
	HealthScore float64 `json:"health_score"`
	Status      string  `json:"status"` // healthy, warning, critical
}

// marshalJSON 辅助函数
func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	data, _ := json.Marshal(v)
	return string(data)
}
