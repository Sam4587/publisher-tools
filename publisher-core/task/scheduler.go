package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"publisher-core/database"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SchedulerService 定时任务调度服务
type SchedulerService struct {
	db          *gorm.DB
	queueService *QueueService
	cron        *cron.Cron
	jobs        map[string]cron.EntryID
	mu          sync.RWMutex
}

// NewSchedulerService 创建调度服务
func NewSchedulerService(db *gorm.DB, queueService *QueueService) *SchedulerService {
	return &SchedulerService{
		db:          db,
		queueService: queueService,
		cron:        cron.New(cron.WithSeconds()),
		jobs:        make(map[string]cron.EntryID),
	}
}

// Start 启动调度服务
func (s *SchedulerService) Start(ctx context.Context) error {
	// 加载所有激活的定时任务
	var scheduledTasks []database.ScheduledTask
	if err := s.db.Where("is_active = ?", true).Find(&scheduledTasks).Error; err != nil {
		return fmt.Errorf("加载定时任务失败: %w", err)
	}

	for _, task := range scheduledTasks {
		if err := s.scheduleTask(&task); err != nil {
			logrus.Errorf("调度任务失败: %s, 错误: %v", task.Name, err)
		}
	}

	s.cron.Start()
	logrus.Infof("定时任务调度服务已启动，已加载 %d 个任务", len(scheduledTasks))

	// 监听上下文取消
	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	return nil
}

// Stop 停止调度服务
func (s *SchedulerService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logrus.Info("定时任务调度服务已停止")
}

// scheduleTask 调度任务
func (s *SchedulerService) scheduleTask(task *database.ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已存在，先移除
	if entryID, exists := s.jobs[task.Name]; exists {
		s.cron.Remove(entryID)
	}

	// 添加新任务
	entryID, err := s.cron.AddFunc(task.CronExpr, func() {
		s.executeScheduledTask(task)
	})

	if err != nil {
		return fmt.Errorf("添加定时任务失败: %w", err)
	}

	s.jobs[task.Name] = entryID
	logrus.Infof("定时任务已调度: %s, Cron: %s", task.Name, task.CronExpr)

	return nil
}

// executeScheduledTask 执行定时任务
func (s *SchedulerService) executeScheduledTask(task *database.ScheduledTask) {
	logrus.Infof("执行定时任务: %s", task.Name)

	// 解析payload
	var payload map[string]interface{}
	if task.Payload != "" {
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			logrus.Errorf("解析任务payload失败: %v", err)
			return
		}
	}

	// 提交任务到队列
	taskReq := &TaskRequest{
		TaskType:  task.TaskType,
		QueueName: task.QueueName,
		Priority:  database.PriorityNormal,
		Payload:   payload,
	}

	_, err := s.queueService.SubmitTask(context.Background(), taskReq)
	if err != nil {
		logrus.Errorf("提交定时任务失败: %v", err)

		// 更新错误信息
		now := time.Now()
		s.db.Model(&database.ScheduledTask{}).
			Where("id = ?", task.ID).
			Updates(map[string]interface{}{
				"last_error": err.Error(),
				"last_run_at": now,
			})
		return
	}

	// 更新执行记录
	now := time.Now()
	nextRun, _ := s.getNextRunTime(task.CronExpr)
	s.db.Model(&database.ScheduledTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]interface{}{
			"last_run_at": now,
			"next_run_at": nextRun,
			"run_count":   gorm.Expr("run_count + 1"),
		})
}

// getNextRunTime 获取下次执行时间
func (s *SchedulerService) getNextRunTime(cronExpr string) (*time.Time, error) {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return nil, err
	}

	next := schedule.Next(time.Now())
	return &next, nil
}

// CreateScheduledTask 创建定时任务
func (s *SchedulerService) CreateScheduledTask(req *ScheduledTaskRequest) (*database.ScheduledTask, error) {
	// 验证Cron表达式
	if _, err := cron.ParseStandard(req.CronExpr); err != nil {
		return nil, fmt.Errorf("无效的Cron表达式: %w", err)
	}

	// 序列化payload
	payloadBytes, _ := json.Marshal(req.Payload)

	task := &database.ScheduledTask{
		Name:      req.Name,
		TaskType:  req.TaskType,
		CronExpr:  req.CronExpr,
		Payload:   string(payloadBytes),
		QueueName: req.QueueName,
		IsActive:  true,
	}

	// 计算下次执行时间
	nextRun, _ := s.getNextRunTime(req.CronExpr)
	task.NextRunAt = nextRun

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建定时任务失败: %w", err)
	}

	// 调度任务
	if err := s.scheduleTask(task); err != nil {
		return nil, err
	}

	logrus.Infof("定时任务已创建: %s", task.Name)
	return task, nil
}

// ScheduledTaskRequest 定时任务请求
type ScheduledTaskRequest struct {
	Name      string                 `json:"name"`
	TaskType  string                 `json:"task_type"`
	CronExpr  string                 `json:"cron_expr"`
	Payload   map[string]interface{} `json:"payload"`
	QueueName string                 `json:"queue_name"`
}

// UpdateScheduledTask 更新定时任务
func (s *SchedulerService) UpdateScheduledTask(name string, req *ScheduledTaskRequest) error {
	var task database.ScheduledTask
	if err := s.db.Where("name = ?", name).First(&task).Error; err != nil {
		return err
	}

	// 验证Cron表达式
	if _, err := cron.ParseStandard(req.CronExpr); err != nil {
		return fmt.Errorf("无效的Cron表达式: %w", err)
	}

	// 更新任务
	payloadBytes, _ := json.Marshal(req.Payload)
	task.TaskType = req.TaskType
	task.CronExpr = req.CronExpr
	task.Payload = string(payloadBytes)
	task.QueueName = req.QueueName

	nextRun, _ := s.getNextRunTime(req.CronExpr)
	task.NextRunAt = nextRun

	if err := s.db.Save(&task).Error; err != nil {
		return err
	}

	// 重新调度
	return s.scheduleTask(&task)
}

// DeleteScheduledTask 删除定时任务
func (s *SchedulerService) DeleteScheduledTask(name string) error {
	s.mu.Lock()
	if entryID, exists := s.jobs[name]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, name)
	}
	s.mu.Unlock()

	result := s.db.Where("name = ?", name).Delete(&database.ScheduledTask{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("定时任务不存在")
	}

	logrus.Infof("定时任务已删除: %s", name)
	return nil
}

// PauseScheduledTask 暂停定时任务
func (s *SchedulerService) PauseScheduledTask(name string) error {
	s.mu.Lock()
	if entryID, exists := s.jobs[name]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, name)
	}
	s.mu.Unlock()

	return s.db.Model(&database.ScheduledTask{}).
		Where("name = ?", name).
		Update("is_active", false).Error
}

// ResumeScheduledTask 恢复定时任务
func (s *SchedulerService) ResumeScheduledTask(name string) error {
	var task database.ScheduledTask
	if err := s.db.Where("name = ?", name).First(&task).Error; err != nil {
		return err
	}

	task.IsActive = true
	nextRun, _ := s.getNextRunTime(task.CronExpr)
	task.NextRunAt = nextRun

	if err := s.db.Save(&task).Error; err != nil {
		return err
	}

	return s.scheduleTask(&task)
}

// GetScheduledTask 获取定时任务
func (s *SchedulerService) GetScheduledTask(name string) (*database.ScheduledTask, error) {
	var task database.ScheduledTask
	if err := s.db.Where("name = ?", name).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListScheduledTasks 列出定时任务
func (s *SchedulerService) ListScheduledTasks() ([]database.ScheduledTask, error) {
	var tasks []database.ScheduledTask
	err := s.db.Order("name ASC").Find(&tasks).Error
	return tasks, err
}

// GetSchedulerStats 获取调度器统计
func (s *SchedulerService) GetSchedulerStats() *SchedulerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeCount, inactiveCount int64
	s.db.Model(&database.ScheduledTask{}).Where("is_active = ?", true).Count(&activeCount)
	s.db.Model(&database.ScheduledTask{}).Where("is_active = ?", false).Count(&inactiveCount)

	return &SchedulerStats{
		TotalJobs:     len(s.jobs),
		ActiveJobs:    int(activeCount),
		InactiveJobs:  int(inactiveCount),
		RunningSince:  time.Now(), // 实际应该记录启动时间
	}
}

// SchedulerStats 调度器统计
type SchedulerStats struct {
	TotalJobs    int       `json:"total_jobs"`
	ActiveJobs   int       `json:"active_jobs"`
	InactiveJobs int       `json:"inactive_jobs"`
	RunningSince time.Time `json:"running_since"`
}

// RunScheduledTaskNow 立即执行定时任务
func (s *SchedulerService) RunScheduledTaskNow(name string) error {
	var task database.ScheduledTask
	if err := s.db.Where("name = ?", name).First(&task).Error; err != nil {
		return err
	}

	go s.executeScheduledTask(&task)
	return nil
}
