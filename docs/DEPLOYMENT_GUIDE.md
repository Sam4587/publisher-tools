# Publisher Tools 部署指南

> 项目部署和运维文档
>
> 文档版本：v1.0
> 创建时间：2026-02-20
> 最后更新：2026-02-20

---

## 目录

- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [详细部署](#详细部署)
- [配置说明](#配置说明)
- [运维管理](#运维管理)
- [故障排查](#故障排查)

---

## 环境要求

### 系统要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+) / macOS / Windows
- **内存**: 最低 2GB，推荐 4GB+
- **磁盘**: 最低 10GB 可用空间
- **网络**: 稳定的互联网连接

### 软件依赖

#### 后端 (Go)
- Go 1.21 或更高版本
- SQLite 3.35 或更高版本

#### 前端 (Node.js)
- Node.js 18.x 或更高版本
- npm 9.x 或更高版本

#### 可选工具
- Docker 和 Docker Compose（推荐）
- Git
- yt-dlp（视频下载功能）
- Faster-Whisper（语音转录功能）

---

## 快速开始

### 方式一：Docker 部署（推荐）

1. **克隆项目**
```bash
git clone <repository-url>
cd publisher-tools
```

2. **构建并启动**
```bash
docker-compose up -d
```

3. **访问应用**
- 前端: http://localhost:3000
- 后端 API: http://localhost:8080

### 方式二：本地开发部署

1. **安装依赖**

后端：
```bash
cd publisher-core
go mod download
```

前端：
```bash
cd publisher-web
npm install
```

2. **配置环境变量**

创建 `.env` 文件：
```env
# 后端配置
PORT=8080
DATABASE_PATH=./data/publisher.db
LOG_LEVEL=info

# AI 服务配置
OPENROUTER_API_KEY=your_openrouter_key
DEEPSEEK_API_KEY=your_deepseek_key
GROQ_API_KEY=your_groq_key
GOOGLE_AI_API_KEY=your_google_key
```

3. **启动服务**

启动后端：
```bash
cd publisher-core
go run cmd/server/main.go
```

启动前端（新终端）：
```bash
cd publisher-web
npm run dev
```

4. **访问应用**
- 前端: http://localhost:5173
- 后端 API: http://localhost:8080

---

## 详细部署

### 1. 后端部署

#### 编译后端
```bash
cd publisher-core
go build -o publisher-server cmd/server/main.go
```

#### 配置后端

创建配置文件 `config.yaml`：
```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  path: "./data/publisher.db"
  auto_migrate: true

logging:
  level: "info"
  file: "./logs/publisher.log"

ai:
  default_provider: "openrouter"
  providers:
    openrouter:
      api_key: "${OPENROUTER_API_KEY}"
      base_url: "https://openrouter.ai/api/v1"
      model: "openai/gpt-4"
    deepseek:
      api_key: "${DEEPSEEK_API_KEY}"
      base_url: "https://api.deepseek.com/v1"
      model: "deepseek-chat"
```

#### 启动后端服务

使用 systemd（Linux）：
```bash
sudo cp scripts/publisher.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start publisher
sudo systemctl enable publisher
```

手动启动：
```bash
./publisher-server
```

### 2. 前端部署

#### 构建前端
```bash
cd publisher-web
npm run build
```

#### 部署到 Nginx

创建 Nginx 配置 `/etc/nginx/sites-available/publisher`：
```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /var/www/publisher-web/dist;
    index index.html;

    # 前端路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

启用配置：
```bash
sudo ln -s /etc/nginx/sites-available/publisher /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 3. Docker 部署

#### Dockerfile（后端）
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY publisher-core .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o publisher-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/publisher-server .
COPY --from=builder /app/config.yaml .

EXPOSE 8080
CMD ["./publisher-server"]
```

#### Dockerfile（前端）
```dockerfile
FROM node:18-alpine AS builder

WORKDIR /app
COPY publisher-web/package*.json ./
RUN npm install
COPY publisher-web ./
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### docker-compose.yml
```yaml
version: '3.8'

services:
  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    environment:
      - OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "3000:80"
    depends_on:
      - backend
    restart: unless-stopped
```

---

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `PORT` | 后端服务端口 | 8080 | 否 |
| `DATABASE_PATH` | SQLite 数据库路径 | ./data/publisher.db | 否 |
| `LOG_LEVEL` | 日志级别 (debug/info/warn/error) | info | 否 |
| `OPENROUTER_API_KEY` | OpenRouter API 密钥 | - | 否 |
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 | - | 否 |
| `GROQ_API_KEY` | Groq API 密钥 | - | 否 |
| `GOOGLE_AI_API_KEY` | Google AI API 密钥 | - | 否 |

### AI 服务配置

系统支持多个 AI 提供商，可以配置优先级和降级策略：

1. **OpenRouter**: 支持多种模型，免费额度
2. **Groq**: 最快的推理速度
3. **DeepSeek**: 国产 AI，成本低
4. **Google AI**: Gemini 模型

配置示例：
```yaml
ai:
  providers:
    - name: "OpenRouter GPT-4"
      provider: "openrouter"
      api_key: "${OPENROUTER_API_KEY}"
      model: "openai/gpt-4"
      priority: 100
      is_default: true

    - name: "Groq Llama 3.3 70B"
      provider: "groq"
      api_key: "${GROQ_API_KEY}"
      model: "llama-3.3-70b-versatile"
      priority: 90
```

### 平台配置

支持的平台：
- 抖音 (douyin)
- 今日头条 (toutiao)
- 小红书 (xiaohongshu)
- 微博 (weibo)
- B站 (bilibili)

每个平台需要配置：
- 平台 ID
- 登录 Cookie
- 发布接口配置

---

## 运维管理

### 日志管理

#### 后端日志
```bash
# 查看实时日志
tail -f logs/publisher.log

# 查看错误日志
grep ERROR logs/publisher.log

# 日志轮转
logrotate -f /etc/logrotate.d/publisher
```

#### 前端日志
```bash
# Nginx 访问日志
tail -f /var/log/nginx/access.log

# Nginx 错误日志
tail -f /var/log/nginx/error.log
```

### 数据库管理

#### 备份数据库
```bash
# 备份
cp data/publisher.db data/publisher.db.backup.$(date +%Y%m%d)

# 压缩备份
gzip data/publisher.db.backup.$(date +%Y%m%d)
```

#### 恢复数据库
```bash
# 停止服务
systemctl stop publisher

# 恢复备份
cp data/publisher.db.backup.20260220 data/publisher.db

# 启动服务
systemctl start publisher
```

#### 数据库维护
```bash
# 进入数据库
sqlite3 data/publisher.db

# 查看表
.tables

# 查看数据库大小
.databases

# 优化数据库
VACUUM;
```

### 监控指标

#### 系统监控
```bash
# CPU 使用率
top

# 内存使用
free -h

# 磁盘使用
df -h

# 网络连接
netstat -tuln
```

#### 应用监控
- API 响应时间
- 数据库查询性能
- AI 调用成功率
- 任务队列积压情况

### 性能优化

#### 后端优化
1. 启用数据库连接池
2. 缓存热点数据
3. 异步处理耗时任务
4. 限制并发请求数

#### 前端优化
1. 启用 CDN 加速
2. 开启 Gzip 压缩
3. 实现代码分割
4. 优化图片资源

---

## 故障排查

### 常见问题

#### 1. 后端无法启动

**症状**: 服务启动失败或立即退出

**排查步骤**:
```bash
# 检查端口占用
lsof -i :8080

# 检查日志
tail -f logs/publisher.log

# 手动启动查看错误
./publisher-server
```

**解决方案**:
- 释放占用端口
- 检查配置文件语法
- 确认数据库文件权限

#### 2. 前端页面空白

**症状**: 访问前端页面显示空白

**排查步骤**:
```bash
# 检查 Nginx 配置
nginx -t

# 查看 Nginx 错误日志
tail -f /var/log/nginx/error.log

# 检查前端文件
ls -la /var/www/publisher-web/dist
```

**解决方案**:
- 重新构建前端
- 检查 Nginx 配置
- 确认文件路径正确

#### 3. API 请求失败

**症状**: 前端无法调用后端 API

**排查步骤**:
```bash
# 测试后端 API
curl http://localhost:8080/api/health

# 检查防火墙
sudo ufw status

# 检查代理配置
cat /etc/nginx/sites-available/publisher
```

**解决方案**:
- 确认后端服务运行
- 配置反向代理
- 检查 CORS 设置

#### 4. AI 调用失败

**症状**: AI 功能无法使用

**排查步骤**:
```bash
# 检查 API 密钥
echo $OPENROUTER_API_KEY

# 查看后端日志
grep "AI" logs/publisher.log | tail -20

# 测试 API 连接
curl -X POST https://openrouter.ai/api/v1/chat/completions \
  -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"openai/gpt-4","messages":[{"role":"user","content":"test"}]}'
```

**解决方案**:
- 验证 API 密钥有效
- 检查网络连接
- 确认 API 配额未超限

### 调试技巧

#### 启用调试模式
```bash
# 设置日志级别为 debug
export LOG_LEVEL=debug
./publisher-server
```

#### 查看详细错误
```bash
# 启用 Go 运行时追踪
export GODEBUG=gctrace=1
./publisher-server
```

#### 性能分析
```bash
# CPU 性能分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap
```

---

## 安全建议

1. **使用 HTTPS**
   - 配置 SSL 证书
   - 强制 HTTPS 重定向

2. **限制访问**
   - 配置防火墙规则
   - 使用 API 认证

3. **数据备份**
   - 定期备份数据库
   - 加密敏感数据

4. **更新维护**
   - 及时更新依赖
   - 监控安全公告

---

## 升级指南

### 后端升级
```bash
# 停止服务
systemctl stop publisher

# 备份数据
cp data/publisher.db data/publisher.db.backup

# 拉取新代码
git pull

# 重新编译
cd publisher-core
go build -o publisher-server cmd/server/main.go

# 启动服务
systemctl start publisher
```

### 前端升级
```bash
# 拉取新代码
git pull

# 安装依赖
cd publisher-web
npm install

# 重新构建
npm run build

# 部署文件
cp -r dist/* /var/www/publisher-web/dist/
```

---

## 支持

如有问题，请：
1. 查看本文档的故障排查部分
2. 检查 GitHub Issues
3. 联系技术支持团队

---

**文档维护**: 开发团队
**最后更新**: 2026-02-20
**下次更新**: 根据版本更新
