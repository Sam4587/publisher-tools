# 架构文档

## 系统架构概述

Publisher Tools 采用微服务架构设计，分为前端、后端核心服务和辅助服务三个主要部分。

## 技术栈

### 前端 (publisher-web)
- **框架**: React 18 + TypeScript
- **构建工具**: Vite 5
- **UI库**: shadcn/ui + Tailwind CSS
- **状态管理**: React Hooks
- **路由**: React Router DOM
- **HTTP客户端**: Axios

### 后端核心 (publisher-core)
- **语言**: Go 1.21+
- **Web框架**: Gorilla Mux
- **浏览器自动化**: Rod (基于Chrome DevTools Protocol)
- **日志**: Logrus
- **错误处理**: pkg/errors

### 辅助服务
- **测试API服务器**: Node.js + Express
- **AI服务**: 支持多种提供商（DeepSeek、Google、Groq等）

## 目录结构

```
publisher-tools/
├── publisher-web/          # React前端应用
│   ├── src/
│   │   ├── components/     # UI组件
│   │   ├── pages/          # 页面组件
│   │   ├── lib/            # 工具库
│   │   └── types/          # TypeScript类型定义
│   └── package.json
├── publisher-core/         # Go后端核心
│   ├── adapters/           # 平台适配器
│   ├── ai/                 # AI服务集成
│   ├── analytics/          # 数据分析模块
│   ├── api/                # REST API服务
│   ├── browser/            # 浏览器自动化
│   ├── cmd/                # 命令行入口
│   ├── cookies/            # Cookie管理
│   ├── hotspot/            # 热点监控
│   ├── interfaces/         # 接口定义
│   ├── storage/            # 文件存储
│   └── task/               # 任务管理
├── server/                 # Node.js测试服务器
├── docs/                   # 文档目录
│   ├── architecture/       # 架构文档
│   ├── api/               # API文档
│   ├── deployment/        # 部署文档
│   └── development/       # 开发文档
└── test-api-server.js     # 测试API服务器
```

## 核心模块说明

### 1. 平台适配器 (Adapters)
负责对接不同平台的发布接口，实现统一的发布接口。

### 2. AI服务 (AI Service)
集成多种AI提供商，提供内容生成、改写、分析等功能。

### 3. 任务管理 (Task Manager)
异步任务处理系统，管理内容发布等耗时操作。

### 4. 浏览器自动化 (Browser Automation)
基于Rod框架实现网页操作自动化。

### 5. 数据分析 (Analytics)
收集和分析各平台发布效果数据。

## 数据流向

```
用户输入 → 前端UI → API服务 → 任务管理 → 平台适配器 → 目标平台
     ↑                                                   ↓
   展示结果 ← 数据分析 ← 浏览器自动化 ← Cookie管理 ← 登录状态
```

## 部署架构

### 开发环境
- 前端: Vite开发服务器 (localhost:5173+)
- 后端: 测试API服务器 (localhost:3001) 或 Go服务 (localhost:8080)

### 生产环境
- 前端: 构建后静态文件部署
- 后端: Go服务 (Docker容器化部署)
- 数据存储: 本地文件系统 + Cookie文件