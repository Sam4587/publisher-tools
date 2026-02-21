package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	providers map[ProviderType]Provider
	stats     map[ProviderType]*ProviderStats
	interval  time.Duration
	stopChan  chan struct{}
	mu        sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(providers map[ProviderType]Provider) *HealthChecker {
	checker := &HealthChecker{
		providers: providers,
		stats:     make(map[ProviderType]*ProviderStats),
		interval:  30 * time.Second,
		stopChan:  make(chan struct{}),
	}

	// 初始化统计
	for pType := range providers {
		checker.stats[pType] = &ProviderStats{
			IsHealthy:    true,
			QualityScore: 100.0,
		}
	}

	return checker
}

// Start 启动健康检查
func (h *HealthChecker) Start() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.checkAll()
		case <-h.stopChan:
			return
		}
	}
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	close(h.stopChan)
}

// checkAll 检查所有提供商
func (h *HealthChecker) checkAll() {
	var wg sync.WaitGroup

	for pType, provider := range h.providers {
		wg.Add(1)
		go func(pt ProviderType, p Provider) {
			defer wg.Done()
			h.checkProvider(pt, p)
		}(pType, provider)
	}

	wg.Wait()
}

// checkProvider 检查单个提供商
func (h *HealthChecker) checkProvider(pType ProviderType, provider Provider) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 发送测试请求
	startTime := time.Now()
	_, err := provider.Generate(ctx, &GenerateOptions{
		Model: provider.DefaultModel(),
		Messages: []Message{
			{Role: RoleUser, Content: "health check"},
		},
		MaxTokens: 10,
	})
	duration := time.Since(startTime)

	h.mu.Lock()
	defer h.mu.Unlock()

	stats, exists := h.stats[pType]
	if !exists {
		return
	}

	stats.LastCallTime = time.Now()
	stats.AvgDuration = duration

	if err != nil {
		stats.IsHealthy = false
		stats.LastError = err.Error()
		stats.FailedCalls++
	} else {
		stats.IsHealthy = true
		stats.SuccessCalls++
		stats.LastError = ""
	}

	stats.TotalCalls++
	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
	}
}

// GetStats 获取健康状态统计
func (h *HealthChecker) GetStats() map[ProviderType]*ProviderStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[ProviderType]*ProviderStats)
	for k, v := range h.stats {
		statsCopy := *v
		result[k] = &statsCopy
	}
	return result
}

// IsHealthy 检查提供商是否健康
func (h *HealthChecker) IsHealthy(pType ProviderType) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if stats, exists := h.stats[pType]; exists {
		return stats.IsHealthy
	}
	return false
}

// MarkUnhealthy 标记提供商为不健康
func (h *HealthChecker) MarkUnhealthy(pType ProviderType, reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if stats, exists := h.stats[pType]; exists {
		stats.IsHealthy = false
		stats.LastError = reason
		stats.LastCallTime = time.Now()
	}
}

// MarkHealthy 标记提供商为健康
func (h *HealthChecker) MarkHealthy(pType ProviderType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if stats, exists := h.stats[pType]; exists {
		stats.IsHealthy = true
		stats.LastError = ""
	}
}

// GetHealthyProviders 获取所有健康的提供商
func (h *HealthChecker) GetHealthyProviders() []ProviderType {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var healthy []ProviderType
	for pType, stats := range h.stats {
		if stats.IsHealthy {
			healthy = append(healthy, pType)
		}
	}
	return healthy
}

// GetUnhealthyProviders 获取所有不健康的提供商
func (h *HealthChecker) GetUnhealthyProviders() []ProviderType {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var unhealthy []ProviderType
	for pType, stats := range h.stats {
		if !stats.IsHealthy {
			unhealthy = append(unhealthy, pType)
		}
	}
	return unhealthy
}

// HealthReport 健康报告
type HealthReport struct {
	Timestamp          time.Time                `json:"timestamp"`
	TotalProviders     int                      `json:"total_providers"`
	HealthyProviders   int                      `json:"healthy_providers"`
	UnhealthyProviders int                      `json:"unhealthy_providers"`
	ProviderStats      map[ProviderType]*ProviderStats `json:"provider_stats"`
}

// GenerateReport 生成健康报告
func (h *HealthChecker) GenerateReport() *HealthReport {
	h.mu.RLock()
	defer h.mu.RUnlock()

	report := &HealthReport{
		Timestamp:     time.Now(),
		TotalProviders: len(h.stats),
		ProviderStats: make(map[ProviderType]*ProviderStats),
	}

	for pType, stats := range h.stats {
		statsCopy := *stats
		report.ProviderStats[pType] = &statsCopy

		if stats.IsHealthy {
			report.HealthyProviders++
		} else {
			report.UnhealthyProviders++
		}
	}

	return report
}

// String 返回健康状态字符串
func (h *HealthChecker) String() string {
	report := h.GenerateReport()
	return fmt.Sprintf(
		"健康检查报告 [%s]: 总计 %d 个提供商, 健康 %d 个, 不健康 %d 个",
		report.Timestamp.Format(time.RFC3339),
		report.TotalProviders,
		report.HealthyProviders,
		report.UnhealthyProviders,
	)
}
