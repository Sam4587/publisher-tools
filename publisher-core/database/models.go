package database

import (
	"time"

	"gorm.io/gorm"
)

// =====================================================
// AI 服务配置模型
// =====================================================

// AIServiceConfig AI 服务配置
type AIServiceConfig struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ServiceType   string         `gorm:"index;not null" json:"service_type"`   // text, image, video, audio
	Name          string         `gorm:"not null" json:"name"`
	Provider      string         `gorm:"index;not null" json:"provider"`       // openai, google, doubao, openrouter, groq, deepseek
	BaseURL       string         `json:"base_url"`
	APIKey        string         `json:"api_key"`
	Model         string         `json:"model"`
	Endpoint      string         `json:"endpoint"`
	QueryEndpoint string         `json:"query_endpoint"`
	Priority      int            `gorm:"default:0" json:"priority"`
	IsDefault     bool           `gorm:"default:false" json:"is_default"`
	IsActive      bool           `gorm:"default:true;index" json:"is_active"`
	Settings      string         `gorm:"type:text" json:"settings"` // JSON 格式额外配置
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AIServiceConfig) TableName() string {
	return "ai_service_configs"
}

// =====================================================
// 热点监控模型
// =====================================================

// Platform 平台
type Platform struct {
	ID        string    `gorm:"primaryKey;size:50" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Platform) TableName() string {
	return "platforms"
}

// Topic 热点话题
type Topic struct {
	ID             string    `gorm:"primaryKey;size:100" json:"id"`
	Title          string    `gorm:"index;size:500;not null" json:"title"`
	Description    string    `gorm:"type:text" json:"description"`
	Category       string    `gorm:"size:50" json:"category"`
	PlatformID     string    `gorm:"index;size:50" json:"platform_id"`
	Platform       *Platform `gorm:"foreignKey:PlatformID" json:"platform,omitempty"`
	URL            string    `gorm:"size:1000" json:"url"`
	Heat           int       `gorm:"default:0;index" json:"heat"`
	Trend          string    `gorm:"size:20;default:new" json:"trend"` // up, down, stable, new, hot
	Source         string    `gorm:"size:50;index" json:"source"`
	SourceID       string    `gorm:"size:100" json:"source_id"`
	SourceURL      string    `gorm:"size:1000" json:"source_url"`
	OriginalURL    string    `gorm:"size:1000" json:"original_url"`
	Keywords       string    `gorm:"type:text" json:"keywords"` // JSON 数组
	Suitability    int       `gorm:"default:0" json:"suitability"`
	PublishedAt    time.Time `json:"published_at"`
	FirstCrawlTime time.Time `json:"first_crawl_time"`
	LastCrawlTime  time.Time `json:"last_crawl_time"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Topic) TableName() string {
	return "topics"
}

// RankHistory 排名历史
type RankHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TopicID   string    `gorm:"index;size:100" json:"topic_id"`
	Topic     *Topic    `gorm:"foreignKey:TopicID" json:"topic,omitempty"`
	Rank      int       `json:"rank"`
	Heat      int       `json:"heat"`
	CrawlTime time.Time `gorm:"index" json:"crawl_time"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (RankHistory) TableName() string {
	return "rank_history"
}

// CrawlRecord 抓取记录
type CrawlRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CrawlTime  time.Time `gorm:"uniqueIndex" json:"crawl_time"`
	TotalItems int       `gorm:"default:0" json:"total_items"`
	Status     string    `gorm:"size:20;default:success" json:"status"` // success, failed, partial
	Error      string    `gorm:"type:text" json:"error"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (CrawlRecord) TableName() string {
	return "crawl_records"
}

// =====================================================
// 视频处理模型
// =====================================================

// Video 视频
type Video struct {
	ID           string       `gorm:"primaryKey;size:100" json:"id"`
	URL          string       `gorm:"size:1000;not null" json:"url"`
	Platform     string       `gorm:"size:50" json:"platform"`
	Title        string       `gorm:"size:500" json:"title"`
	Duration     int          `json:"duration"` // 秒
	Status       string       `gorm:"size:20;default:pending;index" json:"status"` // pending, processing, completed, failed
	Transcript   *Transcript  `gorm:"foreignKey:VideoID" json:"transcript,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// TableName 指定表名
func (Video) TableName() string {
	return "videos"
}

// Transcript 转录文本
type Transcript struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	VideoID   string    `gorm:"uniqueIndex;size:100" json:"video_id"`
	Video     *Video    `gorm:"foreignKey:VideoID" json:"video,omitempty"`
	Language  string    `gorm:"size:20" json:"language"`
	Content   string    `gorm:"type:text" json:"content"`
	Optimized string    `gorm:"type:text" json:"optimized"`
	Summary   string    `gorm:"type:text" json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (Transcript) TableName() string {
	return "transcripts"
}

// =====================================================
// 通知服务模型
// =====================================================

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:20;not null" json:"type"` // feishu, dingtalk, wecom, telegram, email
	Name      string    `gorm:"size:100;not null" json:"name"`
	Webhook   string    `gorm:"size:1000" json:"webhook"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	Config    string    `gorm:"type:text" json:"config"` // JSON 格式配置
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (NotificationChannel) TableName() string {
	return "notification_channels"
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Title     string    `gorm:"size:200" json:"title"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// =====================================================
// 任务管理模型
// =====================================================

// Task 任务
type Task struct {
	ID        string    `gorm:"primaryKey;size:100" json:"id"`
	Type      string    `gorm:"size:50;not null;index" json:"type"`
	Platform  string    `gorm:"size:50" json:"platform"`
	Status    string    `gorm:"size:20;default:pending;index" json:"status"` // pending, running, completed, failed
	Progress  int       `gorm:"default:0" json:"progress"`
	Payload   string    `gorm:"type:text" json:"payload"`  // JSON 格式
	Result    string    `gorm:"type:text" json:"result"`   // JSON 格式
	Error     string    `gorm:"type:text" json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// =====================================================
// Cookie 管理模型
// =====================================================

// Cookie Cookie 存储
type Cookie struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Platform     string    `gorm:"size:50;uniqueIndex;not null" json:"platform"`
	CookieData   string    `gorm:"type:text;not null" json:"cookie_data"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
	IsValid      bool      `gorm:"default:true" json:"is_valid"`
	RefreshToken string    `gorm:"size:500" json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Cookie) TableName() string {
	return "cookies"
}

// =====================================================
// AI 历史记录模型
// =====================================================

// AIHistory AI 调用历史
type AIHistory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Provider     string    `gorm:"size:50;index" json:"provider"`
	Model        string    `gorm:"size:100" json:"model"`
	Prompt       string    `gorm:"type:text" json:"prompt"`
	Response     string    `gorm:"type:text" json:"response"`
	TokensUsed   int       `json:"tokens_used"`
	DurationMs   int       `json:"duration_ms"`
	Success      bool      `gorm:"default:true;index" json:"success"`
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (AIHistory) TableName() string {
	return "ai_history"
}
