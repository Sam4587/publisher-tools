// Package task 提供增强的异步任务管理功能
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

// EnhancedTaskManager 增强版任务管理器
type EnhancedTaskManager struct {
	*TaskManager
	progressTracker    *ProgressTracker
	notificationService *NotificationService
	retryManager       *RetryManager
	metricsCollector   *MetricsCollector
}

// NewEnhancedTaskManager 创建增强版任务管理器
func NewEnhancedTaskManager(storage TaskStorage) *EnhancedTaskManager {
	baseManager := NewTaskManager(storage)

	return &EnhancedTaskManager{
		TaskManager:        baseManager,
		progressTracker:    NewProgressTracker(),
		notificationService: NewNotificationService(),
		retryManager:       NewRetryManager(),
		metricsCollector:   NewMetricsCollector(),
	}
}

// ExecuteWithRetry 带重试的任务执行
func (m *EnhancedTaskManager) ExecuteWithRetry(ctx context.Context, taskID string, maxRetries int) error {
	task, err := m.GetTask(taskID)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			logrus.Infof("重试任务 %s，第 %d 次尝试", taskID, attempt)

			// 计算重试延迟
			delay := m.retryManager.CalculateDelay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// 执行任务
		err = m.Execute(ctx, taskID)
		if err == nil {
			// 成功，记录指标
			m.metricsCollector.RecordSuccess(task.Type, attempt)
			return nil
		}

		lastErr = err
		logrus.Warnf("任务 %s 执行失败（第 %d 次尝试）: %v", taskID, attempt+1, err)

		// 更新进度
		m.progressTracker.UpdateProgress(taskID, ProgressDetail{
			TaskID:    taskID,
			Progress:  0,
			Message:   fmt.Sprintf("重试中... (%d/%d)", attempt+1, maxRetries),
			Timestamp: time.Now(),
		})
	}

	// 所有重试都失败
	m.metricsCollector.RecordFailure(task.Type, maxRetries)
	return fmt.Errorf("任务 %s 执行失败，已重试 %d 次: %w", taskID, maxRetries, lastErr)
}

// ExecuteWithProgress 带进度追踪的任务执行
func (m *EnhancedTaskManager) ExecuteWithProgress(ctx context.Context, taskID string, progressCallback func(int, string)) error {
	// 订阅进度更新
	_, cancel := m.progressTracker.Subscribe(taskID, func(detail ProgressDetail) {
		if progressCallback != nil {
			progressCallback(detail.Progress, detail.Message)
		}
	})
	defer cancel()

	return m.Execute(ctx, taskID)
}

// CreateAndExecute 创建并立即执行任务
func (m *EnhancedTaskManager) CreateAndExecute(ctx context.Context, taskType string, platform string, payload map[string]interface{}) (*Task, error) {
	task, err := m.CreateTask(taskType, platform, payload)
	if err != nil {
		return nil, err
	}

	// 异步执行
	go func() {
		if execErr := m.Execute(ctx, task.ID); execErr != nil {
			logrus.Errorf("任务执行失败: %v", execErr)
		}
	}()

	return task, nil
}

// GetProgress 获取任务进度
func (m *EnhancedTaskManager) GetProgress(taskID string) (*ProgressDetail, error) {
	return m.progressTracker.GetProgress(taskID)
}

// GetMetrics 获取任务指标
func (m *EnhancedTaskManager) GetMetrics(taskType string) (*TaskMetrics, error) {
	return m.metricsCollector.GetMetrics(taskType)
}

// ProgressTracker 进度追踪器
type ProgressTracker struct {
	mu          sync.RWMutex
	progress    map[string]*ProgressDetail
	subscribers map[string][]chan ProgressDetail
}

// NewProgressTracker 创建进度追踪器
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		progress:    make(map[string]*ProgressDetail),
		subscribers: make(map[string][]chan ProgressDetail),
	}
}

// UpdateProgress 更新进度
func (p *ProgressTracker) UpdateProgress(taskID string, detail ProgressDetail) {
	p.mu.Lock()
	defer p.mu.Unlock()

	detail.TaskID = taskID
	p.progress[taskID] = &detail

	// 通知订阅者
	for _, ch := range p.subscribers[taskID] {
		select {
		case ch <- detail:
		default:
			logrus.Warn("进度通知通道已满")
		}
	}
}

// GetProgress 获取进度
func (p *ProgressTracker) GetProgress(taskID string) (*ProgressDetail, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	detail, exists := p.progress[taskID]
	if !exists {
		return nil, fmt.Errorf("任务进度不存在: %s", taskID)
	}

	return detail, nil
}

// Subscribe 订阅进度更新
func (p *ProgressTracker) Subscribe(taskID string, callback func(ProgressDetail)) (func(), error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan ProgressDetail, 10)
	p.subscribers[taskID] = append(p.subscribers[taskID], ch)

	// 启动监听协程
	go func() {
		for detail := range ch {
			callback(detail)
		}
	}()

	// 返回取消订阅函数
	cancel := func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		// 从订阅列表中移除
		for i, subscriberCh := range p.subscribers[taskID] {
			if subscriberCh == ch {
				p.subscribers[taskID] = append(p.subscribers[taskID][:i], p.subscribers[taskID][i+1:]...)
				break
			}
		}

		close(ch)
	}

	return cancel, nil
}

// ProgressDetail 进度详情
type ProgressDetail struct {
	TaskID      string                 `json:"task_id"`
	Progress    int                    `json:"progress"`       // 0-100
	CurrentStep string                 `json:"current_step"`
	TotalSteps  int                    `json:"total_steps"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NotificationService 通知服务
type NotificationService struct {
	wsHub        *WebSocketHub
	emailService *EmailService
	webhookService *WebhookService
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	return &NotificationService{
		wsHub:          NewWebSocketHub(),
		emailService:   NewEmailService(),
		webhookService: NewWebhookService(),
	}
}

// NotifyProgress 通知进度更新
func (n *NotificationService) NotifyProgress(taskID string, detail ProgressDetail) {
	// WebSocket 推送
	n.wsHub.Broadcast(taskID, map[string]interface{}{
		"type": "progress",
		"data": detail,
	})
}

// NotifyCompletion 通知任务完成
func (n *NotificationService) NotifyCompletion(taskID string, task *Task) {
	n.wsHub.Broadcast(taskID, map[string]interface{}{
		"type": "completed",
		"data": task,
	})
}

// NotifyError 通知错误
func (n *NotificationService) NotifyError(taskID string, err error) {
	n.wsHub.Broadcast(taskID, map[string]interface{}{
		"type": "error",
		"data": map[string]interface{}{
			"task_id": taskID,
			"error":   err.Error(),
		},
	})
}

// WebSocketHub WebSocket 连接管理
type WebSocketHub struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewWebSocketHub 创建 WebSocket Hub
func NewWebSocketHub() *WebSocketHub {
	hub := &WebSocketHub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go hub.run()

	return hub
}

// run 运行 Hub
func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			logrus.Infof("WebSocket 客户端已连接: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			logrus.Infof("WebSocket 客户端已断开: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast 广播消息
func (h *WebSocketHub) Broadcast(topic string, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	h.broadcast <- message
}

// Client WebSocket 客户端
type Client struct {
	ID     string
	Send   chan []byte
	Topics []string
}

// EmailService 邮件服务
type EmailService struct {
	// 邮件服务配置
}

// NewEmailService 创建邮件服务
func NewEmailService() *EmailService {
	return &EmailService{}
}

// Send 发送邮件
func (s *EmailService) Send(to string, subject string, body string) error {
	// 实现邮件发送逻辑
	logrus.Infof("发送邮件到 %s: %s", to, subject)
	return nil
}

// WebhookService Webhook 服务
type WebhookService struct {
	// Webhook 配置
}

// NewWebhookService 创建 Webhook 服务
func NewWebhookService() *WebhookService {
	return &WebhookService{}
}

// Trigger 触发 Webhook
func (s *WebhookService) Trigger(url string, data interface{}) error {
	// 实现 Webhook 触发逻辑
	logrus.Infof("触发 Webhook: %s", url)
	return nil
}

// RetryManager 重试管理器
type RetryManager struct {
	strategy   RetryStrategy
	maxRetries int
}

// NewRetryManager 创建重试管理器
func NewRetryManager() *RetryManager {
	return &RetryManager{
		strategy: RetryStrategy{
			Type:          RetryTypeExponential,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
		},
		maxRetries: 3,
	}
}

// CalculateDelay 计算重试延迟
func (r *RetryManager) CalculateDelay(attempt int) time.Duration {
	switch r.strategy.Type {
	case RetryTypeFixed:
		return r.strategy.InitialDelay

	case RetryTypeExponential:
		delay := r.strategy.InitialDelay
		for i := 1; i < attempt; i++ {
			delay = time.Duration(float64(delay) * r.strategy.BackoffFactor)
			if delay > r.strategy.MaxDelay {
				return r.strategy.MaxDelay
			}
		}
		return delay

	case RetryTypeLinear:
		delay := r.strategy.InitialDelay + time.Duration(attempt)*time.Second
		if delay > r.strategy.MaxDelay {
			return r.strategy.MaxDelay
		}
		return delay

	default:
		return r.strategy.InitialDelay
	}
}

// SetStrategy 设置重试策略
func (r *RetryManager) SetStrategy(strategy RetryStrategy) {
	r.strategy = strategy
}

// RetryStrategy 重试策略
type RetryStrategy struct {
	Type          RetryType     `json:"type"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// RetryType 重试类型
type RetryType string

const (
	RetryTypeFixed       RetryType = "fixed"
	RetryTypeExponential RetryType = "exponential"
	RetryTypeLinear      RetryType = "linear"
)

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu       sync.RWMutex
	metrics  map[string]*TaskMetrics
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*TaskMetrics),
	}
}

// RecordSuccess 记录成功
func (m *MetricsCollector) RecordSuccess(taskType string, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.metrics[taskType]; !exists {
		m.metrics[taskType] = &TaskMetrics{}
	}

	metrics := m.metrics[taskType]
	metrics.TotalExecutions++
	metrics.SuccessfulExecutions++
	metrics.SuccessRate = float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions)
}

// RecordFailure 记录失败
func (m *MetricsCollector) RecordFailure(taskType string, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.metrics[taskType]; !exists {
		m.metrics[taskType] = &TaskMetrics{}
	}

	metrics := m.metrics[taskType]
	metrics.TotalExecutions++
	metrics.FailedExecutions++
	metrics.SuccessRate = float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions)
}

// GetMetrics 获取指标
func (m *MetricsCollector) GetMetrics(taskType string) (*TaskMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics, exists := m.metrics[taskType]
	if !exists {
		return nil, fmt.Errorf("任务类型指标不存在: %s", taskType)
	}

	return metrics, nil
}

// TaskMetrics 任务指标
type TaskMetrics struct {
	TotalExecutions      int64   `json:"total_executions"`
	SuccessfulExecutions int64   `json:"successful_executions"`
	FailedExecutions     int64   `json:"failed_executions"`
	SuccessRate          float64 `json:"success_rate"`
}
