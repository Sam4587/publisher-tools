package prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"publisher-tools/database"

	"gorm.io/gorm"
)

// Service 提示词模板服务
type Service struct {
	db *gorm.DB
}

// NewService 创建新的提示词模板服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// VariableDefinition 变量定义
type VariableDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`        // string, number, boolean, array
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

// TemplateInput 模板输入
type TemplateInput struct {
	TemplateID string                 `json:"template_id"`
	Variables  map[string]interface{} `json:"variables"`
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	TemplateID  string               `json:"template_id"`
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Category    string               `json:"category"`
	Description string               `json:"description"`
	Content     string               `json:"content"`
	Variables   []VariableDefinition `json:"variables"`
	IsDefault   bool                 `json:"is_default"`
	Tags        []string             `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Category    string               `json:"category"`
	Description string               `json:"description"`
	Content     string               `json:"content"`
	Variables   []VariableDefinition `json:"variables"`
	IsActive    bool                 `json:"is_active"`
	IsDefault   bool                 `json:"is_default"`
	Tags        []string             `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	ChangeNote  string               `json:"change_note"`
}

// CreateTemplate 创建提示词模板
func (s *Service) CreateTemplate(req *CreateTemplateRequest) (*database.PromptTemplate, error) {
	// 检查模板ID是否已存在
	var count int64
	if err := s.db.Model(&database.PromptTemplate{}).Where("template_id = ?", req.TemplateID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查模板ID失败: %w", err)
	}
	if count > 0 {
		return nil, errors.New("模板ID已存在")
	}

	// 序列化变量定义
	variablesJSON, err := json.Marshal(req.Variables)
	if err != nil {
		return nil, fmt.Errorf("序列化变量定义失败: %w", err)
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("序列化标签失败: %w", err)
	}

	// 序列化元数据
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("序列化元数据失败: %w", err)
	}

	template := &database.PromptTemplate{
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Content:     req.Content,
		Variables:   string(variablesJSON),
		Version:     1,
		IsActive:    true,
		IsDefault:   req.IsDefault,
		IsSystem:    false,
		Tags:        string(tagsJSON),
		Metadata:    string(metadataJSON),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(template).Error; err != nil {
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	// 创建初始版本记录
	if err := s.createVersionRecord(template, "初始版本"); err != nil {
		return nil, fmt.Errorf("创建版本记录失败: %w", err)
	}

	return template, nil
}

// GetTemplate 获取模板
func (s *Service) GetTemplate(templateID string) (*database.PromptTemplate, error) {
	var template database.PromptTemplate
	if err := s.db.Where("template_id = ?", templateID).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("模板不存在")
		}
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	return &template, nil
}

// ListTemplates 列出模板
func (s *Service) ListTemplates(templateType, category string, isActive *bool, page, pageSize int) ([]database.PromptTemplate, int64, error) {
	var templates []database.PromptTemplate
	var total int64

	query := s.db.Model(&database.PromptTemplate{})

	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计模板数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, fmt.Errorf("查询模板列表失败: %w", err)
	}

	return templates, total, nil
}

// UpdateTemplate 更新模板
func (s *Service) UpdateTemplate(templateID string, req *UpdateTemplateRequest) (*database.PromptTemplate, error) {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// 检查是否为系统模板
	if template.IsSystem {
		return nil, errors.New("系统模板不可修改")
	}

	// 序列化变量定义
	variablesJSON, err := json.Marshal(req.Variables)
	if err != nil {
		return nil, fmt.Errorf("序列化变量定义失败: %w", err)
	}

	// 序列化标签
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("序列化标签失败: %w", err)
	}

	// 序列化元数据
	metadataJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return nil, fmt.Errorf("序列化元数据失败: %w", err)
	}

	// 更新模板
	updates := map[string]interface{}{
		"name":        req.Name,
		"type":        req.Type,
		"category":    req.Category,
		"description": req.Description,
		"content":     req.Content,
		"variables":   string(variablesJSON),
		"is_active":   req.IsActive,
		"is_default":  req.IsDefault,
		"tags":        string(tagsJSON),
		"metadata":    string(metadataJSON),
		"version":     template.Version + 1,
		"updated_at":  time.Now(),
	}

	if err := s.db.Model(template).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新模板失败: %w", err)
	}

	// 创建版本记录
	changeNote := req.ChangeNote
	if changeNote == "" {
		changeNote = fmt.Sprintf("更新至版本 %d", template.Version+1)
	}
	if err := s.createVersionRecord(template, changeNote); err != nil {
		return nil, fmt.Errorf("创建版本记录失败: %w", err)
	}

	return s.GetTemplate(templateID)
}

// DeleteTemplate 删除模板
func (s *Service) DeleteTemplate(templateID string) error {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return err
	}

	// 检查是否为系统模板
	if template.IsSystem {
		return errors.New("系统模板不可删除")
	}

	// 软删除
	if err := s.db.Delete(template).Error; err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}

	return nil
}

// RenderTemplate 渲染模板（替换变量）
func (s *Service) RenderTemplate(templateID string, variables map[string]interface{}) (string, error) {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return "", err
	}

	// 检查模板是否激活
	if !template.IsActive {
		return "", errors.New("模板未激活")
	}

	// 解析变量定义
	var varDefs []VariableDefinition
	if template.Variables != "" {
		if err := json.Unmarshal([]byte(template.Variables), &varDefs); err != nil {
			return "", fmt.Errorf("解析变量定义失败: %w", err)
		}
	}

	// 验证必需变量
	for _, varDef := range varDefs {
		if varDef.Required {
			if _, ok := variables[varDef.Name]; !ok {
				if varDef.Default == "" {
					return "", fmt.Errorf("缺少必需变量: %s", varDef.Name)
				}
				variables[varDef.Name] = varDef.Default
			}
		}
	}

	// 替换变量
	content := template.Content
	for varName, varValue := range variables {
		placeholder := fmt.Sprintf("{{%s}}", varName)
		valueStr := fmt.Sprintf("%v", varValue)
		content = strings.ReplaceAll(content, placeholder, valueStr)
	}

	// 检查是否还有未替换的变量
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		unreplacedVars := make([]string, 0)
		for _, match := range matches {
			unreplacedVars = append(unreplacedVars, match[1])
		}
		return "", fmt.Errorf("存在未替换的变量: %s", strings.Join(unreplacedVars, ", "))
	}

	return content, nil
}

// createVersionRecord 创建版本记录
func (s *Service) createVersionRecord(template *database.PromptTemplate, changeNote string) error {
	version := &database.PromptTemplateVersion{
		TemplateID: template.TemplateID,
		Version:    template.Version,
		Content:    template.Content,
		Variables:  template.Variables,
		ChangeNote: changeNote,
		CreatedAt:  time.Now(),
	}

	return s.db.Create(version).Error
}

// GetTemplateVersions 获取模板版本历史
func (s *Service) GetTemplateVersions(templateID string, page, pageSize int) ([]database.PromptTemplateVersion, int64, error) {
	var versions []database.PromptTemplateVersion
	var total int64

	query := s.db.Model(&database.PromptTemplateVersion{}).Where("template_id = ?", templateID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计版本数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("version DESC").Offset(offset).Limit(pageSize).Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("查询版本列表失败: %w", err)
	}

	return versions, total, nil
}

// RestoreTemplateVersion 恢复到指定版本
func (s *Service) RestoreTemplateVersion(templateID string, version int) (*database.PromptTemplate, error) {
	template, err := s.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// 检查是否为系统模板
	if template.IsSystem {
		return nil, errors.New("系统模板不可恢复")
	}

	// 查找指定版本
	var templateVersion database.PromptTemplateVersion
	if err := s.db.Where("template_id = ? AND version = ?", templateID, version).First(&templateVersion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("版本 %d 不存在", version)
		}
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}

	// 更新模板
	updates := map[string]interface{}{
		"content":    templateVersion.Content,
		"variables":  templateVersion.Variables,
		"version":    template.Version + 1,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(template).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("恢复版本失败: %w", err)
	}

	// 创建版本记录
	changeNote := fmt.Sprintf("恢复到版本 %d", version)
	if err := s.createVersionRecord(template, changeNote); err != nil {
		return nil, fmt.Errorf("创建版本记录失败: %w", err)
	}

	return s.GetTemplate(templateID)
}
