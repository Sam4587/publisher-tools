package cookies

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
		return fmt.Errorf("failed to create cookie directory: %w", err)
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
		return fmt.Errorf("failed to marshal cookies: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to save cookie file: %w", err)
	}

	return nil
}

func (m *Manager) Load(ctx context.Context, platform string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.getCookiePath(platform)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	var cookieMap map[string]map[string]interface{}
	if err := json.Unmarshal(data, &cookieMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	result := make(map[string]string)
	for name, cookie := range cookieMap {
		if value, ok := cookie["value"].(string); ok {
			result[name] = value
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
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	var cookieMap map[string]map[string]interface{}
	if err := json.Unmarshal(data, &cookieMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	var result []*proto.NetworkCookieParam
	for _, cookie := range cookieMap {
		param := &proto.NetworkCookieParam{
			Name:  cookie["name"].(string),
			Value: cookie["value"].(string),
		}
		if domain != "" {
			param.Domain = domain
		}
		if d, ok := cookie["domain"].(string); ok && d != "" {
			param.Domain = d
		}
		if p, ok := cookie["path"].(string); ok {
			param.Path = p
		}
		result = append(result, param)
	}

	return result, nil
}

func (m *Manager) Delete(ctx context.Context, platform string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.getCookiePath(platform)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cookie file: %w", err)
	}

	return nil
}

func (m *Manager) List(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files, err := os.ReadDir(m.cookieDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list cookie directory: %w", err)
	}

	var platforms []string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if len(name) > 12 && name[len(name)-12:] == "_cookies.json" {
				platforms = append(platforms, name[:len(name)-12])
			}
		}
	}

	return platforms, nil
}

func ExtractCookies(cookies []*proto.NetworkCookie, keys []string) []*proto.NetworkCookie {
	var result []*proto.NetworkCookie
	keySet := make(map[string]bool)
	for _, k := range keys {
		keySet[k] = true
	}

	for _, c := range cookies {
		if keySet[c.Name] {
			result = append(result, c)
		}
	}

	return result
}

type CookieJar struct {
	mu      sync.RWMutex
	cookies map[string]map[string]*proto.NetworkCookie
}

func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make(map[string]map[string]*proto.NetworkCookie),
	}
}

func (j *CookieJar) SetCookies(urlStr string, cookies []*proto.NetworkCookie) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.cookies[urlStr] == nil {
		j.cookies[urlStr] = make(map[string]*proto.NetworkCookie)
	}

	for _, c := range cookies {
		j.cookies[urlStr][c.Name] = c
	}
}

func (j *CookieJar) Cookies(urlStr string) []*proto.NetworkCookie {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var result []*proto.NetworkCookie
	if cookies, ok := j.cookies[urlStr]; ok {
		for _, c := range cookies {
			result = append(result, c)
		}
	}

	return result
}

func (j *CookieJar) Clear() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cookies = make(map[string]map[string]*proto.NetworkCookie)
}

type Cookie struct {
	Name     string
	Value    string
	Domain   string
	Path     string
	Expires  time.Time
	Secure   bool
	HttpOnly bool
}
