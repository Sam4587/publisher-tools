# 项目分离说明

## 分离时间
- 日期：2025-02-16
- 分支：publisher-tools

## 项目背景

本项目从 TrendRadar（热点监控与内容运营系统）中分离出来，成为独立的多平台发布工具系统。

## 分离原因

### 1. 独立的业务域
- **TrendRadar**：专注于热点监控、内容生成、数据分析
- **Publisher Tools**：专注于多平台内容发布自动化

### 2. 独立的技术栈
- **TrendRadar**：Node.js + React + MongoDB
- **Publisher Tools**：Go + React (独立前端)

### 3. 独立的部署需求
- 发布工具需要独立部署、独立扩展
- 便于后续微服务化架构演进

## 项目结构

```
.
├── publisher-core/        # 核心库（统一架构）
│   ├── interfaces/       # 接口定义
│   ├── adapters/         # 平台适配器
│   ├── browser/          # 浏览器自动化
│   ├── cookies/          # Cookie 管理
│   ├── task/             # 异步任务管理
│   ├── storage/          # 文件存储抽象
│   ├── api/              # REST API
│   └── cmd/
│       ├── server/       # API 服务入口
│       └── cli/          # 命令行入口
├── publisher-web/        # Web 管理界面
│   └── src/
│       ├── pages/        # 页面组件
│       ├── components/   # UI 组件
│       └── lib/          # API 工具函数
├── douyin-toutiao/       # 抖音/头条 CLI 工具
└── xiaohongshu-publisher/ # 小红书 CLI 工具
```

## 与 TrendRadar 的关系

### 分离前
```
TrendRadar/
├── src/              # 前端（包含发布管理页面）
├── server/           # 后端（包含发布集成服务）
└── tools/            # 发布工具（本项目的来源）
```

### 分离后
```
# TrendRadar (master 分支)
TrendRadar/
├── src/              # 前端（热点监控、内容生成等）
├── server/           # 后端（热点数据、AI服务等）
└── PROJECT_SEPARATION.md

# Publisher Tools (publisher-tools 分支)
Publisher-Tools/
├── publisher-core/   # Go 后端核心
├── publisher-web/    # React 前端
├── douyin-toutiao/   # CLI 工具
└── xiaohongshu-publisher/
```

## 后续开发方向

### 短期目标

1. **完善 publisher-core**
   - [ ] 完善平台适配器（抖音、头条、小红书、B站）
   - [ ] 实现异步任务队列
   - [ ] 添加重试机制和错误处理

2. **完善 publisher-web**
   - [ ] 仪表盘页面
   - [ ] 账号管理页面
   - [ ] 内容发布页面
   - [ ] 任务管理页面

3. **测试覆盖**
   - [ ] 单元测试
   - [ ] 集成测试
   - [ ] E2E 测试

### 中期目标

1. **扩展平台支持**
   - [ ] B站视频发布
   - [ ] 微博图文发布
   - [ ] 微信公众号发布

2. **高级功能**
   - [ ] 定时发布
   - [ ] 批量发布
   - [ ] 内容审核
   - [ ] 数据统计

### 长期目标

1. **微服务架构**
   - [ ] 拆分为独立服务
   - [ ] 消息队列集成
   - [ ] 分布式任务调度

2. **AI 增强**
   - [ ] 智能内容适配
   - [ ] 自动标签生成
   - [ ] 发布时间优化

3. **与 TrendRadar 集成**
   - [ ] 提供 REST API 供 TrendRadar 调用
   - [ ] Webhook 支持
   - [ ] 联合工作流

## API 文档

### REST API 端点

```
GET  /api/v1/platforms           # 获取支持的平台列表
GET  /api/v1/accounts            # 获取账号列表
GET  /api/v1/accounts/:platform  # 获取指定平台的登录状态

POST /api/v1/login/:platform     # 登录指定平台
POST /api/v1/publish             # 同步发布
POST /api/v1/publish/async       # 异步发布
GET  /api/v1/tasks               # 获取任务列表
GET  /api/v1/tasks/:taskId       # 查询任务状态
```

## 部署说明

### 开发环境

```bash
# 编译
make build

# 启动后端服务
./bin/publisher-server -port 8080

# 启动前端开发服务器
cd publisher-web && npm run dev
```

### 生产环境

```bash
# 构建
make build

# 使用 PM2 管理后端服务
pm2 start ./bin/publisher-server --name publisher-api

# 构建前端
cd publisher-web && npm run build

# 使用 Nginx 托管前端静态文件
```

## 环境要求

- Go 1.21+
- Node.js 18+
- Chrome/Chromium 浏览器（用于浏览器自动化）

## 给 AI 开发者的说明

### 项目状态
- 本项目已完成从 TrendRadar 的分离
- 分离时间：2025-02-16
- 原始代码位于 `tools/` 目录，现已移至根目录

### 技术栈
- **后端**：Go 1.21+，使用 go-rod 进行浏览器自动化
- **前端**：React 18 + Vite + shadcn/ui + Tailwind CSS

### 开发注意事项
1. 浏览器自动化需要 Chrome/Chromium 环境
2. Cookie 管理在 `./cookies/` 目录
3. 异步任务在 `./task/` 模块管理
4. API 服务入口在 `publisher-core/cmd/server/`

### 相关分支
| 分支 | 说明 |
|------|------|
| `master` | TrendRadar 主项目 |
| `publisher-tools` | 本项目（发布工具） |

## 联系方式

如有问题，请在项目 Issue 中反馈。
