package cookies

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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
		return fmt.Errorf("序列化 cookie 失败: %w", err)
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

func CookieExists(platform string) bool {
	path := filepath.Join("./cookies", fmt.Sprintf("%s_cookies.json", platform))
	_, err := os.Stat(path)
	return err == nil
}
