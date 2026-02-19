# 开发指南

## 概述

本文档为Publisher Tools项目的开发人员提供完整的开发环境搭建、编码规范和工作流程指导。

## 目录

- [环境搭建](#环境搭建)
- [编码规范](#编码规范)
- [开发流程](#开发流程)
- [调试技巧](#调试技巧)
- [测试指南](#测试指南)
- [提交规范](#提交规范)
- [相关文档](#相关文档)

## 环境搭建

### 必需软件

- **Go**: 1.21+
- **Node.js**: 18+
- **浏览器**: Chrome/Chromium（用于浏览器自动化）
- **Git**: 版本控制
- **编辑器**: VS Code推荐（安装Go和TypeScript扩展）

### 项目克隆

```bash
git clone <repository-url>
cd publisher-tools
```

### 依赖安装

```bash
# 安装所有依赖
make deps

# 或分别安装
cd publisher-core && go mod tidy
cd publisher-web && npm install
```

### 环境变量配置

创建 `.env` 文件：

```bash
# 服务配置
PORT=8080
HEADLESS=true
DEBUG=false

# AI配置（可选）
OPENROUTER_API_KEY=your_key
DEEPSEEK_API_KEY=your_key

# 存储配置
STORAGE_DIR=./uploads
DATA_DIR=./data
COOKIE_DIR=./cookies
```

## 编码规范

### Go代码规范

#### 日志记录
使用 `logrus` 进行日志记录：

```go
import "github.com/sirupsen/logrus"

logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
logger.WithField("component", "adapter").Info("登录成功")
```

#### 错误处理
使用 `github.com/pkg/errors` 包装错误：

```go
import "github.com/pkg/errors"

func doSomething() error {
    if err := someOperation(); err != nil {
        return errors.Wrap(err, "执行操作失败")
    }
    return nil
}
```

#### 接口定义
接口统一定义在 `interfaces/` 包中：

```go
// interfaces/publisher.go
type Publisher interface {
    Platform() string
    Publish(ctx context.Context, content *Content) error
    // ...
}
```

#### 代码风格
- 遵循Go官方代码风格
- 使用 `gofmt` 格式化代码
- 函数名使用驼峰命名法
- 变量名要有意义，避免单字母命名

### TypeScript/React规范

#### 类型安全
- 启用TypeScript严格模式
- 为所有props和state定义接口
- 避免使用any类型

```typescript
interface Props {
  title: string;
  onPublish: (content: Content) => void;
}

const PublishForm: React.FC<Props> = ({ title, onPublish }) => {
  // ...
};
```

#### 组件设计
- 优先使用函数式组件 + Hooks
- 组件职责单一，保持简洁
- 合理拆分大型组件

```typescript
// 好的做法
const usePublishLogic = () => {
  const [loading, setLoading] = useState(false);
  const publish = useCallback(async (content: Content) => {
    // 发布逻辑
  }, []);
  return { loading, publish };
};

const PublishForm: React.FC = () => {
  const { loading, publish } = usePublishLogic();
  // JSX渲染
};
```

#### UI组件
使用 shadcn/ui 组件库：

```typescript
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

<Button variant="primary" onClick={handleClick}>
  发布内容
</Button>
```

#### API调用
封装在 `lib/api.ts` 中：

```typescript
// lib/api.ts
export const api = {
  publish: (data: PublishRequest) => 
    axios.post<PublishResponse>('/api/v1/publish', data),
  getTasks: () => 
    axios.get<Task[]>('/api/v1/tasks'),
};
```

## 开发流程

### 1. 启动开发环境

```bash
# 启动完整开发环境
make dev

# 或分别启动
make serve      # 启动后端服务
make serve-web  # 启动前端开发服务器
```

### 2. 代码修改

#### 前端开发
```bash
cd publisher-web
npm run dev  # 热重载开发服务器
```

#### 后端开发
```bash
cd publisher-core
go run cmd/server/main.go
```

### 3. 测试验证

#### 前端测试
- 浏览器访问 http://localhost:5173
- 使用React DevTools检查组件状态
- 查看网络请求和响应

#### 后端测试
```bash
# API测试
curl http://localhost:8080/api/v1/platforms

# 单元测试
cd publisher-core
go test ./... -v
```

## 调试技巧

### 前端调试

#### 浏览器开发者工具
- **Elements**: 检查DOM结构和样式
- **Console**: 查看JavaScript错误和日志
- **Network**: 监控API请求和响应
- **React**: 使用React DevTools插件

#### 常用调试代码
```typescript
// 组件渲染调试
useEffect(() => {
  console.log('组件重新渲染', { props, state });
}, [props, state]);

// API调用调试
const debugApiCall = async (url: string, data: any) => {
  console.log('API请求:', url, data);
  const response = await api.call(url, data);
  console.log('API响应:', response);
  return response;
};
```

### 后端调试

#### 日志级别控制
```go
// 设置不同级别的日志
logger.SetLevel(logrus.DebugLevel)  // 开发时使用
logger.SetLevel(logrus.InfoLevel)   // 生产时使用
```

#### 浏览器自动化调试
```go
// 开发时禁用无头模式
browser := rod.New().MustConnect()
page := browser.MustPage("")
page.MustWindowFullscreen()  // 便于观察自动化过程
```

#### 常见问题排查
1. **端口占用**: 
   ```bash
   netstat -ano | findstr :8080
   taskkill /PID <进程ID> /F
   ```

2. **依赖问题**:
   ```bash
   go mod tidy
   npm install --force
   ```

3. **浏览器问题**:
   ```bash
   # 检查Chrome版本
   chrome --version
   
   # 安装chromedriver（如果需要）
   npm install chromedriver --save-dev
   ```

## 测试指南

### 单元测试

#### Go单元测试
```go
// example_test.go
func TestPublishContent(t *testing.T) {
    // 准备测试数据
    content := &Content{
        Title: "测试标题",
        Body:  "测试内容",
    }
    
    // 执行测试
    result, err := publisher.Publish(context.Background(), content)
    
    // 断言结果
    assert.NoError(t, err)
    assert.Equal(t, PublishStatusSuccess, result.Status)
}
```

#### React组件测试
```typescript
// PublishForm.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import PublishForm from './PublishForm';

test('should submit form with valid data', async () => {
  const mockPublish = jest.fn();
  render(<PublishForm onPublish={mockPublish} />);
  
  fireEvent.change(screen.getByLabelText('标题'), {
    target: { value: '测试标题' }
  });
  
  fireEvent.click(screen.getByText('发布'));
  
  expect(mockPublish).toHaveBeenCalledWith({
    title: '测试标题',
    // ...
  });
});
```

### 集成测试

#### API集成测试
```bash
# 使用curl测试API端点
curl -X POST http://localhost:8080/api/v1/publish \
  -H "Content-Type: application/json" \
  -d '{"platform":"douyin","title":"测试","content":"内容"}'
```

#### 端到端测试
```typescript
// playwright测试示例
import { test, expect } from '@playwright/test';

test('should publish content successfully', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.fill('[name=title]', '测试标题');
  await page.fill('[name=content]', '测试内容');
  await page.click('button[type=submit]');
  
  await expect(page.locator('.success-message')).toBeVisible();
});
```

## 提交规范

### Git提交信息格式
```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type类型说明
- **feat**: 新功能
- **fix**: 修复bug
- **docs**: 文档更新
- **style**: 代码格式调整（不影响功能）
- **refactor**: 代码重构
- **test**: 测试相关
- **chore**: 构建过程或辅助工具变动

### 示例提交信息
```
feat(content): 添加一键发布功能

- 实现内容生成后直接发布到平台
- 添加发布状态反馈
- 优化用户交互体验

Closes #123
```

```
fix(api): 修复异步发布状态查询问题

- 解决任务ID为空时的panic问题
- 添加输入参数验证
- 完善错误处理逻辑

Fixes #456
```

### 分支命名规范
- **feature/功能名称**: 新功能开发
- **fix/问题描述**: bug修复
- **hotfix/紧急修复**: 紧急修复
- **release/v版本号**: 发布准备

## 相关文档

- [架构文档](../architecture/)
- [API参考](../api/)
- [模块文档](../modules/)
- [部署指南](../guides/deployment/)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0