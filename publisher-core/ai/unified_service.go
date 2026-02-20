package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"publisher-core/ai/provider"
	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UnifiedService 统一的 AI 服务管理器
type UnifiedService struct {
	db      *gorm.DB
	mu      sync.RWMutex
	clients map[string]provider.Provider // 客户端缓存
	log     *logrus.Logger
}

// NewUnifiedService 创建统一的 AI 服务
func NewUnifiedService(db *gorm.DB) *UnifiedService {
	if db == nil {
		db = database.GetDB()
	}
	return &UnifiedService{
		db:      db,
		clients: make(map[string]provider.Provider),
		log:     logrus.StandardLogger(),
	}
}

// GetDefaultConfig 获取默认配置
func (s *UnifiedService) GetDefaultConfig(serviceType string) (*database.AIServiceConfig, error) {
	var config database.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_default = ? AND is_active = ?", serviceType, true, true).
		First(&config).Error
	if err != nil {
		// 如果没有默认配置，获取优先级最高的活跃配置
		err = s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
			Order("priority desc").
			First(&config).Error
		if err != nil {
			return nil, fmt.Errorf("no active AI config found for type %s", serviceType)
		}
	}
	return &config, nil
}

// GetActiveConfigs 获取所有活跃配置（按优先级排序）
func (s *UnifiedService) GetActiveConfigs(serviceType string) ([]database.AIServiceConfig, error) {
	var configs []database.AIServiceConfig
	err := s.db.Where("service_type = ? AND is_active = ?", serviceType, true).
		Order("priority desc").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("no active AI config found for type %s", serviceType)
	}
	return configs, nil
}

// GetOrCreateClient 获取或创建客户端
func (s *UnifiedService) GetOrCreateClient(config *database.AIServiceConfig) (provider.Provider, error) {
	key := fmt.Sprintf("%s:%s", config.Provider, config.Model)

	s.mu.RLock()
	client, ok := s.clients[key]
	s.mu.RUnlock()

	if ok {
		return client, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if client, ok := s.clients[key]; ok {
		return client, nil
	}

	// 创建新的客户端
	client, err := s.createProvider(config)
	if err != nil {
		return nil, err
	}

	s.clients[key] = client
	s.log.Infof("Created AI client: %s (%s)", config.Name, key)

	return client, nil
}

// createProvider 根据配置创建 Provider
func (s *UnifiedService) createProvider(config *database.AIServiceConfig) (provider.Provider, error) {
	switch config.Provider {
	case "openrouter":
		return provider.NewOpenRouterProvider(config.APIKey), nil
	case "google":
		return provider.NewGoogleProvider(config.APIKey), nil
	case "groq":
		return provider.NewGroqProvider(config.APIKey), nil
	case "deepseek":
		if config.BaseURL != "" {
			return provider.NewDeepSeekProviderWithBaseURL(config.APIKey, config.BaseURL), nil
		}
		return provider.NewDeepSeekProvider(config.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// Generate 生成文本（支持降级）
func (s *UnifiedService) Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	configs, err := s.GetActiveConfigs("text")
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, config := range configs {
		client, err := s.GetOrCreateClient(&config)
		if err != nil {
			lastErr = err
			s.log.Warnf("Failed to create client for %s: %v", config.Name, err)
			continue
		}

		// 如果没有指定模型，使用配置中的模型
		if opts.Model == "" {
			opts.Model = config.Model
		}

		startTime := time.Now()
		result, err := client.Generate(ctx, opts)
		if err == nil {
			// 记录成功的历史
			s.recordHistory(&config, opts, result, time.Since(startTime), true, "")
			return result, nil
		}

		// 记录失败的历史
		s.recordHistory(&config, opts, nil, time.Since(startTime), false, err.Error())

		s.log.Warnf("AI generation failed for %s, trying next provider: %v", config.Name, err)
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// GenerateStream 流式生成文本（支持降级）
func (s *UnifiedService) GenerateStream(ctx context.Context, opts *provider.GenerateOptions) (<-chan string, error) {
	configs, err := s.GetActiveConfigs("text")
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, config := range configs {
		client, err := s.GetOrCreateClient(&config)
		if err != nil {
			lastErr = err
			continue
		}

		if opts.Model == "" {
			opts.Model = config.Model
		}

		ch, err := client.GenerateStream(ctx, opts)
		if err == nil {
			return ch, nil
		}

		s.log.Warnf("AI stream generation failed for %s, trying next provider: %v", config.Name, err)
		lastErr = err
	}

	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// GenerateWithProvider 使用指定提供商生成
func (s *UnifiedService) GenerateWithProvider(ctx context.Context, providerName string, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	var config database.AIServiceConfig
	err := s.db.Where("provider = ? AND is_active = ?", providerName, true).
		Order("priority desc").
		First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("provider %s not found or inactive", providerName)
	}

	client, err := s.GetOrCreateClient(&config)
	if err != nil {
		return nil, err
	}

	if opts.Model == "" {
		opts.Model = config.Model
	}

	return client.Generate(ctx, opts)
}

// GenerateWithModel 使用指定模型生成
func (s *UnifiedService) GenerateWithModel(ctx context.Context, providerName, model string, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
	var config database.AIServiceConfig
	err := s.db.Where("provider = ? AND model = ? AND is_active = ?", providerName, model, true).
		First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("provider %s with model %s not found or inactive", providerName, model)
	}

	client, err := s.GetOrCreateClient(&config)
	if err != nil {
		return nil, err
	}

	opts.Model = model
	return client.Generate(ctx, opts)
}

// recordHistory 记录 AI 调用历史
func (s *UnifiedService) recordHistory(config *database.AIServiceConfig, opts *provider.GenerateOptions, result *provider.GenerateResult, duration time.Duration, success bool, errMsg string) {
	history := &database.AIHistory{
		Provider:     config.Provider,
		Model:        opts.Model,
		Success:      success,
		ErrorMessage: errMsg,
		DurationMs:   int(duration.Milliseconds()),
		CreatedAt:    time.Now(),
	}

	// 记录提示词（截断过长的内容）
	if len(opts.Messages) > 0 {
		promptBytes, _ := opts.Messages[0].Content, ""
		if len(promptBytes) > 10000 {
			history.Prompt = string(promptBytes[:10000]) + "..."
		} else {
			history.Prompt = string(promptBytes)
		}
	}

	if result != nil {
		history.TokensUsed = result.InputTokens + result.OutputTokens
		// 截断过长的响应
		if len(result.Content) > 10000 {
			history.Response = result.Content[:10000] + "..."
		} else {
			history.Response = result.Content
		}
	}

	if err := s.db.Create(history).Error; err != nil {
		s.log.Warnf("Failed to record AI history: %v", err)
	}
}

// ListProviders 列出所有可用的提供商
func (s *UnifiedService) ListProviders() ([]database.AIServiceConfig, error) {
	var configs []database.AIServiceConfig
	err := s.db.Where("is_active = ?", true).
		Order("service_type, priority desc").
		Find(&configs).Error
	return configs, err
}

// UpdateConfig 更新 AI 配置
func (s *UnifiedService) UpdateConfig(config *database.AIServiceConfig) error {
	config.UpdatedAt = time.Now()

	// 如果设置为默认，取消同类型其他配置的默认状态
	if config.IsDefault {
		s.db.Model(&database.AIServiceConfig{}).
			Where("service_type = ? AND id != ?", config.ServiceType, config.ID).
			Update("is_default", false)
	}

	// 清除客户端缓存
	key := fmt.Sprintf("%s:%s", config.Provider, config.Model)
	s.mu.Lock()
	delete(s.clients, key)
	s.mu.Unlock()

	return s.db.Save(config).Error
}

// CreateConfig 创建新的 AI 配置
func (s *UnifiedService) CreateConfig(config *database.AIServiceConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// 如果设置为默认，取消同类型其他配置的默认状态
	if config.IsDefault {
		s.db.Model(&database.AIServiceConfig{}).
			Where("service_type = ?", config.ServiceType).
			Update("is_default", false)
	}

	return s.db.Create(config).Error
}

// DeleteConfig 删除 AI 配置
func (s *UnifiedService) DeleteConfig(id uint) error {
	// 清除客户端缓存
	var config database.AIServiceConfig
	if err := s.db.First(&config, id).Error; err == nil {
		key := fmt.Sprintf("%s:%s", config.Provider, config.Model)
		s.mu.Lock()
		delete(s.clients, key)
		s.mu.Unlock()
	}

	return s.db.Delete(&database.AIServiceConfig{}, id).Error
}

// GetHistory 获取 AI 调用历史
func (s *UnifiedService) GetHistory(limit, offset int) ([]database.AIHistory, int64, error) {
	var history []database.AIHistory
	var total int64

	s.db.Model(&database.AIHistory{}).Count(&total)

	err := s.db.Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	return history, total, err
}

// GetStats 获取 AI 调用统计
func (s *UnifiedService) GetStats(since time.Time) (*AIStats, error) {
	stats := &AIStats{}

	// 总调用次数
	s.db.Model(&database.AIHistory{}).Where("created_at > ?", since).Count(&stats.TotalCalls)

	// 成功次数
	s.db.Model(&database.AIHistory{}).Where("created_at > ? AND success = ?", since, true).Count(&stats.SuccessCalls)

	// 失败次数
	stats.FailedCalls = stats.TotalCalls - stats.SuccessCalls

	// 总 Token 使用量
	s.db.Model(&database.AIHistory{}).Where("created_at > ?", since).
		Select("COALESCE(SUM(tokens_used), 0)").Scan(&stats.TotalTokens)

	// 平均响应时间
	s.db.Model(&database.AIHistory{}).Where("created_at > ?", since).
		Select("COALESCE(AVG(duration_ms), 0)").Scan(&stats.AvgDurationMs)

	// 按提供商统计
	var providerStats []ProviderStats
	s.db.Model(&database.AIHistory{}).
		Select("provider, COUNT(*) as count, SUM(tokens_used) as tokens, AVG(duration_ms) as avg_duration").
		Where("created_at > ?", since).
		Group("provider").
		Find(&providerStats)
	stats.ByProvider = providerStats

	return stats, nil
}

// AIStats AI 调用统计
type AIStats struct {
	TotalCalls    int64
	SuccessCalls  int64
	FailedCalls   int64
	TotalTokens   int64
	AvgDurationMs float64
	ByProvider    []ProviderStats
}

// ProviderStats 提供商统计
type ProviderStats struct {
	Provider   string
	Count      int64
	Tokens     int64
	AvgDuration float64
}

// ClearCache 清除客户端缓存
func (s *UnifiedService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients = make(map[string]provider.Provider)
	s.log.Info("AI client cache cleared")
}
