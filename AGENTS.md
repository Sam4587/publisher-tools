# AI 开发者指南

本文档为 AI 开发者提供快速上手指南，帮助你尽快进入开发状态。

## 项目状态

- **分支**: publisher-tools
- **来源**: 从 TrendRadar 项目分离
- **分离时间**: 2025-02-16
- **当前状态**: 可编译运行，部分功能待完善

## 快速开始

### 1. 编译项目

```bash
make build
# 或
./dev.sh build
```

### 2. 启动开发环境

```bash
# 方式一: 使用 Makefile
make dev

# 方式二: 使用启动脚本
./dev.sh start

# 方式三: 分别启动
make serve       # 后端 (端口 8080)
make serve-web   # 前端 (端口 5173)
```

### 3. 停止服务

```bash
make stop
# 或
./dev.sh stop
```

## 项目结构速览

```
publisher-core/           # Go 后端核心
├── interfaces/           # 接口定义 (先看这里)
├── adapters/             # 平台适配器
├── api/                  # REST API
├── task/                 # 任务管理
├── browser/              # 浏览器自动化
├── storage/              # 文件存储
├── cookies/              # Cookie 管理
└── cmd/server/           # 服务入口

publisher-web/            # React 前端
├── src/pages/            # 页面组件
├── src/components/       # UI 组件
├── src/lib/api.ts        # API 调用
└── src/types/api.ts      # 类型定义
```

## 关键文件

| 文件 | 用途 |
|------|------|
| `publisher-core/interfaces/publisher.go` | Publisher 接口定义 |
| `publisher-core/adapters/adapters.go` | 平台适配器实现 |
| `publisher-core/api/server.go` | REST API 服务 |
| `publisher-core/cmd/server/main.go` | 服务入口 |
| `publisher-web/src/lib/api.ts` | 前端 API 调用 |
| `publisher-web/src/types/api.ts` | TypeScript 类型 |

## API 端点

| 端点 | 状态 | 说明 |
|------|------|------|
| `GET /api/v1/platforms` | 完成 | 平台列表 |
| `GET /api/v1/platforms/{platform}/check` | 完成 | 登录状态 |
| `POST /api/v1/platforms/{platform}/login` | 完成 | 登录 |
| `GET /api/v1/tasks` | 完成 | 任务列表 |
| `POST /api/v1/publish/async` | 待完善 | 异步发布 |

## 常用开发命令

```bash
# 查看 Makefile 帮助
make help

# 运行测试
make test

# 安装依赖
make deps

# 查看服务状态
make status
./dev.sh status

# 查看日志
make logs
./dev.sh logs
```

## 待开发功能

1. **发布功能完善**
   - `/api/v1/publish` 同步发布
   - `/api/v1/publish/async` 异步发布执行
   - 任务处理器注册

2. **文件上传**
   - `/api/v1/storage/upload`
   - 图片/视频处理

3. **热点监控**
   - 热点数据抓取
   - 趋势分析

## 开发注意事项

1. **浏览器自动化**: 需要 Chrome/Chromium 环境
2. **Cookie 存储**: `./cookies/` 目录
3. **文件上传**: `./uploads/` 目录
4. **日志文件**: `./logs/` 目录
5. **PID 文件**: `./pids/` 目录

## 代码规范

### Go 代码

- 使用 `logrus` 进行日志记录
- 错误使用 `github.com/pkg/errors` 包装
- 接口定义在 `interfaces/` 包

### 前端代码

- 使用 TypeScript
- 组件使用函数式组件 + Hooks
- UI 组件使用 shadcn/ui
- API 调用封装在 `lib/api.ts`

## 测试

```bash
# 后端测试
cd publisher-core && go test ./... -v

# 前端测试 (如有)
cd publisher-web && npm test
```

## 文档位置

- 开发指南: `.monkeycode/docs/DEVELOPMENT.md`
- 架构文档: `.monkeycode/docs/ARCHITECTURE.md`
- 项目说明: `README.md`
