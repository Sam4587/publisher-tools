# Publisher Core - 多平台内容发布核心库

[English](#english) | [中文](#中文)

---

## 中文

### 概述

Publisher Core 是一个统一的多平台内容发布核心库，采用适配器模式设计，支持抖音、今日头条、小红书等多个平台的统一接口调用。

### 架构设计

```
publisher-core/
├── interfaces/          # 接口定义
│   └── publisher.go     # Publisher 接口和公共类型
├── adapters/            # 平台适配器实现
│   └── adapters.go      # 抖音/头条/小红书适配器
├── task/                # 异步任务管理
│   └── manager.go       # 任务创建、执行、状态查询
├── storage/             # 文件存储抽象
│   └── storage.go       # 本地/内存存储实现
├── api/                 # REST API 服务
│   └── server.go        # HTTP 服务和路由
└── cmd/
    ├── server/          # 服务入口
    └── cli/             # 命令行工具
```

### 核心特性

#### 1. 统一发布器接口

所有平台发布器实现统一的 `Publisher` 接口：

```go
type Publisher interface {
    Platform() string                              // 平台名称
    Login(ctx) (*LoginResult, error)               // 登录
    WaitForLogin(ctx) error                        // 等待登录完成
    CheckLoginStatus(ctx) (bool, error)            // 检查登录状态
    Publish(ctx, *Content) (*PublishResult, error) // 同步发布
    PublishAsync(ctx, *Content) (string, error)    // 异步发布
    QueryStatus(ctx, taskID) (*PublishResult, error) // 查询状态
    Cancel(ctx, taskID) error                      // 取消任务
    Close() error                                  // 关闭资源
}
```

#### 2. 适配器模式

采用工厂模式创建各平台发布器：

```go
// 创建工厂
factory := adapters.DefaultFactory()

// 创建抖音发布器
douyinPub, _ := factory.Create("douyin")

// 创建小红书发布器
xhsPub, _ := factory.Create("xiaohongshu")

// 创建今日头条发布器
toutiaoPub, _ := factory.Create("toutiao")
```

#### 3. 异步任务处理

耗时操作支持异步执行，立即返回任务 ID：

```go
// 创建任务管理器
taskMgr := task.NewTaskManager(task.NewMemoryStorage())

// 异步发布
taskID, _ := pub.PublishAsync(ctx, content)

// 查询任务状态
result, _ := pub.QueryStatus(ctx, taskID)

// 取消任务
pub.Cancel(ctx, taskID)
```

#### 4. 文件存储抽象

统一的文件存储接口，支持本地存储和云存储：

```go
// 创建本地存储
store, _ := storage.NewLocalStorage("./uploads", "http://localhost:8080")

// 写入文件
store.Write(ctx, "images/photo.jpg", imageData)

// 读取文件
data, _ := store.Read(ctx, "images/photo.jpg")

// 获取访问 URL
url, _ := store.GetURL(ctx, "images/photo.jpg")
```

### 使用方式

#### 方式一：REST API 服务

```bash
# 启动 API 服务
go run ./cmd/server -port 8080

# 检查健康状态
curl http://localhost:8080/health

# 获取支持的平台列表
curl http://localhost:8080/api/v1/platforms

# 登录平台
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login

# 检查登录状态
curl http://localhost:8080/api/v1/platforms/douyin/check

# 创建发布任务
curl -X POST http://localhost:8080/api/v1/publish/async \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "douyin",
    "type": "images",
    "title": "今日分享",
    "content": "美好的一天",
    "images": ["photo1.jpg", "photo2.jpg"],
    "tags": ["生活", "日常"]
  }'

# 查询任务状态
curl http://localhost:8080/api/v1/tasks/{taskId}
```

#### 方式二：命令行工具

```bash
# 编译
go build -o publisher ./cmd/cli

# 登录
./publisher -platform douyin -login
./publisher -platform xiaohongshu -login

# 检查登录状态
./publisher -platform douyin -check

# 发布图文
./publisher -platform douyin \
  -title "今日分享" \
  -content "美好的一天" \
  -images "photo1.jpg,photo2.jpg" \
  -tags "生活,日常"

# 发布视频
./publisher -platform douyin \
  -title "旅行Vlog" \
  -content "记录美好时光" \
  -video "travel.mp4" \
  -tags "旅行,vlog"

# 异步发布
./publisher -platform douyin \
  -title "标题" \
  -content "内容" \
  -video "video.mp4" \
  -async

# 查询任务状态
./publisher -task-id <task_id> -status

# 列出任务
./publisher -list
```

#### 方式三：作为库使用

```go
package main

import (
    "context"
    "time"

    "github.com/monkeycode/publisher-core/adapters"
    publisher "github.com/monkeycode/publisher-core/interfaces"
)

func main() {
    // 创建工厂
    factory := adapters.DefaultFactory()

    // 创建发布器
    pub, _ := factory.Create("douyin")
    defer pub.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    // 检查登录
    loggedIn, _ := pub.CheckLoginStatus(ctx)
    if !loggedIn {
        // 登录
        result, _ := pub.Login(ctx)
        if result.QrcodeURL != "" {
            println("请扫码登录:", result.QrcodeURL)
        }
        pub.WaitForLogin(ctx)
    }

    // 发布内容
    content := &publisher.Content{
        Type:       publisher.ContentTypeImages,
        Title:      "今日分享",
        Body:       "美好的一天",
        ImagePaths: []string{"photo1.jpg", "photo2.jpg"},
        Tags:       []string{"生活", "日常"},
    }

    result, err := pub.Publish(ctx, content)
    if err != nil {
        panic(err)
    }

    println("发布状态:", string(result.Status))
}
```

### API 文档

#### 平台接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/platforms` | GET | 获取支持的平台列表 |
| `/api/v1/platforms/{platform}` | GET | 获取平台信息 |
| `/api/v1/platforms/{platform}/login` | POST | 登录平台 |
| `/api/v1/platforms/{platform}/check` | GET | 检查登录状态 |

#### 任务接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/tasks` | POST | 创建任务 |
| `/api/v1/tasks` | GET | 列出任务 |
| `/api/v1/tasks/{taskId}` | GET | 获取任务详情 |
| `/api/v1/tasks/{taskId}/cancel` | POST | 取消任务 |

#### 发布接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/publish` | POST | 同步发布 |
| `/api/v1/publish/async` | POST | 异步发布 |

### 内容限制

| 平台 | 标题 | 正文 | 图片 | 视频 |
|------|------|------|------|------|
| 抖音 | 30字 | 2000字 | 最多12张 | 4GB |
| 今日头条 | 30字 | 2000字 | 多张 | 无限制 |
| 小红书 | 20字 | 1000字 | 最多18张 | 500MB |

---

## English

### Overview

Publisher Core is a unified multi-platform content publishing library designed with adapter pattern, supporting unified interface calls for Douyin, Toutiao, Xiaohongshu and other platforms.

### Architecture

```
publisher-core/
├── interfaces/          # Interface definitions
│   └── publisher.go     # Publisher interface and common types
├── adapters/            # Platform adapter implementations
│   └── adapters.go      # Douyin/Toutiao/Xiaohongshu adapters
├── task/                # Async task management
│   └── manager.go       # Task creation, execution, status query
├── storage/             # File storage abstraction
│   └── storage.go       # Local/memory storage implementations
├── api/                 # REST API service
│   └── server.go        # HTTP server and routes
└── cmd/
    ├── server/          # Server entry point
    └── cli/             # Command line tool
```

### Key Features

- **Unified Publisher Interface**: All platforms implement the same `Publisher` interface
- **Adapter Pattern**: Easy to add new platform support
- **Async Task Processing**: Long-running operations return task ID immediately
- **File Storage Abstraction**: Unified interface for local and cloud storage
- **REST API Service**: HTTP endpoints for integration
- **CLI Tool**: Command line interface for quick operations

### License

MIT License
