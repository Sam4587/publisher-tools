package cookies

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-rod/rod/lib/proto"
)

var (
	DouyinCookieKeys = []string{
		"tt_webid",
		"passport_auth",
		"csrf_token",
		"ttcid",
		"sid_guard",
		"uid_tt",
		"sid_tt",
	}

	ToutiaoCookieKeys = []string{
		"sessionid",
		"passport_auth",
		"tt_token",
		"tt_webid",
		"sso_uid_tt",
		"sso_uid_tt_ss",
	}

	XiaohongshuCookieKeys = []string{
		"web_session",
		"webId",
		"web_session_sig",
		"a1",
		"websectiga",
	}
)

type Manager struct {
	mu        sync.RWMutex
	cookieDir string
}

func NewManager(cookieDir string) *Manager {
	return &Manager{
		cookieDir: cookieDir,
	}
}

func (m *Manager) getCookiePath(platform string) string {
	return filepath.Join(m.cookieDir, fmt.Sprintf("%s_cookies.json", platform))
}

func (m *Manager) Exists(ctx context.Context, platform string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.getCookiePath(platform)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *Manager) Save(ctx context.Context, platform string, cookies []*proto.NetworkCookie) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.cookieDir, 0755); err != nil {
		return fmt.Errorf("创建 cookie 目录失败: %w", err)
	}

	path := m.getCookiePath(platform)

	cookieMap := make(map[string]interface{})
	for _, c := range cookies {
		cookieMap[c.Name] = map[string]interface{}{
			"name":     c.Name,
			"value":    c.Value,
			"domain":   c.Domain,
			"path":     c.Path,
			"expires":  c.Expires,
			"httpOnly": c.HTTPOnly,
			"secure":   c.Secure,
			"sameSite": c.SameSite,
		}
	}

	data, err := json.MarshalIndent(cookieMap, "", "  ")
	if err != nil {
		return fmt.Errorf("序列�?cookie 失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("保存 cookie 文件失败: %w", err)
	}

	return nil
}

func (m *Manager) Load(ctx context.Context, platform string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.getCookiePath(platform)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 cookie 文件失败: %w", err)
	}

	var cookieMap map[string]map[string]interface{}
	if err := json.Unmarshal(data, &cookieMap); err != nil {
		return nil, fmt.Errorf("解析 cookie 文件失败: %w", err)
	}

	result := make(map[string]string)
	for name, c := range cookieMap {
		if val, ok := c["value"].(string); ok {
			result[name] = val
		}
	}

	return result, nil
}

func (m *Manager) LoadAsProto(ctx context.Context, platform string, domain string) ([]*proto.NetworkCookieParam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.getCookiePath(platform)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取 cookie 文件失败: %w", err)
	}

	var cookieMap map[string]map[string]interface{}
	if err := json.Unmarshal(data, &cookieMap); err != nil {
		return nil, fmt.Errorf("解析 cookie 文件失败: %w", err)
	}

	var result []*proto.NetworkCookieParam
	for _, c := range cookieMap {
		cookie := &proto.NetworkCookieParam{
			Name:     c["name"].(string),
			Value:    c["value"].(string),
			Domain:   domain,
			Path:     "/",
			HTTPOnly: false,
			Secure:   false,
		}

		if domainVal, ok := c["domain"].(string); ok && domainVal != "" {
			cookie.Domain = domainVal
		}
		if pathVal, ok := c["path"].(string); ok {
			cookie.Path = pathVal
		}
		if httpOnly, ok := c["httpOnly"].(bool); ok {
			cookie.HTTPOnly = httpOnly
		}
		if secure, ok := c["secure"].(bool); ok {
			cookie.Secure = secure
		}

		result = append(result, cookie)
	}

	return result, nil
}

func (m *Manager) Delete(ctx context.Context, platform string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.getCookiePath(platform)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ExtractCookies(cookies []*proto.NetworkCookie, keys []string) map[string]string {
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	result := make(map[string]string)
	for _, c := range cookies {
		if keySet[c.Name] {
			result[c.Name] = c.Value
		}
	}
	return result
}

func InitCookieDir() error {
	return os.MkdirAll("./cookies", 0755)
}

func SaveCookies(cookies map[string]string, platform string) error {
	if err := InitCookieDir(); err != nil {
		return err
	}

	path := filepath.Join("./cookies", fmt.Sprintf("%s_cookies.json", platform))

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return fmt.Errorf("序列�?cookie 失败: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

func LoadCookies(platform string) (map[string]string, error) {
	path := filepath.Join("./cookies", fmt.Sprintf("%s_cookies.json", platform))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
