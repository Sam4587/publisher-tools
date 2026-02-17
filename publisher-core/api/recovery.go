package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ErrorHandler 错误处理�?
type ErrorHandler struct {
	mu           sync.RWMutex
	errorCounts  map[string]int
	lastErrors   map[string]time.Time
	maxErrors    int
	windowPeriod time.Duration
}

// NewErrorHandler 创建错误处理�?
func NewErrorHandler(maxErrors int, windowPeriod time.Duration) *ErrorHandler {
	return &ErrorHandler{
		errorCounts:  make(map[string]int),
		lastErrors:   make(map[string]time.Time),
		maxErrors:    maxErrors,
		windowPeriod: windowPeriod,
	}
}

// RecordError 记录错误
func (h *ErrorHandler) RecordError(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	errKey := err.Error()
	now := time.Now()

	// 清理过期错误
	for key, lastTime := range h.lastErrors {
		if now.Sub(lastTime) > h.windowPeriod {
			delete(h.errorCounts, key)
			delete(h.lastErrors, key)
		}
	}

	h.errorCounts[errKey]++
	h.lastErrors[errKey] = now

	// 检查是否达到阈�?
	if h.errorCounts[errKey] >= h.maxErrors {
		logrus.Errorf("Error threshold reached: %s (count: %d)", errKey, h.errorCounts[errKey])
	}
}

// IsCircuitOpen 检查熔断器是否打开
func (h *ErrorHandler) IsCircuitOpen(err error) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	errKey := err.Error()
	count := h.errorCounts[errKey]
	
	return count >= h.maxErrors
}

// RecoveryMiddleware 恢复中间�?
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				stack := debug.Stack()
				logrus.Errorf("Panic recovered: %v
Stack: %s", recovered, string(stack))

				err := fmt.Errorf("internal server error: %v", recovered)
				jsonError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware 超时中间�?
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan bool, 1)
			go func() {
				next.ServeHTTP(w, r)
				done <- true
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				logrus.Warnf("Request timeout: %s %s", r.Method, r.URL.Path)
				jsonError(w, "TIMEOUT", "Request timeout", http.StatusRequestTimeout)
			}
		})
	}
}

// RateLimitMiddleware 限流中间�?
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	maxReqs  int
	window   time.Duration
}

func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		maxReqs:  maxReqs,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// 清理过期请求
	var validRequests []time.Time
	for _, reqTime := range rl.requests[ip] {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	if len(validRequests) >= rl.maxReqs {
		rl.requests[ip] = validRequests
		return false
	}

	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIPAddress(r)

			if !limiter.Allow(ip) {
				logrus.Warnf("Rate limit exceeded for IP: %s", ip)
				jsonError(w, "RATE_LIMIT_EXCEEDED", "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GracefulShutdown 优雅关闭
func GracefulShutdown(server *http.Server, shutdownTimeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server shutdown error: %v", err)
	}

	logrus.Info("Server stopped")
}

// Retry 重试机制
func Retry(ctx context.Context, maxRetries int, delay time.Duration, fn func() error) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			logrus.Warnf("Retry %d/%d failed: %v", i+1, maxRetries, err)

			// 检查上下文是否已取�?
			if ctx.Err() != nil {
				return ctx.Err()
			}

			time.Sleep(delay)
			continue
		}

		return nil
	}

	return errors.Wrap(lastErr, "max retries exceeded")
}

// getIPAddress 获取客户端IP地址
func getIPAddress(r *http.Request) string {
	// 尝试�?X-Forwarded-For 获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// 尝试�?X-Real-IP 获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return r.RemoteAddr
}

// jsonError 辅助函数
func jsonError(w http.ResponseWriter, code string, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    false,
		"error":      message,
		"error_code": code,
		"timestamp":  time.Now().Unix(),
	})
}
