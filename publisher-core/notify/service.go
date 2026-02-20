package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"publisher-core/database"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service 通知服务
type Service struct {
	db       *gorm.DB
	channels map[string]Notifier
	mu       sync.RWMutex
}

// Notifier 通知器接口
type Notifier interface {
	Send(ctx context.Context, message *Message) error
	GetMaxSize() int
	GetName() string
}

// Message 通知消息
type Message struct {
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NewService 创建通知服务
func NewService(db *gorm.DB) *Service {
	s := &Service{
		db:       db,
		channels: make(map[string]Notifier),
	}

	// 注册默认通知器
	s.RegisterChannel("feishu", &FeishuNotifier{})
	s.RegisterChannel("dingtalk", &DingTalkNotifier{})
	s.RegisterChannel("wecom", &WeComNotifier{})
	s.RegisterChannel("telegram", &TelegramNotifier{})
	s.RegisterChannel("email", &EmailNotifier{})

	return s
}

// RegisterChannel 注册通知渠道
func (s *Service) RegisterChannel(channelType string, notifier Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channelType] = notifier
	logrus.Infof("Registered notify channel: %s", channelType)
}

// Send 发送通知
func (s *Service) Send(ctx context.Context, channelType string, message *Message) error {
	s.mu.RLock()
	notifier, ok := s.channels[channelType]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("channel %s not found", channelType)
	}

	// 获取渠道配置
	config, err := s.getChannelConfig(channelType)
	if err != nil {
		return err
	}

	// 设置配置
	if setter, ok := notifier.(ConfigSetter); ok {
		setter.SetConfig(config)
	}

	// 检查消息大小
	maxSize := notifier.GetMaxSize()
	if len(message.Content) > maxSize {
		// 分批发送
		return s.sendInBatches(ctx, notifier, message, maxSize)
	}

	return notifier.Send(ctx, message)
}

// sendInBatches 分批发送
func (s *Service) sendInBatches(ctx context.Context, notifier Notifier, message *Message, maxSize int) error {
	content := message.Content
	batchSize := maxSize - 200 // 预留头部空间
	batches := splitContent(content, batchSize)
	total := len(batches)

	for i, batch := range batches {
		batchMsg := &Message{
			Title: fmt.Sprintf("%s [%d/%d]", message.Title, i+1, total),
			Content: batch,
			Data: message.Data,
		}

		if err := notifier.Send(ctx, batchMsg); err != nil {
			return fmt.Errorf("batch %d/%d failed: %w", i+1, total, err)
		}

		// 批次间间隔
		if i < total-1 {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

// getChannelConfig 获取渠道配置
func (s *Service) getChannelConfig(channelType string) (*database.NotificationChannel, error) {
	var config database.NotificationChannel
	err := s.db.Where("type = ? AND is_active = ?", channelType, true).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("channel %s not configured", channelType)
	}
	return &config, err
}

// SendWithTemplate 使用模板发送通知
func (s *Service) SendWithTemplate(ctx context.Context, channelType, templateName string, data map[string]interface{}) error {
	// 获取模板
	var tmpl database.NotificationTemplate
	err := s.db.Where("name = ?", templateName).First(&tmpl).Error
	if err != nil {
		return fmt.Errorf("template %s not found: %w", templateName, err)
	}

	// 渲染模板
	title, content, err := s.renderTemplate(&tmpl, data)
	if err != nil {
		return err
	}

	return s.Send(ctx, channelType, &Message{
		Title:   title,
		Content: content,
		Data:    data,
	})
}

// renderTemplate 渲染模板
func (s *Service) renderTemplate(tmpl *database.NotificationTemplate, data map[string]interface{}) (string, string, error) {
	// 解析标题模板
	titleTmpl, err := template.New("title").Parse(tmpl.Title)
	if err != nil {
		return "", "", fmt.Errorf("parse title template: %w", err)
	}

	var titleBuf bytes.Buffer
	if err := titleTmpl.Execute(&titleBuf, data); err != nil {
		return "", "", fmt.Errorf("render title: %w", err)
	}

	// 解析内容模板
	contentTmpl, err := template.New("content").Parse(tmpl.Body)
	if err != nil {
		return "", "", fmt.Errorf("parse content template: %w", err)
	}

	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", "", fmt.Errorf("render content: %w", err)
	}

	return titleBuf.String(), contentBuf.String(), nil
}

// SendToAllChannels 发送到所有活跃渠道
func (s *Service) SendToAllChannels(ctx context.Context, message *Message) map[string]error {
	var channels []database.NotificationChannel
	s.db.Where("is_active = ?", true).Find(&channels)

	results := make(map[string]error)
	for _, ch := range channels {
		err := s.Send(ctx, ch.Type, message)
		results[ch.Type] = err
		if err != nil {
			logrus.Warnf("Failed to send to %s: %v", ch.Type, err)
		}
	}

	return results
}

// CreateChannel 创建通知渠道
func (s *Service) CreateChannel(channel *database.NotificationChannel) error {
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()
	return s.db.Create(channel).Error
}

// UpdateChannel 更新通知渠道
func (s *Service) UpdateChannel(channel *database.NotificationChannel) error {
	channel.UpdatedAt = time.Now()
	return s.db.Save(channel).Error
}

// DeleteChannel 删除通知渠道
func (s *Service) DeleteChannel(id uint) error {
	return s.db.Delete(&database.NotificationChannel{}, id).Error
}

// ListChannels 列出所有渠道
func (s *Service) ListChannels() ([]database.NotificationChannel, error) {
	var channels []database.NotificationChannel
	err := s.db.Find(&channels).Error
	return channels, err
}

// CreateTemplate 创建通知模板
func (s *Service) CreateTemplate(tmpl *database.NotificationTemplate) error {
	tmpl.CreatedAt = time.Now()
	tmpl.UpdatedAt = time.Now()
	return s.db.Create(tmpl).Error
}

// ListTemplates 列出所有模板
func (s *Service) ListTemplates() ([]database.NotificationTemplate, error) {
	var templates []database.NotificationTemplate
	err := s.db.Find(&templates).Error
	return templates, err
}

// ConfigSetter 配置设置接口
type ConfigSetter interface {
	SetConfig(config *database.NotificationChannel)
}

// =====================================================
// 飞书通知器
// =====================================================

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	webhook string
}

// SetConfig 设置配置
func (n *FeishuNotifier) SetConfig(config *database.NotificationChannel) {
	n.webhook = config.Webhook
}

func (n *FeishuNotifier) GetName() string {
	return "feishu"
}

func (n *FeishuNotifier) GetMaxSize() int {
	return 30000 // 飞书限制 30KB
}

func (n *FeishuNotifier) Send(ctx context.Context, message *Message) error {
	if n.webhook == "" {
		return fmt.Errorf("feishu webhook not configured")
	}

	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("%s\n\n%s", message.Title, message.Content),
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", n.webhook, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu API error: %d", resp.StatusCode)
	}

	return nil
}

// =====================================================
// 钉钉通知器
// =====================================================

// DingTalkNotifier 钉钉通知器
type DingTalkNotifier struct {
	webhook string
}

func (n *DingTalkNotifier) SetConfig(config *database.NotificationChannel) {
	n.webhook = config.Webhook
}

func (n *DingTalkNotifier) GetName() string {
	return "dingtalk"
}

func (n *DingTalkNotifier) GetMaxSize() int {
	return 20000 // 钉钉限制 20KB
}

func (n *DingTalkNotifier) Send(ctx context.Context, message *Message) error {
	if n.webhook == "" {
		return fmt.Errorf("dingtalk webhook not configured")
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", message.Title, message.Content),
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", n.webhook, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API error: %d", resp.StatusCode)
	}

	return nil
}

// =====================================================
// 企业微信通知器
// =====================================================

// WeComNotifier 企业微信通知器
type WeComNotifier struct {
	webhook string
}

func (n *WeComNotifier) SetConfig(config *database.NotificationChannel) {
	n.webhook = config.Webhook
}

func (n *WeComNotifier) GetName() string {
	return "wecom"
}

func (n *WeComNotifier) GetMaxSize() int {
	return 4096 // 企业微信限制 4KB
}

func (n *WeComNotifier) Send(ctx context.Context, message *Message) error {
	if n.webhook == "" {
		return fmt.Errorf("wecom webhook not configured")
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("%s\n\n%s", message.Title, message.Content),
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", n.webhook, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom API error: %d", resp.StatusCode)
	}

	return nil
}

// =====================================================
// Telegram 通知器
// =====================================================

// TelegramNotifier Telegram 通知器
type TelegramNotifier struct {
	botToken string
	chatID   string
}

func (n *TelegramNotifier) SetConfig(config *database.NotificationChannel) {
	// 从 config.Config 解析 bot_token 和 chat_id
	var cfg struct {
		BotToken string `json:"bot_token"`
		ChatID   string `json:"chat_id"`
	}
	if config.Config != "" {
		json.Unmarshal([]byte(config.Config), &cfg)
	}
	n.botToken = cfg.BotToken
	n.chatID = cfg.ChatID
}

func (n *TelegramNotifier) GetName() string {
	return "telegram"
}

func (n *TelegramNotifier) GetMaxSize() int {
	return 4096 // Telegram 限制 4KB
}

func (n *TelegramNotifier) Send(ctx context.Context, message *Message) error {
	if n.botToken == "" || n.chatID == "" {
		return fmt.Errorf("telegram not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	payload := map[string]interface{}{
		"chat_id": n.chatID,
		"text":    fmt.Sprintf("*%s*\n\n%s", message.Title, message.Content),
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: %d", resp.StatusCode)
	}

	return nil
}

// =====================================================
// 邮件通知器
// =====================================================

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	from         string
	to           []string
}

func (n *EmailNotifier) SetConfig(config *database.NotificationChannel) {
	var cfg struct {
		SMTPHost     string   `json:"smtp_host"`
		SMTPPort     int      `json:"smtp_port"`
		SMTPUser     string   `json:"smtp_user"`
		SMTPPassword string   `json:"smtp_password"`
		From         string   `json:"from"`
		To           []string `json:"to"`
	}
	if config.Config != "" {
		json.Unmarshal([]byte(config.Config), &cfg)
	}
	n.smtpHost = cfg.SMTPHost
	n.smtpPort = cfg.SMTPPort
	n.smtpUser = cfg.SMTPUser
	n.smtpPassword = cfg.SMTPPassword
	n.from = cfg.From
	n.to = cfg.To
}

func (n *EmailNotifier) GetName() string {
	return "email"
}

func (n *EmailNotifier) GetMaxSize() int {
	return 1000000 // 邮件限制 1MB
}

func (n *EmailNotifier) Send(ctx context.Context, message *Message) error {
	// 邮件发送需要额外的 SMTP 库支持
	// 这里仅作为占位实现
	return fmt.Errorf("email notifier not implemented, please use a proper email library")
}

// =====================================================
// 辅助函数
// =====================================================

// splitContent 分割内容
func splitContent(content string, maxSize int) []string {
	if len(content) <= maxSize {
		return []string{content}
	}

	var batches []string
	lines := strings.Split(content, "\n")
	var currentBatch strings.Builder

	for _, line := range lines {
		if currentBatch.Len()+len(line)+1 > maxSize {
			batches = append(batches, currentBatch.String())
			currentBatch.Reset()
		}
		currentBatch.WriteString(line)
		currentBatch.WriteString("\n")
	}

	if currentBatch.Len() > 0 {
		batches = append(batches, currentBatch.String())
	}

	return batches
}
