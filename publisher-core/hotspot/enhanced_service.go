package hotspot

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"publisher-core/ai/provider"
	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// EnhancedService 增强版热点服务
type EnhancedService struct {
	mu          sync.RWMutex
	sources     map[string]SourceInterface
	storage     *database.HotspotStorage
	aiService   AIAnalyzer
	notifyService Notifier
	db          *gorm.DB
	config      *EnhancedConfig
}

// AIAnalyzer AI 分析接口
type AIAnalyzer interface {
	Generate(ctx context.Context, opts *provider.GenerateOptions) (*provider.GenerateResult, error)
}

// Notifier 通知接口
type Notifier interface {
	Send(ctx context.Context, channelType, title, content string) error
}

// EnhancedConfig 增强服务配置
type EnhancedConfig struct {
	// 热度计算权重
	RankWeight     float64 `json:"rank_weight"`
	FrequencyWeight float64 `json:"frequency_weight"`
	HotnessWeight  float64 `json:"hotness_weight"`
	// 趋势分析窗口
	TrendWindow int `json:"trend_window"` // 分析最近 N 次抓取
	// 自动通知
	AutoNotify      bool   `json:"auto_notify"`
	NotifyChannel   string `json:"notify_channel"`
	NotifyThreshold int    `json:"notify_threshold"` // 热度阈值
}

// DefaultEnhancedConfig 默认配置
func DefaultEnhancedConfig() *EnhancedConfig {
	return &EnhancedConfig{
		RankWeight:      0.6,
		FrequencyWeight: 0.3,
		HotnessWeight:   0.1,
		TrendWindow:     10,
		AutoNotify:      false,
		NotifyChannel:   "feishu",
		NotifyThreshold: 80,
	}
}

// NewEnhancedService 创建增强版热点服务
func NewEnhancedService(db *gorm.DB, config *EnhancedConfig) *EnhancedService {
	if config == nil {
		config = DefaultEnhancedConfig()
	}
	return &EnhancedService{
		sources: make(map[string]SourceInterface),
		storage: database.NewHotspotStorage(db),
		db:      db,
		config:  config,
	}
}

// SetAIService 设置 AI 服务
func (s *EnhancedService) SetAIService(ai AIAnalyzer) {
	s.aiService = ai
}

// SetNotifyService 设置通知服务
func (s *EnhancedService) SetNotifyService(notifier Notifier) {
	s.notifyService = notifier
}

// RegisterSource 注册数据源
func (s *EnhancedService) RegisterSource(source SourceInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sources[source.ID()] = source
	logrus.Infof("Registered hotspot source: %s", source.ID())
}

// FetchAndSave 抓取并保存数据
func (s *EnhancedService) FetchAndSave(ctx context.Context, sourceID string, maxItems int) ([]database.Topic, error) {
	s.mu.RLock()
	source, ok := s.sources[sourceID]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrSourceNotFound
	}

	// 抓取数据
	topics, err := source.Fetch(ctx, maxItems)
	if err != nil {
		return nil, err
	}

	// 转换为数据库模型
	dbTopics := s.convertTopics(topics)

	// 保存到数据库
	if err := s.saveWithHistory(ctx, dbTopics, sourceID); err != nil {
		return nil, err
	}

	return dbTopics, nil
}

// FetchFromAllSources 从所有数据源抓取
func (s *EnhancedService) FetchFromAllSources(ctx context.Context, maxItemsPerSource int) (map[string][]database.Topic, error) {
	s.mu.RLock()
	sources := make([]SourceInterface, 0, len(s.sources))
	for _, src := range s.sources {
		if src.IsEnabled() {
			sources = append(sources, src)
		}
	}
	s.mu.RUnlock()

	results := make(map[string][]database.Topic)
	var allTopics []database.Topic

	for _, source := range sources {
		topics, err := s.FetchAndSave(ctx, source.ID(), maxItemsPerSource)
		if err != nil {
			logrus.Warnf("Fetch from %s failed: %v", source.ID(), err)
			continue
		}
		results[source.ID()] = topics
		allTopics = append(allTopics, topics...)
	}

	// 记录抓取记录
	s.recordCrawl(len(allTopics), "success", "")

	logrus.Infof("Fetched %d topics from %d sources", len(allTopics), len(sources))
	return results, nil
}

// saveWithHistory 保存数据并记录历史
func (s *EnhancedService) saveWithHistory(ctx context.Context, topics []database.Topic, sourceID string) error {
	now := time.Now()

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range topics {
			topic := &topics[i]
			topic.LastCrawlTime = now

			// 检查是否已存在
			var existing database.Topic
			err := tx.Where("id = ?", topic.ID).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// 新话题
				topic.FirstCrawlTime = now
				topic.CreatedAt = now
				topic.UpdatedAt = now
				if err := tx.Create(topic).Error; err != nil {
					return err
				}
			} else if err == nil {
				// 已存在，更新
				topic.CreatedAt = existing.CreatedAt
				topic.FirstCrawlTime = existing.FirstCrawlTime
				topic.UpdatedAt = now
				if err := tx.Save(topic).Error; err != nil {
					return err
				}
			} else {
				return err
			}

			// 记录排名历史
			history := &database.RankHistory{
				TopicID:   topic.ID,
				Rank:      i + 1,
				Heat:      topic.Heat,
				CrawlTime: now,
				CreatedAt: now,
			}
			if err := tx.Create(history).Error; err != nil {
				logrus.Warnf("Failed to save rank history: %v", err)
			}
		}

		return nil
	})
}

// recordCrawl 记录抓取
func (s *EnhancedService) recordCrawl(totalItems int, status, errMsg string) {
	record := &database.CrawlRecord{
		CrawlTime:  time.Now(),
		TotalItems: totalItems,
		Status:     status,
		Error:      errMsg,
		CreatedAt:  time.Now(),
	}
	if err := s.storage.SaveCrawlRecord(record); err != nil {
		logrus.Warnf("Failed to save crawl record: %v", err)
	}
}

// CalculateHeat 计算综合热度
func (s *EnhancedService) CalculateHeat(rank, frequency, hotness int) int {
	// 排名分数
	rankScore := 100 - (rank-1)*2
	if rankScore < 0 {
		rankScore = 0
	}

	// 频次分数
	freqScore := frequency * 20
	if freqScore > 100 {
		freqScore = 100
	}

	// 热度分数
	hotScore := hotness / 10000
	if hotScore > 100 {
		hotScore = 100
	}

	return int(float64(rankScore)*s.config.RankWeight +
		float64(freqScore)*s.config.FrequencyWeight +
		float64(hotScore)*s.config.HotnessWeight)
}

// AnalyzeTrend 分析趋势
func (s *EnhancedService) AnalyzeTrend(topicID string) (string, error) {
	history, err := s.storage.GetRankHistory(topicID, s.config.TrendWindow)
	if err != nil {
		return "", err
	}

	if len(history) < 2 {
		return "new", nil
	}

	// 计算排名变化
	latest := history[0].Rank
	previous := history[1].Rank

	// 计算热度变化趋势
	var heatTrend float64
	if len(history) >= 3 {
		latestHeat := float64(history[0].Heat)
		previousHeat := float64(history[1].Heat)
		if previousHeat > 0 {
			heatTrend = (latestHeat - previousHeat) / previousHeat * 100
		}
	}

	// 综合判断趋势
	if latest < previous-5 || heatTrend > 20 {
		return "up", nil
	} else if latest > previous+5 || heatTrend < -20 {
		return "down", nil
	} else if latest <= 10 {
		return "hot", nil
	}

	return "stable", nil
}

// UpdateAllTrends 更新所有话题趋势
func (s *EnhancedService) UpdateAllTrends() error {
	topics, _, err := s.storage.List(database.TopicFilter{Limit: 1000})
	if err != nil {
		return err
	}

	for _, topic := range topics {
		trend, err := s.AnalyzeTrend(topic.ID)
		if err != nil {
			logrus.Warnf("Failed to analyze trend for %s: %v", topic.ID, err)
			continue
		}

		if trend != topic.Trend {
			if err := s.storage.UpdateTrend(topic.ID, trend); err != nil {
				logrus.Warnf("Failed to update trend for %s: %v", topic.ID, err)
			}
		}
	}

	return nil
}

// AIAnalyze AI 分析热点
func (s *EnhancedService) AIAnalyze(ctx context.Context, topicIDs []string) (*AIAnalysisResult, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI service not configured")
	}

	// 获取话题详情
	var topics []database.Topic
	if err := s.db.Where("id IN ?", topicIDs).Find(&topics).Error; err != nil {
		return nil, err
	}

	if len(topics) == 0 {
		return nil, fmt.Errorf("no topics found")
	}

	// 构建分析提示词
	prompt := s.buildAnalysisPrompt(topics)

	// 调用 AI 服务
	result, err := s.aiService.Generate(ctx, &provider.GenerateOptions{
		Messages: []provider.Message{
			{Role: provider.RoleUser, Content: prompt},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, err
	}

	// 解析结果
	var analysis AIAnalysisResult
	if err := json.Unmarshal([]byte(result.Content), &analysis); err != nil {
		// 如果解析失败，返回原始内容
		analysis.Summary = result.Content
	}

	return &analysis, nil
}

// buildAnalysisPrompt 构建分析提示词
func (s *EnhancedService) buildAnalysisPrompt(topics []database.Topic) string {
	var topicList string
	for i, t := range topics {
		topicList += fmt.Sprintf("%d. %s (热度: %d, 趋势: %s, 来源: %s)\n",
			i+1, t.Title, t.Heat, t.Trend, t.Source)
	}

	return fmt.Sprintf(`请分析以下热点话题，并以 JSON 格式返回结果：

话题列表：
%s

请返回以下 JSON 格式：
{
  "summary": "整体趋势总结",
  "key_points": ["要点1", "要点2", "要点3"],
  "categories": {
    "科技": ["话题标题1"],
    "娱乐": ["话题标题2"]
  },
  "recommendations": [
    {
      "topic_id": "话题ID",
      "reason": "推荐理由",
      "suitability": 85
    }
  ],
  "alerts": ["需要注意的问题"]
}`, topicList)
}

// AIAnalysisResult AI 分析结果
type AIAnalysisResult struct {
	Summary        string                            `json:"summary"`
	KeyPoints      []string                          `json:"key_points"`
	Categories     map[string][]string               `json:"categories"`
	Recommendations []AIRecommendation               `json:"recommendations"`
	Alerts         []string                          `json:"alerts"`
}

// AIRecommendation AI 推荐
type AIRecommendation struct {
	TopicID     string `json:"topic_id"`
	Reason      string `json:"reason"`
	Suitability int    `json:"suitability"`
}

// GetTrendingTopics 获取趋势上升的话题
func (s *EnhancedService) GetTrendingTopics(limit int) ([]database.Topic, error) {
	return s.storage.GetTrendingTopics(limit)
}

// GetTopTopics 获取热门话题
func (s *EnhancedService) GetTopTopics(limit int) ([]database.Topic, error) {
	return s.storage.GetTopTopics(limit)
}

// GetTopicHistory 获取话题历史
func (s *EnhancedService) GetTopicHistory(topicID string, limit int) ([]database.RankHistory, error) {
	return s.storage.GetRankHistory(topicID, limit)
}

// List 列出话题
func (s *EnhancedService) List(filter database.TopicFilter) ([]database.Topic, int64, error) {
	return s.storage.List(filter)
}

// Get 获取单个话题
func (s *EnhancedService) Get(id string) (*database.Topic, error) {
	return s.storage.Get(id)
}

// Delete 删除话题
func (s *EnhancedService) Delete(id string) error {
	return s.storage.Delete(id)
}

// CleanupOldData 清理旧数据
func (s *EnhancedService) CleanupOldData(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	return s.storage.DeleteBefore(cutoff)
}

// convertTopics 转换话题格式
func (s *EnhancedService) convertTopics(topics []Topic) []database.Topic {
	result := make([]database.Topic, len(topics))
	for i, t := range topics {
		keywordsJSON, _ := json.Marshal(t.Keywords)

		result[i] = database.Topic{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Category:    string(t.Category),
			Heat:        t.Heat,
			Trend:       string(t.Trend),
			Source:      t.Source,
			SourceID:    t.SourceID,
			SourceURL:   t.SourceURL,
			OriginalURL: t.OriginalURL,
			Keywords:    string(keywordsJSON),
			Suitability: t.Suitability,
			PublishedAt: t.PublishedAt,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	}
	return result
}

// NotifyHotTopics 通知热门话题
func (s *EnhancedService) NotifyHotTopics(ctx context.Context, threshold int) error {
	if s.notifyService == nil {
		return fmt.Errorf("notify service not configured")
	}

	topics, err := s.storage.GetTopTopics(20)
	if err != nil {
		return err
	}

	var hotTopics []database.Topic
	for _, t := range topics {
		if t.Heat >= threshold {
			hotTopics = append(hotTopics, t)
		}
	}

	if len(hotTopics) == 0 {
		return nil
	}

	// 构建通知内容
	content := s.buildNotifyContent(hotTopics)

	return s.notifyService.Send(ctx, s.config.NotifyChannel, "热点话题推送", content)
}

// buildNotifyContent 构建通知内容
func (s *EnhancedService) buildNotifyContent(topics []database.Topic) string {
	content := fmt.Sprintf("【热点推送】%s\n\n", time.Now().Format("2006-01-02 15:04"))
	for i, t := range topics {
		content += fmt.Sprintf("%d. %s\n   热度: %d | 趋势: %s | 来源: %s\n\n",
			i+1, t.Title, t.Heat, t.Trend, t.Source)
	}
	return content
}

// GetStats 获取统计信息
func (s *EnhancedService) GetStats() (*HotspotStats, error) {
	stats := &HotspotStats{}

	// 总话题数
	total, err := s.storage.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalTopics = total

	// 按来源统计
	bySource, err := s.storage.CountBySource()
	if err != nil {
		return nil, err
	}
	stats.BySource = bySource

	// 趋势统计
	var upCount, downCount, stableCount, newCount int64
	s.db.Model(&database.Topic{}).Where("trend = ?", "up").Count(&upCount)
	s.db.Model(&database.Topic{}).Where("trend = ?", "down").Count(&downCount)
	s.db.Model(&database.Topic{}).Where("trend = ?", "stable").Count(&stableCount)
	s.db.Model(&database.Topic{}).Where("trend = ?", "new").Count(&newCount)

	stats.ByTrend = map[string]int64{
		"up":     upCount,
		"down":   downCount,
		"stable": stableCount,
		"new":    newCount,
	}

	// 最新抓取记录
	latestCrawl, _ := s.storage.GetLatestCrawlRecord()
	stats.LastCrawl = latestCrawl

	return stats, nil
}

// HotspotStats 热点统计
type HotspotStats struct {
	TotalTopics int64                  `json:"total_topics"`
	BySource    map[string]int64       `json:"by_source"`
	ByTrend     map[string]int64       `json:"by_trend"`
	LastCrawl   *database.CrawlRecord  `json:"last_crawl"`
}
