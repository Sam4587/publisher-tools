package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"publisher-core/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// QueueService 任务队列服务
type QueueService struct {
	db       *gorm.DB
	queues   map[string]*TaskQueue
	handlers map[string]QueueTaskHandler
	mu       sync.RWMutex
	config   *QueueConfig
}

// QueueTaskHandler 队列任务处理函数
type QueueTaskHandler func(ctx context.Context, task *database.AsyncTask) error

// TaskQueue 任务队列
type TaskQueue struct {
	Name        string
	Tasks       chan *database.AsyncTask
	Concurrency int
	Workers     int
	IsActive    bool
}

// QueueConfig 队列配置
type QueueConfig struct {
	DefaultConcurrency int           `json:"default_concurrency"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	DefaultMaxRetries  int           `json:"default_max_retries"`
	PollInterval       time.Duration `json:"poll_interval"`
}

// DefaultQueueConfig 默认配置
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		DefaultConcurrency: 5,
		DefaultTimeout:     30 * time.Minute,
		DefaultMaxRetries:  3,
		PollInterval:       1 * time.Second,
	}
}

// NewQueueService 创建队列服务
func NewQueueService(db *gorm.DB, config *QueueConfig) *QueueService {
	if config == nil {
		config = DefaultQueueConfig()
	}
	return &QueueService{
		db:       db,
		queues:   make(map[string]*TaskQueue),
		handlers: make(map[string]QueueTaskHandler),
		config:   config,
	}
}

// RegisterQueue 注册队列
func (s *QueueService) RegisterQueue(name string, concurrency int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.queues[name]; exists {
		return fmt.Errorf("队列 %s 已存在", name)
	}

	queue := &TaskQueue{
		Name:        name,
		Tasks:       make(chan *database.AsyncTask, 1000),
		Concurrency: concurrency,
		IsActive:    true,
	}

	s.queues[name] = queue

	// 保存到数据库
	dbQueue := &database.TaskQueue{
		Name:        name,
		Concurrency: concurrency,
		MaxSize:     1000,
		Priority:    5,
		IsActive:    true,
		Timeout:     int(s.config.DefaultTimeout.Seconds()),
	}

	if err := s.db.Create(dbQueue).Error; err != nil {
		logrus.Errorf("保存队列配置失败: %v", err)
	}

	logrus.Infof("注册队列: %s, 并发数: %d", name, concurrency)
	return nil
}

// RegisterHandler 注册任务处理器
func (s *QueueService) RegisterHandler(taskType string, handler QueueTaskHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[taskType] = handler
	logrus.Infof("注册任务处理器: %s", taskType)
}

// SubmitTask 提交任务
func (s *QueueService) SubmitTask(ctx context.Context, req *TaskRequest) (*database.AsyncTask, error) {
	// 生成任务ID
	taskID := uuid.New().String()

	// 序列化payload
	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("序列化任务数据失败: %w", err)
	}

	// 创建任务记录
	task := &database.AsyncTask{
		TaskID:     taskID,
		TaskType:   req.TaskType,
		QueueName:  req.QueueName,
		Status:     database.TaskStatusPending,
		Priority:   req.Priority,
		Payload:    string(payloadBytes),
		MaxRetries: req.MaxRetries,
		Timeout:    req.Timeout,
		UserID:     req.UserID,
		ProjectID:  req.ProjectID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if task.MaxRetries == 0 {
		task.MaxRetries = s.config.DefaultMaxRetries
	}

	if task.Timeout == 0 {
		task.Timeout = int(s.config.DefaultTimeout.Seconds())
	}

	// 保存到数据库
	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 添加到内存队列
	s.mu.RLock()
	queue, exists := s.queues[req.QueueName]
	s.mu.RUnlock()

	if !exists {
		// 自动创建队列
		if err := s.RegisterQueue(req.QueueName, s.config.DefaultConcurrency); err != nil {
			return nil, err
		}
		queue = s.queues[req.QueueName]
	}

	select {
	case queue.Tasks <- task:
		logrus.Infof("任务已提交: %s, 类型: %s, 队列: %s", taskID, req.TaskType, req.QueueName)
		return task, nil
	default:
		// 队列已满，任务仍在数据库中等待
		logrus.Warnf("队列 %s 已满，任务 %s 将在下次轮询时处理", req.QueueName, taskID)
		return task, nil
	}
}

// TaskRequest 任务请求
type TaskRequest struct {
	TaskType   string                 `json:"task_type"`
	QueueName  string                 `json:"queue_name"`
	Priority   database.TaskPriority  `json:"priority"`
	Payload    map[string]interface{} `json:"payload"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    int                    `json:"timeout"`
	UserID     string                 `json:"user_id"`
	ProjectID  string                 `json:"project_id"`
}

// Start 启动队列服务
func (s *QueueService) Start(ctx context.Context) {
	logrus.Info("任务队列服务已启动")

	// 启动队列工作器
	s.mu.RLock()
	for name, queue := range s.queues {
		for i := 0; i < queue.Concurrency; i++ {
			go s.worker(ctx, name, i)
		}
	}
	s.mu.RUnlock()

	// 启动轮询器
	go s.poller(ctx)
}

// worker 工作器
func (s *QueueService) worker(ctx context.Context, queueName string, workerID int) {
	logrus.Infof("工作器启动: 队列=%s, ID=%d", queueName, workerID)

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("工作器停止: 队列=%s, ID=%d", queueName, workerID)
			return
		case task := <-s.queues[queueName].Tasks:
			s.processTask(ctx, task)
		}
	}
}

// poller 轮询器
func (s *QueueService) poller(ctx context.Context) {
	ticker := time.NewTicker(s.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.pollPendingTasks()
		}
	}
}

// pollPendingTasks 轮询待处理任务
func (s *QueueService) pollPendingTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for queueName, queue := range s.queues {
		// 检查队列是否有空闲位置
		if len(queue.Tasks) >= cap(queue.Tasks) {
			continue
		}

		// 从数据库获取待处理任务
		var tasks []database.AsyncTask
		err := s.db.Where("queue_name = ? AND status = ?", queueName, database.TaskStatusPending).
			Order("priority DESC, created_at ASC").
			Limit(10).
			Find(&tasks).Error

		if err != nil {
			logrus.Errorf("查询待处理任务失败: %v", err)
			continue
		}

		for _, task := range tasks {
			select {
			case queue.Tasks <- &task:
				// 成功添加到队列
			default:
				// 队列已满
				return
			}
		}
	}
}

// processTask 处理任务
func (s *QueueService) processTask(ctx context.Context, task *database.AsyncTask) {
	// 更新任务状态为运行中
	now := time.Now()
	task.Status = database.TaskStatusRunning
	task.StartedAt = &now
	task.UpdatedAt = now

	if err := s.db.Save(task).Error; err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
		return
	}

	// 获取处理器
	s.mu.RLock()
	handler, exists := s.handlers[task.TaskType]
	s.mu.RUnlock()

	if !exists {
		s.handleTaskError(task, fmt.Errorf("未找到任务处理器: %s", task.TaskType))
		return
	}

	// 执行任务
	startTime := time.Now()
	err := handler(ctx, task)
	duration := time.Since(startTime)

	// 记录执行结果
	execution := &database.TaskExecution{
		TaskID:      task.TaskID,
		WorkerID:    fmt.Sprintf("worker-%d", time.Now().UnixNano()),
		Status:      "success",
		DurationMs:  int(duration.Milliseconds()),
		StartedAt:   *task.StartedAt,
		CompletedAt: time.Now(),
	}

	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		s.handleTaskError(task, err)
	} else {
		s.handleTaskSuccess(task)
	}

	// 保存执行记录
	if err := s.db.Create(execution).Error; err != nil {
		logrus.Errorf("保存执行记录失败: %v", err)
	}
}

// handleTaskSuccess 处理任务成功
func (s *QueueService) handleTaskSuccess(task *database.AsyncTask) {
	now := time.Now()
	task.Status = database.TaskStatusCompleted
	task.CompletedAt = &now
	task.UpdatedAt = now

	if err := s.db.Save(task).Error; err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
	}

	logrus.Infof("任务完成: %s", task.TaskID)
}

// handleTaskError 处理任务错误
func (s *QueueService) handleTaskError(task *database.AsyncTask, err error) {
	task.Error = err.Error()
	task.RetryCount++
	task.UpdatedAt = time.Now()

	// 检查是否需要重试
	if task.RetryCount < task.MaxRetries {
		task.Status = database.TaskStatusRetrying
		logrus.Warnf("任务将重试: %s, 重试次数: %d/%d", task.TaskID, task.RetryCount, task.MaxRetries)

		// 计算重试延迟（指数退避）
		delay := time.Duration(task.RetryCount*task.RetryCount) * time.Second
		scheduledAt := time.Now().Add(delay)
		task.ScheduledAt = &scheduledAt
		task.Status = database.TaskStatusPending
	} else {
		task.Status = database.TaskStatusFailed
		logrus.Errorf("任务失败: %s, 错误: %v", task.TaskID, err)
	}

	if err := s.db.Save(task).Error; err != nil {
		logrus.Errorf("更新任务状态失败: %v", err)
	}
}

// GetTask 获取任务
func (s *QueueService) GetTask(taskID string) (*database.AsyncTask, error) {
	var task database.AsyncTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CancelTask 取消任务
func (s *QueueService) CancelTask(taskID string) error {
	result := s.db.Model(&database.AsyncTask{}).
		Where("task_id = ? AND status IN ?", taskID, []database.TaskStatus{
			database.TaskStatusPending,
			database.TaskStatusRetrying,
		}).
		Update("status", database.TaskStatusCancelled)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("任务不存在或无法取消")
	}

	logrus.Infof("任务已取消: %s", taskID)
	return nil
}

// GetQueueStats 获取队列统计
func (s *QueueService) GetQueueStats(queueName string) (*QueueStats, error) {
	var stats QueueStats

	// 统计各状态任务数
	rows, err := s.db.Model(&database.AsyncTask{}).
		Select("status, count(*) as count").
		Where("queue_name = ?", queueName).
		Group("status").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}

		switch database.TaskStatus(status) {
		case database.TaskStatusPending:
			stats.Pending = count
		case database.TaskStatusRunning:
			stats.Running = count
		case database.TaskStatusCompleted:
			stats.Completed = count
		case database.TaskStatusFailed:
			stats.Failed = count
		}
	}

	s.mu.RLock()
	if queue, exists := s.queues[queueName]; exists {
		stats.QueueSize = len(queue.Tasks)
		stats.QueueCapacity = cap(queue.Tasks)
		stats.Concurrency = queue.Concurrency
	}
	s.mu.RUnlock()

	return &stats, nil
}

// QueueStats 队列统计
type QueueStats struct {
	Pending      int64 `json:"pending"`
	Running      int64 `json:"running"`
	Completed    int64 `json:"completed"`
	Failed       int64 `json:"failed"`
	QueueSize    int   `json:"queue_size"`
	QueueCapacity int  `json:"queue_capacity"`
	Concurrency  int   `json:"concurrency"`
}
