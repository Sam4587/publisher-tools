package publisher

import (
	"context"
	"time"
)

type ContentType string

const (
	ContentTypeImages ContentType = "images"
	ContentTypeVideo  ContentType = "video"
)

type PublishStatus string

const (
	StatusPending    PublishStatus = "pending"
	StatusProcessing PublishStatus = "processing"
	StatusSuccess    PublishStatus = "success"
	StatusFailed     PublishStatus = "failed"
)

type Content struct {
	Type        ContentType
	Title       string
	Body        string
	ImagePaths  []string
	VideoPath   string
	Tags        []string
	ScheduleAt  *time.Time
}

type PublishResult struct {
	TaskID     string
	Status     PublishStatus
	Platform   string
	PostID     string
	PostURL    string
	Error      string
	CreatedAt  time.Time
	FinishedAt *time.Time
}

type LoginResult struct {
	Success    bool
	QrcodeURL  string
	Error      string
	ExpiresAt  *time.Time
}

type Publisher interface {
	Platform() string
	Login(ctx context.Context) (*LoginResult, error)
	WaitForLogin(ctx context.Context) error
	CheckLoginStatus(ctx context.Context) (bool, error)
	Publish(ctx context.Context, content *Content) (*PublishResult, error)
	PublishAsync(ctx context.Context, content *Content) (string, error)
	QueryStatus(ctx context.Context, taskID string) (*PublishResult, error)
	Cancel(ctx context.Context, taskID string) error
	Logout(ctx context.Context) error
	Close() error
}

type PublisherFactory interface {
	Create(platform string, opts ...Option) (Publisher, error)
	SupportedPlatforms() []string
}

type Option func(*Options)

type Options struct {
	Headless     bool
	Timeout      time.Duration
	CookieDir    string
	ProxyURL     string
	UserAgent    string
	DebugMode    bool
}

func DefaultOptions() *Options {
	return &Options{
		Headless:  true,
		Timeout:   10 * time.Minute,
		CookieDir: "./cookies",
	}
}

func WithHeadless(headless bool) Option {
	return func(o *Options) {
		o.Headless = headless
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

func WithCookieDir(dir string) Option {
	return func(o *Options) {
		o.CookieDir = dir
	}
}

func WithProxy(proxyURL string) Option {
	return func(o *Options) {
		o.ProxyURL = proxyURL
	}
}

func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.DebugMode = debug
	}
}

type ContentLimits struct {
	TitleMaxLength      int
	BodyMaxLength       int
	MaxImages           int
	MaxVideoSize        int64
	MaxTags             int
	AllowedVideoFormats []string
	AllowedImageFormats []string
}

type PublisherInfo struct {
	Name        string
	Description string
	Limits      ContentLimits
	Features    []string
}
