package middleware

import (
	"net/http"
	"runtime/debug"
	"sync"

	"publisher-core/api/response"
	"github.com/sirupsen/logrus"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				logrus.Errorf("Panic recovered: %v\n%s", err, string(stack))
				
				// 返回500错误
				response.JSONError(w, "INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// ErrorHandlerMiddleware 错误处理中间件
type ErrorHandlerMiddleware struct {
	errorHandler func(error)
}

// NewErrorHandlerMiddleware 创建错误处理中间件
func NewErrorHandlerMiddleware(errorHandler func(error)) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{
		errorHandler: errorHandler,
	}
}

func (m *ErrorHandlerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		// 在这里可以添加全局错误处理逻辑
		if err := recover(); err != nil {
			if m.errorHandler != nil {
				m.errorHandler(err)
			}
		}
	}()
	
	next(w, r)
}

// ErrorCounter 错误计数器
type ErrorCounter struct {
	counts map[string]int
	mu     sync.RWMutex
}

// NewErrorCounter 创建错误计数器
func NewErrorCounter() *ErrorCounter {
	return &ErrorCounter{
		counts: make(map[string]int),
	}
}

// RecordError 记录错误
func (c *ErrorCounter) RecordError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	errorType := getErrorType(err)
	c.counts[errorType]++
	
	// 定期清理过期错误
	if len(c.counts) > 1000 {
		c.cleanupOldErrors()
	}
}

// GetCounts 获取错误计数
func (c *ErrorCounter) GetCounts() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result := make(map[string]int)
	for k, v := range c.counts {
		result[k] = v
	}
	return result
}

// cleanupOldErrors 清理旧错误
func (c *ErrorCounter) cleanupOldErrors() {
	// 简单实现: 清理计数小于5的错误
	for k, v := range c.counts {
		if v < 5 {
			delete(c.counts, k)
		}
	}
}

// getErrorType 获取错误类型
func getErrorType(err error) string {
	if err == nil {
		return "unknown"
	}
	return err.Error()
}
