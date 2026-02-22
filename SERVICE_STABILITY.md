# 服务稳定性改进方案

## 问题分析

### 原有问题
1. **服务启动方式不稳定**
   - 使用 `run_hidden.js` 启动服务，没有进程监控
   - 服务崩溃后不会自动重启
   - 没有健康检查机制

2. **Node.js 后端容易崩溃**
   - 缺少错误处理机制
   - 未捕获的异常会导致进程退出
   - 没有优雅关闭机制

3. **配置不一致**
   - 前端代理指向 3001 端口（Node.js）
   - 启动脚本期望 Go 后端在 8080 端口
   - 导致配置混乱

## 解决方案

### 1. 服务管理器 (`service-manager.js`)

**功能特性：**
- ✅ 自动监控服务状态
- ✅ 服务崩溃自动重启（5秒后）
- ✅ 健康检查机制（每30秒）
- ✅ 日志记录到文件
- ✅ 优雅关闭处理

**使用方法：**
```bash
# 方式1: 直接运行
node service-manager.js

# 方式2: 使用启动脚本
start-stable.bat
```

### 2. 改进的 Node.js 后端

**新增功能：**
- ✅ 捕获未处理的异常（不退出进程）
- ✅ 捕获未处理的 Promise 拒绝
- ✅ 优雅关闭（SIGTERM/SIGINT）
- ✅ 请求日志记录
- ✅ 错误处理中间件
- ✅ 404 处理
- ✅ 心跳检测

### 3. 新的启动脚本 (`start-stable.bat`)

**改进点：**
- 使用服务管理器启动服务
- 自动清理旧进程
- 显示服务状态信息
- 自动打开浏览器

## 使用指南

### 启动服务

**推荐方式：**
```bash
# 双击运行
start-stable.bat
```

**或手动启动：**
```bash
# 1. 停止所有服务
stop.bat

# 2. 启动服务管理器
node service-manager.js
```

### 查看日志

日志文件位置：
```
logs/
├── node-backend.log    # Node.js 后端日志
├── frontend.log        # 前端日志
└── service-manager.log # 服务管理器日志
```

实时查看日志：
```bash
# Windows
Get-Content logs\node-backend.log -Wait

# Linux/Mac
tail -f logs/node-backend.log
```

### 监控服务状态

服务管理器会每 30 秒自动检查服务状态，显示：
- 进程 ID (PID)
- 运行端口
- 运行时间
- 健康状态

### 停止服务

```bash
# 方式1: 使用停止脚本
stop.bat

# 方式2: 在服务管理器窗口按 Ctrl+C
```

## 服务架构

```
┌─────────────────────────────────────┐
│      Service Manager (Node.js)      │
│  - 监控服务状态                      │
│  - 自动重启崩溃的服务                │
│  - 健康检查                          │
└──────────────┬──────────────────────┘
               │
       ┌───────┴────────┐
       │                │
       ▼                ▼
┌─────────────┐  ┌─────────────┐
│   Node.js   │  │   Frontend  │
│   Backend   │  │   (Vite)    │
│  (Port 3001)│  │  (Port 5173)│
└─────────────┘  └─────────────┘
```

## 故障排查

### 问题1: 服务无法启动

**检查步骤：**
1. 确认端口未被占用
   ```bash
   netstat -ano | findstr ":3001"
   netstat -ano | findstr ":5173"
   ```

2. 查看日志文件
   ```bash
   type logs\node-backend.log
   type logs\frontend.log
   ```

3. 手动启动服务测试
   ```bash
   cd server && node simple-server.js
   cd publisher-web && npm run dev
   ```

### 问题2: 服务频繁重启

**可能原因：**
- 端口被占用
- 依赖包缺失
- 配置文件错误

**解决方法：**
1. 检查日志文件
2. 确认依赖已安装
   ```bash
   cd server && npm install
   cd publisher-web && npm install
   ```

### 问题3: API 请求失败

**检查步骤：**
1. 确认后端服务正在运行
   ```bash
   curl http://localhost:3001/api/health
   ```

2. 检查前端代理配置
   ```bash
   # publisher-web/vite.config.ts
   proxy: {
     '/api': {
       target: 'http://localhost:3001',
       changeOrigin: true,
     }
   }
   ```

## 性能优化建议

### 1. 使用 PM2（生产环境推荐）

```bash
# 安装 PM2
npm install -g pm2

# 启动服务
pm2 start server/simple-server.js --name "backend"
pm2 start "npm run dev" --name "frontend" --cwd publisher-web

# 查看状态
pm2 status

# 查看日志
pm2 logs

# 停止服务
pm2 stop all
```

### 2. 使用 Docker（推荐用于部署）

创建 `docker-compose.yml`:
```yaml
version: '3'
services:
  backend:
    build: ./server
    ports:
      - "3001:3001"
    restart: always
    
  frontend:
    build: ./publisher-web
    ports:
      - "5173:5173"
    depends_on:
      - backend
    restart: always
```

## 总结

通过以上改进，服务稳定性得到显著提升：

| 改进项 | 之前 | 之后 |
|--------|------|------|
| 崩溃恢复 | ❌ 手动重启 | ✅ 自动重启 |
| 健康检查 | ❌ 无 | ✅ 每30秒检查 |
| 错误处理 | ❌ 进程退出 | ✅ 捕获并记录 |
| 日志记录 | ⚠️ 部分 | ✅ 完整记录 |
| 监控告警 | ❌ 无 | ✅ 实时监控 |

**建议：**
- 开发环境：使用 `start-stable.bat`
- 生产环境：使用 PM2 或 Docker
- 定期检查日志文件
- 监控服务健康状态
