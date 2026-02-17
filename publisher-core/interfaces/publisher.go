// Package publisher 定义了多平台内容发布器的统一接口
// 采用适配器模式，支持不同平台的统一调用
package publisher

import (
	"context"
	"time"
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeImages ContentType = "images"
	ContentTypeVideo  ContentType = "video"
)

// PublishStatus 发布状�?
type PublishStatus string

const (
	StatusPending    PublishStatus = "pending"    // 等待�?
	StatusProcessing PublishStatus = "processing" // 处理�?
	StatusSuccess    PublishStatus = "success"    // 成功
	StatusFailed     PublishStatus = "failed"     // 失败
)

// Content 发布内容
type Content struct {
	Type        ContentType   // 内容类型: images/video
	Title       string        // 标题
	Body        string        // 正文
	ImagePaths  []string      // 图片路径(图文类型)
	VideoPath   string        // 视频路径(视频类型)
	Tags        []string      // 话题标签
	ScheduleAt  *time.Time    // 定时发布时间
}

// PublishResult 发布结果
type PublishResult struct {
	TaskID     string        // 任务ID
	Status     PublishStatus // 发布状�?
	Platform   string        // 平台名称
	PostID     string        // 发布后的帖子ID
	PostURL    string        // 发布后的帖子链接
	Error      string        // 错误信息
	CreatedAt  time.Time     // 创建时间
	FinishedAt *time.Time    // 完成时间
}

// LoginResult 登录结果
type LoginResult struct {
	Success    bool          // 是否成功
	QrcodeURL  string        // 二维码链�?需要扫码时)
	Error      string        // 错误信息
	ExpiresAt  *time.Time    // Cookie过期时间
}

// Publisher 发布器接�?- 所有平台必须实现此接口
type Publisher interface {
	// Platform 返回平台名称
	Platform() string

	// Login 执行登录操作
	// 返回二维码链接时需要调�?WaitForLogin 等待扫码
	Login(ctx context.Context) (*LoginResult, error)

	// WaitForLogin 等待扫码登录完成
	WaitForLogin(ctx context.Context) error

	// CheckLoginStatus 检查登录状�?
	CheckLoginStatus(ctx context.Context) (bool, error)

	// Publish 发布内容
	// 返回任务ID，可通过 QueryStatus 查询状�?
	Publish(ctx context.Context, content *Content) (*PublishResult, error)

	// PublishAsync 异步发布内容
	// 立即返回任务ID，后台执行发�?
	PublishAsync(ctx context.Context, content *Content) (string, error)

	// QueryStatus 查询发布任务状�?
	QueryStatus(ctx context.Context, taskID string) (*PublishResult, error)

	// Cancel 取消发布任务
	Cancel(ctx context.Context, taskID string) error


	// Logout 登出平台，清除登录状�?
	Logout(ctx context.Context) error

	// Close 关闭发布器，释放资源
	Close() error

}

// PublisherFactory 发布器工厂接�?
type PublisherFactory interface {
	// Create 创建发布器实�?
	Create(platform string, opts ...Option) (Publisher, error)

	// SupportedPlatforms 返回支持的平台列�?
	SupportedPlatforms() []string
}

// Option 发布器配置选项
type Option func(*Options)

// Options 发布器配�?
type Options struct {
	Headless     bool          // 无头模式
	Timeout      time.Duration // 超时时间
	CookieDir    string        // Cookie存储目录
	ProxyURL     string        // 代理地址
	UserAgent    string        // User-Agent
	DebugMode    bool          // 调试模式
}

// DefaultOptions 默认配置
func DefaultOptions() *Options {
	return &Options{
		Headless:  true,
		Timeout:   10 * time.Minute,
		CookieDir: "./cookies",
	}
}

// WithHeadless 设置无头模式
func WithHeadless(headless bool) Option {
	return func(o *Options) {
		o.Headless = headless
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithCookieDir 设置Cookie存储目录
func WithCookieDir(dir string) Option {
	return func(o *Options) {
		o.CookieDir = dir
	}
}

// WithProxy 设置代理
func WithProxy(proxyURL string) Option {
	return func(o *Options) {
		o.ProxyURL = proxyURL
	}
}

// WithDebug 设置调试模式
func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.DebugMode = debug
	}
}

// ContentLimits 内容限制
type ContentLimits struct {
	TitleMaxLength   int   // 标题最大长�?
	BodyMaxLength    int   // 正文最大长�?
	MaxImages        int   // 最大图片数�?
	MaxVideoSize     int64 // 最大视频大�?字节)
	MaxTags          int   // 最大标签数�?
	AllowedVideoFormats []string // 允许的视频格�?
	AllowedImageFormats []string // 允许的图片格�?
}

// PublisherInfo 发布器信�?
type PublisherInfo struct {
	Name        string        // 平台名称
	Description string        // 平台描述
	Limits      ContentLimits // 内容限制
	Features    []string      // 支持的功�?
}
