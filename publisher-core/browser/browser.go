// Package browser 提供浏览器自动化管理功能
package browser

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// Browser 浏览器实例管理
type Browser struct {
	instance *rod.Browser
	headless bool
	once     sync.Once
	mu       sync.Mutex
}

var (
	defaultBrowser *Browser
	browserOnce    sync.Once
)

// Config 浏览器配置
type Config struct {
	Headless  bool
	Proxy     string
	UserAgent string
	Debug     bool
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Headless:  true,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// NewBrowser 创建浏览器实例
func NewBrowser(cfg *Config) *Browser {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	b := &Browser{
		headless: cfg.Headless,
	}

	// 初始化浏览器
	b.init(cfg)

	return b
}

// DefaultBrowser 获取默认浏览器实例（单例）
func DefaultBrowser() *Browser {
	browserOnce.Do(func() {
		defaultBrowser = NewBrowser(DefaultConfig())
	})
	return defaultBrowser
}

func (b *Browser) init(cfg *Config) {
	b.once.Do(func() {
		// 使用 rod 直接连接浏览器
		browser := rod.New()
		if !cfg.Headless {
			// 非无头模式
			browser = browser.NoDefaultDevice()
		}
		b.instance = browser.MustConnect()
		logrus.Info("浏览器实例已创建")
	})
}

// MustPage 创建新页面
func (b *Browser) MustPage() *rod.Page {
	return b.instance.MustPage()
}

// NewPage 创建新页面（带上下文）
func (b *Browser) NewPage(ctx context.Context) (*rod.Page, error) {
	page := b.instance.MustPage()
	return page.Context(ctx), nil
}

// Close 关闭浏览器
func (b *Browser) Close() error {
	if b.instance != nil {
		return b.instance.Close()
	}
	return nil
}

// PageHelper 页面操作辅助方法
type PageHelper struct {
	page *rod.Page
}

// NewPageHelper 创建页面辅助器
func NewPageHelper(page *rod.Page) *PageHelper {
	return &PageHelper{page: page}
}

// Navigate 导航到URL并等待加载
func (h *PageHelper) Navigate(url string) error {
	logrus.Debugf("导航到 %s", url)
	if err := h.page.Navigate(url); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}
	if err := h.page.WaitLoad(); err != nil {
		logrus.Warnf("等待页面加载警告: %v", err)
	}
	h.RandomDelay(1, 2)
	return nil
}

// WaitElement 等待元素出现
func (h *PageHelper) WaitElement(selector string, timeout time.Duration) (*rod.Element, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	page := h.page.Context(ctx)
	has, elem, err := page.Has(selector)
	if err != nil {
		return nil, fmt.Errorf("查找元素失败: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("元素未找到: %s", selector)
	}
	return elem, nil
}

// ClickElement 点击元素
func (h *PageHelper) ClickElement(selector string) error {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return err
	}

	h.RandomDelay(0.3, 1)
	return elem.Click(proto.InputMouseButtonLeft, 1)
}

// InputText 输入文本
func (h *PageHelper) InputText(selector, text string) error {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return err
	}

	// 模拟人工输入
	h.RandomDelay(0.2, 0.5)
	return elem.Input(text)
}

// UploadFile 上传文件
func (h *PageHelper) UploadFile(selector, filePath string) error {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return err
	}

	return elem.SetFiles([]string{filePath})
}

// UploadFiles 上传多个文件
func (h *PageHelper) UploadFiles(selector string, filePaths []string) error {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return err
	}

	return elem.SetFiles(filePaths)
}

// HasElement 检查元素是否存在
func (h *PageHelper) HasElement(selector string) (bool, error) {
	has, _, err := h.page.Has(selector)
	return has, err
}

// GetAttribute 获取元素属性
func (h *PageHelper) GetAttribute(selector, attr string) (string, error) {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return "", err
	}

	val, err := elem.Attribute(attr)
	if err != nil {
		return "", err
	}
	if val == nil {
		return "", nil
	}
	return *val, nil
}

// GetText 获取元素文本
func (h *PageHelper) GetText(selector string) (string, error) {
	elem, err := h.WaitElement(selector, 10*time.Second)
	if err != nil {
		return "", err
	}

	return elem.Text()
}

// WaitDOMStable 等待DOM稳定
func (h *PageHelper) WaitDOMStable(d time.Duration) error {
	return h.page.WaitDOMStable(d, 0.1)
}

// RandomDelay 随机延迟（模拟人工操作）
func (h *PageHelper) RandomDelay(min, max float64) {
	delay := time.Duration((min + rand.Float64()*(max-min)) * float64(time.Second))
	time.Sleep(delay)
}

// GetCookies 获取所有Cookie
func (h *PageHelper) GetCookies() ([]*proto.NetworkCookie, error) {
	return h.page.Cookies([]string{})
}

// SetCookies 设置Cookie
func (h *PageHelper) SetCookies(cookies []*proto.NetworkCookieParam) error {
	return h.page.SetCookies(cookies)
}

// Screenshot 截图
func (h *PageHelper) Screenshot() ([]byte, error) {
	return h.page.Screenshot(false, nil)
}

// Page 返回原始页面
func (h *PageHelper) Page() *rod.Page {
	return h.page
}

// AntiCrawlerStrategies 反爬虫策略
type AntiCrawlerStrategies struct{}

// RandomDelay 全局随机延迟
func RandomDelay(min, max float64) {
	delay := time.Duration((min + rand.Float64()*(max-min)) * float64(time.Second))
	time.Sleep(delay)
}

// SimulateHumanInput 模拟人工输入
func SimulateHumanInput(page *rod.Page, selector, text string) error {
	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	// 清空现有内容
	elem.SelectAllText()
	elem.Input("")

	// 逐字输入，模拟打字
	for i, char := range text {
		if i > 0 && i%5 == 0 {
			RandomDelay(0.05, 0.15)
		}
		elem.Input(string(char))
	}

	return nil
}

// WaitForNavigation 等待页面导航完成
func WaitForNavigation(page *rod.Page, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_ = page.Context(ctx).WaitLoad()
	return nil
}

// ScrollPage 滚动页面
func ScrollPage(page *rod.Page, distance int) error {
	_, err := page.Eval(fmt.Sprintf("window.scrollBy(0, %d)", distance))
	return err
}

// ScrollToBottom 滚动到页面底部
func ScrollToBottom(page *rod.Page) error {
	_, err := page.Eval("window.scrollTo(0, document.body.scrollHeight)")
	return err
}
