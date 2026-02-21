package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MultiModelService 多模型服务
type MultiModelService struct {
	providers map[ProviderType]Provider
	health    *HealthChecker
	selector  *ProviderSelector
}

// NewMultiModelService 创建多模型服务
func NewMultiModelService(providers map[ProviderType]Provider) *MultiModelService {
	service := &MultiModelService{
		providers: providers,
		health:    NewHealthChecker(providers),
		selector:  NewProviderSelector(providers),
	}
	
	// 启动健康检查
	go service.health.Start()
	
	return service
}

// ParallelCallResult 并行调用结果
type ParallelCallResult struct {
	Provider   ProviderType     `json:"provider"`
	Model      string           `json:"model"`
	Result     *GenerateResult  `json:"result"`
	Error      error            `json:"error"`
	Duration   time.Duration    `json:"duration"`
	QualityScore float64        `json:"quality_score"`
}

// ParallelCall 并行调用多个模型
func (s *MultiModelService) ParallelCall(ctx context.Context, opts *GenerateOptions, providers []ProviderType) ([]*ParallelCallResult, error) {
	if len(providers) == 0 {
		// 默认使用所有可用提供商
		for pType := range s.providers {
			providers = append(providers, pType)
		}
	}

	results := make([]*ParallelCallResult, len(providers))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, providerType := range providers {
		wg.Add(1)
		go func(idx int, pType ProviderType) {
			defer wg.Done()

			provider, exists := s.providers[pType]
			if !exists {
				mu.Lock()
				results[idx] = &ParallelCallResult{
					Provider: pType,
					Error:    fmt.Errorf("提供商不存在: %s", pType),
				}
				mu.Unlock()
				return
			}

			startTime := time.Now()
			result, err := provider.Generate(ctx, opts)
			duration := time.Since(startTime)

			mu.Lock()
			results[idx] = &ParallelCallResult{
				Provider: pType,
				Model:    opts.Model,
				Result:   result,
				Error:    err,
				Duration: duration,
			}
			mu.Unlock()
		}(i, providerType)
	}

	wg.Wait()
	return results, nil
}

// CompareResults 对比多个结果
func (s *MultiModelService) CompareResults(results []*ParallelCallResult) *ParallelCallResult {
	var bestResult *ParallelCallResult
	bestScore := -1.0

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		// 计算质量评分
		score := s.calculateQualityScore(result)
		result.QualityScore = score

		if score > bestScore {
			bestScore = score
			bestResult = result
		}
	}

	return bestResult
}

// calculateQualityScore 计算质量评分
func (s *MultiModelService) calculateQualityScore(result *ParallelCallResult) float64 {
	score := 100.0

	// 响应时间评分（越快越好）
	if result.Duration < 1*time.Second {
		score += 10
	} else if result.Duration < 3*time.Second {
		score += 5
	} else if result.Duration > 10*time.Second {
		score -= 10
	}

	// 内容长度评分
	if result.Result != nil {
		contentLen := len(result.Result.Content)
		if contentLen > 100 && contentLen < 2000 {
			score += 5
		} else if contentLen < 50 {
			score -= 5
		}
	}

	// 提供商优先级评分
	switch result.Provider {
	case ProviderOpenRouter:
		score += 5
	case ProviderGoogle:
		score += 3
	case ProviderGroq:
		score += 2
	}

	return score
}

// SmartCall 智能调用（自动选择最佳提供商）
func (s *MultiModelService) SmartCall(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
	// 选择最佳提供商
	provider, err := s.selector.SelectBest(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("选择提供商失败: %w", err)
	}

	// 调用选中的提供商
	result, err := provider.Generate(ctx, opts)
	if err != nil {
		// 降级到其他提供商
		return s.fallbackCall(ctx, opts, provider.Name())
	}

	return result, nil
}

// fallbackCall 降级调用
func (s *MultiModelService) fallbackCall(ctx context.Context, opts *GenerateOptions, failedProvider ProviderType) (*GenerateResult, error) {
	// 获取备用提供商列表
	backupProviders := s.selector.GetBackupProviders(failedProvider)

	for _, providerType := range backupProviders {
		provider, exists := s.providers[providerType]
		if !exists {
			continue
		}

		result, err := provider.Generate(ctx, opts)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("所有提供商都不可用")
}

// GetProviderStats 获取提供商统计
func (s *MultiModelService) GetProviderStats() map[ProviderType]*ProviderStats {
	return s.health.GetStats()
}

// Stop 停止服务
func (s *MultiModelService) Stop() {
	s.health.Stop()
}
