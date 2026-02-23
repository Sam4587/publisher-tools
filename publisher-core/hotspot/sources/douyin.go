package sources

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"publisher-core/hotspot"

	"github.com/sirupsen/logrus"
)

type DouyinSource struct {
	name        string
	displayName string
	enabled     bool
	client      *http.Client
}

func NewDouyinSource() *DouyinSource {
	return &DouyinSource{
		name:        "douyin",
		displayName: "Douyin",
		enabled:     true,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *DouyinSource) Name() string {
	return s.name
}

func (s *DouyinSource) DisplayName() string {
	return s.displayName
}

func (s *DouyinSource) ID() string {
	return s.name
}

func (s *DouyinSource) IsEnabled() bool {
	return s.enabled
}

func (s *DouyinSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *DouyinSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Douyin] Fetching hot topics, maxItems=%d", maxItems)

	// TODO: 实现真实的抖音热点数据抓取
	// 当前返回模拟数据
	topics := []hotspot.Topic{
		{
			ID:        "douyin-1",
			Title:     "抖音热门视频 1",
			Source:    "douyin",
			Heat:      10000000,
			SourceURL: "https://www.douyin.com/hot",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "douyin-2",
			Title:     "抖音热门视频 2",
			Source:    "douyin",
			Heat:      9500000,
			SourceURL: "https://www.douyin.com/hot",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Infof("[Douyin] Fetched %d topics", len(topics))
	return topics, nil
}
