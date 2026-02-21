package asr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// BcutASRProvider 必剪云端ASR提供商
type BcutASRProvider struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
	priority   int
	maxRetries int
	timeout    time.Duration
}

// BcutASRConfig 必剪ASR配置
type BcutASRConfig struct {
	APIKey     string        `json:"api_key"`
	APIURL     string        `json:"api_url"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
	Priority   int           `json:"priority"`
}

// DefaultBcutASRConfig 默认配置
func DefaultBcutASRConfig() *BcutASRConfig {
	return &BcutASRConfig{
		APIURL:     "https://member.bilibili.com/x/bcut/rubick-interface",
		Timeout:    10 * time.Minute,
		MaxRetries: 3,
		Priority:   1, // 最高优先级
	}
}

// NewBcutASRProvider 创建必剪ASR提供商
func NewBcutASRProvider(config *BcutASRConfig) *BcutASRProvider {
	if config == nil {
		config = DefaultBcutASRConfig()
	}

	return &BcutASRProvider{
		apiKey:     config.APIKey,
		apiURL:     config.APIURL,
		priority:   config.Priority,
		maxRetries: config.MaxRetries,
		timeout:    config.Timeout,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// bcutASRRequest ASR请求
type bcutASRRequest struct {
	AudioURL   string `json:"audio_url,omitempty"`
	Language   string `json:"language,omitempty"`
	WithTimestamp bool `json:"with_timestamp"`
}

// bcutASRResponse ASR响应
type bcutASRResponse struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    *bcutASRResultData `json:"data"`
}

// bcutASRResultData 结果数据
type bcutASRResultData struct {
	TaskID    string              `json:"task_id"`
	Status    string              `json:"status"`
	Text      string              `json:"text"`
	Segments  []bcutASRSegment    `json:"segments"`
	Language  string              `json:"language"`
	Duration  float64             `json:"duration"`
}

// bcutASRSegment 片段
type bcutASRSegment struct {
	ID         int     `json:"id"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// Name 返回提供商名称
func (p *BcutASRProvider) Name() ProviderType {
	return ProviderBcutASR
}

// Recognize 执行语音识别
func (p *BcutASRProvider) Recognize(ctx context.Context, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	if opts == nil {
		opts = DefaultRecognizeOptions()
	}

	startTime := time.Now()

	// 读取音频文件
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("read audio file: %w", err)
	}

	// 检查文件大小 (最大500MB)
	if int64(len(audioData)) > p.GetMaxFileSize() {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", len(audioData), p.GetMaxFileSize())
	}

	// 上传并识别
	result, err := p.uploadAndRecognize(ctx, audioPath, audioData, opts)
	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	result.QualityScore = CalculateQualityScore(result)

	return result, nil
}

// uploadAndRecognize 上传并识别
func (p *BcutASRProvider) uploadAndRecognize(ctx context.Context, audioPath string, audioData []byte, opts *RecognizeOptions) (*RecognitionResult, error) {
	var lastErr error

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Second * time.Duration(attempt*2)):
			}
		}

		// 创建上传请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加音频文件
		part, err := writer.CreateFormFile("audio", filepath.Base(audioPath))
		if err != nil {
			lastErr = err
			continue
		}
		if _, err := part.Write(audioData); err != nil {
			lastErr = err
			continue
		}

		// 添加其他参数
		_ = writer.WriteField("language", p.mapLanguage(opts.Language))
		if opts.EnableTimestamps {
			_ = writer.WriteField("with_timestamp", "true")
		}

		if err := writer.Close(); err != nil {
			lastErr = err
			continue
		}

		// 发送请求
		req, err := http.NewRequestWithContext(ctx, "POST", p.apiURL+"/asr/recognize", body)
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		if p.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+p.apiKey)
		}

		resp, err := p.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// 解析响应
		result, err := p.parseResponse(resp)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		return result, nil
	}

	return nil, fmt.Errorf("after %d retries: %w", p.maxRetries, lastErr)
}

// parseResponse 解析响应
func (p *BcutASRProvider) parseResponse(resp *http.Response) (*RecognitionResult, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var asrResp bcutASRResponse
	if err := json.Unmarshal(body, &asrResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if asrResp.Code != 0 {
		return nil, fmt.Errorf("API error code %d: %s", asrResp.Code, asrResp.Message)
	}

	if asrResp.Data == nil {
		return nil, fmt.Errorf("empty result data")
	}

	// 转换结果
	result := &RecognitionResult{
		Provider:     ProviderBcutASR,
		Language:     asrResp.Data.Language,
		Text:         strings.TrimSpace(asrResp.Data.Text),
		AudioDuration: asrResp.Data.Duration,
		WordCount:    len(strings.Fields(asrResp.Data.Text)),
	}

	// 转换片段
	for _, seg := range asrResp.Data.Segments {
		result.Segments = append(result.Segments, RecognitionSegment{
			ID:         seg.ID,
			Start:      seg.StartTime,
			End:        seg.EndTime,
			Text:       strings.TrimSpace(seg.Text),
			Confidence: seg.Confidence,
		})
	}

	return result, nil
}

// mapLanguage 映射语言代码
func (p *BcutASRProvider) mapLanguage(lang string) string {
	langMap := map[string]string{
		"auto": "auto",
		"zh":   "zh-CN",
		"en":   "en-US",
		"ja":   "ja-JP",
		"ko":   "ko-KR",
	}

	if mapped, ok := langMap[lang]; ok {
		return mapped
	}
	return "auto"
}

// IsAvailable 检查服务是否可用
func (p *BcutASRProvider) IsAvailable() bool {
	// 如果没有配置API Key，尝试使用免费接口
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.apiURL+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		logrus.Debugf("BcutASR health check failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound
}

// GetPriority 获取优先级
func (p *BcutASRProvider) GetPriority() int {
	return p.priority
}

// SupportsLanguage 检查是否支持指定语言
func (p *BcutASRProvider) SupportsLanguage(lang string) bool {
	supported := map[string]bool{
		"auto": true,
		"zh":   true,
		"en":   true,
		"ja":   true,
		"ko":   true,
	}
	return supported[lang] || supported["auto"]
}

// GetMaxFileSize 获取最大文件大小 (500MB)
func (p *BcutASRProvider) GetMaxFileSize() int64 {
	return 500 * 1024 * 1024
}

// GetMaxDuration 获取最大音频时长 (2小时)
func (p *BcutASRProvider) GetMaxDuration() float64 {
	return 2 * 60 * 60
}

// SetAPIKey 设置API密钥
func (p *BcutASRProvider) SetAPIKey(apiKey string) {
	p.apiKey = apiKey
}

// SetAPIURL 设置API地址
func (p *BcutASRProvider) SetAPIURL(apiURL string) {
	p.apiURL = apiURL
}
