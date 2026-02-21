// Package pipeline 提供预定义流水线模板
package pipeline

import (
	"fmt"
	"time"
)

// ContentPublishPipeline 内容发布流水线模板
func ContentPublishPipeline() *Pipeline {
	return &Pipeline{
		ID:          "content-publish-v1",
		Name:        "内容发布流水线",
		Description: "从内容生成到多平台发布的完整流程",
		Steps: []PipelineStep{
			{
				ID:       "step-1",
				Name:     "内容生成",
				Type:     StepTypeContentGeneration,
				Handler:  "ai_content_generator",
				Config: map[string]interface{}{
					"model":        "deepseek-chat",
					"max_tokens":   2000,
					"temperature":  0.7,
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-2",
				Name:     "内容优化",
				Type:     StepTypeContentOptimization,
				Handler:  "content_optimizer",
				DependsOn: []string{"step-1"},
				Config: map[string]interface{}{
					"check_spelling":     true,
					"improve_readability": true,
				},
				RetryCount: 2,
				Timeout:    2 * time.Minute,
			},
			{
				ID:       "step-3",
				Name:     "质量评分",
				Type:     StepTypeQualityScoring,
				Handler:  "quality_scorer",
				DependsOn: []string{"step-2"},
				Config: map[string]interface{}{
					"min_score":     0.7,
					"scoring_model": "quality-v2",
				},
				RetryCount: 1,
				Timeout:    1 * time.Minute,
			},
			{
				ID:       "step-4",
				Name:     "发布执行",
				Type:     StepTypePublishExecution,
				Handler:  "platform_publisher",
				DependsOn: []string{"step-3"},
				Config: map[string]interface{}{
					"platforms":  []string{"douyin", "toutiao", "xiaohongshu"},
					"async_mode": true,
				},
				RetryCount: 3,
				Timeout:    10 * time.Minute,
			},
			{
				ID:       "step-5",
				Name:     "数据采集",
				Type:     StepTypeDataCollection,
				Handler:  "analytics_collector",
				DependsOn: []string{"step-4"},
				Config: map[string]interface{}{
					"collect_immediately": true,
					"collect_after_hours": 24,
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
		},
		Config: PipelineConfig{
			ParallelMode: false,
			MaxParallel:  1,
			FailFast:     true,
			RetryStrategy: RetryStrategy{
				Type:          RetryTypeExponential,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Notification: NotificationConfig{
				OnStart:    true,
				OnComplete: true,
				OnError:    true,
				Channels:   []string{"websocket", "email"},
			},
		},
	}
}

// VideoProcessingPipeline 视频处理流水线模板
func VideoProcessingPipeline() *Pipeline {
	return &Pipeline{
		ID:          "video-processing-v1",
		Name:        "视频处理流水线",
		Description: "视频下载、转录、切片、发布的完整流程",
		Steps: []PipelineStep{
			{
				ID:       "step-1",
				Name:     "视频下载",
				Type:     StepTypeDataCollection,
				Handler:  "video_downloader",
				Config: map[string]interface{}{
					"max_retries": 3,
					"timeout":      "10m",
				},
				RetryCount: 3,
				Timeout:    10 * time.Minute,
			},
			{
				ID:       "step-2",
				Name:     "语音转录",
				Type:     StepTypeContentGeneration,
				Handler:  "speech_transcriber",
				DependsOn: []string{"step-1"},
				Config: map[string]interface{}{
					"strategy":      "cloud_first", // cloud_first, local_only, hybrid
					"cloud_service": "bcut_asr",
					"local_service": "whisper",
				},
				RetryCount: 3,
				Timeout:    15 * time.Minute,
			},
			{
				ID:       "step-3",
				Name:     "内容改写",
				Type:     StepTypeContentOptimization,
				Handler:  "content_rewriter",
				DependsOn: []string{"step-2"},
				Config: map[string]interface{}{
					"style":      "casual",
					"max_length": 1000,
				},
				RetryCount: 2,
				Timeout:    3 * time.Minute,
			},
			{
				ID:       "step-4",
				Name:     "视频切片",
				Type:     StepTypeDataCollection,
				Handler:  "video_cutter",
				DependsOn: []string{"step-3"},
				Config: map[string]interface{}{
					"max_duration":  60,
					"output_format": "mp4",
				},
				RetryCount: 2,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-5",
				Name:     "发布执行",
				Type:     StepTypePublishExecution,
				Handler:  "platform_publisher",
				DependsOn: []string{"step-4"},
				Config: map[string]interface{}{
					"platforms": []string{"douyin", "xiaohongshu"},
				},
				RetryCount: 3,
				Timeout:    10 * time.Minute,
			},
		},
		Config: PipelineConfig{
			ParallelMode: false,
			MaxParallel:  1,
			FailFast:     false,
			RetryStrategy: RetryStrategy{
				Type:          RetryTypeExponential,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Notification: NotificationConfig{
				OnStart:    true,
				OnComplete: true,
				OnError:    true,
				Channels:   []string{"websocket"},
			},
		},
	}
}

// HotspotAnalysisPipeline 热点分析流水线模板
func HotspotAnalysisPipeline() *Pipeline {
	return &Pipeline{
		ID:          "hotspot-analysis-v1",
		Name:        "热点分析流水线",
		Description: "抓取热点、分析趋势、生成内容的完整流程",
		Steps: []PipelineStep{
			{
				ID:       "step-1",
				Name:     "热点抓取",
				Type:     StepTypeDataCollection,
				Handler:  "hotspot_fetcher",
				Config: map[string]interface{}{
					"sources": []string{"newsnow", "toutiao"},
					"limit":   20,
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-2",
				Name:     "趋势分析",
				Type:     StepTypeAnalytics,
				Handler:  "trend_analyzer",
				DependsOn: []string{"step-1"},
				Config: map[string]interface{}{
					"analysis_type": "keyword_extraction",
				},
				RetryCount: 2,
				Timeout:    3 * time.Minute,
			},
			{
				ID:       "step-3",
				Name:     "内容生成",
				Type:     StepTypeContentGeneration,
				Handler:  "ai_content_generator",
				DependsOn: []string{"step-2"},
				Config: map[string]interface{}{
					"model": "deepseek-chat",
					"style": "professional",
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-4",
				Name:     "发布执行",
				Type:     StepTypePublishExecution,
				Handler:  "platform_publisher",
				DependsOn: []string{"step-3"},
				Config: map[string]interface{}{
					"platforms": []string{"douyin", "xiaohongshu"},
				},
				RetryCount: 3,
				Timeout:    10 * time.Minute,
			},
		},
		Config: PipelineConfig{
			ParallelMode: false,
			MaxParallel:  1,
			FailFast:     true,
			RetryStrategy: RetryStrategy{
				Type:          RetryTypeExponential,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Notification: NotificationConfig{
				OnStart:    true,
				OnComplete: true,
				OnError:    true,
				Channels:   []string{"websocket"},
			},
		},
	}
}

// DataCollectionPipeline 数据采集流水线模板
func DataCollectionPipeline() *Pipeline {
	return &Pipeline{
		ID:          "data-collection-v1",
		Name:        "数据采集流水线",
		Description: "从多平台采集发布数据和性能指标",
		Steps: []PipelineStep{
			{
				ID:       "step-1",
				Name:     "抖音数据采集",
				Type:     StepTypeDataCollection,
				Handler:  "douyin_collector",
				Config: map[string]interface{}{
					"metrics": []string{"views", "likes", "comments", "shares"},
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-2",
				Name:     "今日头条数据采集",
				Type:     StepTypeDataCollection,
				Handler:  "toutiao_collector",
				Config: map[string]interface{}{
					"metrics": []string{"views", "likes", "comments", "shares"},
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-3",
				Name:     "小红书数据采集",
				Type:     StepTypeDataCollection,
				Handler:  "xiaohongshu_collector",
				Config: map[string]interface{}{
					"metrics": []string{"views", "likes", "comments", "shares"},
				},
				RetryCount: 3,
				Timeout:    5 * time.Minute,
			},
			{
				ID:       "step-4",
				Name:     "数据分析",
				Type:     StepTypeAnalytics,
				Handler:  "data_analyzer",
				DependsOn: []string{"step-1", "step-2", "step-3"},
				Config: map[string]interface{}{
					"analysis_type": "performance_report",
				},
				RetryCount: 2,
				Timeout:    3 * time.Minute,
			},
			{
				ID:       "step-5",
				Name:     "报告生成",
				Type:     StepTypeAnalytics,
				Handler:  "report_generator",
				DependsOn: []string{"step-4"},
				Config: map[string]interface{}{
					"format": "markdown",
				},
				RetryCount: 1,
				Timeout:    2 * time.Minute,
			},
		},
		Config: PipelineConfig{
			ParallelMode: true,
			MaxParallel:  3,
			FailFast:     false,
			RetryStrategy: RetryStrategy{
				Type:          RetryTypeExponential,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Notification: NotificationConfig{
				OnStart:    false,
				OnComplete: true,
				OnError:    true,
				Channels:   []string{"email"},
			},
		},
	}
}

// GetTemplate 获取模板
func GetTemplate(templateID string) (*Pipeline, error) {
	switch templateID {
	case "content-publish-v1":
		return ContentPublishPipeline(), nil
	case "video-processing-v1":
		return VideoProcessingPipeline(), nil
	case "hotspot-analysis-v1":
		return HotspotAnalysisPipeline(), nil
	case "data-collection-v1":
		return DataCollectionPipeline(), nil
	default:
		return nil, fmt.Errorf("模板不存在: %s", templateID)
	}
}

// ListTemplates 列出所有模板
func ListTemplates() []*Pipeline {
	return []*Pipeline{
		ContentPublishPipeline(),
		VideoProcessingPipeline(),
		HotspotAnalysisPipeline(),
		DataCollectionPipeline(),
	}
}
