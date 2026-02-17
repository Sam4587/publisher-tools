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

	topics := []hotspot.Topic{
		{
			ID:        "zhihu-1",
			Title:     "Zhihu Hot Question 1",
			Source:    "zhihu",
			Heat:      8000000,
			SourceURL: "https://www.zhihu.com/hot",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "zhihu-2",
			Title:     "Zhihu Hot Question 2",
			Source:    "zhihu",
			Heat:      7500000,
			SourceURL: "https://www.zhihu.com/hot",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	logrus.Infof("[Zhihu] Fetched %d topics", len(topics))
	return topics, nil
}
