# 开发指南

## 开发环境搭建

### 必需软件
- Go 1.21+
- Node.js 18+
- Chrome/Chromium 浏览器
- Git

### 项目克隆
```bash
git clone <repository-url>
cd publisher-tools
```

## 代码规范

### Go代码规范
- 使用 `logrus` 进行日志记录
- 错误使用 `github.com/pkg/errors` 包装
- 接口定义在 `interfaces/` 包
- 遵循Go官方代码风格

### TypeScript/React规范
- 使用TypeScript严格模式
- 组件使用函数式组件 + Hooks
- UI组件使用shadcn/ui
- API调用封装在 `lib/api.ts`

## 开发流程

### 1. 启动开发环境
```bash
# 启动测试API服务器
node test-api-server.js

# 在新终端启动前端开发服务器
cd publisher-web
npm run dev
```

### 2. 代码修改
- 前端代码: `publisher-web/src/`
- 后端代码: `publisher-core/`
- 测试API: `test-api-server.js`

### 3. 测试验证
- 前端: 浏览器访问开发服务器
- 后端: curl测试API端点
- 功能: 使用应用实际功能测试

## 常用开发命令

```bash
# 前端开发
cd publisher-web
npm run dev        # 启动开发服务器
npm run build      # 构建生产版本
npm run lint       # 代码检查

# 后端开发
cd publisher-core
go run cmd/server/main.go    # 运行服务
go test ./... -v            # 运行测试
go build -o bin/server cmd/server/main.go  # 编译

# 项目管理
make help          # 查看Makefile帮助
make build         # 编译项目
make dev           # 启动开发环境
make test          # 运行测试
```

## 调试技巧

### 前端调试
- 使用浏览器开发者工具
- React DevTools插件
- 查看网络请求和响应

### 后端调试
- 查看日志输出
- 使用Postman测试API
- 检查Chrome浏览器自动化过程

### 常见问题排查
1. **端口占用**: 使用 `netstat` 或 `lsof` 查看端口使用情况
2. **依赖问题**: 重新安装依赖 `npm install` 或 `go mod tidy`
3. **浏览器问题**: 确保Chrome/Chromium已安装且版本兼容

## 提交规范

### Git提交信息格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type类型
- feat: 新功能
- fix: 修复bug
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关
- chore: 构建过程或辅助工具变动

### 示例
```
feat(content): 添加一键发布功能

- 实现内容生成后直接发布到平台
- 添加发布状态反馈
- 优化用户交互体验

Closes #123
```

## 代码审查要点

### 前端审查
- [ ] TypeScript类型正确
- [ ] 组件可复用性
- [ ] 错误处理完善
- [ ] 用户体验流畅

### 后端审查
- [ ] 接口设计合理
- [ ] 错误处理完整
- [ ] 日志记录充分
- [ ] 性能考虑周全

## 版本发布流程

1. **功能开发完成**
   - 代码审查通过
   - 测试验证通过
   - 文档更新完成

2. **版本打包**
   ```bash
   make build
   ```

3. **版本发布**
   - 更新版本号
   - 创建Git标签
   - 推送到远程仓库

4. **部署上线**
   - Docker镜像构建
   - 服务器部署
   - 监控验证