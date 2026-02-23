package api

import (
	"net/http"
	"time"

	"publisher-core/database"
	"publisher-core/task"

	"github.com/gin-gonic/gin"
)

// TaskAPI 任务API
type TaskAPI struct {
	queueService *task.QueueService
	stateService *task.StateService
	scheduler    *task.SchedulerService
}

// NewTaskAPI 创建任务API
func NewTaskAPI(queueService *task.QueueService, stateService *task.StateService, scheduler *task.SchedulerService) *TaskAPI {
	return &TaskAPI{
		queueService: queueService,
		stateService: stateService,
		scheduler:    scheduler,
	}
}

// RegisterRoutes 注册路由
func (api *TaskAPI) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		// 任务管理
		tasks.POST("", api.SubmitTask)
		tasks.GET("/:task_id", api.GetTask)
		tasks.DELETE("/:task_id", api.CancelTask)
		tasks.GET("", api.ListTasks)

		// 进度管理
		tasks.GET("/:task_id/progress", api.GetProgress)
		tasks.GET("/:task_id/result", api.GetResult)

		// 统计
		tasks.GET("/statistics", api.GetStatistics)
	}

	queues := r.Group("/queues")
	{
		queues.GET("/:queue_name/stats", api.GetQueueStats)
		queues.POST("", api.CreateQueue)
	}

	scheduled := r.Group("/scheduled-tasks")
	{
		scheduled.POST("", api.CreateScheduledTask)
		scheduled.GET("", api.ListScheduledTasks)
		scheduled.GET("/:name", api.GetScheduledTask)
		scheduled.PUT("/:name", api.UpdateScheduledTask)
		scheduled.DELETE("/:name", api.DeleteScheduledTask)
		scheduled.POST("/:name/pause", api.PauseScheduledTask)
		scheduled.POST("/:name/resume", api.ResumeScheduledTask)
		scheduled.POST("/:name/run", api.RunScheduledTaskNow)
		scheduled.GET("/stats", api.GetSchedulerStats)
	}
}

// SubmitTask 提交任务
func (api *TaskAPI) SubmitTask(c *gin.Context) {
	var req task.TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.QueueName == "" {
		req.QueueName = "default"
	}
	if req.Priority == 0 {
		req.Priority = database.PriorityNormal
	}

	task, err := api.queueService.SubmitTask(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "任务已提交",
		"task":    task,
	})
}

// GetTask 获取任务
func (api *TaskAPI) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := api.queueService.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CancelTask 取消任务
func (api *TaskAPI) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")

	if err := api.queueService.CancelTask(taskID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已取消"})
}

// ListTasks 列出任务
func (api *TaskAPI) ListTasks(c *gin.Context) {
	var filter task.StateTaskFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认分页
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}

	tasks, total, err := api.stateService.ListTasks(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": total,
		"page":  filter.Page,
		"page_size": filter.PageSize,
	})
}

// GetProgress 获取任务进度
func (api *TaskAPI) GetProgress(c *gin.Context) {
	taskID := c.Param("task_id")

	progress, err := api.stateService.GetProgress(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetResult 获取任务结果
func (api *TaskAPI) GetResult(c *gin.Context) {
	taskID := c.Param("task_id")

	var result map[string]interface{}
	if err := api.stateService.GetTaskResult(taskID, &result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStatistics 获取任务统计
func (api *TaskAPI) GetStatistics(c *gin.Context) {
	// 解析时间范围
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	if start := c.Query("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			startTime = t
		}
	}
	if end := c.Query("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			endTime = t
		}
	}

	stats, err := api.stateService.GetTaskStatistics(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetQueueStats 获取队列统计
func (api *TaskAPI) GetQueueStats(c *gin.Context) {
	queueName := c.Param("queue_name")

	stats, err := api.queueService.GetQueueStats(queueName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateQueue 创建队列
func (api *TaskAPI) CreateQueue(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Concurrency int    `json:"concurrency"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Concurrency == 0 {
		req.Concurrency = 5
	}

	if err := api.queueService.RegisterQueue(req.Name, req.Concurrency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "队列已创建",
		"name":    req.Name,
	})
}

// CreateScheduledTask 创建定时任务
func (api *TaskAPI) CreateScheduledTask(c *gin.Context) {
	var req task.ScheduledTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := api.scheduler.CreateScheduledTask(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// ListScheduledTasks 列出定时任务
func (api *TaskAPI) ListScheduledTasks(c *gin.Context) {
	tasks, err := api.scheduler.ListScheduledTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetScheduledTask 获取定时任务
func (api *TaskAPI) GetScheduledTask(c *gin.Context) {
	name := c.Param("name")

	task, err := api.scheduler.GetScheduledTask(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "定时任务不存在"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateScheduledTask 更新定时任务
func (api *TaskAPI) UpdateScheduledTask(c *gin.Context) {
	name := c.Param("name")

	var req task.ScheduledTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := api.scheduler.UpdateScheduledTask(name, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "定时任务已更新"})
}

// DeleteScheduledTask 删除定时任务
func (api *TaskAPI) DeleteScheduledTask(c *gin.Context) {
	name := c.Param("name")

	if err := api.scheduler.DeleteScheduledTask(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "定时任务已删除"})
}

// PauseScheduledTask 暂停定时任务
func (api *TaskAPI) PauseScheduledTask(c *gin.Context) {
	name := c.Param("name")

	if err := api.scheduler.PauseScheduledTask(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "定时任务已暂停"})
}

// ResumeScheduledTask 恢复定时任务
func (api *TaskAPI) ResumeScheduledTask(c *gin.Context) {
	name := c.Param("name")

	if err := api.scheduler.ResumeScheduledTask(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "定时任务已恢复"})
}

// RunScheduledTaskNow 立即执行定时任务
func (api *TaskAPI) RunScheduledTaskNow(c *gin.Context) {
	name := c.Param("name")

	if err := api.scheduler.RunScheduledTaskNow(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "定时任务已触发执行"})
}

// GetSchedulerStats 获取调度器统计
func (api *TaskAPI) GetSchedulerStats(c *gin.Context) {
	stats := api.scheduler.GetSchedulerStats()
	c.JSON(http.StatusOK, stats)
}
