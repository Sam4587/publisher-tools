package asr

import (
	"context"
	"time"
)

// ProviderType ASR提供商类型
type ProviderType string

const (
	ProviderBcutASR   ProviderType = "bcut-asr"   // 必剪云端ASR
	ProviderWhisper   ProviderType = "whisper"    // OpenAI Whisper本地
	ProviderFaster    ProviderType = "faster"     // Faster-Whisper
	ProviderAuto      ProviderType = "auto"       // 自动选择
)

// RecognitionResult 识别结果
type RecognitionResult struct {
	Provider      ProviderType        `json:"provider"`
	Language      string              `json:"language"`
	Text          string              `json:"text"`
	Segments      []RecognitionSegment `json:"segments"`
	Duration      time.Duration       `json:"duration"`
	AudioDuration float64             `json:"audio_duration"` // 音频时长(秒)
	WordCount     int                 `json:"word_count"`
	QualityScore  float64             `json:"quality_score"`  // 质量评分 0-100
	Cached        bool                `json:"cached"`         // 是否来自缓存
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RecognitionSegment 识别片段
type RecognitionSegment struct {
	ID       int     `json:"id"`
	Start    float64 `json:"start"`    // 开始时间(秒)
	End      float64 `json:"end"`      // 结束时间(秒)
	Text     string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"` // 置信度 0-1
	Words    []Word  `json:"words,omitempty"`
}

// Word 词级别时间戳
type Word struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Provider ASR提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() ProviderType
	
	// Recognize 执行语音识别
	Recognize(ctx context.Context, audioPath string, opts *RecognizeOptions) (*RecognitionResult, error)
	
	// IsAvailable 检查服务是否可用
	IsAvailable() bool
	
	// GetPriority 获取优先级（数值越小优先级越高）
	GetPriority() int
	
	// SupportsLanguage 检查是否支持指定语言
	SupportsLanguage(lang string) bool
	
	// GetMaxFileSize 获取支持的最大文件大小(字节)
	GetMaxFileSize() int64
	
	// GetMaxDuration 获取支持的最大音频时长(秒)
	GetMaxDuration() float64
}

// RecognizeOptions 识别选项
type RecognizeOptions struct {
	Language         string        `json:"language"`          // 语言代码: zh, en, ja, auto等
	Model            string        `json:"model"`             // 模型名称(whisper专用)
	EnableTimestamps bool          `json:"enable_timestamps"` // 是否启用时间戳
	EnableWordLevel  bool          `json:"enable_word_level"` // 是否启用词级时间戳
	EnableGPU        bool          `json:"enable_gpu"`        // 是否启用GPU加速
	Temperature      float64       `json:"temperature"`       // 采样温度
	InitialPrompt    string        `json:"initial_prompt"`    // 初始提示词
	Timeout          time.Duration `json:"timeout"`           // 超时时间
}

// DefaultRecognizeOptions 默认识别选项
func DefaultRecognizeOptions() *RecognizeOptions {
	return &RecognizeOptions{
		Language:         "auto",
		Model:            "base",
		EnableTimestamps: true,
		EnableWordLevel:  false,
		EnableGPU:        false,
		Temperature:      0.0,
		Timeout:          30 * time.Minute,
	}
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Type       ProviderType `json:"type"`
	Enabled    bool         `json:"enabled"`
	Priority   int          `json:"priority"`
	MaxRetries int          `json:"max_retries"`
	Timeout    time.Duration `json:"timeout"`
	
	// BcutASR配置
	BcutAPIKey  string `json:"bcut_api_key,omitempty"`
	BcutAPIURL  string `json:"bcut_api_url,omitempty"`
	
	// Whisper配置
	WhisperPath string `json:"whisper_path,omitempty"`
	WhisperModel string `json:"whisper_model,omitempty"`
	
	// Faster-Whisper配置
	FasterWhisperPath string `json:"faster_whisper_path,omitempty"`
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	AverageConfidence float64 `json:"average_confidence"` // 平均置信度
	SilenceRatio      float64 `json:"silence_ratio"`      // 静音比例
	SpeechRate        float64 `json:"speech_rate"`        // 语速(词/分钟)
	SegmentCount      int     `json:"segment_count"`      // 片段数量
	AvgSegmentLength  float64 `json:"avg_segment_length"` // 平均片段长度
	RepetitionScore   float64 `json:"repetition_score"`   // 重复评分
}

// CalculateQualityScore 计算质量评分
func CalculateQualityScore(result *RecognitionResult) float64 {
	if result == nil || len(result.Segments) == 0 {
		return 0
	}

	var totalConfidence float64
	var validSegments int

	for _, seg := range result.Segments {
		if seg.Confidence > 0 {
			totalConfidence += seg.Confidence
			validSegments++
		}
	}

	// 基础分数来自平均置信度
	var score float64
	if validSegments > 0 {
		score = (totalConfidence / float64(validSegments)) * 60 // 60%权重
	} else {
		score = 50 // 无置信度信息时给中等分数
	}

	// 文本完整性加分 (20%权重)
	if result.WordCount > 0 && result.AudioDuration > 0 {
		wordsPerSecond := float64(result.WordCount) / result.AudioDuration
		// 正常语速范围: 中文2-4词/秒, 英文2-3词/秒
		if wordsPerSecond >= 1 && wordsPerSecond <= 5 {
			score += 20
		} else if wordsPerSecond >= 0.5 && wordsPerSecond <= 6 {
			score += 10
		}
	}

	// 片段连续性加分 (20%权重)
	if len(result.Segments) > 1 {
		continuousScore := 0.0
		for i := 1; i < len(result.Segments); i++ {
			gap := result.Segments[i].Start - result.Segments[i-1].End
			if gap >= -0.1 && gap <= 0.5 { // 允许小重叠和间隙
				continuousScore += 1
			}
		}
		continuityRatio := continuousScore / float64(len(result.Segments)-1)
		score += continuityRatio * 20
	}

	if score > 100 {
		score = 100
	}

	return score
}

// GetSupportedLanguages 获取支持的语言列表
func GetSupportedLanguages() map[string]string {
	return map[string]string{
		"auto": "自动检测",
		"zh":   "中文",
		"en":   "英语",
		"ja":   "日语",
		"ko":   "韩语",
		"fr":   "法语",
		"de":   "德语",
		"es":   "西班牙语",
		"ru":   "俄语",
		"pt":   "葡萄牙语",
		"it":   "意大利语",
		"ar":   "阿拉伯语",
		"th":   "泰语",
		"vi":   "越南语",
	}
}
