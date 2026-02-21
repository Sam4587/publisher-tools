package api

import (
	"encoding/json"
	"net/http"
	"time"

	"publisher-tools/cache"
)

// CacheHandler 缓存管理处理器
type CacheHandler struct {
	cacheService *cache.AICacheService
}

// NewCacheHandler 创建缓存处理器
func NewCacheHandler(cacheService *cache.AICacheService) *CacheHandler {
	return &CacheHandler{
		cacheService: cacheService,
	}
}

// RegisterRoutes 注册路由
func (h *CacheHandler) RegisterRoutes(router interface{}) {
	// 类型断言以支持gorilla/mux
	if r, ok := router.(interface {
		HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route
	}); ok {
		r.HandleFunc("/api/v1/cache/stats", h.GetCacheStats).Methods("GET")
		r.HandleFunc("/api/v1/cache/keys", h.GetCacheKeys).Methods("GET")
		r.HandleFunc("/api/v1/cache/clear", h.ClearCache).Methods("POST")
		r.HandleFunc("/api/v1/cache/warmup", h.WarmupCache).Methods("POST")
		r.HandleFunc("/api/v1/cache/{key}", h.DeleteCache).Methods("DELETE")
	}
}

// GetCacheStats 获取缓存统计
func (h *CacheHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.cacheService.GetStats()

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetCacheKeys 获取所有缓存键
func (h *CacheHandler) GetCacheKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.cacheService.GetCacheKeys()

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"keys":      keys,
			"total":     len(keys),
		},
	})
}

// ClearCache 清空缓存
func (h *CacheHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.cacheService.Clear()

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "缓存已清空",
	})
}

// DeleteCache 删除指定缓存
func (h *CacheHandler) DeleteCache(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if err := h.cacheService.Delete(key); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "缓存已删除",
	})
}

// WarmupCache 预热缓存
func (h *CacheHandler) WarmupCache(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items map[string]interface{} `json:"items"`
		TTL   int                    `json:"ttl"` // 秒
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	ttl := time.Duration(req.TTL) * time.Second
	if ttl == 0 {
		ttl = 5 * time.Minute // 默认5分钟
	}

	if err := h.cacheService.Warmup(req.Items, ttl); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "缓存预热成功",
		"data": map[string]interface{}{
			"items_count": len(req.Items),
			"ttl_seconds": req.TTL,
		},
	})
}

// respondWithError 返回错误响应
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// respondWithJSON 返回JSON响应
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
