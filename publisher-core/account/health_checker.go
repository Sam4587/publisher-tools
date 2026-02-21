// Package account 提供各平台健康检查器实现
package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"publisher-core/database"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// =====================================================
// 抖音健康检查器
// =====================================================

// DouyinHealthChecker 抖音健康检查器
type DouyinHealthChecker struct {
	client *http.Client
}

// NewDouyinHealthChecker 创建抖音健康检查器
func NewDouyinHealthChecker() *DouyinHealthChecker {
	return &DouyinHealthChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Check 执行健康检查
func (c *DouyinHealthChecker) Check(ctx context.Context, account *database.PlatformAccount) error {
	// 解析Cookie
	cookies, err := c.parseCookies(account.CookieData)
	if err != nil {
		return fmt.Errorf("解析Cookie失败: %w", err)
	}

	// 检查必要Cookie是否存在
	requiredKeys := []string{"tt_webid", "passport_auth"}
	for _, key := range requiredKeys {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("缺少必要Cookie: %s", key)
		}
	}

	// 发送请求验证登录状态
	req, err := http.NewRequestWithContext(ctx, "GET", "https://creator.douyin.com/creator-micro/home/", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置Cookie
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("账号未授权或已过期")
	}

	// 读取响应内容检查登录状态
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// 检查是否包含登录相关内容
	if strings.Contains(bodyStr, "请登录") || strings.Contains(bodyStr, "login") {
		return fmt.Errorf("账号需要重新登录")
	}

	return nil
}

func (c *DouyinHealthChecker) parseCookies(cookieData string) (map[string]string, error) {
	// 尝试解析JSON格式
	var cookieMap map[string]interface{}
	if err := json.Unmarshal([]byte(cookieData), &cookieMap); err == nil {
		result := make(map[string]string)
		for k, v := range cookieMap {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result, nil
	}

	// 尝试解析标准Cookie字符串格式
	result := make(map[string]string)
	pairs := strings.Split(cookieData, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result, nil
}

// =====================================================
// 头条健康检查器
// =====================================================

// ToutiaoHealthChecker 头条健康检查器
type ToutiaoHealthChecker struct {
	client *http.Client
}

// NewToutiaoHealthChecker 创建头条健康检查器
func NewToutiaoHealthChecker() *ToutiaoHealthChecker {
	return &ToutiaoHealthChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Check 执行健康检查
func (c *ToutiaoHealthChecker) Check(ctx context.Context, account *database.PlatformAccount) error {
	cookies, err := c.parseCookies(account.CookieData)
	if err != nil {
		return fmt.Errorf("解析Cookie失败: %w", err)
	}

	// 检查必要Cookie
	requiredKeys := []string{"sessionid", "passport_auth"}
	for _, key := range requiredKeys {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("缺少必要Cookie: %s", key)
		}
	}

	// 验证登录状态
	req, err := http.NewRequestWithContext(ctx, "GET", "https://mp.toutiao.com/auth/", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("账号未授权或已过期")
	}

	return nil
}

func (c *ToutiaoHealthChecker) parseCookies(cookieData string) (map[string]string, error) {
	var cookieMap map[string]interface{}
	if err := json.Unmarshal([]byte(cookieData), &cookieMap); err == nil {
		result := make(map[string]string)
		for k, v := range cookieMap {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result, nil
	}

	result := make(map[string]string)
	pairs := strings.Split(cookieData, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

// =====================================================
// 小红书健康检查器
// =====================================================

// XiaohongshuHealthChecker 小红书健康检查器
type XiaohongshuHealthChecker struct {
	client *http.Client
}

// NewXiaohongshuHealthChecker 创建小红书健康检查器
func NewXiaohongshuHealthChecker() *XiaohongshuHealthChecker {
	return &XiaohongshuHealthChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Check 执行健康检查
func (c *XiaohongshuHealthChecker) Check(ctx context.Context, account *database.PlatformAccount) error {
	cookies, err := c.parseCookies(account.CookieData)
	if err != nil {
		return fmt.Errorf("解析Cookie失败: %w", err)
	}

	// 检查必要Cookie
	requiredKeys := []string{"web_session", "a1"}
	for _, key := range requiredKeys {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("缺少必要Cookie: %s", key)
		}
	}

	// 验证登录状态
	req, err := http.NewRequestWithContext(ctx, "GET", "https://creator.xiaohongshu.com/", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	// 设置必要的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("账号未授权或已过期")
	}

	return nil
}

func (c *XiaohongshuHealthChecker) parseCookies(cookieData string) (map[string]string, error) {
	var cookieMap map[string]interface{}
	if err := json.Unmarshal([]byte(cookieData), &cookieMap); err == nil {
		result := make(map[string]string)
		for k, v := range cookieMap {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result, nil
	}

	result := make(map[string]string)
	pairs := strings.Split(cookieData, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

// =====================================================
// B站健康检查器
// =====================================================

// BilibiliHealthChecker B站健康检查器
type BilibiliHealthChecker struct {
	client *http.Client
}

// NewBilibiliHealthChecker 创建B站健康检查器
func NewBilibiliHealthChecker() *BilibiliHealthChecker {
	return &BilibiliHealthChecker{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Check 执行健康检查
func (c *BilibiliHealthChecker) Check(ctx context.Context, account *database.PlatformAccount) error {
	cookies, err := c.parseCookies(account.CookieData)
	if err != nil {
		return fmt.Errorf("解析Cookie失败: %w", err)
	}

	// 检查必要Cookie
	requiredKeys := []string{"SESSDATA", "bili_jct"}
	for _, key := range requiredKeys {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("缺少必要Cookie: %s", key)
		}
	}

	// 验证登录状态 - 使用B站API
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析B站API响应
	var navResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			IsLogin bool `json:"isLogin"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &navResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if navResp.Code != 0 {
		return fmt.Errorf("API错误: %s", navResp.Message)
	}

	if !navResp.Data.IsLogin {
		return fmt.Errorf("账号未登录")
	}

	return nil
}

func (c *BilibiliHealthChecker) parseCookies(cookieData string) (map[string]string, error) {
	var cookieMap map[string]interface{}
	if err := json.Unmarshal([]byte(cookieData), &cookieMap); err == nil {
		result := make(map[string]string)
		for k, v := range cookieMap {
			if str, ok := v.(string); ok {
				result[k] = str
			}
		}
		return result, nil
	}

	result := make(map[string]string)
	pairs := strings.Split(cookieData, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

// =====================================================
// 健康检查器工厂
// =====================================================

// HealthCheckerFactory 健康检查器工厂
type HealthCheckerFactory struct {
	checkers map[string]HealthChecker
}

// NewHealthCheckerFactory 创建健康检查器工厂
func NewHealthCheckerFactory() *HealthCheckerFactory {
	factory := &HealthCheckerFactory{
		checkers: make(map[string]HealthChecker),
	}

	// 注册默认检查器
	factory.Register("douyin", NewDouyinHealthChecker())
	factory.Register("toutiao", NewToutiaoHealthChecker())
	factory.Register("xiaohongshu", NewXiaohongshuHealthChecker())
	factory.Register("bilibili", NewBilibiliHealthChecker())

	return factory
}

// Register 注册健康检查器
func (f *HealthCheckerFactory) Register(platform string, checker HealthChecker) {
	f.checkers[platform] = checker
	logrus.Infof("注册健康检查器: %s", platform)
}

// Get 获取健康检查器
func (f *HealthCheckerFactory) Get(platform string) (HealthChecker, bool) {
	checker, ok := f.checkers[platform]
	return checker, ok
}

// RegisterAllToService 将所有检查器注册到服务
func (f *HealthCheckerFactory) RegisterAllToService(service *AccountService) {
	for platform, checker := range f.checkers {
		service.RegisterHealthChecker(platform, checker)
	}
}

// =====================================================
// Cookie验证工具
// =====================================================

// CookieValidator Cookie验证器
type CookieValidator struct{}

// ValidateCookieFormat 验证Cookie格式
func (v *CookieValidator) ValidateCookieFormat(cookieStr string) error {
	if cookieStr == "" {
		return fmt.Errorf("Cookie不能为空")
	}

	// 尝试解析为JSON
	if strings.HasPrefix(cookieStr, "{") {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(cookieStr), &m); err != nil {
			return fmt.Errorf("JSON格式Cookie无效: %w", err)
		}
		return nil
	}

	// 尝试解析为标准Cookie格式
	pairs := strings.Split(cookieStr, ";")
	if len(pairs) == 0 {
		return fmt.Errorf("Cookie格式无效")
	}

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Cookie格式无效: %s", pair)
		}
		// 验证Cookie名称
		if parts[0] == "" {
			return fmt.Errorf("Cookie名称不能为空")
		}
	}

	return nil
}

// ParseCookieToMap 将Cookie字符串解析为Map
func (v *CookieValidator) ParseCookieToMap(cookieStr string) (map[string]string, error) {
	result := make(map[string]string)

	// 尝试JSON格式
	if strings.HasPrefix(cookieStr, "{") {
		var m map[string]string
		if err := json.Unmarshal([]byte(cookieStr), &m); err == nil {
			return m, nil
		}
	}

	// 标准Cookie格式
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			// URL解码值
			if decoded, err := url.QueryUnescape(parts[1]); err == nil {
				result[parts[0]] = decoded
			} else {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result, nil
}

// CheckRequiredCookies 检查必要Cookie是否存在
func (v *CookieValidator) CheckRequiredCookies(cookies map[string]string, required []string) error {
	for _, key := range required {
		if _, ok := cookies[key]; !ok {
			return fmt.Errorf("缺少必要Cookie: %s", key)
		}
	}
	return nil
}
