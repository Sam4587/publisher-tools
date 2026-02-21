package cost

import (
	"fmt"
	"sort"
	"time"

	"publisher-tools/database"

	"gorm.io/gorm"
)

// OptimizationService 成本优化服务
type OptimizationService struct {
	db          *gorm.DB
	costService *CostService
}

// NewOptimizationService 创建优化服务
func NewOptimizationService(db *gorm.DB, costService *CostService) *OptimizationService {
	return &OptimizationService{
		db:          db,
		costService: costService,
	}
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type        string  `json:"type"`         // model_switch, cache_usage, prompt_optimization, etc.
	Priority    string  `json:"priority"`     // high, medium, low
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PotentialSavings float64 `json:"potential_savings"` // 潜在节省金额
	Impact      string  `json:"impact"`       // high, medium, low
	Action      string `json:"action"`       // 具体操作建议
}

// GenerateSuggestions 生成优化建议
func (s *OptimizationService) GenerateSuggestions(userID, projectID string) ([]OptimizationSuggestion, error) {
	var suggestions []OptimizationSuggestion

	// 获取最近30天的成本数据
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// 1. 分析模型使用情况
	modelStats, err := s.costService.GetCostByModel(userID, projectID, startTime, endTime)
	if err == nil {
		suggestions = append(suggestions, s.analyzeModelUsage(modelStats)...)
	}

	// 2. 分析缓存使用情况
	cacheSuggestions := s.analyzeCacheUsage(userID, projectID, startTime, endTime)
	suggestions = append(suggestions, cacheSuggestions...)

	// 3. 分析提示词效率
	promptSuggestions := s.analyzePromptEfficiency(userID, projectID, startTime, endTime)
	suggestions = append(suggestions, promptSuggestions...)

	// 4. 分析提供商性能
	providerStats, err := s.costService.GetCostByProvider(userID, projectID, startTime, endTime)
	if err == nil {
		suggestions = append(suggestions, s.analyzeProviderPerformance(providerStats)...)
	}

	// 按优先级排序
	sort.Slice(suggestions, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priorityOrder[suggestions[i].Priority] > priorityOrder[suggestions[j].Priority]
	})

	return suggestions, nil
}

// analyzeModelUsage 分析模型使用情况
func (s *OptimizationService) analyzeModelUsage(stats []ModelCostStats) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	for _, stat := range stats {
		// 检查是否使用了昂贵的模型
		if stat.TotalCost > 10.0 { // 总成本超过$10
			// 建议切换到更便宜的模型
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:             "model_switch",
				Priority:         "high",
				Title:            fmt.Sprintf("考虑切换 %s 模型", stat.Model),
				Description:      fmt.Sprintf("模型 %s 在30天内花费 $%.2f，建议评估是否可以使用更便宜的模型", stat.Model, stat.TotalCost),
				PotentialSavings: stat.TotalCost * 0.3, // 假设可以节省30%
				Impact:           "high",
				Action:           fmt.Sprintf("评估是否可以用更便宜的模型替代 %s", stat.Model),
			})
		}

		// 检查平均响应时间
		if stat.AvgDurationMs > 5000 { // 平均响应时间超过5秒
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:             "performance",
				Priority:         "medium",
				Title:            fmt.Sprintf("优化 %s 模型响应时间", stat.Model),
				Description:      fmt.Sprintf("模型 %s 平均响应时间 %.0fms，可能影响用户体验", stat.Model, stat.AvgDurationMs),
				PotentialSavings: 0,
				Impact:           "medium",
				Action:           "考虑使用更快的模型或优化提示词长度",
			})
		}
	}

	return suggestions
}

// analyzeCacheUsage 分析缓存使用情况
func (s *OptimizationService) analyzeCacheUsage(userID, projectID string, startTime, endTime time.Time) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// 查询缓存使用情况
	var totalCalls, cachedCalls int64
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	query.Select("COUNT(*) as total_calls, SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_calls").
		Scan(&struct {
			TotalCalls  int64
			CachedCalls int64
		}{&totalCalls, &cachedCalls})

	if totalCalls > 0 {
		cacheHitRate := float64(cachedCalls) / float64(totalCalls) * 100
		if cacheHitRate < 30 { // 缓存命中率低于30%
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:             "cache_usage",
				Priority:         "high",
				Title:            "提高缓存使用率",
				Description:      fmt.Sprintf("当前缓存命中率仅 %.1f%%，建议优化缓存策略", cacheHitRate),
				PotentialSavings: float64(totalCalls-cachedCalls) * 0.001, // 假设每次调用$0.001
				Impact:           "high",
				Action:           "增加缓存TTL时间，优化缓存键生成策略",
			})
		}
	}

	return suggestions
}

// analyzePromptEfficiency 分析提示词效率
func (s *OptimizationService) analyzePromptEfficiency(userID, projectID string, startTime, endTime time.Time) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// 查询平均token使用量
	var avgTokens float64
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	query.Select("AVG(total_tokens)").Scan(&avgTokens)

	if avgTokens > 1000 { // 平均token超过1000
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:             "prompt_optimization",
			Priority:         "medium",
			Title:            "优化提示词长度",
			Description:      fmt.Sprintf("平均每次调用使用 %.0f tokens，建议精简提示词", avgTokens),
			PotentialSavings: avgTokens * 0.0001 * 100, // 假设100次调用
			Impact:           "medium",
			Action:           "精简提示词，移除冗余内容，使用更简洁的表达",
		})
	}

	return suggestions
}

// analyzeProviderPerformance 分析提供商性能
func (s *OptimizationService) analyzeProviderPerformance(stats []ProviderCostStats) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	for _, stat := range stats {
		// 检查是否有性能问题
		if stat.AvgDurationMs > 3000 { // 平均响应时间超过3秒
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:             "provider_switch",
				Priority:         "low",
				Title:            fmt.Sprintf("考虑切换 %s 提供商", stat.Provider),
				Description:      fmt.Sprintf("提供商 %s 平均响应时间 %.0fms，可能影响性能", stat.Provider, stat.AvgDurationMs),
				PotentialSavings: 0,
				Impact:           "low",
				Action:           fmt.Sprintf("评估是否切换到更快的提供商替代 %s", stat.Provider),
			})
		}
	}

	return suggestions
}

// GetCostReport 生成成本报告
func (s *OptimizationService) GetCostReport(userID, projectID string, startTime, endTime time.Time) (*CostReport, error) {
	// 获取总体统计
	stats, err := s.costService.GetCostStats(userID, projectID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取按提供商统计
	providerStats, err := s.costService.GetCostByProvider(userID, projectID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取按模型统计
	modelStats, err := s.costService.GetCostByModel(userID, projectID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取按功能统计
	functionStats, err := s.costService.GetCostByFunction(userID, projectID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 获取每日成本
	dailyStats, err := s.costService.GetDailyCost(userID, projectID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 计算成功率
	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessCalls) / float64(stats.TotalCalls) * 100
		stats.CacheHitRate = float64(stats.CachedCalls) / float64(stats.TotalCalls) * 100
		stats.AvgCostPerCall = stats.TotalCost / float64(stats.TotalCalls)
	}

	report := &CostReport{
		StartTime:     startTime,
		EndTime:       endTime,
		OverallStats:  stats,
		ProviderStats: providerStats,
		ModelStats:    modelStats,
		FunctionStats: functionStats,
		DailyStats:    dailyStats,
	}

	return report, nil
}

// CostReport 成本报告
type CostReport struct {
	StartTime     time.Time           `json:"start_time"`
	EndTime       time.Time           `json:"end_time"`
	OverallStats  *CostStats          `json:"overall_stats"`
	ProviderStats []ProviderCostStats `json:"provider_stats"`
	ModelStats    []ModelCostStats    `json:"model_stats"`
	FunctionStats []FunctionCostStats `json:"function_stats"`
	DailyStats    []DailyCostStats    `json:"daily_stats"`
}
