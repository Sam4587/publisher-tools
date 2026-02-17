package analytics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type JSONStorage struct {
	mu             sync.RWMutex
	dataDir        string
	postMetrics    map[string]*PostMetrics
	accountMetrics map[string]*AccountMetrics
	dailyStats     map[string]*DailyStats
}

func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if dataDir == "" {
		dataDir = "./data/analytics"
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	s := &JSONStorage{
		dataDir:        dataDir,
		postMetrics:    make(map[string]*PostMetrics),
		accountMetrics: make(map[string]*AccountMetrics),
		dailyStats:     make(map[string]*DailyStats),
	}

	if err := s.load(); err != nil {
		// Ignore load errors, start fresh
	}

	return s, nil
}

func (s *JSONStorage) load() error {
	postFile := filepath.Join(s.dataDir, "post_metrics.json")
	if data, err := os.ReadFile(postFile); err == nil {
		var posts []*PostMetrics
		if err := json.Unmarshal(data, &posts); err == nil {
			for _, p := range posts {
				s.postMetrics[p.PostID] = p
			}
		}
	}

	accountFile := filepath.Join(s.dataDir, "account_metrics.json")
	if data, err := os.ReadFile(accountFile); err == nil {
		var accounts []*AccountMetrics
		if err := json.Unmarshal(data, &accounts); err == nil {
			for _, a := range accounts {
				s.accountMetrics[a.AccountID] = a
			}
		}
	}

	return nil
}

func (s *JSONStorage) save() error {
	posts := make([]*PostMetrics, 0, len(s.postMetrics))
	for _, p := range s.postMetrics {
		posts = append(posts, p)
	}
	if data, err := json.MarshalIndent(posts, "", "  "); err == nil {
		os.WriteFile(filepath.Join(s.dataDir, "post_metrics.json"), data, 0644)
	}

	accounts := make([]*AccountMetrics, 0, len(s.accountMetrics))
	for _, a := range s.accountMetrics {
		accounts = append(accounts, a)
	}
	if data, err := json.MarshalIndent(accounts, "", "  "); err == nil {
		os.WriteFile(filepath.Join(s.dataDir, "account_metrics.json"), data, 0644)
	}

	return nil
}

func (s *JSONStorage) SavePostMetrics(metrics *PostMetrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics.CollectedAt = time.Now()
	s.postMetrics[metrics.PostID] = metrics
	return s.save()
}

func (s *JSONStorage) GetPostMetrics(postID string) (*PostMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.postMetrics[postID], nil
}

func (s *JSONStorage) ListPostMetrics(platform Platform, limit int) ([]*PostMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*PostMetrics
	for _, p := range s.postMetrics {
		if platform == "" || p.Platform == platform {
			result = append(result, p)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CollectedAt.After(result[j].CollectedAt)
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (s *JSONStorage) SaveAccountMetrics(metrics *AccountMetrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics.CollectedAt = time.Now()
	s.accountMetrics[metrics.AccountID] = metrics
	return s.save()
}

func (s *JSONStorage) GetAccountMetrics(accountID string) (*AccountMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accountMetrics[accountID], nil
}

func (s *JSONStorage) SaveDailyStats(stats *DailyStats) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := stats.Date.Format("2006-01-02") + "_" + string(stats.Platform)
	s.dailyStats[key] = stats
	return nil
}

func (s *JSONStorage) GetDailyStats(platform Platform, startDate, endDate time.Time) ([]*DailyStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*DailyStats
	for _, stats := range s.dailyStats {
		if platform != "" && stats.Platform != platform {
			continue
		}
		if stats.Date.Before(startDate) || stats.Date.After(endDate) {
			continue
		}
		result = append(result, stats)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date.Before(result[j].Date)
	})

	return result, nil
}

func (s *JSONStorage) GetTrendData(metricType MetricType, platform Platform, days int) ([]TrendData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	stats, err := s.GetDailyStats(platform, startDate, now)
	if err != nil {
		return nil, err
	}

	result := make([]TrendData, 0, len(stats))
	for _, stat := range stats {
		var value float64
		switch metricType {
		case MetricTypeViews:
			value = float64(stat.TotalViews)
		case MetricTypeLikes:
			value = float64(stat.TotalLikes)
		case MetricTypeComments:
			value = float64(stat.TotalComments)
		case MetricTypeShares:
			value = float64(stat.TotalShares)
		case MetricTypeEngagement:
			value = CalculateEngagement(stat.TotalLikes, stat.TotalComments, stat.TotalShares, stat.TotalViews)
		}
		result = append(result, TrendData{
			Date:  stat.Date.Format("2006-01-02"),
			Value: value,
		})
	}

	return result, nil
}

// GenerateMockData generates mock analytics data for testing
func (s *JSONStorage) GenerateMockData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	platforms := []Platform{PlatformDouyin, PlatformToutiao, PlatformXiaohongshu}

	for i := 0; i < 10; i++ {
		for _, p := range platforms {
			postID := string(p) + "_post_" + string(rune('A'+i))
			s.postMetrics[postID] = &PostMetrics{
				PostID:      postID,
				Platform:    p,
				Title:       "示例内容 " + string(rune('A'+i)),
				Views:       int64(1000 + i*500 + (int(p[0]) % 1000)),
				Likes:       int64(50 + i*20),
				Comments:    int64(10 + i*5),
				Shares:      int64(5 + i*2),
				Favorites:   int64(20 + i*8),
				Engagement:  float64(5 + i),
				PublishedAt: now.AddDate(0, 0, -i),
				CollectedAt: now,
			}
		}
	}

	for i := 0; i < 7; i++ {
		for _, p := range platforms {
			key := now.AddDate(0, 0, -i).Format("2006-01-02") + "_" + string(p)
			s.dailyStats[key] = &DailyStats{
				Date:          now.AddDate(0, 0, -i),
				Platform:      p,
				PostsCount:    2 + i%3,
				TotalViews:    int64(2000 + i*500 + (int(p[0]) % 1000)),
				TotalLikes:    int64(100 + i*20),
				TotalComments: int64(30 + i*5),
				TotalShares:   int64(10 + i*2),
				NewFollowers:  int64(50 + i*10),
			}
		}
	}

	return s.save()
}
