package provider

import (
	"context"
	"time"
)

type ProviderType string

const (
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderGoogle     ProviderType = "google"
	ProviderGroq       ProviderType = "groq"
	ProviderDeepSeek   ProviderType = "deepseek"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type GenerateOptions struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

type GenerateResult struct {
	Content      string    `json:"content"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	FinishedAt   time.Time `json:"finished_at"`
}

type Provider interface {
	Name() ProviderType
	Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error)
	GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error)
	Models() []string
	DefaultModel() string
}

type ProviderConfig struct {
	Type     ProviderType `json:"type"`
	APIKey   string       `json:"api_key"`
	BaseURL  string       `json:"base_url"`
	Model    string       `json:"model"`
	Enabled  bool         `json:"enabled"`
	Priority int          `json:"priority"`
}

type ContentTask string

const (
	TaskContentGenerate  ContentTask = "generate"
	TaskContentRewrite   ContentTask = "rewrite"
	TaskContentExpand    ContentTask = "expand"
	TaskContentSummarize ContentTask = "summarize"
	TaskContentAnalyze   ContentTask = "analyze"
	TaskContentAudit     ContentTask = "audit"
)

type ContentRequest struct {
	Task     ContentTask `json:"task"`
	Input    string      `json:"input"`
	Context  string      `json:"context,omitempty"`
	Style    string      `json:"style,omitempty"`
	Length   int         `json:"length,omitempty"`
	Platform string      `json:"platform,omitempty"`
	Language string      `json:"language,omitempty"`
}

type ContentResult struct {
	Content     string   `json:"content"`
	Title       string   `json:"title,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

type Service interface {
	GenerateContent(ctx context.Context, req *ContentRequest) (*ContentResult, error)
	AnalyzeHotspot(ctx context.Context, title, content string) (*HotspotAnalysis, error)
	RewriteContent(ctx context.Context, content string, style string) (string, error)
	AuditContent(ctx context.Context, content string) (*AuditResult, error)
}

type HotspotAnalysis struct {
	Summary     string   `json:"summary"`
	KeyPoints   []string `json:"key_points"`
	Sentiment   string   `json:"sentiment"`
	Relevance   int      `json:"relevance"`
	Suggestions []string `json:"suggestions"`
	Tags        []string `json:"tags"`
}

type AuditResult struct {
	Passed      bool     `json:"passed"`
	Issues      []string `json:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
	Score       int      `json:"score"`
}
