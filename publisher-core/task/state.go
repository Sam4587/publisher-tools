package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StateService 任务状态服务
type StateService struct {
	db          *gorm.DB
	progressMap map[string]*TaskProgress
	mu          sync.RWMutex
	subscribers map[string][]ProgressSubscriber
	subMu       sync.RWMutex
}

// ProgressSubscriber 进度订阅者
type ProgressSubscriber func(progress *TaskProgress)

// TaskProgress 任务进度
type TaskProgress struct {
	TaskID       string    `json:"task_id"`
	Progress     int       `json:"progress"`      // 0-100
	CurrentStep  string    `json:"current_step"`
	TotalSteps   int       `json:"total_steps"`
	CompletedSteps int     `json:"completed_steps"`
	Message      string    `json:"message"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewStateService 创建状态服务
func NewStateService(db *gorm.DB) *StateService {
	return &StateService{
		db:          db,
		progressMap: make(map[string]*TaskProgress),
		subscribers: make(map[string][]ProgressSubscriber),
	}
}

// UpdateProgress 更新任务进度
func (s *StateService) UpdateProgress(taskID string, progress int, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新内存中的进度
	p := &TaskProgress{
		TaskID:    taskID,
		Progress:  progress,
		Message:   message,
		UpdatedAt: time.Now(),
	}
	s.progressMap[taskID] = p

	// 更新数据库
	if err := s.db.Model(&database.AsyncTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"progress":      progress,
			"progress_text": message,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		logrus.Errorf("更新任务进度失败: %v", err)
		return err
	}

	// 通知订阅者
	s.notifySubscribers(taskID, p)

	logrus.Debugf("任务进度更新: %s, %d%%, %s", taskID, progress, message)
	return nil
}

// UpdateProgressWithSteps 更新任务进度（带步骤）
func (s *StateService) UpdateProgressWithSteps(taskID string, currentStep string, completedSteps, totalSteps int, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	progress := 0
	if totalSteps > 0 {
		progress = (completedSteps * 100) / totalSteps
	}

	p := &TaskProgress{
		TaskID:         taskID,
		Progress:       progress,
		CurrentStep:    currentStep,
		TotalSteps:     totalSteps,
		CompletedSteps: completedSteps,
		Message:        message,
		UpdatedAt:      time.Now(),
	}
	s.progressMap[taskID] = p

	// 更新数据库
	if err := s.db.Model(&database.AsyncTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"progress":      progress,
			"progress_text": message,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		logrus.Errorf("更新任务进度失败: %v", err)
		return err
	}

	// 通知订阅者
	s.notifySubscribers(taskID, p)

	return nil
}

// GetProgress 获取任务进度
func (s *StateService) GetProgress(taskID string) (*TaskProgress, error) {
	s.mu.RLock()
	if p, exists := s.progressMap[taskID]; exists {
		s.mu.RUnlock()
		return p, nil
	}
	s.mu.RUnlock()

	// 从数据库加载
	var task database.AsyncTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	p := &TaskProgress{
		TaskID:    taskID,
		Progress:  task.Progress,
		Message:   task.ProgressText,
		UpdatedAt: task.UpdatedAt,
	}

	s.mu.Lock()
	s.progressMap[taskID] = p
	s.mu.Unlock()

	return p, nil
}

// SubscribeProgress 订阅进度更新
func (s *StateService) SubscribeProgress(taskID string, subscriber ProgressSubscriber) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	s.subscribers[taskID] = append(s.subscribers[taskID], subscriber)
}

// UnsubscribeProgress 取消订阅
func (s *StateService) UnsubscribeProgress(taskID string) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	delete(s.subscribers, taskID)
}

// notifySubscribers 通知订阅者
func (s *StateService) notifySubscribers(taskID string, progress *TaskProgress) {
	s.subMu.RLock()
	subscribers := s.subscribers[taskID]
	s.subMu.RUnlock()

	for _, sub := range subscribers {
		go func(fn ProgressSubscriber) {
			defer func() {
				if r := recover(); r != nil {
					logrus.Errorf("订阅者回调异常: %v", r)
				}
			}()
			fn(progress)
		}(sub)
	}
}

// SetTaskResult 设置任务结果
func (s *StateService) SetTaskResult(taskID string, result interface{}) error {
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}

	now := time.Now()
	if err := s.db.Model(&database.AsyncTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"result":      string(resultBytes),
			"status":      database.TaskStatusCompleted,
			"completed_at": now,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}

	logrus.Infof("任务结果已保存: %s", taskID)
	return nil
}

// GetTaskResult 获取任务结果
func (s *StateService) GetTaskResult(taskID string, result interface{}) error {
	var task database.AsyncTask
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return err
	}

	if task.Status != database.TaskStatusCompleted {
		return fmt.Errorf("任务未完成")
	}

	if task.Result == "" {
		return fmt.Errorf("任务结果为空")
	}

	return json.Unmarshal([]byte(task.Result), result)
}

// ListTasks 列出任务
func (s *StateService) ListTasks(filter *TaskFilter) ([]database.AsyncTask, int64, error) {
	query := s.db.Model(&database.AsyncTask{})

	if filter.QueueName != "" {
		query = query.Where("queue_name = ?", filter.QueueName)
	}
	if filter.TaskType != "" {
		query = query.Where("task_type = ?", filter.TaskType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.ProjectID != "" {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var tasks []database.AsyncTask
	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	QueueName string `form:"queue_name"`
	TaskType  string `form:"task_type"`
	Status    string `form:"status"`
	UserID    string `form:"user_id"`
	ProjectID string `form:"project_id"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

// GetTaskStatistics 获取任务统计
func (s *StateService) GetTaskStatistics(startTime, endTime time.Time) (*TaskStatistics, error) {
	stats := &TaskStatistics{}

	// 总任务数
	if err := s.db.Model(&database.AsyncTask{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 各状态统计
	rows, err := s.db.Model(&database.AsyncTask{}).
		Select("status, count(*) as count").
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
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
		case database.TaskStatusCancelled:
			stats.Cancelled = count
		}
	}

	// 平均执行时间
	var avgDuration float64
	if err := s.db.Model(&database.TaskExecution{}).
		Select("COALESCE(AVG(duration_ms), 0)").
		Where("started_at BETWEEN ? AND ?", startTime, endTime).
		Scan(&avgDuration).Error; err != nil {
		return nil, err
	}
	stats.AvgDurationMs = avgDuration

	// 成功率
	if stats.Total > 0 {
		stats.SuccessRate = float64(stats.Completed) / float64(stats.Total) * 100
	}

	return stats, nil
}

// TaskStatistics 任务统计
type TaskStatistics struct {
	Total         int64   `json:"total"`
	Pending       int64   `json:"pending"`
	Running       int64   `json:"running"`
	Completed     int64   `json:"completed"`
	Failed        int64   `json:"failed"`
	Cancelled     int64   `json:"cancelled"`
	SuccessRate   float64 `json:"success_rate"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
}

// RetryManager 重试管理器
type RetryManager struct {
	db *gorm.DB
}

// NewRetryManager 创建重试管理器
func NewRetryManager(db *gorm.DB) *RetryManager {
	return &RetryManager{db: db}
}

// GetRetryableTasks 获取可重试的任务
func (m *RetryManager) GetRetryableTasks() ([]database.AsyncTask, error) {
	var tasks []database.AsyncTask
	err := m.db.Where("status = ? AND retry_count < max_retries", database.TaskStatusFailed).
		Order("created_at ASC").
		Limit(100).
		Find(&tasks).Error

	return tasks, err
}

// RetryTask 重试任务
func (m *RetryManager) RetryTask(taskID string) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		var task database.AsyncTask
		if err := tx.Where("task_id = ?", taskID).First(&task).Error; err != nil {
			return err
		}

		if task.Status != database.TaskStatusFailed {
			return fmt.Errorf("任务状态不允许重试")
		}

		if task.RetryCount >= task.MaxRetries {
			return fmt.Errorf("已达到最大重试次数")
		}

		// 重置任务状态
		task.Status = database.TaskStatusPending
		task.Error = ""
		task.UpdatedAt = time.Now()

		return tx.Save(&task).Error
	})
}

// CleanupOldTasks 清理旧任务
func (m *RetryManager) CleanupOldTasks(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	result := m.db.Where("status IN ? AND created_at < ?",
		[]database.TaskStatus{
			database.TaskStatusCompleted,
			database.TaskStatusFailed,
			database.TaskStatusCancelled,
		},
		cutoff,
	).Delete(&database.AsyncTask{})

	return result.RowsAffected, result.Error
}
