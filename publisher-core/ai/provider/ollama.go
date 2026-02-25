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

const (
	// OllamaDefaultBaseURL Ollama 云端 API 地址
	OllamaDefaultBaseURL = "https://ollama.com/api"
	// OllamaDefaultModel 默认模型
	OllamaDefaultModel = "gemma3:4b"
)

// OllamaProvider Ollama 提供商
type OllamaProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// OllamaChatRequest Ollama 聊天请求
type OllamaChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Options  *OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions Ollama 选项
type OllamaOptions struct {
	NumPredict int     `json:"num_predict,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP       float64 `json:"top_p,omitempty"`
}

// OllamaChatResponse Ollama 聊天响应
type OllamaChatResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
	DoneReason string   `json:"done_reason,omitempty"`
}

// NewOllamaProvider 创建 Ollama 提供商
func NewOllamaProvider(apiKey, baseURL, model string) (*OllamaProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for Ollama provider")
	}

	if baseURL == "" {
		baseURL = OllamaDefaultBaseURL
	}

	if model == "" {
		model = OllamaDefaultModel
	}

	return &OllamaProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// Name 返回提供商名称
func (p *OllamaProvider) Name() ProviderType {
	return ProviderOllama
}

// Generate 生成内容
func (p *OllamaProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
	req := &OllamaChatRequest{
		Model:    p.model,
		Messages: opts.Messages,
		Stream:   false,
	}

	if opts.MaxTokens > 0 {
		req.Options = &OllamaOptions{
			NumPredict:  opts.MaxTokens,
			Temperature: opts.Temperature,
			TopP:        opts.TopP,
		}
	} else {
		req.Options = &OllamaOptions{
			Temperature: opts.Temperature,
			TopP:        opts.TopP,
		}
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &GenerateResult{
		Content:      ollamaResp.Message.Content,
		Model:        ollamaResp.Model,
		Provider:     "ollama",
		InputTokens:  0, // Ollama 不返回 token 数
		OutputTokens: 0,
		FinishedAt:   time.Now(),
	}, nil
}

// GenerateStream 流式生成内容
func (p *OllamaProvider) GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		req := &OllamaChatRequest{
			Model:    p.model,
			Messages: opts.Messages,
			Stream:   true,
		}

		if opts.MaxTokens > 0 {
			req.Options = &OllamaOptions{
				NumPredict:  opts.MaxTokens,
				Temperature: opts.Temperature,
				TopP:        opts.TopP,
			}
		} else {
			req.Options = &OllamaOptions{
				Temperature: opts.Temperature,
				TopP:        opts.TopP,
			}
		}

		reqBody, err := json.Marshal(req)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat", bytes.NewReader(reqBody))
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			ch <- fmt.Sprintf("Error: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			ch <- fmt.Sprintf("Error: API request failed with status %d: %s", resp.StatusCode, string(body))
			return
		}

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp OllamaChatResponse
			if err := decoder.Decode(&streamResp); err != nil {
				if err == io.EOF {
					break
				}
				ch <- fmt.Sprintf("Error: %v", err)
				return
			}

			ch <- streamResp.Message.Content

			if streamResp.Done {
				break
			}
		}
	}()

	return ch, nil
}

// Models 返回可用模型列表
func (p *OllamaProvider) Models() []string {
	return []string{"gemma3:4b", "llama3", "deepseek-v3.2", "qwen3.5:397b"}
}

// DefaultModel 返回默认模型
func (p *OllamaProvider) DefaultModel() string {
	return p.model
}
