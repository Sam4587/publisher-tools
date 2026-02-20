# 自动化流水线系统 - 快速开始指南

## 📋 概述

本指南将帮助您快速上手自动化流水线系统，实现**内容生成 → 发布 → 监控**的完整业务链。

---

## 🚀 快速开始

### 1. 环境准备

确保您已安装以下依赖：

```bash
# Go 1.21+
go version

# Node.js 18+
node --version

# Redis (可选，用于缓存)
redis-server --version
```

### 2. 启动后端服务

```bash
# 进入项目目录
cd publisher-core

# 编译项目
go build -o ../bin/publisher-server ./cmd/server

# 启动服务
../bin/publisher-server -port 8080
```

### 3. 启动前端服务

```bash
# 进入前端目录
cd publisher-web

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

### 4. 访问系统

打开浏览器访问：http://localhost:5173

---

## 📝 使用预定义模板

### 方式1: 通过代码使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "your-project/publisher-core/pipeline"
)

func main() {
    // 1. 创建编排器
    orchestrator := pipeline.NewPipelineOrchestrator(nil)

    // 2. 获取预定义模板
    template, err := pipeline.GetTemplate("content-publish-v1")
    if err != nil {
        log.Fatal(err)
    }

    // 3. 创建流水线
    if err := orchestrator.CreatePipeline(template); err != nil {
        log.Fatal(err)
    }

    // 4. 准备输入数据
    input := map[string]interface{}{
        "topic": "人工智能最新进展",
        "keywords": []string{"AI", "机器学习", "深度学习"},
        "target_audience": "技术爱好者",
        "platforms": []string{"douyin", "toutiao"},
    }

    // 5. 执行流水线
    execution, err := orchestrator.ExecutePipeline(context.Background(), template.ID, input)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("✅ 流水线已启动\n")
    fmt.Printf("执行ID: %s\n", execution.ID)
    fmt.Printf("状态: %s\n", execution.Status)

    // 6. 监控执行状态
    for {
        status, _ := orchestrator.GetExecutionStatus(execution.ID)
        fmt.Printf("当前状态: %s, 进度: %d%%\n", status.Status, calculateProgress(status))

        if status.Status == pipeline.ExecutionStatusCompleted ||
           status.Status == pipeline.ExecutionStatusFailed {
            break
        }

        time.Sleep(2 * time.Second)
    }
}

func calculateProgress(execution *pipeline.PipelineExecution) int {
    completed := 0
    for _, step := range execution.Steps {
        if step.Status == pipeline.StepStatusCompleted {
            completed++
        }
    }
    return int(float64(completed) / float64(len(execution.Steps)) * 100)
}
```

### 方式2: 通过 API 使用

```bash
# 1. 使用模板创建流水线
curl -X POST http://localhost:8080/api/v1/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "content-publish-v1",
    "name": "我的内容发布流水线",
    "config": {
      "platforms": ["douyin", "toutiao"]
    }
  }'

# 2. 执行流水线
curl -X POST http://localhost:8080/api/v1/pipelines/{pipeline_id}/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "topic": "人工智能最新进展",
      "keywords": ["AI", "机器学习"]
    }
  }'

# 3. 查询执行状态
curl http://localhost:8080/api/v1/executions/{execution_id}
```

### 方式3: 通过 WebSocket 监控

```javascript
// 前端 WebSocket 连接
const ws = new WebSocket('ws://localhost:8080/ws/execution/{execution_id}');

ws.onopen = () => {
  console.log('✅ WebSocket 已连接');

  // 订阅执行进度
  ws.send(JSON.stringify({
    type: 'subscribe',
    topics: ['execution:{execution_id}']
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  switch (message.type) {
    case 'progress':
      console.log(`📊 进度: ${message.data.progress}% - ${message.data.message}`);
      updateProgressUI(message.data);
      break;

    case 'status_change':
      console.log(`🔄 状态变更: ${message.data.status}`);
      updateStatusUI(message.data);
      break;

    case 'error':
      console.error(`❌ 错误: ${message.data.error}`);
      showErrorNotification(message.data);
      break;

    case 'completed':
      console.log('✅ 执行完成', message.data.output);
      showCompletionNotification(message.data);
      break;
  }
};
```

---

## 🎯 可用的预定义模板

### 1. 内容发布流水线 (content-publish-v1)

**描述**: 从内容生成到多平台发布的完整流程

**步骤**:
1. 内容生成 - 使用 AI 生成内容
2. 内容优化 - 优化拼写和可读性
3. 质量评分 - 评估内容质量
4. 发布执行 - 发布到多个平台
5. 数据采集 - 采集发布数据

**预计耗时**: 15-20 分钟

**输入参数**:
```json
{
  "topic": "内容主题",
  "keywords": ["关键词1", "关键词2"],
  "target_audience": "目标受众",
  "platforms": ["douyin", "toutiao", "xiaohongshu"]
}
```

### 2. 视频处理流水线 (video-processing-v1)

**描述**: 视频下载、转录、切片、发布的完整流程

**步骤**:
1. 视频下载 - 下载视频文件
2. 语音转录 - 转录语音为文字
3. 内容改写 - 改写转录内容
4. 视频切片 - 切片视频文件
5. 发布执行 - 发布到平台

**预计耗时**: 20-30 分钟

**输入参数**:
```json
{
  "video_url": "视频URL",
  "output_format": "mp4",
  "max_duration": 60,
  "platforms": ["douyin", "xiaohongshu"]
}
```

### 3. 热点分析流水线 (hotspot-analysis-v1)

**描述**: 抓取热点、分析趋势、生成内容的完整流程

**步骤**:
1. 热点抓取 - 抓取热点数据
2. 趋势分析 - 分析热点趋势
3. 内容生成 - 生成相关内容
4. 发布执行 - 发布到平台

**预计耗时**: 10-15 分钟

**输入参数**:
```json
{
  "keywords": ["AI", "人工智能"],
  "sources": ["newsnow", "toutiao"],
  "platforms": ["douyin", "xiaohongshu"]
}
```

### 4. 数据采集流水线 (data-collection-v1)

**描述**: 从多平台采集发布数据和性能指标

**步骤**:
1. 抖音数据采集 - 采集抖音数据
2. 今日头条数据采集 - 采集头条数据
3. 小红书数据采集 - 采集小红书数据
4. 数据分析 - 分析数据
5. 报告生成 - 生成报告

**预计耗时**: 5-10 分钟

**输入参数**:
```json
{
  "metrics": ["views", "likes", "comments", "shares"],
  "date_range": "7d",
  "format": "markdown"
}
```

---

## 🛠️ 自定义流水线

### 创建自定义流水线

```go
package main

import (
    "context"
    "time"
    "your-project/publisher-core/pipeline"
)

func main() {
    // 创建编排器
    orchestrator := pipeline.NewPipelineOrchestrator(nil)

    // 定义自定义流水线
    customPipeline := &pipeline.Pipeline{
        Name:        "我的自定义流水线",
        Description: "自定义业务流程",
        Steps: []pipeline.PipelineStep{
            {
                ID:      "step-1",
                Name:    "数据采集",
                Type:    pipeline.StepTypeDataCollection,
                Handler: "custom_collector",
                Config: map[string]interface{}{
                    "source": "custom_api",
                    "limit": 100,
                },
                Timeout: 5 * time.Minute,
            },
            {
                ID:        "step-2",
                Name:      "数据处理",
                Type:      pipeline.StepTypeAnalytics,
                Handler:   "custom_processor",
                DependsOn: []string{"step-1"},
                Config: map[string]interface{}{
                    "algorithm": "custom_algo",
                },
                Timeout: 3 * time.Minute,
            },
            {
                ID:        "step-3",
                Name:      "结果输出",
                Type:      pipeline.StepTypePublishExecution,
                Handler:   "custom_publisher",
                DependsOn: []string{"step-2"},
                Config: map[string]interface{}{
                    "output_format": "json",
                },
                Timeout: 2 * time.Minute,
            },
        },
        Config: pipeline.PipelineConfig{
            ParallelMode: false,
            MaxParallel:  1,
            FailFast:     true,
            RetryStrategy: pipeline.RetryStrategy{
                Type:          pipeline.RetryTypeExponential,
                InitialDelay:  1 * time.Second,
                MaxDelay:      30 * time.Second,
                BackoffFactor: 2.0,
            },
        },
    }

    // 注册自定义处理器
    orchestrator.RegisterHandler("custom_collector", &CustomCollector{})
    orchestrator.RegisterHandler("custom_processor", &CustomProcessor{})
    orchestrator.RegisterHandler("custom_publisher", &CustomPublisher{})

    // 创建流水线
    if err := orchestrator.CreatePipeline(customPipeline); err != nil {
        panic(err)
    }

    // 执行流水线
    input := map[string]interface{}{
        "param1": "value1",
        "param2": "value2",
    }

    execution, err := orchestrator.ExecutePipeline(context.Background(), customPipeline.ID, input)
    if err != nil {
        panic(err)
    }

    fmt.Printf("执行ID: %s\n", execution.ID)
}

// 自定义处理器示例
type CustomCollector struct{}

func (h *CustomCollector) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 实现数据采集逻辑
    return map[string]interface{}{
        "data": "collected data",
    }, nil
}

type CustomProcessor struct{}

func (h *CustomProcessor) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 实现数据处理逻辑
    return map[string]interface{}{
        "result": "processed result",
    }, nil
}

type CustomPublisher struct{}

func (h *CustomPublisher) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 实现结果输出逻辑
    return map[string]interface{}{
        "output": "published output",
    }, nil
}
```

---

## 📊 监控与管理

### 查看执行状态

```bash
# 获取执行状态
curl http://localhost:8080/api/v1/executions/{execution_id}

# 获取执行日志
curl http://localhost:8080/api/v1/executions/{execution_id}/logs

# 获取执行进度
curl http://localhost:8080/api/v1/executions/{execution_id}/progress
```

### 管理流水线

```bash
# 列出所有流水线
curl http://localhost:8080/api/v1/pipelines

# 获取流水线详情
curl http://localhost:8080/api/v1/pipelines/{pipeline_id}

# 更新流水线
curl -X PUT http://localhost:8080/api/v1/pipelines/{pipeline_id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新的名称",
    "config": {
      "fail_fast": false
    }
  }'

# 删除流水线
curl -X DELETE http://localhost:8080/api/v1/pipelines/{pipeline_id}
```

### 控制执行

```bash
# 暂停执行
curl -X POST http://localhost:8080/api/v1/executions/{execution_id}/pause

# 恢复执行
curl -X POST http://localhost:8080/api/v1/executions/{execution_id}/resume

# 取消执行
curl -X POST http://localhost:8080/api/v1/executions/{execution_id}/cancel
```

---

## 🔧 配置说明

### 流水线配置

```json
{
  "parallel_mode": false,
  "max_parallel": 1,
  "fail_fast": true,
  "retry_strategy": {
    "type": "exponential",
    "initial_delay": "1s",
    "max_delay": "30s",
    "backoff_factor": 2.0
  },
  "notification": {
    "on_start": true,
    "on_complete": true,
    "on_error": true,
    "channels": ["websocket", "email"]
  }
}
```

**配置说明**:
- `parallel_mode`: 是否并行执行步骤
- `max_parallel`: 最大并行数
- `fail_fast`: 是否在失败时快速终止
- `retry_strategy`: 重试策略
- `notification`: 通知配置

### 步骤配置

```json
{
  "id": "step-1",
  "name": "步骤名称",
  "type": "content_generation",
  "handler": "handler_name",
  "config": {
    "param1": "value1",
    "param2": "value2"
  },
  "depends_on": [],
  "retry_count": 3,
  "timeout": "5m"
}
```

**配置说明**:
- `id`: 步骤唯一标识
- `name`: 步骤名称
- `type`: 步骤类型
- `handler`: 处理器名称
- `config`: 步骤配置
- `depends_on`: 依赖的步骤ID列表
- `retry_count`: 重试次数
- `timeout`: 超时时间

---

## 🐛 故障排查

### 问题1: 流水线执行失败

**症状**: 执行状态显示为 failed

**解决方案**:
1. 查看执行日志：`curl http://localhost:8080/api/v1/executions/{execution_id}/logs`
2. 检查错误信息
3. 确认所有依赖服务正常运行
4. 检查步骤配置是否正确

### 问题2: WebSocket 连接失败

**症状**: 无法连接到 WebSocket

**解决方案**:
1. 检查后端服务是否运行
2. 确认 WebSocket 端口是否开放
3. 检查防火墙设置
4. 查看浏览器控制台错误信息

### 问题3: 进度更新不及时

**症状**: 进度长时间不更新

**解决方案**:
1. 检查 WebSocket 连接状态
2. 确认进度追踪器正常运行
3. 查看后端日志
4. 重启 WebSocket 连接

---

## 📚 更多资源

- [架构设计文档](../architecture/automation-pipeline-design.md)
- [实施总结](../implementation-summary.md)
- [API文档](../api/rest-api.md)
- [开发者指南](../development/developer-guide.md)

---

## 🆘 获取帮助

如果您遇到问题：

1. 查看文档和 FAQ
2. 检查 GitHub Issues
3. 联系技术支持

---

**祝您使用愉快！** 🎉
