package analytics

import (
	"encoding/json"
	"fmt"
	"time"
)

// VisualizationData 可视化数�?
type VisualizationData struct {
	Charts    []Chart     `json:"charts"`
	Summary   Summary     `json:"summary"`
	TimeRange TimeRange   `json:"time_range"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Chart 图表数据
type Chart struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"` // line, bar, pie, area
	Title   string      `json:"title"`
	Data    ChartData   `json:"data"`
	Options ChartOptions `json:"options"`
}

// ChartData 图表数据
type ChartData struct {
	Labels []string      `json:"labels"`
	Series []ChartSeries `json:"series"`
}

// ChartSeries 数据系列
type ChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
	Color  string    `json:"color"`
}

// ChartOptions 图表选项
type ChartOptions struct {
	XAxis      string `json:"x_axis,omitempty"`
	YAxis      string `json:"y_axis,omitempty"`
	ShowLegend bool   `json:"show_legend"`
	ShowGrid   bool   `json:"show_grid"`
	Stacked    bool   `json:"stacked"`
}

// Summary 汇总数�?
type Summary struct {
	TotalViews      int64     `json:"total_views"`
	TotalLikes      int64     `json:"total_likes"`
	AvgEngagement   float64   `json:"avg_engagement"`
	GrowthRate      float64   `json:"growth_rate"`
	TopPlatform     string    `json:"top_platform"`
	TopPostID       string    `json:"top_post_id"`
	BestPerformTime string    `json:"best_perform_time"`
}

// VisualizationGenerator 可视化数据生成器
type VisualizationGenerator struct {
	storage MetricsStorage
}

// NewVisualizationGenerator 创建可视化生成器
func NewVisualizationGenerator(storage MetricsStorage) *VisualizationGenerator {
	return &VisualizationGenerator{
		storage: storage,
	}
}

// GenerateDashboardCharts 生成仪表盘图�?
func (g *VisualizationGenerator) GenerateDashboardCharts(platform string, days int) (*VisualizationData, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	vizData := &VisualizationData{
		Charts:    make([]Chart, 0),
		TimeRange: TimeRange{Start: startTime, End: endTime},
		UpdatedAt: time.Now(),
	}

	// 1. 趋势线图
	trendChart, err := g.generateTrendChart(platform, days)
	if err == nil {
		vizData.Charts = append(vizData.Charts, *trendChart)
	}

	// 2. 平台分布饼图
	pieChart, err := g.generatePlatformPieChart(days)
	if err == nil {
		vizData.Charts = append(vizData.Charts, *pieChart)
	}

	// 3. 互动率柱状图
	barChart, err := g.generateEngagementBarChart(platform, days)
	if err == nil {
		vizData.Charts = append(vizData.Charts, *barChart)
	}

	// 4. 发布时间热力�?
	heatmap, err := g.generatePublishTimeHeatmap(platform)
	if err == nil {
		vizData.Charts = append(vizData.Charts, *heatmap)
	}

	return vizData, nil
}

// generateTrendChart 生成趋势�?
func (g *VisualizationGenerator) generateTrendChart(platform string, days int) (*Chart, error) {
	trends, err := g.storage.GetTrendData(MetricTypeViews, Platform(platform), days)
	if err != nil {
		return nil, err
	}

	chart := &Chart{
		ID:    "views_trend",
		Type:  "line",
		Title: "浏览量趋�?,
		Data: ChartData{
			Labels: make([]string, 0),
			Series: make([]ChartSeries, 0),
		},
		Options: ChartOptions{
			XAxis:      "日期",
			YAxis:      "浏览�?,
			ShowLegend: true,
			ShowGrid:   true,
		},
	}

	viewsSeries := ChartSeries{
		Name:   "浏览�?,
		Values: make([]float64, 0),
		Color:  "#3b82f6",
	}

	for _, t := range trends {
		chart.Data.Labels = append(chart.Data.Labels, t.Date)
		viewsSeries.Values = append(viewsSeries.Values, t.Value)
	}

	chart.Data.Series = append(chart.Data.Series, viewsSeries)
	return chart, nil
}

// generatePlatformPieChart 生成平台分布饼图
func (g *VisualizationGenerator) generatePlatformPieChart(days int) (*Chart, error) {
	chart := &Chart{
		ID:    "platform_distribution",
		Type:  "pie",
		Title: "平台内容分布",
		Data: ChartData{
			Labels: []string{"抖音", "小红�?, "头条"},
			Series: make([]ChartSeries, 0),
		},
		Options: ChartOptions{
			ShowLegend: true,
		},
	}

	platforms := []Platform{PlatformDouyin, PlatformXiaohongshu, PlatformToutiao}
	colors := []string{"#ef4444", "#ec4899", "#f97316"}

	values := make([]float64, 0)
	for i, p := range platforms {
		stats, err := g.storage.GetDailyStats(p, time.Now().AddDate(0, 0, -days), time.Now())
		if err != nil {
			values = append(values, 0)
			continue
		}

		var totalViews int64
		for _, s := range stats {
			totalViews += s.TotalViews
		}
		values = append(values, float64(totalViews))

		_ = colors[i]
	}

	series := ChartSeries{
		Name:   "浏览量分�?,
		Values: values,
	}
	chart.Data.Series = append(chart.Data.Series, series)

	return chart, nil
}

// generateEngagementBarChart 生成互动率柱状图
func (g *VisualizationGenerator) generateEngagementBarChart(platform string, days int) (*Chart, error) {
	chart := &Chart{
		ID:    "engagement_bar",
		Type:  "bar",
		Title: "互动率对�?,
		Data: ChartData{
			Labels: make([]string, 0),
			Series: make([]ChartSeries, 0),
		},
		Options: ChartOptions{
			XAxis:      "日期",
			YAxis:      "互动�?%)",
			ShowLegend: false,
			ShowGrid:   true,
		},
	}

	stats, err := g.storage.GetDailyStats(Platform(platform), time.Now().AddDate(0, 0, -days), time.Now())
	if err != nil {
		return nil, err
	}

	engagementSeries := ChartSeries{
		Name:   "互动�?,
		Values: make([]float64, 0),
		Color:  "#10b981",
	}

	for _, s := range stats {
		chart.Data.Labels = append(chart.Data.Labels, s.Date.Format("01-02"))
		if s.TotalViews > 0 {
			engagement := CalculateEngagement(s.TotalLikes, s.TotalComments, s.TotalShares, s.TotalViews)
			engagementSeries.Values = append(engagementSeries.Values, engagement)
		} else {
			engagementSeries.Values = append(engagementSeries.Values, 0)
		}
	}

	chart.Data.Series = append(chart.Data.Series, engagementSeries)
	return chart, nil
}

// generatePublishTimeHeatmap 生成发布时间热力�?
func (g *VisualizationGenerator) generatePublishTimeHeatmap(platform string) (*Chart, error) {
	chart := &Chart{
		ID:    "publish_heatmap",
		Type:  "heatmap",
		Title: "最佳发布时�?,
		Data: ChartData{
			Labels: []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
			Series: make([]ChartSeries, 0),
		},
		Options: ChartOptions{
			ShowLegend: true,
		},
	}

	// 生成时间段标�?
	for hour := 0; hour < 24; hour++ {
		chart.Data.Labels = append(chart.Data.Labels, fmt.Sprintf("%02d:00", hour))
	}

	// TODO: 从实际数据中统计各时间段的发布效�?
	// 当前使用模拟数据
	for i := 0; i < 7; i++ {
		values := make([]float64, 24)
		for j := 0; j < 24; j++ {
			// 模拟数据：早晚高峰效果更�?
			if j >= 8 && j <= 10 || j >= 18 && j <= 22 {
				values[j] = float64(60 + (i*j)%40)
			} else {
				values[j] = float64(20 + (i*j)%30)
			}
		}

		series := ChartSeries{
			Name:   chart.Data.Labels[i],
			Values: values,
		}
		chart.Data.Series = append(chart.Data.Series, series)
	}

	return chart, nil
}

// ExportChartAsJSON 导出图表为JSON
func (g *VisualizationGenerator) ExportChartAsJSON(chart *Chart) (string, error) {
	data, err := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExportVisualizationAsJSON 导出完整可视化数据为JSON
func (g *VisualizationGenerator) ExportVisualizationAsJSON(viz *VisualizationData) (string, error) {
	data, err := json.MarshalIndent(viz, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
