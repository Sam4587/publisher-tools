# Publisher Tools - 多平台内容发布工具集

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**一站式多平台内容发布自动化系统**

[功能特性](#功能特性) • [快速开始](#快速开始) • [部署](#部署) • [API文档](#api文档) • [开发指南](#开发指南)

</div>

---

## 📖 项目简介

Publisher Tools 是一个支持多平台内容发布的自动化系统，从 TrendRadar 项目分离而来，提供完整的**内容创作→发布→数据分析**闭环能力。

### 当前版本状态

- **分支**: `publisher-tools`
- **状态**: ✅ 可编译运行，核心功能已实现
- **最近更新**: 2026-02-17 完成内容生成页面一键发布功能

### 核心能力

- 🚀 **多平台发布** - 支持抖音、今日头条、小红书三大平台
- 🤖 **AI驱动** - 集成多种AI提供商，智能内容生成
- 📊 **数据分析** - 完整的数据采集和报告生成
- 🎬 **视频转录** - 视频内容AI转录，一键改写发布
- 🔥 **热点监控** - 实时热点抓取，趋势分析
- 💻 **Web管理** - 现代化React前端界面
- 🔄 **一键发布** - 内容生成后直接发布到目标平台

---

## ✨ 功能特性

### 1. 平台发布
- ✅ 统一发布接口
- ✅ 同步/异步发布模式
- ✅ 任务队列管理
- ✅ 发布状态追踪
- ✅ Cookie自动管理

### 2. AI服务集成
- ✅ 多提供商支持（OpenRouter、DeepSeek、Ollama等）
- ✅ 内容生成与改写
- ✅ 内容审核
- ✅ 热点分析

### 3. 视频转录
- ✅ 多平台视频下载
- ✅ AI语音转录
- ✅ 关键词提取
- ✅ 内容改写优化
- ✅ 一键发布

### 4. 热点监控
- ✅ NewsNow聚合源
- ✅ 多数据源支持
- ✅ 热度趋势分析
- ✅ AI内容适配

### 5. 数据分析
- ✅ 三平台数据采集框架
- ✅ 智能报告生成
- ✅ JSON/Markdown导出
- ✅ 数据洞察分析

### 6. Web管理界面
- ✅ 仪表盘总览
- ✅ 账号管理
- ✅ 内容发布
- ✅ 任务历史
- ✅ 数据分析

---

## 🚀 快速开始

### 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **Chrome/Chromium** - 浏览器自动化

### 方式一：直接运行（推荐开发环境）

```bash
# 1. 克隆项目
git clone <repository-url>
cd publisher-tools

# 2. 编译项目
make build

# 3. 启动开发环境
make dev
```

### 服务端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端开发服务器 | 5173+ | Vite自动寻找可用端口 |
| 测试API服务器 | 3001 | 用于前端开发测试 |
| Go后端服务 | 8080 | 生产环境后端服务 |

访问：
- **前端**: http://localhost:5173（或Vite自动分配的端口）
- **测试API**: http://localhost:3001/api/health
- **后端API**: http://localhost:8080

### 方式二：Docker部署（推荐生产环境）

```bash
# 使用 Docker Compose
docker-compose up -d

# 查看日志
docker-compose logs -f
```

访问：http://localhost:8080

### 方式三：手动启动

```bash
# 启动测试API服务器（用于开发测试）
node test-api-server.js

# 或启动Go后端服务
./bin/publisher-server -port 8080

# 前端开发服务器
cd publisher-web
npm install
npm run dev
```

> **注意**: Vite会自动寻找可用端口，实际端口请查看终端输出

---

## 📦 部署

### Docker部署

```bash
# 构建镜像
docker build -t publisher-tools .

# 运行容器
docker run -d \
  --name publisher-server \
  -p 8080:8080 \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/cookies:/app/cookies \
  -v $(pwd)/data:/app/data \
  publisher-tools
```

### Docker Compose部署

```bash
# 启动所有服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart
```

### 生产环境配置

创建 `.env` 文件：

```bash
# 服务配置
PORT=8080
HEADLESS=true
DEBUG=false

# AI配置
OPENROUTER_API_KEY=your_key
DEEPSEEK_API_KEY=your_key

# 存储配置
STORAGE_DIR=/app/uploads
DATA_DIR=/app/data
```

---

## 📚 API文档

### 核心端点

#### 平台管理
```
GET    /api/v1/platforms                    # 平台列表
GET    /api/v1/platforms/{platform}/check   # 登录状态
POST   /api/v1/platforms/{platform}/login   # 登录平台
POST   /api/v1/platforms/{platform}/logout  # 登出平台
```

#### 内容发布
```
POST   /api/v1/publish                       # 同步发布
POST   /api/v1/publish/async                 # 异步发布
GET    /api/v1/tasks                         # 任务列表
GET    /api/v1/tasks/{taskId}                # 任务详情
POST   /api/v1/tasks/{taskId}/cancel         # 取消任务
```

#### 文件存储
```
POST   /api/v1/storage/upload                # 文件上传
GET    /api/v1/storage/download              # 文件下载
GET    /api/v1/storage/list                  # 文件列表
DELETE /api/v1/storage/delete                # 文件删除
```

#### 热点监控
```
GET    /api/hot-topics                       # 热点列表
GET    /api/hot-topics/{id}                  # 热点详情
POST   /api/hot-topics/newsnow/fetch         # 抓取热点
GET    /api/hot-topics/newsnow/sources       # 数据源列表
```

#### 数据分析
```
GET    /api/analytics/dashboard              # 仪表盘数据
GET    /api/analytics/trends                 # 趋势数据
GET    /api/analytics/report/weekly          # 周报
GET    /api/analytics/report/monthly         # 月报
GET    /api/analytics/report/export          # 导出报告
```

#### AI服务
```
GET    /api/v1/ai/providers                  # AI提供商列表
GET    /api/v1/ai/models                     # AI模型列表
POST   /api/v1/ai/generate                   # AI生成
POST   /api/v1/ai/analyze/hotspot            # 热点分析
POST   /api/v1/ai/content/generate           # 内容生成
POST   /api/v1/ai/content/rewrite            # 内容改写
```

### API使用示例

```bash
# 获取平台列表
curl http://localhost:8080/api/v1/platforms

# 登录平台
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login

# 异步发布
curl -X POST http://localhost:8080/api/v1/publish/async \
  -H "Content-Type: application/json" \
  -d '{
    "platform":"douyin",
    "type":"images",
    "title":"标题",
    "content":"正文",
    "images":["uploads/test.jpg"]
  }'

# 获取周报
curl http://localhost:8080/api/analytics/report/weekly

# 导出Markdown报告
curl "http://localhost:8080/api/analytics/report/export?format=markdown"
```

---

## 👨‍💻 开发指南

### 项目结构

```
publisher-tools/
├── publisher-core/           # Go后端核心
│   ├── adapters/            # 平台适配器
│   ├── analytics/           # 数据分析
│   │   └── collectors/     # 数据采集器
│   ├── api/                 # REST API
│   ├── ai/                  # AI服务
│   ├── browser/             # 浏览器自动化
│   ├── cookies/             # Cookie管理
│   ├── hotspot/             # 热点监控
│   ├── interfaces/          # 接口定义
│   ├── storage/             # 文件存储
│   ├── task/                # 任务管理
│   │   └── handlers/       # 任务处理器
│   └── cmd/server/          # 服务入口
│
├── publisher-web/           # React前端
│   ├── src/
│   │   ├── pages/          # 页面组件
│   │   ├── components/     # UI组件
│   │   ├── lib/            # API工具
│   │   └── types/          # 类型定义
│   └── package.json
│
├── Makefile                 # 构建脚本
├── dev.sh                   # 开发脚本
├── Dockerfile               # Docker配置
└── docker-compose.yml       # Docker编排
```

### 开发命令

```bash
# 查看 Makefile 帮助
make help

# 编译项目
make build

# 启动开发环境
make dev

# 运行测试
make test

# 查看服务状态
make status

# 查看日志
make logs

# 停止服务
make stop
```

### 代码规范

#### Go代码
- 使用 `logrus` 进行日志记录
- 错误使用 `github.com/pkg/errors` 包装
- 接口定义在 `interfaces/` 包
- 单元测试覆盖率 > 70%

#### 前端代码
- 使用 TypeScript
- 函数式组件 + Hooks
- UI组件使用 shadcn/ui
- API调用封装在 `lib/api.ts`

---

## 🔧 配置说明

### 平台限制

| 平台 | 标题 | 正文 | 图片 | 视频 |
|------|------|------|------|------|
| 抖音 | 30字 | 2000字 | 12张 | 2GB, MP4 |
| 今日头条 | 30字 | 2000字 | 9张 | MP4 |
| 小红书 | 20字 | 1000字 | 18张 | 500MB, MP4 |

### Cookie存储

- 小红书：`./cookies/xiaohongshu_cookies.json`
- 抖音：`./cookies/douyin_cookies.json`
- 今日头条：`./cookies/toutiao_cookies.json`

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| PORT | API服务端口 | 8080 |
| HEADLESS | 浏览器无头模式 | true |
| DEBUG | 调试模式 | false |
| COOKIE_DIR | Cookie存储目录 | ./cookies |
| STORAGE_DIR | 文件存储目录 | ./uploads |
| DATA_DIR | 数据存储目录 | ./data |

---

## 📈 性能优化

### 并发控制
- 任务队列最大并发：10
- API请求限流：100 req/min
- 浏览器实例池：5个实例

### 缓存策略
- Cookie缓存：内存 + 文件
- 热点数据：5分钟过期
- 报告数据：1小时缓存

---

## 🔒 安全说明

### 注意事项

1. **首次使用**：必须先执行登录操作
2. **Cookie过期**：需要定期重新登录
3. **发布间隔**：建议间隔 >=5 分钟
4. **内容规范**：遵守各平台社区规范
5. **风控风险**：高频操作可能触发限流

### 安全建议

- 不要在公网暴露服务端口
- 定期更换Cookie
- 使用环境变量管理敏感信息
- 启用HTTPS加密传输

---

## 🛠️ 故障排查

### 常见问题

**Q: 浏览器自动化失败**
```bash
# 检查 Chrome 是否安装
which chromium || which google-chrome

# 安装 Chrome
apt-get install chromium-browser  # Ubuntu/Debian
yum install chromium               # CentOS/RHEL
```

**Q: Cookie失效**
```bash
# 重新登录
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login
```

**Q: 任务执行失败**
```bash
# 查看任务状态
curl http://localhost:8080/api/v1/tasks/{taskId}

# 查看日志
make logs
```

**Q: AI服务不可用**
```bash
# 配置 AI API Key
export OPENROUTER_API_KEY=your_key
export DEEPSEEK_API_KEY=your_key

# 重启服务
make restart
```

---

## 🗺️ 路线图

### v1.1 (计划中)
- [ ] 更多平台支持（B站、微博、微信公众号）
- [ ] 定时发布功能
- [ ] 批量发布优化
- [ ] 内容审核增强

### v1.2 (规划中)
- [ ] 微服务架构
- [ ] 消息队列集成
- [ ] 分布式任务调度
- [ ] AI增强功能

---

## 🤝 贡献指南

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

## 🙏 致谢

### 开源依赖

**后端**
- [go-rod/rod](https://github.com/go-rod/rod) - 浏览器自动化
- [gorilla/mux](https://github.com/gorilla/mux) - HTTP路由
- [sirupsen/logrus](https://github.com/sirupsen/logrus) - 日志

**前端**
- [React](https://react.dev/) - UI框架
- [shadcn/ui](https://ui.shadcn.com/) - UI组件库
- [Tailwind CSS](https://tailwindcss.com/) - CSS框架
- [Vite](https://vitejs.dev/) - 构建工具

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给一个 Star ⭐**

Made with ❤️ by MonkeyCode Team

</div>
