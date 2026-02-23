package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrorResponse 标准错误响应
type ErrorResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error"`
	ErrorCode string `json:"error_code,omitempty"`
	Timestamp int64  `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

// AppError 应用错误
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Internal   error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// 预定义错误
var (
	ErrBadRequest     = &AppError{Code: "BAD_REQUEST", Message: "Bad request", StatusCode: http.StatusBadRequest}
	ErrUnauthorized   = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", StatusCode: http.StatusUnauthorized}
	ErrForbidden      = &AppError{Code: "FORBIDDEN", Message: "Forbidden", StatusCode: http.StatusForbidden}
	ErrNotFound       = &AppError{Code: "NOT_FOUND", Message: "Resource not found", StatusCode: http.StatusNotFound}
	ErrConflict       = &AppError{Code: "CONFLICT", Message: "Resource conflict", StatusCode: http.StatusConflict}
	ErrInternalServer = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", StatusCode: http.StatusInternalServerError}
	ErrServiceUnavailable = &AppError{Code: "SERVICE_UNAVAILABLE", Message: "Service unavailable", StatusCode: http.StatusServiceUnavailable}
	ErrRateLimitExceeded = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "Rate limit exceeded", StatusCode: http.StatusTooManyRequests}
	ErrValidationFailed = &AppError{Code: "VALIDATION_FAILED", Message: "Validation failed", StatusCode: http.StatusBadRequest}
)

// NewAppError 创建应用错误
func NewAppError(code string, message string, statusCode int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// WrapError 包装错误
func WrapError(err error, code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Internal:   err,
	}
}

// ErrorHandlerMiddleware 错误处理中间件
func ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				// 记录panic信息
				stack := debug.Stack()
				logrus.Errorf("Panic recovered: %v\nStack: %s", recovered, string(stack))

				// 返回内部服务器错误
				respondWithError(w, ErrInternalServer, r)
			}
		}()

		// 创建响应记录器以捕获响应状态
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		// 如果状态码是错误状态，确保响应格式统一
		if recorder.statusCode >= 400 && recorder.statusCode < 600 {
			// 确保响应头已设置
			if recorder.Header().Get("Content-Type") == "" {
				recorder.Header().Set("Content-Type", "application/json")
			}
		}
	})
}

// responseRecorder 用于记录响应状态码
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// respondWithError 返回错误响应
func respondWithError(w http.ResponseWriter, err error, r *http.Request) {
	var appErr *AppError
	var statusCode int
	var code string
	var message string

	// 判断错误类型
	if errors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		code = appErr.Code
		message = appErr.Message
	} else {
		// 未知错误，使用内部服务器错误
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
		message = "Internal server error"
	}

	// 记录错误日志
	if statusCode >= 500 {
		logrus.Errorf("Server error: %v", err)
	} else if statusCode >= 400 && statusCode < 500 {
		logrus.Warnf("Client error: %v", err)
	}

	// 创建错误响应
	response := ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorCode: code,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(r),
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 写入响应
	json.NewEncoder(w).Encode(response)
}

// getRequestID 获取请求ID
func getRequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	return r.Context().Value("request_id").(string)
}

// RespondWithJSON 返回JSON响应
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
		"timestamp": time.Now().Unix(),
	})
}

// RespondWithError 返回错误响应
func RespondWithError(w http.ResponseWriter, err error) {
	var appErr *AppError
	var statusCode int
	var code string
	var message string

	// 判断错误类型
	if errors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		code = appErr.Code
		message = appErr.Message
	} else {
		// 未知错误，使用内部服务器错误
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
		message = "Internal server error"
	}

	// 记录错误日志
	if statusCode >= 500 {
		logrus.Errorf("Server error: %v", err)
	} else if statusCode >= 400 && statusCode < 500 {
		logrus.Warnf("Client error: %v", err)
	}

	// 创建错误响应
	response := ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorCode: code,
		Timestamp: time.Now().Unix(),
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 写入响应
	json.NewEncoder(w).Encode(response)
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RespondWithValidationErrors 返回验证错误响应
func RespondWithValidationErrors(w http.ResponseWriter, errors []ValidationError) error {
	response := map[string]interface{}{
		"success":   false,
		"error":     "Validation failed",
		"error_code": "VALIDATION_FAILED",
		"errors":    errors,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(response)
}

// NotFoundHandler 404处理器
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithError(w, ErrNotFound)
}

// MethodNotAllowedHandler 405处理器
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithError(w, &AppError{
		Code:       "METHOD_NOT_ALLOWED",
		Message:    fmt.Sprintf("Method %s not allowed", r.Method),
		StatusCode: http.StatusMethodNotAllowed,
	})
}
