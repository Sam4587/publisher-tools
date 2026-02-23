package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"publisher-core/prompt"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// PromptTemplateHandler 提示词模板处理器
type PromptTemplateHandler struct {
	service       *prompt.Service
	abTestService *prompt.ABTestService
}

// NewPromptTemplateHandler 创建提示词模板处理器
func NewPromptTemplateHandler(db *gorm.DB) *PromptTemplateHandler {
	return &PromptTemplateHandler{
		service:       prompt.NewService(db),
		abTestService: prompt.NewABTestService(db),
	}
}

// RegisterRoutes 注册路由
func (h *PromptTemplateHandler) RegisterRoutes(router *mux.Router) {
	// 模板管理路由
	router.HandleFunc("/api/v1/prompt-templates", h.CreateTemplate).Methods("POST")
	router.HandleFunc("/api/v1/prompt-templates", h.ListTemplates).Methods("GET")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}", h.GetTemplate).Methods("GET")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}", h.UpdateTemplate).Methods("PUT")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}", h.DeleteTemplate).Methods("DELETE")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}/render", h.RenderTemplate).Methods("POST")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}/versions", h.GetTemplateVersions).Methods("GET")
	router.HandleFunc("/api/v1/prompt-templates/{templateId}/versions/{version}/restore", h.RestoreTemplateVersion).Methods("POST")

	// A/B测试路由
	router.HandleFunc("/api/v1/prompt-ab-tests", h.CreateABTest).Methods("POST")
	router.HandleFunc("/api/v1/prompt-ab-tests", h.ListABTests).Methods("GET")
	router.HandleFunc("/api/v1/prompt-ab-tests/{testId}", h.GetABTest).Methods("GET")
	router.HandleFunc("/api/v1/prompt-ab-tests/{testId}", h.UpdateABTest).Methods("PUT")
	router.HandleFunc("/api/v1/prompt-ab-tests/{testId}/complete", h.CompleteABTest).Methods("POST")
	router.HandleFunc("/api/v1/prompt-ab-tests/{testId}/stats", h.GetABTestStats).Methods("GET")
}

// CreateTemplate 创建模板
func (h *PromptTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req prompt.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	template, err := h.service.CreateTemplate(&req)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusCreated, template)
}

// GetTemplate 获取模板
func (h *PromptTemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]

	template, err := h.service.GetTemplate(templateID)
	if err != nil {
		promptRespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, template)
}

// ListTemplates 列出模板
func (h *PromptTemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	templateType := query.Get("type")
	category := query.Get("category")
	isActiveStr := query.Get("is_active")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	var isActive *bool
	if isActiveStr != "" {
		val := isActiveStr == "true"
		isActive = &val
	}

	page, _ := strconv.Atoi(pageStr)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize == 0 {
		pageSize = 20
	}

	templates, total, err := h.service.ListTemplates(templateType, category, isActive, page, pageSize)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateTemplate 更新模板
func (h *PromptTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]

	var req prompt.UpdateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	template, err := h.service.UpdateTemplate(templateID, &req)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, template)
}

// DeleteTemplate 删除模板
func (h *PromptTemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]

	if err := h.service.DeleteTemplate(templateID); err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// RenderTemplate 渲染模板
func (h *PromptTemplateHandler) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]

	var req struct {
		Variables map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	rendered, err := h.service.RenderTemplate(templateID, req.Variables)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]string{"content": rendered})
}

// GetTemplateVersions 获取模板版本历史
func (h *PromptTemplateHandler) GetTemplateVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]

	query := r.URL.Query()
	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	page, _ := strconv.Atoi(pageStr)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize == 0 {
		pageSize = 20
	}

	versions, total, err := h.service.GetTemplateVersions(templateID, page, pageSize)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"versions":  versions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RestoreTemplateVersion 恢复模板版本
func (h *PromptTemplateHandler) RestoreTemplateVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["templateId"]
	versionStr := vars["version"]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的版本号")
		return
	}

	template, err := h.service.RestoreTemplateVersion(templateID, version)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, template)
}

// CreateABTest 创建A/B测试
func (h *PromptTemplateHandler) CreateABTest(w http.ResponseWriter, r *http.Request) {
	var req prompt.CreateABTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	abTest, err := h.abTestService.CreateABTest(&req)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusCreated, abTest)
}

// GetABTest 获取A/B测试
func (h *PromptTemplateHandler) GetABTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testIDStr := vars["testId"]

	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的测试ID")
		return
	}

	abTest, err := h.abTestService.GetABTest(uint(testID))
	if err != nil {
		promptRespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, abTest)
}

// ListABTests 列出A/B测试
func (h *PromptTemplateHandler) ListABTests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	status := query.Get("status")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("page_size")

	page, _ := strconv.Atoi(pageStr)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize == 0 {
		pageSize = 20
	}

	abTests, total, err := h.abTestService.ListABTests(status, page, pageSize)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"ab_tests":  abTests,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateABTest 更新A/B测试
func (h *PromptTemplateHandler) UpdateABTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testIDStr := vars["testId"]

	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的测试ID")
		return
	}

	var req prompt.UpdateABTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	abTest, err := h.abTestService.UpdateABTest(uint(testID), &req)
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, abTest)
}

// CompleteABTest 完成A/B测试
func (h *PromptTemplateHandler) CompleteABTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testIDStr := vars["testId"]

	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的测试ID")
		return
	}

	var req struct {
		WinnerTemplateID string `json:"winner_template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := h.abTestService.CompleteABTest(uint(testID), req.WinnerTemplateID); err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, map[string]string{"message": "测试已完成"})
}

// GetABTestStats 获取A/B测试统计
func (h *PromptTemplateHandler) GetABTestStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testIDStr := vars["testId"]

	testID, err := strconv.ParseUint(testIDStr, 10, 32)
	if err != nil {
		promptRespondWithError(w, http.StatusBadRequest, "无效的测试ID")
		return
	}

	stats, err := h.abTestService.GetABTestStats(uint(testID))
	if err != nil {
		promptRespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	promptRespondWithJSON(w, http.StatusOK, stats)
}

// promptRespondWithError 返回错误响应
func promptRespondWithError(w http.ResponseWriter, code int, message string) {
	promptRespondWithJSON(w, code, map[string]string{"error": message})
}

// promptRespondWithJSON 返回JSON响应
func promptRespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
