# API接口文档

## 概述

本文档详细描述Publisher Tools提供的所有REST API接口，包括平台管理、内容发布、任务管理、数据分析等核心功能。

## 目录

- [基础信息](#基础信息)
- [认证机制](#认证机制)
- [平台管理API](#平台管理api)
- [内容发布API](#内容发布api)
- [任务管理API](#任务管理api)
- [文件存储API](#文件存储api)
- [热点监控API](#热点监控api)
- [数据分析API](#数据分析api)
- [AI服务API](#ai服务api)
- [错误处理](#错误处理)

## 基础信息

### 服务端点
- **开发环境**: `http://localhost:8080`
- **生产环境**: `https://your-domain.com`

### 响应格式
所有API响应均采用JSON格式：

```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2026-02-19T10:00:00Z"
}
```

### 分页参数
对于列表接口，支持以下分页参数：
- `page`: 页码（默认1）
- `limit`: 每页数量（默认20，最大100）
- `sort`: 排序字段
- `order`: 排序方向（asc/desc）

## 认证机制

### Cookie认证
大部分接口通过Cookie进行认证，Cookie由各平台登录后自动设置。

### 请求头
```http
Authorization: Bearer <token>
Content-Type: application/json
Accept: application/json
```

## 平台管理API

### 获取平台列表
```http
GET /api/v1/platforms
```

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "name": "douyin",
      "display_name": "抖音",
      "supported_types": ["images", "video"],
      "login_required": true,
      "status": "available"
    },
    {
      "name": "toutiao",
      "display_name": "今日头条",
      "supported_types": ["images", "article"],
      "login_required": true,
      "status": "available"
    }
  ]
}
```

### 检查登录状态
```http
GET /api/v1/platforms/{platform}/check
```

**路径参数**:
- `platform`: 平台标识（douyin/toutiao/xiaohongshu）

**响应示例**:
```json
{
  "success": true,
  "data": {
    "logged_in": true,
    "username": "user123",
    "expires_at": "2026-02-20T10:00:00Z"
  }
}
```

### 平台登录
```http
POST /api/v1/platforms/{platform}/login
```

**请求体**:
```json
{
  "headless": false
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "qrcode_url": "https://qr.douyin.com/xxx",
    "login_timeout": 300,
    "task_id": "login-task-123"
  }
}
```

### 等待登录完成
```http
GET /api/v1/platforms/{platform}/wait-login
```

**查询参数**:
- `task_id`: 登录任务ID

## 内容发布API

### 同步发布
```http
POST /api/v1/publish
```

**请求体**:
```json
{
  "platform": "douyin",
  "type": "images",
  "title": "今日分享",
  "content": "美好的一天从这里开始",
  "images": [
    "uploads/photo1.jpg",
    "uploads/photo2.jpg"
  ],
  "tags": ["生活", "日常"],
  "scheduled_time": "2026-02-19T15:00:00Z"
}
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "status": "published",
    "post_id": "7890123456",
    "url": "https://www.douyin.com/video/7890123456",
    "created_at": "2026-02-19T10:00:00Z"
  }
}
```

### 异步发布
```http
POST /api/v1/publish/async
```

**请求体**: 同同步发布

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": "publish-task-456",
    "status": "pending",
    "estimated_time": 120
  }
}
```

### 发布内容类型说明

#### 图文发布
```json
{
  "type": "images",
  "title": "标题（最多30字）",
  "content": "正文内容（最多2000字）",
  "images": ["图片路径1", "图片路径2"],
  "max_images": 12
}
```

#### 视频发布
```json
{
  "type": "video",
  "title": "标题（最多30字）",
  "content": "视频描述（最多2000字）",
  "video": "uploads/video.mp4",
  "cover": "uploads/cover.jpg",
  "max_size": "4GB"
}
```

#### 文章发布（今日头条）
```json
{
  "type": "article",
  "title": "文章标题（最多30字）",
  "content": "文章正文（支持HTML）",
  "category": "科技",
  "tags": ["AI", "技术"]
}
```

## 任务管理API

### 创建任务
```http
POST /api/v1/tasks
```

**请求体**:
```json
{
  "type": "publish",
  "platform": "douyin",
  "payload": {
    "title": "任务标题",
    "content": "任务内容"
  },
  "priority": "normal",
  "scheduled_time": "2026-02-19T15:00:00Z"
}
```

### 获取任务列表
```http
GET /api/v1/tasks
```

**查询参数**:
- `status`: 任务状态（pending/running/completed/failed）
- `type`: 任务类型
- `platform`: 平台筛选

**响应示例**:
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": "task-123",
        "type": "publish",
        "platform": "douyin",
        "status": "completed",
        "created_at": "2026-02-19T09:00:00Z",
        "updated_at": "2026-02-19T09:05:00Z",
        "result": {
          "post_id": "7890123456",
          "url": "https://www.douyin.com/video/7890123456"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 45,
      "pages": 3
    }
  }
}
```

### 获取任务详情
```http
GET /api/v1/tasks/{taskId}
```

### 取消任务
```http
POST /api/v1/tasks/{taskId}/cancel
```

### 重试任务
```http
POST /api/v1/tasks/{taskId}/retry
```

## 文件存储API

### 上传文件
```http
POST /api/v1/storage/upload
Content-Type: multipart/form-data
```

**表单字段**:
- `file`: 文件数据
- `type`: 文件类型（image/video/audio/document）
- `platform`: 目标平台（可选）

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": "file-789",
    "path": "uploads/images/20260219/photo.jpg",
    "url": "http://localhost:8080/uploads/images/20260219/photo.jpg",
    "size": 1024000,
    "mime_type": "image/jpeg"
  }
}
```

### 获取文件列表
```http
GET /api/v1/storage/list
```

**查询参数**:
- `type`: 文件类型筛选
- `platform`: 平台筛选
- `date_from`: 开始日期
- `date_to`: 结束日期

### 下载文件
```http
GET /api/v1/storage/download/{fileId}
```

### 删除文件
```http
DELETE /api/v1/storage/delete/{fileId}
```

## 热点监控API

### 获取热点列表
```http
GET /api/hot-topics
```

**查询参数**:
- `source`: 数据源（newsnow/baidu/weibo/zhihu）
- `category`: 分类筛选
- `limit`: 返回数量（默认20）

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "id": "hot-123",
      "title": "AI技术最新发展趋势",
      "description": "人工智能领域的重要突破...",
      "source": "newsnow",
      "category": "科技",
      "hot_score": 95,
      "keywords": ["AI", "技术", "发展"],
      "suitability": {
        "douyin": 0.85,
        "toutiao": 0.92,
        "xiaohongshu": 0.78
      },
      "created_at": "2026-02-19T08:00:00Z"
    }
  ]
}
```

### 抓取热点数据
```http
POST /api/hot-topics/newsnow/fetch
```

**请求体**:
```json
{
  "sources": ["newsnow", "baidu"],
  "force_refresh": true
}
```

### 获取数据源列表
```http
GET /api/hot-topics/newsnow/sources
```

## 数据分析API

### 获取仪表盘数据
```http
GET /api/analytics/dashboard
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "summary": {
      "total_posts": 1250,
      "today_posts": 15,
      "total_views": 1250000,
      "engagement_rate": 8.5
    },
    "platform_stats": {
      "douyin": {
        "posts": 500,
        "views": 600000,
        "likes": 45000
      },
      "toutiao": {
        "posts": 400,
        "views": 400000,
        "comments": 12000
      }
    },
    "recent_trends": [...]
  }
}
```

### 获取趋势数据
```http
GET /api/analytics/trends
```

**查询参数**:
- `period`: 时间周期（day/week/month）
- `metrics`: 指标类型（views/likes/comments）

### 生成周报
```http
GET /api/analytics/report/weekly
```

**查询参数**:
- `start_date`: 开始日期
- `end_date`: 结束日期
- `format`: 输出格式（json/markdown）

### 导出报告
```http
GET /api/analytics/report/export
```

**查询参数**:
- `type`: 报告类型（weekly/monthly/custom）
- `format`: 导出格式（json/markdown/csv/pdf）
- `include_charts`: 是否包含图表

## AI服务API

### 获取AI提供商列表
```http
GET /api/v1/ai/providers
```

### 获取模型列表
```http
GET /api/v1/ai/models
```

### AI内容生成
```http
POST /api/v1/ai/content/generate
```

**请求体**:
```json
{
  "provider": "openrouter",
  "model": "gpt-4",
  "template": "social_media_post",
  "parameters": {
    "topic": "科技发展",
    "tone": "专业",
    "length": "medium",
    "platform": "douyin"
  }
}
```

### AI内容改写
```http
POST /api/v1/ai/content/rewrite
```

**请求体**:
```json
{
  "content": "原始内容文本",
  "instructions": "使语气更轻松幽默",
  "target_platform": "xiaohongshu"
}
```

### 热点分析
```http
POST /api/v1/ai/analyze/hotspot
```

## 错误处理

### 错误响应格式
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "标题不能为空",
    "details": {
      "field": "title",
      "value": ""
    }
  },
  "timestamp": "2026-02-19T10:00:00Z"
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| VALIDATION_ERROR | 400 | 参数验证失败 |
| AUTHENTICATION_FAILED | 401 | 认证失败 |
| PERMISSION_DENIED | 403 | 权限不足 |
| RESOURCE_NOT_FOUND | 404 | 资源不存在 |
| PLATFORM_LOGIN_REQUIRED | 401 | 需要平台登录 |
| TASK_EXECUTION_FAILED | 500 | 任务执行失败 |
| RATE_LIMIT_EXCEEDED | 429 | 请求频率超限 |

### 重试策略
```javascript
// 客户端重试逻辑示例
async function apiCallWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, options);
      if (response.ok) return response.json();
      
      if (response.status === 429) {
        // 速率限制，等待后重试
        const retryAfter = response.headers.get('Retry-After') || (2 ** i) * 1000;
        await new Promise(resolve => setTimeout(resolve, retryAfter));
        continue;
      }
      
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}
```

## 相关文档

- [开发指南](../development/)
- [平台适配器文档](../modules/adapters/)
- [任务管理文档](../modules/task/)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0