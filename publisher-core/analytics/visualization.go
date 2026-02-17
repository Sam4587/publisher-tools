package analytics

import (
	"encoding/json"
	"time"
)

type VisualizationData struct {
	Charts    []Chart   `json:"charts"`
	Summary   Summary   `json:"summary"`
	TimeRange TimeRange `json:"time_range"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Chart struct {
	ID      string       `json:"id"`
	Type    string       `json:"type"` // line, bar, pie, area
	Title   string       `json:"title"`
	Data    ChartData    `json:"data"`
	Options ChartOptions `json:"options"`
}

type ChartData struct {
	Labels []string      `json:"labels"`
	Series []ChartSeries `json:"series"`
}

type ChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
	Color  string    `json:"color"`
}

type ChartOptions struct {
	XAxis      string `json:"x_axis,omitempty"`
	YAxis      string `json:"y_axis,omitempty"`
	ShowLegend bool   `json:"show_legend"`
	ShowGrid   bool   `json:"show_grid"`
	Stacked    bool   `json:"stacked"`
}

type Summary struct {
	TotalViews    int64   `json:"total_views"`
	TotalLikes    int64   `json:"total_likes"`
	TotalComments int64   `json:"total_comments"`
	TotalShares   int64   `json:"total_shares"`
	TotalPosts    int64   `json:"total_posts"`
	AvgEngagement float64 `json:"avg_engagement"`
	BestPlatform  string  `json:"best_platform"`
	GrowthRate    float64 `json:"growth_rate"`
}

type Visualizer struct {
	storage MetricsStorage
}

func NewVisualizer(storage MetricsStorage) *Visualizer {
	return &Visualizer{
		storage: storage,
	}
}

func (v *Visualizer) GenerateVisualization(start, end time.Time) (*VisualizationData, error) {
	data := &VisualizationData{
		Charts:    []Chart{},
		Summary:   Summary{},
		TimeRange: TimeRange{Start: start, End: end},
		UpdatedAt: time.Now(),
	}

	platforms := []Platform{PlatformDouyin, PlatformXiaohongshu, PlatformToutiao}

	for _, platform := range platforms {
		stats, err := v.storage.GetDailyStats(platform, start, end)
		if err != nil {
			continue
		}

		for _, s := range stats {
			data.Summary.TotalViews += s.TotalViews
			data.Summary.TotalLikes += s.TotalLikes
			data.Summary.TotalComments += s.TotalComments
			data.Summary.TotalShares += s.TotalShares
			data.Summary.TotalPosts += int64(s.PostsCount)
		}
	}

	if data.Summary.TotalViews > 0 {
		data.Summary.AvgEngagement = CalculateEngagement(
			data.Summary.TotalLikes,
			data.Summary.TotalComments,
			data.Summary.TotalShares,
			data.Summary.TotalViews,
		)
	}

	return data, nil
}

func (v *Visualizer) ExportJSON(data *VisualizationData) (string, error) {
	result, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(result), nil
}
