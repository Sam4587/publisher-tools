package analytics

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type AnalyticsService struct {
	mu         sync.RWMutex
	collectors map[Platform]Collector
	storage    MetricsStorage
}

func NewService(storage MetricsStorage) *AnalyticsService {
	return &AnalyticsService{
		collectors: make(map[Platform]Collector),
		storage:    storage,
	}
}

func (s *AnalyticsService) RegisterCollector(collector Collector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.collectors[collector.Platform()] = collector
	logrus.Infof("Registered analytics collector: %s", collector.Platform())
}

func (s *AnalyticsService) GetCollector(platform Platform) (Collector, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.collectors[platform]
	return c, ok
}

func (s *AnalyticsService) ListCollectors() []Platform {
	s.mu.RLock()
	defer s.mu.RUnlock()

	platforms := make([]Platform, 0, len(s.collectors))
	for p := range s.collectors {
		platforms = append(platforms, p)
	}
	return platforms
}

func (s *AnalyticsService) CollectPostMetrics(ctx context.Context, platform Platform, postID string) (*PostMetrics, error) {
	collector, ok := s.GetCollector(platform)
	if !ok {
		return nil, ErrCollectorNotFound
	}

	if !collector.IsEnabled() {
		return nil, ErrCollectorDisabled
	}

	metrics, err := collector.CollectPostMetrics(ctx, postID)
	if err != nil {
		return nil, err
	}

	if s.storage != nil {
		if err := s.storage.SavePostMetrics(metrics); err != nil {
			logrus.Warnf("Failed to save post metrics: %v", err)
		}
	}

	return metrics, nil
}

func (s *AnalyticsService) CollectAccountMetrics(ctx context.Context, platform Platform, accountID string) (*AccountMetrics, error) {
	collector, ok := s.GetCollector(platform)
	if !ok {
		return nil, ErrCollectorNotFound
	}

	if !collector.IsEnabled() {
		return nil, ErrCollectorDisabled
	}

	metrics, err := collector.CollectAccountMetrics(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if s.storage != nil {
		if err := s.storage.SaveAccountMetrics(metrics); err != nil {
			logrus.Warnf("Failed to save account metrics: %v", err)
		}
	}

	return metrics, nil
}

func (s *AnalyticsService) GetDashboardStats() (*DashboardStats, error) {
	if s.storage == nil {
		return nil, ErrStorageNotInitialized
	}

	stats := &DashboardStats{
		PlatformStats: []PlatformStats{},
		RecentPosts:   []*PostMetrics{},
		TopPerforming: []*PostMetrics{},
		GrowthTrend:   []TrendData{},
	}

	for _, platform := range s.ListCollectors() {
		posts, err := s.storage.ListPostMetrics(platform, 100)
		if err != nil {
			logrus.Warnf("Failed to get posts for %s: %v", platform, err)
			continue
		}

		ps := PlatformStats{Platform: platform}
		for _, p := range posts {
			ps.Posts++
			ps.Views += p.Views
			ps.Likes += p.Likes
			ps.Comments += p.Comments

			stats.TotalPosts++
			stats.TotalViews += p.Views
			stats.TotalLikes += p.Likes
			stats.TotalComments += p.Comments
			stats.TotalShares += p.Shares
		}

		if ps.Posts > 0 {
			ps.Engagement = CalculateEngagement(ps.Likes, ps.Comments, 0, ps.Views)
			stats.PlatformStats = append(stats.PlatformStats, ps)
		}

		if len(posts) > 0 {
			stats.RecentPosts = append(stats.RecentPosts, posts[:min(5, len(posts))]...)
		}
	}

	if stats.TotalViews > 0 {
		stats.AvgEngagement = CalculateEngagement(
			stats.TotalLikes,
			stats.TotalComments,
			stats.TotalShares,
			stats.TotalViews,
		)
	}

	trend, err := s.storage.GetTrendData(MetricTypeViews, "", 7)
	if err == nil {
		stats.GrowthTrend = trend
	}

	return stats, nil
}

func (s *AnalyticsService) GetTrendData(metricType MetricType, platform Platform, days int) ([]TrendData, error) {
	if s.storage == nil {
		return nil, ErrStorageNotInitialized
	}
	return s.storage.GetTrendData(metricType, platform, days)
}

func (s *AnalyticsService) RefreshMetrics(ctx context.Context) error {
	logrus.Info("Starting metrics refresh...")

	// TODO: Implement actual refresh logic
	// This would iterate through known posts and accounts and collect fresh metrics

	return nil
}

func (s *AnalyticsService) GetDailyStats(platform Platform, days int) ([]*DailyStats, error) {
	if s.storage == nil {
		return nil, ErrStorageNotInitialized
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)
	return s.storage.GetDailyStats(platform, start, end)
}

var (
	ErrCollectorNotFound     = &Error{Code: "COLLECTOR_NOT_FOUND", Message: "collector not found"}
	ErrCollectorDisabled     = &Error{Code: "COLLECTOR_DISABLED", Message: "collector is disabled"}
	ErrStorageNotInitialized = &Error{Code: "STORAGE_NOT_INITIALIZED", Message: "storage not initialized"}
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
