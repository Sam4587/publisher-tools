# Publisher Tools 运维文档

## 系统架构

Publisher Tools 是一个多服务架构系统,包含以下组件:

### 服务列表

1. **Hotspot Server (Go)** - 端口 8080
   - 功能: 热点数据抓取和管理
   - 技术: Go + Gorilla Mux
   - 健康检查: `http://localhost:8080/api/health`
   - 主要API:
     - `GET /api/hot-topics` - 获取热点列表
     - `GET /api/hot-topics/newsnow/sources` - 获取数据源
     - `POST /api/hot-topics/newsnow/fetch` - 抓取热点
     - `GET /api/hot-topics/trends/new` - 获取新增热点

2. **Node Backend** - 端口 3001
   - 功能: 业务逻辑和平台发布
   - 技术: Node.js + Express
   - 健康检查: `http://localhost:3001/api/health`
   - 主要API:
     - `GET /api/platforms` - 获取平台列表
     - `POST /api/v1/publisher/platforms/:platform/login` - 平台登录
     - `GET /api/v1/publisher/platforms/:platform/check` - 检查登录状态

3. **Frontend (React)** - 端口 5173
   - 功能: Web界面
   - 技术: React + Vite + TypeScript
   - 访问地址: `http://localhost:5173`
   - API代理: 所有 `/api` 请求代理到 `localhost:8080`

## 快速启动

### 一键启动(推荐)

```bash
# Windows
start-all.bat
```

**启动脚本特性:**
- ✅ 自动清理旧进程
- ✅ 按顺序启动所有服务
- ✅ 后台隐藏运行
- ✅ 自动验证服务状态
- ✅ 完整日志记录
- ✅ 自动打开浏览器

### 手动启动各服务(开发调试)

```bash
# 1. 启动 Go 热点服务器
cd hotspot-server
go run main.go

# 2. 启动 Node 后端
cd server
node simple-server.js

# 3. 启动前端
cd publisher-web
npm run dev
```

## 服务管理

### 查看服务状态

```bash
# 检查端口占用
netstat -ano | findstr "8080"  # Hotspot Server
netstat -ano | findstr "3001"  # Node Backend
netstat -ano | findstr "5173"  # Frontend

# 查看服务日志
type logs\hotspot-server.log
type logs\node-backend.log
type logs\frontend.log
type logs\service-manager.log
```

### 健康检查

```bash
# Hotspot Server
curl http://localhost:8080/api/health

# Node Backend
curl http://localhost:3001/api/health

# 前端(通过浏览器访问)
http://localhost:5173
```

### 停止服务

```bash
# 停止所有服务(如果有 stop.bat)
stop.bat

# 或手动停止
taskkill //F //PID <进程ID>

# 或通过端口停止
netstat -ano | findstr ":8080"
taskkill //F //PID <找到的PID>
```

## 常见问题排查

### 问题1: 热点数据无法加载

**症状:** 前端显示"没有数据"或API返回500错误

**排查步骤:**
1. 检查Go服务是否运行
   ```bash
   curl http://localhost:8080/api/health
   ```
2. 检查端口是否被占用
   ```bash
   netstat -ano | findstr ":8080"
   ```
3. 查看服务日志
   ```bash
   type logs\hotspot-server.log
   ```
4. 重启服务
   ```bash
   # 停止旧进程
   taskkill //F //PID <PID>
   # 重新启动
   cd hotspot-server
   go run main.go
   ```

### 问题2: 前端无法访问API

**症状:** 浏览器控制台显示CORS错误或网络错误

**排查步骤:**
1. 检查Vite代理配置(`vite.config.ts`)
   ```typescript
   proxy: {
     '/api': {
       target: 'http://localhost:8080',
       changeOrigin: true,
     }
   }
   ```
2. 确认Go服务运行在8080端口
3. 检查防火墙设置

### 问题3: 服务频繁重启

**症状:** 服务管理器日志显示服务不断重启

**排查步骤:**
1. 查看服务日志找出崩溃原因
   ```bash
   type logs\hotspot-server.log
   ```
2. 检查端口冲突
3. 检查依赖是否安装完整
   ```bash
   # Go依赖
   cd hotspot-server
   go mod download
   
   # Node依赖
   cd publisher-web
   npm install
   ```

## 日志管理

### 日志位置

所有日志存储在 `logs/` 目录:
- `hotspot-server.log` - Go服务日志
- `node-backend.log` - Node后端日志
- `frontend.log` - 前端开发服务器日志
- `service-manager.log` - 服务管理器日志

### 日志轮转

建议定期清理日志文件:
```bash
# Windows
del logs\*.log

# 或保留最近7天的日志
forfiles /p logs /m *.log /d -7 /c "cmd /c del @path"
```

## 性能优化

### 1. Go服务优化

- 调整热点抓取频率(默认30分钟)
- 限制内存中的热点数量
- 使用数据库持久化存储

### 2. 前端优化

- 启用Vite的构建优化
- 使用React.lazy进行代码分割
- 配置CDN加速静态资源

### 3. 服务管理器优化

- 调整健康检查频率(默认30秒)
- 配置重启延迟时间(默认5秒)
- 设置最大重启次数限制

## 安全建议

1. **生产环境配置**
   - 修改默认端口
   - 启用HTTPS
   - 配置防火墙规则
   - 使用环境变量管理敏感信息

2. **API安全**
   - 添加认证中间件
   - 实施请求频率限制
   - 验证所有输入参数

3. **日志安全**
   - 不记录敏感信息
   - 定期归档和清理日志
   - 限制日志文件访问权限

## 备份和恢复

### 备份策略

1. **配置文件备份**
   - `vite.config.ts`
   - `service-manager.js`
   - `start-stable.bat`

2. **数据备份**
   - 热点数据(如使用数据库)
   - 用户会话和Cookie

### 恢复步骤

1. 停止所有服务
2. 恢复配置文件
3. 恢复数据
4. 重启服务

## 监控和告警

### 推荐监控指标

1. **服务可用性**
   - HTTP状态码
   - 响应时间
   - 服务运行时间

2. **资源使用**
   - CPU使用率
   - 内存使用量
   - 磁盘空间

3. **业务指标**
   - 热点抓取成功率
   - API请求量
   - 错误率

### 告警配置

建议配置以下告警:
- 服务不可用(连续3次健康检查失败)
- 高CPU/内存使用(>80%)
- API错误率过高(>5%)
- 磁盘空间不足(<10GB)

## 更新和维护

### 更新步骤

1. 备份当前版本
2. 拉取最新代码
3. 更新依赖
   ```bash
   cd hotspot-server && go mod download
   cd publisher-web && npm install
   ```
4. 重启服务

### 定期维护任务

- 每周: 检查日志,清理临时文件
- 每月: 更新依赖包,检查安全漏洞
- 每季度: 性能评估和优化

## 联系和支持

如遇到问题,请:
1. 查看本文档的常见问题部分
2. 检查服务日志
3. 提交Issue到项目仓库

---

**最后更新:** 2026-02-22
**版本:** 1.0.0
