package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"publisher-core/database"
	"publisher-core/task"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ProgressService 进度服务
type ProgressService struct {
	db          *gorm.DB
	hub         *Hub
	stateService *task.StateService
	history     map[string][]*ProgressMessage
	mu          sync.RWMutex
	config      *ProgressConfig
}

// ProgressConfig 进度配置
type ProgressConfig struct {
	MaxHistoryPerTask int           `json:"max_history_per_task"`
	HistoryTTL        time.Duration `json:"history_ttl"`
	BroadcastEnabled  bool          `json:"broadcast_enabled"`
}

// DefaultProgressConfig 默认配置
func DefaultProgressConfig() *ProgressConfig {
	return &ProgressConfig{
		MaxHistoryPerTask: 100,
		HistoryTTL:        24 * time.Hour,
		BroadcastEnabled:  true,
	}
}

// NewProgressService 创建进度服务
func NewProgressService(db *gorm.DB, hub *Hub, stateService *task.StateService, config *ProgressConfig) *ProgressService {
	if config == nil {
		config = DefaultProgressConfig()
	}

	service := &ProgressService{
		db:           db,
		hub:          hub,
		stateService: stateService,
		history:      make(map[string][]*ProgressMessage),
		config:       config,
	}

	// 订阅任务进度更新
	if stateService != nil {
		stateService.SubscribeProgress("*", service.handleProgressUpdate)
	}

	return service
}

// handleProgressUpdate 处理进度更新
func (s *ProgressService) handleProgressUpdate(progress *task.TaskProgress) {
	// 转换为WebSocket进度消息
	msg := &ProgressMessage{
		TaskID:         progress.TaskID,
		Progress:       progress.Progress,
		CurrentStep:    progress.CurrentStep,
		TotalSteps:     progress.TotalSteps,
		CompletedSteps: progress.CompletedSteps,
		Message:        progress.Message,
		Timestamp:      progress.UpdatedAt,
	}

	// 获取任务状态
	task, err := s.stateService.GetProgress(progress.TaskID)
	if err == nil {
		msg.Status = string(task.Status)
	}

	// 保存到内存历史
	s.saveHistory(msg)

	// 持久化到数据库
	s.saveToDatabase(msg)

	// 广播进度
	if s.config.BroadcastEnabled && s.hub != nil {
		s.hub.BroadcastProgress(progress.TaskID, msg)
	}
}

// UpdateProgress 更新进度
func (s *ProgressService) UpdateProgress(taskID string, progress int, message string) error {
	// 更新任务状态服务
	if s.stateService != nil {
		if err := s.stateService.UpdateProgress(taskID, progress, message); err != nil {
			return err
		}
	}

	return nil
}

// UpdateProgressWithSteps 更新进度（带步骤）
func (s *ProgressService) UpdateProgressWithSteps(taskID string, currentStep string, completedSteps, totalSteps int, message string) error {
	if s.stateService != nil {
		if err := s.stateService.UpdateProgressWithSteps(taskID, currentStep, completedSteps, totalSteps, message); err != nil {
			return err
		}
	}

	return nil
}

// UpdateTaskStatus 更新任务状态
func (s *ProgressService) UpdateTaskStatus(taskID string, status string, result interface{}, errMsg string) error {
	// 更新数据库
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "completed" {
		now := time.Now()
		updates["completed_at"] = now
		if result != nil {
			resultBytes, _ := json.Marshal(result)
			updates["result"] = string(resultBytes)
		}
	}

	if errMsg != "" {
		updates["error"] = errMsg
	}

	if err := s.db.Model(&database.AsyncTask{}).Where("task_id = ?", taskID).Updates(updates).Error; err != nil {
		return err
	}

	// 广播状态更新
	if s.config.BroadcastEnabled && s.hub != nil {
		s.hub.BroadcastStatus(taskID, &StatusMessage{
			TaskID:    taskID,
			Status:    status,
			Error:     errMsg,
			Result:    result,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// GetProgress 获取进度
func (s *ProgressService) GetProgress(taskID string) (*ProgressMessage, error) {
	if s.stateService != nil {
		progress, err := s.stateService.GetProgress(taskID)
		if err != nil {
			return nil, err
		}

		return &ProgressMessage{
			TaskID:         progress.TaskID,
			Progress:       progress.Progress,
			CurrentStep:    progress.CurrentStep,
			TotalSteps:     progress.TotalSteps,
			CompletedSteps: progress.CompletedSteps,
			Message:        progress.Message,
			Timestamp:      progress.UpdatedAt,
		}, nil
	}

	return nil, nil
}

// GetHistory 获取进度历史
func (s *ProgressService) GetHistory(taskID string, limit int) []*ProgressMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.history[taskID]
	if len(history) == 0 {
		return nil
	}

	if limit > 0 && len(history) > limit {
		return history[len(history)-limit:]
	}

	return history
}

// saveHistory 保存历史
func (s *ProgressService) saveHistory(msg *ProgressMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := s.history[msg.TaskID]
	history = append(history, msg)

	// 限制历史记录数量
	if len(history) > s.config.MaxHistoryPerTask {
		history = history[len(history)-s.config.MaxHistoryPerTask:]
	}

	s.history[msg.TaskID] = history
}

// saveToDatabase 持久化到数据库
func (s *ProgressService) saveToDatabase(msg *ProgressMessage) {
	if s.db == nil {
		return
	}

	history := &database.ProgressHistory{
		TaskID:         msg.TaskID,
		Progress:       msg.Progress,
		CurrentStep:    msg.CurrentStep,
		TotalSteps:     msg.TotalSteps,
		CompletedSteps: msg.CompletedSteps,
		Message:        msg.Message,
		Status:         msg.Status,
		CreatedAt:      msg.Timestamp,
	}

	if err := s.db.Create(history).Error; err != nil {
		logrus.Errorf("保存进度历史到数据库失败: %v", err)
	}
}

// GetHistoryFromDB 从数据库获取历史
func (s *ProgressService) GetHistoryFromDB(taskID string, limit int) ([]*ProgressMessage, error) {
	if s.db == nil {
		return nil, nil
	}

	var records []database.ProgressHistory
	query := s.db.Where("task_id = ?", taskID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	// 转换为ProgressMessage
	messages := make([]*ProgressMessage, len(records))
	for i, record := range records {
		messages[i] = &ProgressMessage{
			TaskID:         record.TaskID,
			Progress:       record.Progress,
			CurrentStep:    record.CurrentStep,
			TotalSteps:     record.TotalSteps,
			CompletedSteps: record.CompletedSteps,
			Message:        record.Message,
			Status:         record.Status,
			Timestamp:      record.CreatedAt,
		}
	}

	// 反转顺序（从旧到新）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// ClearHistory 清除历史
func (s *ProgressService) ClearHistory(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.history, taskID)
}

// CleanupOldHistory 清理旧历史
func (s *ProgressService) CleanupOldHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.config.HistoryTTL)

	for taskID, history := range s.history {
		// 过滤掉过期的记录
		var newHistory []*ProgressMessage
		for _, msg := range history {
			if msg.Timestamp.After(cutoff) {
				newHistory = append(newHistory, msg)
			}
		}

		if len(newHistory) == 0 {
			delete(s.history, taskID)
		} else {
			s.history[taskID] = newHistory
		}
	}
}

// BroadcastToUser 向用户广播进度
func (s *ProgressService) BroadcastToUser(userID string, taskID string, progress *ProgressMessage) {
	if s.hub != nil {
		s.hub.BroadcastToUser(userID, &Message{
			Type:    "progress",
			TaskID:  taskID,
			Payload: progress,
		})
	}
}

// BroadcastToProject 向项目广播进度
func (s *ProgressService) BroadcastToProject(projectID string, taskID string, progress *ProgressMessage) {
	if s.hub != nil {
		s.hub.BroadcastToProject(projectID, &Message{
			Type:    "progress",
			TaskID:  taskID,
			Payload: progress,
		})
	}
}

// NotifyTaskStart 通知任务开始
func (s *ProgressService) NotifyTaskStart(taskID, userID, projectID string) {
	s.UpdateProgress(taskID, 0, "任务已开始")

	logrus.Infof("任务开始: %s, 用户: %s, 项目: %s", taskID, userID, projectID)
}

// NotifyTaskComplete 通知任务完成
func (s *ProgressService) NotifyTaskComplete(taskID string, result interface{}) {
	s.UpdateProgress(taskID, 100, "任务已完成")
	s.UpdateTaskStatus(taskID, "completed", result, "")

	logrus.Infof("任务完成: %s", taskID)
}

// NotifyTaskFailed 通知任务失败
func (s *ProgressService) NotifyTaskFailed(taskID string, err error) {
	s.UpdateTaskStatus(taskID, "failed", nil, err.Error())

	logrus.Errorf("任务失败: %s, 错误: %v", taskID, err)
}

// NotifyTaskCancelled 通知任务取消
func (s *ProgressService) NotifyTaskCancelled(taskID string) {
	s.UpdateTaskStatus(taskID, "cancelled", nil, "用户取消")

	logrus.Infof("任务取消: %s", taskID)
}

// ProgressStats 进度统计
type ProgressStats struct {
	TotalTasks     int `json:"total_tasks"`
	ActiveTasks    int `json:"active_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
}

// GetStats 获取统计
func (s *ProgressService) GetStats() (*ProgressStats, error) {
	stats := &ProgressStats{}

	// 统计各状态任务数
	rows, err := s.db.Model(&database.AsyncTask{}).
		Select("status, count(*) as count").
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

		stats.TotalTasks += int(count)
		switch status {
		case "running", "pending":
			stats.ActiveTasks += int(count)
		case "completed":
			stats.CompletedTasks += int(count)
		case "failed":
			stats.FailedTasks += int(count)
		}
	}

	return stats, nil
}

// GenerateClientID 生成客户端ID
func GenerateClientID() string {
	return uuid.New().String()
}
