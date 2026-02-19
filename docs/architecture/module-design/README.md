# 模块详细设计

## 目录

- [平台适配器设计](#平台适配器设计)
- [AI服务设计](#ai服务设计)
- [任务管理系统设计](#任务管理系统设计)
- [浏览器自动化设计](#浏览器自动化设计)
- [数据分析模块设计](#数据分析模块设计)
- [热点监控设计](#热点监控设计)
- [Cookie管理设计](#cookie管理设计)
- [文件存储设计](#文件存储设计)

## 平台适配器设计

### 设计原则
采用适配器模式，为每个平台实现统一的发布接口。

### 核心接口
```go
type Publisher interface {
    Platform() string
    Login(ctx context.Context) (*LoginResult, error)
    WaitForLogin(ctx context.Context) error
    CheckLoginStatus(ctx context.Context) (bool, error)
    Publish(ctx context.Context, content *Content) (*PublishResult, error)
    PublishAsync(ctx context.Context, content *Content) (string, error)
    QueryStatus(ctx context.Context, taskID string) (*PublishResult, error)
    Cancel(ctx context.Context, taskID string) error
    Close() error
}
```

### 实现架构
```
adapters/
├── adapters.go          # 工厂模式实现
├── douyin/              # 抖音适配器
│   ├── login.go
│   └── publish.go
├── toutiao/             # 今日头条适配器
│   ├── login.go
│   └── publish.go
└── xiaohongshu/         # 小红书适配器
    ├── login.go
    └── publish.go
```

## AI服务设计

### 服务架构
```
ai/
├── provider/            # AI提供商适配器
│   ├── deepseek.go
│   ├── google.go
│   ├── groq.go
│   └── openrouter.go
├── service.go           # AI服务核心
├── template.go          # 提示词模板
└── history.go           # 对话历史管理
```

### 支持的提供商
- DeepSeek
- Google Gemini
- Groq
- OpenRouter
- Ollama (本地)

### 模板系统
提供预定义的提示词模板：
- 内容生成模板
- 内容改写模板
- 热点分析模板
- 标题优化模板

## 任务管理系统设计

### 核心组件
```
task/
├── manager.go           # 任务管理器
├── tracker.go           # 状态追踪器
├── handlers/            # 任务处理器
│   └── publish.go
└── storage.go           # 任务存储
```

### 状态机设计
```
Created → Pending → Running → Completed/Failed
              ↘ Cancelled
```

### 并发控制
- 最大并发任务数：10
- 任务队列长度：100
- 超时设置：30分钟

## 浏览器自动化设计

### 技术选型
基于 [go-rod/rod](https://github.com/go-rod/rod) 框架，利用Chrome DevTools Protocol。

### 核心功能
```
browser/
├── browser.go           # 浏览器实例管理
├── navigation.go        # 页面导航
├── interaction.go       # 元素交互
└── screenshot.go        # 截图功能
```

### 实例池管理
- 最大浏览器实例：5个
- 实例复用机制
- 自动回收策略

## 数据分析模块设计

### 数据采集架构
```
analytics/
├── collectors/          # 数据采集器
│   ├── douyin.go
│   ├── toutiao.go
│   └── xiaohongshu.go
├── storage.go           # 数据存储
├── report.go            # 报告生成
└── visualization.go     # 数据可视化
```

### 报告类型
- 日报
- 周报
- 月报
- 自定义报告

### 输出格式
- JSON格式（程序化处理）
- Markdown格式（人工阅读）
- CSV格式（表格分析）

## 热点监控设计

### 数据源架构
```
hotspot/
├── sources/             # 热点数据源
│   ├── newsnow.go
│   ├── baidu.go
│   ├── weibo.go
│   └── zhihu.go
├── service.go           # 热点服务
├── storage.go           # 数据存储
└── api.go               # API接口
```

### 数据处理流程
1. 定时抓取各数据源
2. 数据清洗和去重
3. 热度计算和排序
4. AI内容适配建议
5. 存储和API暴露

## Cookie管理设计

### 存储策略
```
cookies/
├── cookies.go           # Cookie管理器
├── storage.go           # 存储实现
└── encryption.go        # 加密保护
```

### 功能特性
- 自动加载和保存
- 多平台隔离存储
- 过期检测和刷新
- 加密存储保护

### 存储格式
JSON文件格式，每个平台独立文件：
```
cookies/
├── douyin_cookies.json
├── toutiao_cookies.json
└── xiaohongshu_cookies.json
```

## 文件存储设计

### 抽象层设计
```go
type Storage interface {
    Write(ctx context.Context, path string, data []byte) error
    Read(ctx context.Context, path string) ([]byte, error)
    Delete(ctx context.Context, path string) error
    GetURL(ctx context.Context, path string) (string, error)
    List(ctx context.Context, prefix string) ([]string, error)
}
```

### 实现类型
- 本地文件存储
- 内存存储（测试用）
- 云存储适配器（扩展）

### 目录结构
```
uploads/
├── images/              # 图片文件
├── videos/              # 视频文件
├── audio/               # 音频文件
└── documents/           # 文档文件
```

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0