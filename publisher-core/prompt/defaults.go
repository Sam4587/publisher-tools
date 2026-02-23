package prompt

import (
	"encoding/json"
	"log"
	"publisher-core/database"
	"time"

	"gorm.io/gorm"
)

// DefaultPromptTemplates 默认提示词模板
var DefaultPromptTemplates = []struct {
	TemplateID  string
	Name        string
	Type        string
	Category    string
	Description string
	Content     string
	Variables   []VariableDefinition
	IsDefault   bool
	IsSystem    bool
	Tags        []string
}{
	{
		TemplateID:  "content-generation-v1",
		Name:        "内容生成模板",
		Type:        "content_generation",
		Category:    "content",
		Description: "基于主题和关键词生成高质量内容",
		Content: `请根据以下信息生成一篇高质量的内容：

主题：{{topic}}
关键词：{{keywords}}
目标平台：{{platform}}
内容风格：{{style}}

要求：
1. 标题要吸引人，符合平台调性
2. 内容要有价值，解决用户痛点
3. 结构清晰，易于阅读
4. 适当使用emoji增加趣味性
5. 字数控制在{{word_count}}字左右

请生成完整的内容，包括标题和正文。`,
		Variables: []VariableDefinition{
			{Name: "topic", Type: "string", Required: true, Description: "内容主题"},
			{Name: "keywords", Type: "string", Required: true, Description: "关键词列表"},
			{Name: "platform", Type: "string", Required: true, Default: "抖音", Description: "目标平台"},
			{Name: "style", Type: "string", Required: false, Default: "轻松幽默", Description: "内容风格"},
			{Name: "word_count", Type: "number", Required: false, Default: "500", Description: "字数要求"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"内容生成", "AI创作", "默认模板"},
	},
	{
		TemplateID:  "content-rewrite-v1",
		Name:        "内容改写模板",
		Type:        "content_rewrite",
		Category:    "content",
		Description: "改写现有内容，保持原意的同时提升质量",
		Content: `请改写以下内容，要求：

原始内容：
{{original_content}}

改写要求：
1. 保持原文核心意思不变
2. 改善语言表达，使其更加流畅
3. 优化结构，提升可读性
4. 改写风格：{{rewrite_style}}
5. 目标平台：{{platform}}

请提供改写后的完整内容。`,
		Variables: []VariableDefinition{
			{Name: "original_content", Type: "string", Required: true, Description: "原始内容"},
			{Name: "rewrite_style", Type: "string", Required: false, Default: "简洁明了", Description: "改写风格"},
			{Name: "platform", Type: "string", Required: false, Default: "抖音", Description: "目标平台"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"内容改写", "AI优化", "默认模板"},
	},
	{
		TemplateID:  "hotspot-analysis-v1",
		Name:        "热点分析模板",
		Type:        "hotspot_analysis",
		Category:    "analysis",
		Description: "分析热点话题，提取关键信息和趋势",
		Content: `请分析以下热点话题：

热点标题：{{hotspot_title}}
热点描述：{{hotspot_description}}
热度指数：{{heat_index}}
来源平台：{{source_platform}}

分析要求：
1. 提取核心关键词（至少5个）
2. 分析热点趋势（上升/下降/稳定）
3. 评估内容创作价值（0-10分）
4. 推荐相关内容方向（至少3个）
5. 预测热点持续时间

请提供详细的分析报告。`,
		Variables: []VariableDefinition{
			{Name: "hotspot_title", Type: "string", Required: true, Description: "热点标题"},
			{Name: "hotspot_description", Type: "string", Required: false, Description: "热点描述"},
			{Name: "heat_index", Type: "number", Required: false, Default: "0", Description: "热度指数"},
			{Name: "source_platform", Type: "string", Required: false, Default: "微博", Description: "来源平台"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"热点分析", "趋势预测", "默认模板"},
	},
	{
		TemplateID:  "video-transcription-v1",
		Name:        "视频转录优化模板",
		Type:        "video_transcription",
		Category:    "video",
		Description: "优化视频转录文本，提升可读性",
		Content: `请优化以下视频转录文本：

原始转录：
{{transcription}}

视频标题：{{video_title}}
视频时长：{{duration}}

优化要求：
1. 修正语音识别错误
2. 添加标点符号和段落
3. 删除重复和无效内容
4. 保持口语化风格
5. 提取关键信息点

请提供优化后的文本和关键信息摘要。`,
		Variables: []VariableDefinition{
			{Name: "transcription", Type: "string", Required: true, Description: "原始转录文本"},
			{Name: "video_title", Type: "string", Required: false, Description: "视频标题"},
			{Name: "duration", Type: "string", Required: false, Description: "视频时长"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"视频转录", "文本优化", "默认模板"},
	},
	{
		TemplateID:  "title-generation-v1",
		Name:        "标题生成模板",
		Type:        "title_generation",
		Category:    "content",
		Description: "生成吸引人的标题",
		Content: `请根据以下内容生成吸引人的标题：

内容摘要：{{content_summary}}
目标平台：{{platform}}
目标受众：{{target_audience}}

标题要求：
1. 字数控制在{{title_length}}字以内
2. 使用吸引人的词汇
3. 符合平台调性
4. 避免标题党
5. 生成{{num_titles}}个备选标题

请提供多个标题选项。`,
		Variables: []VariableDefinition{
			{Name: "content_summary", Type: "string", Required: true, Description: "内容摘要"},
			{Name: "platform", Type: "string", Required: false, Default: "抖音", Description: "目标平台"},
			{Name: "target_audience", Type: "string", Required: false, Default: "年轻人", Description: "目标受众"},
			{Name: "title_length", Type: "number", Required: false, Default: "30", Description: "标题字数限制"},
			{Name: "num_titles", Type: "number", Required: false, Default: "5", Description: "生成标题数量"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"标题生成", "AI创作", "默认模板"},
	},
	{
		TemplateID:  "content-review-v1",
		Name:        "内容审核模板",
		Type:        "content_review",
		Category:    "review",
		Description: "审核内容是否符合平台规范",
		Content: `请审核以下内容是否符合平台规范：

待审核内容：
{{content}}
目标平台：{{platform}}

审核要点：
1. 是否包含敏感词汇
2. 是否违反平台规则
3. 是否适合目标受众
4. 是否需要修改
5. 修改建议

请提供详细的审核报告和修改建议。`,
		Variables: []VariableDefinition{
			{Name: "content", Type: "string", Required: true, Description: "待审核内容"},
			{Name: "platform", Type: "string", Required: false, Default: "抖音", Description: "目标平台"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"内容审核", "合规检查", "默认模板"},
	},
	{
		TemplateID:  "keyword-extraction-v1",
		Name:        "关键词提取模板",
		Type:        "keyword_extraction",
		Category:    "analysis",
		Description: "从内容中提取关键词",
		Content: `请从以下内容中提取关键词：

内容：
{{content}}

提取要求：
1. 提取{{num_keywords}}个核心关键词
2. 按重要性排序
3. 包含长尾关键词
4. 适合SEO优化
5. 符合平台搜索习惯

请提供关键词列表和每个关键词的重要性说明。`,
		Variables: []VariableDefinition{
			{Name: "content", Type: "string", Required: true, Description: "待提取内容"},
			{Name: "num_keywords", Type: "number", Required: false, Default: "10", Description: "关键词数量"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"关键词提取", "SEO优化", "默认模板"},
	},
	{
		TemplateID:  "content-summary-v1",
		Name:        "内容摘要模板",
		Type:        "content_summary",
		Category:    "content",
		Description: "生成内容摘要",
		Content: `请为以下内容生成摘要：

原始内容：
{{content}}

摘要要求：
1. 字数控制在{{summary_length}}字左右
2. 突出核心观点
3. 保持逻辑清晰
4. 适合快速阅读
5. 保留关键信息

请提供简洁明了的摘要。`,
		Variables: []VariableDefinition{
			{Name: "content", Type: "string", Required: true, Description: "原始内容"},
			{Name: "summary_length", Type: "number", Required: false, Default: "100", Description: "摘要字数"},
		},
		IsDefault: true,
		IsSystem:  true,
		Tags:      []string{"内容摘要", "AI总结", "默认模板"},
	},
}

// InitializeDefaultTemplates 初始化默认模板
func InitializeDefaultTemplates(db *gorm.DB) error {
	service := NewService(db)

	for _, template := range DefaultPromptTemplates {
		// 检查模板是否已存在
		var count int64
		if err := db.Model(&database.PromptTemplate{}).Where("template_id = ?", template.TemplateID).Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			log.Printf("模板 %s 已存在，跳过初始化", template.TemplateID)
			continue
		}

		// 序列化变量定义
		variablesJSON, err := json.Marshal(template.Variables)
		if err != nil {
			return err
		}

		// 序列化标签
		tagsJSON, err := json.Marshal(template.Tags)
		if err != nil {
			return err
		}

		// 创建模板
		promptTemplate := &database.PromptTemplate{
			TemplateID:  template.TemplateID,
			Name:        template.Name,
			Type:        template.Type,
			Category:    template.Category,
			Description: template.Description,
			Content:     template.Content,
			Variables:   string(variablesJSON),
			Version:     1,
			IsActive:    true,
			IsDefault:   template.IsDefault,
			IsSystem:    template.IsSystem,
			Tags:        string(tagsJSON),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(promptTemplate).Error; err != nil {
			log.Printf("创建模板 %s 失败: %v", template.TemplateID, err)
			continue
		}

		// 创建初始版本记录
		if err := service.createVersionRecord(promptTemplate, "初始版本"); err != nil {
			log.Printf("创建模板 %s 版本记录失败: %v", template.TemplateID, err)
		}

		log.Printf("成功初始化模板: %s", template.TemplateID)
	}

	return nil
}
