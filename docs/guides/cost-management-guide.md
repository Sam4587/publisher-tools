# AI成本追踪与预算控制系统使用指南

## 概述

AI成本追踪与预算控制系统提供完整的成本管理解决方案，包括成本记录、统计分析、预算控制和优化建议功能，帮助用户有效管理AI调用成本。

## 核心功能

### 1. 成本追踪

#### 自动记录成本

系统自动记录每次AI调用的成本信息：

```go
import "publisher-tools/cost"

// 创建成本服务
costService := cost.NewCostService(db)

// 记录AI调用成本
record, err := costService.RecordCost(&cost.CostRecordInput{
    Provider:     "openrouter",
    Model:        "gpt-3.5-turbo",
    UserID:       "user-123",
    ProjectID:    "project-456",
    FunctionType: "content_generation",
    InputTokens:  150,
    OutputTokens: 500,
    DurationMs:   2500,
    Success:      true,
    RequestID:    "req-789",
    Prompt:       "请生成一篇关于AI的文章",
    Cached:       false,
})
```

#### 成本计算

系统支持多种定价模型：

```go
// 计算成本
costPerToken, totalCost := costService.CalculateCost(
    "openrouter",
    "gpt-3.5-turbo",
    150,  // input tokens
    500,  // output tokens
)

fmt.Printf("每token成本: $%.6f\n", costPerToken)
fmt.Printf("总成本: $%.4f\n", totalCost)
```

### 2. 成本统计

#### 总体统计

```go
// 获取最近30天的成本统计
startTime := time.Now().AddDate(0, 0, -30)
endTime := time.Now()

stats, err := costService.GetCostStats("user-123", "project-456", startTime, endTime)

fmt.Printf("总调用次数: %d\n", stats.TotalCalls)
fmt.Printf("总成本: $%.2f\n", stats.TotalCost)
fmt.Printf("总Token数: %d\n", stats.TotalTokens)
fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate)
fmt.Printf("缓存命中率: %.2f%%\n", stats.CacheHitRate)
```

#### 按提供商统计

```go
providerStats, err := costService.GetCostByProvider(userID, projectID, startTime, endTime)

for _, stat := range providerStats {
    fmt.Printf("提供商: %s\n", stat.Provider)
    fmt.Printf("  总调用: %d\n", stat.TotalCalls)
    fmt.Printf("  总成本: $%.2f\n", stat.TotalCost)
    fmt.Printf("  平均响应时间: %.0fms\n", stat.AvgDurationMs)
}
```

#### 按模型统计

```go
modelStats, err := costService.GetCostByModel(userID, projectID, startTime, endTime)

for _, stat := range modelStats {
    fmt.Printf("模型: %s/%s\n", stat.Provider, stat.Model)
    fmt.Printf("  总调用: %d\n", stat.TotalCalls)
    fmt.Printf("  总成本: $%.2f\n", stat.TotalCost)
}
```

#### 每日成本趋势

```go
dailyStats, err := costService.GetDailyCost(userID, projectID, startTime, endTime)

for _, stat := range dailyStats {
    fmt.Printf("日期: %s, 成本: $%.2f, 调用: %d\n",
        stat.Date, stat.TotalCost, stat.TotalCalls)
}
```

### 3. 预算控制

#### 创建预算

```go
budgetService := cost.NewBudgetService(db, costService)

// 创建月度预算
budget, err := budgetService.CreateBudget(&cost.CreateBudgetRequest{
    UserID:         "user-123",
    ProjectID:      "project-456",
    BudgetType:     "monthly",
    BudgetAmount:   100.0,  // $100
    AlertThreshold: 80.0,   // 80%时预警
    StartDate:      time.Now(),
    EndDate:        time.Now().AddDate(0, 1, 0),
})
```

#### 检查预算状态

```go
status, err := budgetService.CheckBudget("user-123", "project-456")

if status.HasBudget {
    fmt.Printf("预算金额: $%.2f\n", status.BudgetAmount)
    fmt.Printf("已使用: $%.2f\n", status.UsedAmount)
    fmt.Printf("剩余: $%.2f\n", status.RemainingAmount)
    fmt.Printf("使用率: %.1f%%\n", status.UsagePercent)
    
    if status.IsExceeded {
        fmt.Println("警告：预算已超限！")
    } else if status.IsWarning {
        fmt.Println("警告：预算使用已达预警阈值！")
    }
}
```

#### 更新使用金额

```go
// 在AI调用后更新预算
err := budgetService.UpdateUsedAmount("user-123", "project-456", 0.05)
```

#### 管理预警

```go
// 获取预警列表
alerts, total, err := budgetService.GetAlerts("user-123", "project-456", nil, 1, 20)

for _, alert := range alerts {
    fmt.Printf("预警类型: %s\n", alert.AlertType)
    fmt.Printf("预警级别: %s\n", alert.AlertLevel)
    fmt.Printf("消息: %s\n", alert.Message)
    fmt.Printf("使用率: %.1f%%\n", alert.UsagePercent)
}

// 标记为已读
budgetService.MarkAlertAsRead(alertID)

// 解决预警
budgetService.ResolveAlert(alertID)
```

### 4. 成本优化建议

#### 生成优化建议

```go
optimizationService := cost.NewOptimizationService(db, costService)

suggestions, err := optimizationService.GenerateSuggestions("user-123", "project-456")

for _, suggestion := range suggestions {
    fmt.Printf("类型: %s\n", suggestion.Type)
    fmt.Printf("优先级: %s\n", suggestion.Priority)
    fmt.Printf("标题: %s\n", suggestion.Title)
    fmt.Printf("描述: %s\n", suggestion.Description)
    fmt.Printf("潜在节省: $%.2f\n", suggestion.PotentialSavings)
    fmt.Printf("建议操作: %s\n", suggestion.Action)
}
```

#### 生成成本报告

```go
report, err := optimizationService.GetCostReport("user-123", "project-456", startTime, endTime)

fmt.Println("=== 成本报告 ===")
fmt.Printf("时间范围: %s 至 %s\n", report.StartTime, report.EndTime)
fmt.Printf("总成本: $%.2f\n", report.OverallStats.TotalCost)
fmt.Printf("总调用: %d\n", report.OverallStats.TotalCalls)
fmt.Printf("平均成本/调用: $%.4f\n", report.OverallStats.AvgCostPerCall)
```

## 数据模型

### 成本记录

```go
type AICostRecord struct {
    ID            uint      // 记录ID
    Provider      string    // 提供商
    Model         string    // 模型名称
    UserID        string    // 用户ID
    ProjectID     string    // 项目ID
    FunctionType  string    // 功能类型
    InputTokens   int       // 输入token数
    OutputTokens  int       // 输出token数
    TotalTokens   int       // 总token数
    CostPerToken  float64   // 每token成本
    TotalCost     float64   // 总成本
    DurationMs    int       // 执行时长
    Success       bool      // 是否成功
    Cached        bool      // 是否来自缓存
    CreatedAt     time.Time // 创建时间
}
```

### 预算配置

```go
type AIBudget struct {
    ID             uint      // 预算ID
    UserID         string    // 用户ID
    ProjectID      string    // 项目ID
    BudgetType     string    // 预算类型：daily, weekly, monthly
    BudgetAmount   float64   // 预算金额
    UsedAmount     float64   // 已使用金额
    AlertThreshold float64   // 预警阈值
    IsActive       bool      // 是否激活
    StartDate      time.Time // 开始日期
    EndDate        time.Time // 结束日期
    LastResetAt    time.Time // 上次重置时间
}
```

## 定价配置

### 默认定价

系统内置了主流AI模型的定价：

| 提供商 | 模型 | 输入价格($/1K) | 输出价格($/1K) |
|--------|------|----------------|----------------|
| OpenRouter | gpt-3.5-turbo | 0.0015 | 0.002 |
| OpenRouter | gpt-4 | 0.03 | 0.06 |
| Google | gemini-pro | 0.00025 | 0.0005 |
| Groq | mixtral-8x7b | 0.00027 | 0.00027 |
| DeepSeek | deepseek-chat | 0.00014 | 0.00028 |
| NVIDIA | llama-3.1-8b | 0.00022 | 0.00022 |
| Mistral | mistral-large | 0.004 | 0.012 |

### 自定义定价

```go
// 添加自定义定价
pricing := &database.AIModelPricing{
    Provider:      "custom-provider",
    Model:         "custom-model",
    InputPrice:    0.001,
    OutputPrice:   0.002,
    Currency:      "USD",
    EffectiveDate: time.Now(),
    IsActive:      true,
}

db.Create(pricing)
```

## 最佳实践

### 1. 成本监控

```go
// 定期检查成本
ticker := time.NewTicker(1 * time.Hour)
go func() {
    for range ticker.C {
        stats, _ := costService.GetCostStats(userID, projectID, startTime, endTime)
        if stats.TotalCost > 50.0 { // 超过$50
            log.Printf("警告：本月成本已达 $%.2f", stats.TotalCost)
        }
    }
}()
```

### 2. 预算管理

```go
// 在AI调用前检查预算
status, _ := budgetService.CheckBudget(userID, projectID)
if status.IsExceeded {
    return errors.New("预算已超限，请充值或调整预算")
}

// 执行AI调用
result, err := aiService.Call(...)

// 记录成本
costService.RecordCost(...)

// 更新预算
budgetService.UpdateUsedAmount(userID, projectID, cost)
```

### 3. 成本优化

```go
// 定期生成优化建议
suggestions, _ := optimizationService.GenerateSuggestions(userID, projectID)

for _, suggestion := range suggestions {
    if suggestion.Priority == "high" {
        log.Printf("高优先级优化建议: %s", suggestion.Title)
        // 执行优化操作
    }
}
```

## 性能指标

### 典型成本数据

- **平均成本/调用**: $0.001 - $0.01
- **缓存命中率**: 30% - 85%
- **成本节省**: 通过优化可节省20% - 50%

### 成本分析示例

```
月度成本报告：
- 总调用次数: 10,000
- 总成本: $50.00
- 平均成本/调用: $0.005
- 缓存命中率: 65%
- 成功调用率: 98.5%

按提供商分布：
- OpenRouter: $30.00 (60%)
- Google: $15.00 (30%)
- NVIDIA: $5.00 (10%)

优化建议：
1. 提高缓存使用率（潜在节省: $10.00）
2. 优化提示词长度（潜在节省: $5.00）
3. 切换到更便宜的模型（潜在节省: $8.00）
```

## 故障排查

### 1. 成本记录不准确

**可能原因**:
- 定价配置错误
- Token计数不准确

**解决方案**:
- 检查AIModelPricing表
- 验证Token计数逻辑

### 2. 预算预警未触发

**可能原因**:
- 预算未激活
- 预警阈值设置过高

**解决方案**:
- 检查预算IsActive状态
- 调整AlertThreshold值

### 3. 成本统计异常

**可能原因**:
- 数据库查询错误
- 时间范围设置不当

**解决方案**:
- 检查数据库连接
- 验证时间范围参数

## 总结

AI成本追踪与预算控制系统通过完整的成本管理功能，帮助用户有效控制AI调用成本。结合缓存优化、模型选择和提示词优化，可以显著降低AI服务费用。

---

**文档版本**: v1.0  
**最后更新**: 2026-02-20  
**维护者**: MonkeyCode Team
