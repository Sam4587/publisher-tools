# 自动化流水线系统集成指南

## 📋 概述

本指南将帮助您将自动化流水线系统集成到现有的 publisher-tools 项目中，实现完整的**内容生成 → 发布 → 监控**业务链。

---

## 🔗 系统集成架构

### 集成点概览

```
现有系统
    ↓
┌─────────────────────────────────────────┐
│  现有组件                               │
│  - TaskManager (任务管理)              │
│  - AIService (AI服务)                  │
│  - PublisherManager (平台发布)         │
│  - AnalyticsService (数据分析)         │
└─────────────────────────────────────────┘
    ↓ 集成
┌─────────────────────────────────────────┐
│  新增组件                               │
│  - EnhancedTaskManager (增强任务管理)  │
│  - PipelineOrchestrator (流水线编排)   │
│  - WebSocketServer (实时通信)          │
│  - PipelineTemplates (预定义模板)      │
└─────────────────────────────────────────┘
    ↓ 前端集成
┌─────────────────────────────────────────┐
│  前端界面                               │
│  - PipelineManagement (流水线管理)     │
│  - MonitoringDashboard (实时监控)       │
│  - ExecutionDetail (执行详情)           │
└─────────────────────────────────────────┘
```

---

## 🚀 集成步骤

### 步骤1: 后端集成

#### 1.1 更新 go.mod 依赖

```bash
cd publisher-core

# 添加新的依赖
go get github.com/gorilla/websocket
go get github.com/google/uuid
```

#### 1.2 集成增强版任务管理器

在 `publisher-core/cmd/server/main.go` 中：

```go
package main

import (
    "your-project/publisher-core/task"
    "your-project/publisher-core/pipeline"
    "your-project/publisher-core/websocket"
    // ... 其他导入
)

func main() {
    // 创建增强版任务管理器
    enhancedTaskManager := task.NewEnhancedTaskManager(task.NewMemoryStorage())

    // 创建流水线编排器
    orchestrator := pipeline.NewPipelineOrchestrator(nil)

    // 注册预定义模板
    templates := pipeline.ListTemplates()
    for _, tmpl := range templates {
        orchestrator.CreatePipeline(tmpl)
    }

    // 创建 WebSocket 服务器
    wsServer := websocket.NewServer()

    // 注册步骤处理器
    registerStepHandlers(orchestrator)

    // 启动 HTTP 服务器
    setupRoutes(orchestrator, enhancedTaskManager, wsServer)
}
```

#### 1.3 注册步骤处理器

创建 `publisher-core/pipeline/handlers.go`：

```go
package pipeline

import (
    "context"
    "fmt"
    "your-project/publisher-core/ai"
    "your-project/publisher-core/adapters"
    "your-project/publisher-core/analytics"
)

// registerStepHandlers 注册所有步骤处理器
func registerStepHandlers(orchestrator *PipelineOrchestrator) {
    // AI 内容生成处理器
    orchestrator.RegisterHandler("ai_content_generator", &AIContentGenerator{})

    // 内容优化处理器
    orchestrator.RegisterHandler("content_optimizer", &ContentOptimizer{})

    // 质量评分处理器
    orchestrator.RegisterHandler("quality_scorer", &QualityScorer{})

    // 平台发布处理器
    orchestrator.RegisterHandler("platform_publisher", &PlatformPublisher{})

    // 数据采集处理器
    orchestrator.RegisterHandler("analytics_collector", &AnalyticsCollector{})
}

// AIContentGenerator AI内容生成处理器
type AIContentGenerator struct {
    aiService *ai.AIService
}

func (h *AIContentGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    model := config["model"].(string)
    topic := input["topic"].(string)
    keywords := input["keywords"].([]string)

    prompt := fmt.Sprintf("主题: %s\n关键词: %v\n请生成一篇相关内容", topic, keywords)

    result, err := h.aiService.Generate(ctx, ai.GenerateRequest{
        Model: model,
        Prompt: prompt,
    })
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "content": result.Content,
        "tokens_used": result.TokensUsed,
    }, nil
}

// ContentOptimizer 内容优化处理器
type ContentOptimizer struct{}

func (h *ContentOptimizer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    content := input["content"].(string)

    // 实现内容优化逻辑
    optimizedContent := content // 实际实现中会进行优化

    return map[string]interface{}{
        "optimized_content": optimizedContent,
    }, nil
}

// QualityScorer 质量评分处理器
type QualityScorer struct{}

func (h *QualityScorer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    content := input["optimized_content"].(string)

    // 实现质量评分逻辑
    score := 0.85 // 实际实现中会计算真实分数

    return map[string]interface{}{
        "score": score,
        "passed": score >= 0.7,
    }, nil
}

// PlatformPublisher 平台发布处理器
type PlatformPublisher struct {
    publisherManager *adapters.PublisherManager
}

func (h *PlatformPublisher) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    platforms := config["platforms"].([]string)
    content := input["optimized_content"].(string)

    results := make(map[string]interface{})
    for _, platform := range platforms {
        adapter := h.publisherManager.GetAdapter(platform)
        result, err := adapter.Publish(ctx, adapters.PublishRequest{
            Type:    "article",
            Title:   input["topic"].(string),
            Content: content,
        })
        if err != nil {
            return nil, fmt.Errorf("发布到 %s 失败: %w", platform, err)
        }
        results[platform] = result
    }

    return results, nil
}

// AnalyticsCollector 数据采集处理器
type AnalyticsCollector struct {
    analyticsService *analytics.AnalyticsService
}

func (h *AnalyticsCollector) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
    // 实现数据采集逻辑
    return map[string]interface{}{
        "collected": true,
    }, nil
}
```

#### 1.4 设置 API 路由

创建 `publisher-core/api/pipeline_routes.go`：

```go
package api

import (
    "encoding/json"
    "net/http"
    "your-project/publisher-core/pipeline"
    "your-project/publisher-core/websocket"
)

func SetupPipelineRoutes(mux *http.ServeMux, orchestrator *pipeline.PipelineOrchestrator, wsServer *websocket.Server) {
    // 流水线管理
    mux.HandleFunc("/api/v1/pipelines", handlePipelines(orchestrator))
    mux.HandleFunc("/api/v1/pipelines/", handlePipelineDetail(orchestrator))

    // 流水线执行
    mux.HandleFunc("/api/v1/pipelines/", func(w http.ResponseWriter, r *http.Request) {
        // 处理执行、暂停、恢复、取消
    })

    // 执行管理
    mux.HandleFunc("/api/v1/executions", handleExecutions(orchestrator))
    mux.HandleFunc("/api/v1/executions/", handleExecutionDetail(orchestrator))

    // WebSocket
    mux.HandleFunc("/ws/monitor", wsServer.HandleWebSocket)
}

func handlePipelines(orchestrator *pipeline.PipelineOrchestrator) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            pipelines, _ := orchestrator.ListPipelines()
            json.NewEncoder(w).Encode(pipelines)
        case "POST":
            var p pipeline.Pipeline
            json.NewDecoder(r.Body).Decode(&p)
            orchestrator.CreatePipeline(&p)
            json.NewEncoder(w).Encode(p)
        }
    }
}

// ... 其他路由处理函数
```

### 步骤2: 前端集成

#### 2.1 安装前端依赖

```bash
cd publisher-web

# 安装 Ant Design
npm install antd @ant-design/icons

# 安装 WebSocket 客户端
npm install @types/ws
```

#### 2.2 更新路由配置

在 `publisher-web/src/App.tsx` 中：

```typescript
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import PipelineManagement from './pages/PipelineManagement';
import MonitoringDashboard from './pages/MonitoringDashboard';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/pipelines" element={<PipelineManagement />} />
        <Route path="/monitoring" element={<MonitoringDashboard />} />
        {/* 其他路由 */}
      </Routes>
    </Router>
  );
}

export default App;
```

#### 2.3 更新导航菜单

在主布局组件中添加导航：

```typescript
import { Menu } from 'antd';
import { Link } from 'react-router-dom';

const menuItems = [
  {
    key: 'pipelines',
    label: <Link to="/pipelines">流水线管理</Link>,
    icon: <AppstoreOutlined />,
  },
  {
    key: 'monitoring',
    label: <Link to="/monitoring">实时监控</Link>,
    icon: <DashboardOutlined />,
  },
  // ... 其他菜单项
];
```

### 步骤3: 测试集成

#### 3.1 后端测试

```bash
# 启动后端服务
cd publisher-core
go run cmd/server/main.go

# 测试 API
curl http://localhost:8080/api/v1/pipelines
curl http://localhost:8080/api/v1/pipeline-templates
```

#### 3.2 前端测试

```bash
# 启动前端服务
cd publisher-web
npm run dev

# 访问应用
open http://localhost:5173
```

#### 3.3 WebSocket 测试

```javascript
// 在浏览器控制台测试
const ws = new WebSocket('ws://localhost:8080/ws/monitor');

ws.onopen = () => {
  console.log('Connected');
  ws.send(JSON.stringify({
    type: 'subscribe',
    topics: ['monitor']
  }));
};

ws.onmessage = (event) => {
  console.log('Message:', JSON.parse(event.data));
};
```

---

## 🔧 配置说明

### 后端配置

在 `publisher-core/config/config.go` 中添加：

```go
type Config struct {
    // ... 现有配置

    // 流水线配置
    Pipeline PipelineConfig `yaml:"pipeline"`
}

type PipelineConfig struct {
    // WebSocket配置
    WebSocket WebSocketConfig `yaml:"websocket"`

    // 默认重试策略
    DefaultRetryStrategy RetryStrategy `yaml:"default_retry_strategy"`

    // 最大并行数
    MaxParallel int `yaml:"max_parallel"`
}

type WebSocketConfig struct {
    Enabled bool   `yaml:"enabled"`
    Port    int    `yaml:"port"`
    Path    string `yaml:"path"`
}
```

### 前端配置

在 `publisher-web/src/config.ts` 中添加：

```typescript
export const config = {
  api: {
    baseURL: 'http://localhost:8080',
    timeout: 30000,
  },
  websocket: {
    url: 'ws://localhost:8080/ws/monitor',
    reconnectInterval: 5000,
    maxReconnectAttempts: 5,
  },
  pipeline: {
    refreshInterval: 5000, // 刷新间隔（毫秒）
  },
};
```

---

## 🧪 集成测试

### 测试清单

- [ ] 后端 API 测试
  - [ ] 流水线列表查询
  - [ ] 流水线创建
  - [ ] 流水线执行
  - [ ] 执行状态查询
  - [ ] 执行日志查询

- [ ] WebSocket 测试
  - [ ] 连接测试
  - [ ] 订阅测试
  - [ ] 消息推送测试
  - [ ] 断线重连测试

- [ ] 前端功能测试
  - [ ] 流水线管理页面
  - [ ] 实时监控面板
  - [ ] 执行详情页面
  - [ ] WebSocket 实时更新

- [ ] 端到端测试
  - [ ] 创建流水线
  - [ ] 执行流水线
  - [ ] 监控进度
  - [ ] 查看结果

### 测试脚本

创建 `scripts/integration-test.sh`：

```bash
#!/bin/bash

echo "开始集成测试..."

# 1. 启动后端
echo "启动后端服务..."
cd publisher-core
go run cmd/server/main.go &
BACKEND_PID=$!
sleep 5

# 2. 启动前端
echo "启动前端服务..."
cd ../publisher-web
npm run dev &
FRONTEND_PID=$!
sleep 5

# 3. 运行测试
echo "运行 API 测试..."
curl http://localhost:8080/api/v1/pipelines
curl http://localhost:8080/api/v1/pipeline-templates

# 4. 清理
echo "清理进程..."
kill $BACKEND_PID $FRONTEND_PID

echo "集成测试完成！"
```

---

## 📚 相关文档

- [快速开始指南](./automation-pipeline-quickstart.md)
- [架构设计文档](../architecture/automation-pipeline-design.md)
- [实施总结](../implementation-summary.md)
- [API文档](../api/rest-api.md)

---

## 🆘 故障排查

### 问题1: WebSocket 连接失败

**症状**: 前端无法连接到 WebSocket

**解决方案**:
1. 检查后端服务是否运行
2. 确认 WebSocket 端口正确
3. 检查防火墙设置
4. 查看浏览器控制台错误

### 问题2: 流水线执行失败

**症状**: 执行状态显示为 failed

**解决方案**:
1. 查看执行日志
2. 检查步骤处理器是否正确注册
3. 确认所有依赖服务正常运行
4. 检查配置参数

### 问题3: 进度更新不及时

**症状**: 进度长时间不更新

**解决方案**:
1. 检查 WebSocket 连接状态
2. 确认进度追踪器正常运行
3. 查看后端日志
4. 重启 WebSocket 连接

---

## 🎯 下一步

集成完成后，您可以：

1. **自定义流水线** - 根据业务需求创建自定义流水线
2. **添加更多步骤处理器** - 扩展功能支持
3. **优化性能** - 根据实际使用情况优化
4. **部署到生产环境** - 配置生产环境部署

---

**祝集成顺利！** 🎉
