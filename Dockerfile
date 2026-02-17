# Publisher Tools - 多平台内容发布工具
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make nodejs npm chromium

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY publisher-core/go.mod publisher-core/go.sum ./
RUN go mod download

# 复制源代码
COPY publisher-core ./

# 构建后端
RUN cd cmd/server && go build -o /build/publisher-server

# 构建前端
FROM node:18-alpine AS frontend-builder
WORKDIR /build
COPY publisher-web/package*.json ./
RUN npm install
COPY publisher-web ./
RUN npm run build

# 运行时镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates chromium tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 publisher &&     adduser -u 1000 -G publisher -s /bin/sh -D publisher

# 创建必要目录
RUN mkdir -p /app/uploads /app/cookies /app/data /app/logs /app/pids &&     chown -R publisher:publisher /app

WORKDIR /app

# 复制构建产物
COPY --from=builder /build/publisher-server /app/
COPY --from=frontend-builder /build/dist /app/web

# 切换用户
USER publisher

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["/app/publisher-server", "-port", "8080", "-headless=true"]
