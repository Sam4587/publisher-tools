package video

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Transcriber 转录器
type Transcriber struct {
	db              *gorm.DB
	whisperPath     string
	model           string
	language        string
	outputDir       string
	enableGPU       bool
}

// TranscriberConfig 转录器配置
type TranscriberConfig struct {
	WhisperPath string `json:"whisper_path"`
	Model       string `json:"model"`        // tiny, base, small, medium, large
	Language    string `json:"language"`     // zh, en, auto
	OutputDir   string `json:"output_dir"`
	EnableGPU   bool   `json:"enable_gpu"`
}

// DefaultTranscriberConfig 默认配置
func DefaultTranscriberConfig() *TranscriberConfig {
	return &TranscriberConfig{
		WhisperPath: "whisper",
		Model:       "base",
		Language:    "auto",
		OutputDir:   "./data/transcripts",
		EnableGPU:   false,
	}
}

// NewTranscriber 创建转录器
func NewTranscriber(db *gorm.DB, config *TranscriberConfig) *Transcriber {
	if config == nil {
		config = DefaultTranscriberConfig()
	}

	os.MkdirAll(config.OutputDir, 0755)

	return &Transcriber{
		db:          db,
		whisperPath: config.WhisperPath,
		model:       config.Model,
		language:    config.Language,
		outputDir:   config.OutputDir,
		enableGPU:   config.EnableGPU,
	}
}

// TranscriptResult 转录结果
type TranscriptResult struct {
	VideoID     string          `json:"video_id"`
	Language    string          `json:"language"`
	Text        string          `json:"text"`
	Segments    []TranscriptSegment `json:"segments"`
	Duration    float64         `json:"duration"`
	WordCount   int             `json:"word_count"`
}

// TranscriptSegment 转录片段
type TranscriptSegment struct {
	ID      int     `json:"id"`
	Start   float64 `json:"start"`
	End     float64 `json:"end"`
	Text    string  `json:"text"`
	Tokens  []int   `json:"tokens,omitempty"`
}

// Transcribe 转录音频/视频
func (t *Transcriber) Transcribe(ctx context.Context, audioPath string, videoID string) (*TranscriptResult, error) {
	// 构建命令参数
	args := []string{
		audioPath,
		"--model", t.model,
		"--output_format", "json",
		"--output_dir", t.outputDir,
		"--word_timestamps", "True",
	}

	// 语言设置
	if t.language != "auto" {
		args = append(args, "--language", t.language)
	}

	// GPU 设置
	if t.enableGPU {
		args = append(args, "--device", "cuda")
	} else {
		args = append(args, "--device", "cpu")
	}

	// 执行转录
	startTime := time.Now()
	cmd := exec.CommandContext(ctx, t.whisperPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper error: %s - %w", string(output), err)
	}

	// 查找输出文件
	outputFile := t.findOutputFile(audioPath)
	if outputFile == "" {
		return nil, fmt.Errorf("output file not found")
	}
	defer os.Remove(outputFile)

	// 解析结果
	result, err := t.parseOutputFile(outputFile)
	if err != nil {
		return nil, err
	}

	result.VideoID = videoID
	result.Duration = time.Since(startTime).Seconds()

	// 保存到数据库
	if t.db != nil && videoID != "" {
		if err := t.saveTranscript(videoID, result); err != nil {
			logrus.Warnf("Failed to save transcript: %v", err)
		}
	}

	return result, nil
}

// TranscribeWithFallback 带降级的转录
func (t *Transcriber) TranscribeWithFallback(ctx context.Context, audioPath string, videoID string) (*TranscriptResult, error) {
	// 尝试使用 Faster-Whisper
	result, err := t.Transcribe(ctx, audioPath, videoID)
	if err == nil {
		return result, nil
	}

	logrus.Warnf("Faster-Whisper failed: %v, trying alternative methods", err)

	// 降级到标准 whisper
	fallbackTranscriber := &Transcriber{
		db:          t.db,
		whisperPath: "whisper",
		model:       t.model,
		language:    t.language,
		outputDir:   t.outputDir,
		enableGPU:   false,
	}

	return fallbackTranscriber.Transcribe(ctx, audioPath, videoID)
}

// findOutputFile 查找输出文件
func (t *Transcriber) findOutputFile(audioPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	outputFile := filepath.Join(t.outputDir, baseName+".json")

	if _, err := os.Stat(outputFile); err == nil {
		return outputFile
	}

	// 尝试其他可能的文件名
	files, _ := os.ReadDir(t.outputDir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), baseName) && strings.HasSuffix(f.Name(), ".json") {
			return filepath.Join(t.outputDir, f.Name())
		}
	}

	return ""
}

// parseOutputFile 解析输出文件
func (t *Transcriber) parseOutputFile(filePath string) (*TranscriptResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read output file: %w", err)
	}

	// Faster-Whisper JSON 格式
	var whisperOutput struct {
		Text     string `json:"text"`
		Language string `json:"language"`
		Segments []struct {
			ID    int     `json:"id"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
			Text  string  `json:"text"`
		} `json:"segments"`
	}

	if err := json.Unmarshal(data, &whisperOutput); err != nil {
		return nil, fmt.Errorf("parse whisper output: %w", err)
	}

	result := &TranscriptResult{
		Language: whisperOutput.Language,
		Text:     strings.TrimSpace(whisperOutput.Text),
	}

	for _, seg := range whisperOutput.Segments {
		result.Segments = append(result.Segments, TranscriptSegment{
			ID:    seg.ID,
			Start: seg.Start,
			End:   seg.End,
			Text:  strings.TrimSpace(seg.Text),
		})
	}

	// 计算字数
	result.WordCount = len(strings.Fields(result.Text))

	return result, nil
}

// saveTranscript 保存转录结果
func (t *Transcriber) saveTranscript(videoID string, result *TranscriptResult) error {
	transcript := &database.Transcript{
		VideoID:  videoID,
		Language: result.Language,
		Content:  result.Text,
	}

	// 检查是否已存在
	var existing database.Transcript
	err := t.db.Where("video_id = ?", videoID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return t.db.Create(transcript).Error
	} else if err == nil {
		transcript.ID = existing.ID
		return t.db.Save(transcript).Error
	}

	return err
}

// TranscribeFromSRT 从 SRT 字幕文件转录
func (t *Transcriber) TranscribeFromSRT(srtPath string, videoID string) (*TranscriptResult, error) {
	data, err := os.ReadFile(srtPath)
	if err != nil {
		return nil, fmt.Errorf("read srt file: %w", err)
	}

	result := &TranscriptResult{
		VideoID: videoID,
	}

	// 解析 SRT 格式
	segments := parseSRT(string(data))
	result.Segments = segments

	// 合并文本
	var texts []string
	for _, seg := range segments {
		texts = append(texts, seg.Text)
	}
	result.Text = strings.Join(texts, " ")
	result.WordCount = len(strings.Fields(result.Text))

	// 保存到数据库
	if t.db != nil && videoID != "" {
		if err := t.saveTranscript(videoID, result); err != nil {
			logrus.Warnf("Failed to save transcript: %v", err)
		}
	}

	return result, nil
}

// parseSRT 解析 SRT 格式
func parseSRT(content string) []TranscriptSegment {
	var segments []TranscriptSegment
	var currentSegment *TranscriptSegment
	id := 0

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if currentSegment != nil {
				segments = append(segments, *currentSegment)
				currentSegment = nil
			}
			continue
		}

		// 序号
		if isNumber(line) && currentSegment == nil {
			id++
			currentSegment = &TranscriptSegment{ID: id}
			continue
		}

		// 时间码
		if currentSegment != nil && strings.Contains(line, "-->") {
			times := strings.Split(line, " --> ")
			if len(times) == 2 {
				currentSegment.Start = parseSRTTime(times[0])
				currentSegment.End = parseSRTTime(times[1])
			}
			continue
		}

		// 文本
		if currentSegment != nil && currentSegment.Start > 0 {
			if currentSegment.Text != "" {
				currentSegment.Text += " "
			}
			currentSegment.Text += line
		}
	}

	if currentSegment != nil {
		segments = append(segments, *currentSegment)
	}

	return segments
}

// isNumber 检查是否为数字
func isNumber(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseSRTTime 解析 SRT 时间格式
func parseSRTTime(timeStr string) float64 {
	// 格式: 00:00:00,000
	re := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d{3})`)
	matches := re.FindStringSubmatch(timeStr)

	if len(matches) != 5 {
		return 0
	}

	hours := parseFloat(matches[1])
	minutes := parseFloat(matches[2])
	seconds := parseFloat(matches[3])
	millis := parseFloat(matches[4])

	return hours*3600 + minutes*60 + seconds + millis/1000
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// IsWhisperAvailable 检查 Whisper 是否可用
func (t *Transcriber) IsWhisperAvailable() bool {
	cmd := exec.Command(t.whisperPath, "--help")
	return cmd.Run() == nil
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
	}
}

// EstimateTokenCount 估算 Token 数量
func EstimateTokenCount(text string) int {
	// 简单估算：中文约 1.5 字/token，英文约 4 字符/token
	chineseCount := 0
	englishCount := 0

	for _, r := range text {
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		}
	}

	return int(float64(chineseCount)/1.5 + float64(englishCount)/4)
}

// SplitLongText 分割长文本
func SplitLongText(text string, maxTokens int) []string {
	tokens := EstimateTokenCount(text)
	if tokens <= maxTokens {
		return []string{text}
	}

	// 按段落分割
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var currentChunk strings.Builder
	currentTokens := 0

	for _, para := range paragraphs {
		paraTokens := EstimateTokenCount(para)

		if currentTokens+paraTokens > maxTokens {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				currentTokens = 0
			}

			// 如果单个段落超过限制，按句子分割
			if paraTokens > maxTokens {
				sentences := splitBySentences(para, maxTokens)
				chunks = append(chunks, sentences...)
				continue
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(para)
		currentTokens += paraTokens
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// splitBySentences 按句子分割
func splitBySentences(text string, maxTokens int) []string {
	sentences := regexp.MustCompile(`[。！？.!?]+`).Split(text, -1)
	var chunks []string
	var currentChunk strings.Builder
	currentTokens := 0

	for _, sent := range sentences {
		sent = strings.TrimSpace(sent)
		if sent == "" {
			continue
		}

		sentTokens := EstimateTokenCount(sent)

		if currentTokens+sentTokens > maxTokens {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				currentTokens = 0
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("。")
		}
		currentChunk.WriteString(sent)
		currentTokens += sentTokens
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
