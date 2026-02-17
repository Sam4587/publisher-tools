package analytics

import (
	"context"
	"time"
)

type MetricType string

const (
	MetricTypeViews      MetricType = "views"
	MetricTypeLikes      MetricType = "likes"
	MetricTypeComments   MetricType = "comments"
	MetricTypeShares     MetricType = "shares"
	MetricTypeFavorites  MetricType = "favorites"
	MetricTypeFollowers  MetricType = "followers"
	MetricTypeEngagement MetricType = "engagement"
)

type Platform string

const (
	PlatformDouyin      Platform = "douyin"
	PlatformToutiao     Platform = "toutiao"
	PlatformXiaohongshu Platform = "xiaohongshu"
	PlatformWeibo       Platform = "weibo"
)

type PostMetrics struct {
	PostID      string    `json:"post_id"`
	Platform    Platform  `json:"platform"`
	Title       string    `json:"title"`
	Views       int64     `json:"views"`
	Likes       int64     `json:"likes"`
	Comments    int64     `json:"comments"`
	Shares      int64     `json:"shares"`
	Favorites   int64     `json:"favorites"`
	Engagement  float64   `json:"engagement"`
	CollectedAt time.Time `json:"collected_at"`
	PublishedAt time.Time `json:"published_at"`
}

type AccountMetrics struct {
	AccountID   string    `json:"account_id"`
	Platform    Platform  `json:"platform"`
	Username    string    `json:"username"`
	Followers   int64     `json:"followers"`
	Following   int64     `json:"following"`
	Posts       int64     `json:"posts"`
	Likes       int64     `json:"likes"`
	CollectedAt time.Time `json:"collected_at"`
}

type DailyStats struct {
	Date          time.Time `json:"date"`
	Platform      Platform  `json:"platform"`
	PostsCount    int       `json:"posts_count"`
	TotalViews    int64     `json:"total_views"`
	TotalLikes    int64     `json:"total_likes"`
	TotalComments int64     `json:"total_comments"`
	TotalShares   int64     `json:"total_shares"`
	NewFollowers  int64     `json:"new_followers"`
}

type TrendData struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type Collector interface {
	Platform() Platform
	CollectPostMetrics(ctx context.Context, postID string) (*PostMetrics, error)
	CollectAccountMetrics(ctx context.Context, accountID string) (*AccountMetrics, error)
	IsEnabled() bool
	SetEnabled(enabled bool)
}

type MetricsStorage interface {
	SavePostMetrics(metrics *PostMetrics) error
	GetPostMetrics(postID string) (*PostMetrics, error)
	ListPostMetrics(platform Platform, limit int) ([]*PostMetrics, error)

	SaveAccountMetrics(metrics *AccountMetrics) error
	GetAccountMetrics(accountID string) (*AccountMetrics, error)

	SaveDailyStats(stats *DailyStats) error
	GetDailyStats(platform Platform, startDate, endDate time.Time) ([]*DailyStats, error)

	GetTrendData(metricType MetricType, platform Platform, days int) ([]TrendData, error)
}

type Service interface {
	CollectPostMetrics(ctx context.Context, platform Platform, postID string) (*PostMetrics, error)
	CollectAccountMetrics(ctx context.Context, platform Platform, accountID string) (*AccountMetrics, error)
	GetDashboardStats() (*DashboardStats, error)
	GetTrendData(metricType MetricType, platform Platform, days int) ([]TrendData, error)
	RefreshMetrics(ctx context.Context) error
}

type DashboardStats struct {
	TotalPosts    int64           `json:"total_posts"`
	TotalViews    int64           `json:"total_views"`
	TotalLikes    int64           `json:"total_likes"`
	TotalComments int64           `json:"total_comments"`
	TotalShares   int64           `json:"total_shares"`
	AvgEngagement float64         `json:"avg_engagement"`
	PlatformStats []PlatformStats `json:"platform_stats"`
	RecentPosts   []*PostMetrics  `json:"recent_posts"`
	TopPerforming []*PostMetrics  `json:"top_performing"`
	GrowthTrend   []TrendData     `json:"growth_trend"`
}

type PlatformStats struct {
	Platform   Platform `json:"platform"`
	Posts      int64    `json:"posts"`
	Views      int64    `json:"views"`
	Likes      int64    `json:"likes"`
	Comments   int64    `json:"comments"`
	Engagement float64  `json:"engagement"`
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func CalculateEngagement(likes, comments, shares, views int64) float64 {
	if views == 0 {
		return 0
	}
	return float64(likes+comments+shares) / float64(views) * 100
}
