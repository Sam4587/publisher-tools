package sources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"publisher-core/hotspot"
	"github.com/sirupsen/logrus"
)

type WeiboSource struct {
	name        string
	displayName string
	enabled     bool
	client      *http.Client
}

func NewWeiboSource() *WeiboSource {
	return &WeiboSource{
		name:        "weibo",
		displayName: "Weibo",
		enabled:     true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *WeiboSource) Name() string {
	return s.name
}

func (s *WeiboSource) DisplayName() string {
	return s.displayName
}

func (s *WeiboSource) ID() string {
	return s.name
}

func (s *WeiboSource) IsEnabled() bool {
	return s.enabled
}

func (s *WeiboSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *WeiboSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Weibo] Fetching hot topics, maxItems=%d", maxItems)

	// 尝试抓取真实数据
	topics, err := s.fetchRealData(ctx, maxItems)
	if err != nil {
		logrus.Warnf("[Weibo] Failed to fetch real data, using fallback: %v", err)
		// 失败时使用模拟数据
		topics = s.getFallbackTopics(maxItems)
	}

	if len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Infof("[Weibo] Fetched %d topics", len(topics))
	return topics, nil
}

// fetchRealData 尝试抓取真实数据
func (s *WeiboSource) fetchRealData(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://s.weibo.com/top/summary", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// TODO: 解析HTML页面提取热点数据
	// 由于微博的反爬虫机制,这里返回空,使用fallback数据
	return nil, fmt.Errorf("parsing not implemented")
}

// getFallbackTopics 获取备用数据
func (s *WeiboSource) getFallbackTopics(maxItems int) []hotspot.Topic {
	now := time.Now()
	return []hotspot.Topic{
		{
			ID:        "weibo-1",
			Title:     "微博热门话题示例 1",
			Source:    "weibo",
			Heat:      5000000,
			SourceURL: "https://s.weibo.com/top/summary",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "weibo-2",
			Title:     "微博热门话题示例 2",
			Source:    "weibo",
			Heat:      4500000,
			SourceURL: "https://s.weibo.com/top/summary",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
