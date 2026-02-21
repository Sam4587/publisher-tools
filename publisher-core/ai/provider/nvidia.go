package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NVIDIAProvider NVIDIA AI提供商
type NVIDIAProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NVIDIAConfig NVIDIA配置
type NVIDIAConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewNVIDIAProvider 创建NVIDIA提供商
func NewNVIDIAProvider(cfg *NVIDIAConfig) *NVIDIAProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://integrate.api.nvidia.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "meta/llama-3.1-8b-instruct"
	}

	return &NVIDIAProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name 返回提供商名称
func (p *NVIDIAProvider) Name() ProviderType {
	return "nvidia"
}

// Generate 生成内容
func (p *NVIDIAProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
	if opts.Model == "" {
		opts.Model = p.model
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"model":    opts.Model,
		"messages": opts.Messages,
	}

	if opts.MaxTokens > 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		reqBody["temperature"] = opts.Temperature
	}
	if opts.TopP > 0 {
		reqBody["top_p"] = opts.TopP
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var nvidiaResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(body, &nvidiaResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(nvidiaResp.Choices) == 0 {
		return nil, fmt.Errorf("无生成结果")
	}

	return &GenerateResult{
		Content:      nvidiaResp.Choices[0].Message.Content,
		Model:        nvidiaResp.Model,
		Provider:     "nvidia",
		InputTokens:  nvidiaResp.Usage.PromptTokens,
		OutputTokens: nvidiaResp.Usage.CompletionTokens,
		FinishedAt:   time.Now(),
	}, nil
}

// GenerateStream 流式生成（NVIDIA暂不支持，返回错误）
func (p *NVIDIAProvider) GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error) {
	return nil, fmt.Errorf("NVIDIA提供商暂不支持流式生成")
}

// Models 返回支持的模型列表
func (p *NVIDIAProvider) Models() []string {
	return []string{
		"meta/llama-3.1-8b-instruct",
		"meta/llama-3.1-70b-instruct",
		"meta/llama-3.1-405b-instruct",
		"nvidia/nemotron-4-340b-instruct",
		"mistralai/mistral-large",
		"mistralai/mixtral-8x7b-instruct",
		"google/gemma-7b",
		"microsoft/phi-3-medium-128k-instruct",
	}
}

// DefaultModel 返回默认模型
func (p *NVIDIAProvider) DefaultModel() string {
	return p.model
}
