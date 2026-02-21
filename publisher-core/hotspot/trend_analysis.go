package hotspot

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"publisher-tools/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TrendAnalysisService 热点趋势分析服务
type TrendAnalysisService struct {
	db        *gorm.DB
	aiService AIAnalyzer
	config    *TrendAnalysisConfig
}

// TrendAnalysisConfig 趋势分析配置
type TrendAnalysisConfig struct {
	KeywordMinFrequency int     `json:"keyword_min_frequency"` // 关键词最小出现频率
	SentimentThreshold  float64 `json:"sentiment_threshold"`   // 情感分析阈值
	RelevanceThreshold  float64 `json:"relevance_threshold"`   // 相关性阈值
	TrendWindowSize     int     `json:"trend_window_size"`     // 趋势分析窗口大小（小时）
	MaxKeywords         int     `json:"max_keywords"`          // 最大关键词数量
}

// DefaultTrendAnalysisConfig 默认配置
func DefaultTrendAnalysisConfig() *TrendAnalysisConfig {
	return &TrendAnalysisConfig{
		KeywordMinFrequency: 2,
		SentimentThreshold:  0.6,
		RelevanceThreshold:  0.5,
		TrendWindowSize:     24,
		MaxKeywords:         20,
	}
}

// NewTrendAnalysisService 创建趋势分析服务
func NewTrendAnalysisService(db *gorm.DB, config *TrendAnalysisConfig) *TrendAnalysisService {
	if config == nil {
		config = DefaultTrendAnalysisConfig()
	}
	return &TrendAnalysisService{
		db:     db,
		config: config,
	}
}

// SetAIService 设置AI服务
func (s *TrendAnalysisService) SetAIService(ai AIAnalyzer) {
	s.aiService = ai
}

// TrendAnalysisResult 趋势分析结果
type TrendAnalysisResult struct {
	TopicID          string             `json:"topic_id"`
	Title            string             `json:"title"`
	Keywords         []KeywordScore     `json:"keywords"`
	Sentiment        SentimentAnalysis  `json:"sentiment"`
	Trend            TrendInfo          `json:"trend"`
	RelevanceScore   float64            `json:"relevance_score"`
	ContentSuggestions []string         `json:"content_suggestions"`
	RelatedTopics    []string           `json:"related_topics"`
	AnalyzedAt       time.Time          `json:"analyzed_at"`
}

// KeywordScore 关键词评分
type KeywordScore struct {
	Keyword   string  `json:"keyword"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
	Type      string  `json:"type"` // person, location, organization, event, concept
}

// SentimentAnalysis 情感分析
type SentimentAnalysis struct {
	Score       float64 `json:"score"`        // -1 到 1
	Label       string  `json:"label"`        // positive, negative, neutral
	Confidence  float64 `json:"confidence"`   // 0 到 1
	Emotions    map[string]float64 `json:"emotions"` // joy, anger, fear, sadness, surprise
}

// TrendInfo 趋势信息
type TrendInfo struct {
	Direction     string    `json:"direction"`      // up, down, stable
	Speed         float64   `json:"speed"`          // 变化速度
	Acceleration  float64   `json:"acceleration"`   // 加速度
	PeakTime      time.Time `json:"peak_time"`      // 预计峰值时间
	DecayRate     float64   `json:"decay_rate"`     // 衰减率
	PredictedLife int       `json:"predicted_life"` // 预计生命周期（小时）
}

// AnalyzeTopic 分析单个热点
func (s *TrendAnalysisService) AnalyzeTopic(ctx context.Context, topicID string) (*TrendAnalysisResult, error) {
	// 获取热点信息
	var topic database.Topic
	if err := s.db.Where("id = ?", topicID).First(&topic).Error; err != nil {
		return nil, fmt.Errorf("获取热点失败: %w", err)
	}

	// 获取历史数据
	var history []database.RankHistory
	if err := s.db.Where("topic_id = ?", topicID).
		Order("crawl_time DESC").
		Limit(s.config.TrendWindowSize).
		Find(&history).Error; err != nil {
		return nil, fmt.Errorf("获取历史数据失败: %w", err)
	}

	result := &TrendAnalysisResult{
		TopicID:    topicID,
		Title:      topic.Title,
		AnalyzedAt: time.Now(),
	}

	// 1. 关键词提取
	keywords := s.extractKeywords(topic.Title, topic.Description)
	result.Keywords = keywords

	// 2. 情感分析
	sentiment := s.analyzeSentiment(topic.Title, topic.Description)
	result.Sentiment = sentiment

	// 3. 趋势分析
	trend := s.analyzeTrend(history)
	result.Trend = trend

	// 4. 相关性评分
	relevance := s.calculateRelevance(topic, keywords, sentiment)
	result.RelevanceScore = relevance

	// 5. 内容建议
	if s.aiService != nil {
		suggestions := s.generateContentSuggestions(ctx, topic, keywords)
		result.ContentSuggestions = suggestions
	}

	// 6. 相关热点
	related := s.findRelatedTopics(topicID, keywords)
	result.RelatedTopics = related

	return result, nil
}

// extractKeywords 提取关键词
func (s *TrendAnalysisService) extractKeywords(title, description string) []KeywordScore {
	text := title + " " + description
	
	// 简单的关键词提取（实际项目中可以使用更复杂的NLP算法）
	words := s.tokenize(text)
	
	// 统计词频
	wordFreq := make(map[string]int)
	for _, word := range words {
		if len(word) >= 2 { // 过滤单字
			wordFreq[word]++
		}
	}

	// 转换为KeywordScore并排序
	var keywords []KeywordScore
	for word, freq := range wordFreq {
		if freq >= s.config.KeywordMinFrequency {
			keywords = append(keywords, KeywordScore{
				Keyword:   word,
				Frequency: freq,
				Score:     float64(freq) / float64(len(words)) * 100,
				Type:      s.classifyKeyword(word),
			})
		}
	}

	// 按频率排序
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Frequency > keywords[j].Frequency
	})

	// 限制数量
	if len(keywords) > s.config.MaxKeywords {
		keywords = keywords[:s.config.MaxKeywords]
	}

	return keywords
}

// tokenize 分词
func (s *TrendAnalysisService) tokenize(text string) []string {
	// 简单的分词（实际项目中可以使用jieba等专业分词库）
	// 移除标点符号
	reg := regexp.MustCompile(`[^\w\s\u4e00-\u9fa5]`)
	text = reg.ReplaceAllString(text, " ")
	
	// 分词
	words := strings.Fields(text)
	return words
}

// classifyKeyword 分类关键词
func (s *TrendAnalysisService) classifyKeyword(keyword string) string {
	// 简单的分类逻辑（实际项目中可以使用NER等技术）
	// 这里使用简单的规则
	if strings.ContainsAny(keyword, "省市县区镇村") {
		return "location"
	}
	if strings.ContainsAny(keyword, "公司集团企业") {
		return "organization"
	}
	return "concept"
}

// analyzeSentiment 情感分析
func (s *TrendAnalysisService) analyzeSentiment(title, description string) SentimentAnalysis {
	// 简单的情感分析（实际项目中可以使用AI模型）
	text := title + " " + description
	
	// 定义情感词汇
	positiveWords := []string{"好", "优秀", "成功", "突破", "创新", "领先", "增长", "提升"}
	negativeWords := []string{"失败", "问题", "危机", "下降", "损失", "风险", "挑战", "困难"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}
	
	// 计算情感分数
	total := positiveCount + negativeCount
	score := 0.0
	if total > 0 {
		score = float64(positiveCount-negativeCount) / float64(total)
	}
	
	// 确定情感标签
	label := "neutral"
	if score > s.config.SentimentThreshold {
		label = "positive"
	} else if score < -s.config.SentimentThreshold {
		label = "negative"
	}
	
	return SentimentAnalysis{
		Score:      score,
		Label:      label,
		Confidence: 0.8, // 简单方法的置信度
		Emotions: map[string]float64{
			"joy":     float64(positiveCount) * 0.1,
			"anger":   float64(negativeCount) * 0.05,
			"fear":    float64(negativeCount) * 0.03,
			"sadness": float64(negativeCount) * 0.02,
			"surprise": 0.1,
		},
	}
}

// analyzeTrend 分析趋势
func (s *TrendAnalysisService) analyzeTrend(history []database.RankHistory) TrendInfo {
	if len(history) < 2 {
		return TrendInfo{
			Direction:     "stable",
			Speed:         0,
			Acceleration:  0,
			PredictedLife: 24,
		}
	}

	// 计算排名变化
	rankChanges := make([]float64, len(history)-1)
	for i := 0; i < len(history)-1; i++ {
		rankChanges[i] = float64(history[i+1].Rank - history[i].Rank)
	}

	// 计算平均速度
	var totalSpeed float64
	for _, change := range rankChanges {
		totalSpeed += change
	}
	avgSpeed := totalSpeed / float64(len(rankChanges))

	// 计算加速度
	var totalAccel float64
	for i := 1; i < len(rankChanges); i++ {
		totalAccel += rankChanges[i] - rankChanges[i-1]
	}
	avgAccel := totalAccel / float64(len(rankChanges)-1)

	// 确定趋势方向
	direction := "stable"
	if avgSpeed < -1 {
		direction = "up" // 排名上升（数值减小）
	} else if avgSpeed > 1 {
		direction = "down" // 排名下降（数值增大）
	}

	// 预测生命周期
	predictedLife := 24 // 默认24小时
	if direction == "up" {
		predictedLife = int(24 + avgSpeed*-2) // 上升趋势，生命周期更长
	} else if direction == "down" {
		predictedLife = int(24 - avgSpeed*2) // 下降趋势，生命周期更短
	}

	// 计算衰减率
	decayRate := 0.1
	if len(history) > 5 {
		recentAvg := float64(history[0].Heat+history[1].Heat+history[2].Heat) / 3
		olderAvg := float64(history[3].Heat+history[4].Heat+history[5].Heat) / 3
		if olderAvg > 0 {
			decayRate = (olderAvg - recentAvg) / olderAvg
		}
	}

	return TrendInfo{
		Direction:     direction,
		Speed:         avgSpeed,
		Acceleration:  avgAccel,
		PredictedLife: predictedLife,
		DecayRate:     decayRate,
	}
}

// calculateRelevance 计算相关性评分
func (s *TrendAnalysisService) calculateRelevance(topic database.Topic, keywords []KeywordScore, sentiment SentimentAnalysis) float64 {
	score := 0.0

	// 1. 热度评分（0-40分）
	heatScore := float64(topic.Heat) / 100 * 40
	if heatScore > 40 {
		heatScore = 40
	}
	score += heatScore

	// 2. 关键词质量评分（0-30分）
	keywordScore := 0.0
	for _, kw := range keywords {
		keywordScore += kw.Score
	}
	keywordScore = keywordScore / float64(len(keywords)) * 30
	if keywordScore > 30 {
		keywordScore = 30
	}
	score += keywordScore

	// 3. 情感强度评分（0-20分）
	sentimentScore := (sentiment.Confidence * 20)
	score += sentimentScore

	// 4. 趋势评分（0-10分）
	if topic.Trend == "up" || topic.Trend == "hot" {
		score += 10
	} else if topic.Trend == "stable" {
		score += 5
	}

	return score
}

// generateContentSuggestions 生成内容建议
func (s *TrendAnalysisService) generateContentSuggestions(ctx context.Context, topic database.Topic, keywords []KeywordScore) []string {
	if s.aiService == nil {
		return []string{}
	}

	// 构建提示词
	keywordList := make([]string, len(keywords))
	for i, kw := range keywords {
		keywordList[i] = kw.Keyword
	}

	prompt := fmt.Sprintf(`基于以下热点话题，生成3个内容创作建议：

热点标题：%s
关键词：%s

请提供具体的内容创作方向和角度。`, topic.Title, strings.Join(keywordList, ", "))

	// 调用AI服务
	result, err := s.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		MaxTokens: 500,
	})

	if err != nil {
		logrus.Errorf("生成内容建议失败: %v", err)
		return []string{}
	}

	// 解析结果
	suggestions := strings.Split(result.Content, "\n")
	var validSuggestions []string
	for _, suggestion := range suggestions {
		suggestion = strings.TrimSpace(suggestion)
		if suggestion != "" && len(suggestion) > 10 {
			validSuggestions = append(validSuggestions, suggestion)
		}
	}

	return validSuggestions
}

// findRelatedTopics 查找相关热点
func (s *TrendAnalysisService) findRelatedTopics(topicID string, keywords []KeywordScore) []string {
	if len(keywords) == 0 {
		return []string{}
	}

	// 构建查询条件
	var relatedTopics []database.Topic
	query := s.db.Model(&database.Topic{}).Where("id != ?", topicID)

	for _, kw := range keywords[:5] { // 只使用前5个关键词
		query = query.Or("title LIKE ? OR description LIKE ?", "%"+kw.Keyword+"%", "%"+kw.Keyword+"%")
	}

	if err := query.Limit(10).Find(&relatedTopics).Error; err != nil {
		logrus.Errorf("查找相关热点失败: %v", err)
		return []string{}
	}

	var topicIDs []string
	for _, topic := range relatedTopics {
		topicIDs = append(topicIDs, topic.ID)
	}

	return topicIDs
}

// BatchAnalyzeTopics 批量分析热点
func (s *TrendAnalysisService) BatchAnalyzeTopics(ctx context.Context, topicIDs []string) ([]*TrendAnalysisResult, error) {
	var results []*TrendAnalysisResult

	for _, topicID := range topicIDs {
		result, err := s.AnalyzeTopic(ctx, topicID)
		if err != nil {
			logrus.Errorf("分析热点 %s 失败: %v", topicID, err)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTrendReport 生成趋势报告
func (s *TrendAnalysisService) GetTrendReport(ctx context.Context, startTime, endTime time.Time) (*TrendReport, error) {
	// 获取时间范围内的热点
	var topics []database.Topic
	if err := s.db.Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Order("heat DESC").
		Limit(50).
		Find(&topics).Error; err != nil {
		return nil, fmt.Errorf("获取热点失败: %w", err)
	}

	report := &TrendReport{
		StartTime: startTime,
		EndTime:   endTime,
		TotalTopics: len(topics),
	}

	// 统计趋势分布
	trendDist := make(map[string]int)
	for _, topic := range topics {
		trendDist[topic.Trend]++
	}
	report.TrendDistribution = trendDist

	// 提取热门关键词
	allKeywords := make(map[string]int)
	for _, topic := range topics {
		keywords := s.extractKeywords(topic.Title, topic.Description)
		for _, kw := range keywords {
			allKeywords[kw.Keyword] += kw.Frequency
		}
	}

	// 排序关键词
	var keywordList []KeywordScore
	for keyword, freq := range allKeywords {
		keywordList = append(keywordList, KeywordScore{
			Keyword:   keyword,
			Frequency: freq,
			Score:     float64(freq),
		})
	}
	sort.Slice(keywordList, func(i, j int) bool {
		return keywordList[i].Frequency > keywordList[j].Frequency
	})

	if len(keywordList) > 20 {
		keywordList = keywordList[:20]
	}
	report.TopKeywords = keywordList

	return report, nil
}

// TrendReport 趋势报告
type TrendReport struct {
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	TotalTopics       int                `json:"total_topics"`
	TrendDistribution map[string]int     `json:"trend_distribution"`
	TopKeywords       []KeywordScore     `json:"top_keywords"`
}
