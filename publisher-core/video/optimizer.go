package video

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"publisher-core/ai/provider"
	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Optimizer 文本优化器
type Optimizer struct {
	db        *gorm.DB
	aiService AIProvider
	config    *OptimizerConfig
}

// AIProvider AI 提供商接口
type AIProvider interface {
	Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error)
}

// OptimizerConfig 优化器配置
type OptimizerConfig struct {
	MaxChunkTokens  int     `json:"max_chunk_tokens"`
	Temperature     float64 `json:"temperature"`
	EnableParallel  bool    `json:"enable_parallel"`
	MaxWorkers      int     `json:"max_workers"`
}

// DefaultOptimizerConfig 默认配置
func DefaultOptimizerConfig() *OptimizerConfig {
	return &OptimizerConfig{
		MaxChunkTokens:  2000,
		Temperature:     0.3,
		EnableParallel:  true,
		MaxWorkers:      3,
	}
}

// NewOptimizer 创建优化器
func NewOptimizer(db *gorm.DB, aiService AIProvider, config *OptimizerConfig) *Optimizer {
	if config == nil {
		config = DefaultOptimizerConfig()
	}
	return &Optimizer{
		db:        db,
		aiService: aiService,
		config:    config,
	}
}

// OptimizeResult 优化结果
type OptimizeResult struct {
	OriginalText  string   `json:"original_text"`
	OptimizedText string   `json:"optimized_text"`
	Summary       string   `json:"summary"`
	KeyPoints     []string `json:"key_points"`
	Topics        []string `json:"topics"`
	WordCount     int      `json:"word_count"`
	Duration      float64  `json:"duration"`
}

// Optimize 优化转录文本
func (o *Optimizer) Optimize(ctx context.Context, transcript string, opts *OptimizeOptions) (*OptimizeResult, error) {
	if opts == nil {
		opts = &OptimizeOptions{}
	}

	startTime := time.Now()
	result := &OptimizeResult{
		OriginalText: transcript,
		WordCount:    len(strings.Fields(transcript)),
	}

	// 分块处理长文本
	chunks := SplitLongText(transcript, o.config.MaxChunkTokens)

	if len(chunks) == 1 {
		// 单块直接处理
		optimized, err := o.optimizeChunk(ctx, chunks[0], opts)
		if err != nil {
			return nil, err
		}
		result.OptimizedText = optimized
	} else {
		// 多块并行处理
		optimizedChunks, err := o.optimizeChunks(ctx, chunks, opts)
		if err != nil {
			return nil, err
		}
		result.OptimizedText = strings.Join(optimizedChunks, "\n\n")
	}

	// 生成摘要
	summary, err := o.generateSummary(ctx, result.OptimizedText, opts)
	if err != nil {
		logrus.Warnf("Failed to generate summary: %v", err)
	} else {
		result.Summary = summary
	}

	// 提取关键点
	keyPoints, err := o.extractKeyPoints(ctx, result.OptimizedText, opts)
	if err != nil {
		logrus.Warnf("Failed to extract key points: %v", err)
	} else {
		result.KeyPoints = keyPoints
	}

	// 提取主题
	topics, err := o.extractTopics(ctx, result.OptimizedText, opts)
	if err != nil {
		logrus.Warnf("Failed to extract topics: %v", err)
	} else {
		result.Topics = topics
	}

	result.Duration = time.Since(startTime).Seconds()
	return result, nil
}

// OptimizeOptions 优化选项
type OptimizeOptions struct {
	Language       string `json:"language"`        // zh, en
	PreserveFormat bool   `json:"preserve_format"` // 保留原始格式
	AddPunctuation bool   `json:"add_punctuation"` // 添加标点
	FixGrammar     bool   `json:"fix_grammar"`     // 修正语法
	RemoveFiller   bool   `json:"remove_filler"`   // 移除填充词
}

// optimizeChunk 优化单个文本块
func (o *Optimizer) optimizeChunk(ctx context.Context, chunk string, opts *OptimizeOptions) (string, error) {
	prompt := o.buildOptimizePrompt(chunk, opts)

	result, err := o.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		Temperature: o.config.Temperature,
	})
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return strings.TrimSpace(result.Content), nil
}

// optimizeChunks 并行优化多个文本块
func (o *Optimizer) optimizeChunks(ctx context.Context, chunks []string, opts *OptimizeOptions) ([]string, error) {
	if !o.config.EnableParallel || len(chunks) <= 1 {
		// 串行处理
		results := make([]string, len(chunks))
		for i, chunk := range chunks {
			optimized, err := o.optimizeChunk(ctx, chunk, opts)
			if err != nil {
				return nil, err
			}
			results[i] = optimized
		}
		return results, nil
	}

	// 并行处理
	results := make([]string, len(chunks))
	errors := make([]error, len(chunks))

	var wg sync.WaitGroup
	sem := make(chan struct{}, o.config.MaxWorkers)

	for i, chunk := range chunks {
		wg.Add(1)
		go func(idx int, c string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			optimized, err := o.optimizeChunk(ctx, c, opts)
			results[idx] = optimized
			errors[idx] = err
		}(i, chunk)
	}

	wg.Wait()

	// 检查错误
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// buildOptimizePrompt 构建优化提示词
func (o *Optimizer) buildOptimizePrompt(text string, opts *OptimizeOptions) string {
	language := "中文"
	if opts.Language == "en" {
		language = "English"
	}

	var instructions []string
	instructions = append(instructions, fmt.Sprintf("请优化以下%s转录文本：", language))
	instructions = append(instructions, "1. 修正错别字和错误词汇")
	instructions = append(instructions, "2. 补全不完整的句子")
	instructions = append(instructions, "3. 按语义合理分段")

	if opts.AddPunctuation {
		instructions = append(instructions, "4. 添加正确的标点符号")
	}
	if opts.FixGrammar {
		instructions = append(instructions, "5. 修正语法错误")
	}
	if opts.RemoveFiller {
		instructions = append(instructions, "6. 移除口语填充词（如嗯、啊、那个等）")
	}

	instructions = append(instructions, "7. 保持原意不变，不要添加新内容")
	instructions = append(instructions, "8. 直接输出优化后的文本，不要添加任何解释")

	return fmt.Sprintf("%s\n\n转录文本：\n%s", strings.Join(instructions, "\n"), text)
}

// generateSummary 生成摘要
func (o *Optimizer) generateSummary(ctx context.Context, text string, opts *OptimizeOptions) (string, error) {
	// 如果文本太长，先截取关键部分
	if EstimateTokenCount(text) > 3000 {
		chunks := SplitLongText(text, 2000)
		text = chunks[0]
		if len(chunks) > 1 {
			text += "\n...\n" + chunks[len(chunks)-1]
		}
	}

	language := "中文"
	if opts.Language == "en" {
		language = "English"
	}

	prompt := fmt.Sprintf(`请用%s为以下内容生成一个简洁的摘要（100-200字）：

%s

直接输出摘要内容，不要添加任何解释。`, language, text)

	result, err := o.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		Temperature: 0.5,
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Content), nil
}

// extractKeyPoints 提取关键点
func (o *Optimizer) extractKeyPoints(ctx context.Context, text string, opts *OptimizeOptions) ([]string, error) {
	language := "中文"
	if opts.Language == "en" {
		language = "English"
	}

	prompt := fmt.Sprintf(`请从以下内容中提取5-10个关键要点，用%s回答：

%s

请按以下格式输出，每行一个要点：
- 要点1
- 要点2
...`, language, text)

	result, err := o.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	// 解析关键点
	lines := strings.Split(result.Content, "\n")
	var keyPoints []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "• ")
		if line != "" && !strings.HasPrefix(line, "要点") {
			keyPoints = append(keyPoints, line)
		}
	}

	return keyPoints, nil
}

// extractTopics 提取主题
func (o *Optimizer) extractTopics(ctx context.Context, text string, opts *OptimizeOptions) ([]string, error) {
	language := "中文"
	if opts.Language == "en" {
		language = "English"
	}

	prompt := fmt.Sprintf(`请从以下内容中提取3-5个主要主题/话题标签，用%s回答：

%s

请直接输出主题标签，用逗号分隔，例如：技术,创新,人工智能`, language, text)

	result, err := o.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return nil, err
	}

	// 解析主题
	topics := strings.Split(result.Content, ",")
	var cleanTopics []string
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			cleanTopics = append(cleanTopics, topic)
		}
	}

	return cleanTopics, nil
}

// SaveOptimizedResult 保存优化结果
func (o *Optimizer) SaveOptimizedResult(videoID string, result *OptimizeResult) error {
	if o.db == nil {
		return nil
	}

	// 更新转录记录
	return o.db.Model(&database.Transcript{}).
		Where("video_id = ?", videoID).
		Updates(map[string]interface{}{
			"optimized": result.OptimizedText,
			"summary":   result.Summary,
		}).Error
}

// OptimizeVideo 优化视频转录（完整流程）
func (o *Optimizer) OptimizeVideo(ctx context.Context, videoID string, transcript string, opts *OptimizeOptions) (*OptimizeResult, error) {
	// 执行优化
	result, err := o.Optimize(ctx, transcript, opts)
	if err != nil {
		return nil, err
	}

	// 保存结果
	if err := o.SaveOptimizedResult(videoID, result); err != nil {
		logrus.Warnf("Failed to save optimized result: %v", err)
	}

	// 更新视频状态
	if o.db != nil {
		o.db.Model(&database.Video{}).Where("id = ?", videoID).Update("status", "completed")
	}

	return result, nil
}
