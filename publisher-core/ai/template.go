package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ContentTemplate 内容模板
type ContentTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Platform    string                 `json:"platform"`
	Category    string                 `json:"category"` // 新闻、教程、生活、娱乐等
	Template    string                 `json:"template"`
	Variables   []TemplateVariable     `json:"variables"`
	Examples    []string               `json:"examples"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UsageCount  int                    `json:"usage_count"`
	Rating      float64                `json:"rating"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // text, number, select, multiselect
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	Options     []string `json:"options,omitempty"`
}

// TemplateManager 模板管理�?
type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*ContentTemplate
	storage   TemplateStorage
}

// TemplateStorage 模板存储接口
type TemplateStorage interface {
	Save(template *ContentTemplate) error
	Load(id string) (*ContentTemplate, error)
	List(filter TemplateFilter) ([]*ContentTemplate, error)
	Delete(id string) error
}

// TemplateFilter 模板过滤�?
type TemplateFilter struct {
	Platform string
	Category string
	Tags     []string
	Limit    int
}

// NewTemplateManager 创建模板管理�?
func NewTemplateManager(storage TemplateStorage) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*ContentTemplate),
		storage:   storage,
	}
	tm.loadDefaults()
	return tm
}

// loadDefaults 加载默认模板
func (tm *TemplateManager) loadDefaults() {
	defaults := []*ContentTemplate{
		{
			ID:          "news-hotspot",
			Name:        "热点新闻评论",
			Description: "针对热点事件生成评论性内�?,
			Platform:    "all",
			Category:    "新闻",
			Template:    "【{title}】{event}

{comment}

#热点解读 #{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "标题", Type: "text", Required: true},
				{Name: "event", Description: "事件描述", Type: "text", Required: true},
				{Name: "comment", Description: "评论内容", Type: "text", Required: true},
				{Name: "tags", Description: "话题标签", Type: "text", Required: false},
			},
			Tags:      []string{"热点", "新闻", "评论"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "tutorial-guide",
			Name:        "教程指南",
			Description: "生成教程类内�?,
			Platform:    "xiaohongshu",
			Category:    "教程",
			Template:    "【{title}�?

�?{intro}

📝 {steps}

💡 {tips}

#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "教程标题", Type: "text", Required: true},
				{Name: "intro", Description: "简�?, Type: "text", Required: true},
				{Name: "steps", Description: "步骤说明", Type: "text", Required: true},
				{Name: "tips", Description: "小贴�?, Type: "text", Required: false},
				{Name: "tags", Description: "话题标签", Type: "text", Required: false},
			},
			Tags:      []string{"教程", "指南", "干货"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "lifestyle-share",
			Name:        "生活分享",
			Description: "生活类内容分享模�?,
			Platform:    "xiaohongshu",
			Category:    "生活",
			Template:    "【{title}�?

{content}

💭 {thoughts}

📍 {location}

#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "标题", Type: "text", Required: true},
				{Name: "content", Description: "内容", Type: "text", Required: true},
				{Name: "thoughts", Description: "感悟", Type: "text", Required: false},
				{Name: "location", Description: "地点", Type: "text", Required: false},
				{Name: "tags", Description: "话题标签", Type: "text", Required: false},
			},
			Tags:      []string{"生活", "分享", "日常"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "entertainment-review",
			Name:        "娱乐测评",
			Description: "娱乐类内容测评模�?,
			Platform:    "douyin",
			Category:    "娱乐",
			Template:    "【{title}�?

🎯 {overview}

�?优点：{pros}

�?缺点：{cons}

💰 价格：{price}

💭 总结：{summary}

#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "测评标题", Type: "text", Required: true},
				{Name: "overview", Description: "概述", Type: "text", Required: true},
				{Name: "pros", Description: "优点", Type: "text", Required: true},
				{Name: "cons", Description: "缺点", Type: "text", Required: true},
				{Name: "price", Description: "价格", Type: "text", Required: false},
				{Name: "summary", Description: "总结", Type: "text", Required: true},
				{Name: "tags", Description: "话题标签", Type: "text", Required: false},
			},
			Tags:      []string{"测评", "娱乐", "推荐"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, t := range defaults {
		tm.templates[t.ID] = t
	}
}

// CreateTemplate 创建模板
func (tm *TemplateManager) CreateTemplate(template *ContentTemplate) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	tm.templates[template.ID] = template

	if tm.storage != nil {
		return tm.storage.Save(template)
	}
	return nil
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(id string) (*ContentTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates 列出模板
func (tm *TemplateManager) ListTemplates(filter TemplateFilter) []*ContentTemplate {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*ContentTemplate, 0)
	for _, t := range tm.templates {
		if filter.Platform != "" && filter.Platform != "all" && t.Platform != "all" && t.Platform != filter.Platform {
			continue
		}
		if filter.Category != "" && t.Category != filter.Category {
			continue
		}
		result = append(result, t)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result
}

// UpdateTemplate 更新模板
func (tm *TemplateManager) UpdateTemplate(template *ContentTemplate) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.templates[template.ID]; !exists {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	template.UpdatedAt = time.Now()
	tm.templates[template.ID] = template

	if tm.storage != nil {
		return tm.storage.Save(template)
	}
	return nil
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.templates, id)

	if tm.storage != nil {
		return tm.storage.Delete(id)
	}
	return nil
}

// ApplyTemplate 应用模板
func (tm *TemplateManager) ApplyTemplate(templateID string, values map[string]string) (string, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return "", err
	}

	result := template.Template
	for _, v := range template.Variables {
		value, ok := values[v.Name]
		if !ok {
			if v.Required {
				return "", fmt.Errorf("missing required variable: %s", v.Name)
			}
			value = v.Default
		}
		result = replaceAll(result, "{"+v.Name+"}", value)
	}

	// 更新使用次数
	tm.mu.Lock()
	template.UsageCount++
	tm.mu.Unlock()

	return result, nil
}

// JSONTemplateStorage JSON文件存储实现
type JSONTemplateStorage struct {
	dataDir string
}

func NewJSONTemplateStorage(dataDir string) (*JSONTemplateStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &JSONTemplateStorage{dataDir: dataDir}, nil
}

func (s *JSONTemplateStorage) Save(template *ContentTemplate) error {
	path := filepath.Join(s.dataDir, template.ID+".json")
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *JSONTemplateStorage) Load(id string) (*ContentTemplate, error) {
	path := filepath.Join(s.dataDir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var template ContentTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}
	return &template, nil
}

func (s *JSONTemplateStorage) List(filter TemplateFilter) ([]*ContentTemplate, error) {
	files, err := filepath.Glob(filepath.Join(s.dataDir, "*.json"))
	if err != nil {
		return nil, err
	}

	templates := make([]*ContentTemplate, 0)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var t ContentTemplate
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		templates = append(templates, &t)
	}

	return templates, nil
}

func (s *JSONTemplateStorage) Delete(id string) error {
	path := filepath.Join(s.dataDir, id+".json")
	return os.Remove(path)
}

func replaceAll(s, old, new string) string {
	for {
		replaced := replaceFirst(s, old, new)
		if replaced == s {
			return s
		}
		s = replaced
	}
}

func replaceFirst(s, old, new string) string {
	idx := indexOf(s, old)
	if idx == -1 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
