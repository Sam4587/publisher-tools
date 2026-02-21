# AI处理流水线编排器

## 概述

AI处理流水线编排器是一个可配置的流水线系统，用于编排和执行复杂的AI处理任务。它支持步骤依赖、条件执行、并行处理、进度追踪等功能。

## 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                    PipelineOrchestrator                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Pipeline     │  │ Step         │  │ Progress     │      │
│  │ Definition   │  │ Handlers     │  │ Tracker      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                 │                 │               │
│         └─────────────────┼─────────────────┘               │
│                           ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              DBStorage (数据库持久化)                 │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 文件结构

```
publisher-core/pipeline/
├── orchestrator.go        # 流水线编排器核心
├── handlers.go            # 基础步骤处理器
├── advanced_handlers.go   # 高级步骤处理器
├── templates.go           # 预定义流水线模板
├── storage.go             # 数据库存储服务
└── README.md              # 本文档
```

## 流水线定义

### Pipeline 结构

```go
type Pipeline struct {
    ID          string         // 流水线ID
    Name        string         // 流水线名称
    Description string         // 描述
    Steps       []PipelineStep // 步骤列表
    Config      PipelineConfig // 配置
    Status      PipelineStatus // 状态
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### PipelineStep 结构

```go
type PipelineStep struct {
    ID         string                 // 步骤ID
    Name       string                 // 步骤名称
    Type       StepType               // 步骤类型
    Handler    string                 // 处理器名称
    Config     map[string]interface{} // 步骤配置
    DependsOn  []string               // 依赖的步骤ID
    RetryCount int                    // 重试次数
    Timeout    time.Duration          // 超时时间
}
```

### PipelineConfig 结构

```go
type PipelineConfig struct {
    ParallelMode  bool               // 是否并行执行
    MaxParallel   int                // 最大并行数
    FailFast      bool               // 是否快速失败
    RetryStrategy RetryStrategy      // 重试策略
    Notification  NotificationConfig // 通知配置
}
```

## 步骤类型

### 基础步骤类型

| 类型 | 说明 | 处理器 |
|------|------|--------|
| content_generation | 内容生成 | ai_content_generator |
| content_optimization | 内容优化 | content_optimizer |
| quality_scoring | 质量评分 | quality_scorer |
| publish_execution | 发布执行 | platform_publisher |
| data_collection | 数据采集 | analytics_collector |
| analytics | 数据分析 | data_analyzer |

### 高级步骤类型

| 类型 | 说明 | 处理器 |
|------|------|--------|
| outline_generation | 大纲生成 | outline_generator |
| content_clustering | 内容聚类 | content_clusterer |
| content_classification | 内容分类 | content_classifier |
| title_generation | 标题生成 | title_generator |
| description_generation | 描述生成 | description_generator |
| content_filtering | 内容筛选 | content_filter |
| conditional_execution | 条件执行 | conditional_executor |
| parallel_execution | 并行执行 | parallel_executor |

## 预定义模板

### 1. 内容发布流水线

```go
pipeline.ContentPublishPipeline()
```

**步骤：**
1. 内容生成 - AI生成内容
2. 内容优化 - 优化可读性
3. 质量评分 - 评估内容质量
4. 发布执行 - 多平台发布
5. 数据采集 - 采集发布数据

### 2. 视频处理流水线

```go
pipeline.VideoProcessingPipeline()
```

**步骤：**
1. 视频下载 - 下载视频文件
2. 语音转录 - 转录音频内容
3. 内容改写 - 改写为发布内容
4. 视频切片 - 切割视频片段
5. 发布执行 - 发布到平台

### 3. 热点分析流水线

```go
pipeline.HotspotAnalysisPipeline()
```

**步骤：**
1. 热点抓取 - 抓取热点数据
2. 趋势分析 - 分析热点趋势
3. 内容生成 - 生成相关内容
4. 发布执行 - 发布内容

### 4. 数据采集流水线

```go
pipeline.DataCollectionPipeline()
```

**步骤：**
1. 抖音数据采集
2. 今日头条数据采集
3. 小红书数据采集
4. 数据分析
5. 报告生成

## API接口

### 流水线管理

#### 创建流水线

```http
POST /api/v1/pipelines
Content-Type: application/json

{
  "name": "我的流水线",
  "description": "自定义流水线",
  "template_id": "content-publish-v1",
  "config": {
    "parallel_mode": false,
    "fail_fast": true
  }
}
```

#### 获取流水线列表

```http
GET /api/v1/pipelines
```

#### 获取流水线详情

```http
GET /api/v1/pipelines/{id}
```

#### 更新流水线

```http
PUT /api/v1/pipelines/{id}
Content-Type: application/json

{
  "name": "更新后的名称",
  "description": "更新后的描述"
}
```

#### 删除流水线

```http
DELETE /api/v1/pipelines/{id}
```

### 流水线执行

#### 执行流水线

```http
POST /api/v1/pipelines/{id}/execute
Content-Type: application/json

{
  "input": {
    "topic": "人工智能发展趋势",
    "keywords": ["AI", "机器学习", "深度学习"],
    "target_audience": "科技爱好者"
  }
}
```

**响应：**
```json
{
  "id": "exec-123456",
  "pipeline_id": "pipeline-001",
  "status": "running",
  "started_at": "2026-02-21T10:00:00Z"
}
```

#### 获取执行状态

```http
GET /api/v1/executions/{id}
```

#### 暂停执行

```http
POST /api/v1/executions/{id}/pause
```

#### 恢复执行

```http
POST /api/v1/executions/{id}/resume
```

#### 取消执行

```http
POST /api/v1/executions/{id}/cancel
```

#### 获取执行进度

```http
GET /api/v1/executions/{id}/progress
```

**响应：**
```json
{
  "execution_id": "exec-123456",
  "progress": 45.5,
  "total_steps": 5,
  "completed_steps": 2,
  "status": "running",
  "steps": [
    {
      "step_id": "step-1",
      "name": "内容生成",
      "status": "completed",
      "progress": 100
    },
    {
      "step_id": "step-2",
      "name": "内容优化",
      "status": "running",
      "progress": 50
    }
  ]
}
```

### 模板管理

#### 获取模板列表

```http
GET /api/v1/pipeline-templates
```

#### 获取模板详情

```http
GET /api/v1/pipeline-templates/{id}
```

#### 使用模板创建流水线

```http
POST /api/v1/pipeline-templates/{id}/use
Content-Type: application/json

{
  "name": "基于模板的流水线",
  "config": {
    "fail_fast": false
  }
}
```

## 自定义步骤处理器

### 实现步骤处理器接口

```go
type StepHandler interface {
    Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error)
}
```

### 示例：自定义处理器

```go
// CustomHandler 自定义处理器
type CustomHandler struct {
    // 依赖的服务
    someService *SomeService
}

func NewCustomHandler(someService *SomeService) *CustomHandler {
    return &CustomHandler{someService: someService}
}

func (h *CustomHandler) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 1. 从input获取参数
    param, ok := input["param"].(string)
    if !ok {
        return nil, fmt.Errorf("缺少 param 参数")
    }

    // 2. 从config获取配置
    option, _ := config["option"].(string)

    // 3. 执行业务逻辑
    result, err := h.someService.DoSomething(ctx, param, option)
    if err != nil {
        return nil, err
    }

    // 4. 返回结果
    return map[string]interface{}{
        "result": result,
        "processed_at": time.Now().Format(time.RFC3339),
    }, nil
}
```

### 注册处理器

```go
// 创建编排器
orchestrator := pipeline.NewPipelineOrchestrator(storage)

// 注册自定义处理器
orchestrator.RegisterHandler("custom_handler", NewCustomHandler(someService))
```

## 条件执行

### 条件语法

支持简单的条件表达式：

- `field == value` - 等于
- `field != value` - 不等于
- `field > value` - 大于
- `field >= value` - 大于等于
- `field < value` - 小于
- `field <= value` - 小于等于

### 示例

```json
{
  "id": "conditional-step",
  "name": "条件执行",
  "handler": "conditional_executor",
  "config": {
    "condition": "score >= 0.7",
    "true_step": {
      "handler": "platform_publisher",
      "config": {
        "platforms": ["douyin", "xiaohongshu"]
      }
    },
    "false_step": {
      "handler": "content_optimizer",
      "config": {
        "improve_readability": true
      }
    }
  }
}
```

## 并行执行

### 配置并行执行

```json
{
  "id": "parallel-step",
  "name": "并行执行",
  "handler": "parallel_executor",
  "config": {
    "max_parallel": 3,
    "steps": [
      {
        "handler": "douyin_collector",
        "config": {"metrics": ["views", "likes"]}
      },
      {
        "handler": "toutiao_collector",
        "config": {"metrics": ["views", "likes"]}
      },
      {
        "handler": "xiaohongshu_collector",
        "config": {"metrics": ["views", "likes"]}
      }
    ]
  }
}
```

## 重试策略

### 重试类型

| 类型 | 说明 |
|------|------|
| none | 不重试 |
| fixed | 固定间隔重试 |
| linear | 线性递增重试 |
| exponential | 指数退避重试 |

### 配置示例

```json
{
  "retry_strategy": {
    "type": "exponential",
    "initial_delay": "1s",
    "max_delay": "30s",
    "backoff_factor": 2.0,
    "max_retries": 3
  }
}
```

## 进度追踪

### WebSocket实时进度

连接到WebSocket端点接收实时进度更新：

```javascript
const ws = new WebSocket('ws://host/api/v1/ws?user_id=user-123')

// 订阅执行进度
ws.send(JSON.stringify({
  action: 'subscribe',
  task_id: 'exec-123456'
}))

// 接收进度消息
ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  if (message.type === 'progress') {
    console.log('进度:', message.payload.progress)
    console.log('当前步骤:', message.payload.current_step)
  }
}
```

### 进度消息格式

```json
{
  "type": "progress",
  "task_id": "exec-123456",
  "payload": {
    "execution_id": "exec-123456",
    "step_id": "step-2",
    "progress": 45,
    "current_step": "步骤 2/5: 内容优化",
    "total_steps": 5,
    "message": "正在执行: 内容优化",
    "timestamp": "2026-02-21T10:01:30Z"
  }
}
```

## 数据库模型

### PipelineDefinition

```go
type PipelineDefinition struct {
    ID          string    // 流水线ID
    Name        string    // 名称
    Description string    // 描述
    Steps       string    // JSON格式的步骤定义
    Config      string    // JSON格式的配置
    IsActive    bool      // 是否激活
    IsSystem    bool      // 是否系统模板
    Version     int       // 版本号
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### PipelineExecutionRecord

```go
type PipelineExecutionRecord struct {
    ID           uint       // 主键
    ExecutionID  string     // 执行ID
    PipelineID   string     // 流水线ID
    Status       string     // 状态
    Input        string     // JSON格式的输入
    Output       string     // JSON格式的输出
    Steps        string     // JSON格式的步骤状态
    Error        string     // 错误信息
    StartedAt    time.Time  // 开始时间
    FinishedAt   *time.Time // 结束时间
    DurationMs   int        // 执行时长（毫秒）
    UserID       string     // 用户ID
    ProjectID    string     // 项目ID
    CreatedAt    time.Time
}
```

## 最佳实践

### 1. 步骤设计原则

- **单一职责**：每个步骤只做一件事
- **幂等性**：步骤可以安全重试
- **可观测性**：记录足够的日志和指标
- **超时控制**：设置合理的超时时间

### 2. 错误处理

```go
// 在步骤处理器中处理错误
func (h *MyHandler) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 可重试的错误
    if isRetryableError(err) {
        return nil, fmt.Errorf("可重试错误: %w", err)
    }
    
    // 不可重试的错误
    if isFatalError(err) {
        return nil, &FatalError{Message: err.Error()}
    }
    
    // 部分成功
    return map[string]interface{}{
        "partial_result": result,
        "warnings": warnings,
    }, nil
}
```

### 3. 资源清理

```go
func (h *MyHandler) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 创建资源
    resource := createResource()
    defer func() {
        // 确保资源被清理
        resource.Close()
    }()
    
    // 执行逻辑...
}
```

### 4. 上下文传递

```go
func (h *MyHandler) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 检查上下文是否已取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // 带超时的子操作
    subCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    return h.doWork(subCtx, input)
}
```

## 故障排查

### 常见问题

1. **步骤执行超时**
   - 检查步骤的Timeout配置
   - 检查下游服务响应时间
   - 考虑增加超时时间或优化处理逻辑

2. **步骤依赖未满足**
   - 检查DependsOn配置是否正确
   - 确认依赖步骤已成功完成
   - 查看执行日志确认依赖关系

3. **处理器未找到**
   - 确认处理器已注册
   - 检查处理器名称是否正确
   - 查看启动日志确认注册成功

### 调试技巧

```go
// 启用详细日志
logrus.SetLevel(logrus.DebugLevel)

// 查看步骤执行详情
execution, _ := orchestrator.GetExecutionStatus(executionID)
for _, step := range execution.Steps {
    logrus.Debugf("步骤 %s: %s, 输出: %v", step.StepID, step.Status, step.Output)
}
```

## 相关文档

- [异步任务系统指南](./async-task-system-guide.md)
- [实时进度追踪](./realtime-progress-tracking.md)
- [API接口文档](../api/rest-api.md)
