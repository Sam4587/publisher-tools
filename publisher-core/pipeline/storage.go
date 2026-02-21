// Package pipeline 提供流水线存储服务
package pipeline

import (
	"encoding/json"
	"fmt"
	"publisher-core/database"
	"time"

	"gorm.io/gorm"
)

// DBStorage 数据库存储实现
type DBStorage struct {
	db *gorm.DB
}

// NewDBStorage 创建数据库存储
func NewDBStorage(db *gorm.DB) *DBStorage {
	return &DBStorage{db: db}
}

// SavePipeline 保存流水线定义
func (s *DBStorage) SavePipeline(pipeline *Pipeline) error {
	stepsJSON, err := json.Marshal(pipeline.Steps)
	if err != nil {
		return fmt.Errorf("序列化步骤失败: %w", err)
	}

	configJSON, err := json.Marshal(pipeline.Config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	definition := &database.PipelineDefinition{
		ID:          pipeline.ID,
		Name:        pipeline.Name,
		Description: pipeline.Description,
		Steps:       string(stepsJSON),
		Config:      string(configJSON),
		IsActive:    pipeline.Status == PipelineStatusActive,
		Version:     1,
		CreatedAt:   pipeline.CreatedAt,
		UpdatedAt:   pipeline.UpdatedAt,
	}

	// 使用事务保存
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否存在
		var existing database.PipelineDefinition
		result := tx.Where("id = ?", pipeline.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新记录
			return tx.Create(definition).Error
		} else if result.Error != nil {
			return result.Error
		}

		// 更新现有记录
		definition.Version = existing.Version + 1
		definition.CreatedAt = existing.CreatedAt
		return tx.Save(definition).Error
	})
}

// LoadPipeline 加载流水线定义
func (s *DBStorage) LoadPipeline(id string) (*Pipeline, error) {
	var definition database.PipelineDefinition
	if err := s.db.Where("id = ?", id).First(&definition).Error; err != nil {
		return nil, err
	}

	// 解析步骤
	var steps []PipelineStep
	if err := json.Unmarshal([]byte(definition.Steps), &steps); err != nil {
		return nil, fmt.Errorf("解析步骤失败: %w", err)
	}

	// 解析配置
	var config PipelineConfig
	if err := json.Unmarshal([]byte(definition.Config), &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &Pipeline{
		ID:          definition.ID,
		Name:        definition.Name,
		Description: definition.Description,
		Steps:       steps,
		Config:      config,
		Status:      map[bool]PipelineStatus{true: PipelineStatusActive, false: PipelineStatusDraft}[definition.IsActive],
		CreatedAt:   definition.CreatedAt,
		UpdatedAt:   definition.UpdatedAt,
	}, nil
}

// SaveExecution 保存执行记录
func (s *DBStorage) SaveExecution(execution *PipelineExecution) error {
	inputJSON, err := json.Marshal(execution.Input)
	if err != nil {
		return fmt.Errorf("序列化输入失败: %w", err)
	}

	outputJSON, err := json.Marshal(execution.Output)
	if err != nil {
		return fmt.Errorf("序列化输出失败: %w", err)
	}

	stepsJSON, err := json.Marshal(execution.Steps)
	if err != nil {
		return fmt.Errorf("序列化步骤失败: %w", err)
	}

	var durationMs int
	if execution.FinishedAt != nil {
		durationMs = int(execution.FinishedAt.Sub(execution.StartedAt).Milliseconds())
	}

	record := &database.PipelineExecutionRecord{
		ExecutionID: execution.ID,
		PipelineID:  execution.PipelineID,
		Status:      string(execution.Status),
		Input:       string(inputJSON),
		Output:      string(outputJSON),
		Steps:       string(stepsJSON),
		Error:       execution.Error,
		StartedAt:   execution.StartedAt,
		FinishedAt:  execution.FinishedAt,
		DurationMs:  durationMs,
	}

	// 使用事务保存
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否存在
		var existing database.PipelineExecutionRecord
		result := tx.Where("execution_id = ?", execution.ID).First(&existing)
		
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新记录
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		} else {
			// 更新现有记录
			record.ID = existing.ID
			record.CreatedAt = existing.CreatedAt
			if err := tx.Save(record).Error; err != nil {
				return err
			}
		}

		// 保存步骤执行记录
		for _, step := range execution.Steps {
			if err := s.saveStepExecution(tx, execution.ID, step); err != nil {
				return err
			}
		}

		return nil
	})
}

// saveStepExecution 保存步骤执行记录
func (s *DBStorage) saveStepExecution(tx *gorm.DB, executionID string, step StepExecution) error {
	inputJSON, _ := json.Marshal(step.Input)
	outputJSON, _ := json.Marshal(step.Output)

	var durationMs int
	if step.FinishedAt != nil {
		durationMs = int(step.FinishedAt.Sub(step.StartedAt).Milliseconds())
	}

	record := &database.PipelineStepExecution{
		ExecutionID: executionID,
		StepID:      step.StepID,
		StepName:    step.Name,
		Status:      string(step.Status),
		Input:       string(inputJSON),
		Output:      string(outputJSON),
		Error:       step.Error,
		Progress:    step.Progress,
		DurationMs:  durationMs,
		StartedAt:   step.StartedAt,
		FinishedAt:  step.FinishedAt,
	}

	// 检查是否存在
	var existing database.PipelineStepExecution
	result := tx.Where("execution_id = ? AND step_id = ?", executionID, step.StepID).First(&existing)
	
	if result.Error == gorm.ErrRecordNotFound {
		return tx.Create(record).Error
	} else if result.Error != nil {
		return result.Error
	}

	record.ID = existing.ID
	return tx.Save(record).Error
}

// LoadExecution 加载执行记录
func (s *DBStorage) LoadExecution(id string) (*PipelineExecution, error) {
	var record database.PipelineExecutionRecord
	if err := s.db.Where("execution_id = ?", id).First(&record).Error; err != nil {
		return nil, err
	}

	// 解析输入
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(record.Input), &input); err != nil {
		input = make(map[string]interface{})
	}

	// 解析输出
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(record.Output), &output); err != nil {
		output = make(map[string]interface{})
	}

	// 解析步骤
	var steps []StepExecution
	if err := json.Unmarshal([]byte(record.Steps), &steps); err != nil {
		steps = make([]StepExecution, 0)
	}

	return &PipelineExecution{
		ID:         record.ExecutionID,
		PipelineID: record.PipelineID,
		Status:     ExecutionStatus(record.Status),
		Input:      input,
		Output:     output,
		Steps:      steps,
		StartedAt:  record.StartedAt,
		FinishedAt: record.FinishedAt,
		Error:      record.Error,
	}, nil
}

// ListPipelines 列出流水线
func (s *DBStorage) ListPipelines(activeOnly bool) ([]*Pipeline, error) {
	query := s.db.Model(&database.PipelineDefinition{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	var definitions []database.PipelineDefinition
	if err := query.Find(&definitions).Error; err != nil {
		return nil, err
	}

	pipelines := make([]*Pipeline, 0, len(definitions))
	for _, def := range definitions {
		pipeline, err := s.LoadPipeline(def.ID)
		if err != nil {
			continue
		}
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

// ListExecutions 列出执行记录
func (s *DBStorage) ListExecutions(pipelineID string, limit int) ([]*PipelineExecution, error) {
	query := s.db.Model(&database.PipelineExecutionRecord{})
	if pipelineID != "" {
		query = query.Where("pipeline_id = ?", pipelineID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	var records []database.PipelineExecutionRecord
	if err := query.Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	executions := make([]*PipelineExecution, 0, len(records))
	for _, record := range records {
		execution, err := s.LoadExecution(record.ExecutionID)
		if err != nil {
			continue
		}
		executions = append(executions, execution)
	}

	return executions, nil
}

// DeletePipeline 删除流水线
func (s *DBStorage) DeletePipeline(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除相关的执行记录
		if err := tx.Where("pipeline_id = ?", id).Delete(&database.PipelineExecutionRecord{}).Error; err != nil {
			return err
		}

		// 删除流水线定义
		return tx.Where("id = ?", id).Delete(&database.PipelineDefinition{}).Error
	})
}

// GetExecutionStats 获取执行统计
func (s *DBStorage) GetExecutionStats(pipelineID string) (*ExecutionStats, error) {
	stats := &ExecutionStats{}

	query := s.db.Model(&database.PipelineExecutionRecord{})
	if pipelineID != "" {
		query = query.Where("pipeline_id = ?", pipelineID)
	}

	// 统计各状态数量
	rows, err := query.Select("status, count(*) as count").Group("status").Rows()
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

		stats.TotalExecutions += int(count)
		switch status {
		case "running":
			stats.RunningExecutions = int(count)
		case "completed":
			stats.CompletedExecutions = int(count)
		case "failed":
			stats.FailedExecutions = int(count)
		}
	}

	// 计算平均执行时间
	var avgDuration float64
	s.db.Model(&database.PipelineExecutionRecord{}).
		Where("pipeline_id = ? AND status = ?", pipelineID, "completed").
		Select("COALESCE(AVG(duration_ms), 0)").
		Scan(&avgDuration)
	stats.AvgDurationMs = avgDuration

	return stats, nil
}

// ExecutionStats 执行统计
type ExecutionStats struct {
	TotalExecutions     int     `json:"total_executions"`
	RunningExecutions   int     `json:"running_executions"`
	CompletedExecutions int     `json:"completed_executions"`
	FailedExecutions    int     `json:"failed_executions"`
	AvgDurationMs       float64 `json:"avg_duration_ms"`
}

// CleanupOldExecutions 清理旧执行记录
func (s *DBStorage) CleanupOldExecutions(maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)

	result := s.db.Where("created_at < ? AND status IN ?", cutoff, []string{"completed", "failed", "cancelled"}).
		Delete(&database.PipelineExecutionRecord{})

	return result.RowsAffected, result.Error
}
