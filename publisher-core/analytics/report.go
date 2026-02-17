package analytics

import (
	"encoding/json"
	"fmt"
	"time"
)

type ReportGenerator struct {
	storage MetricsStorage
}

func NewReportGenerator(storage MetricsStorage) *ReportGenerator {
	return &ReportGenerator{
		storage: storage,
	}
}

type Report struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Period      TimeRange      `json:"period"`
	Summary     ReportSummary  `json:"summary"`
	Platforms   []PlatformData `json:"platforms"`
	TopPosts    []*PostMetrics `json:"top_posts"`
	Insights    []string       `json:"insights"`
}

type ReportSummary struct {
	TotalPosts    int64   `json:"total_posts"`
	TotalViews    int64   `json:"total_views"`
	TotalLikes    int64   `json:"total_likes"`
	TotalComments int64   `json:"total_comments"`
	TotalShares   int64   `json:"total_shares"`
	AvgEngagement float64 `json:"avg_engagement"`
	GrowthRate    float64 `json:"growth_rate"`
}

type PlatformData struct {
	Platform       Platform `json:"platform"`
	Posts          int64    `json:"posts"`
	Views          int64    `json:"views"`
	Likes          int64    `json:"likes"`
	Comments       int64    `json:"comments"`
	Engagement     float64  `json:"engagement"`
	BestPerforming string   `json:"best_performing"`
	GrowthRate     float64  `json:"growth_rate"`
}

func (g *ReportGenerator) GenerateWeeklyReport() (*Report, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -7)
	return g.GenerateReport(TimeRange{Start: start, End: end})
}

func (g *ReportGenerator) GenerateMonthlyReport() (*Report, error) {
	end := time.Now()
	start := end.AddDate(0, -1, 0)
	return g.GenerateReport(TimeRange{Start: start, End: end})
}

func (g *ReportGenerator) GenerateReport(period TimeRange) (*Report, error) {
	report := &Report{
		GeneratedAt: time.Now(),
		Period:      period,
		Platforms:   []PlatformData{},
		TopPosts:    []*PostMetrics{},
		Insights:    []string{},
	}

	platforms := []Platform{PlatformDouyin, PlatformXiaohongshu, PlatformToutiao}

	for _, platform := range platforms {
		stats, err := g.storage.GetDailyStats(platform, period.Start, period.End)
		if err != nil {
			continue
		}

		platformData := PlatformData{
			Platform: platform,
		}

		for _, s := range stats {
			platformData.Posts += int64(s.PostsCount)
			platformData.Views += s.TotalViews
			platformData.Likes += s.TotalLikes
			platformData.Comments += s.TotalComments

			report.Summary.TotalPosts += int64(s.PostsCount)
			report.Summary.TotalViews += s.TotalViews
			report.Summary.TotalLikes += s.TotalLikes
			report.Summary.TotalComments += s.TotalComments
			report.Summary.TotalShares += s.TotalShares
		}

		if platformData.Views > 0 {
			platformData.Engagement = CalculateEngagement(
				platformData.Likes,
				platformData.Comments,
				0,
				platformData.Views,
			)
		}

		report.Platforms = append(report.Platforms, platformData)
	}

	if report.Summary.TotalViews > 0 {
		report.Summary.AvgEngagement = CalculateEngagement(
			report.Summary.TotalLikes,
			report.Summary.TotalComments,
			report.Summary.TotalShares,
			report.Summary.TotalViews,
		)
	}

	for _, platform := range platforms {
		posts, err := g.storage.ListPostMetrics(platform, 5)
		if err == nil && len(posts) > 0 {
			report.TopPosts = append(report.TopPosts, posts...)
		}
	}

	report.Insights = g.generateInsights(report)

	return report, nil
}

func (g *ReportGenerator) generateInsights(report *Report) []string {
	insights := []string{}

	if report.Summary.TotalPosts > 0 {
		avgViews := report.Summary.TotalViews / report.Summary.TotalPosts
		insights = append(insights,
			fmt.Sprintf("Total posts: %d, avg views per post: %d",
				report.Summary.TotalPosts, avgViews))
	}

	if report.Summary.AvgEngagement > 5.0 {
		insights = append(insights, "Good engagement rate")
	} else if report.Summary.AvgEngagement > 2.0 {
		insights = append(insights, "Moderate engagement rate")
	} else {
		insights = append(insights, "Low engagement rate")
	}

	if len(report.Platforms) > 0 {
		bestPlatform := report.Platforms[0]
		for _, p := range report.Platforms {
			if p.Views > bestPlatform.Views {
				bestPlatform = p
			}
		}
		insights = append(insights,
			fmt.Sprintf("Best platform: %s", bestPlatform.Platform))
	}

	return insights
}

func (g *ReportGenerator) ExportJSON(report *Report) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
