package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type OpenRouterProvider struct {
	apiKey       string
	baseURL      string
	models       []string
	defaultModel string
}

type openRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type openRouterResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	return &OpenRouterProvider{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		models: []string{
			"deepseek/deepseek-r1-0528:free",
			"meta-llama/llama-3.3-70b-instruct:free",
			"qwen/qwen3-4b:free",
			"qwen/qwen3-coder:free",
			"google/gemma-3-27b-it:free",
			"mistralai/mistral-small-3.1-24b-instruct:free",
		},
		defaultModel: "deepseek/deepseek-r1-0528:free",
	}
}

func (p *OpenRouterProvider) Name() ProviderType {
	return ProviderOpenRouter
}

func (p *OpenRouterProvider) Models() []string {
	return p.models
}

func (p *OpenRouterProvider) DefaultModel() string {
	return p.defaultModel
}

func (p *OpenRouterProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	req := openRouterRequest{
		Model:    model,
		Messages: opts.Messages,
		Stream:   false,
	}

	if opts.MaxTokens > 0 {
		req.MaxTokens = opts.MaxTokens
	}
	if opts.Temperature > 0 {
		req.Temperature = opts.Temperature
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://publisher-tools.local")
	httpReq.Header.Set("X-Title", "Publisher Tools")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response choices")
	}

	return &GenerateResult{
		Content:      result.Choices[0].Message.Content,
		Model:        model,
		Provider:     string(ProviderOpenRouter),
		InputTokens:  result.Usage.PromptTokens,
		OutputTokens: result.Usage.CompletionTokens,
		FinishedAt:   time.Now(),
	}, nil
}

func (p *OpenRouterProvider) GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error) {
	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	req := openRouterRequest{
		Model:    model,
		Messages: opts.Messages,
		Stream:   true,
	}

	if opts.MaxTokens > 0 {
		req.MaxTokens = opts.MaxTokens
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://publisher-tools.local")
	httpReq.Header.Set("X-Title", "Publisher Tools")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	ch := make(chan string, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var streamResp openRouterResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				logrus.Warnf("parse stream data: %v", err)
				continue
			}

			if len(streamResp.Choices) > 0 {
				content := streamResp.Choices[0].Delta.Content
				if content != "" {
					ch <- content
				}
			}
		}

		if err := scanner.Err(); err != nil {
			logrus.Errorf("stream scanner error: %v", err)
		}
	}()

	return ch, nil
}
