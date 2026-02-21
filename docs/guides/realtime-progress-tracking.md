# 实时进度追踪系统

## 概述

实时进度追踪系统基于WebSocket实现，提供任务进度的实时推送、断线重连、历史记录等功能。该系统与异步任务系统（AI-019）紧密集成，为用户提供实时的任务执行反馈。

## 架构设计

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                      前端应用                                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         RealtimeProgressTracker 组件                 │   │
│  │  - WebSocket连接管理                                 │   │
│  │  - 自动重连机制                                       │   │
│  │  - 进度状态展示                                       │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      后端服务                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ WebSocket    │  │ Progress     │  │ Reconnection │      │
│  │ Hub          │  │ Service      │  │ Manager      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                 │                 │               │
│         └─────────────────┼─────────────────┘               │
│                           ▼                                 │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              数据库 (ProgressHistory)                 │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 文件结构

```
publisher-core/websocket/
├── server.go          # WebSocket服务器基础实现
├── hub.go             # 连接池管理和广播
├── progress.go        # 进度服务（持久化、历史记录）
└── handler.go         # HTTP处理器和API端点

publisher-web/src/components/
└── RealtimeProgressTracker.tsx  # 前端进度组件
```

## WebSocket协议

### 连接端点

```
WebSocket URL: ws://host/api/v1/ws
重连端点: ws://host/api/v1/ws/reconnect
```

### 连接参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | string | 否 | 用户ID，默认为"anonymous" |
| project_id | string | 否 | 项目ID |
| old_client_id | string | 否 | 重连时提供的旧客户端ID |

### 消息格式

所有消息均为JSON格式：

```json
{
  "type": "消息类型",
  "task_id": "任务ID",
  "payload": {}
}
```

### 消息类型

#### 1. 连接消息

**connected** - 连接成功
```json
{
  "type": "connected",
  "task_id": "",
  "payload": {
    "client_id": "client-123456",
    "message": "WebSocket连接成功",
    "timestamp": "2026-02-21T10:00:00Z"
  }
}
```

**reconnected** - 重连成功
```json
{
  "type": "reconnected",
  "task_id": "",
  "payload": {
    "client_id": "client-789012",
    "old_client_id": "client-123456",
    "message": "WebSocket重连成功",
    "subscribed": {"task-001": true},
    "timestamp": "2026-02-21T10:05:00Z"
  }
}
```

#### 2. 订阅消息

**subscribe** - 订阅任务（客户端发送）
```json
{
  "action": "subscribe",
  "task_id": "task-001"
}
```

**subscribed** - 订阅确认（服务端响应）
```json
{
  "type": "subscribed",
  "task_id": "",
  "payload": {
    "topics": ["task-001"]
  }
}
```

**unsubscribe** - 取消订阅（客户端发送）
```json
{
  "action": "unsubscribe",
  "task_id": "task-001"
}
```

#### 3. 进度消息

**progress** - 进度更新
```json
{
  "type": "progress",
  "task_id": "task-001",
  "payload": {
    "task_id": "task-001",
    "progress": 45,
    "current_step": "处理视频片段",
    "total_steps": 10,
    "completed_steps": 4,
    "message": "正在处理第4个片段...",
    "status": "running",
    "timestamp": "2026-02-21T10:01:30Z"
  }
}
```

**status** - 状态变更
```json
{
  "type": "status",
  "task_id": "task-001",
  "payload": {
    "task_id": "task-001",
    "status": "completed",
    "error": "",
    "result": {"output": "..."},
    "timestamp": "2026-02-21T10:02:00Z"
  }
}
```

#### 4. 心跳消息

**ping** - 心跳请求（客户端发送）
```json
{
  "type": "ping"
}
```

**pong** - 心跳响应（服务端响应）
```json
{
  "type": "pong",
  "task_id": "",
  "payload": {
    "timestamp": 1708512000
  }
}
```

#### 5. 错误消息

**error** - 错误通知
```json
{
  "type": "error",
  "task_id": "task-001",
  "payload": {
    "execution_id": "task-001",
    "step_id": "step-3",
    "error": "处理失败：网络超时",
    "timestamp": "2026-02-21T10:01:45Z"
  }
}
```

## REST API

### 进度管理API

#### 获取任务进度

```http
GET /api/v1/progress/:task_id
```

**响应示例：**
```json
{
  "task_id": "task-001",
  "progress": 45,
  "current_step": "处理视频片段",
  "total_steps": 10,
  "completed_steps": 4,
  "message": "正在处理第4个片段...",
  "status": "running",
  "timestamp": "2026-02-21T10:01:30Z"
}
```

#### 获取进度历史（内存）

```http
GET /api/v1/progress/:task_id/history?limit=20
```

**响应示例：**
```json
{
  "task_id": "task-001",
  "source": "memory",
  "history": [
    {
      "task_id": "task-001",
      "progress": 45,
      "current_step": "处理视频片段",
      "total_steps": 10,
      "completed_steps": 4,
      "message": "正在处理第4个片段...",
      "status": "running",
      "timestamp": "2026-02-21T10:01:30Z"
    }
  ]
}
```

#### 获取进度历史（数据库）

```http
GET /api/v1/progress/:task_id/history/db?limit=100
```

**响应示例：**
```json
{
  "task_id": "task-001",
  "source": "database",
  "history": [...]
}
```

#### 清除进度历史

```http
DELETE /api/v1/progress/:task_id/history
```

**响应示例：**
```json
{
  "message": "历史已清除",
  "task_id": "task-001"
}
```

#### 获取统计信息

```http
GET /api/v1/progress/stats
```

**响应示例：**
```json
{
  "total_tasks": 100,
  "active_tasks": 5,
  "completed_tasks": 90,
  "failed_tasks": 5
}
```

#### 获取活跃任务

```http
GET /api/v1/progress/active
```

**响应示例：**
```json
{
  "connected_clients": 10,
  "by_user": {
    "user-001": 3,
    "user-002": 2
  },
  "by_project": {
    "project-001": 5,
    "project-002": 3
  }
}
```

#### 订阅任务

```http
POST /api/v1/progress/:task_id/subscribe?client_id=client-123
```

**响应示例：**
```json
{
  "message": "订阅成功",
  "task_id": "task-001",
  "client_id": "client-123"
}
```

#### 取消订阅

```http
POST /api/v1/progress/:task_id/unsubscribe?client_id=client-123
```

### WebSocket统计API

#### 获取WebSocket统计

```http
GET /api/v1/websocket/stats
```

**响应示例：**
```json
{
  "total_clients": 10,
  "by_user": {
    "user-001": 3,
    "user-002": 2
  },
  "by_project": {
    "project-001": 5,
    "project-002": 3
  }
}
```

## 前端组件使用

### 基本使用

```tsx
import RealtimeProgressTracker from '@/components/RealtimeProgressTracker'

function MyComponent() {
  const handleProgress = (progress) => {
    console.log('进度更新:', progress)
  }

  const handleComplete = (progress) => {
    console.log('任务完成:', progress)
  }

  const handleError = (error) => {
    console.error('错误:', error)
  }

  return (
    <RealtimeProgressTracker
      taskId="task-001"
      userId="user-123"
      projectId="project-456"
      onProgress={handleProgress}
      onComplete={handleComplete}
      onError={handleError}
    />
  )
}
```

### 组件属性

| 属性 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| taskId | string | 是 | - | 任务ID |
| userId | string | 否 | 'anonymous' | 用户ID |
| projectId | string | 否 | - | 项目ID |
| wsUrl | string | 否 | 自动生成 | WebSocket URL |
| onProgress | function | 否 | - | 进度更新回调 |
| onComplete | function | 否 | - | 任务完成回调 |
| onError | function | 否 | - | 错误回调 |
| onReconnect | function | 否 | - | 重连成功回调 |
| showHistory | boolean | 否 | false | 是否显示历史记录 |
| autoReconnect | boolean | 否 | true | 是否自动重连 |
| maxReconnectAttempts | number | 否 | 5 | 最大重连次数 |

### 高级功能

#### 显示进度历史

```tsx
<RealtimeProgressTracker
  taskId="task-001"
  showHistory={true}
/>
```

#### 自定义重连策略

```tsx
<RealtimeProgressTracker
  taskId="task-001"
  autoReconnect={true}
  maxReconnectAttempts={10}
  onReconnect={(clientId) => {
    console.log('重连成功，新客户端ID:', clientId)
  }}
/>
```

## 断线重连机制

### 工作流程

1. **连接断开检测**：WebSocket onclose事件触发
2. **指数退避重连**：延迟时间按指数增长（1s, 2s, 4s, 8s...），最大30秒
3. **会话恢复**：使用old_client_id恢复之前的订阅状态
4. **消息补发**：发送断线期间的消息队列

### 会话状态保存

```go
type SessionState struct {
    ClientID    string
    UserID      string
    ProjectID   string
    Subscribed  map[string]bool  // 订阅的任务ID
    LastActive  time.Time
    MessageQueue [][]byte         // 断线期间的消息队列
}
```

### 重连示例

```javascript
// 前端重连逻辑
const ws = new WebSocket(`${wsUrl}?old_client_id=${oldClientId}`)

ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  if (message.type === 'reconnected') {
    // 恢复订阅状态
    console.log('已恢复订阅:', message.payload.subscribed)
    // 接收断线期间的消息
  }
}
```

## 数据库模型

### ProgressHistory

```go
type ProgressHistory struct {
    ID             uint      `gorm:"primaryKey"`
    TaskID         string    `gorm:"index;size:100;not null"`
    Progress       int       // 进度百分比 (0-100)
    CurrentStep    string    `gorm:"size:200"`
    TotalSteps     int
    CompletedSteps int
    Message        string    `gorm:"size:500"`
    Status         string    `gorm:"size:20;index"`
    Metadata       string    `gorm:"type:text"`
    CreatedAt      time.Time `gorm:"index"`
}
```

### WebSocketSession

```go
type WebSocketSession struct {
    ID             uint       `gorm:"primaryKey"`
    ClientID       string     `gorm:"uniqueIndex;size:100;not null"`
    UserID         string     `gorm:"size:100;index"`
    ProjectID      string     `gorm:"size:100;index"`
    ConnectedAt    time.Time
    DisconnectedAt *time.Time
    LastActiveAt   time.Time
    MessageCount   int
    IPAddress      string     `gorm:"size:50"`
    UserAgent      string     `gorm:"size:500"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

## 性能优化

### 1. 消息批处理

```go
// 批量发送消息，减少系统调用
func (c *Client) WritePump() {
    for {
        select {
        case message := <-c.Send:
            w, _ := c.Conn.NextWriter(websocket.TextMessage)
            w.Write(message)
            
            // 批量发送队列中的消息
            n := len(c.Send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.Send)
            }
            w.Close()
        }
    }
}
```

### 2. 连接池管理

```go
// Hub管理所有客户端连接
type Hub struct {
    Clients    map[string]*Client
    Broadcast  chan *Message
    Register   chan *Client
    Unregister chan *Client
    mu         sync.RWMutex
}
```

### 3. 历史记录限制

```go
// 限制内存中的历史记录数量
const MaxHistoryPerTask = 100

// 定期清理过期记录
func (s *ProgressService) CleanupOldHistory() {
    cutoff := time.Now().Add(-s.config.HistoryTTL)
    // 清理逻辑...
}
```

## 监控与调试

### 日志级别

```go
// 连接日志
logrus.Infof("WebSocket客户端已连接: %s", client.ID)

// 订阅日志
logrus.Infof("客户端 %s 订阅任务: %s", client.ID, taskID)

// 重连日志
logrus.Infof("客户端重连成功: %s -> %s", oldClientID, newClientID)

// 错误日志
logrus.Errorf("WebSocket升级失败: %v", err)
```

### 监控指标

- 连接客户端数量
- 按用户/项目分组的连接数
- 消息发送速率
- 重连成功率
- 平均消息延迟

## 最佳实践

### 1. 错误处理

```javascript
// 前端错误处理
const handleError = (error) => {
  if (error.includes('连接错误')) {
    // 尝试重连
    handleReconnect()
  } else if (error.includes('认证失败')) {
    // 跳转登录
    redirectToLogin()
  }
}
```

### 2. 资源清理

```javascript
// 组件卸载时断开连接
useEffect(() => {
  connect()
  return () => {
    disconnect()
  }
}, [])
```

### 3. 心跳保活

```javascript
// 定期发送心跳
useEffect(() => {
  const interval = setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'ping' }))
    }
  }, 30000)
  
  return () => clearInterval(interval)
}, [])
```

## 故障排查

### 常见问题

1. **连接失败**
   - 检查WebSocket URL是否正确
   - 检查网络连接和防火墙设置
   - 查看服务器日志

2. **频繁断线**
   - 检查心跳间隔设置
   - 检查服务器负载
   - 查看网络稳定性

3. **消息丢失**
   - 确认订阅状态
   - 检查消息队列大小
   - 查看断线重连日志

### 调试技巧

```javascript
// 启用详细日志
const ws = new WebSocket(url)
ws.onopen = () => console.log('WebSocket已连接')
ws.onmessage = (e) => console.log('收到消息:', e.data)
ws.onerror = (e) => console.error('WebSocket错误:', e)
ws.onclose = (e) => console.log('WebSocket已关闭:', e.code, e.reason)
```

## 相关文档

- [异步任务系统指南](./async-task-system-guide.md)
- [API接口文档](../api/rest-api.md)
- [开发者指南](../development/developer-guide.md)
