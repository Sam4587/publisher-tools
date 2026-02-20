package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"publisher-core/database"
	"publisher-core/hotspot"
	"publisher-core/video"
	"publisher-core/notify"

	"gorm.io/gorm"
)

// ToolsRegistry 工具注册器
type ToolsRegistry struct {
	db            *gorm.DB
	hotspotService *hotspot.EnhancedService
	videoService  *video.Service
	notifyService *notify.Service
}

// NewToolsRegistry 创建工具注册器
func NewToolsRegistry(db *gorm.DB) *ToolsRegistry {
	return &ToolsRegistry{
		db: db,
	}
}

// SetHotspotService 设置热点服务
func (r *ToolsRegistry) SetHotspotService(svc *hotspot.EnhancedService) {
	r.hotspotService = svc
}

// SetVideoService 设置视频服务
func (r *ToolsRegistry) SetVideoService(svc *video.Service) {
	r.videoService = svc
}

// SetNotifyService 设置通知服务
func (r *ToolsRegistry) SetNotifyService(svc *notify.Service) {
	r.notifyService = svc
}

// RegisterAllTools 注册所有工具
func (r *ToolsRegistry) RegisterAllTools(server *Server) error {
	// 热点监控工具
	r.registerHotspotTools(server)

	// 视频处理工具
	r.registerVideoTools(server)

	// 通知工具
	r.registerNotifyTools(server)

	// AI 配置工具
	r.registerAIConfigTools(server)

	// 系统工具
	r.registerSystemTools(server)

	return nil
}

// =====================================================
// 热点监控工具
// =====================================================

func (r *ToolsRegistry) registerHotspotTools(server *Server) {
	// 获取热点话题
	server.RegisterTool(Tool{
		Name:        "get_hot_topics",
		Description: "获取指定平台的热点话题列表",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"platform": map[string]interface{}{
					"type":        "string",
					"description": "平台ID（weibo/douyin/zhihu/baidu/toutiao等）",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量，默认20",
					"default":     20,
				},
				"min_heat": map[string]interface{}{
					"type":        "integer",
					"description": "最低热度过滤",
				},
			},
		},
		Handler: r.handleGetHotTopics,
	})

	// 获取趋势话题
	server.RegisterTool(Tool{
		Name:        "get_trending_topics",
		Description: "获取趋势上升的热点话题",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量，默认10",
					"default":     10,
				},
			},
		},
		Handler: r.handleGetTrendingTopics,
	})

	// 搜索话题
	server.RegisterTool(Tool{
		Name:        "search_topics",
		Description: "搜索热点话题",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"keyword": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量，默认20",
					"default":     20,
				},
			},
			"required": []string{"keyword"},
		},
		Handler: r.handleSearchTopics,
	})

	// 获取话题详情
	server.RegisterTool(Tool{
		Name:        "get_topic_detail",
		Description: "获取话题详情和历史",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"topic_id": map[string]interface{}{
					"type":        "string",
					"description": "话题ID",
				},
			},
			"required": []string{"topic_id"},
		},
		Handler: r.handleGetTopicDetail,
	})

	// 刷新热点数据
	server.RegisterTool(Tool{
		Name:        "refresh_hotspots",
		Description: "刷新热点数据",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "数据源ID，不指定则刷新所有",
				},
			},
		},
		Handler: r.handleRefreshHotspots,
	})

	// 获取热点统计
	server.RegisterTool(Tool{
		Name:        "get_hotspot_stats",
		Description: "获取热点监控统计信息",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: r.handleGetHotspotStats,
	})
}

func (r *ToolsRegistry) handleGetHotTopics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	platform, _ := args["platform"].(string)
	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	filter := database.TopicFilter{
		Source: platform,
		Limit:  limit,
		SortBy: "heat",
		SortDesc: true,
	}

	if minHeat, ok := args["min_heat"].(float64); ok {
		filter.MinHeat = int(minHeat)
	}

	topics, total, err := database.NewHotspotStorage(r.db).List(filter)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"total":   total,
		"topics":  topics,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetTrendingTopics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	topics, err := database.NewHotspotStorage(r.db).GetTrendingTopics(limit)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"topics":  topics,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleSearchTopics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	keyword, ok := args["keyword"].(string)
	if !ok {
		return ErrorResult(fmt.Errorf("keyword is required")), nil
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// 搜索话题
	var topics []database.Topic
	r.db.Where("title LIKE ?", "%"+keyword+"%").
		Order("heat desc").
		Limit(limit).
		Find(&topics)

	result := map[string]interface{}{
		"success": true,
		"keyword": keyword,
		"topics":  topics,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetTopicDetail(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	topicID, ok := args["topic_id"].(string)
	if !ok {
		return ErrorResult(fmt.Errorf("topic_id is required")), nil
	}

	topic, err := database.NewHotspotStorage(r.db).Get(topicID)
	if err != nil {
		return ErrorResult(err), nil
	}

	if topic == nil {
		return ErrorResult(fmt.Errorf("topic not found")), nil
	}

	// 获取历史
	history, _ := database.NewHotspotStorage(r.db).GetRankHistory(topicID, 10)

	result := map[string]interface{}{
		"success": true,
		"topic":   topic,
		"history": history,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleRefreshHotspots(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.hotspotService == nil {
		return ErrorResult(fmt.Errorf("hotspot service not configured")), nil
	}

	source, _ := args["source"].(string)

	var count int

	if source != "" {
		topics, err := r.hotspotService.FetchAndSave(ctx, source, 0)
		if err != nil {
			return ErrorResult(err), nil
		}
		count = len(topics)
	} else {
		results, err := r.hotspotService.FetchFromAllSources(ctx, 0)
		if err != nil {
			return ErrorResult(err), nil
		}
		for _, topics := range results {
			count += len(topics)
		}
	}

	result := map[string]interface{}{
		"success": true,
		"count":   count,
		"message": fmt.Sprintf("刷新完成，共获取 %d 条热点", count),
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetHotspotStats(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.hotspotService == nil {
		return ErrorResult(fmt.Errorf("hotspot service not configured")), nil
	}

	stats, err := r.hotspotService.GetStats()
	if err != nil {
		return ErrorResult(err), nil
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	return SuccessResult(string(data)), nil
}

// =====================================================
// 视频处理工具
// =====================================================

func (r *ToolsRegistry) registerVideoTools(server *Server) {
	// 提交视频处理任务
	server.RegisterTool(Tool{
		Name:        "submit_video_task",
		Description: "提交视频处理任务",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "视频URL",
				},
				"transcribe": map[string]interface{}{
					"type":        "boolean",
					"description": "是否转录",
					"default":     true,
				},
				"optimize": map[string]interface{}{
					"type":        "boolean",
					"description": "是否优化文本",
					"default":     true,
				},
			},
			"required": []string{"url"},
		},
		Handler: r.handleSubmitVideoTask,
	})

	// 获取视频任务状态
	server.RegisterTool(Tool{
		Name:        "get_video_task_status",
		Description: "获取视频处理任务状态",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "任务ID",
				},
			},
			"required": []string{"task_id"},
		},
		Handler: r.handleGetVideoTaskStatus,
	})

	// 获取视频列表
	server.RegisterTool(Tool{
		Name:        "list_videos",
		Description: "获取视频列表",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"description": "状态过滤（pending/completed/failed）",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "返回数量",
					"default":     20,
				},
			},
		},
		Handler: r.handleListVideos,
	})

	// 获取视频详情
	server.RegisterTool(Tool{
		Name:        "get_video_detail",
		Description: "获取视频详情和转录内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"video_id": map[string]interface{}{
					"type":        "string",
					"description": "视频ID",
				},
			},
			"required": []string{"video_id"},
		},
		Handler: r.handleGetVideoDetail,
	})
}

func (r *ToolsRegistry) handleSubmitVideoTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.videoService == nil {
		return ErrorResult(fmt.Errorf("video service not configured")), nil
	}

	url, ok := args["url"].(string)
	if !ok {
		return ErrorResult(fmt.Errorf("url is required")), nil
	}

	opts := video.DefaultProcessOptions()
	if transcribe, ok := args["transcribe"].(bool); ok {
		opts.Transcribe = transcribe
	}
	if optimize, ok := args["optimize"].(bool); ok {
		opts.Optimize = optimize
	}

	task, err := r.videoService.SubmitTask(url, opts)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"task_id": task.ID,
		"status":  task.Status,
		"message": "视频处理任务已提交",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetVideoTaskStatus(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.videoService == nil {
		return ErrorResult(fmt.Errorf("video service not configured")), nil
	}

	taskID, ok := args["task_id"].(string)
	if !ok {
		return ErrorResult(fmt.Errorf("task_id is required")), nil
	}

	task, err := r.videoService.GetTask(taskID)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success":  true,
		"task":     task,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleListVideos(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.videoService == nil {
		return ErrorResult(fmt.Errorf("video service not configured")), nil
	}

	status, _ := args["status"].(string)
	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	videos, total, err := r.videoService.ListVideos(status, limit, 0)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"total":   total,
		"videos":  videos,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetVideoDetail(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.videoService == nil {
		return ErrorResult(fmt.Errorf("video service not configured")), nil
	}

	videoID, ok := args["video_id"].(string)
	if !ok {
		return ErrorResult(fmt.Errorf("video_id is required")), nil
	}

	video, err := r.videoService.GetVideo(videoID)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"video":   video,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

// =====================================================
// 通知工具
// =====================================================

func (r *ToolsRegistry) registerNotifyTools(server *Server) {
	// 发送通知
	server.RegisterTool(Tool{
		Name:        "send_notification",
		Description: "发送通知消息",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"channel": map[string]interface{}{
					"type":        "string",
					"description": "通知渠道（feishu/dingtalk/wecom/telegram）",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "通知标题",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "通知内容",
				},
			},
			"required": []string{"channel", "title", "content"},
		},
		Handler: r.handleSendNotification,
	})

	// 发送热点通知
	server.RegisterTool(Tool{
		Name:        "notify_hot_topics",
		Description: "发送热点话题通知",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"channel": map[string]interface{}{
					"type":        "string",
					"description": "通知渠道",
				},
				"threshold": map[string]interface{}{
					"type":        "integer",
					"description": "热度阈值",
					"default":     80,
				},
			},
			"required": []string{"channel"},
		},
		Handler: r.handleNotifyHotTopics,
	})
}

func (r *ToolsRegistry) handleSendNotification(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.notifyService == nil {
		return ErrorResult(fmt.Errorf("notify service not configured")), nil
	}

	channel, _ := args["channel"].(string)
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)

	if channel == "" || title == "" || content == "" {
		return ErrorResult(fmt.Errorf("channel, title and content are required")), nil
	}

	err := r.notifyService.Send(ctx, channel, &notify.Message{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success": true,
		"message": "通知发送成功",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleNotifyHotTopics(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	if r.hotspotService == nil {
		return ErrorResult(fmt.Errorf("hotspot service not configured")), nil
	}

	threshold := 80
	if t, ok := args["threshold"].(float64); ok {
		threshold = int(t)
	}

	err := r.hotspotService.NotifyHotTopics(ctx, threshold)
	if err != nil {
		return ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"success":   true,
		"threshold": threshold,
		"message":   "热点通知发送成功",
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

// =====================================================
// AI 配置工具
// =====================================================

func (r *ToolsRegistry) registerAIConfigTools(server *Server) {
	// 列出 AI 配置
	server.RegisterTool(Tool{
		Name:        "list_ai_configs",
		Description: "列出所有 AI 服务配置",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: r.handleListAIConfigs,
	})

	// 获取 AI 统计
	server.RegisterTool(Tool{
		Name:        "get_ai_stats",
		Description: "获取 AI 调用统计",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"days": map[string]interface{}{
					"type":        "integer",
					"description": "统计天数",
					"default":     7,
				},
			},
		},
		Handler: r.handleGetAIStats,
	})
}

func (r *ToolsRegistry) handleListAIConfigs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	var configs []database.AIServiceConfig
	r.db.Where("is_active = ?", true).Order("priority desc").Find(&configs)

	result := map[string]interface{}{
		"success": true,
		"configs": configs,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

func (r *ToolsRegistry) handleGetAIStats(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	days := 7
	if d, ok := args["days"].(float64); ok {
		days = int(d)
	}

	since := fmt.Sprintf("-%d days", days)

	var stats struct {
		TotalCalls   int64
		SuccessCalls int64
		FailedCalls  int64
		TotalTokens  int64
	}

	r.db.Model(&database.AIHistory{}).Where("created_at > datetime('now', ?)", since).Count(&stats.TotalCalls)
	r.db.Model(&database.AIHistory{}).Where("created_at > datetime('now', ?) AND success = ?", since, true).Count(&stats.SuccessCalls)
	r.db.Model(&database.AIHistory{}).Where("created_at > datetime('now', ?) AND success = ?", since, false).Count(&stats.FailedCalls)
	r.db.Model(&database.AIHistory{}).Where("created_at > datetime('now', ?)", since).Select("COALESCE(SUM(tokens_used), 0)").Scan(&stats.TotalTokens)

	result := map[string]interface{}{
		"success": true,
		"days":    days,
		"stats":   stats,
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return SuccessResult(string(data)), nil
}

// =====================================================
// 系统工具
// =====================================================

func (r *ToolsRegistry) registerSystemTools(server *Server) {
	// 获取系统状态
	server.RegisterTool(Tool{
		Name:        "get_system_status",
		Description: "获取系统状态",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: r.handleGetSystemStatus,
	})
}

func (r *ToolsRegistry) handleGetSystemStatus(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	status := map[string]interface{}{
		"success": true,
		"database": "connected",
		"services": map[string]bool{
			"hotspot": r.hotspotService != nil,
			"video":   r.videoService != nil,
			"notify":  r.notifyService != nil,
		},
	}

	data, _ := json.MarshalIndent(status, "", "  ")
	return SuccessResult(string(data)), nil
}
