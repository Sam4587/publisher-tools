# 快速开始指南

## 🚀 快速启动（推荐）

### Windows 用户

#### 1. 双击运行部署脚本
```
deploy.bat
```

这个脚本会自动：
- ✅ 检查环境（Go、Node.js、npm）
- ✅ 创建必要的目录
- ✅ 编译后端服务
- ✅ 安装前端依赖
- ✅ 启动后端和前端服务
- ✅ 测试 API 连接

#### 2. 访问系统
- **前端界面**: http://localhost:5173
- **后端 API**: http://localhost:8080
- **API 文档**: http://localhost:8080/api/v1/pipeline-templates

#### 3. 停止服务
```
stop.bat
```

---

## 📋 环境要求

### 必需软件
- **Go** 1.21+ - [下载地址](https://golang.org/dl/)
- **Node.js** 18+ - [下载地址](https://nodejs.org/)
- **npm** (随 Node.js 一起安装)

### 可选软件
- **curl** - 用于 API 测试（Windows 10+ 已内置）

---

## 🛠️ 手动启动

### 启动后端服务

```bash
# 1. 进入后端目录
cd publisher-core

# 2. 编译后端
go build -o ../bin/publisher-server.exe ./cmd/server

# 3. 启动后端
../bin/publisher-server.exe -port 8080
```

### 启动前端服务

```bash
# 1. 进入前端目录
cd publisher-web

# 2. 安装依赖（首次运行）
npm install

# 3. 启动前端
npm run dev
```

---

## 🧪 测试 API

### 测试流水线模板列表

```bash
curl http://localhost:8080/api/v1/pipeline-templates
```

### 测试创建流水线

```bash
curl -X POST http://localhost:8080/api/v1/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "content-publish-v1",
    "name": "测试流水线",
    "config": {
      "platforms": ["douyin"]
    }
  }'
```

### 测试执行流水线

```bash
curl -X POST http://localhost:8080/api/v1/pipelines/content-publish-v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "input": {
      "topic": "测试主题",
      "keywords": ["测试"],
      "platforms": ["douyin"]
    }
  }'
```

---

## 📚 预定义流水线模板

### 1. 内容发布流水线 (content-publish-v1)
- **功能**: 从内容生成到多平台发布的完整流程
- **步骤**: 内容生成 → 内容优化 → 质量评分 → 发布执行 → 数据采集
- **预计耗时**: 15-20 分钟

### 2. 视频处理流水线 (video-processing-v1)
- **功能**: 视频下载、转录、切片、发布的完整流程
- **步骤**: 视频下载 → 语音转录 → 内容改写 → 视频切片 → 发布执行
- **预计耗时**: 20-30 分钟

### 3. 热点分析流水线 (hotspot-analysis-v1)
- **功能**: 抓取热点、分析趋势、生成内容的完整流程
- **步骤**: 热点抓取 → 趋势分析 → 内容生成 → 发布执行
- **预计耗时**: 10-15 分钟

### 4. 数据采集流水线 (data-collection-v1)
- **功能**: 从多平台采集发布数据和性能指标
- **步骤**: 多平台数据采集 → 数据分析 → 报告生成
- **预计耗时**: 5-10 分钟

---

## 🔧 故障排查

### 问题1: deploy.bat 一闪而过

**解决方案**:
- 确保已安装 Go 和 Node.js
- 检查是否有足够的磁盘空间
- 查看 `logs/build.log` 和 `logs/npm-install.log` 了解详细错误

### 问题2: 后端服务启动失败

**解决方案**:
- 检查端口 8080 是否被占用
- 查看 `logs/publisher-server.log` 了解详细错误
- 尝试手动编译和启动后端

### 问题3: 前端服务启动失败

**解决方案**:
- 检查端口 5173 是否被占用
- 删除 `publisher-web/node_modules` 重新安装依赖
- 查看 `logs/vite-dev.log` 了解详细错误

### 问题4: API 测试失败

**解决方案**:
- 确保后端服务已启动
- 等待 5-10 秒让服务完全启动
- 检查防火墙设置
- 查看后端日志了解错误原因

---

## 📖 更多文档

- [架构设计文档](docs/architecture/automation-pipeline-design.md)
- [快速开始指南](docs/guides/automation-pipeline-quickstart.md)
- [集成指南](docs/guides/integration-guide.md)
- [实施总结](docs/implementation-summary.md)

---

## 🆘 获取帮助

如果遇到问题：

1. 查看日志文件（`logs/` 目录）
2. 检查环境要求
3. 查看故障排查章节
4. 提交 Issue 或联系技术支持

---

**祝您使用愉快！** 🎉
