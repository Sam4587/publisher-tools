# Publisher Tools 开发指南

## 项目概述

Publisher Tools 是一个多平台内容发布自动化系统，支持抖音、今日头条、小红书等平台的内容发布。

### 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21+, go-rod (浏览器自动化), gorilla/mux |
| 前端 | React 18, TypeScript, Vite, shadcn/ui, Tailwind CSS |
| 自动化 | Chrome/Chromium 浏览器 |

### 项目结构

```
/workspace/
├── publisher-core/          # 核心库 (Go) - 统一架构
│   ├── interfaces/          # Publisher 接口定义
│   ├── adapters/            # 平台适配器 (抖音/头条/小红书)
│   ├── api/                 # REST API 服务
│   ├── task/                # 异步任务管理
│   ├── browser/             # 浏览器自动化封装
│   ├── storage/             # 文件存储抽象
│   ├── cookies/             # Cookie 管理
│   └── cmd/
│       ├── server/          # API 服务入口
│       └── cli/             # 命令行工具入口
│
├── publisher-web/           # Web 管理界面 (React)
│   └── src/
│       ├── pages/           # 页面组件
│       ├── components/      # UI 组件
│       ├── lib/             # API 工具函数
│       └── types/           # TypeScript 类型定义
│
├── douyin-toutiao/          # 独立 CLI 工具 (Go)
│   ├── douyin/              # 抖音登录/发布
│   ├── toutiao/             # 今日头条登录/发布
│   └── cookies/             # Cookie 管理
│
├── xiaohongshu-publisher/   # 小红书独立 CLI 工具 (Go)
│
├── bin/                     # 编译输出目录
├── cookies/                 # Cookie 存储目录 (运行时生成)
└── uploads/                 # 文件上传目录 (运行时生成)
```

---

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Chrome/Chromium 浏览器

### 一键启动

```bash
# 编译所有服务
make build

# 启动后端服务 (端口 8080)
make serve

# 启动前端服务 (端口 5173) - 新终端
make serve-web

# 或同时启动前后端
make dev
```

### 手动启动

```bash
# 后端
./bin/publisher-server -port 8080

# 前端
cd publisher-web && npm run dev
```

---

## API 文档

### 基础信息

- 基础路径: `/api/v1`
- 响应格式: JSON
- 默认端口: 8080

### 通用响应格式

```json
{
  "success": true,
  "data": {},
  "error": "",
  "error_code": "",
  "timestamp": 1234567890
}
```

### 端点列表

| 端点 | 方法 | 状态 | 说明 |
|------|------|------|------|
| `/health` | GET | 完成 | 健康检查 |
| `/api/v1/platforms` | GET | 完成 | 获取平台列表 |
| `/api/v1/platforms/{platform}` | GET | 完成 | 获取平台信息 |
| `/api/v1/platforms/{platform}/login` | POST | 完成 | 登录平台 |
| `/api/v1/platforms/{platform}/check` | GET | 完成 | 检查登录状态 |
| `/api/v1/tasks` | GET | 完成 | 任务列表 |
| `/api/v1/tasks` | POST | 完成 | 创建任务 |
| `/api/v1/tasks/{taskId}` | GET | 完成 | 任务详情 |
| `/api/v1/tasks/{taskId}/cancel` | POST | 完成 | 取消任务 |
| `/api/v1/publish` | POST | 待完善 | 同步发布 |
| `/api/v1/publish/async` | POST | 待完善 | 异步发布 |
| `/api/v1/storage/*` | * | 待实现 | 文件上传下载 |

### 使用示例

```bash
# 健康检查
curl http://localhost:8080/health

# 获取平台列表
curl http://localhost:8080/api/v1/platforms

# 检查登录状态
curl http://localhost:8080/api/v1/platforms/douyin/check

# 登录平台
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login

# 异步发布
curl -X POST http://localhost:8080/api/v1/publish/async \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "douyin",
    "type": "images",
    "title": "标题",
    "body": "正文",
    "images": ["img1.jpg", "img2.jpg"],
    "tags": ["标签1", "标签2"]
  }'

# 查询任务状态
curl http://localhost:8080/api/v1/tasks/{taskId}
```

---

## 核心模块说明

### 1. Publisher 接口 (`publisher-core/interfaces/publisher.go`)

所有平台发布器必须实现此接口：

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

### 2. 平台适配器 (`publisher-core/adapters/adapters.go`)

- `DouyinAdapter` - 抖音发布器
- `ToutiaoAdapter` - 今日头条发布器
- `XiaohongshuAdapter` - 小红书发布器

使用工厂模式创建：

```go
factory := adapters.DefaultFactory()
pub, _ := factory.Create("douyin")
```

### 3. 任务管理 (`publisher-core/task/manager.go`)

```go
taskMgr := task.NewTaskManager(task.NewMemoryStorage())
task, _ := taskMgr.CreateTask("publish", "douyin", payload)
taskMgr.Execute(ctx, task.ID)
```

### 4. Cookie 管理 (`publisher-core/cookies/cookies.go`)

```go
cookieMgr := cookies.NewManager("./cookies")
cookieMgr.Save(ctx, "douyin", cookies)
cookieMgr.Load(ctx, "douyin")
cookieMgr.Exists(ctx, "douyin")
```

### 5. 文件存储 (`publisher-core/storage/storage.go`)

```go
store, _ := storage.NewLocalStorage("./uploads", "http://localhost:8080")
store.Write(ctx, "images/photo.jpg", data)
store.Read(ctx, "images/photo.jpg")
store.GetURL(ctx, "images/photo.jpg")
```

---

## 前端开发

### 目录结构

```
publisher-web/src/
├── pages/              # 页面组件
│   ├── Dashboard.tsx   # 仪表盘
│   ├── Accounts.tsx    # 账号管理
│   ├── Publish.tsx     # 内容发布
│   ├── History.tsx     # 发布历史
│   ├── HotTopics.tsx   # 热点监控
│   └── VideoTranscription.tsx  # 视频转录
├── components/         # 组件
│   ├── ui/            # shadcn/ui 组件
│   └── *.tsx          # 业务组件
├── lib/
│   ├── api.ts         # API 调用封装
│   └── utils.ts       # 工具函数
└── types/
    └── api.ts         # TypeScript 类型定义
```

### 开发命令

```bash
cd publisher-web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build

# 代码检查
npm run lint
```

### API 调用

```typescript
import { getPlatforms, checkLogin, publishAsync } from '@/lib/api'

// 获取平台列表
const { data } = await getPlatforms()

// 检查登录状态
const { data } = await checkLogin('douyin')

// 异步发布
const { data } = await publishAsync({
  platform: 'douyin',
  type: 'images',
  title: '标题',
  body: '正文',
  images: ['img.jpg'],
  tags: ['标签']
})
```

---

## 平台限制

| 平台 | 标题 | 正文 | 图片 | 视频 |
|------|------|------|------|------|
| 抖音 | 30字 | 2000字 | 12张 | 4GB |
| 今日头条 | 30字 | 2000字 | 多张 | 无限制 |
| 小红书 | 20字 | 1000字 | 18张 | 500MB |

---

## 待开发功能

### 高优先级

1. **完善发布 API**
   - 实现同步发布逻辑
   - 完善异步发布任务执行
   - 添加发布进度追踪

2. **文件上传功能**
   - 实现 `/api/v1/storage/upload`
   - 图片/视频文件处理
   - 文件类型验证

3. **任务执行器**
   - 注册 publish 任务处理器
   - 任务重试机制
   - 任务超时处理

### 中优先级

1. **热点监控 API**
   - 热点数据抓取
   - 热点分析
   - 趋势追踪

2. **视频转录功能**
   - 视频下载
   - 语音转文字
   - 字幕生成

### 低优先级

1. **扩展平台**
   - B站视频发布
   - 微博图文发布
   - 微信公众号发布

2. **高级功能**
   - 定时发布
   - 批量发布
   - 内容审核

---

## 注意事项

1. **首次使用**: 必须先执行登录操作
2. **Cookie 过期**: 需要定期重新登录
3. **发布间隔**: 建议间隔 >= 5 分钟
4. **内容规范**: 遵守各平台社区规范
5. **风控风险**: 高频操作可能触发限流

---

## 常见问题

### Q: 后端启动失败？

检查端口是否被占用：
```bash
lsof -i :8080
```

### Q: 前端代理不工作？

确认 `vite.config.ts` 中的代理配置：
```typescript
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
  }
}
```

### Q: 浏览器自动化不工作？

确保已安装 Chrome/Chromium：
```bash
# Linux
which chromium-browser || which google-chrome

# macOS
ls /Applications/Google\ Chrome.app
```

### Q: Cookie 保存失败？

检查目录权限：
```bash
mkdir -p ./cookies
chmod 755 ./cookies
```
