package asr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// WhisperProvider Whisper本地识别提供商
type WhisperProvider struct {
	whisperPath string
	model       string
	outputDir   string
	enableGPU   bool
	priority    int
	maxRetries  int
	timeout     time.Duration
}

// WhisperConfig Whisper配置
type WhisperConfig struct {
	WhisperPath string        `json:"whisper_path"`
	Model       string        `json:"model"`        // tiny, base, small, medium, large
	OutputDir   string        `json:"output_dir"`
	EnableGPU   bool          `json:"enable_gpu"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	Priority    int           `json:"priority"`
}

// DefaultWhisperConfig 默认配置
func DefaultWhisperConfig() *WhisperConfig {
	return &WhisperConfig{
		WhisperPath: "whisper",
		Model:       "base",
		OutputDir:   "./data/transcripts",
		EnableGPU:   false,
		Timeout:     30 * time.Minute,
		MaxRetries:  2,
		Priority:    2, // 次高优先级
	}
}

// NewWhisperProvider 创建Whisper提供商
func NewWhisperProvider(config *WhisperConfig) *WhisperProvider {
	if config == nil {
		config = DefaultWhisperConfig()
	}

	// 确保输出目录存在
	os.MkdirAll(config.OutputDir, 0755)

	return &WhisperProvider{
		whisperPath: config.WhisperPath,
		model:       config.Model,
		outputDir:   config.OutputDir,
		enableGPU:   config.EnableGPU,
		priority:    config.Priority,
		maxRetries:  config.MaxRetries,
		timeout:     config.Timeout,
	}
}

// Name 返回提供商名称
func (p *WhisperProvider) Name() ProviderType {
	return ProviderWhisper
}

// Recognize 执行语音识别
func (p *WhisperProvider) Recognize(ctx context.Context, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error) {
	if opts == nil {
		opts = DefaultRecognizeOptions()
	}

	// 检查文件是否存在
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file not found: %s", audioPath)
	}

	// 检查文件大小
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return nil, fmt.Errorf("stat audio file: %w", err)
	}
	if fileInfo.Size() > p.GetMaxFileSize() {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", fileInfo.Size(), p.GetMaxFileSize())
	}

	startTime := time.Now()

	// 构建命令参数
	args := p.buildArgs(audioPath, opts)

	// 执行识别
	var lastErr error
	var result *RecognitionResult

	for attempt := 0; attempt < p.maxRetries; attempt++ {
		if attempt > 0 {
			logrus.Debugf("Whisper retry attempt %d", attempt+1)
			time.Sleep(time.Second * time.Duration(attempt))
		}

		result, lastErr = p.executeWhisper(ctx, audioPath, args, opts)
		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	result.Duration = time.Since(startTime)
	result.QualityScore = CalculateQualityScore(result)

	return result, nil
}

// buildArgs 构建命令参数
func (p *WhisperProvider) buildArgs(audioPath string, opts *RecognizeOptions) []string {
	args := []string{
		audioPath,
		"--model", p.getModel(opts),
		"--output_format", "json",
		"--output_dir", p.outputDir,
	}

	// 时间戳
	if opts.EnableTimestamps {
		args = append(args, "--word_timestamps", "True")
	}

	// 语言设置
	if opts.Language != "auto" && opts.Language != "" {
		args = append(args, "--language", opts.Language)
	}

	// 设备设置
	if p.enableGPU || opts.EnableGPU {
		args = append(args, "--device", "cuda")
	} else {
		args = append(args, "--device", "cpu")
	}

	// 温度
	if opts.Temperature > 0 {
		args = append(args, "--temperature", fmt.Sprintf("%.1f", opts.Temperature))
	}

	// 初始提示词
	if opts.InitialPrompt != "" {
		args = append(args, "--initial_prompt", opts.InitialPrompt)
	}

	return args
}

// getModel 获取模型名称
func (p *WhisperProvider) getModel(opts *RecognizeOptions) string {
	if opts.Model != "" {
		return opts.Model
	}
	return p.model
}

// executeWhisper 执行Whisper命令
func (p *WhisperProvider) executeWhisper(ctx context.Context, audioPath string, args []string, opts *RecognizeOptions) (*RecognitionResult, error) {
	// 设置超时
	timeout := p.timeout
	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(ctx, p.whisperPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper error: %s - %w", string(output), err)
	}

	// 查找输出文件
	outputFile := p.findOutputFile(audioPath)
	if outputFile == "" {
		return nil, fmt.Errorf("output file not found")
	}
	defer os.Remove(outputFile)

	// 解析结果
	return p.parseOutputFile(outputFile)
}

// findOutputFile 查找输出文件
func (p *WhisperProvider) findOutputFile(audioPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	outputFile := filepath.Join(p.outputDir, baseName+".json")

	if _, err := os.Stat(outputFile); err == nil {
		return outputFile
	}

	// 尝试其他可能的文件名
	files, _ := os.ReadDir(p.outputDir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), baseName) && strings.HasSuffix(f.Name(), ".json") {
			return filepath.Join(p.outputDir, f.Name())
		}
	}

	return ""
}

// parseOutputFile 解析输出文件
func (p *WhisperProvider) parseOutputFile(filePath string) (*RecognitionResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read output file: %w", err)
	}

	// Whisper JSON 格式
	var whisperOutput struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			ID    int     `json:"id"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
			Tokens []struct {
				ID    int     `json:"id"`
				Start float64 `json:"start"`
				End   float64 `json:"end"`
				Word  string  `json:"word"`
			} `json:"tokens,omitempty"`
		} `json:"segments"`
	}

	if err := json.Unmarshal(data, &whisperOutput); err != nil {
		return nil, fmt.Errorf("parse whisper output: %w", err)
	}

	result := &RecognitionResult{
		Provider: ProviderWhisper,
		Language: whisperOutput.Language,
		Text:     strings.TrimSpace(whisperOutput.Text),
	}

	// 计算音频时长
	var maxEnd float64
	for _, seg := range whisperOutput.Segments {
		if seg.End > maxEnd {
			maxEnd = seg.End
		}

		segment := RecognitionSegment{
			ID:    seg.ID,
			Start: seg.Start,
			End:   seg.End,
			Text:  strings.TrimSpace(seg.Text),
		}

		// 提取词级时间戳
		if len(seg.Tokens) > 0 {
			for _, tok := range seg.Tokens {
				if tok.Word != "" {
					segment.Words = append(segment.Words, Word{
						Word:  tok.Word,
						Start: tok.Start,
						End:   tok.End,
					})
				}
			}
		}

		result.Segments = append(result.Segments, segment)
	}

	result.AudioDuration = maxEnd
	result.WordCount = len(strings.Fields(result.Text))

	return result, nil
}

// IsAvailable 检查服务是否可用
func (p *WhisperProvider) IsAvailable() bool {
	cmd := exec.Command(p.whisperPath, "--help")
	if err := cmd.Run(); err != nil {
		logrus.Debugf("Whisper not available: %v", err)
		return false
	}
	return true
}

// GetPriority 获取优先级
func (p *WhisperProvider) GetPriority() int {
	return p.priority
}

// SupportsLanguage 检查是否支持指定语言
func (p *WhisperProvider) SupportsLanguage(lang string) bool {
	// Whisper支持所有主流语言
	return true
}

// GetMaxFileSize 获取最大文件大小 (2GB)
func (p *WhisperProvider) GetMaxFileSize() int64 {
	return 2 * 1024 * 1024 * 1024
}

// GetMaxDuration 获取最大音频时长 (无限制)
func (p *WhisperProvider) GetMaxDuration() float64 {
	return 0 // 无限制
}

// SetModel 设置模型
func (p *WhisperProvider) SetModel(model string) {
	p.model = model
}

// SetEnableGPU 设置GPU加速
func (p *WhisperProvider) SetEnableGPU(enable bool) {
	p.enableGPU = enable
}

// GetAvailableModels 获取可用模型列表
func GetAvailableModels() []string {
	return []string{
		"tiny",
		"tiny.en",
		"base",
		"base.en",
		"small",
		"small.en",
		"medium",
		"medium.en",
		"large",
		"large-v1",
		"large-v2",
		"large-v3",
	}
}

// GetModelInfo 获取模型信息
func GetModelInfo(model string) map[string]interface{} {
	models := map[string]map[string]interface{}{
		"tiny": {
			"params":      "39M",
			"vram":        "~1GB",
			"speed":       "~32x",
			"accuracy":    "较低",
			"recommended": "快速预览",
		},
		"base": {
			"params":      "74M",
			"vram":        "~1GB",
			"speed":       "~16x",
			"accuracy":    "中等",
			"recommended": "日常使用",
		},
		"small": {
			"params":      "244M",
			"vram":        "~2GB",
			"speed":       "~6x",
			"accuracy":    "良好",
			"recommended": "平衡选择",
		},
		"medium": {
			"params":      "769M",
			"vram":        "~5GB",
			"speed":       "~2x",
			"accuracy":    "优秀",
			"recommended": "高质量需求",
		},
		"large": {
			"params":      "1550M",
			"vram":        "~10GB",
			"speed":       "~1x",
			"accuracy":    "最佳",
			"recommended": "专业场景",
		},
	}

	if info, ok := models[model]; ok {
		return info
	}
	return models["base"]
}
