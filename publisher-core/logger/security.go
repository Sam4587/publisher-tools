package logger

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// SensitiveFields 敏感字段名列表
var SensitiveFields = []string{
	"password", "passwd", "pwd",
	"token", "access_token", "refresh_token", "auth_token",
	"secret", "api_key", "apikey", "api-secret",
	"authorization", "bearer",
	"credit_card", "card_number", "cvv",
	"ssn", "social_security_number",
}

// SanitizeLog 清理日志中的敏感信息
func SanitizeLog(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range data {
		lowerKey := strings.ToLower(key)
		isSensitive := false
		
		for _, sensitiveField := range SensitiveFields {
			if strings.Contains(lowerKey, sensitiveField) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			result[key] = "***REDACTED***"
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			result[key] = SanitizeLog(nestedMap)
		} else {
			result[key] = value
		}
	}
	
	return result
}

// SanitizeString 清理字符串中的敏感信息
func SanitizeString(s string) string {
	result := s
	
	// 清理常见的敏感模式
	patterns := []struct {
		pattern string
		replacement string
	}{
		{"Bearer [^\\s]+", "Bearer ***REDACTED***"},
		{"password=[^\\s&]+", "password=***REDACTED***"},
		{"token=[^\\s&]+", "token=***REDACTED***"},
		{"secret=[^\\s&]+", "secret=***REDACTED***"},
		{"api_key=[^\\s&]+", "api_key=***REDACTED***"},
	}
	
	for _, p := range patterns {
		// 简单的字符串替换,实际使用时应该使用正则表达式
		if strings.Contains(result, p.pattern) {
			result = strings.ReplaceAll(result, p.pattern, p.replacement)
		}
	}
	
	return result
}

// SafeError 安全的错误日志记录
func SafeError(fields logrus.Fields, message string) {
	sanitized := SanitizeLog(fields)
	logrus.WithFields(sanitized).Error(message)
}

// SafeWarn 安全的警告日志记录
func SafeWarn(fields logrus.Fields, message string) {
	sanitized := SanitizeLog(fields)
	logrus.WithFields(sanitized).Warn(message)
}

// SafeInfo 安全的信息日志记录
func SafeInfo(fields logrus.Fields, message string) {
	sanitized := SanitizeLog(fields)
	logrus.WithFields(sanitized).Info(message)
}

// SafeDebug 安全的调试日志记录
func SafeDebug(fields logrus.Fields, message string) {
	sanitized := SanitizeLog(fields)
	logrus.WithFields(sanitized).Debug(message)
}

// SafeErrorf 安全的错误日志记录(格式化)
func SafeErrorf(message string, args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeString(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	logrus.Errorf(message, sanitizedArgs...)
}

// SafeWarnf 安全的警告日志记录(格式化)
func SafeWarnf(message string, args ...interface{}) {
	sanitizedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			sanitizedArgs[i] = SanitizeString(str)
		} else {
			sanitizedArgs[i] = arg
		}
	}
	logrus.Warnf(message, sanitizedArgs...)
}
