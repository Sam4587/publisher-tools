# Cookie管理模块

## 概述

Cookie管理模块负责统一管理各平台的登录凭证，提供Cookie的持久化存储、自动刷新、加密保护等功能，确保用户登录状态的有效性和安全性。

## 目录

- [核心功能](#核心功能)
- [技术架构](#技术架构)
- [使用指南](#使用指南)
- [API参考](#api参考)
- [安全考虑](#安全考虑)
- [迁移指南](#迁移指南)

## 核心功能

### 1. Cookie持久化
- 自动保存登录后的Cookie信息
- 支持JSON格式存储
- 每个平台独立存储文件

### 2. 自动刷新
- 检测Cookie过期时间
- 自动触发重新登录流程
- 无缝续期用户体验

### 3. 加密保护
- 敏感信息加密存储
- 支持自定义加密密钥
- 防止本地文件泄露风险

### 4. 多平台隔离
- 各平台Cookie独立管理
- 避免跨平台污染
- 支持同时登录多个账号

## 技术架构

### 核心组件

```
cookies/
├── cookies.go           # Cookie管理器核心
├── storage.go           # 存储实现
├── encryption.go        # 加密解密
├── loader.go            # Cookie加载器
└── validator.go         # 有效性验证
```

### 存储结构

```json
{
  "creation_time": "2026-02-19T10:00:00Z",
  "last_access": "2026-02-19T15:30:00Z",
  "cookies": [
    {
      "name": "sessionid",
      "value": "encrypted_value",
      "domain": ".douyin.com",
      "path": "/",
      "expires": "2026-03-19T10:00:00Z",
      "secure": true,
      "httponly": true
    }
  ]
}
```

## 使用指南

### 基本使用

```go
import "github.com/monkeycode/publisher-core/cookies"

// 创建Cookie管理器
manager := cookies.NewManager(cookies.Config{
    StorageDir: "./cookies",
    Encrypt:    true,
    Key:        "your-encryption-key",
})

// 保存Cookie
cookies := []*http.Cookie{
    {
        Name:   "sessionid",
        Value:  "abc123",
        Domain: ".douyin.com",
        Path:   "/",
    },
}
err := manager.SaveCookies("douyin", cookies)

// 加载Cookie
loadedCookies, err := manager.LoadCookies("douyin")
if err != nil {
    // 处理加载失败
    log.Printf("加载Cookie失败: %v", err)
}
```

### 高级功能

#### 自动刷新机制
```go
// 设置自动刷新回调
manager.SetRefreshCallback(func(platform string) error {
    log.Printf("平台 %s Cookie即将过期，执行刷新", platform)
    
    // 执行重新登录逻辑
    result, err := publisher.Login(context.Background())
    if err != nil {
        return err
    }
    
    // 保存新的Cookie
    return manager.SaveCookies(platform, result.Cookies)
})

// 启用自动刷新检查
manager.EnableAutoRefresh(24 * time.Hour) // 提前24小时刷新
```

#### 批量操作
```go
// 批量保存多个平台Cookie
batchCookies := map[string][]*http.Cookie{
    "douyin":    douyinCookies,
    "toutiao":   toutiaoCookies,
    "xiaohongshu": xhsCookies,
}

for platform, cookies := range batchCookies {
    if err := manager.SaveCookies(platform, cookies); err != nil {
        log.Printf("保存 %s Cookie失败: %v", platform, err)
    }
}
```

## API参考

### CookieManager

```go
type Config struct {
    StorageDir string        // Cookie存储目录
    Encrypt    bool          // 是否加密存储
    Key        string        // 加密密钥
    Logger     *log.Logger   // 日志记录器
}

type CookieManager struct {
    // 包含配置和内部状态
}

// 创建新的Cookie管理器
func NewManager(config Config) *CookieManager

// 保存Cookie
func (cm *CookieManager) SaveCookies(platform string, cookies []*http.Cookie) error

// 加载Cookie
func (cm *CookieManager) LoadCookies(platform string) ([]*http.Cookie, error)

// 删除Cookie
func (cm *CookieManager) DeleteCookies(platform string) error

// 检查Cookie是否存在
func (cm *CookieManager) HasCookies(platform string) bool

// 获取所有平台列表
func (cm *CookieManager) ListPlatforms() []string
```

### Cookie实体

```go
type StoredCookie struct {
    CreationTime time.Time     `json:"creation_time"`
    LastAccess   time.Time     `json:"last_access"`
    Cookies      []CookieData  `json:"cookies"`
}

type CookieData struct {
    Name     string    `json:"name"`
    Value    string    `json:"value"`
    Domain   string    `json:"domain"`
    Path     string    `json:"path"`
    Expires  time.Time `json:"expires,omitempty"`
    Secure   bool      `json:"secure"`
    HttpOnly bool      `json:"httponly"`
}
```

## 安全考虑

### 1. 加密存储

```go
// 启用加密存储
config := cookies.Config{
    StorageDir: "./cookies",
    Encrypt:    true,
    Key:        os.Getenv("COOKIE_ENCRYPTION_KEY"), // 从环境变量获取
}

manager := cookies.NewManager(config)
```

### 2. 权限控制

```bash
# 设置适当的文件权限
chmod 700 ./cookies
chmod 600 ./cookies/*.json

# 确保只有应用用户可以访问
chown appuser:appgroup ./cookies
```

### 3. 敏感信息处理

```go
// 避免在日志中记录完整Cookie值
func logSafeCookie(cookie *http.Cookie) {
    // 只记录Cookie名称和其他非敏感信息
    log.Printf("Cookie[name=%s, domain=%s, expires=%v]", 
        cookie.Name, cookie.Domain, cookie.Expires)
}
```

### 4. 定期清理

```go
// 定期清理过期Cookie
func cleanupExpiredCookies(manager *CookieManager) {
    platforms := manager.ListPlatforms()
    for _, platform := range platforms {
        cookies, _ := manager.LoadCookies(platform)
        if len(cookies) == 0 {
            manager.DeleteCookies(platform) // 清理空文件
        }
    }
}

// 每天执行一次清理
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        cleanupExpiredCookies(manager)
    }
}()
```

## 迁移指南

### 从旧版本迁移

#### 1. 文件结构调整
```bash
# 旧结构
./douyin_cookies.json
./toutiao_cookies.txt
./xhs_session.data

# 新结构
./cookies/
├── douyin_cookies.json
├── toutiao_cookies.json
└── xiaohongshu_cookies.json
```

#### 2. 格式转换脚本
```go
// 转换旧格式Cookie文件
func migrateOldCookies() error {
    oldFiles := []string{"douyin_cookies.txt", "toutiao_session.data"}
    
    for _, file := range oldFiles {
        if _, err := os.Stat(file); err == nil {
            // 解析旧格式
            cookies, err := parseOldFormat(file)
            if err != nil {
                return err
            }
            
            // 保存为新格式
            platform := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
            platform = strings.Replace(platform, "_cookies", "", -1)
            platform = strings.Replace(platform, "_session", "", -1)
            
            if err := manager.SaveCookies(platform, cookies); err != nil {
                return err
            }
            
            // 删除旧文件
            os.Remove(file)
        }
    }
    return nil
}
```

#### 3. 数据验证
```go
// 验证迁移后的Cookie有效性
func validateMigratedCookies() error {
    platforms := []string{"douyin", "toutiao", "xiaohongshu"}
    
    for _, platform := range platforms {
        cookies, err := manager.LoadCookies(platform)
        if err != nil {
            return fmt.Errorf("加载 %s Cookie失败: %v", platform, err)
        }
        
        if len(cookies) == 0 {
            log.Printf("警告: %s 没有有效的Cookie", platform)
            continue
        }
        
        // 验证关键Cookie是否存在
        hasSession := false
        for _, cookie := range cookies {
            if strings.Contains(strings.ToLower(cookie.Name), "session") ||
               strings.Contains(strings.ToLower(cookie.Name), "token") {
                hasSession = true
                break
            }
        }
        
        if !hasSession {
            log.Printf("警告: %s 缺少会话Cookie", platform)
        }
    }
    return nil
}
```

## 最佳实践

### 1. 错误处理
```go
// 妥善处理Cookie相关错误
func handleWithCookies(platform string, fn func([]*http.Cookie) error) error {
    cookies, err := manager.LoadCookies(platform)
    if err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("请先登录 %s 平台", platform)
        }
        return fmt.Errorf("加载Cookie失败: %v", err)
    }
    
    if len(cookies) == 0 {
        return fmt.Errorf("%s Cookie为空，请重新登录", platform)
    }
    
    return fn(cookies)
}
```

### 2. 日志记录
```go
// 记录Cookie操作日志
func logCookieOperation(operation, platform string, count int) {
    logger.WithFields(logrus.Fields{
        "operation": operation,
        "platform":  platform,
        "count":     count,
    }).Info("Cookie操作")
}
```

### 3. 监控告警
```go
// 监控Cookie状态
func monitorCookieHealth() {
    platforms := manager.ListPlatforms()
    for _, platform := range platforms {
        cookies, _ := manager.LoadCookies(platform)
        if len(cookies) == 0 {
            alertManager.SendAlert(fmt.Sprintf("%s Cookie缺失", platform))
        }
        
        // 检查是否即将过期
        for _, cookie := range cookies {
            if cookie.Expires.Sub(time.Now()) < 24*time.Hour {
                alertManager.SendAlert(fmt.Sprintf("%s Cookie即将过期", platform))
            }
        }
    }
}
```

## 相关文档

- [平台适配器文档](../adapters/)
- [浏览器自动化文档](../browser/)
- [安全配置指南](../../guides/security/)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0