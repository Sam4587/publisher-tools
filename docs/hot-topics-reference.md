# 热点监控功能 - 项目借鉴文档

> 基于 [TrendRadar](https://github.com/sansan0/TrendRadar) 项目分析
> 
> TrendRadar 是一个拥有 46k+ stars 的热点监控开源项目，提供了完整的热点数据采集、分析、推送解决方案。

---

## 一、项目概述

### 1.1 TrendRadar 核心特性

- **多平台数据源**：支持微博、抖音、知乎、百度、今日头条、网易、新浪、腾讯等 8+ 主流平台
- **RSS 订阅支持**：支持自定义 RSS 源，扩展数据来源
- **AI 智能分析**：基于 LiteLLM 统一接口，支持 100+ AI 提供商进行热点分析
- **多渠道推送**：支持飞书、钉钉、企业微信、Telegram、邮件、ntfy、Bark、Slack 等多种通知方式
- **MCP Server**：提供 Model Context Protocol 服务器，支持 AI 助手直接调用
- **灵活调度**：支持定时任务、GitHub Actions、Docker 部署

### 1.2 技术栈

- **语言**：Python 3.8+
- **数据存储**：SQLite（本地）/ 远程存储
- **AI 接口**：LiteLLM（统一接口）
- **通知服务**：多渠道 Webhook
- **部署方式**：Docker / GitHub Actions / 本地运行

---

## 二、核心架构分析

### 2.1 项目结构

```
trendradar/
├── core/                    # 核心模块
│   ├── analyzer.py         # 数据分析器
│   ├── config.py           # 配置管理
│   ├── data.py             # 数据处理
│   ├── frequency.py        # 频次统计
│   ├── loader.py           # 数据加载
│   └── scheduler.py        # 调度系统
├── crawler/                 # 爬虫模块
│   ├── fetcher.py          # 数据获取器
│   └── rss/                # RSS 爬虫
│       ├── fetcher.py
│       └── parser.py
├── ai/                      # AI 分析模块
│   ├── analyzer.py         # AI 分析器
│   ├── client.py           # AI 客户端
│   ├── formatter.py        # 格式化器
│   └── translator.py       # 翻译器
├── notification/            # 通知模块
│   ├── batch.py            # 批量处理
│   ├── dispatcher.py       # 分发器
│   ├── formatters.py       # 格式化
│   ├── renderer.py         # 渲染器
│   ├── senders.py          # 发送器
│   └── splitter.py         # 分割器
├── report/                  # 报告模块
│   ├── formatter.py
│   ├── generator.py
│   ├── helpers.py
│   ├── html.py
│   └── rss_html.py
├── storage/                 # 存储模块
│   ├── base.py
│   ├── local.py
│   ├── manager.py
│   ├── remote.py
│   └── schema.sql          # 数据库结构
└── utils/                   # 工具模块
    ├── time.py
    └── url.py
```

### 2.2 数据流架构

```
┌─────────────┐
│  数据源层   │
│ (NewsNow API│
│  + RSS源)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  爬虫层     │
│ (Fetcher)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  存储层     │
│ (SQLite)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  分析层     │
│ (Analyzer)  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  AI 分析    │
│ (LiteLLM)   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  报告生成   │
│ (Report)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  推送层     │
│ (Notification)│
└─────────────┘
```

---

## 三、核心功能实现

### 3.1 数据采集（crawler/fetcher.py）

**关键特性：**
- 使用 NewsNow API 作为数据源
- 支持代理配置
- 自动重试机制
- 请求间隔控制

**核心代码：**
```python
class DataFetcher:
    DEFAULT_API_URL = "https://newsnow.busiyi.world/api/s"
    
    def fetch_data(self, id_info, max_retries=2):
        """获取指定平台数据，支持重试"""
        url = f"{self.api_url}?id={id_value}&latest"
        
        retries = 0
        while retries <= max_retries:
            try:
                response = requests.get(url, proxies=proxies, timeout=10)
                data_json = response.json()
                
                if data_json.get("status") in ["success", "cache"]:
                    return data_json["data"]
            except Exception as e:
                retries += 1
                wait_time = random.randint(min_retry_wait, max_retry_wait)
                time.sleep(wait_time)
```

**借鉴要点：**
1. 使用稳定的第三方 API（NewsNow）而非直接爬取各平台
2. 实现重试机制提高稳定性
3. 支持代理配置应对网络限制
4. 请求间隔控制避免被封

### 3.2 数据存储（storage/schema.sql）

**数据库设计：**

```sql
-- 平台信息表
CREATE TABLE platforms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active INTEGER DEFAULT 1
);

-- 新闻条目表
CREATE TABLE news_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT,
    platform_id TEXT NOT NULL,
    rank INTEGER,
    hot_value INTEGER,
    first_crawl_time TEXT,
    last_crawl_time TEXT,
    FOREIGN KEY (platform_id) REFERENCES platforms(id)
);

-- 排名历史表（记录排名变化）
CREATE TABLE rank_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    news_item_id INTEGER NOT NULL,
    rank INTEGER NOT NULL,
    crawl_time TEXT NOT NULL,
    FOREIGN KEY (news_item_id) REFERENCES news_items(id)
);

-- 抓取记录表
CREATE TABLE crawl_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    crawl_time TEXT NOT NULL UNIQUE,
    total_items INTEGER DEFAULT 0
);
```

**借鉴要点：**
1. **平台信息独立表**：支持平台名称变更，ID 保持不变
2. **排名历史记录**：记录每次抓取的排名，支持趋势分析
3. **URL + platform_id 唯一索引**：实现跨平台去重
4. **抓取记录表**：记录每次抓取状态，便于监控

### 3.3 数据分析（core/analyzer.py）

**分析维度：**
1. **频次统计**：统计标题在多平台出现的次数
2. **排名分析**：分析排名变化趋势
3. **热度计算**：综合排名、频次、热度值计算综合热度
4. **新增检测**：检测新出现的热点

**热度计算公式：**
```python
# 权重配置
weight = {
    "rank": 0.6,      # 排名权重
    "frequency": 0.3, # 频次权重
    "hotness": 0.1    # 热度权重
}

def calculate_score(rank, frequency, hotness):
    """计算综合热度分数"""
    rank_score = 100 - (rank - 1) * 2  # 排名越靠前分数越高
    freq_score = min(frequency * 20, 100)  # 频次越高分数越高
    hot_score = min(hotness / 10000, 100)  # 热度值归一化
    
    return (
        rank_score * weight["rank"] +
        freq_score * weight["frequency"] +
        hot_score * weight["hotness"]
    )
```

**借鉴要点：**
1. 多维度热度计算，而非单一指标
2. 支持权重配置，灵活调整
3. 记录排名时间线，支持趋势分析

### 3.4 AI 分析（ai/analyzer.py）

**核心特性：**
- 基于 LiteLLM 统一接口，支持 100+ AI 提供商
- 可配置的提示词模板
- 支持多语言翻译
- 分析结果结构化输出

**分析结果结构：**
```python
@dataclass
class AIAnalysisResult:
    success: bool
    total_news: int
    analyzed_news: int
    hotlist_count: int
    rss_count: int
    
    # 5 大核心板块
    overview: str              # 整体概览
    hot_events: List[Dict]     # 热点事件
    trends: List[str]          # 趋势分析
    insights: List[str]        # 深度洞察
    suggestions: List[str]     # 行动建议
    
    ai_mode: str               # AI 模式
    model: str                 # 使用的模型
```

**提示词模板结构：**
```
[system]
你是一个专业的新闻分析师...

[user]
请分析以下热点新闻数据：
{data}

请从以下维度进行分析：
1. 整体概览
2. 热点事件
3. 趋势分析
4. 深度洞察
5. 行动建议
```

**借鉴要点：**
1. 使用 LiteLLM 统一接口，避免绑定单一 AI 提供商
2. 提示词模板化，便于优化和调整
3. 结构化输出，便于后续处理
4. 支持多语言翻译

### 3.5 通知推送（notification/senders.py）

**支持的推送渠道：**
- 飞书（Feishu/Lark）
- 钉钉（DingTalk）
- 企业微信（WeCom）
- Telegram
- 邮件（Email）
- ntfy
- Bark
- Slack

**消息分批发送机制：**
```python
def send_to_feishu(webhook_url, content, batch_size=30000):
    """飞书消息发送（支持分批）"""
    
    # 1. 预留头部空间
    header_reserve = get_max_batch_header_size("feishu")
    
    # 2. 分割内容
    batches = split_content(
        content,
        max_bytes=batch_size - header_reserve
    )
    
    # 3. 添加批次头部
    batches = add_batch_headers(batches)
    
    # 4. 逐批发送
    for i, batch in enumerate(batches, 1):
        payload = {
            "msg_type": "text",
            "content": {"text": batch}
        }
        response = requests.post(webhook_url, json=payload)
        
        # 5. 批次间间隔
        if i < len(batches):
            time.sleep(batch_interval)
```

**借鉴要点：**
1. 消息分批发送，避免超出平台限制
2. 预留头部空间，确保添加批次号后不超限
3. 批次间间隔，避免频率限制
4. 统一的消息格式化接口

---

## 四、配置系统（config/config.yaml）

### 4.1 配置结构

```yaml
# 基础设置
app:
  timezone: "Asia/Shanghai"
  language: "zh-CN"

# 数据源配置
sources:
  # 热榜平台
  hotlist:
    - id: "weibo"
      name: "微博热搜"
      enabled: true
    - id: "douyin"
      name: "抖音热点"
      enabled: true
    # ... 更多平台
  
  # RSS 源
  rss:
    - name: "技术博客"
      url: "https://example.com/feed"
      enabled: true

# AI 配置
ai:
  enabled: true
  provider: "openai"  # openai / anthropic / deepseek / ...
  model: "gpt-4"
  api_key: "${AI_API_KEY}"
  api_base: "https://api.openai.com/v1"
  
  analysis:
    max_news: 50
    include_rss: true
    include_rank_timeline: true

# 通知配置
notification:
  feishu:
    enabled: true
    webhook_url: "${FEISHU_WEBHOOK}"
  
  dingtalk:
    enabled: false
    webhook_url: "${DINGTALK_WEBHOOK}"

# 调度配置
schedule:
  - time: "09:00"
    actions: ["fetch", "analyze", "push"]
  - time: "12:00"
    actions: ["fetch", "push"]
  - time: "18:00"
    actions: ["fetch", "analyze", "push"]
```

### 4.2 配置亮点

1. **环境变量支持**：`${VAR_NAME}` 格式，支持敏感信息保护
2. **可视化配置编辑器**：https://sansan0.github.io/TrendRadar/
3. **多账号支持**：每个通知渠道支持多个账号
4. **灵活调度**：支持按时间段配置不同操作

---

## 五、MCP Server（mcp_server/）

### 5.1 MCP 架构

```
mcp_server/
├── server.py           # MCP 服务器主入口
├── services/
│   ├── cache_service.py    # 缓存服务
│   ├── data_service.py     # 数据服务
│   └── parser_service.py   # 解析服务
├── tools/
│   ├── analytics.py        # 分析工具
│   ├── article_reader.py   # 文章阅读
│   ├── config_mgmt.py      # 配置管理
│   ├── data_query.py       # 数据查询
│   ├── notification.py     # 通知工具
│   ├── search_tools.py     # 搜索工具
│   ├── storage_sync.py     # 存储同步
│   └── system.py           # 系统工具
└── utils/
    ├── date_parser.py
    ├── errors.py
    └── validators.py
```

### 5.2 MCP 工具列表

```python
# 数据查询
@tool
def get_hot_topics(platform: str, limit: int = 20):
    """获取指定平台的热点话题"""

@tool
def search_topics(keyword: str, days: int = 7):
    """搜索包含关键词的话题"""

@tool
def get_topic_trend(topic_id: str):
    """获取话题趋势数据"""

# 分析工具
@tool
def analyze_hotness(topic_ids: List[str]):
    """分析话题热度"""

@tool
def compare_platforms(topic: str):
    """对比话题在不同平台的表现"""

# 通知工具
@tool
def send_notification(channel: str, message: str):
    """发送通知消息"""

# 配置管理
@tool
def get_config():
    """获取当前配置"""

@tool
def update_config(key: str, value: Any):
    """更新配置项"""
```

**借鉴要点：**
1. MCP 协议让 AI 助手可以直接调用热点监控功能
2. 工具化设计，每个功能独立封装
3. 支持缓存、数据服务、解析服务分层

---

## 六、部署方案

### 6.1 Docker 部署

```dockerfile
FROM python:3.9-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

CMD ["python", "-m", "trendradar"]
```

```yaml
# docker-compose.yml
version: '3'
services:
  trendradar:
    build: .
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    environment:
      - AI_API_KEY=${AI_API_KEY}
      - FEISHU_WEBHOOK=${FEISHU_WEBHOOK}
    restart: unless-stopped
```

### 6.2 GitHub Actions 部署

```yaml
name: Fetch and Push
on:
  schedule:
    - cron: '0 */1 * * *'  # 每小时执行
  workflow_dispatch:

jobs:
  fetch-push:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.9'
      
      - name: Install dependencies
        run: pip install -r requirements.txt
      
      - name: Run TrendRadar
        env:
          AI_API_KEY: ${{ secrets.AI_API_KEY }}
          FEISHU_WEBHOOK: ${{ secrets.FEISHU_WEBHOOK }}
        run: python -m trendradar
```

---

## 七、与当前项目的整合建议

### 7.1 架构对比

| 维度 | 当前项目 (Go) | TrendRadar (Python) | 整合建议 |
|------|--------------|---------------------|----------|
| 数据源 | NewsNow API | NewsNow API + RSS | ✅ 增加 RSS 支持 |
| 数据存储 | JSON 文件 | SQLite | ✅ 迁移到 SQLite |
| AI 分析 | 基础实现 | LiteLLM 统一接口 | ✅ 采用 LiteLLM 模式 |
| 通知推送 | 无 | 多渠道支持 | ✅ 增加通知模块 |
| MCP 支持 | 无 | 完整 MCP Server | ✅ 增加 MCP 支持 |
| 前端展示 | React | 无（纯推送） | ✅ 保留前端优势 |

### 7.2 整合路线图

#### Phase 1: 数据层优化（1-2 周）
- [ ] 迁移到 SQLite 存储
- [ ] 实现排名历史记录
- [ ] 增加 RSS 数据源支持
- [ ] 实现数据去重机制

#### Phase 2: 分析能力增强（2-3 周）
- [ ] 实现多维度热度计算
- [ ] 增加趋势分析功能
- [ ] 集成 LiteLLM 统一接口
- [ ] 优化 AI 分析提示词

#### Phase 3: 通知系统（1-2 周）
- [ ] 实现消息分批发送
- [ ] 支持飞书/钉钉/企业微信
- [ ] 支持邮件通知
- [ ] 实现通知模板系统

#### Phase 4: MCP Server（1-2 周）
- [ ] 实现 MCP 协议支持
- [ ] 封装核心功能为 MCP 工具
- [ ] 编写 MCP 文档
- [ ] 测试 AI 助手集成

#### Phase 5: 前端优化（1-2 周）
- [ ] 增加趋势图表展示
- [ ] 实现排名时间线可视化
- [ ] 增加 AI 分析结果展示
- [ ] 优化数据筛选和搜索

### 7.3 关键代码借鉴

#### 7.3.1 数据库 Schema

```sql
-- 建议在当前项目中添加
CREATE TABLE rank_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id TEXT NOT NULL,
    rank INTEGER NOT NULL,
    heat INTEGER,
    crawl_time TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(_id)
);

CREATE INDEX idx_rank_history_topic ON rank_history(topic_id);
CREATE INDEX idx_rank_history_time ON rank_history(crawl_time);
```

#### 7.3.2 热度计算

```go
// 建议在 hotspot/service.go 中添加
type HeatCalculator struct {
    RankWeight      float64 // 0.6
    FrequencyWeight float64 // 0.3
    HotnessWeight   float64 // 0.1
}

func (c *HeatCalculator) Calculate(rank, frequency, hotness int) int {
    rankScore := 100 - (rank-1)*2
    freqScore := min(frequency*20, 100)
    hotScore := min(hotness/10000, 100)
    
    return int(
        float64(rankScore)*c.RankWeight +
        float64(freqScore)*c.FrequencyWeight +
        float64(hotScore)*c.HotnessWeight,
    )
}
```

#### 7.3.3 AI 分析接口

```go
// 建议在 ai/provider/ 中添加 LiteLLM 适配器
type LiteLLMProvider struct {
    apiBase string
    apiKey  string
    model   string
}

func (p *LiteLLMProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
    // LiteLLM 统一接口格式
    payload := map[string]interface{}{
        "model": p.model,
        "messages": []map[string]string{
            {"role": "system", "content": opts.SystemPrompt},
            {"role": "user", "content": opts.UserPrompt},
        },
    }
    
    // 调用 LiteLLM API
    // ...
}
```

---

## 八、总结

### 8.1 TrendRadar 的核心优势

1. **成熟稳定**：46k+ stars，经过大量用户验证
2. **架构清晰**：模块化设计，易于理解和扩展
3. **功能完整**：从数据采集到通知推送的完整链路
4. **AI 集成**：基于 LiteLLM 的统一 AI 接口
5. **MCP 支持**：让 AI 助手可以直接调用

### 8.2 值得借鉴的关键点

1. **数据源策略**：使用稳定的第三方 API（NewsNow）而非直接爬取
2. **存储设计**：SQLite + 排名历史记录，支持趋势分析
3. **热度计算**：多维度加权计算，更科学
4. **AI 接口**：LiteLLM 统一接口，避免绑定单一提供商
5. **通知系统**：分批发送 + 多渠道支持
6. **MCP 协议**：让 AI 助手可以直接调用功能

### 8.3 下一步行动

1. **立即实施**：
   - 迁移到 SQLite 存储
   - 增加排名历史记录
   - 实现 RSS 数据源

2. **短期规划**：
   - 集成 LiteLLM 统一接口
   - 实现通知推送系统
   - 增加 MCP Server

3. **长期优化**：
   - 完善趋势分析功能
   - 优化前端可视化
   - 增加 AI 分析深度

---

## 附录

### A. 参考链接

- TrendRadar GitHub: https://github.com/sansan0/TrendRadar
- TrendRadar 在线演示: https://sansan0.github.io/TrendRadar/
- LiteLLM 文档: https://docs.litellm.ai/
- MCP 协议: https://modelcontextprotocol.io/

### B. 相关文件

- 配置示例: `config/config.yaml`
- 数据库结构: `trendradar/storage/schema.sql`
- AI 分析器: `trendradar/ai/analyzer.py`
- 通知发送: `trendradar/notification/senders.py`
- MCP Server: `mcp_server/server.py`

---

*文档创建时间：2026-02-20*
*最后更新：2026-02-20*
