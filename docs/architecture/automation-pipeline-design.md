# 自动化流水线架构设计文档

## 📋 文档信息

- **文档版本**: v1.0.0
- **创建日期**: 2026-02-20
- **最后更新**: 2026-02-20
- **作者**: AI开发团队
- **状态**: 设计阶段

---

## 🎯 设计目标

基于 autoclip 项目分析，设计一个兼顾**后端可扩展性**和**前端用户体验**的自动化流水线系统，实现：

1. **内容生成 → 发布 → 监控** 的完整业务链
2. 简单方便的操作界面
3. 高度可扩展的后端架构
4. 实时进度追踪和状态监控
5. 支持多平台、多任务并发处理

---

## 🏗️ 架构概览

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端层 (React)                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ 流水线管理   │  │ 实时监控面板 │  │ 任务历史     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            │ WebSocket                          │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                      API 网关层 (Go)                            │
├────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ REST API     │  │ WebSocket    │  │ 任务管理API  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                    业务逻辑层 (Go)                              │
├────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              流水线编排器 (Pipeline Orchestrator)         │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │  │
│  │  │ 步骤1    │→ │ 步骤2    │→ │ 步骤3    │→ │ 步骤N    │ │  │
│  │  │ 内容生成 │  │ 内容优化 │  │ 质量评分 │  │ 发布执行 │ │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │  │
│  └──────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ AI服务集成   │  │ 内容处理     │  │ 平台发布     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                   任务队列层 (增强版)                           │
├────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ 任务管理器   │  │ 进度追踪器   │  │ 通知服务     │          │
│  │ (TaskManager)│  │ (Progress)   │  │ (Notify)     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            │ Redis                              │
└────────────────────────────┼────────────────────────────────────┘
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                    数据存储层                                   │
├────────────────────────────┼────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ 任务存储     │  │ 内容存储     │  │ 监控数据     │          │
│  │ (JSON/DB)    │  │ (File/DB)    │  │ (InfluxDB)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 核心组件设计

### 1. 流水线编排器 (Pipeline Orchestrator)

#### 数据结构

```go
// Pipeline 流水线定义
type Pipeline struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Steps       []PipelineStep         `json:"steps"`
    Config      PipelineConfig         `json:"config"`
    Status      PipelineStatus         `json:"status"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        StepType               `json:"type"`
    Handler     string                 `json:"handler"`
    Config      map[string]interface{} `json:"config"`
    DependsOn   []string               `json:"depends_on"`   // 依赖的步骤ID
    RetryCount  int                    `json:"retry_count"`  // 重试次数
    Timeout     time.Duration          `json:"timeout"`      // 超时时间
}

// StepType 步骤类型
type StepType string

const (
    StepTypeContentGeneration  StepType = "content_generation"  // 内容生成
    StepTypeContentOptimization StepType = "content_optimization" // 内容优化
    StepTypeQualityScoring    StepType = "quality_scoring"     // 质量评分
    StepTypePublishExecution  StepType = "publish_execution"   // 发布执行
    StepTypeDataCollection    StepType = "data_collection"     // 数据采集
    StepTypeAnalytics        StepType = "analytics"           // 数据分析
)

// PipelineConfig 流水线配置
type PipelineConfig struct {
    ParallelMode    bool                   `json:"parallel_mode"`    // 并行模式
    MaxParallel     int                    `json:"max_parallel"`     // 最大并行数
    FailFast        bool                   `json:"fail_fast"`        // 快速失败
    RetryStrategy   RetryStrategy          `json:"retry_strategy"`   // 重试策略
    Notification    NotificationConfig     `json:"notification"`     // 通知配置
}

// PipelineStatus 流水线状态
type PipelineStatus string

const (
    PipelineStatusDraft      PipelineStatus = "draft"
    PipelineStatusActive     PipelineStatus = "active"
    PipelineStatusRunning    PipelineStatus = "running"
    PipelineStatusCompleted  PipelineStatus = "completed"
    PipelineStatusFailed     PipelineStatus = "failed"
    PipelineStatusPaused     PipelineStatus = "paused"
)
```

#### 编排器接口

```go
// Orchestrator 编排器接口
type Orchestrator interface {
    // 创建流水线
    CreatePipeline(pipeline *Pipeline) error

    // 执行流水线
    ExecutePipeline(ctx context.Context, pipelineID string, input map[string]interface{}) (*PipelineExecution, error)

    // 暂停流水线
    PausePipeline(pipelineID string) error

    // 恢复流水线
    ResumePipeline(pipelineID string) error

    // 取消流水线
    CancelPipeline(pipelineID string) error

    // 获取执行状态
    GetExecutionStatus(executionID string) (*PipelineExecution, error)

    // 获取执行日志
    GetExecutionLogs(executionID string) ([]ExecutionLog, error)
}

// PipelineExecution 流水线执行实例
type PipelineExecution struct {
    ID          string                 `json:"id"`
    PipelineID  string                 `json:"pipeline_id"`
    Status      ExecutionStatus        `json:"status"`
    Input       map[string]interface{} `json:"input"`
    Output      map[string]interface{} `json:"output"`
    Steps       []StepExecution        `json:"steps"`
    StartedAt   time.Time              `json:"started_at"`
    FinishedAt  *time.Time             `json:"finished_at,omitempty"`
    Error       string                 `json:"error,omitempty"`
}

// StepExecution 步骤执行实例
type StepExecution struct {
    StepID      string                 `json:"step_id"`
    Status      StepStatus             `json:"status"`
    Input       map[string]interface{} `json:"input"`
    Output      map[string]interface{} `json:"output"`
    Progress    int                    `json:"progress"`
    StartedAt   time.Time              `json:"started_at"`
    FinishedAt  *time.Time             `json:"finished_at,omitempty"`
    Error       string                 `json:"error,omitempty"`
    Logs        []string               `json:"logs"`
}
```

### 2. 增强版任务管理器

基于现有的 `TaskManager`，增加以下功能：

```go
// EnhancedTaskManager 增强版任务管理器
type EnhancedTaskManager struct {
    *TaskManager
    progressTracker *ProgressTracker
    notificationService *NotificationService
    retryManager *RetryManager
}

// ProgressTracker 进度追踪器
type ProgressTracker struct {
    mu         sync.RWMutex
    progress   map[string]*ProgressDetail
    subscribers map[string][]chan *ProgressDetail
}

// ProgressDetail 进度详情
type ProgressDetail struct {
    TaskID      string                 `json:"task_id"`
    Progress    int                    `json:"progress"`       // 0-100
    CurrentStep string                 `json:"current_step"`
    TotalSteps  int                    `json:"total_steps"`
    Message     string                 `json:"message"`
    Data        map[string]interface{} `json:"data,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
}

// NotificationService 通知服务
type NotificationService struct {
    wsHub *WebSocketHub
    emailService *EmailService
    webhookService *WebhookService
}

// WebSocketHub WebSocket连接管理
type WebSocketHub struct {
    clients    map[string]*Client
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

// Client WebSocket客户端
type Client struct {
    ID     string
    Conn   *websocket.Conn
    Send   chan []byte
    Topics []string
}

// RetryManager 重试管理器
type RetryManager struct {
    strategy RetryStrategy
    maxRetries int
}

// RetryStrategy 重试策略
type RetryStrategy struct {
    Type           RetryType `json:"type"`
    InitialDelay   time.Duration `json:"initial_delay"`
    MaxDelay       time.Duration `json:"max_delay"`
    BackoffFactor  float64 `json:"backoff_factor"`
}

type RetryType string

const (
    RetryTypeFixed     RetryType = "fixed"
    RetryTypeExponential RetryType = "exponential"
    RetryTypeLinear    RetryType = "linear"
)
```

### 3. 预定义流水线模板

#### 内容发布流水线

```go
// ContentPublishPipeline 内容发布流水线模板
func ContentPublishPipeline() *Pipeline {
    return &Pipeline{
        ID:          "content-publish-v1",
        Name:        "内容发布流水线",
        Description: "从内容生成到多平台发布的完整流程",
        Steps: []PipelineStep{
            {
                ID:       "step-1",
                Name:     "内容生成",
                Type:     StepTypeContentGeneration,
                Handler:  "ai_content_generator",
                Config: map[string]interface{}{
                    "model": "deepseek-chat",
                    "max_tokens": 2000,
                    "temperature": 0.7,
                },
                RetryCount: 3,
                Timeout:    5 * time.Minute,
            },
            {
                ID:       "step-2",
                Name:     "内容优化",
                Type:     StepTypeContentOptimization,
                Handler:  "content_optimizer",
                DependsOn: []string{"step-1"},
                Config: map[string]interface{}{
                    "check_spelling": true,
                    "improve_readability": true,
                },
                RetryCount: 2,
                Timeout:    2 * time.Minute,
            },
            {
                ID:       "step-3",
                Name:     "质量评分",
                Type:     StepTypeQualityScoring,
                Handler:  "quality_scorer",
                DependsOn: []string{"step-2"},
                Config: map[string]interface{}{
                    "min_score": 0.7,
                    "scoring_model": "quality-v2",
                },
                RetryCount: 1,
                Timeout:    1 * time.Minute,
            },
            {
                ID:       "step-4",
                Name:     "发布执行",
                Type:     StepTypePublishExecution,
                Handler:  "platform_publisher",
                DependsOn: []string{"step-3"},
                Config: map[string]interface{}{
                    "platforms": []string{"douyin", "toutiao", "xiaohongshu"},
                    "async_mode": true,
                },
                RetryCount: 3,
                Timeout:    10 * time.Minute,
            },
            {
                ID:       "step-5",
                Name:     "数据采集",
                Type:     StepTypeDataCollection,
                Handler:  "analytics_collector",
                DependsOn: []string{"step-4"},
                Config: map[string]interface{}{
                    "collect_immediately": true,
                    "collect_after_hours": 24,
                },
                RetryCount: 3,
                Timeout:    5 * time.Minute,
            },
        },
        Config: PipelineConfig{
            ParallelMode:  false,
            MaxParallel:   1,
            FailFast:      true,
            RetryStrategy: RetryStrategy{
                Type:          RetryTypeExponential,
                InitialDelay:  1 * time.Second,
                MaxDelay:      30 * time.Second,
                BackoffFactor: 2.0,
            },
            Notification: NotificationConfig{
                OnStart:    true,
                OnComplete: true,
                OnError:    true,
                Channels:   []string{"websocket", "email"},
            },
        },
    }
}
```

#### 视频处理流水线

```go
// VideoProcessingPipeline 视频处理流水线模板
func VideoProcessingPipeline() *Pipeline {
    return &Pipeline{
        ID:          "video-processing-v1",
        Name:        "视频处理流水线",
        Description: "视频下载、转录、切片、发布的完整流程",
        Steps: []PipelineStep{
            {
                ID:       "step-1",
                Name:     "视频下载",
                Type:     StepTypeDataCollection,
                Handler:  "video_downloader",
                Config: map[string]interface{}{
                    "max_retries": 3,
                    "timeout": "10m",
                },
            },
            {
                ID:       "step-2",
                Name:     "语音转录",
                Type:     StepTypeContentGeneration,
                Handler:  "speech_transcriber",
                DependsOn: []string{"step-1"},
                Config: map[string]interface{}{
                    "strategy": "cloud_first", // cloud_first, local_only, hybrid
                    "cloud_service": "bcut_asr",
                    "local_service": "whisper",
                },
            },
            {
                ID:       "step-3",
                Name:     "内容改写",
                Type:     StepTypeContentOptimization,
                Handler:  "content_rewriter",
                DependsOn: []string{"step-2"},
                Config: map[string]interface{}{
                    "style": "casual",
                    "max_length": 1000,
                },
            },
            {
                ID:       "step-4",
                Name:     "视频切片",
                Type:     StepTypeDataCollection,
                Handler:  "video_cutter",
                DependsOn: []string{"step-3"},
                Config: map[string]interface{}{
                    "max_duration": 60,
                    "output_format": "mp4",
                },
            },
            {
                ID:       "step-5",
                Name:     "发布执行",
                Type:     StepTypePublishExecution,
                Handler:  "platform_publisher",
                DependsOn: []string{"step-4"},
                Config: map[string]interface{}{
                    "platforms": []string{"douyin", "xiaohongshu"},
                },
            },
        },
        Config: PipelineConfig{
            ParallelMode:  false,
            MaxParallel:   1,
            FailFast:      false,
        },
    }
}
```

---

## 📡 API 设计

### REST API 端点

#### 流水线管理

```go
// 流水线管理 API
POST   /api/v1/pipelines                    // 创建流水线
GET    /api/v1/pipelines                    // 获取流水线列表
GET    /api/v1/pipelines/{id}               // 获取流水线详情
PUT    /api/v1/pipelines/{id}               // 更新流水线
DELETE /api/v1/pipelines/{id}               // 删除流水线

// 流水线执行 API
POST   /api/v1/pipelines/{id}/execute       // 执行流水线
GET    /api/v1/executions/{id}              // 获取执行状态
POST   /api/v1/executions/{id}/pause        // 暂停执行
POST   /api/v1/executions/{id}/resume       // 恢复执行
POST   /api/v1/executions/{id}/cancel       // 取消执行
GET    /api/v1/executions/{id}/logs         // 获取执行日志

// 预定义模板 API
GET    /api/v1/pipeline-templates           // 获取模板列表
GET    /api/v1/pipeline-templates/{id}      // 获取模板详情
POST   /api/v1/pipeline-templates/{id}/use  // 使用模板创建流水线
```

#### 实时进度 API

```go
// 进度追踪 API
GET    /api/v1/tasks/{id}/progress          // 获取任务进度
GET    /api/v1/executions/{id}/progress     // 获取执行进度
GET    /api/v1/pipelines/{id}/executions    // 获取流水线执行历史
```

### WebSocket API

#### 连接端点

```
ws://localhost:8080/ws/pipeline/{pipeline_id}
ws://localhost:8080/ws/execution/{execution_id}
ws://localhost:8080/ws/task/{task_id}
```

#### 消息格式

```json
// 客户端订阅
{
  "type": "subscribe",
  "topics": ["pipeline:123", "execution:456"]
}

// 服务器推送 - 进度更新
{
  "type": "progress",
  "data": {
    "execution_id": "exec-123",
    "step_id": "step-1",
    "progress": 50,
    "current_step": "内容生成中",
    "total_steps": 5,
    "message": "正在生成内容...",
    "timestamp": "2026-02-20T10:30:00Z"
  }
}

// 服务器推送 - 状态变更
{
  "type": "status_change",
  "data": {
    "execution_id": "exec-123",
    "status": "running",
    "timestamp": "2026-02-20T10:30:00Z"
  }
}

// 服务器推送 - 错误通知
{
  "type": "error",
  "data": {
    "execution_id": "exec-123",
    "step_id": "step-2",
    "error": "AI服务调用失败: 超时",
    "timestamp": "2026-02-20T10:35:00Z"
  }
}

// 服务器推送 - 完成
{
  "type": "completed",
  "data": {
    "execution_id": "exec-123",
    "status": "completed",
    "output": {
      "published_urls": ["https://douyin.com/123", "https://xiaohongshu.com/456"]
    },
    "timestamp": "2026-02-20T10:40:00Z"
  }
}
```

---

## 🎨 前端界面设计

### 1. 流水线管理页面

#### 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│  流水线管理                                [+ 创建流水线]    │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │  搜索: [输入关键词]  状态: [全部▼]  排序: [创建时间▼] │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  流水线卡片列表                                       │  │
│  │  ┌────────────────────────────────────────────────┐ │  │
│  │  │ 📝 内容发布流水线                    [编辑] [删除]│ │  │
│  │  │ 状态: ✅ 活跃  执行次数: 156  成功率: 98%        │ │  │
│  │  │ 步骤: 5个  预计耗时: 15分钟                      │ │  │
│  │  │ [立即执行] [查看历史]                            │ │  │
│  │  └────────────────────────────────────────────────┘ │  │
│  │  ┌────────────────────────────────────────────────┐ │  │
│  │  │ 🎬 视频处理流水线                    [编辑] [删除]│ │  │
│  │  │ 状态: ✅ 活跃  执行次数: 89  成功率: 95%         │ │  │
│  │  │ 步骤: 5个  预计耗时: 20分钟                      │ │  │
│  │  │ [立即执行] [查看历史]                            │ │  │
│  │  └────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

#### 组件结构

```typescript
// PipelineManagement.tsx
interface Pipeline {
  id: string;
  name: string;
  description: string;
  status: PipelineStatus;
  steps: PipelineStep[];
  stats: {
    totalExecutions: number;
    successRate: number;
    avgDuration: number;
  };
}

const PipelineManagement = () => {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [selectedPipeline, setSelectedPipeline] = useState<Pipeline | null>(null);
  const [isExecuting, setIsExecuting] = useState(false);

  return (
    <div className="pipeline-management">
      <PipelineHeader onCreate={handleCreate} />
      <PipelineFilter onFilter={handleFilter} />
      <PipelineList
        pipelines={pipelines}
        onSelect={setSelectedPipeline}
        onExecute={handleExecute}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />
      {selectedPipeline && (
        <PipelineDetail
          pipeline={selectedPipeline}
          onClose={() => setSelectedPipeline(null)}
        />
      )}
    </div>
  );
};
```

### 2. 实时监控面板

#### 页面布局

```
┌─────────────────────────────────────────────────────────────┐
│  实时监控面板                                [全屏] [刷新]   │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │ 运行中: 3   │  │ 今日: 45    │  │ 成功率: 96% │          │
│  │ 等待中: 12  │  │ 本周: 234   │  │ 失败: 2     │          │
│  └─────────────┘  └─────────────┘  └─────────────┘          │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  执行中任务                                           │  │
│  │  ┌────────────────────────────────────────────────┐ │  │
│  │  │ exec-123  内容发布流水线     [暂停] [取消]      │ │  │
│  │  │ ████████████░░░░░░░░░  60%  步骤3/5: 质量评分   │ │  │
│  │  │ 开始: 10:30  预计完成: 10:45  耗时: 8分钟       │ │  │
│  │  └────────────────────────────────────────────────┘ │  │
│  │  ┌────────────────────────────────────────────────┐ │  │
│  │  │ exec-124  视频处理流水线     [暂停] [取消]      │ │  │
│  │  │ ████████████████░░░░░░  70%  步骤3/5: 内容改写   │ │  │
│  │  │ 开始: 10:35  预计完成: 10:55  耗时: 12分钟      │ │  │
│  │  └────────────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  最近