package sources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"publisher-core/hotspot"

	"github.com/sirupsen/logrus"
)

type ZhihuSource struct {
	name        string
	displayName string
	enabled     bool
	client      *http.Client
}

func NewZhihuSource() *ZhihuSource {
	return &ZhihuSource{
		name:        "zhihu",
		displayName: "Zhihu",
		enabled:     true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *ZhihuSource) Name() string {
	return s.name
}

func (s *ZhihuSource) DisplayName() string {
	return s.displayName
}

func (s *ZhihuSource) ID() string {
	return s.name
}

func (s *ZhihuSource) IsEnabled() bool {
	return s.enabled
}

func (s *ZhihuSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *ZhihuSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Zhihu] Fetching hot topics, maxItems=%d", maxItems)

	// 尝试抓取真实数据
	topics, err := s.fetchRealData(ctx, maxItems)
	if err != nil {
		logrus.Warnf("[Zhihu] Failed to fetch real data, using fallback: %v", err)
		// 失败时使用模拟数据
		topics = s.getFallbackTopics(maxItems)
	}

	if len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Infof("[Zhihu] Fetched %d topics", len(topics))
	return topics, nil
}

// fetchRealData 尝试抓取真实数据
func (s *ZhihuSource) fetchRealData(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// TODO: 解析JSON响应提取热点数据
	// 由于知乎的反爬虫机制,这里返回空,使用fallback数据
	return nil, fmt.Errorf("parsing not implemented")
}

// getFallbackTopics 获取备用数据
func (s *ZhihuSource) getFallbackTopics(maxItems int) []hotspot.Topic {
	now := time.Now()
	return []hotspot.Topic{
		{
			ID:        "zhihu-1",
			Title:     "知乎热门问题示例 1",
			Source:    "zhihu",
			Heat:      8000000,
			SourceURL: "https://www.zhihu.com/hot",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "zhihu-2",
			Title:     "知乎热门问题示例 2",
			Source:    "zhihu",
			Heat:      7500000,
			SourceURL: "https://www.zhihu.com/hot",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
