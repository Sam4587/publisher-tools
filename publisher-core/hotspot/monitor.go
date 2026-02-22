package hotspot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MonitorService 热点监控服务
type MonitorService struct {
	db          *gorm.DB
	notifyService Notifier
	config      *MonitorConfig
	alerts      map[string]*Alert
	mu          sync.RWMutex
	stopChan    chan struct{}
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	CheckInterval      time.Duration `json:"check_interval"`       // 检查间隔
	AlertThreshold     int           `json:"alert_threshold"`      // 热度预警阈值
	TrendAlertEnabled  bool          `json:"trend_alert_enabled"`  // 趋势预警开关
	KeywordAlerts      []string      `json:"keyword_alerts"`       // 关键词预警
	MaxAlertsPerHour   int           `json:"max_alerts_per_hour"`  // 每小时最大预警数
}

// DefaultMonitorConfig 默认配置
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		CheckInterval:      5 * time.Minute,
		AlertThreshold:     80,
		TrendAlertEnabled:  true,
		KeywordAlerts:      []string{},
		MaxAlertsPerHour:   10,
	}
}

// NewMonitorService 创建监控服务
func NewMonitorService(db *gorm.DB, config *MonitorConfig) *MonitorService {
	if config == nil {
		config = DefaultMonitorConfig()
	}
	return &MonitorService{
		db:       db,
		config:   config,
		alerts:   make(map[string]*Alert),
		stopChan: make(chan struct{}),
	}
}

// SetNotifyService 设置通知服务
func (s *MonitorService) SetNotifyService(notifier Notifier) {
	s.notifyService = notifier
}

// Alert 预警
type Alert struct {
	ID          string    `json:"id"`
	TopicID     string    `json:"topic_id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`        // heat, trend, keyword
	Level       string    `json:"level"`       // info, warning, critical
	Message     string    `json:"message"`
	Heat        int       `json:"heat"`
	Trend       string    `json:"trend"`
	Keywords    []string  `json:"keywords"`
	CreatedAt   time.Time `json:"created_at"`
	IsRead      bool      `json:"is_read"`
	IsResolved  bool      `json:"is_resolved"`
}

// Start 启动监控
func (s *MonitorService) Start() {
	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	logrus.Info("热点监控服务已启动")

	for {
		select {
		case <-ticker.C:
			s.check()
		case <-s.stopChan:
			logrus.Info("热点监控服务已停止")
			return
		}
	}
}

// Stop 停止监控
func (s *MonitorService) Stop() {
	close(s.stopChan)
}

// check 执行检查
func (s *MonitorService) check() {
	ctx := context.Background()

	// 1. 检查热度预警
	s.checkHeatAlerts(ctx)

	// 2. 检查趋势预警
	if s.config.TrendAlertEnabled {
		s.checkTrendAlerts(ctx)
	}

	// 3. 检查关键词预警
	if len(s.config.KeywordAlerts) > 0 {
		s.checkKeywordAlerts(ctx)
	}

	// 4. 清理过期预警
	s.cleanupOldAlerts()
}

// checkHeatAlerts 检查热度预警
func (s *MonitorService) checkHeatAlerts(ctx context.Context) {
	var topics []database.Topic
	if err := s.db.Where("heat >= ?", s.config.AlertThreshold).
		Order("heat DESC").
		Limit(20).
		Find(&topics).Error; err != nil {
		logrus.Errorf("查询高热度热点失败: %v", err)
		return
	}

	for _, topic := range topics {
		alertID := fmt.Sprintf("heat-%s", topic.ID)
		if _, exists := s.alerts[alertID]; exists {
			continue // 已存在预警
		}

		alert := &Alert{
			ID:        alertID,
			TopicID:   topic.ID,
			Title:     topic.Title,
			Type:      "heat",
			Level:     s.getHeatAlertLevel(topic.Heat),
			Message:   fmt.Sprintf("热点热度达到 %d，超过预警阈值 %d", topic.Heat, s.config.AlertThreshold),
			Heat:      topic.Heat,
			Trend:     topic.Trend,
			CreatedAt: time.Now(),
		}

		s.mu.Lock()
		s.alerts[alertID] = alert
		s.mu.Unlock()

		// 发送通知
		s.sendAlert(ctx, alert)
	}
}

// checkTrendAlerts 检查趋势预警
func (s *MonitorService) checkTrendAlerts(ctx context.Context) {
	var topics []database.Topic
	if err := s.db.Where("trend IN ?", []string{"up", "hot"}).
		Order("heat DESC").
		Limit(20).
		Find(&topics).Error; err != nil {
		logrus.Errorf("查询趋势热点失败: %v", err)
		return
	}

	for _, topic := range topics {
		alertID := fmt.Sprintf("trend-%s", topic.ID)
		if _, exists := s.alerts[alertID]; exists {
			continue
		}

		alert := &Alert{
			ID:        alertID,
			TopicID:   topic.ID,
			Title:     topic.Title,
			Type:      "trend",
			Level:     "warning",
			Message:   fmt.Sprintf("热点趋势：%s，值得关注", topic.Trend),
			Heat:      topic.Heat,
			Trend:     topic.Trend,
			CreatedAt: time.Now(),
		}

		s.mu.Lock()
		s.alerts[alertID] = alert
		s.mu.Unlock()

		s.sendAlert(ctx, alert)
	}
}

// checkKeywordAlerts 检查关键词预警
func (s *MonitorService) checkKeywordAlerts(ctx context.Context) {
	for _, keyword := range s.config.KeywordAlerts {
		var topics []database.Topic
		if err := s.db.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
			Order("heat DESC").
			Limit(10).
			Find(&topics).Error; err != nil {
			logrus.Errorf("查询关键词热点失败: %v", err)
			continue
		}

		for _, topic := range topics {
			alertID := fmt.Sprintf("keyword-%s-%s", keyword, topic.ID)
			if _, exists := s.alerts[alertID]; exists {
				continue
			}

			alert := &Alert{
				ID:        alertID,
				TopicID:   topic.ID,
				Title:     topic.Title,
				Type:      "keyword",
				Level:     "info",
				Message:   fmt.Sprintf("发现包含关键词 '%s' 的热点", keyword),
				Heat:      topic.Heat,
				Trend:     topic.Trend,
				Keywords:  []string{keyword},
				CreatedAt: time.Now(),
			}

			s.mu.Lock()
			s.alerts[alertID] = alert
			s.mu.Unlock()

			s.sendAlert(ctx, alert)
		}
	}
}

// getHeatAlertLevel 获取热度预警级别
func (s *MonitorService) getHeatAlertLevel(heat int) string {
	if heat >= 95 {
		return "critical"
	} else if heat >= 90 {
		return "warning"
	}
	return "info"
}

// sendAlert 发送预警
func (s *MonitorService) sendAlert(ctx context.Context, alert *Alert) {
	logrus.Infof("热点预警 [%s]: %s", alert.Level, alert.Message)

	if s.notifyService != nil {
		title := fmt.Sprintf("[热点预警] %s", alert.Title)
		content := fmt.Sprintf("%s\n\n热度: %d\n趋势: %s\n时间: %s",
			alert.Message, alert.Heat, alert.Trend, alert.CreatedAt.Format("2006-01-02 15:04:05"))

		if err := s.notifyService.Send(ctx, "feishu", title, content); err != nil {
			logrus.Errorf("发送预警通知失败: %v", err)
		}
	}
}

// cleanupOldAlerts 清理过期预警
func (s *MonitorService) cleanupOldAlerts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, alert := range s.alerts {
		if now.Sub(alert.CreatedAt) > 24*time.Hour {
			delete(s.alerts, id)
		}
	}
}

// GetAlerts 获取预警列表
func (s *MonitorService) GetAlerts(alertType string, limit int) []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alerts []*Alert
	for _, alert := range s.alerts {
		if alertType == "" || alert.Type == alertType {
			alerts = append(alerts, alert)
		}
	}

	// 按时间排序
	// sort.Slice(alerts, func(i, j int) bool {
	// 	return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	// })

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	return alerts
}

// MarkAlertAsRead 标记预警为已读
func (s *MonitorService) MarkAlertAsRead(alertID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alert, exists := s.alerts[alertID]; exists {
		alert.IsRead = true
	}
}

// ResolveAlert 解决预警
func (s *MonitorService) ResolveAlert(alertID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if alert, exists := s.alerts[alertID]; exists {
		alert.IsResolved = true
	}
}

// CompetitorAnalysisService 竞品分析服务
type CompetitorAnalysisService struct {
	db *gorm.DB
}

// NewCompetitorAnalysisService 创建竞品分析服务
func NewCompetitorAnalysisService(db *gorm.DB) *CompetitorAnalysisService {
	return &CompetitorAnalysisService{db: db}
}

// CompetitorAnalysis 竞品分析
type CompetitorAnalysis struct {
	TopicID          string           `json:"topic_id"`
	Title            string           `json:"title"`
	Competitors      []Competitor     `json:"competitors"`
	MarketPosition   string           `json:"market_position"`
	Opportunities    []string         `json:"opportunities"`
	Threats          []string         `json:"threats"`
	Recommendations  []string         `json:"recommendations"`
}

// Competitor 竞品信息
type Competitor struct {
	Name        string  `json:"name"`
	Heat        int     `json:"heat"`
	Trend       string  `json:"trend"`
	MarketShare float64 `json:"market_share"`
}

// AnalyzeCompetitors 分析竞品
func (s *CompetitorAnalysisService) AnalyzeCompetitors(ctx context.Context, topicID string) (*CompetitorAnalysis, error) {
	// 获取热点信息
	var topic database.Topic
	if err := s.db.Where("id = ?", topicID).First(&topic).Error; err != nil {
		return nil, fmt.Errorf("获取热点失败: %w", err)
	}

	// 查找相关竞品（基于关键词相似度）
	var relatedTopics []database.Topic
	if err := s.db.Where("id != ? AND category = ?", topicID, topic.Category).
		Order("heat DESC").
		Limit(10).
		Find(&relatedTopics).Error; err != nil {
		return nil, fmt.Errorf("查找竞品失败: %w", err)
	}

	// 构建竞品列表
	var competitors []Competitor
	for _, related := range relatedTopics {
		competitors = append(competitors, Competitor{
			Name:  related.Title,
			Heat:  related.Heat,
			Trend: related.Trend,
		})
	}

	// 分析市场位置
	marketPosition := s.analyzeMarketPosition(topic, relatedTopics)

	// 识别机会和威胁
	opportunities := s.identifyOpportunities(topic, relatedTopics)
	threats := s.identifyThreats(topic, relatedTopics)

	// 生成建议
	recommendations := s.generateRecommendations(topic, competitors)

	return &CompetitorAnalysis{
		TopicID:         topicID,
		Title:           topic.Title,
		Competitors:     competitors,
		MarketPosition:  marketPosition,
		Opportunities:   opportunities,
		Threats:         threats,
		Recommendations: recommendations,
	}, nil
}

// analyzeMarketPosition 分析市场位置
func (s *CompetitorAnalysisService) analyzeMarketPosition(topic database.Topic, competitors []database.Topic) string {
	// 简单的市场位置分析
	rank := 1
	for _, comp := range competitors {
		if comp.Heat > topic.Heat {
			rank++
		}
	}

	if rank == 1 {
		return "市场领导者"
	} else if rank <= 3 {
		return "市场挑战者"
	} else if rank <= 5 {
		return "市场追随者"
	}
	return "市场补缺者"
}

// identifyOpportunities 识别机会
func (s *CompetitorAnalysisService) identifyOpportunities(topic database.Topic, competitors []database.Topic) []string {
	var opportunities []string

	// 检查是否有上升趋势的竞品较少
	upwardCompetitors := 0
	for _, comp := range competitors {
		if comp.Trend == "up" {
			upwardCompetitors++
		}
	}

	if upwardCompetitors < 3 {
		opportunities = append(opportunities, "市场上升空间较大，竞争相对较小")
	}

	// 检查热度差距
	if len(competitors) > 0 && topic.Heat > competitors[0].Heat*0.8 {
		opportunities = append(opportunities, "与领先者差距较小，有机会超越")
	}

	return opportunities
}

// identifyThreats 识别威胁
func (s *CompetitorAnalysisService) identifyThreats(topic database.Topic, competitors []database.Topic) []string {
	var threats []string

	// 检查是否有强劲的上升竞品
	for _, comp := range competitors {
		if comp.Trend == "up" && comp.Heat > topic.Heat*0.9 {
			threats = append(threats, fmt.Sprintf("竞品 '%s' 正在快速上升", comp.Title))
		}
	}

	// 检查自身趋势
	if topic.Trend == "down" {
		threats = append(threats, "自身热度呈下降趋势")
	}

	return threats
}

// generateRecommendations 生成建议
func (s *CompetitorAnalysisService) generateRecommendations(topic database.Topic, competitors []Competitor) []string {
	var recommendations []string

	// 基于市场位置生成建议
	if topic.Heat > 80 {
		recommendations = append(recommendations, "保持内容质量，巩固市场地位")
	} else if topic.Heat > 50 {
		recommendations = append(recommendations, "加大内容投入，提升热度")
	} else {
		recommendations = append(recommendations, "寻找差异化角度，突破重围")
	}

	// 基于趋势生成建议
	if topic.Trend == "up" {
		recommendations = append(recommendations, "抓住上升趋势，快速扩大影响力")
	} else if topic.Trend == "down" {
		recommendations = append(recommendations, "分析下降原因，调整内容策略")
	}

	return recommendations
}
