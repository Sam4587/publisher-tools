.PHONY: build clean test all build-core build-web dev serve serve-web stop deps help install

PID_DIR := ./pids
LOG_DIR := ./logs

all: build

build: build-core
	@echo "Build complete!"

build-core:
	@echo "Building publisher-core..."
	@mkdir -p bin $(PID_DIR) $(LOG_DIR) cookies uploads
	cd publisher-core && go mod tidy
	cd publisher-core && go build -o ../bin/publisher ./cmd/cli
	cd publisher-core && go build -o ../bin/publisher-server ./cmd/server
	@echo "Backend built: bin/publisher-server, bin/publisher"

build-web:
	@echo "Building publisher-web..."
	cd publisher-web && npm install
	cd publisher-web && npm run build
	@echo "Frontend built: publisher-web/dist/"

build-xiaohongshu:
	@echo "Building xiaohongshu-publisher..."
	cd xiaohongshu-publisher && go build -o xhs-publisher .

build-douyin-toutiao:
	@echo "Building douyin-toutiao publisher..."
	cd douyin-toutiao && go build -o publisher .

clean:
	rm -rf bin/
	rm -rf $(PID_DIR)/
	rm -rf $(LOG_DIR)/
	rm -rf publisher-web/dist/
	rm -f xiaohongshu-publisher/xhs-publisher
	rm -f douyin-toutiao/publisher

test:
	cd publisher-core && go test ./... -v

test-core:
	cd publisher-core && go test ./... -v

deps:
	cd publisher-core && go mod tidy
	cd xiaohongshu-publisher && go mod tidy
	cd douyin-toutiao && go mod tidy
	cd publisher-web && npm install

install: deps build
	@mkdir -p cookies uploads

serve:
	@mkdir -p $(PID_DIR) $(LOG_DIR) cookies uploads
	@if [ -f $(PID_DIR)/server.pid ]; then \
		echo "Server already running (PID: $$(cat $(PID_DIR)/server.pid))"; \
		exit 1; \
	fi
	./bin/publisher-server -port 8080 2>&1 | tee $(LOG_DIR)/server.log &
	echo $$! > $(PID_DIR)/server.pid
	@echo "Server started on port 8080 (PID: $$(cat $(PID_DIR)/server.pid))"

serve-web:
	@mkdir -p $(PID_DIR) $(LOG_DIR)
	@if [ -f $(PID_DIR)/web.pid ]; then \
		echo "Web server already running (PID: $$(cat $(PID_DIR)/web.pid))"; \
		exit 1; \
	fi
	cd publisher-web && npm run dev 2>&1 | tee ../$(LOG_DIR)/web.log &
	echo $$! > $(PID_DIR)/web.pid
	@echo "Web server started on port 5173 (PID: $$(cat $(PID_DIR)/web.pid))"

stop:
	@if [ -f $(PID_DIR)/server.pid ]; then \
		kill $$(cat $(PID_DIR)/server.pid) 2>/dev/null || true; \
		rm -f $(PID_DIR)/server.pid; \
		echo "Server stopped"; \
	fi
	@if [ -f $(PID_DIR)/web.pid ]; then \
		kill $$(cat $(PID_DIR)/web.pid) 2>/dev/null || true; \
		rm -f $(PID_DIR)/web.pid; \
		echo "Web server stopped"; \
	fi

status:
	@echo "=== Server Status ==="
	@if [ -f $(PID_DIR)/server.pid ]; then \
		if kill -0 $$(cat $(PID_DIR)/server.pid) 2>/dev/null; then \
			echo "Backend: RUNNING (PID: $$(cat $(PID_DIR)/server.pid))"; \
		else \
			echo "Backend: STOPPED (stale PID file)"; \
		fi \
	else \
		echo "Backend: STOPPED"; \
	fi
	@if [ -f $(PID_DIR)/web.pid ]; then \
		if kill -0 $$(cat $(PID_DIR)/web.pid) 2>/dev/null; then \
			echo "Frontend: RUNNING (PID: $$(cat $(PID_DIR)/web.pid))"; \
		else \
			echo "Frontend: STOPPED (stale PID file)"; \
		fi \
	else \
		echo "Frontend: STOPPED"; \
	fi

dev: stop build
	@mkdir -p $(PID_DIR) $(LOG_DIR) cookies uploads
	@echo "Starting development servers..."
	./bin/publisher-server -port 8080 2>&1 | tee $(LOG_DIR)/server.log &
	echo $$! > $(PID_DIR)/server.pid
	@sleep 1
	cd publisher-web && npm run dev 2>&1 | tee ../$(LOG_DIR)/web.log &
	echo $$! > $(PID_DIR)/web.pid
	@echo ""
	@echo "===================================="
	@echo "  Development servers started!"
	@echo "  Backend:  http://localhost:8080"
	@echo "  Frontend: http://localhost:5173"
	@echo "  Logs:     $(LOG_DIR)/"
	@echo "===================================="
	@echo ""
	@echo "Use 'make stop' to stop all servers"
	@echo "Use 'make status' to check server status"

logs:
	@echo "=== Server Logs (Ctrl+C to exit) ==="
	@if [ -f $(LOG_DIR)/server.log ]; then \
		tail -f $(LOG_DIR)/server.log; \
	else \
		echo "No server log found. Start server with 'make serve'"; \
	fi

help:
	@echo "Publisher Tools - 开发命令"
	@echo ""
	@echo "构建命令:"
	@echo "  make build              - 编译后端服务"
	@echo "  make build-web          - 编译前端生产版本"
	@echo "  make deps               - 安装所有依赖"
	@echo "  make install            - 安装依赖并编译"
	@echo "  make clean              - 清理编译产物"
	@echo ""
	@echo "运行命令:"
	@echo "  make dev                - 启动开发环境 (前后端)"
	@echo "  make serve              - 仅启动后端服务"
	@echo "  make serve-web          - 仅启动前端服务"
	@echo "  make stop               - 停止所有服务"
	@echo "  make status             - 查看服务状态"
	@echo "  make logs               - 查看服务日志"
	@echo ""
	@echo "测试命令:"
	@echo "  make test               - 运行所有测试"
	@echo "  make test-core          - 测试核心库"
	@echo ""
	@echo "端口:"
	@echo "  后端: 8080"
	@echo "  前端: 5173"
