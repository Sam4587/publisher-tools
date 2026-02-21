package cache

import (
	"encoding/json"
	"fmt"
	"time"
)

// AICacheService AI缓存服务
type AICacheService struct {
	cache *MemoryCache
}

// NewAICacheService 创建AI缓存服务
func NewAICacheService(config *CacheConfig) *AICacheService {
	return &AICacheService{
		cache: NewMemoryCache(config),
	}
}

// AICacheKey AI缓存键
type AICacheKey struct {
	Provider string      `json:"provider"`
	Model    string      `json:"model"`
	Prompt   string      `json:"prompt"`
	Options  interface{} `json:"options,omitempty"`
}

// AICacheValue AI缓存值
type AICacheValue struct {
	Response   string      `json:"response"`
	TokensUsed int         `json:"tokens_used"`
	Model      string      `json:"model"`
	Provider   string      `json:"provider"`
	CreatedAt  time.Time   `json:"created_at"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// GetCachedResponse 获取缓存的AI响应
func (s *AICacheService) GetCachedResponse(provider, model, prompt string, options interface{}) (*AICacheValue, bool) {
	key := &AICacheKey{
		Provider: provider,
		Model:    model,
		Prompt:   prompt,
		Options:  options,
	}

	cacheKey, err := GenerateCacheKey("ai_response", key)
	if err != nil {
		return nil, false
	}

	value, exists := s.cache.Get(cacheKey)
	if !exists {
		return nil, false
	}

	cacheValue, ok := value.(*AICacheValue)
	if !ok {
		return nil, false
	}

	return cacheValue, true
}

// SetCachedResponse 缓存AI响应
func (s *AICacheService) SetCachedResponse(provider, model, prompt string, options interface{}, response *AICacheValue, ttl time.Duration) error {
	key := &AICacheKey{
		Provider: provider,
		Model:    model,
		Prompt:   prompt,
		Options:  options,
	}

	cacheKey, err := GenerateCacheKey("ai_response", key)
	if err != nil {
		return err
	}

	response.CreatedAt = time.Now()
	return s.cache.Set(cacheKey, response, ttl)
}

// GetCachedPromptTemplate 获取缓存的提示词模板
func (s *AICacheService) GetCachedPromptTemplate(templateID string) (string, bool) {
	cacheKey := fmt.Sprintf("prompt_template:%s", templateID)
	value, exists := s.cache.Get(cacheKey)
	if !exists {
		return "", false
	}

	template, ok := value.(string)
	if !ok {
		return "", false
	}

	return template, true
}

// SetCachedPromptTemplate 缓存提示词模板
func (s *AICacheService) SetCachedPromptTemplate(templateID, content string, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("prompt_template:%s", templateID)
	return s.cache.Set(cacheKey, content, ttl)
}

// GetCachedHotspotAnalysis 获取缓存的热点分析
func (s *AICacheService) GetCachedHotspotAnalysis(hotspotID string) (interface{}, bool) {
	cacheKey := fmt.Sprintf("hotspot_analysis:%s", hotspotID)
	return s.cache.Get(cacheKey)
}

// SetCachedHotspotAnalysis 缓存热点分析
func (s *AICacheService) SetCachedHotspotAnalysis(hotspotID string, analysis interface{}, ttl time.Duration) error {
	cacheKey := fmt.Sprintf("hotspot_analysis:%s", hotspotID)
	return s.cache.Set(cacheKey, analysis, ttl)
}

// GetCachedContentGeneration 获取缓存的内容生成
func (s *AICacheService) GetCachedContentGeneration(topic, keywords string) (string, bool) {
	key := map[string]string{
		"topic":    topic,
		"keywords": keywords,
	}

	cacheKey, err := GenerateCacheKey("content_generation", key)
	if err != nil {
		return "", false
	}

	value, exists := s.cache.Get(cacheKey)
	if !exists {
		return "", false
	}

	content, ok := value.(string)
	if !ok {
		return "", false
	}

	return content, true
}

// SetCachedContentGeneration 缓存内容生成
func (s *AICacheService) SetCachedContentGeneration(topic, keywords, content string, ttl time.Duration) error {
	key := map[string]string{
		"topic":    topic,
		"keywords": keywords,
	}

	cacheKey, err := GenerateCacheKey("content_generation", key)
	if err != nil {
		return err
	}

	return s.cache.Set(cacheKey, content, ttl)
}

// InvalidateTemplateCache 使模板缓存失效
func (s *AICacheService) InvalidateTemplateCache(templateID string) error {
	cacheKey := fmt.Sprintf("prompt_template:%s", templateID)
	return s.cache.Delete(cacheKey)
}

// InvalidateHotspotCache 使热点缓存失效
func (s *AICacheService) InvalidateHotspotCache(hotspotID string) error {
	cacheKey := fmt.Sprintf("hotspot_analysis:%s", hotspotID)
	return s.cache.Delete(cacheKey)
}

// GetStats 获取缓存统计
func (s *AICacheService) GetStats() CacheStats {
	return s.cache.GetStats()
}

// Clear 清空所有缓存
func (s *AICacheService) Clear() {
	s.cache.Clear()
}

// Warmup 预热缓存
func (s *AICacheService) Warmup(items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := s.cache.Set(key, value, ttl); err != nil {
			return fmt.Errorf("预热缓存失败 [%s]: %w", key, err)
		}
	}
	return nil
}

// WarmupFromJSON 从JSON预热缓存
func (s *AICacheService) WarmupFromJSON(jsonData string, ttl time.Duration) error {
	var items map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &items); err != nil {
		return fmt.Errorf("解析JSON失败: %w", err)
	}
	return s.Warmup(items, ttl)
}

// GetCacheKeys 获取所有缓存键
func (s *AICacheService) GetCacheKeys() []string {
	return s.cache.GetKeys()
}

// Delete 删除指定缓存
func (s *AICacheService) Delete(key string) error {
	return s.cache.Delete(key)
}

// Stop 停止缓存服务
func (s *AICacheService) Stop() {
	s.cache.Stop()
}
