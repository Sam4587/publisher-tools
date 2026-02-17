# 全链路AI创作系统 - 开发计划

## 项目目标

实现从热点发现到内容发布的全流程自动化：
```
热点发现 → AI内容生成 → 平台发布 → 数据分析
```

---

## 当前完成状态

| 模块 | 完成度 | 说明 |
|------|--------|------|
| 平台发布框架 | 70% | 接口完整，任务执行待完善 |
| AI 服务集成 | 100% | 多提供商支持，API 完成 |
| 热点发现 | 30% | 前端完成，后端缺失 |
| 内容生成 | 50% | API 完成，前端待对接 |
| 数据分析 | 5% | 仅有框架 |

---

## Phase 1: 核心功能完善 (优先级 P0)

### 1.1 热点抓取模块

**目标**: 实现多平台热点数据抓取和存储

**目录结构**:
```
publisher-core/hotspot/
├── hotspot.go           # 热点服务接口
├── source.go            # 数据源接口定义
├── sources/
│   ├── newsnow.go       # NewsNow 聚合源 (主要)
│   ├── weibo.go         # 微博热搜
│   ├── douyin.go        # 抖音热点
│   ├── zhihu.go         # 知乎热榜
│   └── baidu.go         # 百度热搜
├── analyzer.go          # 趋势分析
├── storage.go           # 热点存储
└── api.go               # 热点 API (已有前端调用)
```

**数据模型**:
```go
type HotTopic struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Category    string    `json:"category"`
    Heat        int       `json:"heat"`         // 热度值
    Trend       string    `json:"trend"`        // up/down/stable/new
    Source      string    `json:"source"`       // 来源平台
    SourceURL   string    `json:"source_url"`
    Keywords    []string  `json:"keywords"`
    PublishedAt time.Time `json:"published_at"`
    CreatedAt   time.Time `json:"created_at"`
}

type HotSource struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Enabled bool   `json:"enabled"`
}
```

**API 端点** (对接现有前端):
| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/hot-topics` | GET | 热点列表 |
| `/api/hot-topics/{id}` | GET | 热点详情 |
| `/api/hot-topics/newsnow/sources` | GET | 数据源列表 |
| `/api/hot-topics/newsnow/fetch` | POST | 抓取热点 |
| `/api/hot-topics/update` | POST | 刷新热点 |
| `/api/hot-topics/trends/new` | GET | 新增热点 |
| `/api/hot-topics/ai/analyze` | POST | AI 分析 (已完成) |

**开发任务**:
- [ ] 创建 hotspot 模块目录结构
- [ ] 实现 Source 接口
- [ ] 实现 NewsNow 数据源抓取
- [ ] 实现热点存储 (JSON 文件 / SQLite)
- [ ] 实现热点 API handlers
- [ ] 集成到主服务

---

### 1.2 发布功能完善

**目标**: 实现完整的发布流程

**需要完善的文件**:
```
publisher-core/
├── api/server.go              # 完善 publish/uploadFile handlers
├── task/handlers/
│   └── publish.go             # 发布任务处理器 (新建)
├── storage/
│   └── local.go               # 本地存储实现
└── scheduler/
    ├── scheduler.go           # 定时调度器 (新建)
    └── queue.go               # 任务队列 (新建)
```

**开发任务**:
- [ ] 实现 `/api/v1/storage/upload` 文件上传
- [ ] 实现 publish 任务处理器
- [ ] 注册任务处理器到 TaskManager
- [ ] 实现定时发布调度器

---

## Phase 2: 内容生成增强 (优先级 P1)

### 2.1 AI 内容生成页面

**目标**: 前端页面对接 AI 内容生成 API

**新建文件**:
```
publisher-web/src/
├── pages/ContentGeneration.tsx    # 内容生成页面
├── components/content/
│   ├── AIGenerator.tsx           # AI 生成面板
│   ├── PromptEditor.tsx          # 提示词编辑
│   └── ContentPreview.tsx        # 内容预览
└── hooks/
    └── use-ai.ts                 # AI 相关 hooks
```

**功能**:
- 主题输入 → AI 生成内容
- 风格选择 (正式/幽默/感性/理性)
- 平台适配 (抖音/头条/小红书)
- 一键改写/扩写/摘要
- 内容审核检查

---

### 2.2 热点驱动内容生成

**目标**: 从热点自动生成内容

**流程**:
```
热点列表 → 选择热点 → AI 分析 → 生成内容草稿 → 人工审核 → 发布
```

**开发任务**:
- [ ] 热点页面添加"生成内容"按钮
- [ ] 实现热点到内容的转换流程
- [ ] 内容草稿管理

---

## Phase 3: 数据分析模块 (优先级 P2)

### 3.1 数据采集器

**目录结构**:
```
publisher-core/analytics/
├── analytics.go          # 分析服务
├── collector.go          # 采集器接口
├── collectors/
│   ├── douyin.go         # 抖音数据采集
│   ├── xiaohongshu.go    # 小红书数据采集
│   └── toutiao.go        # 头条数据采集
├── metrics/
│   ├── engagement.go     # 互动指标
│   └── reach.go          # 曝光指标
└── report/
    └── report.go         # 报告生成
```

### 3.2 前端分析页面

```
publisher-web/src/
├── pages/Analytics.tsx           # 数据分析页面
└── components/analytics/
    ├── Charts.tsx                # 图表组件
    ├── MetricsCard.tsx           # 指标卡片
    └── ReportExport.tsx          # 报告导出
```

---

## 开发排期

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| Phase 1.1 | 热点抓取模块 | 2-3 天 |
| Phase 1.2 | 发布功能完善 | 1-2 天 |
| Phase 2.1 | AI 内容生成页面 | 1-2 天 |
| Phase 2.2 | 热点驱动内容生成 | 1 天 |
| Phase 3 | 数据分析模块 | 2-3 天 |

---

## 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 热点存储 | SQLite / JSON | 轻量级，开发阶段足够 |
| 任务调度 | 内置 scheduler | 简单场景无需外部依赖 |
| 数据采集 | go-rod 浏览器自动化 | 已有基础设施 |
| 图表库 | Recharts | React 生态成熟方案 |

---

## 下一步行动

**立即开始: Phase 1.1 - 热点抓取模块**

1. 创建 hotspot 目录结构
2. 实现 Source 接口
3. 实现 NewsNow 数据源
4. 实现热点 API
5. 测试验证

是否开始执行？
