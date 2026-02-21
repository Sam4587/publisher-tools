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

// =====================================================
// AI 提示词模板模型
// =====================================================

// PromptTemplate 提示词模板
type PromptTemplate struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TemplateID  string         `gorm:"uniqueIndex;size:100;not null" json:"template_id"` // 唯一模板ID
	Name        string         `gorm:"size:200;not null" json:"name"`                    // 模板名称
	Type        string         `gorm:"size:50;not null;index" json:"type"`               // 模板类型：content_generation, content_rewrite, hotspot_analysis, etc.
	Category    string         `gorm:"size:50;index" json:"category"`                    // 分类标签
	Description string         `gorm:"type:text" json:"description"`                     // 模板描述
	Content     string         `gorm:"type:text;not null" json:"content"`                // 模板内容（支持变量占位符）
	Variables   string         `gorm:"type:text" json:"variables"`                       // JSON格式的变量定义
	Version     int            `gorm:"default:1" json:"version"`                         // 版本号
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`              // 是否激活
	IsDefault   bool           `gorm:"default:false" json:"is_default"`                  // 是否默认模板
	IsSystem    bool           `gorm:"default:false" json:"is_system"`                   // 是否系统模板（不可删除）
	Tags        string         `gorm:"type:text" json:"tags"`                            // JSON格式的标签数组
	Metadata    string         `gorm:"type:text" json:"metadata"`                        // JSON格式的元数据
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (PromptTemplate) TableName() string {
	return "prompt_templates"
}

// PromptTemplateVersion 提示词模板版本历史
type PromptTemplateVersion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TemplateID string    `gorm:"index;size:100;not null" json:"template_id"`
	Version    int       `gorm:"not null" json:"version"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Variables  string    `gorm:"type:text" json:"variables"`
	ChangeNote string    `gorm:"type:text" json:"change_note"` // 变更说明
	CreatedBy  string    `gorm:"size:100" json:"created_by"`   // 创建者
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (PromptTemplateVersion) TableName() string {
	return "prompt_template_versions"
}

// PromptTemplateABTest 提示词模板A/B测试
type PromptTemplateABTest struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TestName       string    `gorm:"size:200;not null" json:"test_name"`        // 测试名称
	TemplateAID    string    `gorm:"size:100;not null" json:"template_a_id"`    // 模板A的ID
	TemplateBID    string    `gorm:"size:100;not null" json:"template_b_id"`    // 模板B的ID
	Status         string    `gorm:"size:20;default:running;index" json:"status"` // running, completed, paused
	TrafficSplit   int       `gorm:"default:50" json:"traffic_split"`           // A/B流量分配（0-100，表示A的流量百分比）
	TotalCalls     int       `gorm:"default:0" json:"total_calls"`              // 总调用次数
	CallsA         int       `gorm:"default:0" json:"calls_a"`                  // 模板A调用次数
	CallsB         int       `gorm:"default:0" json:"calls_b"`                  // 模板B调用次数
	SuccessA       int       `gorm:"default:0" json:"success_a"`                // 模板A成功次数
	SuccessB       int       `gorm:"default:0" json:"success_b"`                // 模板B成功次数
	AvgDurationA   float64   `gorm:"default:0" json:"avg_duration_a"`           // 模板A平均响应时间（毫秒）
	AvgDurationB   float64   `gorm:"default:0" json:"avg_duration_b"`           // 模板B平均响应时间（毫秒）
	UserRatingA    float64   `gorm:"default:0" json:"user_rating_a"`            // 模板A用户评分（0-5）
	UserRatingB    float64   `gorm:"default:0" json:"user_rating_b"`            // 模板B用户评分（0-5）
	StartTime      time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	WinnerTemplate string    `gorm:"size:100" json:"winner_template"`            // 获胜模板ID
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PromptTemplateABTest) TableName() string {
	return "prompt_template_ab_tests"
}

// PromptTemplateUsage 提示词模板使用记录
type PromptTemplateUsage struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TemplateID   string    `gorm:"index;size:100;not null" json:"template_id"`
	UserID       string    `gorm:"size:100;index" json:"user_id"`              // 用户ID（可选）
	SessionID    string    `gorm:"size:100;index" json:"session_id"`           // 会话ID
	InputData    string    `gorm:"type:text" json:"input_data"`                // JSON格式的输入数据
	OutputData   string    `gorm:"type:text" json:"output_data"`               // JSON格式的输出数据
	Success      bool      `gorm:"default:true;index" json:"success"`
	DurationMs   int       `json:"duration_ms"`                                // 执行时长（毫秒）
	TokensUsed   int       `json:"tokens_used"`                                // 使用的token数
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	UserRating   int       `json:"user_rating"`                                // 用户评分（1-5）
	UserFeedback string    `gorm:"type:text" json:"user_feedback"`             // 用户反馈
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (PromptTemplateUsage) TableName() string {
	return "prompt_template_usage"
}

// =====================================================
// AI成本追踪模型
// =====================================================

// AICostRecord AI成本记录
type AICostRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Provider      string    `gorm:"size:50;index" json:"provider"`           // 提供商
	Model         string    `gorm:"size:100;index" json:"model"`             // 模型名称
	UserID        string    `gorm:"size:100;index" json:"user_id"`           // 用户ID
	ProjectID     string    `gorm:"size:100;index" json:"project_id"`        // 项目ID
	FunctionType  string    `gorm:"size:50;index" json:"function_type"`      // 功能类型：generation, rewrite, analysis, etc.
	InputTokens   int       `json:"input_tokens"`                            // 输入token数
	OutputTokens  int       `json:"output_tokens"`                           // 输出token数
	TotalTokens   int       `json:"total_tokens"`                            // 总token数
	CostPerToken  float64   `json:"cost_per_token"`                          // 每token成本（美元）
	TotalCost     float64   `json:"total_cost"`                              // 总成本（美元）
	DurationMs    int       `json:"duration_ms"`                             // 执行时长（毫秒）
	Success       bool      `gorm:"default:true;index" json:"success"`       // 是否成功
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`          // 错误信息
	RequestID     string    `gorm:"size:100;index" json:"request_id"`        // 请求ID
	PromptHash    string    `gorm:"size:64;index" json:"prompt_hash"`        // 提示词哈希（用于缓存分析）
	Cached        bool      `gorm:"default:false;index" json:"cached"`       // 是否来自缓存
	Metadata      string    `gorm:"type:text" json:"metadata"`               // JSON格式的元数据
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (AICostRecord) TableName() string {
	return "ai_cost_records"
}

// AIBudget AI预算配置
type AIBudget struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       string    `gorm:"size:100;index" json:"user_id"`           // 用户ID
	ProjectID    string    `gorm:"size:100;index" json:"project_id"`        // 项目ID
	BudgetType   string    `gorm:"size:20;not null" json:"budget_type"`     // 预算类型：daily, weekly, monthly
	BudgetAmount float64   `json:"budget_amount"`                           // 预算金额（美元）
	UsedAmount   float64   `json:"used_amount"`                             // 已使用金额（美元）
	AlertThreshold float64 `json:"alert_threshold"`                         // 预警阈值（百分比，如80表示80%）
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`     // 是否激活
	StartDate    time.Time `json:"start_date"`                              // 开始日期
	EndDate      time.Time `json:"end_date"`                                // 结束日期
	LastResetAt  time.Time `json:"last_reset_at"`                           // 上次重置时间
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIBudget) TableName() string {
	return "ai_budgets"
}

// AICostAlert AI成本预警记录
type AICostAlert struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	BudgetID    uint      `gorm:"index" json:"budget_id"`                  // 预算ID
	UserID      string    `gorm:"size:100;index" json:"user_id"`           // 用户ID
	ProjectID   string    `gorm:"size:100;index" json:"project_id"`        // 项目ID
	AlertType   string    `gorm:"size:20;not null" json:"alert_type"`      // 预警类型：warning, critical, exceeded
	AlertLevel  string    `gorm:"size:20;not null" json:"alert_level"`     // 预警级别：info, warning, error
	Message     string    `gorm:"type:text" json:"message"`                // 预警消息
	UsagePercent float64  `json:"usage_percent"`                           // 使用百分比
	BudgetAmount float64  `json:"budget_amount"`                           // 预算金额
	UsedAmount  float64   `json:"used_amount"`                             // 已使用金额
	IsRead      bool      `gorm:"default:false;index" json:"is_read"`      // 是否已读
	IsResolved  bool      `gorm:"default:false;index" json:"is_resolved"`  // 是否已解决
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// TableName 指定表名
func (AICostAlert) TableName() string {
	return "ai_cost_alerts"
}

// AIModelPricing AI模型定价配置
type AIModelPricing struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Provider        string    `gorm:"size:50;uniqueIndex:provider_model;not null" json:"provider"`  // 提供商
	Model           string    `gorm:"size:100;uniqueIndex:provider_model;not null" json:"model"`    // 模型名称
	InputPrice      float64   `json:"input_price"`                                                   // 输入价格（美元/1K tokens）
	OutputPrice     float64   `json:"output_price"`                                                  // 输出价格（美元/1K tokens）
	Currency        string    `gorm:"size:10;default:USD" json:"currency"`                           // 货币单位
	EffectiveDate   time.Time `json:"effective_date"`                                                // 生效日期
	ExpiryDate      *time.Time `json:"expiry_date"`                                                  // 过期日期
	IsActive        bool      `gorm:"default:true;index" json:"is_active"`                           // 是否激活
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIModelPricing) TableName() string {
	return "ai_model_pricing"
}

// =====================================================
// 异步任务系统模型
// =====================================================

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusRunning   TaskStatus = "running"   // 运行中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
	TaskStatusCancelled TaskStatus = "cancelled" // 已取消
	TaskStatusRetrying  TaskStatus = "retrying"  // 重试中
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityNormal TaskPriority = 5
	PriorityHigh   TaskPriority = 10
	PriorityUrgent TaskPriority = 20
)

// AsyncTask 异步任务
type AsyncTask struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	TaskID       string       `gorm:"uniqueIndex;size:100;not null" json:"task_id"` // 唯一任务ID
	TaskType     string       `gorm:"size:50;not null;index" json:"task_type"`      // 任务类型
	QueueName    string       `gorm:"size:50;not null;index" json:"queue_name"`     // 队列名称
	Status       TaskStatus   `gorm:"size:20;not null;index" json:"status"`         // 任务状态
	Priority     TaskPriority `json:"priority"`                                     // 优先级
	Payload      string       `gorm:"type:text" json:"payload"`                     // JSON格式的任务数据
	Result       string       `gorm:"type:text" json:"result"`                      // JSON格式的执行结果
	Error        string       `gorm:"type:text" json:"error"`                       // 错误信息
	Progress     int          `gorm:"default:0" json:"progress"`                    // 进度百分比 (0-100)
	ProgressText string       `gorm:"size:500" json:"progress_text"`                // 进度描述
	RetryCount   int          `gorm:"default:0" json:"retry_count"`                 // 重试次数
	MaxRetries   int          `gorm:"default:3" json:"max_retries"`                 // 最大重试次数
	Timeout      int          `json:"timeout"`                                      // 超时时间（秒）
	UserID       string       `gorm:"size:100;index" json:"user_id"`                // 用户ID
	ProjectID    string       `gorm:"size:100;index" json:"project_id"`             // 项目ID
	ParentTaskID string       `gorm:"size:100;index" json:"parent_task_id"`         // 父任务ID
	ScheduledAt  *time.Time   `json:"scheduled_at"`                                 // 计划执行时间
	StartedAt    *time.Time   `json:"started_at"`                                   // 开始时间
	CompletedAt  *time.Time   `json:"completed_at"`                                 // 完成时间
	ExpiredAt    *time.Time   `json:"expired_at"`                                   // 过期时间
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// TableName 指定表名
func (AsyncTask) TableName() string {
	return "async_tasks"
}

// TaskQueue 任务队列配置
type TaskQueue struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:50;not null" json:"name"` // 队列名称
	Description  string    `gorm:"type:text" json:"description"`              // 队列描述
	Concurrency  int       `gorm:"default:5" json:"concurrency"`              // 并发数
	MaxSize      int       `gorm:"default:1000" json:"max_size"`              // 最大队列大小
	Priority     int       `gorm:"default:5" json:"priority"`                 // 队列优先级
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`       // 是否激活
	RetryPolicy  string    `gorm:"type:text" json:"retry_policy"`             // JSON格式的重试策略
	Timeout      int       `json:"timeout"`                                   // 默认超时时间（秒）
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (TaskQueue) TableName() string {
	return "task_queues"
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      string    `gorm:"index;size:100;not null" json:"task_id"` // 任务ID
	WorkerID    string    `gorm:"size:100" json:"worker_id"`              // 执行者ID
	Status      string    `gorm:"size:20;not null;index" json:"status"`   // 执行状态
	Result      string    `gorm:"type:text" json:"result"`                // 执行结果
	Error       string    `gorm:"type:text" json:"error"`                 // 错误信息
	DurationMs  int       `json:"duration_ms"`                            // 执行时长（毫秒）
	MemoryMB    int       `json:"memory_mb"`                              // 内存使用（MB）
	CPUUsage    float64   `json:"cpu_usage"`                              // CPU使用率
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// TableName 指定表名
func (TaskExecution) TableName() string {
	return "task_executions"
}

// ScheduledTask 定时任务
type ScheduledTask struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:100;not null" json:"name"` // 任务名称
	TaskType     string    `gorm:"size:50;not null" json:"task_type"`         // 任务类型
	CronExpr     string    `gorm:"size:100;not null" json:"cron_expr"`        // Cron表达式
	Payload      string    `gorm:"type:text" json:"payload"`                  // JSON格式的任务数据
	QueueName    string    `gorm:"size:50;not null" json:"queue_name"`        // 目标队列
	IsActive     bool      `gorm:"default:true;index" json:"is_active"`       // 是否激活
	LastRunAt    *time.Time `json:"last_run_at"`                             // 上次执行时间
	NextRunAt    *time.Time `json:"next_run_at"`                             // 下次执行时间
	RunCount     int       `json:"run_count"`                                 // 执行次数
	LastError    string    `gorm:"type:text" json:"last_error"`               // 最后错误
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

// =====================================================
// 进度追踪模型
// =====================================================

// ProgressHistory 进度历史记录
type ProgressHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TaskID         string    `gorm:"index;size:100;not null" json:"task_id"`       // 任务ID
	Progress       int       `json:"progress"`                                      // 进度百分比 (0-100)
	CurrentStep    string    `gorm:"size:200" json:"current_step"`                  // 当前步骤
	TotalSteps     int       `json:"total_steps"`                                   // 总步骤数
	CompletedSteps int       `json:"completed_steps"`                               // 已完成步骤数
	Message        string    `gorm:"size:500" json:"message"`                       // 进度消息
	Status         string    `gorm:"size:20;index" json:"status"`                   // 状态：pending, running, completed, failed
	Metadata       string    `gorm:"type:text" json:"metadata"`                     // JSON格式的元数据
	CreatedAt      time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (ProgressHistory) TableName() string {
	return "progress_history"
}

// WebSocketSession WebSocket会话
type WebSocketSession struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ClientID     string    `gorm:"uniqueIndex;size:100;not null" json:"client_id"` // 客户端ID
	UserID       string    `gorm:"size:100;index" json:"user_id"`                   // 用户ID
	ProjectID    string    `gorm:"size:100;index" json:"project_id"`                // 项目ID
	ConnectedAt  time.Time `json:"connected_at"`                                    // 连接时间
	DisconnectedAt *time.Time `json:"disconnected_at"`                              // 断开时间
	LastActiveAt time.Time `json:"last_active_at"`                                  // 最后活跃时间
	MessageCount int       `json:"message_count"`                                   // 消息计数
	IPAddress    string    `gorm:"size:50" json:"ip_address"`                       // IP地址
	UserAgent    string    `gorm:"size:500" json:"user_agent"`                      // User Agent
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName 指定表名
func (WebSocketSession) TableName() string {
	return "websocket_sessions"
}

// =====================================================
// 流水线系统模型
// =====================================================

// PipelineDefinition 流水线定义
type PipelineDefinition struct {
	ID          string    `gorm:"primaryKey;size:100" json:"id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Steps       string    `gorm:"type:text;not null" json:"steps"`       // JSON格式的步骤定义
	Config      string    `gorm:"type:text" json:"config"`               // JSON格式的配置
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`        // 是否系统模板
	Version     int       `gorm:"default:1" json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (PipelineDefinition) TableName() string {
	return "pipeline_definitions"
}

// PipelineExecutionRecord 流水线执行记录
type PipelineExecutionRecord struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ExecutionID  string     `gorm:"uniqueIndex;size:100;not null" json:"execution_id"`
	PipelineID   string     `gorm:"index;size:100;not null" json:"pipeline_id"`
	Status       string     `gorm:"size:20;not null;index" json:"status"` // pending, running, completed, failed, paused, cancelled
	Input        string     `gorm:"type:text" json:"input"`               // JSON格式的输入
	Output       string     `gorm:"type:text" json:"output"`              // JSON格式的输出
	Steps        string     `gorm:"type:text" json:"steps"`               // JSON格式的步骤执行状态
	Error        string     `gorm:"type:text" json:"error"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	DurationMs   int        `json:"duration_ms"`                          // 执行时长（毫秒）
	UserID       string     `gorm:"size:100;index" json:"user_id"`
	ProjectID    string     `gorm:"size:100;index" json:"project_id"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TableName 指定表名
func (PipelineExecutionRecord) TableName() string {
	return "pipeline_execution_records"
}

// PipelineStepExecution 流水线步骤执行记录
type PipelineStepExecution struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ExecutionID  string     `gorm:"index;size:100;not null" json:"execution_id"`
	StepID       string     `gorm:"size:100;not null" json:"step_id"`
	StepName     string     `gorm:"size:200" json:"step_name"`
	Handler      string     `gorm:"size:100" json:"handler"`
	Status       string     `gorm:"size:20;not null;index" json:"status"` // pending, running, completed, failed, skipped
	Input        string     `gorm:"type:text" json:"input"`
	Output       string     `gorm:"type:text" json:"output"`
	Error        string     `gorm:"type:text" json:"error"`
	Progress     int        `gorm:"default:0" json:"progress"`            // 进度百分比
	RetryCount   int        `gorm:"default:0" json:"retry_count"`
	DurationMs   int        `json:"duration_ms"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TableName 指定表名
func (PipelineStepExecution) TableName() string {
	return "pipeline_step_executions"
}

// =====================================================
// 账号管理系统模型
// =====================================================

// AccountStatus 账号状态
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"   // 活跃
	AccountStatusInactive AccountStatus = "inactive" // 不活跃
	AccountStatusExpired  AccountStatus = "expired"  // 已过期
	AccountStatusBanned   AccountStatus = "banned"   // 已封禁
	AccountStatusPending  AccountStatus = "pending"  // 待验证
)

// PlatformAccount 平台账号
type PlatformAccount struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	AccountID    string        `gorm:"uniqueIndex;size:100;not null" json:"account_id"` // 唯一账号ID
	Platform     string        `gorm:"index;size:50;not null" json:"platform"`          // 平台：douyin, toutiao, xiaohongshu, bilibili
	AccountName  string        `gorm:"size:200" json:"account_name"`                    // 账号名称
	AccountType  string        `gorm:"size:20;default:personal" json:"account_type"`    // 账号类型：personal, business
	CookieData   string        `gorm:"type:text" json:"-"`                              // 加密的Cookie数据
	CookieHash   string        `gorm:"size:64" json:"cookie_hash"`                      // Cookie哈希（用于检测变化）
	Status       AccountStatus `gorm:"size:20;default:pending;index" json:"status"`
	Priority     int           `gorm:"default:5" json:"priority"`                       // 优先级（1-10，数字越大优先级越高）
	LastUsedAt   *time.Time    `json:"last_used_at"`
	LastCheckAt  *time.Time    `json:"last_check_at"`
	ExpiresAt    *time.Time    `json:"expires_at"`
	UseCount     int           `gorm:"default:0" json:"use_count"`                      // 使用次数
	SuccessCount int           `gorm:"default:0" json:"success_count"`                  // 成功次数
	FailCount    int           `gorm:"default:0" json:"fail_count"`                     // 失败次数
	LastError    string        `gorm:"type:text" json:"last_error"`
	Tags         string        `gorm:"type:text" json:"tags"`                           // JSON格式的标签
	Metadata     string        `gorm:"type:text" json:"metadata"`                       // JSON格式的元数据
	UserID       string        `gorm:"size:100;index" json:"user_id"`                   // 所属用户
	ProjectID    string        `gorm:"size:100;index" json:"project_id"`                // 所属项目
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// TableName 指定表名
func (PlatformAccount) TableName() string {
	return "platform_accounts"
}

// AccountUsageLog 账号使用日志
type AccountUsageLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AccountID   string    `gorm:"index;size:100;not null" json:"account_id"`
	Action      string    `gorm:"size:50;not null" json:"action"`        // 操作类型：login, publish, check
	Success     bool      `gorm:"default:false;index" json:"success"`
	Error       string    `gorm:"type:text" json:"error"`
	DurationMs  int       `json:"duration_ms"`                          // 执行时长（毫秒）
	IPAddress   string    `gorm:"size:50" json:"ip_address"`
	UserAgent   string    `gorm:"size:500" json:"user_agent"`
	TaskID      string    `gorm:"size:100;index" json:"task_id"`        // 关联的任务ID
	Metadata    string    `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (AccountUsageLog) TableName() string {
	return "account_usage_logs"
}

// AccountPool 账号池
type AccountPool struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PoolID      string    `gorm:"uniqueIndex;size:100;not null" json:"pool_id"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Platform    string    `gorm:"size:50;not null;index" json:"platform"`
	Description string    `gorm:"type:text" json:"description"`
	Strategy    string    `gorm:"size:20;default:round_robin" json:"strategy"` // 负载均衡策略：round_robin, random, priority, least_used
	MaxSize     int       `gorm:"default:10" json:"max_size"`
	IsActive    bool      `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AccountPool) TableName() string {
	return "account_pools"
}

// AccountPoolMember 账号池成员
type AccountPoolMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PoolID    string    `gorm:"index;size:100;not null" json:"pool_id"`
	AccountID string    `gorm:"index;size:100;not null" json:"account_id"`
	Priority  int       `gorm:"default:5" json:"priority"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (AccountPoolMember) TableName() string {
	return "account_pool_members"
}

