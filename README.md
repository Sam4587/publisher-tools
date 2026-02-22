# Publisher Tools

多平台内容发布和热点监控系统

## 快速开始

### 一键启动

```bash
# Windows
start-all.bat
```

启动后自动打开浏览器: http://localhost:5173

**特性:**
- ✅ 后台隐藏运行,无干扰
- ✅ 自动验证服务状态
- ✅ 完整日志记录
- ✅ 自动打开浏览器

### 停止服务

```bash
stop.bat
```

### 健康检查

```bash
health-check.bat
```

### 手动启动(开发调试)

```bash
# 1. 启动热点服务器
cd hotspot-server && go run main.go

# 2. 启动后端服务
cd server && node simple-server.js

# 3. 启动前端
cd publisher-web && npm run dev
```

## 系统组件

| 服务 | 端口 | 说明 |
|------|------|------|
| Hotspot Server | 8080 | Go语言热点数据服务 |
| Node Backend | 3001 | Node.js业务后端 |
| Frontend | 5173 | React前端界面 |

## 常用命令

```bash
# 健康检查
health-check.bat

# 停止所有服务
stop.bat

# 查看日志
type logs\hotspot-server.log
type logs\node-backend.log
type logs\frontend.log
```

## 功能特性

### 热点监控
- 实时抓取多平台热点(微博、抖音、今日头条、知乎、B站)
- 热度趋势分析
- AI智能分析
- 每30分钟自动更新

### 内容发布
- 多平台支持(抖音、小红书、今日头条、B站)
- 扫码登录
- 内容管理
- 发布任务队列

## 故障排查

### 热点数据无法加载

1. 运行健康检查: `health-check.bat`
2. 检查Go服务: `curl http://localhost:8080/api/health`
3. 查看日志: `type logs\hotspot-server.log`
4. 重启服务: `stop.bat && start-stable.bat`

### 前端无法访问

1. 确认前端服务运行: `netstat -ano | findstr ":5173"`
2. 检查浏览器控制台错误
3. 清除浏览器缓存

## 开发指南

### 环境要求

- Node.js >= 16
- Go >= 1.19
- npm >= 8

### 项目结构

```
publisher-tools/
├── hotspot-server/      # Go热点服务
├── server/              # Node后端
├── publisher-web/       # React前端
├── service-manager.js   # 服务管理器
├── start-stable.bat     # 启动脚本
├── stop.bat             # 停止脚本
├── health-check.bat     # 健康检查
└── OPERATIONS.md        # 运维文档
```

### API文档

#### 热点API

```bash
# 获取热点列表
GET http://localhost:8080/api/hot-topics?limit=50

# 获取数据源
GET http://localhost:8080/api/hot-topics/newsnow/sources

# 抓取最新热点
POST http://localhost:8080/api/hot-topics/newsnow/fetch

# 获取新增热点
GET http://localhost:8080/api/hot-topics/trends/new?hours=24
```

#### 平台API

```bash
# 获取平台列表
GET http://localhost:3001/api/platforms

# 平台登录
POST http://localhost:3001/api/v1/publisher/platforms/:platform/login

# 检查登录状态
GET http://localhost:3001/api/v1/publisher/platforms/:platform/check
```

## 更多信息

详细运维文档请查看: [OPERATIONS.md](./OPERATIONS.md)

## 许可证

MIT License
