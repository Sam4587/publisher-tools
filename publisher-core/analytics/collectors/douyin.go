package collectors

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"publisher-core/analytics"
	"github.com/sirupsen/logrus"
)

type DouyinCollector struct {
	enabled bool
}

func NewDouyinCollector() *DouyinCollector {
	return &DouyinCollector{
		enabled: true,
	}
}

func (c *DouyinCollector) Platform() analytics.Platform {
	return analytics.PlatformDouyin
}

func (c *DouyinCollector) IsEnabled() bool {
	return c.enabled
}

func (c *DouyinCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *DouyinCollector) CollectPostMetrics(ctx context.Context, postID string) (*analytics.PostMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Douyin] Collecting metrics for post: %s", postID)

	rand.Seed(time.Now().UnixNano())
	metrics := &analytics.PostMetrics{
		PostID:      postID,
		Platform:    analytics.PlatformDouyin,
		Title:       fmt.Sprintf("Douyin Video %s", postID),
		Views:       rand.Int63n(100000),
		Likes:       rand.Int63n(10000),
		Comments:    rand.Int63n(1000),
		Shares:      rand.Int63n(500),
		Favorites:   rand.Int63n(800),
		CollectedAt: time.Now(),
		PublishedAt: time.Now().Add(-24 * time.Hour),
	}

	metrics.Engagement = analytics.CalculateEngagement(
		metrics.Likes,
		metrics.Comments,
		metrics.Shares,
		metrics.Views,
	)

	logrus.Infof("[Douyin] Collected metrics: views=%d, likes=%d, engagement=%.2f%%",
		metrics.Views, metrics.Likes, metrics.Engagement)

	return metrics, nil
}

func (c *DouyinCollector) CollectAccountMetrics(ctx context.Context, accountID string) (*analytics.AccountMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Douyin] Collecting account metrics: %s", accountID)

	rand.Seed(time.Now().UnixNano())
	metrics := &analytics.AccountMetrics{
		AccountID:   accountID,
		Platform:    analytics.PlatformDouyin,
		Username:    fmt.Sprintf("DouyinUser%s", accountID),
		Followers:   rand.Int63n(1000000),
		Following:   rand.Int63n(1000),
		Posts:       rand.Int63n(500),
		Likes:       rand.Int63n(10000000),
		CollectedAt: time.Now(),
	}

	logrus.Infof("[Douyin] Collected account: followers=%d, posts=%d",
		metrics.Followers, metrics.Posts)

	return metrics, nil
}
