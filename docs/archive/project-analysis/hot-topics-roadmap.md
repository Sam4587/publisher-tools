# 热点监控功能 - 开发路线图

> 基于 TrendRadar 项目分析制定的功能增强计划
> 
> 文档版本：v1.0
> 创建时间：2026-02-20

---

## 一、当前状态

### 1.1 已实现功能

- ✅ 基础数据采集（NewsNow API）
- ✅ JSON 文件存储
- ✅ 基础 API 接口
- ✅ 前端展示页面
- ✅ 数据源管理
- ✅ Mock 数据源

### 1.2 存在的问题

- ❌ 数据存储在 JSON 文件，不支持复杂查询
- ❌ 没有排名历史记录，无法分析趋势
- ❌ 缺少 RSS 数据源支持
- ❌ AI 分析功能基础，缺少深度
- ❌ 没有通知推送功能
- ❌ 不支持 MCP 协议

---

## 二、开发路线图

### Phase 1: 数据层优化（优先级：高）

**目标**：建立稳定、高效的数据存储和查询基础

**时间**：1-2 周

**任务清单**：

#### 1.1 SQLite 数据库迁移
- [ ] 设计数据库 Schema
  - [ ] platforms 表（平台信息）
  - [ ] topics 表（热点话题）
  - [ ] rank_history 表（排名历史）
  - [ ] crawl_records 表（抓取记录）
  - [ ] rss_sources 表（RSS 源配置）
  - [ ] rss_items 表（RSS 条目）
- [ ] 实现数据库初始化
- [ ] 实现数据迁移工具（JSON → SQLite）
- [ ] 更新存储接口实现
- [ ] 编写单元测试

**参考代码**：
```sql
-- 从 TrendRadar 借鉴的 Schema
CREATE TABLE rank_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id TEXT NOT NULL,
    rank INTEGER NOT NULL,
    heat INTEGER,
    crawl_time TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(_id)
);
```

#### 1.2 排名历史记录
- [ ] 实现排名历史记录功能
- [ ] 每次抓取时记录当前排名
- [ ] 实现排名变化检测
- [ ] 实现趋势计算算法

**关键算法**：
```go
// 趋势判断
func CalculateTrend(history []RankRecord) Trend {
    if len(history) < 2 {
        return TrendNew
    }
    
    latest := history[len(history)-1].Rank
    previous := history[len(history)-2].Rank
    
    if latest < previous {
        return TrendUp
    } else if latest > previous {
        return TrendDown
    }
    return TrendStable
}
```

#### 1.3 RSS 数据源支持
- [ ] 实现 RSS 源配置管理
- [ ] 实现 RSS 抓取器
- [ ] 实现 RSS 解析器
- [ ] 实现 RSS 条目存储
- [ ] 集成到主抓取流程

**参考实现**：
```go
type RSSFetcher struct {
    client *http.Client
    parser *RSSParser
}

func (f *RSSFetcher) Fetch(url string) ([]RSSEntry, error) {
    resp, err := f.client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return f.parser.Parse(resp.Body)
}
```

#### 1.4 数据去重机制
- [ ] 实现 URL + 平台去重
- [ ] 实现标题相似度去重
- [ ] 实现跨平台去重
- [ ] 添加去重配置选项

**验收标准**：
- ✅ 数据库迁移完成，所有历史数据保留
- ✅ 排名历史记录正常工作
- ✅ RSS 数据源可以正常抓取
- ✅ 去重机制有效，无重复数据

---

### Phase 2: 分析能力增强（优先级：高）

**目标**：提供深度、智能的热点分析能力

**时间**：2-3 周

**任务清单**：

#### 2.1 多维度热度计算
- [ ] 实现热度计算器
- [ ] 支持权重配置
- [ ] 实现排名分数计算
- [ ] 实现频次分数计算
- [ ] 实现热度值归一化

**核心算法**：
```go
type HeatCalculator struct {
    RankWeight      float64 // 0.6
    FrequencyWeight float64 // 0.3
    HotnessWeight   float64 // 0.1
}

func (c *HeatCalculator) Calculate(rank, frequency, hotness int) int {
    // 排名分数：排名越靠前分数越高
    rankScore := 100 - (rank-1)*2
    if rankScore < 0 {
        rankScore = 0
    }
    
    // 频次分数：频次越高分数越高
    freqScore := frequency * 20
    if freqScore > 100 {
        freqScore = 100
    }
    
    // 热度值归一化
    hotScore := hotness / 10000
    if hotScore > 100 {
        hotScore = 100
    }
    
    // 加权计算
    return int(
        float64(rankScore)*c.RankWeight +
        float64(freqScore)*c.FrequencyWeight +
        float64(hotScore)*c.HotnessWeight,
    )
}
```

#### 2.2 趋势分析功能
- [ ] 实现趋势检测算法
- [ ] 实现热度曲线生成
- [ ] 实现趋势预测（可选）
- [ ] 实现异常检测

**趋势分析维度**：
1. **短期趋势**：最近 1-3 小时
2. **中期趋势**：最近 24 小时
3. **长期趋势**：最近 7 天
4. **爆发检测**：短时间内热度急剧上升

#### 2.3 LiteLLM 集成
- [ ] 研究 LiteLLM Go SDK
- [ ] 实现统一 AI 接口
- [ ] 支持多 AI 提供商
- [ ] 实现模型配置管理

**支持的 AI 提供商**：
- OpenAI (GPT-4, GPT-3.5)
- Anthropic (Claude)
- DeepSeek
- Google AI (Gemini)
- 本地模型 (Ollama)

#### 2.4 AI 分析提示词优化
- [ ] 设计分析提示词模板
- [ ] 实现提示词变量替换
- [ ] 支持自定义提示词
- [ ] 实现提示词版本管理

**提示词模板示例**：
```
[system]
你是一个专业的新闻分析师，擅长从热点话题中发现趋势和洞察。

[user]
请分析以下热点新闻数据：

## 数据概览
- 总话题数：{total_count}
- 时间范围：{time_range}
- 平台覆盖：{platforms}

## 热点列表
{hot_topics}

请从以下维度进行分析：

### 1. 整体概览
总结当前热点整体情况...

### 2. 热点事件
列出最重要的 3-5 个热点事件...

### 3. 趋势分析
分析热点发展趋势...

### 4. 深度洞察
提供独特的见解和洞察...

### 5. 行动建议
给出具体的行动建议...
```

**验收标准**：
- ✅ 热度计算准确，符合预期
- ✅ 趋势分析功能正常
- ✅ AI 分析结果结构化、有价值
- ✅ 支持至少 3 个 AI 提供商

---

### Phase 3: 通知系统（优先级：中）

**目标**：实现多渠道、智能化的通知推送

**时间**：1-2 周

**任务清单**：

#### 3.1 通知框架设计
- [ ] 设计通知接口
- [ ] 实现通知管理器
- [ ] 实现消息队列
- [ ] 实现重试机制

#### 3.2 消息分批发送
- [ ] 实现消息分割算法
- [ ] 实现批次头部添加
- [ ] 实现批次间隔控制
- [ ] 实现发送状态跟踪

**关键实现**：
```go
type NotificationSender interface {
    Send(ctx context.Context, message string) error
    GetMaxSize() int
    GetName() string
}

type BatchSender struct {
    sender    NotificationSender
    batchSize int
    interval  time.Duration
}

func (s *BatchSender) Send(ctx context.Context, content string) error {
    // 1. 预留头部空间
    headerReserve := s.getHeaderReserve()
    
    // 2. 分割内容
    batches := s.splitContent(content, s.batchSize-headerReserve)
    
    // 3. 添加批次头部
    batches = s.addHeaders(batches)
    
    // 4. 逐批发送
    for i, batch := range batches {
        if err := s.sender.Send(ctx, batch); err != nil {
            return err
        }
        
        // 5. 批次间间隔
        if i < len(batches)-1 {
            time.Sleep(s.interval)
        }
    }
    
    return nil
}
```

#### 3.3 支持的通知渠道
- [ ] 飞书（Feishu）
- [ ] 钉钉（DingTalk）
- [ ] 企业微信（WeCom）
- [ ] Telegram
- [ ] 邮件（Email）
- [ ] Webhook（通用）

#### 3.4 通知模板系统
- [ ] 设计通知模板格式
- [ ] 实现模板渲染引擎
- [ ] 支持变量替换
- [ ] 支持条件渲染

**通知模板示例**：
```yaml
templates:
  daily_report:
    title: "📊 每日热点报告"
    body: |
      ## 📈 今日热点概览
      
      总话题数：{total_count}
      新增话题：{new_count}
      热门平台：{top_platforms}
      
      ## 🔥 Top 10 热点
      
      {#each topics as topic, i}
      {i+1}. {topic.title}
         热度：{topic.heat} | 趋势：{topic.trend}
      {/each}
      
      ## 🤖 AI 分析
      
      {ai_analysis}
```

**验收标准**：
- ✅ 至少支持 3 个通知渠道
- ✅ 消息分批发送正常
- ✅ 通知模板系统可用
- ✅ 发送失败有重试机制

---

### Phase 4: MCP Server（优先级：中）

**目标**：让 AI 助手可以直接调用热点监控功能

**时间**：1-2 周

**任务清单**：

#### 4.1 MCP 协议实现
- [ ] 研究 MCP 协议规范
- [ ] 实现 MCP Server 框架
- [ ] 实现工具注册机制
- [ ] 实现请求处理流程

#### 4.2 MCP 工具封装
- [ ] 数据查询工具
  - [ ] `get_hot_topics` - 获取热点话题
  - [ ] `search_topics` - 搜索话题
  - [ ] `get_topic_detail` - 获取话题详情
  - [ ] `get_topic_trend` - 获取话题趋势
- [ ] 分析工具
  - [ ] `analyze_hotness` - 分析热度
  - [ ] `compare_platforms` - 对比平台
  - [ ] `get_trend_report` - 获取趋势报告
- [ ] 通知工具
  - [ ] `send_notification` - 发送通知
  - [ ] `schedule_notification` - 定时通知
- [ ] 配置工具
  - [ ] `get_config` - 获取配置
  - [ ] `update_config` - 更新配置

**MCP 工具示例**：
```go
// 数据查询工具
mcp.Tool{
    Name: "get_hot_topics",
    Description: "获取指定平台的热点话题",
    Parameters: map[string]any{
        "platform": map[string]any{
            "type": "string",
            "description": "平台ID（weibo/douyin/zhihu等）",
        },
        "limit": map[string]any{
            "type": "integer",
            "description": "返回数量，默认20",
            "default": 20,
        },
    },
    Handler: func(args map[string]any) (any, error) {
        platform := args["platform"].(string)
        limit := args["limit"].(int)
        
        topics, err := service.GetHotTopics(platform, limit)
        if err != nil {
            return nil, err
        }
        
        return map[string]any{
            "success": true,
            "data": topics,
        }, nil
    },
}
```

#### 4.3 MCP 文档编写
- [ ] 编写 MCP Server 使用文档
- [ ] 编写工具调用示例
- [ ] 编写集成指南
- [ ] 编写最佳实践

**验收标准**：
- ✅ MCP Server 可以正常启动
- ✅ 至少实现 10 个 MCP 工具
- ✅ AI 助手可以成功调用工具
- ✅ 文档完整清晰

---

### Phase 5: 前端优化（优先级：低）

**目标**：提供更好的用户体验和数据可视化

**时间**：1-2 周

**任务清单**：

#### 5.1 趋势图表展示
- [ ] 集成图表库（ECharts / Recharts）
- [ ] 实现热度曲线图
- [ ] 实现排名变化图
- [ ] 实现平台对比图

#### 5.2 排名时间线可视化
- [ ] 实现时间线组件
- [ ] 显示排名变化轨迹
- [ ] 支持时间范围选择
- [ ] 支持动画效果

#### 5.3 AI 分析结果展示
- [ ] 设计分析结果展示组件
- [ ] 实现结构化内容渲染
- [ ] 支持 Markdown 渲染
- [ ] 支持导出功能

#### 5.4 数据筛选和搜索优化
- [ ] 实现高级筛选功能
- [ ] 实现全文搜索
- [ ] 实现筛选条件保存
- [ ] 实现筛选历史记录

**验收标准**：
- ✅ 图表展示正常，交互流畅
- ✅ 时间线可视化清晰
- ✅ AI 分析结果展示美观
- ✅ 搜索和筛选功能完善

---

## 三、技术选型

### 3.1 数据库
- **选择**：SQLite
- **理由**：
  - 轻量级，无需额外服务
  - 支持复杂查询
  - 易于备份和迁移
  - 性能足够（单机部署）

### 3.2 AI 接口
- **选择**：LiteLLM 模式（统一接口）
- **理由**：
  - 支持 100+ AI 提供商
  - 避免绑定单一提供商
  - 统一的调用接口
  - 易于切换和测试

### 3.3 通知服务
- **选择**：自研通知框架
- **理由**：
  - 完全可控
  - 易于扩展
  - 支持自定义渠道
  - 支持消息模板

### 3.4 MCP 实现
- **选择**：Go MCP SDK
- **理由**：
  - 与项目技术栈一致
  - 性能优秀
  - 易于集成

---

## 四、风险评估

### 4.1 技术风险

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| NewsNow API 不稳定 | 高 | 中 | 实现多数据源备份 |
| AI API 调用失败 | 中 | 中 | 实现重试和降级 |
| 数据库性能问题 | 中 | 低 | 优化查询，添加索引 |
| 通知渠道限制 | 低 | 中 | 实现消息分批和限流 |

### 4.2 进度风险

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| 需求变更 | 高 | 中 | 采用敏捷开发，快速迭代 |
| 技术难点 | 中 | 中 | 提前调研，准备备选方案 |
| 人员变动 | 高 | 低 | 完善文档，知识共享 |

---

## 五、资源需求

### 5.1 人力需求
- 后端开发：1 人
- 前端开发：1 人（Phase 5）
- 测试：0.5 人

### 5.2 硬件需求
- 开发环境：标准开发机
- 测试环境：1 台服务器
- 生产环境：1 台服务器（2核4G即可）

### 5.3 第三方服务
- AI API：OpenAI / DeepSeek / 其他
- 通知服务：各平台 Webhook（免费）

---

## 六、里程碑

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| M1: 数据层优化完成 | 第 2 周 | SQLite 存储、排名历史、RSS 支持 |
| M2: 分析能力增强完成 | 第 5 周 | 热度计算、趋势分析、AI 分析 |
| M3: 通知系统完成 | 第 7 周 | 多渠道通知、消息模板 |
| M4: MCP Server 完成 | 第 9 周 | MCP 工具、集成文档 |
| M5: 前端优化完成 | 第 11 周 | 图表展示、时间线、AI 结果展示 |

---

## 七、后续规划

### 7.1 短期（3 个月内）
- 完成所有 5 个 Phase
- 发布 v2.0 版本
- 完善文档和示例

### 7.2 中期（6 个月内）
- 增加更多数据源
- 优化 AI 分析深度
- 增加用户反馈机制
- 支持多租户

### 7.3 长期（1 年内）
- 构建热点预测模型
- 实现自动化内容生成
- 支持企业级部署
- 开放 API 平台

---

## 八、参考资源

### 8.1 TrendRadar 项目
- GitHub: https://github.com/sansan0/TrendRadar
- 在线演示: https://sansan0.github.io/TrendRadar/
- 详细分析: [hot-topics-reference.md](./hot-topics-reference.md)

### 8.2 技术文档
- LiteLLM: https://docs.litellm.ai/
- MCP 协议: https://modelcontextprotocol.io/
- SQLite: https://www.sqlite.org/docs.html

### 8.3 相关工具
- ECharts: https://echarts.apache.org/
- Recharts: https://recharts.org/
- Gorilla Mux: https://github.com/gorilla/mux

---

*文档维护：开发团队*
*最后更新：2026-02-20*
