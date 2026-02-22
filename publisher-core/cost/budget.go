package cost

import (
	"fmt"
	"time"

	"publisher-core/database"

	"gorm.io/gorm"
)

// BudgetService 预算服务
type BudgetService struct {
	db         *gorm.DB
	costService *CostService
}

// NewBudgetService 创建预算服务
func NewBudgetService(db *gorm.DB, costService *CostService) *BudgetService {
	return &BudgetService{
		db:          db,
		costService: costService,
	}
}

// CreateBudgetRequest 创建预算请求
type CreateBudgetRequest struct {
	UserID         string    `json:"user_id"`
	ProjectID      string    `json:"project_id"`
	BudgetType     string    `json:"budget_type"`      // daily, weekly, monthly
	BudgetAmount   float64   `json:"budget_amount"`
	AlertThreshold float64   `json:"alert_threshold"`  // 0-100
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}

// CreateBudget 创建预算
func (s *BudgetService) CreateBudget(req *CreateBudgetRequest) (*database.AIBudget, error) {
	budget := &database.AIBudget{
		UserID:         req.UserID,
		ProjectID:      req.ProjectID,
		BudgetType:     req.BudgetType,
		BudgetAmount:   req.BudgetAmount,
		UsedAmount:     0,
		AlertThreshold: req.AlertThreshold,
		IsActive:       true,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		LastResetAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(budget).Error; err != nil {
		return nil, fmt.Errorf("创建预算失败: %w", err)
	}

	return budget, nil
}

// CheckBudget 检查预算
func (s *BudgetService) CheckBudget(userID, projectID string) (*BudgetStatus, error) {
	// 查找活跃的预算
	var budget database.AIBudget
	err := s.db.Where("user_id = ? AND project_id = ? AND is_active = ?", userID, projectID, true).
		Where("start_date <= ? AND end_date >= ?", time.Now(), time.Now()).
		First(&budget).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &BudgetStatus{HasBudget: false}, nil
		}
		return nil, fmt.Errorf("查询预算失败: %w", err)
	}

	// 检查是否需要重置
	if s.needReset(&budget) {
		if err := s.resetBudget(&budget); err != nil {
			return nil, err
		}
	}

	// 计算使用百分比
	usagePercent := 0.0
	if budget.BudgetAmount > 0 {
		usagePercent = (budget.UsedAmount / budget.BudgetAmount) * 100
	}

	status := &BudgetStatus{
		HasBudget:     true,
		BudgetID:      budget.ID,
		BudgetAmount:  budget.BudgetAmount,
		UsedAmount:    budget.UsedAmount,
		RemainingAmount: budget.BudgetAmount - budget.UsedAmount,
		UsagePercent:  usagePercent,
		IsExceeded:    budget.UsedAmount >= budget.BudgetAmount,
		IsWarning:     usagePercent >= budget.AlertThreshold,
		BudgetType:    budget.BudgetType,
		EndDate:       budget.EndDate,
	}

	return status, nil
}

// BudgetStatus 预算状态
type BudgetStatus struct {
	HasBudget       bool      `json:"has_budget"`
	BudgetID        uint      `json:"budget_id"`
	BudgetAmount    float64   `json:"budget_amount"`
	UsedAmount      float64   `json:"used_amount"`
	RemainingAmount float64   `json:"remaining_amount"`
	UsagePercent    float64   `json:"usage_percent"`
	IsExceeded      bool      `json:"is_exceeded"`
	IsWarning       bool      `json:"is_warning"`
	BudgetType      string    `json:"budget_type"`
	EndDate         time.Time `json:"end_date"`
}

// UpdateUsedAmount 更新已使用金额
func (s *BudgetService) UpdateUsedAmount(userID, projectID string, cost float64) error {
	var budget database.AIBudget
	err := s.db.Where("user_id = ? AND project_id = ? AND is_active = ?", userID, projectID, true).
		Where("start_date <= ? AND end_date >= ?", time.Now(), time.Now()).
		First(&budget).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // 没有预算限制
		}
		return fmt.Errorf("查询预算失败: %w", err)
	}

	// 更新已使用金额
	newUsedAmount := budget.UsedAmount + cost
	if err := s.db.Model(&budget).Update("used_amount", newUsedAmount).Error; err != nil {
		return fmt.Errorf("更新预算失败: %w", err)
	}

	// 检查是否需要发送预警
	if err := s.checkAndSendAlert(&budget, newUsedAmount); err != nil {
		return err
	}

	return nil
}

// checkAndSendAlert 检查并发送预警
func (s *BudgetService) checkAndSendAlert(budget *database.AIBudget, usedAmount float64) error {
	usagePercent := 0.0
	if budget.BudgetAmount > 0 {
		usagePercent = (usedAmount / budget.BudgetAmount) * 100
	}

	// 检查是否超限
	if usedAmount >= budget.BudgetAmount {
		return s.createAlert(budget, "exceeded", "error", "预算已超限", usagePercent)
	}

	// 检查是否达到预警阈值
	if usagePercent >= budget.AlertThreshold {
		return s.createAlert(budget, "warning", "warning", fmt.Sprintf("预算使用已达%.1f%%", usagePercent), usagePercent)
	}

	return nil
}

// createAlert 创建预警
func (s *BudgetService) createAlert(budget *database.AIBudget, alertType, alertLevel, message string, usagePercent float64) error {
	alert := &database.AICostAlert{
		BudgetID:     budget.ID,
		UserID:       budget.UserID,
		ProjectID:    budget.ProjectID,
		AlertType:    alertType,
		AlertLevel:   alertLevel,
		Message:      message,
		UsagePercent: usagePercent,
		BudgetAmount: budget.BudgetAmount,
		UsedAmount:   budget.UsedAmount,
		IsRead:       false,
		IsResolved:   false,
		CreatedAt:    time.Now(),
	}

	return s.db.Create(alert).Error
}

// needReset 检查是否需要重置预算
func (s *BudgetService) needReset(budget *database.AIBudget) bool {
	now := time.Now()
	switch budget.BudgetType {
	case "daily":
		return now.Sub(budget.LastResetAt) >= 24*time.Hour
	case "weekly":
		return now.Sub(budget.LastResetAt) >= 7*24*time.Hour
	case "monthly":
		return now.Sub(budget.LastResetAt) >= 30*24*time.Hour
	}
	return false
}

// resetBudget 重置预算
func (s *BudgetService) resetBudget(budget *database.AIBudget) error {
	return s.db.Model(budget).Updates(map[string]interface{}{
		"used_amount":   0,
		"last_reset_at": time.Now(),
		"updated_at":    time.Now(),
	}).Error
}

// GetAlerts 获取预警列表
func (s *BudgetService) GetAlerts(userID, projectID string, isRead *bool, page, pageSize int) ([]database.AICostAlert, int64, error) {
	var alerts []database.AICostAlert
	var total int64

	query := s.db.Model(&database.AICostAlert{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计预警数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("查询预警列表失败: %w", err)
	}

	return alerts, total, nil
}

// MarkAlertAsRead 标记预警为已读
func (s *BudgetService) MarkAlertAsRead(alertID uint) error {
	return s.db.Model(&database.AICostAlert{}).Where("id = ?", alertID).Update("is_read", true).Error
}

// ResolveAlert 解决预警
func (s *BudgetService) ResolveAlert(alertID uint) error {
	now := time.Now()
	return s.db.Model(&database.AICostAlert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"is_resolved": true,
		"resolved_at": &now,
	}).Error
}
