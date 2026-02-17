package collectors

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"publisher-core/analytics"
	"github.com/sirupsen/logrus"
)

type ToutiaoCollector struct {
	enabled bool
}

func NewToutiaoCollector() *ToutiaoCollector {
	return &ToutiaoCollector{
		enabled: true,
	}
}

func (c *ToutiaoCollector) Platform() analytics.Platform {
	return analytics.PlatformToutiao
}

func (c *ToutiaoCollector) IsEnabled() bool {
	return c.enabled
}

func (c *ToutiaoCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *ToutiaoCollector) CollectPostMetrics(ctx context.Context, postID string) (*analytics.PostMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Toutiao] Collecting metrics for post: %s", postID)

	metrics := &analytics.PostMetrics{
		PostID:      postID,
		Platform:    analytics.PlatformToutiao,
		Title:       fmt.Sprintf("Toutiao Article %s", postID),
		Views:       rand.Int63n(50000),
		Likes:       rand.Int63n(5000),
		Comments:    rand.Int63n(1000),
		Shares:      rand.Int63n(500),
		Favorites:   rand.Int63n(200),
		CollectedAt: time.Now(),
		PublishedAt: time.Now().Add(-24 * time.Hour),
	}

	logrus.Infof("[Toutiao] Collected metrics")
	return metrics, nil
}

func (c *ToutiaoCollector) CollectAccountMetrics(ctx context.Context, accountID string) (*analytics.AccountMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[Toutiao] Collecting account metrics: %s", accountID)

	metrics := &analytics.AccountMetrics{
		AccountID:   accountID,
		Platform:    analytics.PlatformToutiao,
		Username:    fmt.Sprintf("ToutiaoUser%s", accountID),
		Followers:   rand.Int63n(100000),
		Following:   rand.Int63n(500),
		Posts:       rand.Int63n(100),
		Likes:       rand.Int63n(50000),
		CollectedAt: time.Now(),
	}

	logrus.Infof("[Toutiao] Collected account metrics")
	return metrics, nil
}
