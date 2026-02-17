package analytics

import (
	"testing"
	"time"
)

func TestCalculateEngagement(t *testing.T) {
	tests := []struct {
		name     string
		likes    int64
		comments int64
		shares   int64
		views    int64
		want     float64
	}{
		{
			name:     "Normal case",
			likes:    100,
			comments: 50,
			shares:   30,
			views:    1000,
			want:     18.0,
		},
		{
			name:     "Zero views",
			likes:    100,
			comments: 50,
			shares:   30,
			views:    0,
			want:     0.0,
		},
		{
			name:     "Zero interactions",
			likes:    0,
			comments: 0,
			shares:   0,
			views:    1000,
			want:     0.0,
		},
		{
			name:     "Perfect engagement",
			likes:    500,
			comments: 300,
			shares:   200,
			views:    1000,
			want:     100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEngagement(tt.likes, tt.comments, tt.shares, tt.views)
			if got != tt.want {
				t.Errorf("CalculateEngagement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformConstants(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name:     "Douyin",
			platform: PlatformDouyin,
			want:     "douyin",
		},
		{
			name:     "Xiaohongshu",
			platform: PlatformXiaohongshu,
			want:     "xiaohongshu",
		},
		{
			name:     "Toutiao",
			platform: PlatformToutiao,
			want:     "toutiao",
		},
		{
			name:     "Weibo",
			platform: PlatformWeibo,
			want:     "weibo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.platform) != tt.want {
				t.Errorf("Platform %s = %v, want %v", tt.name, tt.platform, tt.want)
			}
		})
	}
}

func TestPostMetricsValidation(t *testing.T) {
	now := time.Now()

	metrics := &PostMetrics{
		PostID:      "test-post-1",
		Platform:    PlatformDouyin,
		Title:       "Test Post",
		Views:       1000,
		Likes:       100,
		Comments:    50,
		Shares:      30,
		Favorites:   20,
		CollectedAt: now,
		PublishedAt: now.Add(-24 * time.Hour),
	}

	if metrics.PostID != "test-post-1" {
		t.Error("PostID should be set correctly")
	}

	if metrics.Platform != PlatformDouyin {
		t.Error("Platform should be set correctly")
	}

	engagement := CalculateEngagement(metrics.Likes, metrics.Comments, metrics.Shares, metrics.Views)
	if engagement <= 0 {
		t.Error("Engagement should be positive")
	}
}

func TestJSONStorage(t *testing.T) {
	tempDir := t.TempDir()
	storage, err := NewJSONStorage(tempDir)
	if err != nil {
		t.Fatalf("NewJSONStorage failed: %v", err)
	}

	postMetrics := &PostMetrics{
		PostID:      "test-1",
		Platform:    PlatformDouyin,
		Title:       "Test",
		Views:       100,
		CollectedAt: time.Now(),
	}

	err = storage.SavePostMetrics(postMetrics)
	if err != nil {
		t.Errorf("SavePostMetrics failed: %v", err)
	}

	retrieved, err := storage.GetPostMetrics("test-1")
	if err != nil {
		t.Errorf("GetPostMetrics failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved post should not be nil")
	}

	if retrieved.PostID != postMetrics.PostID {
		t.Errorf("Retrieved post ID mismatch, got %s, want %s", retrieved.PostID, postMetrics.PostID)
	}
}
