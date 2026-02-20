# AI 服务开发指南

> 统一的 AI 服务集成和开发文档
> 
> 文档版本：v1.0
> 创建时间：2026-02-20
> 最后更新：2026-02-20

---

## 📋 文档说明

本文档是 AI 服务开发的**统一入口和权威指南**，整合了：
- 免费和付费 AI API 资源
- LiteLLM 统一接口集成方案
- 多提供商接入指南
- 最佳实践和代码示例

**目标读者**：AI 助手、开发者

**使用方式**：
1. AI 助手阅读本文档了解 AI 服务架构和开发规范
2. 按照本文档实现 AI 功能
3. 完成任务后在"开发进度"部分记录
4. 下一个 AI 助手可以从进度记录继续工作

---

## 一、项目背景

### 1.1 为什么需要统一 AI 接口？

**问题**：
- 不同 AI 提供商 API 格式不统一
- 切换提供商需要大量代码修改
- 免费额度分散，难以充分利用
- 缺少统一的错误处理和重试机制

**解决方案**：
- 采用 **LiteLLM 统一接口**模式
- 封装统一的 AI 服务层
- 支持多提供商动态切换
- 实现智能降级和重试

### 1.2 参考项目

#### TrendRadar (46k+ stars)
- **GitHub**: https://github.com/sansan0/TrendRadar
- **借鉴点**：LiteLLM 统一接口、AI 分析提示词设计
- **详细分析**: [hot-topics-reference.md](./hot-topics-reference.md)

#### Free LLM API Resources (11k+ stars)
- **GitHub**: https://github.com/cheahjs/free-llm-api-resources
- **借鉴点**：免费 AI 资源汇总、提供商限制信息
- **核心价值**：提供 20+ 个免费 AI API 资源

---

## 二、免费 AI 资源汇总

> 数据来源：[free-llm-api-resources](https://github.com/cheahjs/free-llm-api-resources)
> 
> 更新时间：2026-02-20

### 2.1 完全免费提供商

#### 2.1.1 OpenRouter ⭐ 推荐

**官网**: https://openrouter.ai

**限制**：
- 20 请求/分钟
- 50 请求/天
- 充值 $10 可提升至 1000 请求/天

**免费模型**：
```
- google/gemma-3-12b-it:free
- google/gemma-3-27b-it:free
- google/gemma-3-4b-it:free
- meta-llama/llama-3.2-3b-instruct:free
- meta-llama/llama-3.3-70b-instruct:free
- mistralai/mistral-small-3.1-24b-instruct:free
- deepseek/deepseek-r1-0528:free
- qwen/qwen-2.5-72b-instruct:free
```

**API 格式**：OpenAI 兼容
```bash
curl https://openrouter.ai/api/v1/chat/completions \
  -H "Authorization: Bearer $OPENROUTER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta-llama/llama-3.3-70b-instruct:free",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ 测试和开发
- ✅ 低频调用
- ✅ 多模型对比

#### 2.1.2 Google AI Studio ⭐ 推荐

**官网**: https://aistudio.google.com

**限制**：
| 模型 | Tokens/分钟 | 请求/天 | 请求/分钟 |
|------|------------|--------|----------|
| Gemini 3 Flash | 250,000 | 20 | 5 |
| Gemini 2.5 Flash | 250,000 | 20 | 5 |
| Gemma 3 27B | 15,000 | 14,400 | 30 |
| Gemma 3 12B | 15,000 | 14,400 | 30 |

**免费模型**：
```
- gemini-3-flash
- gemini-2.5-flash
- gemini-2.5-flash-lite
- gemma-3-27b-it
- gemma-3-12b-it
- gemma-3-4b-it
- gemma-3-1b-it
```

**API 格式**：自定义（但兼容 OpenAI）
```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=$GOOGLE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{"parts": [{"text": "Hello"}]}]
  }'
```

**适用场景**：
- ✅ 高频调用（Gemma 系列）
- ✅ 长文本处理（Gemini Flash）
- ⚠️ 注意：数据可能用于训练（非欧盟地区）

#### 2.1.3 NVIDIA NIM

**官网**: https://build.nvidia.com/explore/discover

**限制**：5000 tokens/分钟

**免费模型**：
```
- meta/llama-3.1-405b-instruct
- meta/llama-3.1-70b-instruct
- meta/llama-3.1-8b-instruct
- meta/llama-3.2-11b-vision-instruct
- meta/llama-3.2-3b-instruct
- meta/llama-3.2-1b-instruct
- mistralai/mistral-large
- mistralai/mixtral-8x7b-instruct
- google/gemma-2-27b-it
- google/gemma-2-9b-it
```

**API 格式**：OpenAI 兼容
```bash
curl https://integrate.api.nvidia.com/v1/chat/completions \
  -H "Authorization: Bearer $NVIDIA_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "meta/llama-3.1-405b-instruct",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ 大模型推理（Llama 405B）
- ✅ 视觉模型
- ✅ 高质量输出

#### 2.1.4 Groq ⚡ 最快推理

**官网**: https://console.groq.com

**限制**：
- 30 请求/分钟（免费）
- 7000 tokens/分钟

**免费模型**：
```
- llama-3.3-70b-versatile
- llama-3.1-8b-instant
- mixtral-8x7b-32768
- gemma2-9b-it
```

**API 格式**：OpenAI 兼容
```bash
curl https://api.groq.com/openai/v1/chat/completions \
  -H "Authorization: Bearer $GROQ_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.3-70b-versatile",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ 实时对话
- ✅ 快速响应
- ✅ 流式输出

#### 2.1.5 Mistral

**官网**: https://console.mistral.ai

**限制**：
- 1 请求/秒
- 500,000 tokens/月

**免费模型**：
```
- mistral-small-latest
- codestral-latest
- mistral-7b-instruct
```

**API 格式**：OpenAI 兼容
```bash
curl https://api.mistral.ai/v1/chat/completions \
  -H "Authorization: Bearer $MISTRAL_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "mistral-small-latest",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ 代码生成（Codestral）
- ✅ 欧洲合规需求

#### 2.1.6 Cloudflare Workers AI

**官网**: https://developers.cloudflare.com/workers-ai

**限制**：10,000 neurons/天

**免费模型**：
```
- @cf/meta/llama-3.1-8b-instruct
- @cf/meta/llama-3.2-3b-instruct
- @cf/mistral/mistral-7b-instruct
- @cf/google/gemma-3-12b-it
- @cf/qwen/qwen-2.5-72b-instruct
- @cf/deepseek/r1-distill-qwen-32b
```

**API 格式**：自定义
```bash
curl https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/ai/run/@cf/meta/llama-3.1-8b-instruct \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ 边缘计算
- ✅ Cloudflare 生态集成

#### 2.1.7 GitHub Models

**官网**: https://github.com/marketplace/models

**限制**：
- 低速率限制（具体未公开）
- 需 GitHub 账号

**免费模型**：
```
- gpt-4o
- gpt-4o-mini
- o1-preview
- o1-mini
- phi-3.5-mini
- phi-3-medium
- meta-llama-3.1-405b-instruct
- meta-llama-3.1-70b-instruct
- mistral-large
- cohere-command-r
```

**API 格式**：OpenAI 兼容
```bash
curl https://models.inference.ai.azure.com/chat/completions \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**适用场景**：
- ✅ GitHub 生态集成
- ✅ 开发者友好

#### 2.1.8 Cohere

**官网**: https://cohere.com

**限制**：
- 1000 次/月（免费试用）
- 需信用卡验证

**免费模型**：
```
- command
- command-light
- command-r
- command-r-plus
```

**API 格式**：自定义
```bash
curl https://api.cohere.ai/v1/chat \
  -H "Authorization: Bearer $COHERE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "command-r-plus",
    "message": "Hello"
  }'
```

**适用场景**：
- ✅ 企业级应用
- ✅ RAG 应用

### 2.2 试用额度提供商

#### 2.2.1 Fireworks AI

**官网**: https://fireworks.ai

**试用额度**：$1 免费额度

**模型**：
```
- llama-3.1-405b-instruct
- llama-3.1-70b-instruct
- qwen2.5-72b-instruct
- mixtral-8x7b-instruct
```

**API 格式**：OpenAI 兼容

#### 2.2.2 Together AI

**官网**: https://together.ai

**试用额度**：$1 免费额度

**模型**：
```
- meta-llama/Llama-3-70b-chat-hf
- mistralai/Mixtral-8x7B-Instruct-v0.1
- togethercomputer/CodeLlama-34b-Instruct
```

**API 格式**：OpenAI 兼容

#### 2.2.3 其他试用提供商

| 提供商 | 试用额度 | 特点 |
|--------|---------|------|
| Baseten | $10 | 自定义模型部署 |
| Nebius | $10 | 云计算平台 |
| Novita | $5 | 多模型支持 |
| AI21 | $10 | Jurassic 系列 |
| Upstage | $10 | 韩国模型 |
| NLP Cloud | $10 | 企业级 |
| Modal | $10 | 无服务器部署 |
| Hyperbolic | $10 | 去中心化 |
| SambaNova | $10 | 企业 AI |

---

## 三、LiteLLM 统一接口方案

### 3.1 LiteLLM 简介

**LiteLLM** 是一个统一的 LLM 接口库，提供：
- 统一的 API 调用格式（OpenAI 格式）
- 支持 100+ AI 提供商
- 自动重试和降级
- 成本追踪
- 缓存支持

**GitHub**: https://github.com/BerriAI/litellm

### 3.2 为什么选择 LiteLLM 模式？

| 特性 | 直接调用 | LiteLLM 模式 |
|------|---------|-------------|
| API 格式 | 每个提供商不同 | 统一 OpenAI 格式 |
| 切换提供商 | 需修改代码 | 仅改配置 |
| 错误处理 | 手动实现 | 自动处理 |
| 重试机制 | 手动实现 | 内置支持 |
| 成本追踪 | 手动实现 | 自动记录 |
| 免费额度利用 | 难以管理 | 统一管理 |

### 3.3 Go 实现方案

由于项目使用 Go 语言，我们需要实现类似 LiteLLM 的统一接口。

#### 3.3.1 核心接口设计

```go
// ai/provider/interface.go

package provider

import "context"

// GenerateOptions 生成选项
type GenerateOptions struct {
    Model       string
    Messages    []Message
    Temperature float64
    MaxTokens   int
    Stream      bool
}

// Message 消息
type Message struct {
    Role    Role
    Content string
}

// Role 角色
type Role string

const (
    RoleSystem Role = "system"
    RoleUser   Role = "user"
    RoleAssistant Role = "assistant"
)

// GenerateResult 生成结果
type GenerateResult struct {
    Content      string
    TokensUsed   int
    Model        string
    Provider     string
    FinishReason string
}

// Provider AI 提供商接口
type Provider interface {
    // Name 提供商名称
    Name() string
    
    // Generate 生成内容
    Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error)
    
    // GenerateStream 流式生成
    GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error)
    
    // ListModels 列出可用模型
    ListModels() []string
    
    // IsAvailable 检查是否可用
    IsAvailable() bool
}
```

#### 3.3.2 OpenAI 兼容提供商实现

```go
// ai/provider/openai_compatible.go

package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

// OpenAICompatibleProvider OpenAI 兼容提供商
type OpenAICompatibleProvider struct {
    name     string
    apiBase  string
    apiKey   string
    models   []string
    client   *http.Client
}

func NewOpenAICompatibleProvider(name, apiBase, apiKey string, models []string) *OpenAICompatibleProvider {
    return &OpenAICompatibleProvider{
        name:    name,
        apiBase: apiBase,
        apiKey:  apiKey,
        models:  models,
        client:  &http.Client{},
    }
}

func (p *OpenAICompatibleProvider) Name() string {
    return p.name
}

func (p *OpenAICompatibleProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
    // 构建请求
    reqBody := map[string]interface{}{
        "model": opts.Model,
        "messages": opts.Messages,
    }
    
    if opts.Temperature > 0 {
        reqBody["temperature"] = opts.Temperature
    }
    if opts.MaxTokens > 0 {
        reqBody["max_tokens"] = opts.MaxTokens
    }
    
    body, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }
    
    // 发送请求
    url := fmt.Sprintf("%s/chat/completions", p.apiBase)
    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
    }
    
    // 解析响应
    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
            FinishReason string `json:"finish_reason"`
        } `json:"choices"`
        Usage struct {
            TotalTokens int `json:"total_tokens"`
        } `json:"usage"`
        Model string `json:"model"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    
    if len(result.Choices) == 0 {
        return nil, fmt.Errorf("no choices in response")
    }
    
    return &GenerateResult{
        Content:      result.Choices[0].Message.Content,
        TokensUsed:   result.Usage.TotalTokens,
        Model:        result.Model,
        Provider:     p.name,
        FinishReason: result.Choices[0].FinishReason,
    }, nil
}

func (p *OpenAICompatibleProvider) ListModels() []string {
    return p.models
}

func (p *OpenAICompatibleProvider) IsAvailable() bool {
    return p.apiKey != ""
}
```

#### 3.3.3 提供商注册

```go
// ai/provider/registry.go

package provider

import (
    "os"
    "sync"
)

// Registry 提供商注册表
type Registry struct {
    providers map[string]Provider
    priority  []string // 优先级顺序
    mu        sync.RWMutex
}

func NewRegistry() *Registry {
    return &Registry{
        providers: make(map[string]Provider),
        priority:  []string{},
    }
}

// Register 注册提供商
func (r *Registry) Register(p Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.providers[p.Name()] = p
    r.priority = append(r.priority, p.Name())
}

// Get 获取提供商
func (r *Registry) Get(name string) (Provider, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    p, ok := r.providers[name]
    return p, ok
}

// GetAvailable 获取可用的提供商（按优先级）
func (r *Registry) GetAvailable() []Provider {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var available []Provider
    for _, name := range r.priority {
        if p, ok := r.providers[name]; ok && p.IsAvailable() {
            available = append(available, p)
        }
    }
    return available
}

// InitFromEnv 从环境变量初始化
func (r *Registry) InitFromEnv() {
    // OpenRouter
    if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
        r.Register(NewOpenAICompatibleProvider(
            "openrouter",
            "https://openrouter.ai/api/v1",
            key,
            []string{
                "meta-llama/llama-3.3-70b-instruct:free",
                "google/gemma-3-27b-it:free",
                "deepseek/deepseek-r1-0528:free",
            },
        ))
    }
    
    // Groq
    if key := os.Getenv("GROQ_API_KEY"); key != "" {
        r.Register(NewOpenAICompatibleProvider(
            "groq",
            "https://api.groq.com/openai/v1",
            key,
            []string{
                "llama-3.3-70b-versatile",
                "llama-3.1-8b-instant",
                "mixtral-8x7b-32768",
            },
        ))
    }
    
    // NVIDIA NIM
    if key := os.Getenv("NVIDIA_API_KEY"); key != "" {
        r.Register(NewOpenAICompatibleProvider(
            "nvidia",
            "https://integrate.api.nvidia.com/v1",
            key,
            []string{
                "meta/llama-3.1-405b-instruct",
                "meta/llama-3.1-70b-instruct",
            },
        ))
    }
    
    // Mistral
    if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
        r.Register(NewOpenAICompatibleProvider(
            "mistral",
            "https://api.mistral.ai/v1",
            key,
            []string{
                "mistral-small-latest",
                "codestral-latest",
            },
        ))
    }
    
    // Google AI Studio
    if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
        r.Register(NewGoogleAIProvider(key))
    }
    
    // DeepSeek
    if key := os.Getenv("DEEPSEEK_API_KEY"); key != "" {
        r.Register(NewOpenAICompatibleProvider(
            "deepseek",
            "https://api.deepseek.com/v1",
            key,
            []string{
                "deepseek-chat",
                "deepseek-coder",
            },
        ))
    }
}
```

#### 3.3.4 统一服务层

```go
// ai/service.go

package ai

import (
    "context"
    "fmt"
    
    "publisher-core/ai/provider"
)

// Service AI 服务
type Service struct {
    registry *provider.Registry
    config   *Config
}

type Config struct {
    DefaultProvider string
    DefaultModel    string
    MaxRetries      int
    FallbackEnabled bool
}

func NewService(registry *provider.Registry, config *Config) *Service {
    return &Service{
        registry: registry,
        config:   config,
    }
}

// Generate 生成内容（自动选择提供商）
func (s *Service) Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
    // 如果指定了提供商，直接使用
    if opts.Model != "" && s.config.DefaultProvider != "" {
        p, ok := s.registry.Get(s.config.DefaultProvider)
        if !ok {
            return nil, fmt.Errorf("provider %s not found", s.config.DefaultProvider)
        }
        return s.generateWithRetry(ctx, p, opts)
    }
    
    // 否则按优先级尝试
    providers := s.registry.GetAvailable()
    if len(providers) == 0 {
        return nil, fmt.Errorf("no available providers")
    }
    
    var lastErr error
    for _, p := range providers {
        result, err := s.generateWithRetry(ctx, p, opts)
        if err == nil {
            return result, nil
        }
        lastErr = err
        
        // 如果不启用降级，直接返回错误
        if !s.config.FallbackEnabled {
            break
        }
    }
    
    return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// generateWithRetry 带重试的生成
func (s *Service) generateWithRetry(ctx context.Context, p provider.Provider, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
    var lastErr error
    
    for i := 0; i <= s.config.MaxRetries; i++ {
        result, err := p.Generate(ctx, opts)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 检查是否可重试
        if !s.isRetryableError(err) {
            break
        }
    }
    
    return nil, lastErr
}

// isRetryableError 判断是否可重试
func (s *Service) isRetryableError(err error) bool {
    // 网络错误、超时、429 等可重试
    // 其他错误不重试
    return true // 简化实现
}
```

### 3.4 配置管理

```yaml
# config/ai.yaml

ai:
  # 默认提供商
  default_provider: "openrouter"
  
  # 默认模型
  default_model: "meta-llama/llama-3.3-70b-instruct:free"
  
  # 重试次数
  max_retries: 3
  
  # 启用降级
  fallback_enabled: true
  
  # 提供商优先级
  priority:
    - openrouter  # 免费，多模型
    - groq        # 最快
    - nvidia      # 大模型
    - mistral     # 代码
    - google      # Gemini
    - deepseek    # 备用

# 提供商配置
providers:
  openrouter:
    api_key: "${OPENROUTER_API_KEY}"
    models:
      - "meta-llama/llama-3.3-70b-instruct:free"
      - "google/gemma-3-27b-it:free"
      - "deepseek/deepseek-r1-0528:free"
  
  groq:
    api_key: "${GROQ_API_KEY}"
    models:
      - "llama-3.3-70b-versatile"
      - "llama-3.1-8b-instant"
  
  nvidia:
    api_key: "${NVIDIA_API_KEY}"
    models:
      - "meta/llama-3.1-405b-instruct"
      - "meta/llama-3.1-70b-instruct"
```

---

## 四、开发进度记录

> **重要**：完成任务后在此记录，下一个 AI 助手可以继续

### 4.1 已完成任务

#### ✅ 2026-02-20: 文档创建
- **任务**：创建统一的 AI 服务开发指南
- **完成内容**：
  - 整合 free-llm-api-resources 项目分析
  - 整合 TrendRadar LiteLLM 方案
  - 创建统一的接口设计
  - 提供完整的代码示例
- **负责人**：AI 助手
- **状态**：✅ 完成
- **产出文档**：
  - `docs/ai-service-development-guide.md`（本文档）
  - `docs/hot-topics-reference.md`
  - `docs/hot-topics-roadmap.md`

### 4.2 待完成任务

#### 📋 Phase 1: 基础框架实现
- [ ] 实现 Provider 接口
- [ ] 实现 OpenAI 兼容提供商
- [ ] 实现提供商注册表
- [ ] 实现统一服务层
- [ ] 编写单元测试

**预计时间**：1 周
**优先级**：高
**依赖**：无

#### 📋 Phase 2: 提供商集成
- [ ] 集成 OpenRouter
- [ ] 集成 Groq
- [ ] 集成 NVIDIA NIM
- [ ] 集成 Mistral
- [ ] 集成 Google AI Studio
- [ ] 集成 DeepSeek

**预计时间**：1 周
**优先级**：高
**依赖**：Phase 1 完成

#### 📋 Phase 3: 高级功能
- [ ] 实现流式输出
- [ ] 实现智能降级
- [ ] 实现成本追踪
- [ ] 实现缓存机制
- [ ] 实现速率限制

**预计时间**：1 周
**优先级**：中
**依赖**：Phase 2 完成

#### 📋 Phase 4: 热点监控集成
- [ ] 集成到热点监控 AI 分析
- [ ] 优化 AI 分析提示词
- [ ] 实现多模型对比
- [ ] 实现分析结果缓存

**预计时间**：1 周
**优先级**：中
**依赖**：Phase 3 完成

---

## 五、最佳实践

### 5.1 提供商选择策略

```go
// 根据任务类型选择提供商
func SelectProvider(taskType string) string {
    switch taskType {
    case "code_generation":
        return "mistral"  // Codestral
    case "fast_response":
        return "groq"     // 最快
    case "large_model":
        return "nvidia"   // Llama 405B
    case "free_tier":
        return "openrouter" // 免费
    default:
        return "openrouter" // 默认
    }
}
```

### 5.2 错误处理

```go
func (s *Service) Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error) {
    result, err := s.generateWithRetry(ctx, p, opts)
    if err != nil {
        // 记录错误
        log.Errorf("AI generation failed: %v", err)
        
        // 降级到其他提供商
        if s.config.FallbackEnabled {
            return s.fallback(ctx, opts)
        }
        
        return nil, err
    }
    return result, nil
}
```

### 5.3 成本优化

```go
// 使用免费额度优先
func (s *Service) selectFreeProvider() Provider {
    providers := s.registry.GetAvailable()
    
    // 优先选择免费模型
    for _, p := range providers {
        if strings.Contains(p.Name(), "free") {
            return p
        }
    }
    
    // 否则选择第一个可用的
    if len(providers) > 0 {
        return providers[0]
    }
    
    return nil
}
```

### 5.4 速率限制

```go
type RateLimiter struct {
    requests map[string]*rate.Limiter
    mu       sync.RWMutex
}

func (l *RateLimiter) Wait(ctx context.Context, provider string) error {
    l.mu.RLock()
    limiter, ok := l.requests[provider]
    l.mu.RUnlock()
    
    if !ok {
        return nil
    }
    
    return limiter.Wait(ctx)
}
```

---

## 六、常见问题

### Q1: 如何获取免费 API Key？

**A**: 按以下步骤：
1. OpenRouter: https://openrouter.ai/keys
2. Groq: https://console.groq.com/keys
3. NVIDIA: https://build.nvidia.com/api-key
4. Mistral: https://console.mistral.ai/api-keys
5. Google AI: https://aistudio.google.com/apikey

### Q2: 免费额度用完了怎么办？

**A**: 
1. 切换到其他免费提供商
2. 使用试用额度提供商
3. 考虑付费（OpenRouter $10 可用很久）

### Q3: 如何选择合适的模型？

**A**: 
- **代码生成**: Codestral, DeepSeek Coder
- **快速响应**: Groq (Llama 3.3 70B)
- **高质量输出**: NVIDIA (Llama 405B)
- **长文本**: Google Gemini Flash
- **免费测试**: OpenRouter 免费模型

### Q4: 如何处理 API 限流？

**A**: 
1. 实现指数退避重试
2. 使用多个提供商轮换
3. 实现请求队列
4. 缓存重复请求

---

## 七、参考资源

### 7.1 官方文档
- LiteLLM: https://docs.litellm.ai/
- OpenRouter: https://openrouter.ai/docs
- Groq: https://console.groq.com/docs
- NVIDIA NIM: https://build.nvidia.com/docs

### 7.2 参考项目
- TrendRadar: https://github.com/sansan0/TrendRadar
- Free LLM API Resources: https://github.com/cheahjs/free-llm-api-resources
- LiteLLM: https://github.com/BerriAI/litellm

### 7.3 相关文档
- [热点监控借鉴文档](./hot-topics-reference.md)
- [热点监控开发路线图](./hot-topics-roadmap.md)

---

## 八、更新日志

| 日期 | 版本 | 更新内容 |
|------|------|----------|
| 2026-02-20 | v1.0 | 初始版本，整合免费 AI 资源和 LiteLLM 方案 |

---

**文档维护**：开发团队
**最后更新**：2026-02-20
**下次更新**：根据开发进度更新"开发进度记录"部分
