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

type GoogleProvider struct {
	apiKey       string
	baseURL      string
	models       []string
	defaultModel string
}

type googleRequest struct {
	Contents         []googleContent `json:"contents"`
	GenerationConfig googleConfig    `json:"generationConfig,omitempty"`
}

type googleContent struct {
	Role  string       `json:"role"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text"`
}

type googleConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
}

type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func NewGoogleProvider(apiKey string) *GoogleProvider {
	return &GoogleProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		models: []string{
			"gemini-2.5-flash",
			"gemini-2.5-flash-lite",
			"gemini-3-flash",
			"gemma-3-27b-it",
			"gemma-3-12b-it",
		},
		defaultModel: "gemini-2.5-flash",
	}
}

func (p *GoogleProvider) Name() ProviderType {
	return ProviderGoogle
}

func (p *GoogleProvider) Models() []string {
	return p.models
}

func (p *GoogleProvider) DefaultModel() string {
	return p.defaultModel
}

func (p *GoogleProvider) Generate(ctx context.Context, opts *GenerateOptions) (*GenerateResult, error) {
	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	contents := make([]googleContent, 0, len(opts.Messages))
	for _, msg := range opts.Messages {
		role := string(msg.Role)
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, googleContent{
			Role:  role,
			Parts: []googlePart{{Text: msg.Content}},
		})
	}

	req := googleRequest{
		Contents: contents,
		GenerationConfig: googleConfig{
			MaxOutputTokens: opts.MaxTokens,
			Temperature:     opts.Temperature,
			TopP:            opts.TopP,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, model, p.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result googleResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates")
	}

	content := ""
	if len(result.Candidates[0].Content.Parts) > 0 {
		content = result.Candidates[0].Content.Parts[0].Text
	}

	return &GenerateResult{
		Content:      content,
		Model:        model,
		Provider:     string(ProviderGoogle),
		InputTokens:  result.UsageMetadata.PromptTokenCount,
		OutputTokens: result.UsageMetadata.CandidatesTokenCount,
		FinishedAt:   time.Now(),
	}, nil
}

func (p *GoogleProvider) GenerateStream(ctx context.Context, opts *GenerateOptions) (<-chan string, error) {
	model := opts.Model
	if model == "" {
		model = p.defaultModel
	}

	contents := make([]googleContent, 0, len(opts.Messages))
	for _, msg := range opts.Messages {
		role := string(msg.Role)
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, googleContent{
			Role:  role,
			Parts: []googlePart{{Text: msg.Content}},
		})
	}

	req := googleRequest{
		Contents: contents,
		GenerationConfig: googleConfig{
			MaxOutputTokens: opts.MaxTokens,
			Temperature:     opts.Temperature,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.baseURL, model, p.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
			if data == "" {
				continue
			}

			var streamResp googleResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				logrus.Warnf("parse stream data: %v", err)
				continue
			}

			if streamResp.Error != nil {
				logrus.Errorf("stream error: %s", streamResp.Error.Message)
				return
			}

			if len(streamResp.Candidates) > 0 && len(streamResp.Candidates[0].Content.Parts) > 0 {
				content := streamResp.Candidates[0].Content.Parts[0].Text
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
