# CGO 配置指南

> 启用 CGO 以运行 SQLite 相关测试
>
> 文档版本：v1.0
> 创建时间：2026-02-20

---

## 什么是 CGO？

CGO 是 Go 语言的 C 语言接口，允许 Go 程序调用 C 代码。SQLite 的 Go 驱动需要 CGO 才能正常工作。

---

## 为什么需要 CGO？

Publisher Tools 项目使用 SQLite 作为数据库，而 SQLite 的 Go 驱动（`go-sqlite3`）需要 CGO 才能编译和运行。

### 影响 CGO 的功能

- ✅ 数据库初始化
- ✅ 数据库迁移
- ✅ 数据库连接
- ✅ 数据库查询
- ✅ 事务处理

---

## 启用 CGO 的方法

### Windows

#### 方法 1: 使用环境变量（推荐）

```powershell
# PowerShell
$env:CGO_ENABLED = "1"

# 验证
go env CGO_ENABLED
```

#### 方法 2: 使用命令行参数

```powershell
# 临时启用（当前会话）
go test -tags sqlite3 ./database/...

# 或者在命令中指定
set CGO_ENABLED=1 && go test ./database/...
```

#### 方法 3: 使用 Makefile

创建 `Makefile`：

```makefile
# 启用 CGO 的测试
test-cgo:
	set CGO_ENABLED=1
	go test ./...

# 禁用 CGO 的测试
test-no-cgo:
	set CGO_ENABLED=0
	go test ./...
```

使用：
```powershell
make test-cgo
```

### Linux / macOS

#### 方法 1: 使用环境变量（推荐）

```bash
# Bash/Zsh
export CGO_ENABLED=1

# 验证
go env CGO_ENABLED
```

#### 方法 2: 使用命令行参数

```bash
# 临时启用
CGO_ENABLED=1 go test ./database/...
```

#### 方法 3: 使用 Makefile

创建 `Makefile`：

```makefile
# 启用 CGO 的测试
test-cgo:
	CGO_ENABLED=1 go test ./...

# 禁用 CGO 的测试
test-no-cgo:
	CGO_ENABLED=0 go test ./...
```

使用：
```bash
make test-cgo
```

#### 方法 4: 使用 .bashrc 或 .zshrc

在 `~/.bashrc` 或 `~/.zshrc` 中添加：

```bash
# Go CGO 配置
export CGO_ENABLED=1
```

然后重新加载配置：

```bash
source ~/.bashrc
# 或
source ~/.zshrc
```

### Docker 环境

如果使用 Docker，需要在 Dockerfile 中启用 CGO：

```dockerfile
FROM golang:1.21-alpine AS builder

# 安装 CGO 依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 设置环境变量
ENV CGO_ENABLED=1

WORKDIR /app
COPY . .
RUN go mod download
RUN go test ./...

# 最终镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
COPY --from=builder /app/publisher-server .
CMD ["./publisher-server"]
```

---

## 验证 CGO 是否启用

### 检查环境变量

```bash
go env CGO_ENABLED
```

输出应该是 `1`（已启用）或 `0`（未启用）。

### 测试编译

```bash
# 编译一个简单的 SQLite 程序
go build -o test_sqlite ./cmd/test_sqlite/

# 如果编译成功，说明 CGO 已正确启用
```

### 运行测试

```bash
# 运行数据库测试
go test ./database/... -v

# 如果没有 CGO 相关错误，说明 CGO 已正确启用
```

---

## 常见问题

### 问题 1: 找不到 gcc 编译器

**错误信息**:
```
gcc: command not found
```

**解决方案**:

#### Windows
安装 TDM-GCC 或 MinGW-w64：
- 访问 https://jmeubank.github.io/tdm-gcc/
- 下载并安装 TDM-GCC
- 确保 gcc 在 PATH 中

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install build-essential

# CentOS/RHEL
sudo yum groupinstall "Development Tools"
```

#### macOS
```bash
# 安装 Xcode Command Line Tools
xcode-select --install

# 或使用 Homebrew
brew install gcc
```

### 问题 2: 找不到 sqlite3.h

**错误信息**:
```
fatal error: sqlite3.h: No such file or directory
```

**解决方案**:

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get install libsqlite3-dev

# CentOS/RHEL
sudo yum install sqlite-devel
```

#### macOS
```bash
# 使用 Homebrew
brew install sqlite3
```

#### Windows
- TDM-GCC 通常包含 SQLite 头文件
- 或者手动下载 SQLite 源码

### 问题 3: 链接错误

**错误信息**:
```
undefined reference to `sqlite3_open`
```

**解决方案**:

确保在编译时启用了 CGO：
```bash
CGO_ENABLED=1 go build
```

### 问题 4: 测试仍然失败

**错误信息**:
```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work
```

**解决方案**:

1. 清理构建缓存：
```bash
go clean -cache
go clean -modcache
```

2. 重新编译：
```bash
CGO_ENABLED=1 go build
```

3. 确认环境变量：
```bash
go env CGO_ENABLED
```

---

## 最佳实践

### 1. 使用 Makefile

创建一个 `Makefile` 来管理 CGO 配置：

```makefile
# Go 配置
GO := go
CGO_ENABLED ?= 1

# 构建配置
BUILD_FLAGS := -v
TEST_FLAGS := -v -cover

# 构建二进制文件
build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(BUILD_FLAGS) ./cmd/server

# 运行测试
test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_FLAGS) ./...

# 运行特定测试
test-db:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_FLAGS) ./database/...

# 运行测试并生成覆盖率
test-coverage:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_FLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 清理
clean:
	$(GO) clean -cache
	$(GO) clean -modcache
	rm -f coverage.out coverage.html

.PHONY: build test test-db test-coverage clean
```

使用：
```bash
make test
make test-coverage
```

### 2. 使用脚本

创建 `scripts/setup-cgo.sh`：

```bash
#!/bin/bash

# CGO 设置脚本

echo "配置 CGO 环境..."

# 设置环境变量
export CGO_ENABLED=1

# 验证 GCC
if ! command -v gcc &> /dev/null; then
    echo "错误: 未找到 gcc 编译器"
    echo "请安装 gcc: sudo apt-get install build-essential"
    exit 1
fi

# 验证 SQLite 头文件
if ! [ -f "/usr/include/sqlite3.h" ] && ! [ -f "/usr/local/include/sqlite3.h" ]; then
    echo "错误: 未找到 sqlite3.h"
    echo "请安装 SQLite 开发库: sudo apt-get install libsqlite3-dev"
    exit 1
fi

echo "CGO 环境配置完成！"
echo "CGO_ENABLED: $CGO_ENABLED"
echo "GCC: $(gcc --version | head -n 1)"
```

使用：
```bash
chmod +x scripts/setup-cgo.sh
./scripts/setup-cgo.sh
```

### 3. 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - CGO_ENABLED=1
    volumes:
      - .:/app
    working_dir: /app
    command: go test ./...
```

使用：
```bash
docker-compose up
```

---

## CI/CD 配置

### GitHub Actions

创建 `.github/workflows/test.yml`：

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y build-essential libsqlite3-dev

    - name: Run tests
      env:
        CGO_ENABLED: 1
      run: go test ./... -v -cover

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
```

### GitLab CI

创建 `.gitlab-ci.yml`：

```yaml
test:
  image: golang:1.21

  before_script:
    - apt-get update && apt-get install -y build-essential libsqlite3-dev

  variables:
    CGO_ENABLED: "1"

  script:
    - go test ./... -v -cover

  coverage: '/coverage: \d+.\d+% of statements/'
```

---

## 总结

### 关键要点

1. **CGO 必需**: SQLite 驱动需要 CGO
2. **环境变量**: 使用 `CGO_ENABLED=1` 启用
3. **编译器**: 需要安装 gcc 或等效编译器
4. **依赖库**: 需要安装 SQLite 开发库

### 快速启动

```bash
# 启用 CGO
export CGO_ENABLED=1

# 安装依赖（Linux）
sudo apt-get install build-essential libsqlite3-dev

# 运行测试
go test ./... -v
```

---

**文档维护**: 开发团队
**最后更新**: 2026-02-20
**相关文档**: [测试报告](./TESTING_REPORT.md)
