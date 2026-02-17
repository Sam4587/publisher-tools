# 项目架构文档

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (React + Vite)                   │
│                      Port: 5173                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐            │
│  │Dashboard│ │Accounts │ │ Publish │ │ History │            │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘            │
│       │           │           │           │                  │
│       └───────────┴─────┬─────┴───────────┘                  │
│                         │ API Calls                          │
│                         ▼                                    │
│                  ┌──────────────┐                            │
│                  │   lib/api.ts │                            │
│                  └──────────────┘                            │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP + Proxy
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   Backend (Go + gorilla/mux)                 │
│                      Port: 8080                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                    API Server                         │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │/platforms  │ │  /tasks    │ │  /publish  │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘        │   │
│  └────────────────────────┬─────────────────────────────┘   │
│                           │                                 │
│  ┌────────────────────────┴─────────────────────────────┐   │
│  │              Publisher Factory                        │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │  Douyin    │ │  Toutiao   │ │Xiaohongshu │        │   │
│  │  │  Adapter   │ │  Adapter   │ │  Adapter   │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘        │   │
│  └────────────────────────┬─────────────────────────────┘   │
│                           │                                 │
│  ┌────────────────────────┴─────────────────────────────┐   │
│  │                 Task Manager                          │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │  Create    │ │  Execute   │ │   Query    │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘        │   │
│  └──────────────────────────────────────────────────────┘   │
│                           │                                 │
│  ┌────────────────────────┴─────────────────────────────┐   │
│  │              Browser Automation (go-rod)              │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │   Login    │ │  Publish   │ │   Cookie   │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘        │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
              ┌───────────────────────┐
              │   Chrome/Chromium     │
              │   (Browser Instance)  │
              └───────────────────────┘
```

## 数据流

### 1. 发布流程

```
用户 → 前端表单 → API请求 → TaskManager创建任务
                                      ↓
                                返回任务ID
                                      ↓
                            后台执行发布任务
                                      ↓
                    Adapter加载Cookie → 启动浏览器 → 导航到发布页
                                      ↓
                              上传文件 → 填写内容 → 点击发布
                                      ↓
                              更新任务状态 → 用户查询结果
```

### 2. 登录流程

```
用户 → 点击登录 → API请求 → Adapter启动浏览器
                                    ↓
                            导航到登录页 → 获取二维码
                                    ↓
                            返回二维码URL → 前端展示
                                    ↓
                            用户扫码 → 检测登录成功
                                    ↓
                            提取Cookie → 保存到本地
```

## 核心接口定义

### Publisher 接口

```go
type Publisher interface {
    // 基础信息
    Platform() string
    
    // 登录相关
    Login(ctx context.Context) (*LoginResult, error)
    WaitForLogin(ctx context.Context) error
    CheckLoginStatus(ctx context.Context) (bool, error)
    
    // 发布相关
    Publish(ctx context.Context, content *Content) (*PublishResult, error)
    PublishAsync(ctx context.Context, content *Content) (string, error)
    QueryStatus(ctx context.Context, taskID string) (*PublishResult, error)
    Cancel(ctx context.Context, taskID string) error
    
    // 资源管理
    Close() error
}
```

### Storage 接口

```go
type Storage interface {
    Write(ctx context.Context, path string, data []byte) error
    Read(ctx context.Context, path string) ([]byte, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    List(ctx context.Context, prefix string) ([]string, error)
    GetURL(ctx context.Context, path string) (string, error)
}
```

### TaskStorage 接口

```go
type TaskStorage interface {
    Save(task *Task) error
    Load(id string) (*Task, error)
    List(filter TaskFilter) ([]*Task, error)
    Delete(id string) error
}
```

## 配置管理

### 后端配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `port` | 8080 | API 服务端口 |
| `headless` | true | 浏览器无头模式 |
| `cookie-dir` | ./cookies | Cookie 存储目录 |
| `storage-dir` | ./uploads | 文件存储目录 |
| `debug` | false | 调试模式 |

### 前端配置

| 参数 | 值 | 说明 |
|------|-----|------|
| API 代理 | localhost:8080 | 后端服务地址 |
| 端口 | 5173 | 开发服务器端口 |

## 错误处理

### API 错误码

| 错误码 | HTTP 状态 | 说明 |
|--------|-----------|------|
| `SERVICE_UNAVAILABLE` | 503 | 服务未初始化 |
| `PLATFORM_NOT_FOUND` | 404 | 平台不存在 |
| `LOGIN_FAILED` | 500 | 登录失败 |
| `TASK_NOT_FOUND` | 404 | 任务不存在 |
| `INVALID_REQUEST` | 400 | 请求格式错误 |
| `NOT_IMPLEMENTED` | 501 | 功能未实现 |

### 错误响应格式

```json
{
  "success": false,
  "error": "错误描述",
  "error_code": "ERROR_CODE",
  "timestamp": 1234567890
}
```

## 扩展指南

### 添加新平台

1. 在 `publisher-core/adapters/adapters.go` 中创建新适配器：

```go
type NewPlatformAdapter struct {
    BaseAdapter
}

func NewNewPlatformAdapter(opts *publisher.Options) *NewPlatformAdapter {
    base := NewBaseAdapter("newplatform", opts)
    base.loginURL = "https://..."
    base.publishURL = "https://..."
    base.domain = ".newplatform.com"
    base.cookieKeys = cookies.NewPlatformCookieKeys
    // 设置内容限制
    return &NewPlatformAdapter{BaseAdapter: *base}
}

// 实现必要的方法
func (a *NewPlatformAdapter) doPublish(ctx context.Context, content *publisher.Content) error {
    // 实现发布逻辑
}
```

2. 在工厂中注册：

```go
func DefaultFactory() *PublisherFactory {
    f := NewPublisherFactory()
    f.Register("newplatform", func(opts *publisher.Options) publisher.Publisher {
        return NewNewPlatformAdapter(opts)
    })
    return f
}
```

### 添加新 API 端点

1. 在 `publisher-core/api/server.go` 中添加路由：

```go
func (s *Server) setupRoutes() {
    // ...
    s.router.HandleFunc("/api/v1/new-endpoint", s.newHandler).Methods("GET")
}
```

2. 实现处理函数：

```go
func (s *Server) newHandler(w http.ResponseWriter, r *http.Request) {
    // 处理逻辑
    s.jsonSuccess(w, result)
}
```

### 添加前端页面

1. 在 `publisher-web/src/pages/` 创建组件：

```tsx
export default function NewPage() {
  return <div>New Page</div>
}
```

2. 在 `App.tsx` 中添加路由：

```tsx
<Route path="/new-page" element={<NewPage />} />
```

3. 在 `Navbar.tsx` 中添加导航链接。
