# AI服务缓存系统使用指南

## 概述

AI服务缓存系统是一个高性能的内存缓存解决方案，专门为AI调用优化设计。通过智能缓存机制，减少重复AI调用，显著提升性能并降低成本。

## 核心功能

### 1. 智能缓存策略

- **基于输入Hash**: 使用SHA256哈希生成唯一缓存键
- **LRU驱逐策略**: 自动驱逐最少使用的缓存项
- **TTL过期机制**: 支持自定义过期时间
- **大小限制**: 支持缓存大小和项数限制

### 2. 缓存统计

- **命中率统计**: 实时追踪缓存命中率和未命中率
- **大小监控**: 监控缓存总大小和项数
- **驱逐统计**: 记录缓存驱逐次数
- **性能指标**: 追踪缓存性能指标

### 3. 缓存管理

- **手动清除**: 支持清空所有缓存或指定缓存
- **缓存预热**: 批量预加载缓存数据
- **键管理**: 查看和管理所有缓存键

## 配置说明

### 默认配置

```go
config := &cache.CacheConfig{
    MaxSize:         100 * 1024 * 1024, // 100MB
    MaxItems:        10000,              // 最多10000项
    DefaultTTL:      5 * time.Minute,    // 默认5分钟过期
    CleanupInterval: 1 * time.Minute,    // 每分钟清理一次
    EnableStats:     true,               // 启用统计
}
```

### 自定义配置

```go
config := &cache.CacheConfig{
    MaxSize:         200 * 1024 * 1024, // 200MB
    MaxItems:        20000,              // 最多20000项
    DefaultTTL:      10 * time.Minute,   // 默认10分钟过期
    CleanupInterval: 2 * time.Minute,    // 每2分钟清理一次
    EnableStats:     true,               // 启用统计
}

cacheService := cache.NewAICacheService(config)
```

## API使用示例

### 1. 缓存AI响应

#### 设置缓存

```go
// 创建AI缓存服务
cacheService := cache.NewAICacheService(nil)

// 缓存AI响应
response := &cache.AICacheValue{
    Response:   "这是AI生成的响应内容",
    TokensUsed: 150,
    Model:      "gpt-3.5-turbo",
    Provider:   "openai",
}

err := cacheService.SetCachedResponse(
    "openai",           // 提供商
    "gpt-3.5-turbo",    // 模型
    "请生成一篇关于AI的文章", // 提示词
    nil,                // 选项
    response,           // 响应
    10 * time.Minute,   // 过期时间
)
```

#### 获取缓存

```go
// 获取缓存的响应
cachedResponse, exists := cacheService.GetCachedResponse(
    "openai",
    "gpt-3.5-turbo",
    "请生成一篇关于AI的文章",
    nil,
)

if exists {
    fmt.Println("缓存命中:", cachedResponse.Response)
} else {
    fmt.Println("缓存未命中，需要调用AI")
}
```

### 2. 缓存提示词模板

```go
// 缓存模板
err := cacheService.SetCachedPromptTemplate(
    "content-generation-v1",
    "模板内容：{{topic}}",
    30 * time.Minute,
)

// 获取缓存的模板
template, exists := cacheService.GetCachedPromptTemplate("content-generation-v1")
if exists {
    fmt.Println("模板:", template)
}
```

### 3. 缓存热点分析

```go
// 缓存热点分析结果
analysis := map[string]interface{}{
    "keywords": []string{"AI", "机器学习"},
    "trend":    "上升",
    "score":    8.5,
}

err := cacheService.SetCachedHotspotAnalysis(
    "hotspot-123",
    analysis,
    15 * time.Minute,
)

// 获取缓存的分析
cachedAnalysis, exists := cacheService.GetCachedHotspotAnalysis("hotspot-123")
```

### 4. 缓存内容生成

```go
// 缓存生成的内容
err := cacheService.SetCachedContentGeneration(
    "人工智能发展趋势",
    "AI, 机器学习, 深度学习",
    "生成的内容...",
    20 * time.Minute,
)

// 获取缓存的内容
content, exists := cacheService.GetCachedContentGeneration(
    "人工智能发展趋势",
    "AI, 机器学习, 深度学习",
)
```

## REST API接口

### 1. 获取缓存统计

**端点**: `GET /api/v1/cache/stats`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "total_items": 1250,
    "total_size": 52428800,
    "total_hits": 8500,
    "total_misses": 1500,
    "hit_rate": 85.0,
    "evictions": 120,
    "last_clear_time": "2026-02-20T10:00:00Z"
  }
}
```

### 2. 获取所有缓存键

**端点**: `GET /api/v1/cache/keys`

**响应示例**:
```json
{
  "success": true,
  "data": {
    "keys": [
      "ai_response:abc123...",
      "prompt_template:content-generation-v1",
      "hotspot_analysis:hotspot-123"
    ],
    "total": 1250
  }
}
```

### 3. 清空缓存

**端点**: `POST /api/v1/cache/clear`

**响应示例**:
```json
{
  "success": true,
  "message": "缓存已清空"
}
```

### 4. 删除指定缓存

**端点**: `DELETE /api/v1/cache/{key}`

**响应示例**:
```json
{
  "success": true,
  "message": "缓存已删除"
}
```

### 5. 预热缓存

**端点**: `POST /api/v1/cache/warmup`

**请求示例**:
```json
{
  "items": {
    "prompt_template:content-generation-v1": "模板内容...",
    "hotspot_analysis:hotspot-123": {"keywords": ["AI"]},
    "ai_response:abc123": {"response": "AI响应..."}
  },
  "ttl": 600
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "缓存预热成功",
  "data": {
    "items_count": 3,
    "ttl_seconds": 600
  }
}
```

## 性能优化建议

### 1. 合理设置TTL

- **高频调用**: 设置较短的TTL（1-5分钟）
- **低频调用**: 设置较长的TTL（10-30分钟）
- **静态内容**: 设置更长的TTL（1-24小时）

### 2. 监控缓存命中率

```go
stats := cacheService.GetStats()
if stats.HitRate < 70.0 {
    fmt.Println("警告：缓存命中率过低，建议优化缓存策略")
}
```

### 3. 定期清理

系统会自动清理过期缓存，但也可以手动触发：

```go
// 清空所有缓存
cacheService.Clear()

// 删除特定缓存
cacheService.Delete("cache-key")

// 使模板缓存失效
cacheService.InvalidateTemplateCache("template-id")
```

### 4. 缓存预热

在系统启动或低峰期预加载常用数据：

```go
items := map[string]interface{}{
    "prompt_template:content-generation-v1": "模板内容",
    "hotspot_analysis:popular-topic": analysisData,
}

err := cacheService.Warmup(items, 30 * time.Minute)
```

## 最佳实践

### 1. 缓存键设计

使用有意义的缓存键前缀：
- `ai_response:` - AI响应缓存
- `prompt_template:` - 提示词模板缓存
- `hotspot_analysis:` - 热点分析缓存
- `content_generation:` - 内容生成缓存

### 2. 错误处理

```go
cachedResponse, exists := cacheService.GetCachedResponse(...)
if !exists {
    // 缓存未命中，调用AI服务
    response, err := aiService.Call(...)
    if err != nil {
        return err
    }

    // 缓存结果
    cacheService.SetCachedResponse(..., response, ttl)
    return response
}

// 使用缓存结果
return cachedResponse, nil
```

### 3. 监控和告警

```go
// 定期检查缓存状态
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := cacheService.GetStats()
        if stats.HitRate < 80.0 {
            log.Printf("缓存命中率警告: %.2f%%", stats.HitRate)
        }
        if stats.TotalSize > 80*1024*1024 { // 80MB
            log.Printf("缓存大小警告: %d bytes", stats.TotalSize)
        }
    }
}()
```

## 性能指标

### 典型性能数据

- **缓存命中**: < 1ms
- **缓存未命中**: < 5ms（包括键生成）
- **内存占用**: 约100MB（10000项）
- **并发支持**: 支持高并发读写

### 成本节省

假设：
- AI调用成本：$0.002/次
- 缓存命中率：85%
- 每日调用：100,000次

**每日节省**: 100,000 × 85% × $0.002 = **$170**
**每月节省**: $170 × 30 = **$5,100**

## 故障排查

### 1. 缓存命中率低

**可能原因**:
- TTL设置过短
- 缓存大小不足
- 输入参数变化频繁

**解决方案**:
- 增加TTL时间
- 增加缓存大小限制
- 优化输入参数标准化

### 2. 内存占用过高

**可能原因**:
- 缓存项过多
- 单个缓存项过大

**解决方案**:
- 减少MaxItems限制
- 减少MaxSize限制
- 定期清理缓存

### 3. 缓存失效

**可能原因**:
- 系统重启
- 手动清除
- TTL过期

**解决方案**:
- 实现缓存持久化
- 使用缓存预热
- 监控缓存状态

## 总结

AI服务缓存系统通过智能缓存策略，显著提升AI调用性能并降低成本。结合合理的配置和最佳实践，可以实现高达85%以上的缓存命中率，大幅减少AI服务调用次数和费用。

---

**文档版本**: v1.0  
**最后更新**: 2026-02-20  
**维护者**: MonkeyCode Team
