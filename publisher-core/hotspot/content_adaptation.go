package hotspot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"publisher-tools/ai/provider"
	"publisher-tools/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ContentAdaptationService 内容适配服务
type ContentAdaptationService struct {
	db        *gorm.DB
	aiService AIAnalyzer
	config    *ContentAdaptationConfig
}

// ContentAdaptationConfig 内容适配配置
type ContentAdaptationConfig struct {
	MaxContentLength   int               `json:"max_content_length"`   // 最大内容长度
	TargetPlatforms    []string          `json:"target_platforms"`     // 目标平台
	ContentStyles      map[string]string `json:"content_styles"`       // 内容风格
	AdaptationDepth    string            `json:"adaptation_depth"`     // 适配深度：light, medium, deep
}

// DefaultContentAdaptationConfig 默认配置
func DefaultContentAdaptationConfig() *ContentAdaptationConfig {
	return &ContentAdaptationConfig{
		MaxContentLength: 2000,
		TargetPlatforms: []string{"douyin", "xiaohongshu", "toutiao"},
		ContentStyles: map[string]string{
			"douyin":      "轻松幽默，吸引眼球",
			"xiaohongshu": "精致优雅，实用分享",
			"toutiao":     "专业深度，信息丰富",
		},
		AdaptationDepth: "medium",
	}
}

// NewContentAdaptationService 创建内容适配服务
func NewContentAdaptationService(db *gorm.DB, config *ContentAdaptationConfig) *ContentAdaptationService {
	if config == nil {
		config = DefaultContentAdaptationConfig()
	}
	return &ContentAdaptationService{
		db:     db,
		config: config,
	}
}

// SetAIService 设置AI服务
func (s *ContentAdaptationService) SetAIService(ai AIAnalyzer) {
	s.aiService = ai
}

// AdaptationRequest 适配请求
type AdaptationRequest struct {
	TopicID      string            `json:"topic_id"`
	Platform     string            `json:"platform"`
	Style        string            `json:"style"`
	Keywords     []string          `json:"keywords"`
	CustomPrompt string            `json:"custom_prompt"`
	Options      map[string]interface{} `json:"options"`
}

// AdaptationResult 适配结果
type AdaptationResult struct {
	TopicID      string    `json:"topic_id"`
	Platform     string    `json:"platform"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Summary      string    `json:"summary"`
	Keywords     []string  `json:"keywords"`
	Tags         []string  `json:"tags"`
	QualityScore float64   `json:"quality_score"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdaptContent 适配内容
func (s *ContentAdaptationService) AdaptContent(ctx context.Context, req *AdaptationRequest) (*AdaptationResult, error) {
	// 获取热点信息
	var topic database.Topic
	if err := s.db.Where("id = ?", req.TopicID).First(&topic).Error; err != nil {
		return nil, fmt.Errorf("获取热点失败: %w", err)
	}

	// 确定平台和风格
	platform := req.Platform
	if platform == "" {
		platform = s.config.TargetPlatforms[0]
	}

	style := req.Style
	if style == "" {
		style = s.config.ContentStyles[platform]
	}

	// 构建适配提示词
	prompt := s.buildAdaptationPrompt(topic, platform, style, req.Keywords, req.CustomPrompt)

	// 调用AI生成内容
	if s.aiService == nil {
		return nil, fmt.Errorf("AI服务未配置")
	}

	result, err := s.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		MaxTokens:   s.config.MaxContentLength,
		Temperature: 0.8,
	})

	if err != nil {
		return nil, fmt.Errorf("AI生成失败: %w", err)
	}

	// 解析生成的内容
	adaptedContent := s.parseGeneratedContent(result.Content, platform)

	// 计算质量评分
	qualityScore := s.calculateQualityScore(adaptedContent, topic)

	adaptationResult := &AdaptationResult{
		TopicID:      req.TopicID,
		Platform:     platform,
		Title:        adaptedContent.Title,
		Content:      adaptedContent.Content,
		Summary:      adaptedContent.Summary,
		Keywords:     adaptedContent.Keywords,
		Tags:         adaptedContent.Tags,
		QualityScore: qualityScore,
		CreatedAt:    time.Now(),
	}

	return adaptationResult, nil
}

// buildAdaptationPrompt 构建适配提示词
func (s *ContentAdaptationService) buildAdaptationPrompt(topic database.Topic, platform, style string, keywords []string, customPrompt string) string {
	var promptBuilder strings.Builder

	// 基础信息
	promptBuilder.WriteString(fmt.Sprintf("请基于以下热点话题，为%s平台创作内容：\n\n", platform))
	promptBuilder.WriteString(fmt.Sprintf("热点标题：%s\n", topic.Title))
	if topic.Description != "" {
		promptBuilder.WriteString(fmt.Sprintf("热点描述：%s\n", topic.Description))
	}

	// 平台要求
	promptBuilder.WriteString(fmt.Sprintf("\n平台要求：\n"))
	switch platform {
	case "douyin":
		promptBuilder.WriteString("- 标题：30字以内，吸引眼球\n")
		promptBuilder.WriteString("- 内容：轻松幽默，适合短视频\n")
		promptBuilder.WriteString("- 风格：口语化，有互动性\n")
	case "xiaohongshu":
		promptBuilder.WriteString("- 标题：20字以内，精致优雅\n")
		promptBuilder.WriteString("- 内容：实用分享，有情感共鸣\n")
		promptBuilder.WriteString("- 风格：图文并茂，有生活气息\n")
	case "toutiao":
		promptBuilder.WriteString("- 标题：30字以内，专业深度\n")
		promptBuilder.WriteString("- 内容：信息丰富，有观点\n")
		promptBuilder.WriteString("- 风格：专业严谨，有深度\n")
	}

	// 风格要求
	if style != "" {
		promptBuilder.WriteString(fmt.Sprintf("\n内容风格：%s\n", style))
	}

	// 关键词要求
	if len(keywords) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("\n请包含以下关键词：%s\n", strings.Join(keywords, ", ")))
	}

	// 自定义提示
	if customPrompt != "" {
		promptBuilder.WriteString(fmt.Sprintf("\n额外要求：%s\n", customPrompt))
	}

	// 输出格式
	promptBuilder.WriteString("\n请按以下格式输出：\n")
	promptBuilder.WriteString("标题：[标题内容]\n")
	promptBuilder.WriteString("内容：[正文内容]\n")
	promptBuilder.WriteString("摘要：[内容摘要]\n")
	promptBuilder.WriteString("关键词：[关键词1, 关键词2, ...]\n")
	promptBuilder.WriteString("标签：[标签1, 标签2, ...]\n")

	return promptBuilder.String()
}

// AdaptedContent 解析后的内容
type AdaptedContent struct {
	Title    string
	Content  string
	Summary  string
	Keywords []string
	Tags     []string
}

// parseGeneratedContent 解析生成的内容
func (s *ContentAdaptationService) parseGeneratedContent(content, platform string) *AdaptedContent {
	result := &AdaptedContent{
		Keywords: []string{},
		Tags:     []string{},
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "标题：") || strings.HasPrefix(line, "标题:") {
			result.Title = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "标题："), "标题:"))
		} else if strings.HasPrefix(line, "内容：") || strings.HasPrefix(line, "内容:") {
			result.Content = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "内容："), "内容:"))
		} else if strings.HasPrefix(line, "摘要：") || strings.HasPrefix(line, "摘要:") {
			result.Summary = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "摘要："), "摘要:"))
		} else if strings.HasPrefix(line, "关键词：") || strings.HasPrefix(line, "关键词:") {
			kwStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "关键词："), "关键词:"))
			result.Keywords = strings.Split(kwStr, ",")
			for i, kw := range result.Keywords {
				result.Keywords[i] = strings.TrimSpace(kw)
			}
		} else if strings.HasPrefix(line, "标签：") || strings.HasPrefix(line, "标签:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "标签："), "标签:"))
			result.Tags = strings.Split(tagStr, ",")
			for i, tag := range result.Tags {
				result.Tags[i] = strings.TrimSpace(tag)
			}
		} else if result.Content != "" && line != "" {
			// 继续添加内容
			result.Content += "\n" + line
		}
	}

	// 如果没有解析到标题，使用第一行作为标题
	if result.Title == "" && len(lines) > 0 {
		result.Title = lines[0]
	}

	return result
}

// calculateQualityScore 计算质量评分
func (s *ContentAdaptationService) calculateQualityScore(content *AdaptedContent, topic database.Topic) float64 {
	score := 0.0

	// 1. 标题质量（0-25分）
	if content.Title != "" {
		titleLen := len([]rune(content.Title))
		if titleLen >= 10 && titleLen <= 30 {
			score += 25
		} else if titleLen >= 5 && titleLen <= 50 {
			score += 15
		} else {
			score += 5
		}
	}

	// 2. 内容长度（0-25分）
	if content.Content != "" {
		contentLen := len([]rune(content.Content))
		if contentLen >= 100 && contentLen <= 2000 {
			score += 25
		} else if contentLen >= 50 && contentLen <= 3000 {
			score += 15
		} else {
			score += 5
		}
	}

	// 3. 关键词相关性（0-25分）
	if len(content.Keywords) > 0 {
		// 检查关键词是否与热点相关
		topicText := topic.Title + " " + topic.Description
		matchedKeywords := 0
		for _, kw := range content.Keywords {
			if strings.Contains(topicText, kw) {
				matchedKeywords++
			}
		}
		keywordScore := float64(matchedKeywords) / float64(len(content.Keywords)) * 25
		score += keywordScore
	}

	// 4. 标签完整性（0-25分）
	if len(content.Tags) >= 3 {
		score += 25
	} else if len(content.Tags) >= 1 {
		score += 15
	}

	return score
}

// BatchAdaptContent 批量适配内容
func (s *ContentAdaptationService) BatchAdaptContent(ctx context.Context, requests []*AdaptationRequest) ([]*AdaptationResult, error) {
	var results []*AdaptationResult

	for _, req := range requests {
		result, err := s.AdaptContent(ctx, req)
		if err != nil {
			logrus.Errorf("适配内容失败 [topic=%s]: %v", req.TopicID, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// GetAdaptationHistory 获取适配历史
func (s *ContentAdaptationService) GetAdaptationHistory(topicID string, limit int) ([]*AdaptationResult, error) {
	// 这里可以从数据库查询历史记录
	// 暂时返回空列表
	return []*AdaptationResult{}, nil
}
