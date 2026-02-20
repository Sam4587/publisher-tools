# 项目架构优化 - 统一实施方案

> 基于当前项目架构分析和四个借鉴项目的深度整合
> 
> 文档版本：v2.0
> 创建时间：2026-02-20
> 最后更新：2026-02-20

---

## 📋 文档说明

本文档是项目架构优化的**统一实施方案**，基于：
- 当前项目架构的深度分析
- 四个优秀开源项目的借鉴整合
- 统一的需求分析和架构设计
- 可执行的开发路线图

**目标读者**：AI 助手、开发者、项目经理

**使用方式**：
1. AI 助手阅读本文档了解完整方案
2. 按照优先级选择任务执行
3. 完成后在"开发进度"部分记录
4. 下一个 AI 助手从进度记录继续

---

## 一、当前项目架构分析

### 1.1 项目概况

**项目名称**：Publisher Tools
**项目定位**：多平台内容发布系统
**技术栈**：Go + React + SQLite/JSON

**核心功能**：
- ✅ 多平台内容发布（抖音、今日头条、小红书）
- ✅ 浏览器自动化（Rod 框架）
- ✅ Cookie 管理
- ✅ 任务管理系统
- ✅ 基础热点监控（NewsNow API）
- ✅ 基础 AI 服务（OpenRouter、DeepSeek）
- ✅ 文件存储抽象层
- ✅ 智能启动系统

### 1.2 目录结构

```
publisher-tools/
├── publisher-core/           # Go 后端核心
│   ├── adapters/            # 平台适配器
│   ├── ai/                  # AI 服务
│   ├── analytics/           # 数据分析
│   ├── api/                 # API 路由
│   ├── hotspot/             # 热点监控
│   ├── storage/             # 文件存储
│   ├── task/                # 任务管理
│   └── cmd/                 # 入口程序
├── publisher-web/           # React 前端
│   ├── src/
│   │   ├── components/     # 组件
│   │   ├── pages/          # 页面
│   │   ├── lib/            # 工具库
│   │   └── types/          # 类型定义
│   └── package.json
├── server/                  # Node.js 辅助服务
├── docs/                    # 文档中心
├── bin/                     # 编译产物
├── logs/                    # 日志文件
├── data/                    # 数据文件
├── cookies/                 # Cookie 存储
└── uploads/                 # 上传文件
```

### 1.3 技术栈清单

#### 后端技术
| 技术 | 版本 | 用途 | 状态 |
|------|------|------|------|
| Go | 1.21+ | 主要语言 | ✅ 已使用 |
| Gorilla Mux | - | HTTP 路由 | ✅ 已使用 |
| Rod | - | 浏览器自动化 | ✅ 已使用 |
| GORM | - | ORM（未使用） | ⚠️ 可用 |
| SQLite | - | 数据库（未使用） | ⚠️ 可用 |

#### 前端技术
| 技术 | 版本 | 用途 | 状态 |
|------|------|------|------|
| React | 18 | UI 框架 | ✅ 已使用 |
| TypeScript | 5.x | 类型安全 | ✅ 已使用 |
| Vite | 5.x | 构建工具 | ✅ 已使用 |
| Tailwind CSS | 3.x | 样式框架 | ✅ 已使用 |
| shadcn/ui | - | 组件库 | ✅ 已使用 |

#### AI 服务
| 提供商 | 用途 | 状态 |
|--------|------|------|
| OpenRouter | 文本生成 | ✅ 已集成 |
| DeepSeek | 文本生成 | ✅ 已集成 |
| Google AI | 文本生成 | ⚠️ 未集成 |
| Groq | 快速推理 | ⚠️ 未集成 |

### 1.4 架构优缺点分析

#### 优点
1. ✅ **技术栈现代化**：Go + React + TypeScript
2. ✅ **模块化设计**：清晰的模块划分
3. ✅ **功能完整**：发布、监控、AI 集成
4. ✅ **部署灵活**：支持多种部署方式
5. ✅ **文档完善**：详细的开发文档
6. ✅ **智能启动**：完善的启动脚本系统

#### 缺点
1. ❌ **数据存储简单**：JSON 文件存储，不支持复杂查询
2. ❌ **AI 服务未统一**：缺少统一的 AI 接口层
3. ❌ **热点监控不完善**：无趋势分析、无通知推送
4. ❌ **缺少视频处理**：无视频内容处理能力
5. ❌ **无 MCP 支持**：AI 助手无法直接调用
6. ❌ **无消息队列**：异步任务处理能力有限

---

## 二、借鉴项目整合分析

### 2.1 四个借鉴项目对比

| 项目 | Stars | 技术栈 | 核心价值 | 契合度 |
|------|-------|--------|---------|--------|
| **TrendRadar** | 46k+ | Python | 热点监控完整方案 | ⭐⭐⭐⭐ |
| **Free LLM API Resources** | 11k+ | - | 免费 AI 资源汇总 | ⭐⭐⭐⭐⭐ |
| **AI-Video-Transcriber** | 2k+ | Python | 视频转录方案 | ⭐⭐⭐ |
| **Huobao Drama** | 7.6k+ | Go + Vue3 | AI 短剧生成平台 | ⭐⭐⭐⭐⭐ |

### 2.2 核心借鉴内容

#### 从 TrendRadar 借鉴
1. **数据采集架构**
   - NewsNow API 集成方式
   - RSS 数据源支持
   - 重试机制和代理配置

2. **数据存储设计**
   - SQLite 数据库 Schema
   - 排名历史记录表
   - 抓取记录表

3. **AI 分析方案**
   - LiteLLM 统一接口
   - 提示词模板设计
   - 结构化输出

4. **通知推送系统**
   - 多渠道支持
   - 消息分批发送
   - 通知模板

5. **MCP Server**
   - 工具化接口设计
   - 数据查询工具
   - 分析工具

#### 从 Free LLM API Resources 借鉴
1. **免费 AI 资源**
   - 20+ 免费 AI 提供商
   - 提供商限制信息
   - API Key 获取方式

2. **提供商选择策略**
   - 根据任务类型选择
   - 免费额度优先
   - 智能降级

#### 从 AI-Video-Transcriber 借鉴
1. **视频处理流程**
   - yt-dlp 集成
   - Faster-Whisper 转录
   - AI 文本优化

2. **长文本处理**
   - 自动分块算法
   - Token 估算
   - 多语言支持

#### 从 Huobao Drama 借鉴（重点）
1. **AI 服务统一管理** ⭐⭐⭐⭐⭐
   - `AIServiceConfig` 数据模型
   - 多提供商端点配置
   - 优先级和默认配置
   - 服务类型抽象（text、image、video）

2. **架构分层设计** ⭐⭐⭐⭐⭐
   - API 层（handlers）
   - 应用层（services）
   - 领域层（models）
   - 基础设施层（pkg）

3. **任务管理机制** ⭐⭐⭐⭐
   - 异步任务处理
   - 进度追踪
   - 错误处理

4. **提示词国际化** ⭐⭐⭐⭐
   - 多语言提示词支持
   - 提示词模板管理

### 2.3 技术栈契合度分析

#### Huobao Drama（最高契合度）
**契合点**：
- ✅ Go 后端 + Vue3/React 前端
- ✅ GORM ORM
- ✅ SQLite 数据库
- ✅ AI 服务统一管理
- ✅ 任务管理系统

**可直接借鉴**：
- AI 服务配置管理（100% 可用）
- 数据模型设计（90% 可用）
- 服务层架构（95% 可用）
- 任务管理机制（90% 可用）

#### TrendRadar（高契合度）
**契合点**：
- ✅ 热点监控功能
- ✅ AI 分析方案
- ✅ 通知推送系统

**需要适配**：
- Python → Go 语言转换
- 架构模式调整

---

## 三、统一架构设计

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        前端层 (React + TypeScript)               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ 内容发布 │  │ 热点监控 │  │ 视频处理 │  │ AI 创作  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────────────────────────────────────────────┘
                              │ HTTP/REST + WebSocket
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      API 网关层 (Go + Gorilla Mux)               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  REST API + WebSocket + MCP Server + 中间件              │  │
│  │  (CORS、日志、认证、限流)                                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       应用层 (Services)                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ 发布服务 │  │ 热点服务 │  │ 视频服务 │  │ AI 服务  │       │
│  │Publisher │  │ Hotspot  │  │  Video   │  │   AI     │       │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                     │
│  │ 任务服务 │  │ 通知服务 │  │ 分析服务 │                     │
│  │  Task    │  │NotifySvc │  │Analytics │                     │
│  │ Service  │  │          │  │ Service  │                     │
│  └──────────┘  └──────────┘  └──────────┘                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       领域层 (Models)                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ Platform │  │  Topic   │  │  Video   │  │AIConfig  │       │
│  │  Task    │  │  Rank    │  │Transcript│  │  Prompt  │       │
│  │  Cookie  │  │  Trend   │  │  Audio   │  │  Result  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    基础设施层 (Infrastructure)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ SQLite   │  │ 文件存储 │  │ AI 客户端│  │ 任务队列 │       │
│  │ Database │  │ Storage  │  │AIClients │  │  Queue   │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ 浏览器   │  │ 视频工具 │  │ 通知渠道 │  │ 缓存系统 │       │
│  │ Browser  │  │VideoTools│  │ Notifier │  │  Cache   │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 核心模块设计

#### 模块 1: AI 服务模块（统一管理）⭐ 优先级最高

**借鉴来源**：Huobao Drama

**数据模型**：
```go
// domain/models/ai_service_config.go

type AIServiceConfig struct {
    ID            uint           `gorm:"primaryKey" json:"id"`
    ServiceType   string         `json:"service_type"`   // text, image, video, audio
    Name          string         `json:"name"`
    Provider      string         `json:"provider"`       // openai, google, doubao, openrouter, groq
    BaseURL       string         `json:"base_url"`
    APIKey        string         `json:"api_key"`
    Model         string         `json:"model"`
    Endpoint      string         `json:"endpoint"`
    QueryEndpoint string         `json:"query_endpoint"`
    Priority      int            `json:"priority"`
    IsDefault     bool           `json:"is_default"`
    IsActive      bool           `json:"is_active"`
    Settings      datatypes.JSON `json:"settings"`       // 额外配置
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
}

// 预定义提供商配置
var DefaultProviders = []AIServiceConfig{
    {
        ServiceType: "text",
        Provider:    "openrouter",
        Name:        "OpenRouter GPT-4",
        BaseURL:     "https://openrouter.ai/api/v1",
        Model:       "openai/gpt-4",
        Endpoint:    "/chat/completions",
        Priority:    100,
        IsDefault:   true,
    },
    {
        ServiceType: "text",
        Provider:    "groq",
        Name:        "Groq Llama 3.3 70B",
        BaseURL:     "https://api.groq.com/openai/v1",
        Model:       "llama-3.3-70b-versatile",
        Endpoint:    "/chat/completions",
        Priority:    90,
    },
    {
        ServiceType: "text",
        Provider:    "google",
        Name:        "Google Gemini Flash",
        BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
        Model:       "gemini-2.5-flash",
        Endpoint:    "/models/{model}:generateContent",
        Priority:    80,
    },
    // ... 更多提供商
}
```

**服务实现**：
```go
// application/services/ai_service.go

type AIService struct {
    db      *gorm.DB
    log     *logger.Logger
    clients map[string]*ai.Client  // 客户端缓存
    mu      sync.RWMutex
}

// 获取默认客户端
func (s *AIService) GetDefaultClient(serviceType string) (*ai.Client, error) {
    config, err := s.GetDefaultConfig(serviceType)
    if err != nil {
        return nil, err
    }
    return s.GetOrCreateClient(config)
}

// 获取或创建客户端
func (s *AIService) GetOrCreateClient(config *models.AIServiceConfig) (*ai.Client, error) {
    key := fmt.Sprintf("%s:%s", config.Provider, config.Model)
    
    s.mu.RLock()
    client, ok := s.clients[key]
    s.mu.RUnlock()
    
    if ok {
        return client, nil
    }
    
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // 双重检查
    if client, ok := s.clients[key]; ok {
        return client, nil
    }
    
    client = ai.NewClient(&ai.ClientConfig{
        Provider:      config.Provider,
        BaseURL:       config.BaseURL,
        APIKey:        config.APIKey,
        Model:         config.Model,
        Endpoint:      config.Endpoint,
        QueryEndpoint: config.QueryEndpoint,
    })
    
    s.clients[key] = client
    return client, nil
}

// 生成文本（支持降级）
func (s *AIService) GenerateText(ctx context.Context, prompt string, opts ...ai.Option) (string, error) {
    configs, err := s.GetActiveConfigs("text")
    if err != nil {
        return "", err
    }
    
    var lastErr error
    for _, config := range configs {
        client, err := s.GetOrCreateClient(&config)
        if err != nil {
            lastErr = err
            continue
        }
        
        result, err := client.GenerateText(ctx, prompt, opts...)
        if err == nil {
            return result, nil
        }
        
        s.log.Warnw("AI generation failed, trying next provider",
            "provider", config.Provider,
            "error", err)
        lastErr = err
    }
    
    return "", fmt.Errorf("all providers failed: %w", lastErr)
}
```

#### 模块 2: 热点监控模块（增强）

**借鉴来源**：TrendRadar + Huobao Drama

**数据模型**：
```go
// domain/models/hotspot.go

type Platform struct {
    ID        string    `gorm:"primaryKey" json:"id"`
    Name      string    `json:"name"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Topic struct {
    ID             string    `gorm:"primaryKey" json:"id"`
    Title          string    `gorm:"index" json:"title"`
    Description    string    `json:"description"`
    Category       string    `json:"category"`
    PlatformID     string    `json:"platform_id"`
    Platform       Platform  `gorm:"foreignKey:PlatformID" json:"platform"`
    URL            string    `json:"url"`
    Heat           int       `json:"heat"`
    Trend          string    `json:"trend"` // up, down, stable, new, hot
    FirstCrawlTime time.Time `json:"first_crawl_time"`
    LastCrawlTime  time.Time `json:"last_crawl_time"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}

type RankHistory struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    TopicID    string    `gorm:"index" json:"topic_id"`
    Topic      Topic     `gorm:"foreignKey:TopicID" json:"topic"`
    Rank       int       `json:"rank"`
    Heat       int       `json:"heat"`
    CrawlTime  time.Time `json:"crawl_time"`
    CreatedAt  time.Time `json:"created_at"`
}

type CrawlRecord struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    CrawlTime  time.Time `gorm:"uniqueIndex" json:"crawl_time"`
    TotalItems int       `json:"total_items"`
    Status     string    `json:"status"` // success, failed, partial
    CreatedAt  time.Time `json:"created_at"`
}
```

**服务实现**：
```go
// application/services/hotspot_service.go

type HotspotService struct {
    db            *gorm.DB
    aiService     *AIService
    notifyService *NotifyService
    sources       map[string]Source
    log           *logger.Logger
}

// 计算综合热度
func (s *HotspotService) CalculateHeat(rank, frequency, hotness int) int {
    // 权重配置
    rankWeight := 0.6
    freqWeight := 0.3
    hotWeight := 0.1
    
    // 排名分数
    rankScore := 100 - (rank-1)*2
    if rankScore < 0 {
        rankScore = 0
    }
    
    // 频次分数
    freqScore := frequency * 20
    if freqScore > 100 {
        freqScore = 100
    }
    
    // 热度分数
    hotScore := hotness / 10000
    if hotScore > 100 {
        hotScore = 100
    }
    
    return int(float64(rankScore)*rankWeight +
        float64(freqScore)*freqWeight +
        float64(hotScore)*hotWeight)
}

// 分析趋势
func (s *HotspotService) AnalyzeTrend(topicID string) (string, error) {
    var history []models.RankHistory
    err := s.db.Where("topic_id = ?", topicID).
        Order("crawl_time desc").
        Limit(10).
        Find(&history).Error
    if err != nil {
        return "", err
    }
    
    if len(history) < 2 {
        return "new", nil
    }
    
    latest := history[0].Rank
    previous := history[1].Rank
    
    if latest < previous {
        return "up", nil
    } else if latest > previous {
        return "down", nil
    }
    return "stable", nil
}

// AI 分析热点
func (s *HotspotService) AIAnalyze(topics []models.Topic) (*AIAnalysisResult, error) {
    // 构建提示词
    prompt := s.buildAnalysisPrompt(topics)
    
    // 调用 AI 服务
    result, err := s.aiService.GenerateText(context.Background(), prompt)
    if err != nil {
        return nil, err
    }
    
    // 解析结果
    var analysis AIAnalysisResult
    if err := json.Unmarshal([]byte(result), &analysis); err != nil {
        return nil, err
    }
    
    return &analysis, nil
}
```

#### 模块 3: 视频处理模块（新增）

**借鉴来源**：AI-Video-Transcriber

**数据模型**：
```go
// domain/models/video.go

type Video struct {
    ID           string      `gorm:"primaryKey" json:"id"`
    URL          string      `json:"url"`
    Platform     string      `json:"platform"`
    Title        string      `json:"title"`
    Duration     int         `json:"duration"` // 秒
    Status       string      `json:"status"` // pending, processing, completed, failed
    Transcript   *Transcript `gorm:"foreignKey:VideoID" json:"transcript"`
    CreatedAt    time.Time   `json:"created_at"`
    UpdatedAt    time.Time   `json:"updated_at"`
}

type Transcript struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    VideoID      string    `gorm:"uniqueIndex" json:"video_id"`
    Language     string    `json:"language"`
    Content      string    `gorm:"type:text" json:"content"`
    Optimized    string    `gorm:"type:text" json:"optimized"`
    Summary      string    `gorm:"type:text" json:"summary"`
    CreatedAt    time.Time `json:"created_at"`
}
```

**服务实现**：
```go
// application/services/video_service.go

type VideoService struct {
    db        *gorm.DB
    aiService *AIService
    log       *logger.Logger
}

// 处理视频
func (s *VideoService) ProcessVideo(videoURL string) (*models.Video, error) {
    // 1. 创建视频记录
    video := &models.Video{
        ID:     uuid.New().String(),
        URL:    videoURL,
        Status: "pending",
    }
    s.db.Create(video)
    
    // 2. 下载视频
    videoPath, err := s.downloadVideo(videoURL)
    if err != nil {
        video.Status = "failed"
        s.db.Save(video)
        return nil, err
    }
    
    // 3. 转录音频
    transcript, err := s.transcribeAudio(videoPath)
    if err != nil {
        video.Status = "failed"
        s.db.Save(video)
        return nil, err
    }
    
    // 4. AI 优化文本
    optimized, err := s.optimizeTranscript(transcript)
    if err != nil {
        video.Status = "failed"
        s.db.Save(video)
        return nil, err
    }
    
    // 5. 生成摘要
    summary, err := s.generateSummary(optimized)
    if err != nil {
        video.Status = "failed"
        s.db.Save(video)
        return nil, err
    }
    
    // 6. 保存结果
    video.Status = "completed"
    video.Transcript = &models.Transcript{
        VideoID:   video.ID,
        Content:   transcript,
        Optimized: optimized,
        Summary:   summary,
    }
    s.db.Save(video)
    
    return video, nil
}

// 转录音频（使用 Faster-Whisper）
func (s *VideoService) transcribeAudio(videoPath string) (string, error) {
    // 调用 Faster-Whisper 进行转录
    // 这里需要集成 Python 的 Faster-Whisper 或使用 Go 绑定
    // 简化实现：调用外部服务
    cmd := exec.Command("whisper", videoPath, "--model", "base", "--output_format", "txt")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}

// AI 优化文本
func (s *VideoService) optimizeTranscript(transcript string) (string, error) {
    prompt := fmt.Sprintf(`
请优化以下转录文本：
1. 修正错别字
2. 补全不完整的句子
3. 按语义分段
4. 保持原意不变

转录文本：
%s
`, transcript)
    
    return s.aiService.GenerateText(context.Background(), prompt)
}
```

#### 模块 4: 通知服务模块（新增）

**借鉴来源**：TrendRadar

**数据模型**：
```go
// domain/models/notification.go

type NotificationChannel struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Type      string    `json:"type"` // feishu, dingtalk, wecom, telegram, email
    Name      string    `json:"name"`
    Webhook   string    `json:"webhook"`
    IsActive  bool      `json:"is_active"`
    Config    datatypes.JSON `json:"config"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type NotificationTemplate struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `json:"name"`
    Title     string    `json:"title"`
    Body      string    `gorm:"type:text" json:"body"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**服务实现**：
```go
// application/services/notification_service.go

type NotificationService struct {
    db        *gorm.DB
    channels  map[string]Notifier
    log       *logger.Logger
}

type Notifier interface {
    Send(ctx context.Context, message string) error
    GetMaxSize() int
    GetName() string
}

// 发送通知（支持分批）
func (s *NotificationService) Send(ctx context.Context, channelType, content string) error {
    notifier, ok := s.channels[channelType]
    if !ok {
        return fmt.Errorf("channel %s not found", channelType)
    }
    
    // 分批发送
    maxSize := notifier.GetMaxSize()
    if len(content) <= maxSize {
        return notifier.Send(ctx, content)
    }
    
    // 分割内容
    batches := s.splitContent(content, maxSize-100) // 预留头部空间
    
    for i, batch := range batches {
        // 添加批次头部
        header := fmt.Sprintf("[%d/%d]\n", i+1, len(batches))
        message := header + batch
        
        if err := notifier.Send(ctx, message); err != nil {
            return err
        }
        
        // 批次间间隔
        if i < len(batches)-1 {
            time.Sleep(3 * time.Second)
        }
    }
    
    return nil
}

// 飞书通知器
type FeishuNotifier struct {
    webhook string
}

func (n *FeishuNotifier) Send(ctx context.Context, message string) error {
    payload := map[string]interface{}{
        "msg_type": "text",
        "content": map[string]string{
            "text": message,
        },
    }
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST", n.webhook, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("feishu notification failed: %s", resp.Status)
    }
    
    return nil
}

func (n *FeishuNotifier) GetMaxSize() int {
    return 30000 // 飞书限制 30KB
}

func (n *FeishuNotifier) GetName() string {
    return "feishu"
}
```

#### 模块 5: MCP Server 模块（新增）

**借鉴来源**：TrendRadar

**服务实现**：
```go
// mcp/server.go

type MCPServer struct {
    tools map[string]Tool
    log   *logger.Logger
}

type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
    Handler     func(args map[string]interface{}) (interface{}, error)
}

// 注册工具
func (s *MCPServer) RegisterTool(tool Tool) {
    s.tools[tool.Name] = tool
}

// 处理请求
func (s *MCPServer) HandleRequest(req *Request) (*Response, error) {
    tool, ok := s.tools[req.Tool]
    if !ok {
        return nil, fmt.Errorf("tool %s not found", req.Tool)
    }
    
    result, err := tool.Handler(req.Arguments)
    if err != nil {
        return &Response{
            Success: false,
            Error:   err.Error(),
        }, nil
    }
    
    return &Response{
        Success: true,
        Data:    result,
    }, nil
}

// 注册热点监控工具
func (s *MCPServer) registerHotspotTools(hotspotService *HotspotService) {
    // 获取热点话题
    s.RegisterTool(Tool{
        Name:        "get_hot_topics",
        Description: "获取指定平台的热点话题",
        Parameters: map[string]interface{}{
            "platform": map[string]string{
                "type":        "string",
                "description": "平台ID（weibo/douyin/zhihu等）",
            },
            "limit": map[string]interface{}{
                "type":        "integer",
                "description": "返回数量，默认20",
                "default":     20,
            },
        },
        Handler: func(args map[string]interface{}) (interface{}, error) {
            platform := args["platform"].(string)
            limit := args["limit"].(int)
            
            topics, err := hotspotService.GetTopics(platform, limit)
            if err != nil {
                return nil, err
            }
            
            return map[string]interface{}{
                "success": true,
                "data":    topics,
            }, nil
        },
    })
    
    // 分析热点
    s.RegisterTool(Tool{
        Name:        "analyze_hotness",
        Description: "分析热点话题的热度",
        Parameters: map[string]interface{}{
            "topic_ids": map[string]string{
                "type":        "array",
                "description": "话题ID列表",
            },
        },
        Handler: func(args map[string]interface{}) (interface{}, error) {
            topicIDs := args["topic_ids"].([]string)
            
            analysis, err := hotspotService.AIAnalyze(topicIDs)
            if err != nil {
                return nil, err
            }
            
            return map[string]interface{}{
                "success": true,
                "data":    analysis,
            }, nil
        },
    })
}
```

---

## 四、数据库设计

### 4.1 完整 Schema

```sql
-- =====================================================
-- AI 服务配置表
-- =====================================================
CREATE TABLE ai_service_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_type TEXT NOT NULL,      -- text, image, video, audio
    name TEXT NOT NULL,
    provider TEXT NOT NULL,          -- openai, google, doubao, openrouter, groq
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    model TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    query_endpoint TEXT,
    priority INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT 0,
    is_active BOOLEAN DEFAULT 1,
    settings TEXT,                   -- JSON 格式额外配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ai_service_type ON ai_service_configs(service_type);
CREATE INDEX idx_ai_provider ON ai_service_configs(provider);
CREATE INDEX idx_ai_active ON ai_service_configs(is_active);

-- =====================================================
-- 热点监控表
-- =====================================================
CREATE TABLE platforms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE topics (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT,
    platform_id TEXT NOT NULL,
    url TEXT,
    heat INTEGER DEFAULT 0,
    trend TEXT DEFAULT 'new',
    first_crawl_time TIMESTAMP,
    last_crawl_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (platform_id) REFERENCES platforms(id)
);

CREATE INDEX idx_topics_title ON topics(title);
CREATE INDEX idx_topics_platform ON topics(platform_id);
CREATE INDEX idx_topics_heat ON topics(heat DESC);
CREATE INDEX idx_topics_crawl_time ON topics(last_crawl_time);

CREATE TABLE rank_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id TEXT NOT NULL,
    rank INTEGER NOT NULL,
    heat INTEGER,
    crawl_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (topic_id) REFERENCES topics(id)
);

CREATE INDEX idx_rank_topic ON rank_history(topic_id);
CREATE INDEX idx_rank_time ON rank_history(crawl_time);

CREATE TABLE crawl_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    crawl_time TIMESTAMP NOT NULL UNIQUE,
    total_items INTEGER DEFAULT 0,
    status TEXT DEFAULT 'success',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 视频处理表
-- =====================================================
CREATE TABLE videos (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    platform TEXT,
    title TEXT,
    duration INTEGER,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_videos_status ON videos(status);

CREATE TABLE transcripts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    video_id TEXT NOT NULL UNIQUE,
    language TEXT,
    content TEXT,
    optimized TEXT,
    summary TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (video_id) REFERENCES videos(id)
);

-- =====================================================
-- 通知服务表
-- =====================================================
CREATE TABLE notification_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,              -- feishu, dingtalk, wecom, telegram, email
    name TEXT NOT NULL,
    webhook TEXT,
    is_active BOOLEAN DEFAULT 1,
    config TEXT,                     -- JSON 格式配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    title TEXT,
    body TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 任务管理表（扩展现有）
-- =====================================================
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    platform TEXT,
    status TEXT DEFAULT 'pending',
    progress INTEGER DEFAULT 0,
    payload TEXT,                    -- JSON 格式
    result TEXT,                     -- JSON 格式
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(type);
```

---

## 五、开发路线图（统一）

### Phase 1: 数据层优化（1-2 周）⭐ 最高优先级

**目标**：建立稳定的数据存储基础

**任务清单**：
- [ ] 设计并创建 SQLite 数据库 Schema
- [ ] 实现 GORM 模型定义
- [ ] 实现数据库初始化和迁移
- [ ] 实现数据迁移工具（JSON → SQLite）
- [ ] 更新所有存储接口实现
- [ ] 编写单元测试

**验收标准**：
- ✅ 数据库 Schema 创建完成
- ✅ 所有模型定义完成
- ✅ 数据迁移成功，历史数据保留
- ✅ 单元测试通过

### Phase 2: AI 服务统一化（1-2 周）⭐ 最高优先级

**目标**：实现统一的 AI 调用接口

**任务清单**：
- [ ] 实现 `AIServiceConfig` 数据模型
- [ ] 实现 `AIService` 服务层
- [ ] 实现多提供商客户端
  - [ ] OpenRouter
  - [ ] Groq
  - [ ] Google AI
  - [ ] DeepSeek
  - [ ] NVIDIA NIM
- [ ] 实现智能降级和重试
- [ ] 实现客户端缓存
- [ ] 编写单元测试

**验收标准**：
- ✅ AI 服务统一接口完成
- ✅ 至少支持 5 个提供商
- ✅ 智能降级正常工作
- ✅ 单元测试通过

### Phase 3: 热点监控增强（2-3 周）⭐ 高优先级

**目标**：完善热点监控功能

**任务清单**：
- [ ] 实现排名历史记录
- [ ] 实现多维度热度计算
- [ ] 实现趋势分析
- [ ] 实现 RSS 数据源支持
- [ ] 实现 AI 分析功能
- [ ] 实现通知推送系统
- [ ] 编写单元测试

**验收标准**：
- ✅ 排名历史记录正常
- ✅ 热度计算准确
- ✅ 趋势分析功能正常
- ✅ AI 分析结果有价值
- ✅ 通知推送成功

### Phase 4: 视频处理模块（2-3 周）⭐ 中优先级

**目标**：实现视频内容处理能力

**任务清单**：
- [ ] 集成 yt-dlp
- [ ] 集成 Faster-Whisper
- [ ] 实现转录器
- [ ] 实现文本优化器
- [ ] 实现摘要生成器
- [ ] 实现异步任务处理
- [ ] 编写单元测试

**验收标准**：
- ✅ 视频下载成功
- ✅ 转录准确
- ✅ 文本优化有效
- ✅ 摘要生成合理

### Phase 5: MCP Server（1-2 周）⭐ 中优先级

**目标**：让 AI 助手可以直接调用项目功能

**任务清单**：
- [ ] 实现 MCP 协议
- [ ] 实现工具注册机制
- [ ] 封装核心功能为 MCP 工具
  - [ ] 数据查询工具
  - [ ] 分析工具
  - [ ] 通知工具
  - [ ] 视频处理工具
- [ ] 编写 MCP 文档
- [ ] 测试 AI 助手集成

**验收标准**：
- ✅ MCP Server 正常启动
- ✅ 至少实现 10 个工具
- ✅ AI 助手可以成功调用
- ✅ 文档完整

### Phase 6: 前端优化（1-2 周）⭐ 低优先级

**目标**：提供更好的用户体验

**任务清单**：
- [ ] 实现热点趋势图表
- [ ] 实现排名时间线可视化
- [ ] 实现 AI 分析结果展示
- [ ] 实现视频处理进度展示
- [ ] 优化数据筛选和搜索

**验收标准**：
- ✅ 图表展示正常
- ✅ 时间线可视化清晰
- ✅ AI 结果展示美观
- ✅ 搜索功能完善

---

## 六、开发进度记录

> **重要**：完成任务后在此记录，下一个 AI 助手可以继续

### 6.1 已完成任务

#### ✅ 2026-02-20: 架构分析和方案制定
- **任务**：创建统一的项目架构优化方案
- **完成内容**：
  - 分析当前项目架构
  - 整合四个借鉴项目
  - 设计统一架构
  - 制定开发路线图
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文档**：
  - `docs/project-architecture-unified-implementation.md`（本文档）

#### ✅ 2026-02-20: Phase 1 数据层优化
- **任务**：建立稳定的数据存储基础
- **完成内容**：
  - ✅ 设计并创建 SQLite 数据库 Schema
  - ✅ 实现 GORM 模型定义（12 个模型）
  - ✅ 实现数据库初始化和自动迁移
  - ✅ 实现数据迁移工具（JSON → SQLite）
  - ✅ 实现默认数据填充
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文件**：
  - `publisher-core/database/models.go` - 数据模型定义
  - `publisher-core/database/database.go` - 数据库初始化
  - `publisher-core/database/defaults.go` - 默认配置
  - `publisher-core/database/migration.go` - 数据迁移工具
  - `publisher-core/database/hotspot_storage.go` - 热点存储实现

#### ✅ 2026-02-20: Phase 2 AI 服务统一化
- **任务**：实现统一的 AI 调用接口
- **完成内容**：
  - ✅ 实现 `AIServiceConfig` 数据模型
  - ✅ 实现 `UnifiedService` 服务层
  - ✅ 实现多提供商客户端管理
  - ✅ 实现智能降级和重试机制
  - ✅ 实现客户端缓存
  - ✅ 实现 AI 调用历史记录
  - ✅ 实现调用统计功能
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文件**：
  - `publisher-core/ai/unified_service.go` - 统一 AI 服务


#### ✅ 2026-02-20: Phase 3 热点监控增强
- **任务**：完善热点监控功能
- **完成内容**：
  - ✅ 实现排名历史记录功能
  - ✅ 实现多维度热度计算（排名、频次、热度值）
  - ✅ 实现趋势分析功能（up/down/stable/new/hot）
  - ✅ 实现 RSS 数据源支持（支持 RSS 2.0 和 Atom）
  - ✅ 实现 AI 分析功能（热点分析、推荐、分类）
  - ✅ 实现通知推送系统（飞书、钉钉、企业微信、Telegram）
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文件**：
  - `publisher-core/hotspot/enhanced_service.go` - 增强版热点服务
  - `publisher-core/hotspot/sources/rss.go` - RSS 数据源
  - `publisher-core/notify/service.go` - 通知推送服务

#### ✅ 2026-02-20: Phase 4 视频处理模块
- **任务**：实现视频内容处理能力
- **完成内容**：
  - ✅ 集成 yt-dlp 视频下载（支持 30+ 平台）
  - ✅ 实现语音转录功能（Faster-Whisper 集成）
  - ✅ 实现 AI 文本优化器（错字修正、语法修复、分段）
  - ✅ 实现摘要生成器（关键点提取、主题识别）
  - ✅ 实现异步任务处理（任务队列、进度追踪）
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文件**：

#### ✅ 2026-02-20: Phase 5 MCP Server
- **任务**：让 AI 助手可以直接调用项目功能
- **完成内容**：
  - ✅ 实现 MCP 协议基础（JSON-RPC 2.0）
  - ✅ 实现工具注册机制
  - ✅ 实现数据查询工具（热点话题、视频列表）
  - ✅ 实现分析工具（趋势分析、统计信息）
  - ✅ 实现通知工具（发送通知、热点推送）
  - ✅ 实现视频处理工具（任务提交、状态查询）
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文件**：
  - `publisher-core/mcp/server.go` - MCP 服务器
  - `publisher-core/mcp/tools.go` - 工具注册器
  - `publisher-core/video/downloader.go` - 视频下载器
  - `publisher-core/video/transcriber.go` - 语音转录器
  - `publisher-core/video/optimizer.go` - 文本优化器
  - `publisher-core/video/service.go` - 视频处理服务
### 6.2 待完成任务

#### 📋 Phase 1: 数据层优化
- **预计时间**：1-2 周
- **优先级**：最高
- **依赖**：无
- **详细任务**：见第五部分 Phase 1

#### 📋 Phase 2: AI 服务统一化
- **预计时间**：1-2 周
- **优先级**：最高
- **依赖**：Phase 1
- **详细任务**：见第五部分 Phase 2

#### 📋 Phase 3: 热点监控增强
- **预计时间**：2-3 周
- **优先级**：高
- **依赖**：Phase 1, Phase 2
- **详细任务**：见第五部分 Phase 3

#### 📋 Phase 4: 视频处理模块
- **预计时间**：2-3 周
- **优先级**：中
- **依赖**：Phase 2
- **详细任务**：见第五部分 Phase 4

#### 📋 Phase 5: MCP Server
- **预计时间**：1-2 周
- **优先级**：中
- **依赖**：Phase 1, Phase 2, Phase 3
- **详细任务**：见第五部分 Phase 5

#### 📋 Phase 6: 前端优化
- **预计时间**：1-2 周
- **优先级**：低
- **依赖**：Phase 3, Phase 4
- **详细任务**：见第五部分 Phase 6

---

## 七、技术选型总结

### 7.1 后端技术栈

| 技术 | 用途 | 选择理由 | 状态 |
|------|------|---------|------|
| Go 1.21+ | 主要语言 | 高性能、并发友好 | ✅ 已使用 |
| Gorilla Mux | HTTP 路由 | 成熟稳定 | ✅ 已使用 |
| GORM | ORM | 功能强大、易用 | ✅ 已启用 |
| SQLite | 数据库 | 轻量级、无需额外服务 | ✅ 已启用 |
| Rod | 浏览器自动化 | 已有基础 | ✅ 已使用 |
| yt-dlp | 视频下载 | 支持 30+ 平台 | ⚠️ 待集成 |
| Faster-Whisper | 语音转录 | 高精度、多语言 | ⚠️ 待集成 |

### 7.2 AI 技术栈

| 技术 | 用途 | 选择理由 | 状态 |
|------|------|---------|------|
| 统一 AI 接口 | AI 服务管理 | 借鉴 Huobao Drama | ✅ 已实现 |
| OpenRouter | 免费 AI | 多模型、免费额度 | ✅ 已集成 |
| Groq | 快速推理 | 最快响应速度 | ✅ 已集成 |
| Google AI | 大模型 | Gemini Flash | ✅ 已集成 |
| DeepSeek | 国产 AI | 成本低 | ✅ 已集成 |

### 7.3 前端技术栈

| 技术 | 用途 | 选择理由 | 状态 |
|------|------|---------|------|
| React 18 | UI 框架 | 生态丰富 | ✅ 已使用 |
| TypeScript | 类型安全 | 提高代码质量 | ✅ 已使用 |
| Vite | 构建工具 | 快速开发 | ✅ 已使用 |
| Tailwind CSS | 样式框架 | 快速开发 | ✅ 已使用 |
| ECharts | 图表库 | 功能强大 | ⚠️ 待集成 |

---

## 八、参考资源

### 8.1 借鉴项目
- TrendRadar: https://github.com/sansan0/TrendRadar
- Free LLM API Resources: https://github.com/cheahjs/free-llm-api-resources
- AI-Video-Transcriber: https://github.com/wendy7756/AI-Video-Transcriber
- Huobao Drama: https://github.com/chatfire-AI/huobao-drama

### 8.2 技术文档
- LiteLLM: https://docs.litellm.ai/
- Faster-Whisper: https://github.com/SYSTRAN/faster-whisper
- yt-dlp: https://github.com/yt-dlp/yt-dlp
- MCP 协议: https://modelcontextprotocol.io/
- GORM: https://gorm.io/docs/

### 8.3 相关文档
- [AI 服务开发指南](./ai-service-development-guide.md)
- [热点监控借鉴文档](./hot-topics-reference.md)
- [热点监控开发路线图](./hot-topics-roadmap.md)
- [Huobao Drama 项目分析](./huobao-drama-analysis.md)
- [智能启动系统实施报告](./SMART_LAUNCHER_IMPLEMENTATION_REPORT.md)

---

## 九、更新日志

| 日期 | 版本 | 更新内容 |
|------|------|----------|
| 2026-02-20 | v2.0 | 创建统一实施方案，整合当前架构和四个借鉴项目 |
| 2026-02-20 | v1.1 | 新增 Huobao Drama 项目分析 |
| 2026-02-20 | v1.0 | 初始版本，整合三个借鉴项目 |

---

**文档维护**：开发团队
**最后更新**：2026-02-20
**下次更新**：根据开发进度更新"开发进度记录"部分
