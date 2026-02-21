package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	HitCount   int         `json:"hit_count"`
	Size       int64       `json:"size"` // 字节数
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalItems    int64   `json:"total_items"`
	TotalSize     int64   `json:"total_size"`     // 总大小（字节）
	TotalHits     int64   `json:"total_hits"`     // 总命中次数
	TotalMisses   int64   `json:"total_misses"`   // 总未命中次数
	HitRate       float64 `json:"hit_rate"`       // 命中率
	Evictions     int64   `json:"evictions"`      // 驱逐次数
	LastClearTime string  `json:"last_clear_time"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxSize         int64         `json:"max_size"`          // 最大缓存大小（字节），0表示无限制
	MaxItems        int           `json:"max_items"`         // 最大缓存项数，0表示无限制
	DefaultTTL      time.Duration `json:"default_ttl"`       // 默认过期时间
	CleanupInterval time.Duration `json:"cleanup_interval"`  // 清理间隔
	EnableStats     bool          `json:"enable_stats"`      // 是否启用统计
}

// DefaultCacheConfig 默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxSize:         100 * 1024 * 1024, // 100MB
		MaxItems:        10000,
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		EnableStats:     true,
	}
}

// MemoryCache 内存缓存服务
type MemoryCache struct {
	items    map[string]*CacheItem
	mu       sync.RWMutex
	config   *CacheConfig
	stats    CacheStats
	stopChan chan struct{}
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(config *CacheConfig) *MemoryCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache := &MemoryCache{
		items:    make(map[string]*CacheItem),
		config:   config,
		stopChan: make(chan struct{}),
	}

	// 启动清理协程
	if config.CleanupInterval > 0 {
		go cache.cleanupLoop()
	}

	return cache
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 计算过期时间
	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}
	expiresAt := time.Now().Add(ttl)

	// 序列化值以计算大小
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存值失败: %w", err)
	}

	// 检查是否需要驱逐
	if err := c.evictIfNeeded(int64(len(valueBytes))); err != nil {
		return err
	}

	// 创建缓存项
	item := &CacheItem{
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		HitCount:  0,
		Size:      int64(len(valueBytes)),
	}

	c.items[key] = item
	return nil
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		if c.config.EnableStats {
			c.stats.TotalMisses++
		}
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		if c.config.EnableStats {
			c.stats.TotalMisses++
		}
		return nil, false
	}

	// 更新命中计数
	item.HitCount++
	if c.config.EnableStats {
		c.stats.TotalHits++
	}

	return item.Value, true
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		return nil
	}

	return fmt.Errorf("缓存键不存在: %s", key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	if c.config.EnableStats {
		c.stats.LastClearTime = time.Now().Format(time.RFC3339)
	}
}

// GetStats 获取缓存统计
func (c *MemoryCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.TotalItems = int64(len(c.items))

	// 计算总大小
	var totalSize int64
	for _, item := range c.items {
		totalSize += item.Size
	}
	stats.TotalSize = totalSize

	// 计算命中率
	totalRequests := stats.TotalHits + stats.TotalMisses
	if totalRequests > 0 {
		stats.HitRate = float64(stats.TotalHits) / float64(totalRequests) * 100
	}

	return stats
}

// GetKeys 获取所有缓存键
func (c *MemoryCache) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// Exists 检查缓存是否存在
func (c *MemoryCache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// 检查是否过期
	return !time.Now().After(item.ExpiresAt)
}

// evictIfNeeded 根据需要驱逐缓存
func (c *MemoryCache) evictIfNeeded(newItemSize int64) error {
	// 检查项数限制
	if c.config.MaxItems > 0 && len(c.items) >= c.config.MaxItems {
		if err := c.evictLRU(); err != nil {
			return err
		}
	}

	// 检查大小限制
	if c.config.MaxSize > 0 {
		var currentSize int64
		for _, item := range c.items {
			currentSize += item.Size
		}

		// 如果添加新项会超过限制，驱逐旧项
		for currentSize+newItemSize > c.config.MaxSize && len(c.items) > 0 {
			if err := c.evictLRU(); err != nil {
				return err
			}
			currentSize = 0
			for _, item := range c.items {
				currentSize += item.Size
			}
		}
	}

	return nil
}

// evictLRU 驱逐最少使用的缓存项
func (c *MemoryCache) evictLRU() error {
	if len(c.items) == 0 {
		return nil
	}

	// 找到最少使用的项（命中次数最少）
	var lruKey string
	minHits := int(^uint(0) >> 1) // 最大int值

	for key, item := range c.items {
		if item.HitCount < minHits {
			minHits = item.HitCount
			lruKey = key
		}
	}

	if lruKey != "" {
		delete(c.items, lruKey)
		if c.config.EnableStats {
			c.stats.Evictions++
		}
	}

	return nil
}

// cleanupLoop 定期清理过期缓存
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup 清理过期缓存
func (c *MemoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}

// Stop 停止缓存服务
func (c *MemoryCache) Stop() {
	close(c.stopChan)
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(prefix string, data interface{}) (string, error) {
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化数据失败: %w", err)
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(jsonData)
	hashStr := hex.EncodeToString(hash[:])

	// 返回带前缀的缓存键
	return fmt.Sprintf("%s:%s", prefix, hashStr), nil
}
