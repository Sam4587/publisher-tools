package database

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// DefaultAIConfigs 返回默认的 AI 服务配置
func DefaultAIConfigs() []AIServiceConfig {
	return []AIServiceConfig{
		{
			ServiceType: "text",
			Provider:    "openrouter",
			Name:        "OpenRouter GPT-4",
			BaseURL:     "https://openrouter.ai/api/v1",
			Model:       "openai/gpt-4",
			Endpoint:    "/chat/completions",
			Priority:    100,
			IsDefault:   true,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "openrouter",
			Name:        "OpenRouter Claude 3.5 Sonnet",
			BaseURL:     "https://openrouter.ai/api/v1",
			Model:       "anthropic/claude-3.5-sonnet",
			Endpoint:    "/chat/completions",
			Priority:    95,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "groq",
			Name:        "Groq Llama 3.3 70B",
			BaseURL:     "https://api.groq.com/openai/v1",
			Model:       "llama-3.3-70b-versatile",
			Endpoint:    "/chat/completions",
			Priority:    90,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "groq",
			Name:        "Groq Mixtral 8x7B",
			BaseURL:     "https://api.groq.com/openai/v1",
			Model:       "mixtral-8x7b-32768",
			Endpoint:    "/chat/completions",
			Priority:    85,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "google",
			Name:        "Google Gemini Flash",
			BaseURL:     "https://generativelanguage.googleapis.com/v1beta",
			Model:       "gemini-2.0-flash",
			Endpoint:    "/models/{model}:generateContent",
			Priority:    80,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "deepseek",
			Name:        "DeepSeek Chat",
			BaseURL:     "https://api.deepseek.com",
			Model:       "deepseek-chat",
			Endpoint:    "/chat/completions",
			Priority:    75,
			IsDefault:   false,
			IsActive:    true,
		},
		{
			ServiceType: "text",
			Provider:    "deepseek",
			Name:        "DeepSeek Reasoner",
			BaseURL:     "https://api.deepseek.com",
			Model:       "deepseek-reasoner",
			Endpoint:    "/chat/completions",
			Priority:    70,
			IsDefault:   false,
			IsActive:    true,
		},
	}
}

// DefaultPlatforms 返回默认的平台配置
func DefaultPlatforms() []Platform {
	return []Platform{
		{ID: "weibo", Name: "微博", IsActive: true},
		{ID: "zhihu", Name: "知乎", IsActive: true},
		{ID: "baidu", Name: "百度", IsActive: true},
		{ID: "douyin", Name: "抖音", IsActive: true},
		{ID: "toutiao", Name: "今日头条", IsActive: true},
		{ID: "xiaohongshu", Name: "小红书", IsActive: true},
		{ID: "bilibili", Name: "B站", IsActive: true},
		{ID: "newsnow", Name: "NewsNow", IsActive: true},
	}
}

// DefaultNotificationTemplates 返回默认的通知模板
func DefaultNotificationTemplates() []NotificationTemplate {
	return []NotificationTemplate{
		{
			Name:  "hotspot_alert",
			Title: "热点提醒",
			Body: `【热点提醒】
发现新热点：{{.Title}}
热度：{{.Heat}}
趋势：{{.Trend}}
来源：{{.Source}}
时间：{{.Time}}
链接：{{.URL}}`,
		},
		{
			Name:  "task_completed",
			Title: "任务完成",
			Body: `【任务完成】
任务ID：{{.TaskID}}
类型：{{.Type}}
状态：{{.Status}}
耗时：{{.Duration}}
结果：{{.Result}}`,
		},
		{
			Name:  "task_failed",
			Title: "任务失败",
			Body: `【任务失败】
任务ID：{{.TaskID}}
类型：{{.Type}}
错误：{{.Error}}
时间：{{.Time}}`,
		},
		{
			Name:  "daily_report",
			Title: "每日报告",
			Body: `【每日报告】
日期：{{.Date}}
新增热点：{{.NewTopics}}
完成任务：{{.CompletedTasks}}
失败任务：{{.FailedTasks}}
AI 调用：{{.AICalls}}`,
		},
	}
}

// SeedDefaultData 填充默认数据
func SeedDefaultData(db *gorm.DB) error {
	// 填充 AI 服务配置
	var count int64
	db.Model(&AIServiceConfig{}).Count(&count)
	if count == 0 {
		configs := DefaultAIConfigs()
		for i := range configs {
			configs[i].CreatedAt = time.Now()
			configs[i].UpdatedAt = time.Now()
		}
		if err := db.Create(&configs).Error; err != nil {
			return err
		}
	}

	// 填充平台配置
	db.Model(&Platform{}).Count(&count)
	if count == 0 {
		platforms := DefaultPlatforms()
		for i := range platforms {
			platforms[i].CreatedAt = time.Now()
			platforms[i].UpdatedAt = time.Now()
		}
		if err := db.Create(&platforms).Error; err != nil {
			return err
		}
	}

	// 填充通知模板
	db.Model(&NotificationTemplate{}).Count(&count)
	if count == 0 {
		templates := DefaultNotificationTemplates()
		for i := range templates {
			templates[i].CreatedAt = time.Now()
			templates[i].UpdatedAt = time.Now()
		}
		if err := db.Create(&templates).Error; err != nil {
			return err
		}
	}

	return nil
}

// SettingsToJSON 将设置转换为 JSON 字符串
func SettingsToJSON(settings map[string]interface{}) (string, error) {
	if settings == nil {
		return "", nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SettingsFromJSON 从 JSON 字符串解析设置
func SettingsFromJSON(settings string) (map[string]interface{}, error) {
	if settings == "" {
		return make(map[string]interface{}), nil
	}
	var result map[string]interface{}
	err := json.Unmarshal([]byte(settings), &result)
	return result, err
}

// KeywordsToJSON 将关键词数组转换为 JSON 字符串
func KeywordsToJSON(keywords []string) (string, error) {
	if keywords == nil {
		return "[]", nil
	}
	data, err := json.Marshal(keywords)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// KeywordsFromJSON 从 JSON 字符串解析关键词数组
func KeywordsFromJSON(keywords string) ([]string, error) {
	if keywords == "" {
		return []string{}, nil
	}
	var result []string
	err := json.Unmarshal([]byte(keywords), &result)
	return result, err
}
