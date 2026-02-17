package sources

import (
	"context"
	"fmt"
	"time"

	"publisher-core/hotspot"

	"github.com/sirupsen/logrus"
)

type BaiduSource struct {
	enabled bool
}

func NewBaiduSource() *BaiduSource {
	return &BaiduSource{
		enabled: true,
	}
}

func (s *BaiduSource) Name() string {
	return "baidu"
}

func (s *BaiduSource) IsEnabled() bool {
	return s.enabled
}

func (s *BaiduSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *BaiduSource) Fetch(ctx context.Context, limit int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, fmt.Errorf("source is disabled")
	}

	logrus.Infof("[Baidu] Fetching hot topics, limit=%d", limit)

	topics := []hotspot.Topic{
		{
			ID:        "baidu-1",
			Title:     "Baidu Hot Topic 1",
			Source:    "baidu",
			Heat:      1000000,
			SourceURL: "https://top.baidu.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "baidu-2",
			Title:     "Baidu Hot Topic 2",
			Source:    "baidu",
			Heat:      900000,
			SourceURL: "https://top.baidu.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	logrus.Infof("[Baidu] Fetched %d topics", len(topics))
	return topics, nil
}
