package database

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// HotspotStorage 基于数据库的热点存储实现
type HotspotStorage struct {
	db *gorm.DB
}

// NewHotspotStorage 创建数据库热点存储
func NewHotspotStorage(db *gorm.DB) *HotspotStorage {
	if db == nil {
		db = GetDB()
	}
	return &HotspotStorage{db: db}
}

// TopicFilter 话题过滤条件
type TopicFilter struct {
	Category string
	Source   string
	Platform string
	MinHeat  int
	MaxHeat  int
	Trend    string
	Limit    int
	Offset   int
	SortBy   string
	SortDesc bool
}

// Save 保存话题列表
func (s *HotspotStorage) Save(topics []Topic) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range topics {
			topic := &topics[i]
			if topic.CreatedAt.IsZero() {
				topic.CreatedAt = time.Now()
			}
			topic.UpdatedAt = time.Now()

			// 使用 upsert
			if err := tx.Where("id = ?", topic.ID).Assign(topic).FirstOrCreate(&Topic{}).Error; err != nil {
				return err
			}
		}
		logrus.Infof("Saved %d topics", len(topics))
		return nil
	})
}

// SaveOne 保存单个话题
func (s *HotspotStorage) SaveOne(topic *Topic) error {
	if topic.CreatedAt.IsZero() {
		topic.CreatedAt = time.Now()
	}
	topic.UpdatedAt = time.Now()

	return s.db.Where("id = ?", topic.ID).Assign(topic).FirstOrCreate(&Topic{}).Error
}

// Get 获取单个话题
func (s *HotspotStorage) Get(id string) (*Topic, error) {
	var topic Topic
	err := s.db.Where("id = ?", id).First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// List 列出话题
func (s *HotspotStorage) List(filter TopicFilter) ([]Topic, int64, error) {
	query := s.db.Model(&Topic{})

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.Platform != "" {
		query = query.Where("platform_id = ?", filter.Platform)
	}
	if filter.MinHeat > 0 {
		query = query.Where("heat >= ?", filter.MinHeat)
	}
	if filter.MaxHeat > 0 {
		query = query.Where("heat <= ?", filter.MaxHeat)
	}
	if filter.Trend != "" {
		query = query.Where("trend = ?", filter.Trend)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	order := "heat desc"
	if filter.SortBy != "" {
		direction := "asc"
		if filter.SortDesc {
			direction = "desc"
		}
		order = filter.SortBy + " " + direction
	}
	query = query.Order(order)

	// 分页
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var topics []Topic
	err := query.Find(&topics).Error
	return topics, total, err
}

// Delete 删除话题
func (s *HotspotStorage) Delete(id string) error {
	return s.db.Delete(&Topic{}, "id = ?", id).Error
}

// DeleteBefore 删除指定时间之前的话题
func (s *HotspotStorage) DeleteBefore(t time.Time) (int64, error) {
	result := s.db.Where("created_at < ?", t).Delete(&Topic{})
	return result.RowsAffected, result.Error
}

// GetByTitle 根据标题获取话题
func (s *HotspotStorage) GetByTitle(title string) (*Topic, error) {
	var topic Topic
	err := s.db.Where("title LIKE ?", "%"+title+"%").First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

// GetNewSince 获取指定时间之后的新话题
func (s *HotspotStorage) GetNewSince(t time.Time) ([]Topic, error) {
	var topics []Topic
	err := s.db.Where("created_at > ?", t).Order("created_at desc").Find(&topics).Error
	return topics, err
}

// Count 获取话题总数
func (s *HotspotStorage) Count() (int64, error) {
	var count int64
	err := s.db.Model(&Topic{}).Count(&count).Error
	return count, err
}

// CountBySource 按来源统计话题数量
func (s *HotspotStorage) CountBySource() (map[string]int64, error) {
	type CountResult struct {
		Source string
		Count  int64
	}
	var results []CountResult
	err := s.db.Model(&Topic{}).Select("source, count(*) as count").Group("source").Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Source] = r.Count
	}
	return counts, nil
}

// SaveRankHistory 保存排名历史
func (s *HotspotStorage) SaveRankHistory(history *RankHistory) error {
	history.CreatedAt = time.Now()
	return s.db.Create(history).Error
}

// GetRankHistory 获取话题的排名历史
func (s *HotspotStorage) GetRankHistory(topicID string, limit int) ([]RankHistory, error) {
	var history []RankHistory
	query := s.db.Where("topic_id = ?", topicID).Order("crawl_time desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}

// SaveCrawlRecord 保存抓取记录
func (s *HotspotStorage) SaveCrawlRecord(record *CrawlRecord) error {
	record.CreatedAt = time.Now()
	return s.db.Create(record).Error
}

// GetLatestCrawlRecord 获取最新的抓取记录
func (s *HotspotStorage) GetLatestCrawlRecord() (*CrawlRecord, error) {
	var record CrawlRecord
	err := s.db.Order("crawl_time desc").First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateTrend 更新话题趋势
func (s *HotspotStorage) UpdateTrend(topicID string, trend string) error {
	return s.db.Model(&Topic{}).Where("id = ?", topicID).Updates(map[string]interface{}{
		"trend":     trend,
		"updated_at": time.Now(),
	}).Error
}

// GetTopTopics 获取热门话题
func (s *HotspotStorage) GetTopTopics(limit int) ([]Topic, error) {
	var topics []Topic
	err := s.db.Order("heat desc").Limit(limit).Find(&topics).Error
	return topics, err
}

// GetTrendingTopics 获取趋势上升的话题
func (s *HotspotStorage) GetTrendingTopics(limit int) ([]Topic, error) {
	var topics []Topic
	err := s.db.Where("trend = ?", "up").Order("heat desc").Limit(limit).Find(&topics).Error
	return topics, err
}

// TopicWithKeywords 带关键词的话题
type TopicWithKeywords struct {
	Topic
	KeywordsList []string
}

// GetTopicWithKeywords 获取话题并解析关键词
func (s *HotspotStorage) GetTopicWithKeywords(id string) (*TopicWithKeywords, error) {
	topic, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if topic == nil {
		return nil, nil
	}

	var keywords []string
	if topic.Keywords != "" {
		json.Unmarshal([]byte(topic.Keywords), &keywords)
	}

	return &TopicWithKeywords{
		Topic:        *topic,
		KeywordsList: keywords,
	}, nil
}

// ListWithKeywords 列出话题并解析关键词
func (s *HotspotStorage) ListWithKeywords(filter TopicFilter) ([]TopicWithKeywords, int64, error) {
	topics, total, err := s.List(filter)
	if err != nil {
		return nil, 0, err
	}

	result := make([]TopicWithKeywords, len(topics))
	for i, topic := range topics {
		var keywords []string
		if topic.Keywords != "" {
			json.Unmarshal([]byte(topic.Keywords), &keywords)
		}
		result[i] = TopicWithKeywords{
			Topic:        topic,
			KeywordsList: keywords,
		}
	}

	return result, total, nil
}
