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

type ContentTemplate struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Platform    string             `json:"platform"`
	Category    string             `json:"category"`
	Template    string             `json:"template"`
	Variables   []TemplateVariable `json:"variables"`
	Examples    []string           `json:"examples"`
	Tags        []string           `json:"tags"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	UsageCount  int                `json:"usage_count"`
	Rating      float64            `json:"rating"`
}

type TemplateVariable struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     string   `json:"default"`
	Options     []string `json:"options,omitempty"`
}

type TemplateManager struct {
	mu        sync.RWMutex
	templates map[string]*ContentTemplate
	storage   TemplateStorage
}

type TemplateStorage interface {
	Save(template *ContentTemplate) error
	Load(id string) (*ContentTemplate, error)
	List(filter TemplateFilter) ([]*ContentTemplate, error)
	Delete(id string) error
}

type TemplateFilter struct {
	Platform string
	Category string
	Tags     []string
	Limit    int
}

func NewTemplateManager(storage TemplateStorage) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*ContentTemplate),
		storage:   storage,
	}
	tm.loadDefaults()
	return tm
}

func (tm *TemplateManager) loadDefaults() {
	defaults := []*ContentTemplate{
		{
			ID:          "news-hotspot",
			Name:        "Hotspot News Commentary",
			Description: "Generate commentary content for hotspot events",
			Platform:    "all",
			Category:    "news",
			Template:    "[{title}] {event}\n\n{comment}\n\n#hotspot #{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "Title", Type: "text", Required: true},
				{Name: "event", Description: "Event description", Type: "text", Required: true},
				{Name: "comment", Description: "Comment content", Type: "text", Required: true},
				{Name: "tags", Description: "Topic tags", Type: "text", Required: false},
			},
			Tags:      []string{"hotspot", "news", "commentary"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "tutorial-guide",
			Name:        "Tutorial Guide",
			Description: "Generate tutorial content",
			Platform:    "xiaohongshu",
			Category:    "tutorial",
			Template:    "[{title}]\n\n{intro}\n\nSteps: {steps}\n\nTips: {tips}\n\n#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "Tutorial title", Type: "text", Required: true},
				{Name: "intro", Description: "Introduction", Type: "text", Required: true},
				{Name: "steps", Description: "Step instructions", Type: "text", Required: true},
				{Name: "tips", Description: "Tips", Type: "text", Required: false},
				{Name: "tags", Description: "Topic tags", Type: "text", Required: false},
			},
			Tags:      []string{"tutorial", "guide", "tips"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "lifestyle-share",
			Name:        "Lifestyle Sharing",
			Description: "Lifestyle content sharing template",
			Platform:    "xiaohongshu",
			Category:    "lifestyle",
			Template:    "[{title}]\n\n{content}\n\nThoughts: {thoughts}\n\nLocation: {location}\n\n#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "Title", Type: "text", Required: true},
				{Name: "content", Description: "Content", Type: "text", Required: true},
				{Name: "thoughts", Description: "Thoughts", Type: "text", Required: false},
				{Name: "location", Description: "Location", Type: "text", Required: false},
				{Name: "tags", Description: "Topic tags", Type: "text", Required: false},
			},
			Tags:      []string{"lifestyle", "sharing", "daily"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "entertainment-review",
			Name:        "Entertainment Review",
			Description: "Entertainment content review template",
			Platform:    "douyin",
			Category:    "entertainment",
			Template:    "[{title}]\n\nOverview: {overview}\n\nPros: {pros}\n\nCons: {cons}\n\nPrice: {price}\n\nSummary: {summary}\n\n#{tags}",
			Variables: []TemplateVariable{
				{Name: "title", Description: "Review title", Type: "text", Required: true},
				{Name: "overview", Description: "Overview", Type: "text", Required: true},
				{Name: "pros", Description: "Pros", Type: "text", Required: true},
				{Name: "cons", Description: "Cons", Type: "text", Required: true},
				{Name: "price", Description: "Price", Type: "text", Required: false},
				{Name: "summary", Description: "Summary", Type: "text", Required: true},
				{Name: "tags", Description: "Topic tags", Type: "text", Required: false},
			},
			Tags:      []string{"review", "entertainment", "recommendation"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, t := range defaults {
		tm.templates[t.ID] = t
	}
}

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

func (tm *TemplateManager) GetTemplate(id string) (*ContentTemplate, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

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

func (tm *TemplateManager) DeleteTemplate(id string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.templates, id)

	if tm.storage != nil {
		return tm.storage.Delete(id)
	}
	return nil
}

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

	tm.mu.Lock()
	template.UsageCount++
	tm.mu.Unlock()

	return result, nil
}

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
