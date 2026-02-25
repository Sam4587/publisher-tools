package sources

import (
	"context"
	"time"

	"publisher-core/hotspot"
)

type MockSource struct {
	id      string
	name    string
	enabled bool
}

func NewMockSource(id, name string) *MockSource {
	return &MockSource{
		id:      id,
		name:    name,
		enabled: true,
	}
}

func (s *MockSource) ID() string {
	return s.id
}

func (s *MockSource) Name() string {
	return s.name
}

func (s *MockSource) IsEnabled() bool {
	return s.enabled
}

func (s *MockSource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

func (s *MockSource) Fetch(ctx context.Context, maxItems int) ([]hotspot.Topic, error) {
	if !s.enabled {
		return nil, nil
	}

	now := time.Now()
	// 使用固定的 ID，避免重复数据
	topics := []hotspot.Topic{
		{
			ID:        s.id + "-1",
			Title:     "Test Hotspot 1: This is a mock hotspot topic",
			Category:  hotspot.CategoryTech,
			Heat:      95,
			Trend:     hotspot.TrendHot,
			Source:    s.id,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        s.id + "-2",
			Title:     "Test Hotspot 2: AI Technology Development Trend Analysis",
			Category:  hotspot.CategoryTech,
			Heat:      88,
			Trend:     hotspot.TrendUp,
			Source:    s.id,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        s.id + "-3",
			Title:     "Test Hotspot 3: Technology Innovation Drives the Future",
			Category:  hotspot.CategoryTech,
			Heat:      75,
			Trend:     hotspot.TrendNew,
			Source:    s.id,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	if maxItems > 0 && len(topics) > maxItems {
		topics = topics[:maxItems]
	}

	return topics, nil
}
