package api

import (
	"bytes"
	"net/http"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// Cache 缓存
type Cache struct {
	mu       sync.RWMutex
	items    map[string]*CacheItem
	maxSize  int
	disabled bool
}

// NewCache 创建缓存
func NewCache(maxSize int) *Cache {
	cache := &Cache{
		items:   make(map[string]*CacheItem),
		maxSize: maxSize,
	}

	// 启动后台清理协程
	go cache.cleanup()

	return cache
}

// Set 设置缓存
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	if c.disabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否超过最大大小
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}

	logrus.Debugf("Cache set: %s (TTL: %v)", key, ttl)
	return nil
}

// Get 获取缓存
func (c *Cache) Get(key string) (interface{}, bool) {
	if c.disabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	logrus.Debugf("Cache hit: %s", key)
	return item.Data, true
}

// Delete 删除缓存
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	logrus.Debugf("Cache deleted: %s", key)
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	logrus.Info("Cache cleared")
}

// Size 获取缓存大小
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats 获取缓存统计
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var expired int
	now := time.Now()

	for _, item := range c.items {
		if now.After(item.ExpiresAt) {
			expired++
		}
	}

	return CacheStats{
		TotalItems:   len(c.items),
		ExpiredItems: expired,
		MaxSize:      c.maxSize,
	}
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalItems   int `json:"total_items"`
	ExpiredItems int `json:"expired_items"`
	MaxSize      int `json:"max_size"`
}

// evictOldest 淘汰最旧的缓存
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
		logrus.Debugf("Evicted oldest cache: %s", oldestKey)
	}
}

// cleanup 定期清理过期缓存
func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// CacheMiddleware 缓存中间件
func CacheMiddleware(cache *Cache, ttl time.Duration, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 只缓存GET请求
			if r.Method != "GET" {
				next.ServeHTTP(w, r)
				return
			}

			key := keyFunc(r)

			// 尝试从缓存获取
			if data, found := cache.Get(key); found {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(data)
				return
			}

			// 创建响应记录器
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			// 如果响应成功，缓存结果
			if recorder.statusCode == http.StatusOK {
				var data interface{}
				if err := json.Unmarshal(recorder.body.Bytes(), &data); err == nil {
					cache.Set(key, data, ttl)
					w.Header().Set("X-Cache", "MISS")
				}
			}
		})
	}
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// DefaultCacheKeyFunc 默认缓存键生成函数
func DefaultCacheKeyFunc(r *http.Request) string {
	return r.Method + ":" + r.URL.Path + ":" + r.URL.RawQuery
}
