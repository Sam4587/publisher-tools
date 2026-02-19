# 浏览器自动化模块

## 概述

浏览器自动化模块基于 [go-rod/rod](https://github.com/go-rod/rod) 框架实现，负责模拟用户在各平台上的操作行为，包括登录、内容发布、数据抓取等。

## 目录

- [核心功能](#核心功能)
- [技术架构](#技术架构)
- [使用指南](#使用指南)
- [API参考](#api参考)
- [最佳实践](#最佳实践)
- [故障排除](#故障排除)

## 核心功能

### 1. 浏览器实例管理
- 自动启动和关闭浏览器实例
- 实例池管理，提高资源利用率
- 支持有头/无头模式切换

### 2. 页面操作
- 页面导航和等待
- 元素定位和交互
- 表单填写和提交
- 截图和PDF导出

### 3. 用户行为模拟
- 鼠标点击和拖拽
- 键盘输入和快捷键
- 滚动和页面操作
- 等待和超时控制

## 技术架构

### 核心组件

```
browser/
├── browser.go           # 浏览器管理器
├── navigation.go        # 页面导航
├── interaction.go       # 元素交互
├── screenshot.go        # 截图功能
└── pool.go             # 实例池管理
```

### 依赖关系
- **底层**: Chrome DevTools Protocol
- **框架**: go-rod/rod
- **上游**: 平台适配器
- **下游**: 任务管理系统

## 使用指南

### 基本使用

```go
import (
    "github.com/go-rod/rod"
    "github.com/go-rod/rod/lib/launcher"
)

// 创建浏览器实例
browser := rod.New().MustConnect()
defer browser.MustClose()

// 打开页面
page := browser.MustPage("https://www.example.com")

// 等待元素加载
page.MustWaitLoad()

// 元素交互
page.MustElement("#username").MustInput("用户名")
page.MustElement("#password").MustInput("密码")
page.MustElement("button[type=submit]").MustClick()

// 等待页面跳转
page.MustWaitNavigation()
```

### 高级功能

#### 等待条件
```go
// 等待元素可见
page.MustElement(".loading").MustWaitInvisible()

// 等待文本出现
page.MustElement("body").MustWaitText("登录成功")

// 自定义等待条件
page.MustWait(`() => document.querySelector('.ready').innerText === '完成'`)
```

#### 元素操作
```go
// 获取元素属性
text := page.MustElement("#title").MustProperty("textContent").Str()
href := page.MustElement("a").MustAttribute("href")

// 执行JavaScript
result := page.MustEval(`() => document.title`)

// 截图
page.MustScreenshot("screenshot.png")
page.MustElement("#chart").MustScreenshot("chart.png")
```

### 实例池使用

```go
// 创建实例池
pool := NewBrowserPool(BrowserPoolConfig{
    MaxInstances: 5,
    Headless:     true,
    Timeout:      30 * time.Second,
})

// 获取浏览器实例
browser, release := pool.Acquire()
defer release()

// 使用浏览器
page := browser.MustPage(url)
// ... 执行操作
```

## API参考

### BrowserManager

```go
type BrowserManager struct {
    // 配置选项
    Headless    bool
    UserDataDir string
    Proxy       string
}

// 创建浏览器管理器
func NewBrowserManager(config BrowserConfig) *BrowserManager

// 启动浏览器
func (bm *BrowserManager) Launch() (*rod.Browser, error)

// 关闭浏览器
func (bm *BrowserManager) Close(browser *rod.Browser) error
```

### Page操作

```go
// 导航
func Navigate(page *rod.Page, url string) error
func WaitLoad(page *rod.Page) error
func WaitNavigation(page *rod.Page) error

// 元素查找
func FindElement(page *rod.Page, selector string) (*rod.Element, error)
func FindElements(page *rod.Page, selector string) ([]*rod.Element, error)

// 元素交互
func Click(element *rod.Element) error
func Input(element *rod.Element, text string) error
func SelectOption(element *rod.Element, value string) error

// 等待条件
func WaitForVisible(element *rod.Element) error
func WaitForInvisible(element *rod.Element) error
func WaitForText(element *rod.Element, text string) error
```

## 最佳实践

### 1. 等待策略

```go
// ❌ 不好的做法 - 固定等待时间
time.Sleep(5 * time.Second)

// ✅ 好的做法 - 等待具体条件
page.MustElement("#button").MustWaitEnabled()
page.MustWaitStable()

// ✅ 更好的做法 - 带超时的等待
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
page.WithContext(ctx).MustElement("#dynamic-content").MustWaitVisible()
```

### 2. 错误处理

```go
// 重试机制
func retry(maxRetries int, fn func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        if err = fn(); err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return err
}

// 使用示例
err := retry(3, func() error {
    return page.MustElement("#retry-button").MustClick()
})
```

### 3. 资源管理

```go
// 正确的资源释放
func processWithBrowser(url string) error {
    browser := rod.New().MustConnect()
    defer browser.MustClose() // 确保释放
    
    page := browser.MustPage(url)
    defer page.MustClose() // 页面级别释放
    
    // 执行操作...
    return nil
}
```

### 4. 调试技巧

```go
// 开发模式启用详细日志
if os.Getenv("DEBUG") == "true" {
    launcher.New().Headless(false).MustLaunch()
    page.MustWindowMaximize()
    page.MustScreenshot("debug.png")
}

// 元素高亮调试
elements := page.MustElements(".item")
for _, el := range elements {
    el.MustEval(`this.style.border = "2px solid red"`)
}
```

## 故障排除

### 常见问题

#### 1. 元素找不到
```go
// 问题：Element not found
// 解决方案：
// 1. 检查选择器是否正确
// 2. 确认页面已完全加载
// 3. 添加适当的等待时间

page.MustWaitLoad()
page.MustElement("selector").MustWaitVisible()
```

#### 2. 点击无效
```go
// 问题：Click intercepted
// 解决方案：
// 1. 确保元素可见且可点击
// 2. 滚动到元素位置
// 3. 使用JavaScript强制点击

el.MustScrollIntoView()
el.MustWaitEnabled()
el.MustClick()

// 或使用JavaScript
el.MustEval(`this.click()`)
```

#### 3. 页面加载超时
```go
// 问题：Navigation timeout
// 解决方案：
// 1. 增加超时时间
// 2. 检查网络连接
// 3. 确认目标URL可达

page.Timeout(60 * time.Second).MustNavigate(url)
```

#### 4. 内存泄漏
```go
// 问题：内存持续增长
// 解决方案：
// 1. 及时关闭页面和浏览器
// 2. 使用实例池管理
// 3. 定期重启浏览器进程

// 使用完成后立即释放
defer page.MustClose()
defer browser.MustClose()
```

### 调试工具

#### 1. Chrome DevTools
```bash
# 启用调试端口
chrome --remote-debugging-port=9222 --user-data-dir=/tmp/chrome-debug

# 连接到调试端口
browser := rod.New().ControlURL("ws://localhost:9222").MustConnect()
```

#### 2. 日志诊断
```go
// 启用详细日志
rod.TryTrace(true)
rod.TryDevtools(true)

// 自定义日志输出
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
```

## 相关文档

- [平台适配器文档](../adapters/)
- [任务管理文档](../task/)
- [Cookie管理文档](../cookies/)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0