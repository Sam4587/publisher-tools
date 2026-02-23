package util

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// 验证错误
var (
	ErrEmptyString        = errors.New("string cannot be empty")
	ErrStringTooShort     = errors.New("string is too short")
	ErrStringTooLong      = errors.New("string is too long")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrInvalidURL         = errors.New("invalid URL format")
	ErrInvalidFormat      = errors.New("invalid format")
	ErrContainsInvalidChar = errors.New("contains invalid characters")
	ErrContainsSQLInjection = errors.New("potential SQL injection detected")
	ErrContainsXSS        = errors.New("potential XSS attack detected")
)

// StringValidator 字符串验证器
type StringValidator struct {
	value  string
	errors []error
}

// NewStringValidator 创建字符串验证器
func NewStringValidator(value string) *StringValidator {
	return &StringValidator{
		value:  value,
		errors: make([]error, 0),
	}
}

// Required 验证字符串不为空
func (v *StringValidator) Required() *StringValidator {
	if strings.TrimSpace(v.value) == "" {
		v.errors = append(v.errors, ErrEmptyString)
	}
	return v
}

// MinLength 验证最小长度
func (v *StringValidator) MinLength(min int) *StringValidator {
	if len(v.value) < min {
		v.errors = append(v.errors, fmt.Errorf("%w: minimum length is %d", ErrStringTooShort, min))
	}
	return v
}

// MaxLength 验证最大长度
func (v *StringValidator) MaxLength(max int) *StringValidator {
	if len(v.value) > max {
		v.errors = append(v.errors, fmt.Errorf("%w: maximum length is %d", ErrStringTooLong, max))
	}
	return v
}

// Email 验证邮箱格式
func (v *StringValidator) Email() *StringValidator {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(v.value) {
		v.errors = append(v.errors, ErrInvalidEmail)
	}
	return v
}

// URL 验证URL格式
func (v *StringValidator) URL() *StringValidator {
	if _, err := url.ParseRequestURI(v.value); err != nil {
		v.errors = append(v.errors, ErrInvalidURL)
	}
	return v
}

// AlphaNumeric 验证只包含字母和数字
func (v *StringValidator) AlphaNumeric() *StringValidator {
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(v.value) {
		v.errors = append(v.errors, ErrContainsInvalidChar)
	}
	return v
}

// NoSQLInjection 检测SQL注入
func (v *StringValidator) NoSQLInjection() *StringValidator {
	sqlPatterns := []string{
		`(?i)\b(select|insert|update|delete|drop|create|alter|truncate|union|exec|execute)\b`,
		`['";]`,
		`--`,
		`/\*.*\*/`,
	}

	lowerValue := strings.ToLower(v.value)
	for _, pattern := range sqlPatterns {
		if regexp.MustCompile(pattern).MatchString(lowerValue) {
			v.errors = append(v.errors, ErrContainsSQLInjection)
			return v
		}
	}
	return v
}

// NoXSS 检测XSS攻击
func (v *StringValidator) NoXSS() *StringValidator {
	xssPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`javascript:`,
		`on\w+\s*=`,
		`<iframe`,
		`<object`,
		`<embed`,
	}

	lowerValue := strings.ToLower(v.value)
	for _, pattern := range xssPatterns {
		if regexp.MustCompile(pattern).MatchString(lowerValue) {
			v.errors = append(v.errors, ErrContainsXSS)
			return v
		}
	}
	return v
}

// Custom 自定义验证
func (v *StringValidator) Custom(fn func(string) error) *StringValidator {
	if err := fn(v.value); err != nil {
		v.errors = append(v.errors, err)
	}
	return v
}

// Validate 执行验证并返回错误
func (v *StringValidator) Validate() error {
	if len(v.errors) > 0 {
		return v.errors[0]
	}
	return nil
}

// ValidateAll 执行验证并返回所有错误
func (v *StringValidator) ValidateAll() error {
	if len(v.errors) > 0 {
		return fmt.Errorf("validation failed: %v", v.errors)
	}
	return nil
}

// IntValidator 整数验证器
type IntValidator struct {
	value  int
	errors []error
}

// NewIntValidator 创建整数验证器
func NewIntValidator(value int) *IntValidator {
	return &IntValidator{
		value:  value,
		errors: make([]error, 0),
	}
}

// Min 验证最小值
func (v *IntValidator) Min(min int) *IntValidator {
	if v.value < min {
		v.errors = append(v.errors, fmt.Errorf("value is less than minimum %d", min))
	}
	return v
}

// Max 验证最大值
func (v *IntValidator) Max(max int) *IntValidator {
	if v.value > max {
		v.errors = append(v.errors, fmt.Errorf("value is greater than maximum %d", max))
	}
	return v
}

// Positive 验证正数
func (v *IntValidator) Positive() *IntValidator {
	if v.value <= 0 {
		v.errors = append(v.errors, errors.New("value must be positive"))
	}
	return v
}

// NonNegative 验证非负数
func (v *IntValidator) NonNegative() *IntValidator {
	if v.value < 0 {
		v.errors = append(v.errors, errors.New("value must be non-negative"))
	}
	return v
}

// Validate 执行验证并返回错误
func (v *IntValidator) Validate() error {
	if len(v.errors) > 0 {
		return v.errors[0]
	}
	return nil
}

// PasswordValidator 密码验证器
type PasswordValidator struct {
	password string
	errors   []error
}

// NewPasswordValidator 创建密码验证器
func NewPasswordValidator(password string) *PasswordValidator {
	return &PasswordValidator{
		password: password,
		errors:   make([]error, 0),
	}
}

// MinLength 验证最小长度
func (v *PasswordValidator) MinLength(min int) *PasswordValidator {
	if len(v.password) < min {
		v.errors = append(v.errors, fmt.Errorf("password must be at least %d characters", min))
	}
	return v
}

// HasUpperCase 验证包含大写字母
func (v *PasswordValidator) HasUpperCase() *PasswordValidator {
	hasUpper := false
	for _, r := range v.password {
		if unicode.IsUpper(r) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		v.errors = append(v.errors, errors.New("password must contain at least one uppercase letter"))
	}
	return v
}

// HasLowerCase 验证包含小写字母
func (v *PasswordValidator) HasLowerCase() *PasswordValidator {
	hasLower := false
	for _, r := range v.password {
		if unicode.IsLower(r) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		v.errors = append(v.errors, errors.New("password must contain at least one lowercase letter"))
	}
	return v
}

// HasDigit 验证包含数字
func (v *PasswordValidator) HasDigit() *PasswordValidator {
	hasDigit := false
	for _, r := range v.password {
		if unicode.IsDigit(r) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		v.errors = append(v.errors, errors.New("password must contain at least one digit"))
	}
	return v
}

// HasSpecialChar 验证包含特殊字符
func (v *PasswordValidator) HasSpecialChar() *PasswordValidator {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	hasSpecial := false
	for _, r := range v.password {
		if strings.ContainsRune(specialChars, r) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial {
		v.errors = append(v.errors, errors.New("password must contain at least one special character"))
	}
	return v
}

// Validate 执行验证并返回错误
func (v *PasswordValidator) Validate() error {
	if len(v.errors) > 0 {
		return v.errors[0]
	}
	return nil
}

// SanitizeString 清理字符串中的危险字符
func SanitizeString(input string) string {
	// 移除潜在的XSS攻击向量
	input = regexp.MustCompile(`<script[^>]*>.*?</script>`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`javascript:`).ReplaceAllString(input, "")
	input = regexp.MustCompile(`on\w+\s*=`).ReplaceAllString(input, "")

	// 移除潜在的SQL注入字符
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "\"", "\"\"")

	return strings.TrimSpace(input)
}

// ValidatePlatform 验证平台名称
func ValidatePlatform(platform string) error {
	validPlatforms := []string{"douyin", "toutiao", "xiaohongshu", "bilibili"}
	for _, p := range validPlatforms {
		if platform == p {
			return nil
		}
	}
	return fmt.Errorf("invalid platform: %s", platform)
}
