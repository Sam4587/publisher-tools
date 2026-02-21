// Package pipeline 提供流水线编排功能
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Orchestrator 编排器接口
type Orchestrator interface {
	// CreatePipeline 创建流水线
	CreatePipeline(pipeline *Pipeline) error

	// ExecutePipeline 执行流水线
	ExecutePipeline(ctx context.Context, pipelineID string, input map[string]interface{}) (*PipelineExecution, error)

	// PausePipeline 暂停流水线
	PausePipeline(executionID string) error

	// ResumePipeline 恢复流水线
	ResumePipeline(executionID string) error

	// CancelPipeline 取消流水线
	CancelPipeline(executionID string) error

	// GetExecutionStatus 获取执行状态
	GetExecutionStatus(executionID string) (*PipelineExecution, error)

	// GetExecutionLogs 获取执行日志
	GetExecutionLogs(executionID string) ([]ExecutionLog, error)

	// ListPipelines 列出流水线
	ListPipelines() ([]*Pipeline, error)

	// GetPipeline 获取流水线
	GetPipeline(pipelineID string) (*Pipeline, error)
}

// PipelineOrchestrator 流水线编排器实现
type PipelineOrchestrator struct {
	mu              sync.RWMutex
	pipelines       map[string]*Pipeline
	executions      map[string]*PipelineExecution
	stepHandlers    map[string]StepHandler
	progressTracker *ProgressTracker
	notificationService *NotificationService
	storage         PipelineStorage
}

// NewPipelineOrchestrator 创建流水线编排器
func NewPipelineOrchestrator(storage PipelineStorage) *PipelineOrchestrator {
	orchestrator := &PipelineOrchestrator{
		pipelines:           make(map[string]*Pipeline),
		executions:          make(map[string]*PipelineExecution),
		stepHandlers:        make(map[string]StepHandler),
		progressTracker:     NewProgressTracker(),
		notificationService: NewNotificationService(),
		storage:             storage,
	}

	// 注册预定义处理器
	orchestrator.registerDefaultHandlers()

	return orchestrator
}

// CreatePipeline 创建流水线
func (o *PipelineOrchestrator) CreatePipeline(pipeline *Pipeline) error {
	if pipeline.ID == "" {
		pipeline.ID = uuid.New().String()
	}

	pipeline.Status = PipelineStatusDraft
	pipeline.CreatedAt = time.Now()
	pipeline.UpdatedAt = time.Now()

	o.mu.Lock()
	o.pipelines[pipeline.ID] = pipeline
	o.mu.Unlock()

	if o.storage != nil {
		if err := o.storage.SavePipeline(pipeline); err != nil {
			logrus.Warnf("保存流水线失败: %v", err)
		}
	}

	logrus.Infof("创建流水线: %s (%s)", pipeline.ID, pipeline.Name)
	return nil
}

// ExecutePipeline 执行流水线
func (o *PipelineOrchestrator) ExecutePipeline(ctx context.Context, pipelineID string, input map[string]interface{}) (*PipelineExecution, error) {
	o.mu.RLock()
	pipeline, exists := o.pipelines[pipelineID]
	o.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("流水线不存在: %s", pipelineID)
	}

	// 创建执行实例
	execution := &PipelineExecution{
		ID:         uuid.New().String(),
		PipelineID: pipelineID,
		Status:     ExecutionStatusRunning,
		Input:      input,
		Output:     make(map[string]interface{}),
		Steps:      make([]StepExecution, 0, len(pipeline.Steps)),
		StartedAt:  time.Now(),
	}

	// 初始化步骤执行状态
	for _, step := range pipeline.Steps {
		execution.Steps = append(execution.Steps, StepExecution{
			StepID:  step.ID,
			Name:    step.Name,
			Status:  StepStatusPending,
			Input:   make(map[string]interface{}),
			Output:  make(map[string]interface{}),
			Logs:    make([]string, 0),
		})
	}

	o.mu.Lock()
	o.executions[execution.ID] = execution
	o.mu.Unlock()

	// 异步执行
	go o.executeSteps(ctx, pipeline, execution)

	logrus.Infof("开始执行流水线: %s (执行ID: %s)", pipelineID, execution.ID)
	return execution, nil
}

// executeSteps 执行步骤
func (o *PipelineOrchestrator) executeSteps(ctx context.Context, pipeline *Pipeline, execution *PipelineExecution) {
	defer func() {
		finishedAt := time.Now()
		execution.FinishedAt = &finishedAt

		if execution.Status == ExecutionStatusRunning {
			execution.Status = ExecutionStatusCompleted
		}

		o.notificationService.NotifyCompletion(execution.ID, execution)

		if o.storage != nil {
			o.storage.SaveExecution(execution)
		}
	}()

	// 构建步骤依赖图
	dependencyGraph := o.buildDependencyGraph(pipeline.Steps)

	// 按顺序执行步骤
	for i, step := range pipeline.Steps {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			execution.Status = ExecutionStatusCancelled
			execution.Error = ctx.Err().Error()
			return
		default:
		}

		// 检查是否已暂停或取消
		if execution.Status == ExecutionStatusPaused || execution.Status == ExecutionStatusCancelled {
			return
		}

		// 更新执行状态
		stepExecution := &execution.Steps[i]
		stepExecution.Status = StepStatusRunning
		stepExecution.StartedAt = time.Now()

		// 通知进度
		o.progressTracker.UpdateProgress(execution.ID, ProgressDetail{
			ExecutionID: execution.ID,
			StepID:      step.ID,
			Progress:    int(float64(i) / float64(len(pipeline.Steps)) * 100),
			CurrentStep: fmt.Sprintf("步骤 %d/%d: %s", i+1, len(pipeline.Steps), step.Name),
			TotalSteps:  len(pipeline.Steps),
			Message:     fmt.Sprintf("正在执行: %s", step.Name),
			Timestamp:   time.Now(),
		})

		// 执行步骤
		output, err := o.executeStep(ctx, step, execution.Input, execution)
		if err != nil {
			stepExecution.Status = StepStatusFailed
			stepExecution.Error = err.Error()
			execution.Error = fmt.Sprintf("步骤 %s 失败: %v", step.Name, err)

			if pipeline.Config.FailFast {
				execution.Status = ExecutionStatusFailed
				o.notificationService.NotifyError(execution.ID, err)
				return
			}

			// 记录日志并继续
			stepExecution.Logs = append(stepExecution.Logs, fmt.Sprintf("错误: %v", err))
		} else {
			stepExecution.Status = StepStatusCompleted
			stepExecution.Output = output

			// 合并输出到执行结果
			for k, v := range output {
				execution.Output[k] = v
			}
		}

		finishedAt := time.Now()
		stepExecution.FinishedAt = &finishedAt

		// 通知进度
		o.progressTracker.UpdateProgress(execution.ID, ProgressDetail{
			ExecutionID: execution.ID,
			StepID:      step.ID,
			Progress:    int(float64(i+1) / float64(len(pipeline.Steps)) * 100),
			CurrentStep: fmt.Sprintf("步骤 %d/%d: %s", i+1, len(pipeline.Steps), step.Name),
			TotalSteps:  len(pipeline.Steps),
			Message:     fmt.Sprintf("完成: %s", step.Name),
			Timestamp:   time.Now(),
		})
	}

	execution.Status = ExecutionStatusCompleted
	logrus.Infof("流水线执行完成: %s (执行ID: %s)", pipeline.ID, execution.ID)
}

// executeStep 执行单个步骤
func (o *PipelineOrchestrator) executeStep(ctx context.Context, step PipelineStep, input map[string]interface{}, execution *PipelineExecution) (map[string]interface{}, error) {
	// 获取步骤处理器
	handler, exists := o.stepHandlers[step.Handler]
	if !exists {
		return nil, fmt.Errorf("未找到步骤处理器: %s", step.Handler)
	}

	// 创建带超时的上下文
	stepCtx, cancel := context.WithTimeout(ctx, step.Timeout)
	defer cancel()

	// 执行步骤
	output, err := handler.Execute(stepCtx, step.Config, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// buildDependencyGraph 构建依赖图
func (o *PipelineOrchestrator) buildDependencyGraph(steps []PipelineStep) map[string][]string {
	graph := make(map[string][]string)
	for _, step := range steps {
		graph[step.ID] = step.DependsOn
	}
	return graph
}

// PausePipeline 暂停流水线
func (o *PipelineOrchestrator) PausePipeline(executionID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if execution.Status != ExecutionStatusRunning {
		return fmt.Errorf("只能暂停正在运行的执行")
	}

	execution.Status = ExecutionStatusPaused
	logrus.Infof("暂停流水线执行: %s", executionID)
	return nil
}

// ResumePipeline 恢复流水线
func (o *PipelineOrchestrator) ResumePipeline(executionID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if execution.Status != ExecutionStatusPaused {
		return fmt.Errorf("只能恢复已暂停的执行")
	}

	execution.Status = ExecutionStatusRunning
	logrus.Infof("恢复流水线执行: %s", executionID)
	return nil
}

// CancelPipeline 取消流水线
func (o *PipelineOrchestrator) CancelPipeline(executionID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return fmt.Errorf("执行不存在: %s", executionID)
	}

	if execution.Status == ExecutionStatusCompleted {
		return fmt.Errorf("无法取消已完成的执行")
	}

	execution.Status = ExecutionStatusCancelled
	logrus.Infof("取消流水线执行: %s", executionID)
	return nil
}

// GetExecutionStatus 获取执行状态
func (o *PipelineOrchestrator) GetExecutionStatus(executionID string) (*PipelineExecution, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("执行不存在: %s", executionID)
	}

	return execution, nil
}

// GetExecutionLogs 获取执行日志
func (o *PipelineOrchestrator) GetExecutionLogs(executionID string) ([]ExecutionLog, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	execution, exists := o.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("执行不存在: %s", executionID)
	}

	logs := make([]ExecutionLog, 0)
	for _, step := range execution.Steps {
		for _, log := range step.Logs {
			logs = append(logs, ExecutionLog{
				ExecutionID: executionID,
				StepID:      step.StepID,
				Message:     log,
				Timestamp:   step.StartedAt,
			})
		}
	}

	return logs, nil
}

// ListPipelines 列出流水线
func (o *PipelineOrchestrator) ListPipelines() ([]*Pipeline, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pipelines := make([]*Pipeline, 0, len(o.pipelines))
	for _, pipeline := range o.pipelines {
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// GetPipeline 获取流水线
func (o *PipelineOrchestrator) GetPipeline(pipelineID string) (*Pipeline, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pipeline, exists := o.pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("流水线不存在: %s", pipelineID)
	}

	return pipeline, nil
}

// RegisterHandler 注册步骤处理器
func (o *PipelineOrchestrator) RegisterHandler(handlerName string, handler StepHandler) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.stepHandlers[handlerName] = handler
	logrus.Infof("注册步骤处理器: %s", handlerName)
}

// registerDefaultHandlers 注册默认处理器
func (o *PipelineOrchestrator) registerDefaultHandlers() {
	// 这里注册预定义的步骤处理器
	// 实际实现中会集成现有的 AI 服务、发布器等
}

// Pipeline 流水线定义
type Pipeline struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []PipelineStep `json:"steps"`
	Config      PipelineConfig `json:"config"`
	Status      PipelineStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// PipelineStep 流水线步骤
type PipelineStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       StepType               `json:"type"`
	Handler    string                 `json:"handler"`
	Config     map[string]interface{} `json:"config"`
	DependsOn  []string               `json:"depends_on"`
	RetryCount int                    `json:"retry_count"`
	Timeout    time.Duration          `json:"timeout"`
}

// PipelineConfig 流水线配置
type PipelineConfig struct {
	ParallelMode  bool               `json:"parallel_mode"`
	MaxParallel   int                `json:"max_parallel"`
	FailFast      bool               `json:"fail_fast"`
	RetryStrategy RetryStrategy      `json:"retry_strategy"`
	Notification  NotificationConfig `json:"notification"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	OnStart    bool     `json:"on_start"`
	OnComplete bool     `json:"on_complete"`
	OnError    bool     `json:"on_error"`
	Channels   []string `json:"channels"`
}

// RetryStrategy 重试策略
type RetryStrategy struct {
	Type          RetryType     `json:"type"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	MaxRetries    int           `json:"max_retries"`
}

// RetryType 重试类型
type RetryType string

const (
	RetryTypeNone        RetryType = "none"
	RetryTypeFixed       RetryType = "fixed"
	RetryTypeLinear      RetryType = "linear"
	RetryTypeExponential RetryType = "exponential"
)

// PipelineStatus 流水线状态
type PipelineStatus string

const (
	PipelineStatusDraft     PipelineStatus = "draft"
	PipelineStatusActive    PipelineStatus = "active"
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusCompleted PipelineStatus = "completed"
	PipelineStatusFailed    PipelineStatus = "failed"
	PipelineStatusPaused    PipelineStatus = "paused"
)

// StepType 步骤类型
type StepType string

const (
	StepTypeContentGeneration   StepType = "content_generation"
	StepTypeContentOptimization StepType = "content_optimization"
	StepTypeQualityScoring      StepType = "quality_scoring"
	StepTypePublishExecution    StepType = "publish_execution"
	StepTypeDataCollection      StepType = "data_collection"
	StepTypeAnalytics           StepType = "analytics"
)

// PipelineExecution 流水线执行实例
type PipelineExecution struct {
	ID         string                 `json:"id"`
	PipelineID string                 `json:"pipeline_id"`
	Status     ExecutionStatus        `json:"status"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output"`
	Steps      []StepExecution        `json:"steps"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// ExecutionStatus 执行状态
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusPaused    ExecutionStatus = "paused"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// StepExecution 步骤执行实例
type StepExecution struct {
	StepID    string                 `json:"step_id"`
	Name      string                 `json:"name"`
	Status    StepStatus             `json:"status"`
	Input     map[string]interface{} `json:"input"`
	Output    map[string]interface{} `json:"output"`
	Progress  int                    `json:"progress"`
	StartedAt time.Time              `json:"started_at"`
	FinishedAt *time.Time            `json:"finished_at,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Logs      []string               `json:"logs"`
}

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// ExecutionLog 执行日志
type ExecutionLog struct {
	ExecutionID string    `json:"execution_id"`
	StepID      string    `json:"step_id"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// StepHandler 步骤处理器接口
type StepHandler interface {
	Execute(ctx context.Context, config map[string]interface{}, input map[string]interface{}) (map[string]interface{}, error)
}

// PipelineStorage 流水线存储接口
type PipelineStorage interface {
	SavePipeline(pipeline *Pipeline) error
	LoadPipeline(id string) (*Pipeline, error)
	SaveExecution(execution *PipelineExecution) error
	LoadExecution(id string) (*PipelineExecution, error)
}

// ProgressDetail 进度详情
type ProgressDetail struct {
	ExecutionID string                 `json:"execution_id"`
	StepID      string                 `json:"step_id"`
	Progress    int                    `json:"progress"`
	CurrentStep string                 `json:"current_step"`
	TotalSteps  int                    `json:"total_steps"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ProgressTracker 进度追踪器
type ProgressTracker struct {
	mu       sync.RWMutex
	progress map[string]*ProgressDetail
}

// NewProgressTracker 创建进度追踪器
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		progress: make(map[string]*ProgressDetail),
	}
}

// UpdateProgress 更新进度
func (p *ProgressTracker) UpdateProgress(executionID string, detail ProgressDetail) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.progress[executionID] = &detail
}

// GetProgress 获取进度
func (p *ProgressTracker) GetProgress(executionID string) (*ProgressDetail, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	detail, exists := p.progress[executionID]
	if !exists {
		return nil, fmt.Errorf("执行进度不存在: %s", executionID)
	}

	return detail, nil
}

// NotificationService 通知服务
type NotificationService struct {
	// 通知服务实现
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

// NotifyCompletion 通知完成
func (n *NotificationService) NotifyCompletion(executionID string, execution *PipelineExecution) {
	logrus.Infof("执行完成: %s", executionID)
}

// NotifyError 通知错误
func (n *NotificationService) NotifyError(executionID string, err error) {
	logrus.Errorf("执行错误: %s, 错误: %v", executionID, err)
}

// ToJSON 转换为JSON
func (p *Pipeline) ToJSON() string {
	data, _ := json.MarshalIndent(p, "", "  ")
	return string(data)
}

func (e *PipelineExecution) ToJSON() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}
