package hotspot

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"publisher-core/ai/provider"
	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TrendAnalyzer 趋势分析器
type TrendAnalyzer struct {
	db        *gorm.DB
	aiService AIAnalyzer
	config    *TrendAnalyzerConfig
}

// TrendAnalyzerConfig 趋势分析器配置
type TrendAnalyzerConfig struct {
	KeywordMinFrequency int     `json:"keyword_min_frequency"` // 关键词最小出现频率
	SentimentThreshold  float64 `json:"sentiment_threshold"`   // 情感分析阈值
	TrendWindowSize     int     `json:"trend_window_size"`     // 趋势分析窗口大小（小时）
	MinHotnessScore     float64 `json:"min_hotness_score"`     // 最小热度评分
}

// DefaultTrendAnalyzerConfig 默认配置
func DefaultTrendAnalyzerConfig() *TrendAnalyzerConfig {
	return &TrendAnalyzerConfig{
		KeywordMinFrequency: 2,
		SentimentThreshold:  0.6,
		TrendWindowSize:     24,
		MinHotnessScore:     50.0,
	}
}

// NewTrendAnalyzer 创建趋势分析器
func NewTrendAnalyzer(db *gorm.DB, aiService AIAnalyzer, config *TrendAnalyzerConfig) *TrendAnalyzer {
	if config == nil {
		config = DefaultTrendAnalyzerConfig()
	}
	return &TrendAnalyzer{
		db:        db,
		aiService: aiService,
		config:    config,
	}
}

// AnalyzerTrendResult 分析器趋势分析结果
type AnalyzerTrendResult struct {
	TopicID          string                `json:"topic_id"`
	Keywords         []AnalyzerKeywordScore `json:"keywords"`
	Sentiment        AnalyzerSentiment     `json:"sentiment"`
	Trend            AnalyzerTrendInfo     `json:"trend"`
	RelevanceScore   float64               `json:"relevance_score"`
	ContentSuggestions []ContentSuggestion  `json:"content_suggestions"`
	CompetitorInsights []CompetitorInsight  `json:"competitor_insights"`
	AnalyzedAt       time.Time             `json:"analyzed_at"`
}

// AnalyzerKeywordScore 分析器关键词评分
type AnalyzerKeywordScore struct {
	Keyword   string  `json:"keyword"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
	Category  string  `json:"category"` // 技术、娱乐、社会等
}

// AnalyzerSentiment 分析器情感分析
type AnalyzerSentiment struct {
	Score       float64 `json:"score"`        // -1 到 1
	Label       string  `json:"label"`        // positive, negative, neutral
	Confidence  float64 `json:"confidence"`   // 0 到 1
	EmotionTags []string `json:"emotion_tags"` // 情感标签
}

// AnalyzerTrendInfo 分析器趋势信息
type AnalyzerTrendInfo struct {
	Direction     string    `json:"direction"`      // up, down, stable
	ChangeRate    float64   `json:"change_rate"`    // 变化率
	PeakTime      time.Time `json:"peak_time"`      // 峰值时间
	PredictedTrend string   `json:"predicted_trend"` // 预测趋势
	Confidence    float64   `json:"confidence"`
}

// ContentSuggestion 内容建议
type ContentSuggestion struct {
	Type        string   `json:"type"`         // article, video, image
	Title       string   `json:"title"`
	Keywords    []string `json:"keywords"`
	TargetPlatform string `json:"target_platform"`
	Priority    string   `json:"priority"`     // high, medium, low
	Reason      string   `json:"reason"`
}

// CompetitorInsight 竞品洞察
type CompetitorInsight struct {
	CompetitorName string   `json:"competitor_name"`
	RelatedTopics  []string `json:"related_topics"`
	ContentCount   int      `json:"content_count"`
	EngagementRate float64  `json:"engagement_rate"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
}

// AnalyzeTrend 分析热点趋势
func (a *TrendAnalyzer) AnalyzeTrend(ctx context.Context, topicID string) (*AnalyzerTrendResult, error) {
	// 获取热点数据
	var topic database.Topic
	if err := a.db.Where("id = ?", topicID).First(&topic).Error; err != nil {
		return nil, fmt.Errorf("获取热点失败: %w", err)
	}

	// 获取历史数据
	history, err := a.getTopicHistory(topicID, a.config.TrendWindowSize)
	if err != nil {
		logrus.Warnf("获取历史数据失败: %v", err)
	}

	// 1. 关键词提取
	keywords := a.extractKeywords(&topic, history)

	// 2. 情感分析
	sentiment := a.analyzeSentiment(ctx, &topic)

	// 3. 趋势分析
	trend := a.analyzeTrendDirection(&topic, history)

	// 4. 相关性评分
	relevanceScore := a.calculateRelevanceScore(&topic, keywords, sentiment)

	// 5. 内容建议
	contentSuggestions := a.generateContentSuggestions(&topic, keywords, sentiment)

	// 6. 竞品洞察
	competitorInsights := a.analyzeCompetitors(ctx, &topic, keywords)

	result := &AnalyzerTrendResult{
		TopicID:            topicID,
		Keywords:           keywords,
		Sentiment:          sentiment,
		Trend:              trend,
		RelevanceScore:     relevanceScore,
		ContentSuggestions: contentSuggestions,
		CompetitorInsights: competitorInsights,
		AnalyzedAt:         time.Now(),
	}

	// 保存分析结果
	if err := a.saveAnalysisResult(result); err != nil {
		logrus.Warnf("保存分析结果失败: %v", err)
	}

	return result, nil
}

// extractKeywords 提取关键词
func (a *TrendAnalyzer) extractKeywords(topic *database.Topic, history []database.RankHistory) []AnalyzerKeywordScore {
	// 合并标题和描述
	text := topic.Title + " " + topic.Description

	// 使用正则提取中文词汇
	re := regexp.MustCompile(`[\p{Han}]+`)
	words := re.FindAllString(text, -1)

	// 统计词频
	wordCount := make(map[string]int)
	for _, word := range words {
		if len(word) >= 2 { // 至少2个字
			wordCount[word]++
		}
	}

	// 转换为AnalyzerKeywordScore并排序
	var keywords []AnalyzerKeywordScore
	for word, count := range wordCount {
		if count >= a.config.KeywordMinFrequency {
			score := float64(count) * float64(topic.Heat) / 1000.0
			keywords = append(keywords, AnalyzerKeywordScore{
				Keyword:   word,
				Frequency: count,
				Score:     score,
				Category:  a.categorizeKeyword(word),
			})
		}
	}

	// 按评分排序
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Score > keywords[j].Score
	})

	// 返回前10个关键词
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}

	return keywords
}

// categorizeKeyword 分类关键词
func (a *TrendAnalyzer) categorizeKeyword(keyword string) string {
	// 简单的关键词分类
	techKeywords := []string{"AI", "人工智能", "技术", "科技", "互联网", "数据"}
	entertainmentKeywords := []string{"娱乐", "明星", "电影", "音乐", "综艺"}
	socialKeywords := []string{"社会", "民生", "政策", "经济", "教育"}

	for _, kw := range techKeywords {
		if strings.Contains(keyword, kw) {
			return "技术"
		}
	}
	for _, kw := range entertainmentKeywords {
		if strings.Contains(keyword, kw) {
			return "娱乐"
		}
	}
	for _, kw := range socialKeywords {
		if strings.Contains(keyword, kw) {
			return "社会"
		}
	}

	return "其他"
}

// analyzeSentiment 分析情感
func (a *TrendAnalyzer) analyzeSentiment(ctx context.Context, topic *database.Topic) AnalyzerSentiment {
	// 如果有AI服务，使用AI分析
	if a.aiService != nil {
		return a.analyzeSentimentWithAI(ctx, topic)
	}

	// 否则使用简单的规则分析
	return a.analyzeSentimentWithRules(topic)
}

// analyzeSentimentWithAI 使用AI分析情感
func (a *TrendAnalyzer) analyzeSentimentWithAI(ctx context.Context, topic *database.Topic) AnalyzerSentiment {
	prompt := fmt.Sprintf(`请分析以下热点话题的情感倾向：

标题：%s
描述：%s

请返回JSON格式的分析结果：
{
  "score": 0.5,  // -1到1之间，-1表示负面，1表示正面
  "label": "positive",  // positive, negative, neutral
  "confidence": 0.8,  // 0到1之间
  "emotion_tags": ["兴奋", "期待"]  // 情感标签数组
}`, topic.Title, topic.Description)

	result, err := a.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		MaxTokens: 200,
	})

	if err != nil {
		logrus.Warnf("AI情感分析失败: %v", err)
		return a.analyzeSentimentWithRules(topic)
	}

	var sentiment AnalyzerSentiment
	if err := json.Unmarshal([]byte(result.Content), &sentiment); err != nil {
		logrus.Warnf("解析AI结果失败: %v", err)
		return a.analyzeSentimentWithRules(topic)
	}

	return sentiment
}

// analyzeSentimentWithRules 使用规则分析情感
func (a *TrendAnalyzer) analyzeSentimentWithRules(topic *database.Topic) AnalyzerSentiment {
	// 简单的情感词典
	positiveWords := []string{"好", "优秀", "成功", "突破", "创新", "进步", "希望"}
	negativeWords := []string{"问题", "危机", "失败", "风险", "担忧", "困难", "挑战"}

	text := topic.Title + " " + topic.Description
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}
	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}

	score := 0.0
	if positiveCount+negativeCount > 0 {
		score = float64(positiveCount-negativeCount) / float64(positiveCount+negativeCount)
	}

	label := "neutral"
	if score > 0.3 {
		label = "positive"
	} else if score < -0.3 {
		label = "negative"
	}

	return AnalyzerSentiment{
		Score:      score,
		Label:      label,
		Confidence: 0.6,
		EmotionTags: []string{},
	}
}

// analyzeTrendDirection 分析趋势方向
func (a *TrendAnalyzer) analyzeTrendDirection(topic *database.Topic, history []database.RankHistory) AnalyzerTrendInfo {
	if len(history) < 2 {
		return AnalyzerTrendInfo{
			Direction:      "stable",
			ChangeRate:     0,
			PredictedTrend: "stable",
			Confidence:     0.5,
		}
	}

	// 计算热度变化率
	recentHeat := history[len(history)-1].Heat
	previousHeat := history[0].Heat
	changeRate := 0.0
	if previousHeat > 0 {
		changeRate = float64(recentHeat-previousHeat) / float64(previousHeat) * 100
	}

	// 判断趋势方向
	direction := "stable"
	if changeRate > 10 {
		direction = "up"
	} else if changeRate < -10 {
		direction = "down"
	}

	// 预测趋势
	predictedTrend := "stable"
	if changeRate > 20 {
		predictedTrend = "持续上升"
	} else if changeRate > 0 {
		predictedTrend = "缓慢上升"
	} else if changeRate < -20 {
		predictedTrend = "快速下降"
	} else if changeRate < 0 {
		predictedTrend = "缓慢下降"
	}

	// 找到峰值时间
	peakTime := topic.CreatedAt
	maxHeat := 0
	for _, h := range history {
		if h.Heat > maxHeat {
			maxHeat = h.Heat
			peakTime = h.CrawlTime
		}
	}

	return AnalyzerTrendInfo{
		Direction:       direction,
		ChangeRate:      changeRate,
		PeakTime:        peakTime,
		PredictedTrend:  predictedTrend,
		Confidence:      0.7,
	}
}

// calculateRelevanceScore 计算相关性评分
func (a *TrendAnalyzer) calculateRelevanceScore(topic *database.Topic, keywords []AnalyzerKeywordScore, sentiment AnalyzerSentiment) float64 {
	score := 0.0

	// 热度评分（0-40分）
	heatScore := float64(topic.Heat) / 100.0 * 40.0
	if heatScore > 40 {
		heatScore = 40
	}
	score += heatScore

	// 关键词评分（0-30分）
	keywordScore := 0.0
	for _, kw := range keywords {
		keywordScore += kw.Score
	}
	if keywordScore > 30 {
		keywordScore = 30
	}
	score += keywordScore

	// 情感评分（0-30分）
	sentimentScore := 0.0
	if sentiment.Label == "positive" {
		sentimentScore = 30.0 * sentiment.Confidence
	} else if sentiment.Label == "neutral" {
		sentimentScore = 20.0 * sentiment.Confidence
	} else {
		sentimentScore = 10.0 * sentiment.Confidence
	}
	score += sentimentScore

	return score
}

// generateContentSuggestions 生成内容建议
func (a *TrendAnalyzer) generateContentSuggestions(topic *database.Topic, keywords []AnalyzerKeywordScore, sentiment AnalyzerSentiment) []ContentSuggestion {
	var suggestions []ContentSuggestion

	// 提取前3个关键词
	topKeywords := make([]string, 0, 3)
	for i, kw := range keywords {
		if i >= 3 {
			break
		}
		topKeywords = append(topKeywords, kw.Keyword)
	}

	// 生成文章建议
	suggestions = append(suggestions, ContentSuggestion{
		Type:           "article",
		Title:          fmt.Sprintf("深度解析：%s", topic.Title),
		Keywords:       topKeywords,
		TargetPlatform: "今日头条",
		Priority:       "high",
		Reason:         "热点话题适合深度分析文章",
	})

	// 生成视频建议
	suggestions = append(suggestions, ContentSuggestion{
		Type:           "video",
		Title:          fmt.Sprintf("3分钟看懂：%s", topic.Title),
		Keywords:       topKeywords,
		TargetPlatform: "抖音",
		Priority:       "high",
		Reason:         "短视频形式适合快速传播热点",
	})

	// 生成图文建议
	suggestions = append(suggestions, ContentSuggestion{
		Type:           "image",
		Title:          fmt.Sprintf("一图看懂：%s", topic.Title),
		Keywords:       topKeywords,
		TargetPlatform: "小红书",
		Priority:       "medium",
		Reason:         "图文形式适合小红书平台",
	})

	return suggestions
}

// analyzeCompetitors 分析竞品
func (a *TrendAnalyzer) analyzeCompetitors(ctx context.Context, topic *database.Topic, keywords []AnalyzerKeywordScore) []CompetitorInsight {
	// 简化的竞品分析
	var insights []CompetitorInsight

	// 模拟竞品数据
	competitors := []string{"竞品A", "竞品B", "竞品C"}
	for _, comp := range competitors {
		insight := CompetitorInsight{
			CompetitorName: comp,
			RelatedTopics:  []string{topic.Title},
			ContentCount:   5 + len(keywords),
			EngagementRate: 0.05 + float64(len(keywords))*0.01,
			Strengths:      []string{"内容质量高", "更新及时"},
			Weaknesses:     []string{"互动较少"},
		}
		insights = append(insights, insight)
	}

	return insights
}

// getTopicHistory 获取热点历史数据
func (a *TrendAnalyzer) getTopicHistory(topicID string, hours int) ([]database.RankHistory, error) {
	var history []database.RankHistory
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	err := a.db.Where("topic_id = ? AND crawl_time >= ?", topicID, startTime).
		Order("crawl_time ASC").
		Find(&history).Error

	return history, err
}

// saveAnalysisResult 保存分析结果
func (a *TrendAnalyzer) saveAnalysisResult(result *AnalyzerTrendResult) error {
	// 将结果序列化为JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// 更新热点的元数据
	return a.db.Model(&database.Topic{}).
		Where("id = ?", result.TopicID).
		Update("keywords", string(resultJSON)).Error
}
