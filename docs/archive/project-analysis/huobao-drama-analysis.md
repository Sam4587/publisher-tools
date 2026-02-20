# Huobao Drama 项目分析

> AI 短剧自动化生产平台深度分析
> 
> 文档版本：v1.0
> 创建时间：2026-02-20

---

## 一、项目概述

### 1.1 基本信息
- **GitHub**: https://github.com/chatfire-AI/huobao-drama
- **Stars**: 7,664
- **语言**: Go + Vue3
- **许可证**: CC BY-NC-SA 4.0
- **核心功能**: AI 短剧自动化生产平台

### 1.2 项目定位
"一句话生成完整短剧" - 从剧本到成片的全流程自动化

---

## 二、核心特性

### 2.1 全栈架构
- **后端**: Go 1.23+（高性能、并发友好）
- **前端**: Vue 3 + TypeScript + Vite
- **数据库**: SQLite / PostgreSQL
- **ORM**: GORM

### 2.2 AI 多模态生成

#### 文本生成
- 剧本生成
- 角色提取
- 场景描述
- 对话生成

#### 图像生成
- 角色形象设计
- 场景画面生成
- 道具设计

#### 视频生成
- AI 视频合成
- 视频剪辑
- 特效添加

#### 音频生成
- 配音生成
- 背景音乐
- 音效添加

### 2.3 统一 AI 服务管理

**支持的提供商**：
- OpenAI
- Google AI
- Doubao（字节跳动）
- Chatfire
- Volcengine（火山引擎）
- 自定义提供商

**服务类型**：
- text（文本生成）
- image（图像生成）
- video（视频生成）

### 2.4 完整工作流

```
剧本生成 → 角色提取 → 场景设计 → 分镜制作 → 视频合成
    ↓           ↓           ↓           ↓           ↓
  AI生成     AI生成      AI生成      AI生成      AI合成
```

### 2.5 任务管理
- 异步任务处理
- 实时进度追踪
- 任务队列管理
- 失败重试机制

### 2.6 资源管理
- 角色库
- 道具库
- 场景库
- 素材库

---

## 三、技术架构

### 3.1 项目结构

```
huobao-drama/
├── api/                        # API 层
│   ├── handlers/              # API 处理器
│   │   ├── ai_config.go              # AI 配置管理
│   │   ├── script_generation.go      # 剧本生成
│   │   ├── character_library.go      # 角色库
│   │   ├── character_library_gen.go  # 角色生成
│   │   ├── image_generation.go       # 图像生成
│   │   ├── video_generation.go       # 视频生成
│   │   ├── video_merge.go            # 视频合并
│   │   ├── audio_extraction.go       # 音频提取
│   │   ├── storyboard.go             # 分镜管理
│   │   ├── scene.go                  # 场景管理
│   │   ├── prop.go                   # 道具管理
│   │   ├── task.go                   # 任务管理
│   │   ├── upload.go                 # 文件上传
│   │   └── settings.go               # 系统设置
│   ├── middlewares/           # 中间件
│   │   ├── cors.go                   # CORS 处理
│   │   ├── logger.go                 # 日志记录
│   │   └── ratelimit.go              # 速率限制
│   └── routes/                # 路由配置
│       └── routes.go
├── application/                # 应用层
│   └── services/              # 业务服务
│       ├── ai_service.go                    # AI 服务统一管理
│       ├── script_generation_service.go     # 剧本生成
│       ├── character_library_service.go     # 角色库服务
│       ├── image_generation_service.go      # 图像生成
│       ├── video_generation_service.go      # 视频生成
│       ├── video_merge_service.go           # 视频合并
│       ├── audio_extraction_service.go      # 音频提取
│       ├── storyboard_service.go            # 分镜服务
│       ├── storyboard_composition_service.go # 分镜合成
│       ├── frame_prompt_service.go          # 帧提示词服务
│       ├── prop_service.go                  # 道具服务
│       ├── task_service.go                  # 任务服务
│       ├── upload_service.go                # 上传服务
│       ├── prompt_i18n.go                   # 提示词国际化
│       └── data_migration_service.go        # 数据迁移
├── domain/                     # 领域层
│   └── models/                # 数据模型
│       ├── ai_service_config.go
│       ├── drama.go
│       ├── character.go
│       ├── scene.go
│       ├── storyboard.go
│       ├── task.go
│       └── ...
├── pkg/                        # 基础设施层
│   ├── ai/                    # AI 客户端封装
│   ├── config/                # 配置管理
│   ├── logger/                # 日志系统
│   └── utils/                 # 工具函数
└── web/                        # 前端代码
    └── src/
        ├── components/        # 组件
        ├── views/             # 页面
        ├── stores/            # 状态管理
        └── i18n/              # 国际化
```

### 3.2 架构分层

```
┌─────────────────────────────────────────┐
│          前端层 (Vue3 + TypeScript)       │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ 剧本编辑 │  │ 角色管理 │  │ 视频预览│ │
│  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────┘
                    ↓ HTTP/REST
┌─────────────────────────────────────────┐
│              API 层 (Go)                 │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ Handlers │  │Middleware│  │ Routes │ │
│  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│            应用层 (Services)             │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ AI服务   │  │ 任务服务 │  │ 资源服务│ │
│  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│            领域层 (Models)               │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ Drama    │  │Character │  │ Task   │ │
│  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         基础设施层 (pkg/)                │
│  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │ AI客户端 │  │ 数据库   │  │ 工具   │ │
│  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────┘
```

---

## 四、核心代码分析

### 4.1 AI 服务统一管理

**数据模型**：
```go
type AIServiceConfig struct {
    ID            uint           `gorm:"primaryKey" json:"id"`
    ServiceType   string         `json:"service_type"`   // text, image, video
    Name          string         `json:"name"`
    Provider      string         `json:"provider"`       // openai, google, doubao
    BaseURL       string         `json:"base_url"`
    APIKey        string         `json:"api_key"`
    Model         string         `json:"model"`
    Endpoint      string         `json:"endpoint"`
    QueryEndpoint string         `json:"query_endpoint"`
    Priority      int            `json:"priority"`
    IsDefault     bool           `json:"is_default"`
    IsActive      bool           `json:"is_active"`
    Settings      datatypes.JSON `json:"settings"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
}
```

**服务实现**：
```go
type AIService struct {
    db  *gorm.DB
    log *logger.Logger
}

// 创建 AI 配置
func (s *AIService) CreateConfig(req CreateAIConfigRequest) (*models.AIServiceConfig, error) {
    var endpoint, queryEndpoint string
    
    // 根据提供商类型设置端点
    switch req.Provider {
    case "google":
        if req.ServiceType == "text" {
            endpoint = "/v1beta/models/{model}:generateContent"
        }
    case "openai":
        if req.ServiceType == "text" {
            endpoint = "/chat/completions"
        } else if req.ServiceType == "image" {
            endpoint = "/images/generations"
        } else if req.ServiceType == "video" {
            endpoint = "/videos"
            queryEndpoint = "/videos/{taskId}"
        }
    case "chatfire":
        if req.ServiceType == "text" {
            endpoint = "/chat/completions"
        } else if req.ServiceType == "image" {
            endpoint = "/images/generations"
        } else if req.ServiceType == "video" {
            endpoint = "/video/generations"
            queryEndpoint = "/video/task/{taskId}"
        }
    case "doubao", "volcengine":
        if req.ServiceType == "video" {
            endpoint = "/contents/generations/tasks"
            queryEndpoint = "/generations/tasks/{taskId}"
        }
    default:
        if req.ServiceType == "text" {
            endpoint = "/chat/completions"
        }
    }
    
    config := &models.AIServiceConfig{
        ServiceType:   req.ServiceType,
        Provider:      req.Provider,
        BaseURL:       req.BaseURL,
        APIKey:        req.APIKey,
        Model:         req.Model,
        Endpoint:      endpoint,
        QueryEndpoint: queryEndpoint,
        Priority:      req.Priority,
        IsDefault:     req.IsDefault,
        IsActive:      true,
    }
    
    return config, s.db.Create(config).Error
}

// 获取默认配置
func (s *AIService) GetDefaultConfig(serviceType string) (*models.AIServiceConfig, error) {
    var config models.AIServiceConfig
    err := s.db.Where("service_type = ? AND is_default = ? AND is_active = ?", 
        serviceType, true, true).
        Order("priority desc").
        First(&config).Error
    if err != nil {
        return nil, err
    }
    return &config, nil
}

// 获取指定模型的客户端
func (s *AIService) GetAIClientForModel(serviceType, model string) (*ai.Client, error) {
    var config models.AIServiceConfig
    err := s.db.Where("service_type = ? AND model = ? AND is_active = ?", 
        serviceType, model, true).
        First(&config).Error
    if err != nil {
        return nil, err
    }
    
    return ai.NewClient(&ai.ClientConfig{
        BaseURL: config.BaseURL,
        APIKey:  config.APIKey,
        Model:   config.Model,
    }), nil
}
```

### 4.2 剧本生成服务

```go
type ScriptGenerationService struct {
    db          *gorm.DB
    aiService   *AIService
    log         *logger.Logger
    config      *config.Config
    promptI18n  *PromptI18n
    taskService *TaskService
}

// 生成角色
func (s *ScriptGenerationService) GenerateCharacters(req GenerateCharactersRequest, taskID string) {
    // 1. 获取 drama 信息
    var drama models.Drama
    if err := s.db.Where("id = ?", req.DramaID).First(&drama).Error; err != nil {
        s.taskService.UpdateTaskStatus(taskID, "failed", 0, "剧本信息不存在")
        return
    }
    
    // 2. 构建提示词
    systemPrompt := s.promptI18n.GetCharacterExtractionPrompt(drama.Style)
    userPrompt := s.promptI18n.FormatUserPrompt("character_request", req.Outline, req.Count)
    
    // 3. 调用 AI 生成
    var text string
    var err error
    if req.Model != "" {
        client, getErr := s.aiService.GetAIClientForModel("text", req.Model)
        if getErr != nil {
            text, err = s.aiService.GenerateText(userPrompt, systemPrompt)
        } else {
            text, err = client.GenerateText(userPrompt, systemPrompt)
        }
    } else {
        text, err = s.aiService.GenerateText(userPrompt, systemPrompt)
    }
    
    if err != nil {
        s.taskService.UpdateTaskStatus(taskID, "failed", 0, "AI生成失败: "+err.Error())
        return
    }
    
    // 4. 解析 AI 返回
    var result []struct {
        Name        string `json:"name"`
        Role        string `json:"role"`
        Description string `json:"description"`
        Personality string `json:"personality"`
        Appearance  string `json:"appearance"`
        VoiceStyle  string `json:"voice_style"`
    }
    
    if err := utils.SafeParseAIJSON(text, &result); err != nil {
        s.taskService.UpdateTaskStatus(taskID, "failed", 0, "解析AI返回结果失败")
        return
    }
    
    // 5. 保存到数据库
    var characters []models.Character
    for _, char := range result {
        character := models.Character{
            DramaID:     dramaID,
            Name:        char.Name,
            Role:        char.Role,
            Description: char.Description,
            Personality: char.Personality,
            Appearance:  char.Appearance,
            VoiceStyle:  char.VoiceStyle,
        }
        s.db.Create(&character)
        characters = append(characters, character)
    }
    
    // 6. 更新任务状态
    s.taskService.UpdateTaskStatus(taskID, "completed", 100, "生成成功")
}
```

### 4.3 提示词国际化

```go
type PromptI18n struct {
    prompts map[string]map[string]string
}

func (p *PromptI18n) GetCharacterExtractionPrompt(style string) string {
    return p.prompts["character_extraction"][style]
}

func (p *PromptI18n) FormatUserPrompt(template string, args ...interface{}) string {
    tmpl := p.prompts["user_prompts"][template]
    return fmt.Sprintf(tmpl, args...)
}
```

---

## 五、与当前项目的整合建议

### 5.1 高度契合点

1. **技术栈一致**
   - Go 后端 + Vue3 前端
   - GORM ORM
   - SQLite 数据库

2. **AI 服务统一管理**
   - 与我们的 LiteLLM 方案高度一致
   - 支持多种提供商
   - 统一的 API 接口

3. **多模态生成**
   - 文本、图像、视频生成能力
   - 可用于内容创作

4. **任务管理**
   - 异步任务处理机制
   - 进度追踪

### 5.2 可直接借鉴

1. **AI 配置管理数据模型**
   - `AIServiceConfig` 模型设计
   - 多提供商端点配置
   - 优先级和默认配置

2. **AI 服务统一接口**
   - `AIService` 服务实现
   - 多提供商适配逻辑
   - 灵活的模型选择

3. **异步任务处理**
   - 任务状态管理
   - 进度追踪
   - 错误处理

4. **提示词国际化**
   - 多语言提示词支持
   - 提示词模板管理

5. **AI 返回结果解析**
   - `SafeParseAIJSON` 工具函数
   - 错误处理和降级

### 5.3 整合方案

#### 方案 1: 直接借鉴代码结构
- 借鉴 `AIServiceConfig` 数据模型
- 借鉴 `AIService` 服务实现
- 借鉴任务管理机制

#### 方案 2: 集成到现有架构
- 在现有 AI 服务基础上增加多模态支持
- 扩展任务管理支持视频生成
- 增加资源管理模块

---

## 六、参考资源

### 6.1 项目链接
- GitHub: https://github.com/chatfire-AI/huobao-drama
- 在线演示: 暂无

### 6.2 相关文档
- [项目架构优化文档](./project-architecture-optimization.md)
- [AI 服务开发指南](./ai-service-development-guide.md)

---

**文档维护**：开发团队
**最后更新**：2026-02-20
