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

// MistralProvider Mistral AI提供商
type MistralProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// MistralConfig Mistral配置
type MistralConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// NewMistralProvider 创建Mistral提供商
func NewMistralProvider(cfg *MistralConfig) *MistralProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.mistral.ai/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "mistral-large-latest"
	}

	return &MistralProvider{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Name 返回提供商名称
func (p *MistralProvider) Name() ProviderType {
	return "mistral"
}

// Generate 生成内容
func (p *MistralProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
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
	var mistralResp struct {
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

	if err := json.Unmarshal(body, &mistralResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(mistralResp.Choices) == 0 {
		return nil, fmt.Errorf("无生成结果")
	}

	return &GenerateResult{
		Content:      mistralResp.Choices[0].Message.Content,
		Model:        mistralResp.Model,
		Provider:     "mistral",
		InputTokens:  mistralResp.Usage.PromptTokens,
		OutputTokens: mistralResp.Usage.CompletionTokens,
		FinishedAt:   time.Now(),
	}, nil
}

// GenerateStream 流式生成
func (p *MistralProvider) GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error) {
	if opts.Model == "" {
		opts.Model = p.model
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"model":    opts.Model,
		"messages": opts.Messages,
		"stream":   true,
	}

	if opts.MaxTokens > 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		reqBody["temperature"] = opts.Temperature
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

	// 创建输出通道
	output := make(chan string, 100)

	// 启动goroutine处理流式响应
	go func() {
		defer close(output)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					break
				}
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				output <- streamResp.Choices[0].Delta.Content
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].FinishReason != "" {
				break
			}
		}
	}()

	return output, nil
}

// Models 返回支持的模型列表
func (p *MistralProvider) Models() []string {
	return []string{
		"mistral-large-latest",
		"mistral-medium-latest",
		"mistral-small-latest",
		"open-mistral-7b",
		"open-mixtral-8x7b",
		"open-mixtral-8x22b",
		"codestral-latest",
	}
}

// DefaultModel 返回默认模型
func (p *MistralProvider) DefaultModel() string {
	return p.model
}
