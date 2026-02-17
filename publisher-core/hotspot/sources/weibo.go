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

	topics := []hotspot.Topic{
		{
			ID:        "weibo-1",
			Title:     "Weibo Hot Topic 1",
			Source:    "weibo",
			Heat:      5000000,
			SourceURL: "https://s.weibo.com/top/summary",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "weibo-2",
			Title:     "Weibo Hot Topic 2",
			Source:    "weibo",
			Heat:      4500000,
			SourceURL: "https://s.weibo.com/top/summary",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Infof("[Weibo] Fetched %d topics", len(topics))
	return topics, nil
}
