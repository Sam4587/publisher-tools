package cost

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"publisher-tools/database"

	"gorm.io/gorm"
)

// CostService 成本服务
type CostService struct {
	db *gorm.DB
}

// NewCostService 创建成本服务
func NewCostService(db *gorm.DB) *CostService {
	return &CostService{db: db}
}

// CostRecordInput 成本记录输入
type CostRecordInput struct {
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	UserID       string                 `json:"user_id"`
	ProjectID    string                 `json:"project_id"`
	FunctionType string                 `json:"function_type"`
	InputTokens  int                    `json:"input_tokens"`
	OutputTokens int                    `json:"output_tokens"`
	DurationMs   int                    `json:"duration_ms"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message"`
	RequestID    string                 `json:"request_id"`
	Prompt       string                 `json:"prompt"`
	Cached       bool                   `json:"cached"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// RecordCost 记录成本
func (s *CostService) RecordCost(input *CostRecordInput) (*database.AICostRecord, error) {
	// 计算成本
	totalTokens := input.InputTokens + input.OutputTokens
	costPerToken, totalCost := s.CalculateCost(input.Provider, input.Model, input.InputTokens, input.OutputTokens)

	// 生成提示词哈希
	promptHash := ""
	if input.Prompt != "" {
		hash := sha256.Sum256([]byte(input.Prompt))
		promptHash = hex.EncodeToString(hash[:])
	}

	// 序列化元数据
	metadataJSON := ""
	if input.Metadata != nil {
		metadataBytes, _ := json.Marshal(input.Metadata)
		metadataJSON = string(metadataBytes)
	}

	record := &database.AICostRecord{
		Provider:      input.Provider,
		Model:         input.Model,
		UserID:        input.UserID,
		ProjectID:     input.ProjectID,
		FunctionType:  input.FunctionType,
		InputTokens:   input.InputTokens,
		OutputTokens:  input.OutputTokens,
		TotalTokens:   totalTokens,
		CostPerToken:  costPerToken,
		TotalCost:     totalCost,
		DurationMs:    input.DurationMs,
		Success:       input.Success,
		ErrorMessage:  input.ErrorMessage,
		RequestID:     input.RequestID,
		PromptHash:    promptHash,
		Cached:        input.Cached,
		Metadata:      metadataJSON,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("记录成本失败: %w", err)
	}

	return record, nil
}

// CalculateCost 计算成本
func (s *CostService) CalculateCost(provider, model string, inputTokens, outputTokens int) (float64, float64) {
	// 查询模型定价
	var pricing database.AIModelPricing
	err := s.db.Where("provider = ? AND model = ? AND is_active = ?", provider, model, true).
		Order("effective_date DESC").
		First(&pricing).Error

	if err != nil {
		// 使用默认定价
		return s.getDefaultPricing(provider, model, inputTokens, outputTokens)
	}

	// 计算成本
	inputCost := float64(inputTokens) / 1000 * pricing.InputPrice
	outputCost := float64(outputTokens) / 1000 * pricing.OutputPrice
	totalCost := inputCost + outputCost

	costPerToken := 0.0
	if inputTokens+outputTokens > 0 {
		costPerToken = totalCost / float64(inputTokens+outputTokens)
	}

	return costPerToken, totalCost
}

// getDefaultPricing 获取默认定价
func (s *CostService) getDefaultPricing(provider, model string, inputTokens, outputTokens int) (float64, float64) {
	// 默认定价表（美元/1K tokens）
	defaultPricing := map[string]struct {
		InputPrice  float64
		OutputPrice float64
	}{
		"openrouter:gpt-3.5-turbo":      {0.0015, 0.002},
		"openrouter:gpt-4":              {0.03, 0.06},
		"openrouter:gpt-4-turbo":        {0.01, 0.03},
		"google:gemini-pro":             {0.00025, 0.0005},
		"groq:mixtral-8x7b-32768":       {0.00027, 0.00027},
		"deepseek:deepseek-chat":        {0.00014, 0.00028},
		"nvidia:meta/llama-3.1-8b-instruct": {0.00022, 0.00022},
		"mistral:mistral-large-latest":  {0.004, 0.012},
	}

	key := fmt.Sprintf("%s:%s", provider, model)
	pricing, exists := defaultPricing[key]
	if !exists {
		// 使用默认价格
		pricing = struct {
			InputPrice  float64
			OutputPrice float64
		}{InputPrice: 0.001, OutputPrice: 0.002}
	}

	inputCost := float64(inputTokens) / 1000 * pricing.InputPrice
	outputCost := float64(outputTokens) / 1000 * pricing.OutputPrice
	totalCost := inputCost + outputCost

	costPerToken := 0.0
	if inputTokens+outputTokens > 0 {
		costPerToken = totalCost / float64(inputTokens+outputTokens)
	}

	return costPerToken, totalCost
}

// GetCostStats 获取成本统计
func (s *CostService) GetCostStats(userID, projectID string, startTime, endTime time.Time) (*CostStats, error) {
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var stats CostStats
	if err := query.Select(`
		COUNT(*) as total_calls,
		COALESCE(SUM(total_cost), 0) as total_cost,
		COALESCE(SUM(total_tokens), 0) as total_tokens,
		COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) as success_calls,
		COALESCE(SUM(CASE WHEN cached THEN 1 ELSE 0 END), 0) as cached_calls,
		COALESCE(AVG(duration_ms), 0) as avg_duration_ms
	`).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取成本统计失败: %w", err)
	}

	return &stats, nil
}

// CostStats 成本统计
type CostStats struct {
	TotalCalls     int64   `json:"total_calls"`
	TotalCost      float64 `json:"total_cost"`
	TotalTokens    int64   `json:"total_tokens"`
	SuccessCalls   int64   `json:"success_calls"`
	CachedCalls    int64   `json:"cached_calls"`
	AvgDurationMs  float64 `json:"avg_duration_ms"`
	SuccessRate    float64 `json:"success_rate"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	AvgCostPerCall float64 `json:"avg_cost_per_call"`
}

// GetCostByProvider 按提供商统计成本
func (s *CostService) GetCostByProvider(userID, projectID string, startTime, endTime time.Time) ([]ProviderCostStats, error) {
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var stats []ProviderCostStats
	if err := query.Select(`
		provider,
		COUNT(*) as total_calls,
		COALESCE(SUM(total_cost), 0) as total_cost,
		COALESCE(SUM(total_tokens), 0) as total_tokens,
		COALESCE(AVG(duration_ms), 0) as avg_duration_ms
	`).Group("provider").Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取提供商成本统计失败: %w", err)
	}

	return stats, nil
}

// ProviderCostStats 提供商成本统计
type ProviderCostStats struct {
	Provider      string  `json:"provider"`
	TotalCalls    int64   `json:"total_calls"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int64   `json:"total_tokens"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
}

// GetCostByModel 按模型统计成本
func (s *CostService) GetCostByModel(userID, projectID string, startTime, endTime time.Time) ([]ModelCostStats, error) {
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var stats []ModelCostStats
	if err := query.Select(`
		provider,
		model,
		COUNT(*) as total_calls,
		COALESCE(SUM(total_cost), 0) as total_cost,
		COALESCE(SUM(total_tokens), 0) as total_tokens,
		COALESCE(AVG(duration_ms), 0) as avg_duration_ms
	`).Group("provider, model").Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取模型成本统计失败: %w", err)
	}

	return stats, nil
}

// ModelCostStats 模型成本统计
type ModelCostStats struct {
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	TotalCalls    int64   `json:"total_calls"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int64   `json:"total_tokens"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
}

// GetCostByFunction 按功能类型统计成本
func (s *CostService) GetCostByFunction(userID, projectID string, startTime, endTime time.Time) ([]FunctionCostStats, error) {
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var stats []FunctionCostStats
	if err := query.Select(`
		function_type,
		COUNT(*) as total_calls,
		COALESCE(SUM(total_cost), 0) as total_cost,
		COALESCE(SUM(total_tokens), 0) as total_tokens
	`).Group("function_type").Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取功能成本统计失败: %w", err)
	}

	return stats, nil
}

// FunctionCostStats 功能成本统计
type FunctionCostStats struct {
	FunctionType string  `json:"function_type"`
	TotalCalls   int64   `json:"total_calls"`
	TotalCost    float64 `json:"total_cost"`
	TotalTokens  int64   `json:"total_tokens"`
}

// GetDailyCost 获取每日成本
func (s *CostService) GetDailyCost(userID, projectID string, startTime, endTime time.Time) ([]DailyCostStats, error) {
	query := s.db.Model(&database.AICostRecord{}).Where("created_at BETWEEN ? AND ?", startTime, endTime)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	var stats []DailyCostStats
	if err := query.Select(`
		DATE(created_at) as date,
		COUNT(*) as total_calls,
		COALESCE(SUM(total_cost), 0) as total_cost,
		COALESCE(SUM(total_tokens), 0) as total_tokens
	`).Group("DATE(created_at)").Order("date").Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("获取每日成本统计失败: %w", err)
	}

	return stats, nil
}

// DailyCostStats 每日成本统计
type DailyCostStats struct {
	Date        string  `json:"date"`
	TotalCalls  int64   `json:"total_calls"`
	TotalCost   float64 `json:"total_cost"`
	TotalTokens int64   `json:"total_tokens"`
}
