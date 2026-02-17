package collectors

import (
	"context"
	"fmt"
	"time"

	"publisher-core/analytics"
	"publisher-core/browser"
	"publisher-core/cookies"

	"github.com/sirupsen/logrus"
)

type RealDouyinCollector struct {
	enabled   bool
	browser   *browser.Browser
	cookieMgr *cookies.Manager
}

func NewRealDouyinCollector(cookieMgr *cookies.Manager) *RealDouyinCollector {
	return &RealDouyinCollector{
		enabled:   true,
		cookieMgr: cookieMgr,
	}
}

func (c *RealDouyinCollector) Platform() analytics.Platform {
	return analytics.PlatformDouyin
}

func (c *RealDouyinCollector) IsEnabled() bool {
	return c.enabled
}

func (c *RealDouyinCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

func (c *RealDouyinCollector) initBrowser() error {
	if c.browser == nil {
		c.browser = browser.NewBrowser(&browser.Config{
			Headless: true,
		})
	}
	return nil
}

func (c *RealDouyinCollector) CollectPostMetrics(ctx context.Context, postID string) (*analytics.PostMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[RealDouyin] Collecting metrics for post: %s", postID)

	if err := c.initBrowser(); err != nil {
		return nil, err
	}

	metrics := &analytics.PostMetrics{
		PostID:      postID,
		Platform:    analytics.PlatformDouyin,
		Title:       fmt.Sprintf("Douyin Video %s", postID),
		Views:       0,
		Likes:       0,
		Comments:    0,
		Shares:      0,
		Favorites:   0,
		CollectedAt: time.Now(),
		PublishedAt: time.Now().Add(-24 * time.Hour),
	}

	logrus.Infof("[RealDouyin] Collected metrics (placeholder)")
	return metrics, nil
}

func (c *RealDouyinCollector) CollectAccountMetrics(ctx context.Context, accountID string) (*analytics.AccountMetrics, error) {
	if !c.enabled {
		return nil, fmt.Errorf("collector is disabled")
	}

	logrus.Infof("[RealDouyin] Collecting account metrics: %s", accountID)

	metrics := &analytics.AccountMetrics{
		AccountID:   accountID,
		Platform:    analytics.PlatformDouyin,
		Username:    fmt.Sprintf("DouyinUser%s", accountID),
		Followers:   0,
		Following:   0,
		Posts:       0,
		Likes:       0,
		CollectedAt: time.Now(),
	}

	logrus.Infof("[RealDouyin] Collected account metrics (placeholder)")
	return metrics, nil
}
