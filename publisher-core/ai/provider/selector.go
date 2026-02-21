package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProviderSelector 提供商选择器
type ProviderSelector struct {
	providers map[ProviderType]Provider
	stats     map[ProviderType]*ProviderStats
	mu        sync.RWMutex
}

// ProviderStats 提供商统计
type ProviderStats struct {
	TotalCalls     int64         `json:"total_calls"`
	SuccessCalls   int64         `json:"success_calls"`
	FailedCalls    int64         `json:"failed_calls"`
	AvgDuration    time.Duration `json:"avg_duration"`
	LastCallTime   time.Time     `json:"last_call_time"`
	LastError      string        `json:"last_error"`
	SuccessRate    float64       `json:"success_rate"`
	QualityScore   float64       `json:"quality_score"`
	Priority       int           `json:"priority"`
	IsHealthy      bool          `json:"is_healthy"`
}

// NewProviderSelector 创建提供商选择器
func NewProviderSelector(providers map[ProviderType]Provider) *ProviderSelector {
	selector := &ProviderSelector{
		providers: providers,
		stats:     make(map[ProviderType]*ProviderStats),
	}

	// 初始化统计
	for pType := range providers {
		selector.stats[pType] = &ProviderStats{
			Priority:     getDefaultPriority(pType),
			IsHealthy:    true,
			QualityScore: 100.0,
		}
	}

	return selector
}

// SelectBest 选择最佳提供商
func (s *ProviderSelector) SelectBest(ctx context.Context, opts *GenerateOptions) (Provider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestProvider Provider
	bestScore := -1.0

	for pType, provider := range s.providers {
		stats, exists := s.stats[pType]
		if !exists || !stats.IsHealthy {
			continue
		}

		// 计算综合评分
		score := s.calculateProviderScore(stats, opts)
		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	if bestProvider == nil {
		return nil, fmt.Errorf("没有可用的提供商")
	}

	return bestProvider, nil
}

// calculateProviderScore 计算提供商评分
func (s *ProviderSelector) calculateProviderScore(stats *ProviderStats, opts *GenerateOptions) float64 {
	score := 100.0

	// 优先级评分
	score += float64(stats.Priority * 10)

	// 成功率评分
	if stats.TotalCalls > 0 {
		score += stats.SuccessRate * 20
	}

	// 响应时间评分
	if stats.AvgDuration > 0 {
		if stats.AvgDuration < 1*time.Second {
			score += 15
		} else if stats.AvgDuration < 3*time.Second {
			score += 10
		} else if stats.AvgDuration > 10*time.Second {
			score -= 10
		}
	}

	// 质量评分
	score += stats.QualityScore * 0.1

	// 最近失败惩罚
	if stats.LastError != "" && time.Since(stats.LastCallTime) < 5*time.Minute {
		score -= 20
	}

	return score
}

// GetBackupProviders 获取备用提供商列表
func (s *ProviderSelector) GetBackupProviders(exclude ProviderType) []ProviderType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var backups []ProviderType
	for pType, stats := range s.stats {
		if pType != exclude && stats.IsHealthy {
			backups = append(backups, pType)
		}
	}

	// 按优先级排序
	sortProvidersByPriority(backups, s.stats)
	return backups
}

// UpdateStats 更新统计
func (s *ProviderSelector) UpdateStats(pType ProviderType, success bool, duration time.Duration, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats, exists := s.stats[pType]
	if !exists {
		return
	}

	stats.TotalCalls++
	stats.LastCallTime = time.Now()

	if success {
		stats.SuccessCalls++
		// 更新平均响应时间
		if stats.AvgDuration == 0 {
			stats.AvgDuration = duration
		} else {
			stats.AvgDuration = (stats.AvgDuration + duration) / 2
		}
	} else {
		stats.FailedCalls++
		if err != nil {
			stats.LastError = err.Error()
		}
	}

	// 计算成功率
	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
	}

	// 更新健康状态
	stats.IsHealthy = stats.SuccessRate > 50 || stats.TotalCalls < 5
}

// UpdateQualityScore 更新质量评分
func (s *ProviderSelector) UpdateQualityScore(pType ProviderType, score float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stats, exists := s.stats[pType]; exists {
		stats.QualityScore = score
	}
}

// GetStats 获取所有统计
func (s *ProviderSelector) GetStats() map[ProviderType]*ProviderStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[ProviderType]*ProviderStats)
	for k, v := range s.stats {
		statsCopy := *v
		result[k] = &statsCopy
	}
	return result
}

// getDefaultPriority 获取默认优先级
func getDefaultPriority(pType ProviderType) int {
	priorities := map[ProviderType]int{
		ProviderOpenRouter: 5,
		ProviderGoogle:     4,
		ProviderGroq:       3,
		ProviderDeepSeek:   3,
		"nvidia":           4,
		"mistral":          4,
	}

	if priority, exists := priorities[pType]; exists {
		return priority
	}
	return 1
}

// sortProvidersByPriority 按优先级排序提供商
func sortProvidersByPriority(providers []ProviderType, stats map[ProviderType]*ProviderStats) {
	// 简单冒泡排序
	for i := 0; i < len(providers); i++ {
		for j := i + 1; j < len(providers); j++ {
			if stats[providers[i]].Priority < stats[providers[j]].Priority {
				providers[i], providers[j] = providers[j], providers[i]
			}
		}
	}
}
