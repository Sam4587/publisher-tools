# 🚀 快速启动指南

## 方案0：桌面快捷方式启动（最简单推荐）

### 创建桌面快捷方式
```bash
# 在项目根目录运行
powershell -ExecutionPolicy Bypass -File create-desktop-shortcuts.ps1
```

运行后将在桌面创建三个快捷方式：
- **Publisher Tools Start.lnk** - PowerShell版本启动器（推荐）
- **Publisher Tools Start(BAT).lnk** - 批处理版本启动器
- **Publisher Tools Stop.lnk** - 服务停止器

### 使用方法
1. 双击 **Publisher Tools Start** 快捷方式启动服务
2. 服务启动后浏览器会自动打开前端界面
3. 使用完毕后双击 **Publisher Tools Stop** 停止服务

> **注意**: 启动后不要关闭命令行窗口，那是服务运行的载体

## ⚠️ 当前环境说明

由于当前环境限制，前端开发服务器可能无法直接运行（Bus error）。
请使用以下替代方案：

## 方案1：使用Docker部署（推荐）

```bash
# 安装Docker和Docker Compose
# 然后运行：
docker-compose up -d

# 访问服务
# http://localhost:8080
```

## 方案2：本地完整环境启动

### 后端启动

```bash
cd publisher-core

# 首次需要编译（需要Go环境）
go build -o bin/publisher-server cmd/server/main.go

# 启动服务
./bin/publisher-server -port 8080
```

### 前端启动

```bash
cd publisher-web

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 如果遇到Bus error，尝试：
npm run build
npm run preview
```

## 方案3：使用生产构建

```bash
# 构建前端
cd publisher-web
npm run build

# 使用静态文件服务器
npx serve -s dist -p 5173
```

## 验证服务

```bash
# 检查后端
# 如果使用测试服务器:
curl http://localhost:3001/api/health

# 如果使用Go后端:
curl http://localhost:8080/health

# 检查前端（浏览器）
# http://localhost:5173 或自动分配的可用端口（如5174）
# Vite会自动寻找可用端口，实际端口请查看终端输出
```

## 常见问题

### 1. 端口被占用

```bash
# 查找并杀死占用进程
lsof -i :8080
kill -9 <PID>
```

### 2. 前端启动失败

```bash
# 清除缓存重试
rm -rf node_modules package-lock.json
npm install
```

### 3. Go环境缺失

```bash
# Ubuntu/Debian
sudo apt-get install golang-go

# macOS
brew install go

# 或使用Docker
```

## 📞 需要帮助？

- 查看 README.md
- 查看日志: `./start.sh --logs`
- 提交Issue到GitHub
