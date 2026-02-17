package sources

import (
	"net/http"
	"context"
	"fmt"
	"time"

	"publisher-core/hotspot"
	"github.com/sirupsen/logrus"
)

// WeiboSource 微博热搜数据�?
type WeiboSource struct {
	name        string
	displayName string
	enabled     bool
	client      *http.Client
}

// NewWeiboSource 创建微博数据�?
func NewWeiboSource() *WeiboSource {
	return &WeiboSource{
		name:        "weibo",
		displayName: "微博热搜",
		enabled:     true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回数据源名�?
func (s *WeiboSource) Name() string {
	return s.name
}

// DisplayName 返回显示名称
func (s *WeiboSource) DisplayName() string {
	return s.displayName
}

// IsEnabled 检查是否启�?
func (s *WeiboSource) IsEnabled() bool {
	return s.enabled
}

// SetEnabled 设置启用状�?
func (s *WeiboSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// Fetch 抓取数据
func (s *WeiboSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Weibo] Fetching hot topics, maxItems=%d", maxItems)

	// TODO: 实现真实的微博热搜抓�?
	// 可以使用以下方式�?
	// 1. 微博API（需要申请）
	// 2. 网页抓取
	// 3. 第三方聚合API
	
	// 当前返回模拟数据
	topics := s.generateMockTopics(maxItems)
	
	logrus.Infof("[Weibo] Fetched %d topics", len(topics))
	return topics, nil
}

// generateMockTopics 生成模拟数据
func (s *WeiboSource) generateMockTopics(count int) []hotspot.Topic {
	var topics []hotspot.Topic
	
	mockData := []struct {
		title string
		heat  int
		url   string
	}{
		{"微博热搜榜更�?, 999999, "https://s.weibo.com/weibo?q=热搜"},
		{"明星动态新�?, 888888, "https://s.weibo.com/weibo?q=明星"},
		{"科技前沿资讯", 777777, "https://s.weibo.com/weibo?q=科技"},
		{"社会热点事件", 666666, "https://s.weibo.com/weibo?q=社会"},
		{"体育赛事报道", 555555, "https://s.weibo.com/weibo?q=体育"},
	}

	for i := 0; i < count && i < len(mockData); i++ {
		topics = append(topics, hotspot.Topic{
			ID:          fmt.Sprintf("weibo_%d", time.Now().UnixNano()+int64(i)),
			Title:       mockData[i].title,
			Description: fmt.Sprintf("微博热搜话题�?s", mockData[i].title),
			Category:    hotspot.CategoryEntertainment,
			Heat:        mockData[i].heat,
			Trend:       "up",
			Source:      "weibo",
			SourceURL:   mockData[i].url,
			Keywords:    []string{"微博", "热搜", "热点"},
			PublishedAt: time.Now(),
			CreatedAt:   time.Now(),
		})
	}

	return topics
}
