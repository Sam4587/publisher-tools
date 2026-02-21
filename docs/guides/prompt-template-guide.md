# AI 提示词模板系统使用指南

## 概述

AI 提示词模板系统是一个强大的提示词管理工具，支持模板创建、版本控制、变量替换和A/B测试功能。该系统帮助用户标准化AI调用流程，提升内容生成质量。

## 核心功能

### 1. 模板管理

#### 创建模板

**API端点**: `POST /api/v1/prompt-templates`

**请求示例**:
```json
{
  "template_id": "my-content-template",
  "name": "我的内容生成模板",
  "type": "content_generation",
  "category": "content",
  "description": "用于生成高质量内容的模板",
  "content": "请根据主题：{{topic}}，关键词：{{keywords}}，生成一篇{{word_count}}字的内容。",
  "variables": [
    {
      "name": "topic",
      "type": "string",
      "required": true,
      "description": "内容主题"
    },
    {
      "name": "keywords",
      "type": "string",
      "required": true,
      "description": "关键词列表"
    },
    {
      "name": "word_count",
      "type": "number",
      "required": false,
      "default": "500",
      "description": "字数要求"
    }
  ],
  "is_default": false,
  "tags": ["内容生成", "自定义模板"]
}
```

**响应示例**:
```json
{
  "id": 1,
  "template_id": "my-content-template",
  "name": "我的内容生成模板",
  "type": "content_generation",
  "category": "content",
  "description": "用于生成高质量内容的模板",
  "content": "请根据主题：{{topic}}，关键词：{{keywords}}，生成一篇{{word_count}}字的内容。",
  "variables": "[{\"name\":\"topic\",\"type\":\"string\",\"required\":true,\"description\":\"内容主题\"},{\"name\":\"keywords\",\"type\":\"string\",\"required\":true,\"description\":\"关键词列表\"},{\"name\":\"word_count\",\"type\":\"number\",\"required\":false,\"default\":\"500\",\"description\":\"字数要求\"}]",
  "version": 1,
  "is_active": true,
  "is_default": false,
  "is_system": false,
  "tags": "[\"内容生成\",\"自定义模板\"]",
  "created_at": "2026-02-20T10:00:00Z",
  "updated_at": "2026-02-20T10:00:00Z"
}
```

#### 获取模板

**API端点**: `GET /api/v1/prompt-templates/{templateId}`

**响应示例**:
```json
{
  "id": 1,
  "template_id": "my-content-template",
  "name": "我的内容生成模板",
  "type": "content_generation",
  "content": "请根据主题：{{topic}}，关键词：{{keywords}}，生成一篇{{word_count}}字的内容。",
  "version": 1,
  "is_active": true
}
```

#### 列出模板

**API端点**: `GET /api/v1/prompt-templates`

**查询参数**:
- `type`: 模板类型（可选）
- `category`: 分类（可选）
- `is_active`: 是否激活（可选）
- `page`: 页码（默认1）
- `page_size`: 每页数量（默认20）

**响应示例**:
```json
{
  "templates": [
    {
      "id": 1,
      "template_id": "content-generation-v1",
      "name": "内容生成模板",
      "type": "content_generation",
      "is_active": true
    }
  ],
  "total": 8,
  "page": 1,
  "page_size": 20
}
```

#### 更新模板

**API端点**: `PUT /api/v1/prompt-templates/{templateId}`

**请求示例**:
```json
{
  "name": "更新后的模板名称",
  "content": "更新后的模板内容：{{topic}}",
  "change_note": "优化模板内容"
}
```

#### 删除模板

**API端点**: `DELETE /api/v1/prompt-templates/{templateId}`

**注意**: 系统模板（`is_system=true`）不可删除。

### 2. 模板渲染

#### 渲染模板

**API端点**: `POST /api/v1/prompt-templates/{templateId}/render`

**请求示例**:
```json
{
  "variables": {
    "topic": "人工智能发展趋势",
    "keywords": "AI, 机器学习, 深度学习",
    "word_count": 800
  }
}
```

**响应示例**:
```json
{
  "content": "请根据主题：人工智能发展趋势，关键词：AI, 机器学习, 深度学习，生成一篇800字的内容。"
}
```

### 3. 版本控制

#### 获取版本历史

**API端点**: `GET /api/v1/prompt-templates/{templateId}/versions`

**响应示例**:
```json
{
  "versions": [
    {
      "id": 1,
      "template_id": "my-content-template",
      "version": 1,
      "content": "初始版本内容",
      "change_note": "初始版本",
      "created_at": "2026-02-20T10:00:00Z"
    },
    {
      "id": 2,
      "template_id": "my-content-template",
      "version": 2,
      "content": "更新后的内容",
      "change_note": "优化模板内容",
      "created_at": "2026-02-20T11:00:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 20
}
```

#### 恢复到指定版本

**API端点**: `POST /api/v1/prompt-templates/{templateId}/versions/{version}/restore`

**响应示例**:
```json
{
  "id": 1,
  "template_id": "my-content-template",
  "version": 3,
  "content": "初始版本内容",
  "change_note": "恢复到版本 1"
}
```

### 4. A/B测试

#### 创建A/B测试

**API端点**: `POST /api/v1/prompt-ab-tests`

**请求示例**:
```json
{
  "test_name": "内容生成模板对比测试",
  "template_a_id": "content-generation-v1",
  "template_b_id": "content-generation-v2",
  "traffic_split": 50
}
```

**响应示例**:
```json
{
  "id": 1,
  "test_name": "内容生成模板对比测试",
  "template_a_id": "content-generation-v1",
  "template_b_id": "content-generation-v2",
  "status": "running",
  "traffic_split": 50,
  "total_calls": 0,
  "calls_a": 0,
  "calls_b": 0,
  "start_time": "2026-02-20T10:00:00Z"
}
```

#### 获取A/B测试统计

**API端点**: `GET /api/v1/prompt-ab-tests/{testId}/stats`

**响应示例**:
```json
{
  "test_id": 1,
  "test_name": "内容生成模板对比测试",
  "status": "running",
  "total_calls": 100,
  "template_a": {
    "template_id": "content-generation-v1",
    "calls": 52,
    "success": 50,
    "success_rate": 96.15,
    "avg_duration": 1250.5,
    "user_rating": 4.2
  },
  "template_b": {
    "template_id": "content-generation-v2",
    "calls": 48,
    "success": 46,
    "success_rate": 95.83,
    "avg_duration": 1180.3,
    "user_rating": 4.5
  },
  "traffic_split": 50
}
```

#### 完成A/B测试

**API端点**: `POST /api/v1/prompt-ab-tests/{testId}/complete`

**请求示例**:
```json
{
  "winner_template_id": "content-generation-v2"
}
```

## 预定义模板

系统提供以下预定义模板：

### 1. 内容生成模板 (content-generation-v1)
- **用途**: 基于主题和关键词生成高质量内容
- **变量**: topic, keywords, platform, style, word_count
- **标签**: 内容生成, AI创作, 默认模板

### 2. 内容改写模板 (content-rewrite-v1)
- **用途**: 改写现有内容，保持原意的同时提升质量
- **变量**: original_content, rewrite_style, platform
- **标签**: 内容改写, AI优化, 默认模板

### 3. 热点分析模板 (hotspot-analysis-v1)
- **用途**: 分析热点话题，提取关键信息和趋势
- **变量**: hotspot_title, hotspot_description, heat_index, source_platform
- **标签**: 热点分析, 趋势预测, 默认模板

### 4. 视频转录优化模板 (video-transcription-v1)
- **用途**: 优化视频转录文本，提升可读性
- **变量**: transcription, video_title, duration
- **标签**: 视频转录, 文本优化, 默认模板

### 5. 标题生成模板 (title-generation-v1)
- **用途**: 生成吸引人的标题
- **变量**: content_summary, platform, target_audience, title_length, num_titles
- **标签**: 标题生成, AI创作, 默认模板

### 6. 内容审核模板 (content-review-v1)
- **用途**: 审核内容是否符合平台规范
- **变量**: content, platform
- **标签**: 内容审核, 合规检查, 默认模板

### 7. 关键词提取模板 (keyword-extraction-v1)
- **用途**: 从内容中提取关键词
- **变量**: content, num_keywords
- **标签**: 关键词提取, SEO优化, 默认模板

### 8. 内容摘要模板 (content-summary-v1)
- **用途**: 生成内容摘要
- **变量**: content, summary_length
- **标签**: 内容摘要, AI总结, 默认模板

## 最佳实践

### 1. 模板设计原则

- **变量命名清晰**: 使用有意义的变量名，如`topic`、`keywords`
- **提供默认值**: 为非必需变量提供合理的默认值
- **详细描述**: 为每个变量提供清晰的描述
- **版本控制**: 每次重要修改都添加变更说明

### 2. A/B测试建议

- **明确测试目标**: 确定要测试的指标（成功率、响应时间、用户评分）
- **合理分配流量**: 根据测试重要性设置流量分配
- **收集足够数据**: 确保每个模板有足够的调用次数
- **及时分析结果**: 定期查看统计数据，及时调整策略

### 3. 性能优化

- **缓存常用模板**: 对高频使用的模板进行缓存
- **批量渲染**: 支持批量模板渲染，减少API调用
- **异步处理**: 对于耗时操作使用异步处理

## 错误处理

### 常见错误码

- `400 Bad Request`: 请求参数无效
- `404 Not Found`: 模板或测试不存在
- `500 Internal Server Error`: 服务器内部错误

### 错误响应示例

```json
{
  "error": "模板不存在"
}
```

## 集成示例

### Python示例

```python
import requests

# 创建模板
response = requests.post('http://localhost:8080/api/v1/prompt-templates', json={
    'template_id': 'my-template',
    'name': '我的模板',
    'type': 'content_generation',
    'content': '主题：{{topic}}',
    'variables': [
        {'name': 'topic', 'type': 'string', 'required': True}
    ]
})

# 渲染模板
response = requests.post('http://localhost:8080/api/v1/prompt-templates/my-template/render', json={
    'variables': {'topic': '人工智能'}
})
print(response.json()['content'])
```

### JavaScript示例

```javascript
// 创建模板
const response = await fetch('http://localhost:8080/api/v1/prompt-templates', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    template_id: 'my-template',
    name: '我的模板',
    type: 'content_generation',
    content: '主题：{{topic}}',
    variables: [
      {name: 'topic', type: 'string', required: true}
    ]
  })
});

// 渲染模板
const renderResponse = await fetch('http://localhost:8080/api/v1/prompt-templates/my-template/render', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    variables: {topic: '人工智能'}
  })
});
const data = await renderResponse.json();
console.log(data.content);
```

## 总结

AI 提示词模板系统提供了完整的模板管理解决方案，通过标准化提示词管理、版本控制和A/B测试功能，帮助用户提升AI内容生成的效率和质量。结合预定义模板和自定义模板，可以满足各种内容生成场景的需求。

---

**文档版本**: v1.0  
**最后更新**: 2026-02-20  
**维护者**: MonkeyCode Team
