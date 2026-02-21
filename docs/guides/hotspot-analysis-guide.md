# 热点趋势分析与内容适配系统使用指南

## 概述

热点趋势分析与内容适配系统提供完整的热点监控、趋势分析、内容适配和竞品分析功能，帮助用户快速响应热点事件并生成高质量内容。

## 核心功能

### 1. 热点趋势分析

#### 分析单个热点

```go
import "publisher-tools/hotspot"

// 创建趋势分析服务
trendService := hotspot.NewTrendAnalysisService(db, nil)
trendService.SetAIService(aiService)

// 分析热点
result, err := trendService.AnalyzeTopic(ctx, "topic-123")

fmt.Printf("热点标题: %s\n", result.Title)
fmt.Printf("关键词: %v\n", result.Keywords)
fmt.Printf("情感分析: %s (%.2f)\n", result.Sentiment.Label, result.Sentiment.Score)
fmt.Printf("趋势方向: %s\n", result.Trend.Direction)
fmt.Printf("相关性评分: %.2f\n", result.RelevanceScore)
```

#### 趋势分析结果

```json
{
  "topic_id": "topic-123",
  "title": "AI技术突破性进展",
  "keywords": [
    {"keyword": "AI", "frequency": 15, "score": 8.5, "type": "concept"},
    {"keyword": "技术", "frequency": 12, "score": 6.8, "type": "concept"},
    {"keyword": "突破", "frequency": 8, "score": 4.5, "type": "concept"}
  ],
  "sentiment": {
    "score": 0.75,
    "label": "positive",
    "confidence": 0.85,
    "emotions": {
      "joy": 0.8,
      "surprise": 0.6
    }
  },
  "trend": {
    "direction": "up",
    "speed": -2.5,
    "acceleration": 0.3,
    "predicted_life": 48
  },
  "relevance_score": 85.5,
  "content_suggestions": [
    "从技术角度分析AI突破的意义",
    "探讨AI技术对未来行业的影响"
  ]
}
```

#### 批量分析热点

```go
topicIDs := []string{"topic-1", "topic-2", "topic-3"}
results, err := trendService.BatchAnalyzeTopics(ctx, topicIDs)

for _, result := range results {
    fmt.Printf("热点: %s, 评分: %.2f\n", result.Title, result.RelevanceScore)
}
```

#### 生成趋势报告

```go
startTime := time.Now().AddDate(0, 0, -7)
endTime := time.Now()

report, err := trendService.GetTrendReport(ctx, startTime, endTime)

fmt.Printf("报告时间范围: %s 至 %s\n", report.StartTime, report.EndTime)
fmt.Printf("总热点数: %d\n", report.TotalTopics)
fmt.Printf("趋势分布: %v\n", report.TrendDistribution)
fmt.Printf("热门关键词: %v\n", report.TopKeywords)
```

### 2. 内容智能适配

#### 适配内容到平台

```go
// 创建内容适配服务
adaptService := hotspot.NewContentAdaptationService(db, nil)
adaptService.SetAIService(aiService)

// 适配内容
result, err := adaptService.AdaptContent(ctx, &hotspot.AdaptationRequest{
    TopicID:  "topic-123",
    Platform: "douyin",
    Style:    "轻松幽默",
    Keywords: []string{"AI", "技术", "创新"},
})

fmt.Printf("标题: %s\n", result.Title)
fmt.Printf("内容: %s\n", result.Content)
fmt.Printf("质量评分: %.2f\n", result.QualityScore)
```

#### 适配结果示例

```json
{
  "topic_id": "topic-123",
  "platform": "douyin",
  "title": "AI技术大突破！未来已来🚀",
  "content": "大家好！今天给大家分享一个超级重磅的消息...\n\nAI技术又有了新的突破...",
  "summary": "AI技术取得重大突破，将改变未来生活方式",
  "keywords": ["AI", "技术", "创新", "未来"],
  "tags": ["科技", "AI", "创新", "热点"],
  "quality_score": 92.5
}
```

#### 批量适配内容

```go
requests := []*hotspot.AdaptationRequest{
    {TopicID: "topic-1", Platform: "douyin"},
    {TopicID: "topic-2", Platform: "xiaohongshu"},
    {TopicID: "topic-3", Platform: "toutiao"},
}

results, err := adaptService.BatchAdaptContent(ctx, requests)
```

### 3. 热点监控和预警

#### 启动监控服务

```go
// 创建监控服务
monitorService := hotspot.NewMonitorService(db, &hotspot.MonitorConfig{
    CheckInterval:     5 * time.Minute,
    AlertThreshold:    80,
    TrendAlertEnabled: true,
    KeywordAlerts:     []string{"AI", "技术"},
})
monitorService.SetNotifyService(notifyService)

// 启动监控
go monitorService.Start()

// 停止监控
defer monitorService.Stop()
```

#### 获取预警列表

```go
// 获取所有预警
alerts := monitorService.GetAlerts("", 20)

for _, alert := range alerts {
    fmt.Printf("预警 [%s]: %s\n", alert.Level, alert.Message)
    fmt.Printf("热点: %s, 热度: %d\n", alert.Title, alert.Heat)
}

// 获取特定类型的预警
heatAlerts := monitorService.GetAlerts("heat", 10)
trendAlerts := monitorService.GetAlerts("trend", 10)
```

#### 管理预警

```go
// 标记为已读
monitorService.MarkAlertAsRead("alert-123")

// 解决预警
monitorService.ResolveAlert("alert-123")
```

### 4. 竞品分析

#### 分析竞品

```go
// 创建竞品分析服务
competitorService := hotspot.NewCompetitorAnalysisService(db)

// 分析竞品
analysis, err := competitorService.AnalyzeCompetitors(ctx, "topic-123")

fmt.Printf("市场位置: %s\n", analysis.MarketPosition)
fmt.Printf("竞品数量: %d\n", len(analysis.Competitors))

for _, comp := range analysis.Competitors {
    fmt.Printf("竞品: %s, 热度: %d, 趋势: %s\n", 
        comp.Name, comp.Heat, comp.Trend)
}

fmt.Printf("机会: %v\n", analysis.Opportunities)
fmt.Printf("威胁: %v\n", analysis.Threats)
fmt.Printf("建议: %v\n", analysis.Recommendations)
```

#### 竞品分析结果

```json
{
  "topic_id": "topic-123",
  "title": "AI技术突破性进展",
  "competitors": [
    {"name": "竞品A", "heat": 95, "trend": "up"},
    {"name": "竞品B", "heat": 88, "trend": "stable"},
    {"name": "竞品C", "heat": 82, "trend": "down"}
  ],
  "market_position": "市场挑战者",
  "opportunities": [
    "市场上升空间较大，竞争相对较小",
    "与领先者差距较小，有机会超越"
  ],
  "threats": [
    "竞品 '竞品A' 正在快速上升"
  ],
  "recommendations": [
    "加大内容投入，提升热度",
    "抓住上升趋势，快速扩大影响力"
  ]
}
```

## 配置说明

### 趋势分析配置

```go
config := &hotspot.TrendAnalysisConfig{
    KeywordMinFrequency: 2,      // 关键词最小出现频率
    SentimentThreshold:  0.6,    // 情感分析阈值
    RelevanceThreshold:  0.5,    // 相关性阈值
    TrendWindowSize:     24,     // 趋势分析窗口（小时）
    MaxKeywords:         20,     // 最大关键词数量
}
```

### 内容适配配置

```go
config := &hotspot.ContentAdaptationConfig{
    MaxContentLength: 2000,
    TargetPlatforms: []string{"douyin", "xiaohongshu", "toutiao"},
    ContentStyles: map[string]string{
        "douyin":      "轻松幽默，吸引眼球",
        "xiaohongshu": "精致优雅，实用分享",
        "toutiao":     "专业深度，信息丰富",
    },
    AdaptationDepth: "medium",
}
```

### 监控配置

```go
config := &hotspot.MonitorConfig{
    CheckInterval:      5 * time.Minute,
    AlertThreshold:     80,
    TrendAlertEnabled:  true,
    KeywordAlerts:      []string{"AI", "技术", "创新"},
    MaxAlertsPerHour:   10,
}
```

## 最佳实践

### 1. 热点监控

```go
// 定期检查热点趋势
ticker := time.NewTicker(1 * time.Hour)
go func() {
    for range ticker.C {
        // 分析最近的热点
        report, _ := trendService.GetTrendReport(ctx, startTime, endTime)
        
        // 检查是否有高价值热点
        for _, topic := range hotTopics {
            if topic.Heat > 80 && topic.Trend == "up" {
                // 自动适配内容
                adaptService.AdaptContent(ctx, &hotspot.AdaptationRequest{
                    TopicID: topic.ID,
                    Platform: "douyin",
                })
            }
        }
    }
}()
```

### 2. 内容优化

```go
// 基于分析结果优化内容
result, _ := trendService.AnalyzeTopic(ctx, topicID)

// 使用高价值关键词
var topKeywords []string
for _, kw := range result.Keywords {
    if kw.Score > 5.0 {
        topKeywords = append(topKeywords, kw.Keyword)
    }
}

// 适配内容时使用这些关键词
adaptResult, _ := adaptService.AdaptContent(ctx, &hotspot.AdaptationRequest{
    TopicID:  topicID,
    Platform: "douyin",
    Keywords: topKeywords,
})
```

### 3. 竞品跟踪

```go
// 定期分析竞品
go func() {
    ticker := time.NewTicker(6 * time.Hour)
    for range ticker.C {
        for _, topicID := range importantTopics {
            analysis, _ := competitorService.AnalyzeCompetitors(ctx, topicID)
            
            // 检查威胁
            if len(analysis.Threats) > 0 {
                log.Printf("发现威胁: %v", analysis.Threats)
                // 发送通知
            }
        }
    }
}()
```

## 性能指标

### 分析性能

- **单热点分析**: < 2秒
- **批量分析**: 10个热点 < 15秒
- **趋势报告生成**: < 5秒

### 内容适配性能

- **单次适配**: < 5秒
- **批量适配**: 10个内容 < 40秒
- **质量评分**: 实时计算

### 监控性能

- **检查间隔**: 可配置（默认5分钟）
- **预警延迟**: < 10秒
- **通知发送**: < 3秒

## 故障排查

### 1. 趋势分析不准确

**可能原因**:
- 历史数据不足
- 关键词提取算法需要优化

**解决方案**:
- 增加历史数据采集频率
- 调整关键词提取参数

### 2. 内容适配质量低

**可能原因**:
- AI模型选择不当
- 提示词设计不合理

**解决方案**:
- 使用更强大的AI模型
- 优化适配提示词模板

### 3. 预警过多

**可能原因**:
- 预警阈值设置过低
- 关键词匹配过于宽泛

**解决方案**:
- 提高预警阈值
- 精确关键词匹配规则

## 总结

热点趋势分析与内容适配系统通过智能化的热点监控、趋势分析和内容适配，帮助用户快速响应热点事件并生成高质量内容。结合竞品分析和预警机制，可以有效提升内容创作效率和市场竞争力。

---

**文档版本**: v1.0  
**最后更新**: 2026-02-20  
**维护者**: MonkeyCode Team
