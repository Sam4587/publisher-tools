package collectors

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"publisher-core/analytics"
	"github.com/sirupsen/logrus"
)

type XiaohongshuCollector struct {
	enabled bool
}

func NewXiaohongshuCollector() *XiaohongshuCollector {
	return &XiaohongshuCollector{
		enabled: true,
	}
}

func (c *XiaohongshuCollector) Platform() analytics.Platform {
	return analytics.PlatformXiaohongshu
}

func (c *XiaohongshuCollector) IsEnabled() bool {
	return c.enabled
}

func (c *XiaohongshuCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *XiaohongshuCollector) CollectPostMetrics(ctx context.Context, postID string) (*analytics.PostMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Xiaohongshu] Collecting metrics for post: %s", postID)

	rand.Seed(time.Now().UnixNano())
	metrics := &analytics.PostMetrics{
		PostID:      postID,
		Platform:    analytics.PlatformXiaohongshu,
		Title:       fmt.Sprintf("Xiaohongshu Note %s", postID),
		Views:       rand.Int63n(50000),
		Likes:       rand.Int63n(5000),
		Comments:    rand.Int63n(500),
		Shares:      rand.Int63n(200),
		Favorites:   rand.Int63n(1000),
		CollectedAt: time.Now(),
		PublishedAt: time.Now().Add(-24 * time.Hour),
	}

	metrics.Engagement = analytics.CalculateEngagement(
		metrics.Likes,
		metrics.Comments,
		metrics.Shares,
		metrics.Views,
	)

	logrus.Infof("[Xiaohongshu] Collected metrics: views=%d, likes=%d, engagement=%.2f%%",
		metrics.Views, metrics.Likes, metrics.Engagement)

	return metrics, nil
}

func (c *XiaohongshuCollector) CollectAccountMetrics(ctx context.Context, accountID string) (*analytics.AccountMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Xiaohongshu] Collecting account metrics: %s", accountID)

	rand.Seed(time.Now().UnixNano())
	metrics := &analytics.AccountMetrics{
		AccountID:   accountID,
		Platform:    analytics.PlatformXiaohongshu,
		Username:    fmt.Sprintf("XiaohongshuUser%s", accountID),
		Followers:   rand.Int63n(500000),
		Following:   rand.Int63n(500),
		Posts:       rand.Int63n(300),
		Likes:       rand.Int63n(5000000),
		CollectedAt: time.Now(),
	}

	logrus.Infof("[Xiaohongshu] Collected account metrics")
	return metrics, nil
}
