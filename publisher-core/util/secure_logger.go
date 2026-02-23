package util

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// 敏感字段列表
var sensitiveFields = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"authorization",
	"cookie",
	"session",
	"credit_card",
	"ssn",
	"jwt",
}

// 敏感字段正则表达式
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`(?i)secret["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`(?i)token["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`(?i)api[_-]?key["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`(?i)authorization["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`(?i)cookie["\s]*[:=]["\s]*[^"\s,}]+`),
	regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-._~+/]+=*`),
}

// SanitizeLogEntry 清理日志条目中的敏感信息
func SanitizeLogEntry(entry string) string {
	result := entry

	// 使用正则表达式替换敏感模式
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// 保留字段名，替换值为 ***
			parts := strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ":***"
			}
			parts = strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=***"
			}
			return "***"
		})
	}

	return result
}

// SanitizeURL 清理URL中的敏感参数
func SanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := u.Query()
	for key := range query {
		for _, field := range sensitiveFields {
			if strings.EqualFold(key, field) {
				query.Set(key, "***")
				break
			}
		}
	}

	u.RawQuery = query.Encode()
	return u.String()
}

// SanitizeMap 清理map中的敏感字段
func SanitizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		for _, field := range sensitiveFields {
			if strings.EqualFold(key, field) {
				result[key] = "***"
				break
			} else {
				result[key] = value
			}
		}
	}
	return result
}

// SecureLogger 安全日志记录器
type SecureLogger struct {
	logger *logrus.Logger
}

// NewSecureLogger 创建安全日志记录器
func NewSecureLogger(logger *logrus.Logger) *SecureLogger {
	return &SecureLogger{logger: logger}
}

// Info 记录信息日志（自动过滤敏感信息）
func (sl *SecureLogger) Info(args ...interface{}) {
	filteredArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			filteredArgs[i] = SanitizeLogEntry(str)
		} else {
			filteredArgs[i] = arg
		}
	}
	sl.logger.Info(filteredArgs...)
}

// Infof 记录格式化信息日志（自动过滤敏感信息）
func (sl *SecureLogger) Infof(format string, args ...interface{}) {
	sanitizedFormat := SanitizeLogEntry(format)
	sl.logger.Infof(sanitizedFormat, args...)
}

// Warn 记录警告日志（自动过滤敏感信息）
func (sl *SecureLogger) Warn(args ...interface{}) {
	filteredArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			filteredArgs[i] = SanitizeLogEntry(str)
		} else {
			filteredArgs[i] = arg
		}
	}
	sl.logger.Warn(filteredArgs...)
}

// Warnf 记录格式化警告日志（自动过滤敏感信息）
func (sl *SecureLogger) Warnf(format string, args ...interface{}) {
	sanitizedFormat := SanitizeLogEntry(format)
	sl.logger.Warnf(sanitizedFormat, args...)
}

// Error 记录错误日志（自动过滤敏感信息）
func (sl *SecureLogger) Error(args ...interface{}) {
	filteredArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			filteredArgs[i] = SanitizeLogEntry(str)
		} else {
			filteredArgs[i] = arg
		}
	}
	sl.logger.Error(filteredArgs...)
}

// Errorf 记录格式化错误日志（自动过滤敏感信息）
func (sl *SecureLogger) Errorf(format string, args ...interface{}) {
	sanitizedFormat := SanitizeLogEntry(format)
	sl.logger.Errorf(sanitizedFormat, args...)
}

// Debug 记录调试日志（自动过滤敏感信息）
func (sl *SecureLogger) Debug(args ...interface{}) {
	filteredArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			filteredArgs[i] = SanitizeLogEntry(str)
		} else {
			filteredArgs[i] = arg
		}
	}
	sl.logger.Debug(filteredArgs...)
}

// Debugf 记录格式化调试日志（自动过滤敏感信息）
func (sl *SecureLogger) Debugf(format string, args ...interface{}) {
	sanitizedFormat := SanitizeLogEntry(format)
	sl.logger.Debugf(sanitizedFormat, args...)
}
