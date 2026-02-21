package asr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Selector ASR提供商选择器
type Selector struct {
	providers    []Provider
	cache        *ResultCache
	cacheEnabled bool
	mu           sync.RWMutex
	stats        *SelectorStats
}

// SelectorConfig 选择器配置
type SelectorConfig struct {
	Providers     []ProviderConfig `json:"providers"`
	CacheEnabled  bool             `json:"cache_enabled"`
	CacheDir      string           `json:"cache_dir"`
	CacheTTL      time.Duration    `json:"cache_ttl"`
	CacheMaxSize  int64            `json:"cache_max_size"`
}

// DefaultSelectorConfig 默认配置
func DefaultSelectorConfig() *SelectorConfig {
	return &SelectorConfig{
		CacheEnabled: true,
		CacheDir:     "./data/asr_cache",
		CacheTTL:     24 * time.Hour,
		CacheMaxSize: 1024 * 1024 * 1024, // 1GB
	}
}

// SelectorStats 选择器统计
type SelectorStats struct {
	mu             sync.RWMutex
	TotalRequests  int64            `json:"total_requests"`
	CacheHits      int64            `json:"cache_hits"`
	CacheMisses    int64            `json:"cache_misses"`
	ProviderStats  map[ProviderType]*ProviderStats `json:"provider_stats"`
	FallbackCount  int64            `json:"fallback_count"`
	TotalDuration  time.Duration    `json:"total_duration"`
}

// ProviderStats 提供商统计
type ProviderStats struct {
	Requests    int64         `json:"requests"`
	Successes   int64         `json:"successes"`
	Failures    int64         `json:"failures"`
	TotalTime   time.Duration `json:"total_time"`
	AvgTime     time.Duration `json:"avg_time"`
	LastUsed    time.Time     `json:"last_used"`
	LastFailure time.Time     `json:"last_failure"`
}

// NewSelector 创建选择器
func NewSelector(config *SelectorConfig) *Selector {
	if config == nil {
		config = DefaultSelectorConfig()
	}

	s := &Selector{
		providers:    make([]Provider, 0),
		cacheEnabled: config.CacheEnabled,
		stats: &SelectorStats{
			ProviderStats: make(map[ProviderType]*ProviderStats),
		},
	}

	// 初始化缓存
	if config.CacheEnabled {
		s.cache = NewResultCache(config.CacheDir, config.CacheTTL, config.CacheMaxSize)
	}

	return s
}

// RegisterProvider 注册提供商
func (s *Selector) RegisterProvider(provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.providers = append(s.providers, provider)
	// 按优先级排序
	sort.Slice(s.providers, func(i, j int) bool {
		return s.providers[i].GetPriority() < s.providers[j].GetPriority()
	})

	// 初始化统计
	s.stats.ProviderStats[provider.Name()] = &ProviderStats{}
}

// Recognize 执行语音识别（自动选择提供商）
func (s *Selector) Recognize(ctx context.Context, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	if opts == nil {
		opts = DefaultRecognizeOptions()
	}

	s.stats.mu.Lock()
	s.stats.TotalRequests++
	s.stats.mu.Unlock()

	startTime := time.Now()

	// 检查缓存
	if s.cacheEnabled && s.cache != nil {
		cacheKey := s.generateCacheKey(audioPath, opts)
		if cached, err := s.cache.Get(cacheKey); err == nil {
			cached.Cached = true
			s.stats.mu.Lock()
			s.stats.CacheHits++
			s.stats.mu.Unlock()
			logrus.Debugf("ASR cache hit for %s", audioPath)
			return cached, nil
		}
		s.stats.mu.Lock()
		s.stats.CacheMisses++
		s.stats.mu.Unlock()
	}

	// 获取可用提供商
	providers := s.getAvailableProviders(opts.Language)
	if len(providers) == 0 {
		return nil, fmt.Errorf("no available ASR provider")
	}

	// 尝试识别（带降级）
	var lastErr error
	var result *RecognitionResult

	for _, provider := range providers {
		logrus.Debugf("Trying ASR provider: %s", provider.Name())

		result, lastErr = s.tryProvider(ctx, provider, audioPath, opts)
		if lastErr == nil {
			break
		}

		logrus.Warnf("ASR provider %s failed: %v, trying next", provider.Name(), lastErr)
		s.stats.mu.Lock()
		s.stats.FallbackCount++
		s.stats.mu.Unlock()
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all ASR providers failed: %w", lastErr)
	}

	// 保存到缓存
	if s.cacheEnabled && s.cache != nil {
		cacheKey := s.generateCacheKey(audioPath, opts)
		if err := s.cache.Set(cacheKey, result); err != nil {
			logrus.Warnf("Failed to cache ASR result: %v", err)
		}
	}

	// 更新统计
	s.stats.mu.Lock()
	s.stats.TotalDuration += time.Since(startTime)
	s.stats.mu.Unlock()

	return result, nil
}

// tryProvider 尝试使用指定提供商
func (s *Selector) tryProvider(ctx context.Context, provider Provider, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	startTime := time.Now()

	result, err := provider.Recognize(ctx, audioPath, opts)

	// 更新提供商统计
	stats := s.stats.ProviderStats[provider.Name()]
	stats.mu.Lock()
	stats.Requests++
	stats.LastUsed = time.Now()

	if err != nil {
		stats.Failures++
		stats.LastFailure = time.Now()
	} else {
		stats.Successes++
		stats.TotalTime += time.Since(startTime)
		stats.AvgTime = time.Duration(int64(stats.TotalTime) / stats.Successes)
	}
	stats.mu.Unlock()

	return result, err
}

// getAvailableProviders 获取可用提供商列表
func (s *Selector) getAvailableProviders(language string) []Provider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var available []Provider
	for _, p := range s.providers {
		if p.IsAvailable() && p.SupportsLanguage(language) {
			available = append(available, p)
		}
	}
	return available
}

// generateCacheKey 生成缓存键
func (s *Selector) generateCacheKey(audioPath string, opts *RecognizeOptions) string {
	// 获取文件信息
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return ""
	}

	// 生成唯一键
	keyData := fmt.Sprintf("%s:%d:%d:%s:%s",
		filepath.Base(audioPath),
		fileInfo.Size(),
		fileInfo.ModTime().Unix(),
		opts.Language,
		opts.Model,
	)

	hash := sha256.Sum256([]byte(keyData))
	return hex.EncodeToString(hash[:])
}

// RecognizeWithProvider 使用指定提供商识别
func (s *Selector) RecognizeWithProvider(ctx context.Context, providerType ProviderType, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.providers {
		if p.Name() == providerType {
			return s.tryProvider(ctx, p, audioPath, opts)
		}
	}

	return nil, fmt.Errorf("provider not found: %s", providerType)
}

// GetStats 获取统计信息
func (s *Selector) GetStats() *SelectorStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 复制统计信息
	stats := &SelectorStats{
		TotalRequests: s.stats.TotalRequests,
		CacheHits:     s.stats.CacheHits,
		CacheMisses:   s.stats.CacheMisses,
		FallbackCount: s.stats.FallbackCount,
		TotalDuration: s.stats.TotalDuration,
		ProviderStats: make(map[ProviderType]*ProviderStats),
	}

	for k, v := range s.stats.ProviderStats {
		v.mu.RLock()
		stats.ProviderStats[k] = &ProviderStats{
			Requests:    v.Requests,
			Successes:   v.Successes,
			Failures:    v.Failures,
			TotalTime:   v.TotalTime,
			AvgTime:     v.AvgTime,
			LastUsed:    v.LastUsed,
			LastFailure: v.LastFailure,
		}
		v.mu.RUnlock()
	}

	return stats
}

// ClearCache 清除缓存
func (s *Selector) ClearCache() error {
	if s.cache != nil {
		return s.cache.Clear()
	}
	return nil
}

// GetProviders 获取所有提供商
func (s *Selector) GetProviders() []ProviderInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var infos []ProviderInfo
	for _, p := range s.providers {
		infos = append(infos, ProviderInfo{
			Name:        p.Name(),
			Available:   p.IsAvailable(),
			Priority:    p.GetPriority(),
			MaxFileSize: p.GetMaxFileSize(),
			MaxDuration: p.GetMaxDuration(),
		})
	}
	return infos
}

// ProviderInfo 提供商信息
type ProviderInfo struct {
	Name        ProviderType `json:"name"`
	Available   bool         `json:"available"`
	Priority    int          `json:"priority"`
	MaxFileSize int64        `json:"max_file_size"`
	MaxDuration float64      `json:"max_duration"`
}

// ResultCache 结果缓存
type ResultCache struct {
	dir      string
	ttl      time.Duration
	maxSize  int64
	mu       sync.RWMutex
}

// NewResultCache 创建结果缓存
func NewResultCache(dir string, ttl time.Duration, maxSize int64) *ResultCache {
	os.MkdirAll(dir, 0755)
	return &ResultCache{
		dir:     dir,
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get 获取缓存
func (c *ResultCache) Get(key string) (*RecognitionResult, error) {
	if key == "" {
		return nil, fmt.Errorf("empty cache key")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := filepath.Join(c.dir, key+".json")

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if time.Since(fileInfo.ModTime()) > c.ttl {
		os.Remove(filePath)
		return nil, fmt.Errorf("cache expired")
	}

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var result RecognitionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Set 设置缓存
func (c *ResultCache) Set(key string, result *RecognitionResult) error {
	if key == "" || result == nil {
		return fmt.Errorf("invalid cache parameters")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查缓存大小
	if err := c.checkSize(); err != nil {
		logrus.Warnf("Cache size check failed: %v", err)
	}

	filePath := filepath.Join(c.dir, key+".json")

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// checkSize 检查并清理缓存大小
func (c *ResultCache) checkSize() error {
	var totalSize int64
	var files []os.FileInfo

	// 遍历缓存目录
	filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			files = append(files, info)
		}
		return nil
	})

	// 如果超过最大大小，删除最旧的文件
	if totalSize > c.maxSize {
		// 按修改时间排序
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})

		// 删除旧文件直到大小符合要求
		for _, f := range files {
			if totalSize <= c.maxSize*9/10 { // 保留90%
				break
			}
			filePath := filepath.Join(c.dir, f.Name())
			os.Remove(filePath)
			totalSize -= f.Size()
		}
	}

	return nil
}

// Clear 清除所有缓存
func (c *ResultCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, f := range dir {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			os.Remove(filepath.Join(c.dir, f.Name()))
		}
	}

	return nil
}

// GetCacheStats 获取缓存统计
func (c *ResultCache) GetCacheStats() (int64, int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalSize int64
	var count int

	filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
			count++
		}
		return nil
	})

	return totalSize, count, nil
}
