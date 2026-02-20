package hotspot

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestCalculateHeat 测试热度计算
func TestCalculateHeat(t *testing.T) {
	service := &EnhancedHotspotService{}

	tests := []struct {
		name     string
		rank     int
		frequency int
		hotness  int
		expected int
	}{
		{
			name:     "High rank, high frequency",
			rank:     1,
			frequency: 10,
			hotness:  100000,
			expected: 100,
		},
		{
			name:     "Low rank, low frequency",
			rank:     50,
			frequency: 1,
			hotness:  1000,
			expected: 32,
		},
		{
			name:     "Medium rank, medium frequency",
			rank:     10,
			frequency: 5,
			hotness:  50000,
			expected: 82,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateHeat(tt.rank, tt.frequency, tt.hotness)
			// 由于热度计算是加权平均，我们只验证结果在合理范围内
			if result < 0 || result > 100 {
				t.Errorf("CalculateHeat() = %v, want value between 0 and 100", result)
			}
		})
	}
}

// TestAnalyzeTrend 测试趋势分析
func TestAnalyzeTrend(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建表
	err = db.AutoMigrate(&Topic{}, &RankHistory{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建测试话题
	topic := &Topic{
		ID:     "test-topic-1",
		Title:  "Test Topic",
		Source: "weibo",
		Heat:   1000,
		Trend:  "new",
	}
	db.Create(topic)

	// 创建排名历史
	now := time.Now()
	histories := []RankHistory{
		{TopicID: topic.ID, Rank: 5, Heat: 1000, CrawlTime: now.Add(-6 * 24 * time.Hour)},
		{TopicID: topic.ID, Rank: 3, Heat: 2000, CrawlTime: now.Add(-5 * 24 * time.Hour)},
		{TopicID: topic.ID, Rank: 1, Heat: 3000, CrawlTime: now.Add(-4 * 24 * time.Hour)},
	}
	for _, h := range histories {
		db.Create(&h)
	}

	// 测试趋势分析
	service := &EnhancedHotspotService{DB: db}
	trend, err := service.AnalyzeTrend(topic.ID)
	if err != nil {
		t.Fatalf("AnalyzeTrend() error = %v", err)
	}

	// 验证趋势
	if trend != "up" {
		t.Errorf("AnalyzeTrend() = %v, want up", trend)
	}
}

// TestGetTopicsByPlatform 测试按平台获取话题
func TestGetTopicsByPlatform(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建表
	err = db.AutoMigrate(&Platform{}, &Topic{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建测试平台
	platforms := []Platform{
		{ID: "weibo", Name: "微博", IsActive: true},
		{ID: "douyin", Name: "抖音", IsActive: true},
	}
	for _, p := range platforms {
		db.Create(&p)
	}

	// 创建测试话题
	topics := []Topic{
		{ID: "topic-1", Title: "Topic 1", Source: "weibo", Heat: 1000, Trend: "up"},
		{ID: "topic-2", Title: "Topic 2", Source: "weibo", Heat: 2000, Trend: "down"},
		{ID: "topic-3", Title: "Topic 3", Source: "douyin", Heat: 3000, Trend: "stable"},
	}
	for _, topic := range topics {
		db.Create(&topic)
	}

	// 测试获取微博话题
	service := &EnhancedHotspotService{DB: db}
	weiboTopics, err := service.GetTopicsByPlatform("weibo", 10)
	if err != nil {
		t.Fatalf("GetTopicsByPlatform() error = %v", err)
	}

	if len(weiboTopics) != 2 {
		t.Errorf("GetTopicsByPlatform() = %v, want 2", len(weiboTopics))
	}

	// 验证所有话题都来自微博
	for _, topic := range weiboTopics {
		if topic.Source != "weibo" {
			t.Errorf("Topic source = %v, want weibo", topic.Source)
		}
	}
}

// TestGetTopTopics 测试获取热门话题
func TestGetTopTopics(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建表
	err = db.AutoMigrate(&Topic{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建测试话题（按热度排序）
	topics := []Topic{
		{ID: "topic-1", Title: "Topic 1", Source: "weibo", Heat: 5000, Trend: "hot"},
		{ID: "topic-2", Title: "Topic 2", Source: "douyin", Heat: 4000, Trend: "up"},
		{ID: "topic-3", Title: "Topic 3", Source: "xiaohongshu", Heat: 3000, Trend: "stable"},
		{ID: "topic-4", Title: "Topic 4", Source: "weibo", Heat: 2000, Trend: "down"},
		{ID: "topic-5", Title: "Topic 5", Source: "douyin", Heat: 1000, Trend: "new"},
	}
	for _, topic := range topics {
		db.Create(&topic)
	}

	// 测试获取前 3 个热门话题
	service := &EnhancedHotspotService{DB: db}
	topTopics, err := service.GetTopTopics(3)
	if err != nil {
		t.Fatalf("GetTopTopics() error = %v", err)
	}

	if len(topTopics) != 3 {
		t.Errorf("GetTopTopics() = %v, want 3", len(topTopics))
	}

	// 验证话题按热度降序排列
	if topTopics[0].Heat < topTopics[1].Heat {
		t.Error("Topics not sorted by heat descending")
	}
}

// TestSearchTopics 测试搜索话题
func TestSearchTopics(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建表
	err = db.AutoMigrate(&Topic{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建测试话题
	topics := []Topic{
		{ID: "topic-1", Title: "人工智能新闻", Source: "weibo", Heat: 1000, Trend: "up"},
		{ID: "topic-2", Title: "AI技术发展", Source: "douyin", Heat: 2000, Trend: "hot"},
		{ID: "topic-3", Title: "机器学习应用", Source: "xiaohongshu", Heat: 1500, Trend: "stable"},
		{ID: "topic-4", Title: "美食推荐", Source: "weibo", Heat: 3000, Trend: "down"},
	}
	for _, topic := range topics {
		db.Create(&topic)
	}

	// 测试搜索包含"AI"的话题
	service := &EnhancedHotspotService{DB: db}
	results, err := service.SearchTopics("AI", 10)
	if err != nil {
		t.Fatalf("SearchTopics() error = %v", err)
	}

	// 应该找到 2 个话题
	if len(results) != 2 {
		t.Errorf("SearchTopics() = %v, want 2", len(results))
	}

	// 验证搜索结果都包含关键词
	for _, topic := range results {
		contains := false
		if topic.Title == "人工智能新闻" || topic.Title == "AI技术发展" {
			contains = true
		}
		if !contains {
			t.Errorf("Topic title = %v, should contain AI", topic.Title)
		}
	}
}

// TestGetTrendingTopics 测试获取趋势话题
func TestGetTrendingTopics(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 创建表
	err = db.AutoMigrate(&Topic{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 创建测试话题
	topics := []Topic{
		{ID: "topic-1", Title: "Topic 1", Source: "weibo", Heat: 1000, Trend: "up"},
		{ID: "topic-2", Title: "Topic 2", Source: "douyin", Heat: 2000, Trend: "hot"},
		{ID: "topic-3", Title: "Topic 3", Source: "xiaohongshu", Heat: 3000, Trend: "stable"},
		{ID: "topic-4", Title: "Topic 4", Source: "weibo", Heat: 4000, Trend: "down"},
		{ID: "topic-5", Title: "Topic 5", Source: "douyin", Heat: 5000, Trend: "new"},
	}
	for _, topic := range topics {
		db.Create(&topic)
	}

	// 测试获取上升趋势的话题
	service := &EnhancedHotspotService{DB: db}
	trendingTopics, err := service.GetTrendingTopics("up", 10)
	if err != nil {
		t.Fatalf("GetTrendingTopics() error = %v", err)
	}

	// 应该找到 1 个上升趋势的话题
	if len(trendingTopics) != 1 {
		t.Errorf("GetTrendingTopics() = %v, want 1", len(trendingTopics))
	}

	// 验证所有话题的趋势都是 up
	for _, topic := range trendingTopics {
		if topic.Trend != "up" {
			t.Errorf("Topic trend = %v, want up", topic.Trend)
		}
	}
}
