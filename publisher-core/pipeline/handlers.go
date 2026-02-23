// Package pipeline 提供流水线步骤处理器实现
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"publisher-core/ai"
	"publisher-core/ai/provider"
	"publisher-core/adapters"
	"publisher-core/analytics"
	publisher "publisher-core/interfaces"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// HandlerRegistry 处理器注册表
type HandlerRegistry struct {
	orchestrator *PipelineOrchestrator
	aiService   *ai.Service
	publisher   *adapters.PublisherFactory
	analytics   *analytics.AnalyticsService
}

// NewHandlerRegistry 创建处理器注册表
func NewHandlerRegistry(
	orchestrator *PipelineOrchestrator,
	aiService *ai.Service,
	publisher *adapters.PublisherFactory,
	analyticsService *analytics.AnalyticsService,
) *HandlerRegistry {
	registry := &HandlerRegistry{
		orchestrator: orchestrator,
		aiService:   aiService,
		publisher:   publisher,
		analytics:   analyticsService,
	}

	registry.registerAllHandlers()

	return registry
}

// registerAllHandlers 注册所有处理器
func (r *HandlerRegistry) registerAllHandlers() {
	// AI 内容生成处理器
	r.orchestrator.RegisterHandler("ai_content_generator", &AIContentGenerator{
		aiService: r.aiService,
	})

	// 内容优化处理器
	r.orchestrator.RegisterHandler("content_optimizer", &ContentOptimizer{
		aiService: r.aiService,
	})

	// 质量评分处理器
	r.orchestrator.RegisterHandler("quality_scorer", &QualityScorer{
		aiService: r.aiService,
	})

	// 平台发布处理器
	r.orchestrator.RegisterHandler("platform_publisher", &PlatformPublisher{
		publisher: r.publisher,
	})

	// 数据采集处理器
	r.orchestrator.RegisterHandler("analytics_collector", &AnalyticsCollector{
		analyticsService: r.analytics,
	})

	// 视频下载处理器
	r.orchestrator.RegisterHandler("video_downloader", &VideoDownloader{})

	// 语音转录处理器
	r.orchestrator.RegisterHandler("speech_transcriber", &SpeechTranscriber{})

	// 内容改写处理器
	r.orchestrator.RegisterHandler("content_rewriter", &ContentRewriter{
		aiService: r.aiService,
	})

	// 视频切片处理器
	r.orchestrator.RegisterHandler("video_cutter", &VideoCutter{})

	// 热点抓取处理器
	r.orchestrator.RegisterHandler("hotspot_fetcher", &HotspotFetcher{})

	// 趋势分析处理器
	r.orchestrator.RegisterHandler("trend_analyzer", &TrendAnalyzer{
		aiService: r.aiService,
	})

	// 数据分析处理器
	r.orchestrator.RegisterHandler("data_analyzer", &DataAnalyzer{
		aiService: r.aiService,
	})

	// 报告生成处理器
	r.orchestrator.RegisterHandler("report_generator", &ReportGenerator{})

	logrus.Info("所有步骤处理器已注册")
}

// AIContentGenerator AI内容生成处理器
type AIContentGenerator struct {
	aiService *ai.Service
}

func (h *AIContentGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	model, _ := config["model"].(string)
	if model == "" {
		model = "deepseek-chat"
	}

	topic, ok := input["topic"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 topic 参数")
	}

	keywords, _ := input["keywords"].([]string)
	targetAudience, _ := input["target_audience"].(string)

	// 构建提示词
	prompt := fmt.Sprintf(`请根据以下信息生成一篇高质量的内容：

主题: %s
关键词: %s
目标受众: %s

要求:
1. 内容要吸引人，有吸引力
2. 结构清晰，逻辑连贯
3. 符合目标受众的阅读习惯
4. 包含适当的emoji表情
5. 字数控制在500-1000字

请直接生成内容，不要包含任何解释。`,
		topic, strings.Join(keywords, ", "), targetAudience)

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:    model,
		Messages: []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens: 2000,
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("AI生成失败: %w", err)
	}

	return map[string]interface{}{
		"content":     result.Content,
		"tokens_used": result.InputTokens + result.OutputTokens,
		"model":       model,
	}, nil
}

// ContentOptimizer 内容优化处理器
type ContentOptimizer struct {
	aiService *ai.Service
}

func (h *ContentOptimizer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	content, ok := input["content"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 content 参数")
	}

	checkSpelling, _ := config["check_spelling"].(bool)
	improveReadability, _ := config["improve_readability"].(bool)

	// 构建优化提示词
	prompt := fmt.Sprintf(`请优化以下内容：

%s

优化要求:
%s%s

请直接输出优化后的内容，不要包含任何解释。`,
		content,
		map[bool]string{true: "1. 检查并修正拼写错误\n"}[checkSpelling],
		map[bool]string{true: "2. 改善可读性和流畅度\n"}[improveReadability])

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:       "deepseek-chat",
		Messages:    []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:   2000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("内容优化失败: %w", err)
	}

	return map[string]interface{}{
		"optimized_content": result.Content,
		"tokens_used":       result.InputTokens + result.OutputTokens,
	}, nil
}

// QualityScorer 质量评分处理器
type QualityScorer struct {
	aiService *ai.Service
}

func (h *QualityScorer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	content, ok := input["optimized_content"].(string)
	if !ok {
		content, ok = input["content"].(string)
		if !ok {
			return nil, fmt.Errorf("缺少 content 或 optimized_content 参数")
		}
	}

	minScore, _ := config["min_score"].(float64)
	if minScore == 0 {
		minScore = 0.7
	}

	// 构建评分提示词
	prompt := fmt.Sprintf(`请对以下内容进行质量评分（0-1分）：

%s

评分维度:
1. 内容质量（30%）
2. 吸引力（25%）
3. 可读性（25%）
4. 完整性（20%）

请以JSON格式返回评分结果，格式如下:
{
  "overall_score": 0.85,
  "content_quality": 0.9,
  "attractiveness": 0.85,
  "readability": 0.8,
  "completeness": 0.85,
  "reasoning": "内容质量优秀，逻辑清晰，具有较强的吸引力"
}`,
		content)

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:       "deepseek-chat",
		Messages:    []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:   500,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("质量评分失败: %w", err)
	}

	// 解析评分结果
	var scoreResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &scoreResult); err != nil {
		return nil, fmt.Errorf("解析评分结果失败: %w", err)
	}

	overallScore, _ := scoreResult["overall_score"].(float64)
	passed := overallScore >= minScore

	return map[string]interface{}{
		"score":   overallScore,
		"passed":  passed,
		"details": scoreResult,
	}, nil
}

// PlatformPublisher 平台发布处理器
type PlatformPublisher struct {
	publisher *adapters.PublisherFactory
}

func (h *PlatformPublisher) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	platforms, ok := config["platforms"].([]string)
	if !ok {
		return nil, fmt.Errorf("缺少 platforms 参数")
	}

	contentStr, ok := input["optimized_content"].(string)
	if !ok {
		contentStr, ok = input["content"].(string)
		if !ok {
			return nil, fmt.Errorf("缺少 content 或 optimized_content 参数")
		}
	}

	title, _ := input["topic"].(string)

	results := make(map[string]interface{})
	successCount := 0

	for _, platform := range platforms {
		logrus.Infof("开始发布到平台: %s", platform)

		// 获取适配器
		adapter, err := h.publisher.Create(platform, nil)
		if err != nil {
			results[platform] = map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("适配器创建失败: %v", err),
			}
			continue
		}

		// 构建发布内容
		publishContent := &publisher.Content{
			Title: title,
			Body:  contentStr,
		}

		// 执行发布
		result, err := adapter.Publish(ctx, publishContent)
		if err != nil {
			logrus.Errorf("发布到 %s 失败: %v", platform, err)
			results[platform] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			continue
		}

		results[platform] = map[string]interface{}{
			"success": true,
			"url":     result.PostURL,
			"post_id": result.PostID,
		}
		successCount++

		logrus.Infof("成功发布到 %s: %s", platform, result.PostURL)
	}

	return map[string]interface{}{
		"results":      results,
		"success_count": successCount,
		"total_count":  len(platforms),
	}, nil
}

// AnalyticsCollector 数据采集处理器
type AnalyticsCollector struct {
	analyticsService *analytics.AnalyticsService
}

func (h *AnalyticsCollector) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	collectImmediately, _ := config["collect_immediately"].(bool)
	collectAfterHours, _ := config["collect_after_hours"].(int)

	results := make(map[string]interface{})

	if collectImmediately {
		// 立即采集数据
		logrus.Info("开始立即采集数据")

		// 调用分析服务采集数据
		// 这里需要根据实际的分析服务接口实现
		results["immediate_collected"] = true
		results["collected_at"] = time.Now().Format(time.RFC3339)
	}

	if collectAfterHours > 0 {
		// 计划延迟采集
		logrus.Infof("计划在 %d 小时后采集数据", collectAfterHours)

		results["scheduled_collect"] = true
		results["scheduled_at"] = time.Now().Add(time.Duration(collectAfterHours) * time.Hour).Format(time.RFC3339)
	}

	return results, nil
}

// VideoDownloader 视频下载处理器
type VideoDownloader struct{}

func (h *VideoDownloader) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	videoURL, ok := input["video_url"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 video_url 参数")
	}

	maxRetries, _ := config["max_retries"].(int)
	if maxRetries == 0 {
		maxRetries = 3
	}

	logrus.Infof("开始下载视频: %s (最大重试次数: %d)", videoURL, maxRetries)

	// 这里实现视频下载逻辑
	// 可以使用 yt-dlp 或其他下载工具

	// 模拟下载
	time.Sleep(2 * time.Second)

	return map[string]interface{}{
		"video_path":   "/downloads/video.mp4",
		"duration":     180, // 3分钟
		"file_size":    "25.5MB",
		"downloaded_at": time.Now().Format(time.RFC3339),
	}, nil
}

// SpeechTranscriber 语音转录处理器
type SpeechTranscriber struct{}

func (h *SpeechTranscriber) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	videoPath, ok := input["video_path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 video_path 参数")
	}

	strategy, _ := config["strategy"].(string)
	if strategy == "" {
		strategy = "cloud_first"
	}

	logrus.Infof("开始语音转录: %s (策略: %s)", videoPath, strategy)

	// 这里实现语音转录逻辑
	// 可以使用 bcut-asr 或 whisper

	// 模拟转录
	time.Sleep(3 * time.Second)

	transcript := `这是一个示例转录文本。
实际使用时，这里会是真正的语音转录结果。`

	return map[string]interface{}{
		"transcript":   transcript,
		"srt_path":     "/downloads/subtitles.srt",
		"word_count":   100,
		"duration":     180,
		"transcribed_at": time.Now().Format(time.RFC3339),
	}, nil
}

// ContentRewriter 内容改写处理器
type ContentRewriter struct {
	aiService *ai.Service
}

func (h *ContentRewriter) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	transcript, ok := input["transcript"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 transcript 参数")
	}

	style, _ := config["style"].(string)
	if style == "" {
		style = "casual"
	}

	maxLength, _ := config["max_length"].(int)
	if maxLength == 0 {
		maxLength = 1000
	}

	// 构建改写提示词
	prompt := fmt.Sprintf(`请将以下转录文本改写为适合发布的内容：

%s

改写要求:
1. 风格: %s
2. 字数限制: %d字以内
3. 语言要生动有趣
4. 保留核心信息
5. 适当添加emoji表情

请直接输出改写后的内容，不要包含任何解释。`,
		transcript, style, maxLength)

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:       "deepseek-chat",
		Messages:    []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:   1500,
		Temperature: 0.8,
	})
	if err != nil {
		return nil, fmt.Errorf("内容改写失败: %w", err)
	}

	return map[string]interface{}{
		"rewritten_content": result.Content,
		"tokens_used":       result.InputTokens + result.OutputTokens,
	}, nil
}

// VideoCutter 视频切片处理器
type VideoCutter struct{}

func (h *VideoCutter) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	videoPath, ok := input["video_path"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 video_path 参数")
	}

	maxDuration, _ := config["max_duration"].(int)
	if maxDuration == 0 {
		maxDuration = 60
	}

	outputFormat, _ := config["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "mp4"
	}

	logrus.Infof("开始视频切片: %s (最大时长: %d秒, 输出格式: %s)", videoPath, maxDuration, outputFormat)

	// 这里实现视频切片逻辑
	// 可以使用 FFmpeg

	// 模拟切片
	time.Sleep(2 * time.Second)

	clips := []map[string]interface{}{
		{
			"clip_id":   "clip-1",
			"path":      "/downloads/clip-1.mp4",
			"start_time": 0,
			"end_time":   45,
			"duration":   45,
		},
		{
			"clip_id":   "clip-2",
			"path":      "/downloads/clip-2.mp4",
			"start_time": 45,
			"end_time":   90,
			"duration":   45,
		},
	}

	return map[string]interface{}{
		"clips":        clips,
		"clip_count":   len(clips),
		"cut_at":       time.Now().Format(time.RFC3339),
	}, nil
}

// HotspotFetcher 热点抓取处理器
type HotspotFetcher struct{}

func (h *HotspotFetcher) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	sources, ok := config["sources"].([]string)
	if !ok {
		sources = []string{"newsnow", "toutiao"}
	}

	limit, _ := config["limit"].(int)
	if limit == 0 {
		limit = 20
	}

	logrus.Infof("开始抓取热点数据 (来源: %v, 限制: %d)", sources, limit)

	// 这里实现热点抓取逻辑
	// 可以调用现有的热点服务

	// 模拟抓取
	time.Sleep(2 * time.Second)

	hotspots := []map[string]interface{}{
		{
			"id":          "hot-1",
			"title":       "人工智能最新突破",
			"source":      "newsnow",
			"url":         "https://example.com/news/1",
			"hot_score":   95,
			"published_at": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":          "hot-2",
			"title":       "机器学习新算法",
			"source":      "toutiao",
			"url":         "https://example.com/news/2",
			"hot_score":   88,
			"published_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		},
	}

	return map[string]interface{}{
		"hotspots":    hotspots,
		"count":       len(hotspots),
		"fetched_at":  time.Now().Format(time.RFC3339),
	}, nil
}

// TrendAnalyzer 趋势分析处理器
type TrendAnalyzer struct {
	aiService *ai.Service
}

func (h *TrendAnalyzer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	hotspots, ok := input["hotspots"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 hotspots 参数")
	}

	analysisType, _ := config["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "keyword_extraction"
	}

	logrus.Infof("开始趋势分析 (分析类型: %s)", analysisType)

	// 提取热点标题
	var titles []string
	for _, hotspot := range hotspots {
		if title, ok := hotspot["title"].(string); ok {
			titles = append(titles, title)
		}
	}

	// 构建分析提示词
	prompt := fmt.Sprintf(`请分析以下热点标题，提取关键趋势和关键词：

%s

分析要求:
1. 提取主要关键词（最多10个）
2. 识别热门话题
3. 分析趋势方向
4. 给出推荐的主题

请以JSON格式返回分析结果。`,
		strings.Join(titles, "\n"))

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:       "deepseek-chat",
		Messages:    []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:   1000,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("趋势分析失败: %w", err)
	}

	var analysisResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Content), &analysisResult); err != nil {
		// 如果解析失败，返回原始结果
		analysisResult = map[string]interface{}{
			"raw_analysis": result.Content,
		}
	}

	return map[string]interface{}{
		"analysis":    analysisResult,
		"analyzed_at": time.Now().Format(time.RFC3339),
	}, nil
}

// DataAnalyzer 数据分析处理器
type DataAnalyzer struct {
	aiService *ai.Service
}

func (h *DataAnalyzer) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	data, ok := input["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少 data 参数")
	}

	analysisType, _ := config["analysis_type"].(string)
	if analysisType == "" {
		analysisType = "performance_report"
	}

	logrus.Infof("开始数据分析 (分析类型: %s)", analysisType)

	// 将数据转换为 JSON 字符串
	dataJSON, _ := json.Marshal(data)

	// 构建分析提示词
	prompt := fmt.Sprintf(`请分析以下数据，生成性能报告：

%s

报告要求:
1. 总结关键指标
2. 识别趋势和模式
3. 提供优化建议
4. 给出行动建议

请以Markdown格式输出报告。`,
		string(dataJSON))

	// 调用 AI 服务
	result, err := h.aiService.Generate(ctx, &provider.GenerateOptions{
		Model:       "deepseek-chat",
		Messages:    []provider.Message{{Role: provider.RoleUser, Content: prompt}},
		MaxTokens:   2000,
		Temperature: 0.4,
	})
	if err != nil {
		return nil, fmt.Errorf("数据分析失败: %w", err)
	}

	return map[string]interface{}{
		"report":      result.Content,
		"analyzed_at": time.Now().Format(time.RFC3339),
	}, nil
}

// ReportGenerator 报告生成处理器
type ReportGenerator struct{}

func (h *ReportGenerator) Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error) {
	_, ok := input["report"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少 report 参数")
	}

	format, _ := config["format"].(string)
	if format == "" {
		format = "markdown"
	}

	logrus.Infof("开始生成报告 (格式: %s)", format)

	// 保存报告
	reportPath := fmt.Sprintf("/reports/report-%d.%s", time.Now().Unix(), format)

	// 这里实现报告保存逻辑
	// 可以将报告保存到文件系统或数据库

	return map[string]interface{}{
		"report_path": reportPath,
		"format":      format,
		"generated_at": time.Now().Format(time.RFC3339),
	}, nil
}
