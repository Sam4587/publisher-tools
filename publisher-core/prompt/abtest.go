package prompt

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"publisher-core/database"

	"gorm.io/gorm"
)

// ABTestService A/B测试服务
type ABTestService struct {
	db *gorm.DB
}

// NewABTestService 创建A/B测试服务
func NewABTestService(db *gorm.DB) *ABTestService {
	return &ABTestService{db: db}
}

// CreateABTestRequest 创建A/B测试请求
type CreateABTestRequest struct {
	TestName     string `json:"test_name"`
	TemplateAID  string `json:"template_a_id"`
	TemplateBID  string `json:"template_b_id"`
	TrafficSplit int    `json:"traffic_split"` // 0-100，表示模板A的流量百分比
}

// UpdateABTestRequest 更新A/B测试请求
type UpdateABTestRequest struct {
	TrafficSplit int    `json:"traffic_split"`
	Status       string `json:"status"` // running, paused, completed
}

// CreateABTest 创建A/B测试
func (s *ABTestService) CreateABTest(req *CreateABTestRequest) (*database.PromptTemplateABTest, error) {
	// 验证流量分配
	if req.TrafficSplit < 0 || req.TrafficSplit > 100 {
		return nil, errors.New("流量分配必须在0-100之间")
	}

	// 检查模板是否存在
	templateService := NewService(s.db)
	if _, err := templateService.GetTemplate(req.TemplateAID); err != nil {
		return nil, fmt.Errorf("模板A不存在: %w", err)
	}
	if _, err := templateService.GetTemplate(req.TemplateBID); err != nil {
		return nil, fmt.Errorf("模板B不存在: %w", err)
	}

	// 检查是否已有同名测试
	var count int64
	if err := s.db.Model(&database.PromptTemplateABTest{}).Where("test_name = ? AND status = ?", req.TestName, "running").Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查测试名称失败: %w", err)
	}
	if count > 0 {
		return nil, errors.New("已存在同名的运行中测试")
	}

	abTest := &database.PromptTemplateABTest{
		TestName:     req.TestName,
		TemplateAID:  req.TemplateAID,
		TemplateBID:  req.TemplateBID,
		Status:       "running",
		TrafficSplit: req.TrafficSplit,
		StartTime:    time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(abTest).Error; err != nil {
		return nil, fmt.Errorf("创建A/B测试失败: %w", err)
	}

	return abTest, nil
}

// GetABTest 获取A/B测试
func (s *ABTestService) GetABTest(testID uint) (*database.PromptTemplateABTest, error) {
	var abTest database.PromptTemplateABTest
	if err := s.db.First(&abTest, testID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("A/B测试不存在")
		}
		return nil, fmt.Errorf("查询A/B测试失败: %w", err)
	}
	return &abTest, nil
}

// ListABTests 列出A/B测试
func (s *ABTestService) ListABTests(status string, page, pageSize int) ([]database.PromptTemplateABTest, int64, error) {
	var abTests []database.PromptTemplateABTest
	var total int64

	query := s.db.Model(&database.PromptTemplateABTest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计测试数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&abTests).Error; err != nil {
		return nil, 0, fmt.Errorf("查询测试列表失败: %w", err)
	}

	return abTests, total, nil
}

// UpdateABTest 更新A/B测试
func (s *ABTestService) UpdateABTest(testID uint, req *UpdateABTestRequest) (*database.PromptTemplateABTest, error) {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	if req.TrafficSplit >= 0 && req.TrafficSplit <= 100 {
		updates["traffic_split"] = req.TrafficSplit
	}

	if req.Status != "" {
		if req.Status == "completed" {
			now := time.Now()
			updates["status"] = "completed"
			updates["end_time"] = &now
		} else if req.Status == "paused" || req.Status == "running" {
			updates["status"] = req.Status
		}
	}

	if err := s.db.Model(abTest).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新A/B测试失败: %w", err)
	}

	return s.GetABTest(testID)
}

// SelectTemplateForTest 为A/B测试选择模板
func (s *ABTestService) SelectTemplateForTest(testID uint) (string, error) {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return "", err
	}

	if abTest.Status != "running" {
		return "", errors.New("A/B测试未运行")
	}

	// 根据流量分配随机选择模板
	if rand.Intn(100) < abTest.TrafficSplit {
		return abTest.TemplateAID, nil
	}
	return abTest.TemplateBID, nil
}

// RecordTestCall 记录测试调用
func (s *ABTestService) RecordTestCall(testID uint, templateID string, success bool, durationMs int) error {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"total_calls": abTest.TotalCalls + 1,
		"updated_at":  time.Now(),
	}

	// 更新对应模板的统计
	if templateID == abTest.TemplateAID {
		updates["calls_a"] = abTest.CallsA + 1
		if success {
			updates["success_a"] = abTest.SuccessA + 1
		}
		// 更新平均响应时间
		newAvgDurationA := (abTest.AvgDurationA*float64(abTest.CallsA) + float64(durationMs)) / float64(abTest.CallsA+1)
		updates["avg_duration_a"] = newAvgDurationA
	} else if templateID == abTest.TemplateBID {
		updates["calls_b"] = abTest.CallsB + 1
		if success {
			updates["success_b"] = abTest.SuccessB + 1
		}
		// 更新平均响应时间
		newAvgDurationB := (abTest.AvgDurationB*float64(abTest.CallsB) + float64(durationMs)) / float64(abTest.CallsB+1)
		updates["avg_duration_b"] = newAvgDurationB
	}

	return s.db.Model(abTest).Updates(updates).Error
}

// RecordUserRating 记录用户评分
func (s *ABTestService) RecordUserRating(testID uint, templateID string, rating float64) error {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return err
	}

	if rating < 0 || rating > 5 {
		return errors.New("评分必须在0-5之间")
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}

	// 更新对应模板的用户评分
	if templateID == abTest.TemplateAID {
		// 计算新的平均评分
		newRatingA := (abTest.UserRatingA*float64(abTest.CallsA) + rating) / float64(abTest.CallsA+1)
		updates["user_rating_a"] = newRatingA
	} else if templateID == abTest.TemplateBID {
		// 计算新的平均评分
		newRatingB := (abTest.UserRatingB*float64(abTest.CallsB) + rating) / float64(abTest.CallsB+1)
		updates["user_rating_b"] = newRatingB
	}

	return s.db.Model(abTest).Updates(updates).Error
}

// CompleteABTest 完成A/B测试并选择获胜者
func (s *ABTestService) CompleteABTest(testID uint, winnerTemplateID string) error {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return err
	}

	if winnerTemplateID != abTest.TemplateAID && winnerTemplateID != abTest.TemplateBID {
		return errors.New("获胜模板必须是测试中的模板之一")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":          "completed",
		"end_time":        &now,
		"winner_template": winnerTemplateID,
		"updated_at":      now,
	}

	return s.db.Model(abTest).Updates(updates).Error
}

// GetABTestStats 获取A/B测试统计信息
func (s *ABTestService) GetABTestStats(testID uint) (map[string]interface{}, error) {
	abTest, err := s.GetABTest(testID)
	if err != nil {
		return nil, err
	}

	// 计算成功率
	successRateA := 0.0
	if abTest.CallsA > 0 {
		successRateA = float64(abTest.SuccessA) / float64(abTest.CallsA) * 100
	}

	successRateB := 0.0
	if abTest.CallsB > 0 {
		successRateB = float64(abTest.SuccessB) / float64(abTest.CallsB) * 100
	}

	stats := map[string]interface{}{
		"test_id":       abTest.ID,
		"test_name":     abTest.TestName,
		"status":        abTest.Status,
		"total_calls":   abTest.TotalCalls,
		"template_a": map[string]interface{}{
			"template_id":   abTest.TemplateAID,
			"calls":         abTest.CallsA,
			"success":       abTest.SuccessA,
			"success_rate":  successRateA,
			"avg_duration":  abTest.AvgDurationA,
			"user_rating":   abTest.UserRatingA,
		},
		"template_b": map[string]interface{}{
			"template_id":   abTest.TemplateBID,
			"calls":         abTest.CallsB,
			"success":       abTest.SuccessB,
			"success_rate":  successRateB,
			"avg_duration":  abTest.AvgDurationB,
			"user_rating":   abTest.UserRatingB,
		},
		"traffic_split":  abTest.TrafficSplit,
		"winner_template": abTest.WinnerTemplate,
	}

	return stats, nil
}
