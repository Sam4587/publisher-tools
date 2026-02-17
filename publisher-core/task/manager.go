// Package task 提供异步任务管理功能
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TaskStatus 任务状�?
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task 任务定义
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      TaskStatus             `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Progress    int                    `json:"progress"`     // 0-100
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`
	Platform    string                 `json:"platform"`
}

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *Task) error

// TaskManager 任务管理�?
type TaskManager struct {
	mu       sync.RWMutex
	tasks    map[string]*Task
	handlers map[string]TaskHandler
	storage  TaskStorage
	notifyCh chan *Task
}

// TaskStorage 任务存储接口
type TaskStorage interface {
	Save(task *Task) error
	Load(id string) (*Task, error)
	List(filter TaskFilter) ([]*Task, error)
	Delete(id string) error
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	Status   TaskStatus
	Platform string
	Type     string
	Limit    int
}

// NewTaskManager 创建任务管理�?
func NewTaskManager(storage TaskStorage) *TaskManager {
	return &TaskManager{
		tasks:    make(map[string]*Task),
		handlers: make(map[string]TaskHandler),
		storage:  storage,
		notifyCh: make(chan *Task, 100),
	}
}

// RegisterHandler 注册任务处理�?
func (m *TaskManager) RegisterHandler(taskType string, handler TaskHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[taskType] = handler
	logrus.Infof("已注册任务处理器: %s", taskType)
}

// CreateTask 创建新任�?
func (m *TaskManager) CreateTask(taskType string, platform string, payload map[string]interface{}) (*Task, error) {
	task := &Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		Status:    TaskStatusPending,
		Payload:   payload,
		Progress:  0,
		CreatedAt: time.Now(),
		Platform:  platform,
	}

	m.mu.Lock()
	m.tasks[task.ID] = task
	m.mu.Unlock()

	if m.storage != nil {
		if err := m.storage.Save(task); err != nil {
			logrus.Warnf("保存任务失败: %v", err)
		}
	}

	logrus.Infof("创建任务: %s, 类型: %s, 平台: %s", task.ID, taskType, platform)
	return task, nil
}

// Execute 执行任务
func (m *TaskManager) Execute(ctx context.Context, taskID string) error {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	handler := m.handlers[task.Type]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("任务不存�? %s", taskID)
	}

	if handler == nil {
		return fmt.Errorf("未注册任务处理器: %s", task.Type)
	}

	// 更新状态为运行�?
	m.updateTaskStatus(task, TaskStatusRunning)
	now := time.Now()
	task.StartedAt = &now

	// 执行任务
	err := handler(ctx, task)

	// 更新最终状�?
	if err != nil {
		task.Error = err.Error()
		m.updateTaskStatus(task, TaskStatusFailed)
	} else {
		m.updateTaskStatus(task, TaskStatusCompleted)
	}

	finishedAt := time.Now()
	task.FinishedAt = &finishedAt

	// 保存结果
	if m.storage != nil {
		if saveErr := m.storage.Save(task); saveErr != nil {
			logrus.Warnf("保存任务结果失败: %v", saveErr)
		}
	}

	return err
}

// ExecuteAsync 异步执行任务
func (m *TaskManager) ExecuteAsync(ctx context.Context, taskID string) <-chan error {
	resultCh := make(chan error, 1)

	go func() {
		defer close(resultCh)
		resultCh <- m.Execute(ctx, taskID)
	}()

	return resultCh
}

// GetTask 获取任务
func (m *TaskManager) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	m.mu.RUnlock()

	if exists {
		return task, nil
	}

	// 从存储加�?
	if m.storage != nil {
		task, err := m.storage.Load(taskID)
		if err != nil {
			return nil, fmt.Errorf("加载任务失败: %w", err)
		}
		m.mu.Lock()
		m.tasks[taskID] = task
		m.mu.Unlock()
		return task, nil
	}

	return nil, fmt.Errorf("任务不存�? %s", taskID)
}

// ListTasks 列出任务
func (m *TaskManager) ListTasks(filter TaskFilter) ([]*Task, error) {
	if m.storage != nil {
		return m.storage.List(filter)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Task
	for _, task := range m.tasks {
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Platform != "" && task.Platform != filter.Platform {
			continue
		}
		if filter.Type != "" && task.Type != filter.Type {
			continue
		}
		result = append(result, task)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}

	return result, nil
}

// Cancel 取消任务
func (m *TaskManager) Cancel(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存�? %s", taskID)
	}

	if task.Status == TaskStatusRunning {
		return fmt.Errorf("无法取消正在运行的任�?)
	}

	task.Status = TaskStatusCancelled
	now := time.Now()
	task.FinishedAt = &now

	if m.storage != nil {
		if err := m.storage.Save(task); err != nil {
			logrus.Warnf("保存任务状态失�? %v", err)
		}
	}

	return nil
}

// UpdateProgress 更新任务进度
func (m *TaskManager) UpdateProgress(taskID string, progress int, result map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task, exists := m.tasks[taskID]; exists {
		task.Progress = progress
		if result != nil {
			task.Result = result
		}

		// 通知进度更新
		select {
		case m.notifyCh <- task:
		default:
			logrus.Warn("通知通道已满，跳过进度通知")
		}
	}
}

// Notify 返回任务通知通道
func (m *TaskManager) Notify() <-chan *Task {
	return m.notifyCh
}

func (m *TaskManager) updateTaskStatus(task *Task, status TaskStatus) {
	m.mu.Lock()
	task.Status = status
	m.mu.Unlock()
}

// ToJSON 转换为JSON
func (t *Task) ToJSON() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		tasks: make(map[string]*Task),
	}
}

func (s *MemoryStorage) Save(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return nil
}

func (s *MemoryStorage) Load(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("任务不存�? %s", id)
	}
	return task, nil
}

func (s *MemoryStorage) List(filter TaskFilter) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Task
	for _, task := range s.tasks {
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Platform != "" && task.Platform != filter.Platform {
			continue
		}
		if filter.Type != "" && task.Type != filter.Type {
			continue
		}
		result = append(result, task)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}

	return result, nil
}

func (s *MemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	return nil
}
