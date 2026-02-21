# AI提供商扩展系统使用指南

## 概述

AI提供商扩展系统支持多个AI提供商的集成、智能选择和健康监控。通过统一的接口和智能降级策略，确保AI服务的高可用性和最优性能。

## 新增提供商

### 1. NVIDIA AI提供商

NVIDIA提供多个高性能AI模型，包括Llama、Mistral等。

#### 支持的模型

- `meta/llama-3.1-8b-instruct` - Llama 3.1 8B（默认）
- `meta/llama-3.1-70b-instruct` - Llama 3.1 70B
- `meta/llama-3.1-405b-instruct` - Llama 3.1 405B
- `nvidia/nemotron-4-340b-instruct` - Nemotron 4 340B
- `mistralai/mistral-large` - Mistral Large
- `google/gemma-7b` - Gemma 7B

#### 使用示例

```go
import "publisher-tools/ai/provider"

// 创建NVIDIA提供商
nvidiaProvider := provider.NewNVIDIAProvider(&provider.NVIDIAConfig{
    APIKey: "your-nvidia-api-key",
    Model:  "meta/llama-3.1-8b-instruct",
})

// 生成内容
result, err := nvidiaProvider.Generate(ctx, &provider.GenerateOptions{
    Model: "meta/llama-3.1-8b-instruct",
    Messages: []provider.Message{
        {Role: provider.RoleUser, Content: "请生成一篇关于AI的文章"},
    },
    MaxTokens: 500,
})
```

### 2. Mistral AI提供商

Mistral提供高性能的开源和商业AI模型。

#### 支持的模型

- `mistral-large-latest` - Mistral Large（默认）
- `mistral-medium-latest` - Mistral Medium
- `mistral-small-latest` - Mistral Small
- `open-mistral-7b` - 开源 Mistral 7B
- `open-mixtral-8x7b` - 开源 Mixtral 8x7B
- `open-mixtral-8x22b` - 开源 Mixtral 8x22B
- `codestral-latest` - Codestral（代码专用）

#### 使用示例

```go
// 创建Mistral提供商
mistralProvider := provider.NewMistralProvider(&provider.MistralConfig{
    APIKey: "your-mistral-api-key",
    Model:  "mistral-large-latest",
})

// 生成内容
result, err := mistralProvider.Generate(ctx, &provider.GenerateOptions{
    Model: "mistral-large-latest",
    Messages: []provider.Message{
        {Role: provider.RoleUser, Content: "请生成一篇关于AI的文章"},
    },
    MaxTokens: 500,
})

// 流式生成
stream, err := mistralProvider.GenerateStream(ctx, &provider.GenerateOptions{
    Model: "mistral-large-latest",
    Messages: []provider.Message{
        {Role: provider.RoleUser, Content: "请生成一篇关于AI的文章"},
    },
})

for chunk := range stream {
    fmt.Print(chunk)
}
```

## 多模型并行调用

### 功能特性

- **并行调用**: 同时调用多个AI模型
- **结果对比**: 自动对比多个模型的结果
- **质量评分**: 基于响应时间、内容质量等指标评分
- **最优选择**: 自动选择最佳结果

### 使用示例

```go
// 创建多模型服务
providers := map[provider.ProviderType]provider.Provider{
    provider.ProviderOpenRouter: openrouterProvider,
    provider.ProviderGoogle:     googleProvider,
    "nvidia":                    nvidiaProvider,
    "mistral":                   mistralProvider,
}

multiService := provider.NewMultiModelService(providers)

// 并行调用多个模型
results, err := multiService.ParallelCall(ctx, &provider.GenerateOptions{
    Messages: []provider.Message{
        {Role: provider.RoleUser, Content: "请生成一篇关于AI的文章"},
    },
    MaxTokens: 500,
}, []provider.ProviderType{
    provider.ProviderOpenRouter,
    "nvidia",
    "mistral",
})

// 对比结果，选择最佳
bestResult := multiService.CompareResults(results)
fmt.Println("最佳结果:", bestResult.Result.Content)
fmt.Println("提供商:", bestResult.Provider)
fmt.Println("质量评分:", bestResult.QualityScore)
```

### 并行调用结果

```json
[
  {
    "provider": "openrouter",
    "model": "gpt-3.5-turbo",
    "result": {
      "content": "生成的内容...",
      "input_tokens": 50,
      "output_tokens": 500
    },
    "duration": "2.5s",
    "quality_score": 95.5
  },
  {
    "provider": "nvidia",
    "model": "meta/llama-3.1-8b-instruct",
    "result": {
      "content": "生成的内容...",
      "input_tokens": 50,
      "output_tokens": 480
    },
    "duration": "1.8s",
    "quality_score": 92.3
  }
]
```

## 智能提供商选择

### 选择策略

系统基于以下指标智能选择最佳提供商：

1. **优先级**: 预设的提供商优先级
2. **成功率**: 历史调用成功率
3. **响应时间**: 平均响应时间
4. **质量评分**: 用户反馈的质量评分
5. **健康状态**: 实时健康检查结果

### 使用示例

```go
// 智能调用（自动选择最佳提供商）
result, err := multiService.SmartCall(ctx, &provider.GenerateOptions{
    Messages: []provider.Message{
        {Role: provider.RoleUser, Content: "请生成一篇关于AI的文章"},
    },
    MaxTokens: 500,
})

if err != nil {
    // 系统会自动降级到其他提供商
    fmt.Println("所有提供商都不可用")
    return
}

fmt.Println("生成结果:", result.Content)
fmt.Println("使用提供商:", result.Provider)
```

### 降级策略

当首选提供商失败时，系统自动降级：

1. 尝试优先级次高的提供商
2. 跳过最近失败的提供商
3. 确保至少有一个可用提供商

## 健康检查机制

### 功能特性

- **定期检查**: 每30秒检查所有提供商
- **自动标记**: 自动标记不健康的提供商
- **统计监控**: 实时监控成功率、响应时间
- **健康报告**: 生成详细的健康报告

### 使用示例

```go
// 获取健康状态
stats := multiService.GetProviderStats()

for pType, stat := range stats {
    fmt.Printf("提供商: %s\n", pType)
    fmt.Printf("  健康状态: %v\n", stat.IsHealthy)
    fmt.Printf("  成功率: %.2f%%\n", stat.SuccessRate)
    fmt.Printf("  平均响应时间: %v\n", stat.AvgDuration)
    fmt.Printf("  总调用次数: %d\n", stat.TotalCalls)
    fmt.Printf("  失败次数: %d\n", stat.FailedCalls)
}

// 获取健康报告
healthChecker := provider.NewHealthChecker(providers)
report := healthChecker.GenerateReport()

fmt.Printf("健康报告: %s\n", healthChecker.String())
fmt.Printf("健康提供商: %d\n", report.HealthyProviders)
fmt.Printf("不健康提供商: %d\n", report.UnhealthyProviders)
```

### 健康报告示例

```json
{
  "timestamp": "2026-02-20T10:00:00Z",
  "total_providers": 6,
  "healthy_providers": 5,
  "unhealthy_providers": 1,
  "provider_stats": {
    "openrouter": {
      "is_healthy": true,
      "success_rate": 98.5,
      "avg_duration": "2.1s",
      "total_calls": 1000,
      "failed_calls": 15
    },
    "nvidia": {
      "is_healthy": true,
      "success_rate": 95.2,
      "avg_duration": "1.8s",
      "total_calls": 500,
      "failed_calls": 24
    },
    "mistral": {
      "is_healthy": false,
      "success_rate": 45.0,
      "avg_duration": "5.2s",
      "total_calls": 100,
      "failed_calls": 55,
      "last_error": "API timeout"
    }
  }
}
```

## 配置管理

### 提供商配置

```go
type ProviderConfig struct {
    Type     ProviderType `json:"type"`
    APIKey   string       `json:"api_key"`
    BaseURL  string       `json:"base_url"`
    Model    string       `json:"model"`
    Enabled  bool         `json:"enabled"`
    Priority int          `json:"priority"`
}
```

### 配置示例

```json
[
  {
    "type": "openrouter",
    "api_key": "sk-or-...",
    "model": "gpt-3.5-turbo",
    "enabled": true,
    "priority": 5
  },
  {
    "type": "nvidia",
    "api_key": "nvapi-...",
    "model": "meta/llama-3.1-8b-instruct",
    "enabled": true,
    "priority": 4
  },
  {
    "type": "mistral",
    "api_key": "sk-...",
    "model": "mistral-large-latest",
    "enabled": true,
    "priority": 4
  }
]
```

## 性能优化

### 1. 缓存集成

```go
// 结合缓存服务使用
cacheService := cache.NewAICacheService(nil)

// 先检查缓存
cachedResponse, exists := cacheService.GetCachedResponse(provider, model, prompt, nil)
if exists {
    return cachedResponse, nil
}

// 缓存未命中，调用AI
result, err := multiService.SmartCall(ctx, opts)
if err == nil {
    // 缓存结果
    cacheService.SetCachedResponse(provider, model, prompt, nil, result, 10*time.Minute)
}
```

### 2. 超时控制

```go
// 设置超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := multiService.SmartCall(ctx, opts)
```

### 3. 并发控制

```go
// 限制并发调用数量
semaphore := make(chan struct{}, 5) // 最多5个并发

for _, prompt := range prompts {
    semaphore <- struct{}{}
    go func(p string) {
        defer func() { <-semaphore }()
        result, _ := multiService.SmartCall(ctx, &provider.GenerateOptions{
            Messages: []provider.Message{
                {Role: provider.RoleUser, Content: p},
            },
        })
    }(prompt)
}
```

## 最佳实践

### 1. 提供商选择

- **生产环境**: 使用多个提供商确保高可用
- **开发环境**: 使用免费提供商降低成本
- **关键任务**: 使用高质量提供商（如GPT-4）

### 2. 错误处理

```go
result, err := multiService.SmartCall(ctx, opts)
if err != nil {
    // 记录错误
    log.Printf("AI调用失败: %v", err)
    
    // 返回降级响应
    return &provider.GenerateResult{
        Content: "抱歉，AI服务暂时不可用",
    }, nil
}
```

### 3. 监控告警

```go
// 定期检查健康状态
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := multiService.GetProviderStats()
        for pType, stat := range stats {
            if !stat.IsHealthy {
                log.Printf("警告: 提供商 %s 不健康", pType)
            }
            if stat.SuccessRate < 80 {
                log.Printf("警告: 提供商 %s 成功率过低: %.2f%%", pType, stat.SuccessRate)
            }
        }
    }
}()
```

## 故障排查

### 1. 提供商不可用

**症状**: 所有提供商都返回错误

**排查步骤**:
1. 检查API密钥是否有效
2. 检查网络连接
3. 查看健康检查报告
4. 检查API配额是否用尽

### 2. 响应时间过长

**症状**: AI调用超时或响应缓慢

**解决方案**:
1. 减少MaxTokens参数
2. 使用更快的模型
3. 增加超时时间
4. 检查网络延迟

### 3. 成功率下降

**症状**: 提供商成功率突然下降

**排查步骤**:
1. 查看错误日志
2. 检查API状态页面
3. 验证请求参数
4. 联系提供商支持

## 总结

AI提供商扩展系统通过多提供商集成、智能选择和健康监控，确保AI服务的高可用性和最优性能。结合缓存和并发控制，可以显著提升系统性能并降低成本。

---

**文档版本**: v1.0  
**最后更新**: 2026-02-20  
**维护者**: MonkeyCode Team
