#!/bin/bash

# 自动化流水线系统集成测试脚本

set -e

echo "========================================="
echo "  自动化流水线系统集成测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果统计
TESTS_PASSED=0
TESTS_FAILED=0

# 函数：打印成功消息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
    ((TESTS_PASSED++))
}

# 函数：打印失败消息
print_failure() {
    echo -e "${RED}✗ $1${NC}"
    ((TESTS_FAILED++))
}

# 函数：打印警告消息
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# 函数：打印信息
print_info() {
    echo -e "ℹ $1"
}

# 函数：测试 API 端点
test_api() {
    local url=$1
    local method=$2
    local data=$3
    local description=$4

    print_info "测试: $description"

    if [ -n "$data" ]; then
        response=$(curl -s -X $method -H "Content-Type: application/json" -d "$data" "$url")
    else
        response=$(curl -s -X $method "$url")
    fi

    # 检查响应
    if echo "$response" | grep -q "error\|Error\|ERROR"; then
        print_failure "$description - 响应包含错误"
        echo "响应: $response"
        return 1
    else
        print_success "$description"
        return 0
    fi
}

# 函数：等待服务就绪
wait_for_service() {
    local url=$1
    local max_attempts=30
    local attempt=1

    print_info "等待服务就绪: $url"

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            print_success "服务已就绪"
            return 0
        fi

        print_info "等待服务启动... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done

    print_failure "服务启动超时"
    return 1
}

# ========================================
# 1. 环境检查
# ========================================
echo "========================================="
echo "1. 环境检查"
echo "========================================="

# 检查 Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    print_success "Go 已安装: $GO_VERSION"
else
    print_failure "Go 未安装"
    exit 1
fi

# 检查 Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    print_success "Node.js 已安装: $NODE_VERSION"
else
    print_failure "Node.js 未安装"
    exit 1
fi

# 检查 npm
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    print_success "npm 已安装: $NPM_VERSION"
else
    print_failure "npm 未安装"
    exit 1
fi

# ========================================
# 2. 后端服务启动
# ========================================
echo ""
echo "========================================="
echo "2. 启动后端服务"
echo "========================================="

cd publisher-core

# 编译后端
print_info "编译后端服务..."
if go build -o ../bin/publisher-server ./cmd/server 2>&1; then
    print_success "后端编译成功"
else
    print_failure "后端编译失败"
    exit 1
fi

# 启动后端服务
print_info "启动后端服务..."
../bin/publisher-server -port 8080 > ../logs/publisher-server.log 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > ../pids/backend.pid

# 等待后端服务就绪
if wait_for_service "http://localhost:8080/api/v1/platforms"; then
    print_success "后端服务已就绪"
else
    print_failure "后端服务启动失败"
    cat ../logs/publisher-server.log
    exit 1
fi

cd ..

# ========================================
# 3. 前端服务启动
# ========================================
echo ""
echo "========================================="
echo "3. 启动前端服务"
echo "========================================="

cd publisher-web

# 检查依赖
if [ ! -d "node_modules" ]; then
    print_info "安装前端依赖..."
    if npm install > ../logs/npm-install.log 2>&1; then
        print_success "前端依赖安装成功"
    else
        print_failure "前端依赖安装失败"
        cat ../logs/npm-install.log
        exit 1
    fi
fi

# 启动前端开发服务器
print_info "启动前端开发服务器..."
npm run dev > ../logs/vite-dev.log 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > ../pids/frontend.pid

# 等待前端服务就绪
sleep 5
if curl -s -f "http://localhost:5173" > /dev/null 2>&1; then
    print_success "前端服务已就绪"
else
    print_warning "前端服务可能未就绪，但继续测试"
fi

cd ..

# ========================================
# 4. API 测试
# ========================================
echo ""
echo "========================================="
echo "4. API 测试"
echo "========================================="

BASE_URL="http://localhost:8080"

# 测试流水线模板列表
test_api "$BASE_URL/api/v1/pipeline-templates" "GET" "" "获取流水线模板列表"

# 测试获取内容发布模板
test_api "$BASE_URL/api/v1/pipeline-templates/content-publish-v1" "GET" "" "获取内容发布模板"

# 测试使用模板创建流水线
CREATE_PIPELINE_DATA='{"name":"测试流水线","config":{"platforms":["douyin"]}}'
test_api "$BASE_URL/api/v1/pipeline-templates/content-publish-v1/use" "POST" "$CREATE_PIPELINE_DATA" "使用模板创建流水线"

# 测试获取流水线列表
test_api "$BASE_URL/api/v1/pipelines" "GET" "" "获取流水线列表"

# 测试执行流水线（使用第一个流水线ID）
PIPELINE_ID="content-publish-v1"  # 使用预定义模板ID
EXECUTE_DATA='{"input":{"topic":"测试主题","keywords":["测试"],"platforms":["douyin"]}}'
test_api "$BASE_URL/api/v1/pipelines/$PIPELINE_ID/execute" "POST" "$EXECUTE_DATA" "执行流水线"

# 测试获取执行列表
test_api "$BASE_URL/api/v1/executions" "GET" "" "获取执行列表"

# 测试获取监控统计
test_api "$BASE_URL/api/v1/monitoring/stats" "GET" "" "获取监控统计"

# ========================================
# 5. WebSocket 测试
# ========================================
echo ""
echo "========================================="
echo "5. WebSocket 测试"
echo "========================================="

print_info "测试 WebSocket 连接..."

# 使用 Python 测试 WebSocket（如果可用）
if command -v python3 &> /dev/null; then
    python3 << 'EOF' || true
import asyncio
import websockets
import json

async def test_websocket():
    try:
        uri = "ws://localhost:8080/ws/monitor"
        async with websockets.connect(uri) as websocket:
            print("✓ WebSocket 连接成功")

            # 发送订阅消息
            subscribe_msg = {"type": "subscribe", "topics": ["monitor"]}
            await websocket.send(json.dumps(subscribe_msg))
            print("✓ 发送订阅消息成功")

            # 等待响应
            response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
            print(f"✓ 收到响应: {response[:50]}...")

    except Exception as e:
        print(f"✗ WebSocket 测试失败: {e}")

asyncio.run(test_websocket())
EOF
else
    print_warning "Python3 未安装，跳过 WebSocket 测试"
fi

# ========================================
# 6. 清理
# ========================================
echo ""
echo "========================================="
echo "6. 清理测试环境"
echo "========================================="

# 停止后端服务
if [ -f "pids/backend.pid" ]; then
    BACKEND_PID=$(cat pids/backend.pid)
    print_info "停止后端服务 (PID: $BACKEND_PID)"
    kill $BACKEND_PID 2>/dev/null || true
    rm pids/backend.pid
fi

# 停止前端服务
if [ -f "pids/frontend.pid" ]; then
    FRONTEND_PID=$(cat pids/frontend.pid)
    print_info "停止前端服务 (PID: $FRONTEND_PID)"
    kill $FRONTEND_PID 2>/dev/null || true
    rm pids/frontend.pid
fi

# 清理可能残留的进程
pkill -f "publisher-server" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true

# ========================================
# 7. 测试结果汇总
# ========================================
echo ""
echo "========================================="
echo "7. 测试结果汇总"
echo "========================================="

echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================="
    echo "  所有测试通过！"
    echo "=========================================${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}========================================="
    echo "  部分测试失败，请查看日志"
    echo "=========================================${NC}"
    exit 1
fi
