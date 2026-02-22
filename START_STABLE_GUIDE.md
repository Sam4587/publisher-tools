# start-stable.bat - 完善版使用说明

## 改进内容

### 1. 后台运行服务管理器
- 使用 `run_hidden.js` 启动服务管理器
- 完全隐藏窗口，不显示任何控制台
- 服务在后台静默运行

### 2. 自动关闭启动窗口
- 启动完成后自动打开浏览器
- 3秒后自动关闭启动窗口
- 无需手动关闭

### 3. 服务状态验证
- 检查前端服务（端口 5173）
- 检查后端服务（端口 3001）
- 显示服务就绪状态

## 启动流程

```
[1/4] Cleaning old processes...
  └─ 停止端口 8080, 3001, 5173 上的进程

[2/4] Starting service manager...
  └─ 使用 run_hidden.js 后台启动服务管理器
  └─ 日志输出到 logs/service-manager.log

[3/4] Waiting for services to start...
  └─ 等待 5 秒让服务启动

[4/4] Verifying service status...
  └─ 检查前端端口 5173
  └─ 检查后端端口 3001
  └─ 显示服务状态

Service Startup Complete!
  └─ 自动打开浏览器
  └─ 3秒后关闭启动窗口
```

## 与 start.bat 的对比

| 特性 | start.bat | start-stable.bat |
|------|-----------|------------------|
| 服务监控 | ❌ 无 | ✅ 自动监控（每30秒） |
| 崩溃重启 | ❌ 手动 | ✅ 自动重启（5秒延迟） |
| 健康检查 | ❌ 无 | ✅ HTTP健康检查 |
| 错误处理 | ⚠️ 基础 | ✅ 完整错误处理 |
| 日志管理 | ⚠️ 基础 | ✅ 完整日志记录 |
| 窗口管理 | ✅ 后台运行 | ✅ 后台运行 |
| 服务验证 | ✅ 端口检查 | ✅ 端口检查 |

## 使用方法

### 方式1: 双击运行（推荐）
```
双击 start-stable.bat
```

### 方式2: 命令行运行
```bash
start-stable.bat
```

## 启动后的状态

### 可见窗口
- ✅ 浏览器自动打开 http://localhost:5173
- ❌ 无其他可见窗口（服务管理器完全隐藏）

### 后台进程
- Node.js 进程（服务管理器）
  - 监控服务状态
  - 自动重启崩溃的服务
  - 健康检查

- Node.js 进程（后端服务）
  - 端口 3001
  - 提供 API 服务

- Node.js 进程（前端服务）
  - 端口 5173
  - Vite 开发服务器

## 日志文件

所有日志保存在 `logs/` 目录：

```
logs/
├── service-manager.log  # 服务管理器日志
├── node-backend.log     # 后端服务日志
└── frontend.log         # 前端服务日志
```

### 查看日志

**实时查看：**
```bash
# Windows PowerShell
Get-Content logs\service-manager.log -Wait

# Git Bash / WSL
tail -f logs/service-manager.log
```

**查看历史：**
```bash
type logs\service-manager.log
```

## 停止服务

### 方式1: 使用 stop.bat
```bash
stop.bat
```

### 方式2: 手动停止
```bash
# 停止所有 Node.js 进程
taskkill /F /IM node.exe

# 或停止特定端口
netstat -ano | findstr ":3001"
taskkill /PID <PID> /F
```

## 故障排查

### 问题1: 服务未启动

**检查步骤：**
1. 查看日志文件
   ```bash
   type logs\service-manager.log
   ```

2. 检查端口占用
   ```bash
   netstat -ano | findstr ":3001"
   netstat -ano | findstr ":5173"
   ```

3. 手动启动测试
   ```bash
   node service-manager.js
   ```

### 问题2: 浏览器未自动打开

**解决方法：**
1. 手动访问 http://localhost:5173
2. 检查默认浏览器设置
3. 查看启动日志

### 问题3: 服务频繁重启

**可能原因：**
- 端口被占用
- 依赖包缺失
- 配置错误

**解决方法：**
1. 查看日志文件
2. 安装依赖
   ```bash
   cd server && npm install
   cd publisher-web && npm install
   ```
3. 清理并重启
   ```bash
   stop.bat
   start-stable.bat
   ```

## 服务管理器特性

### 自动监控
- 每 30 秒检查服务状态
- 显示进程 ID、端口、运行时间
- 健康状态检查

### 自动重启
- 服务崩溃后 5 秒自动重启
- 防止服务长时间不可用
- 记录重启事件

### 健康检查
- HTTP 请求到 `/api/health`
- 2 秒超时
- 检测服务是否真正可用

### 日志记录
- 时间戳记录
- 错误追踪
- 性能监控

## 推荐使用场景

### 开发环境
- ✅ 使用 `start-stable.bat`
- 自动监控和重启
- 完整的日志记录
- 后台静默运行

### 生产环境
- ⚠️ 建议使用 PM2
- 或使用 Docker 容器
- 配置监控告警

## 总结

改进后的 `start-stable.bat` 提供：

✅ **完全后台运行** - 无可见窗口，静默运行
✅ **自动打开浏览器** - 启动后自动访问前端
✅ **自动关闭窗口** - 3秒后自动关闭启动窗口
✅ **服务状态验证** - 检查服务是否真正启动
✅ **自动监控重启** - 服务崩溃自动恢复
✅ **完整日志记录** - 所有事件都有记录

现在你可以：
1. 双击 `start-stable.bat`
2. 等待浏览器自动打开
3. 开始使用应用
4. 无需关心后台服务，它们会自动运行和恢复
