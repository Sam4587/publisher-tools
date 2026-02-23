// Package api 提供流水线管理 API
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"publisher-core/pipeline"
	"publisher-core/websocket"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// PipelineAPI 流水线 API
type PipelineAPI struct {
	orchestrator *pipeline.PipelineOrchestrator
	wsServer    *websocket.Server
}

// NewPipelineAPI 创建流水线 API
func NewPipelineAPI(orchestrator *pipeline.PipelineOrchestrator, wsServer *websocket.Server) *PipelineAPI {
	return &PipelineAPI{
		orchestrator: orchestrator,
		wsServer:    wsServer,
	}
}

// RegisterRoutes 注册路由
func (api *PipelineAPI) RegisterRoutes(router *mux.Router) {
	// 流水线管理
	router.HandleFunc("/api/v1/pipelines", api.handlePipelines).Methods("GET", "POST")
	router.HandleFunc("/api/v1/pipelines/{id}", api.handlePipelineDetail).Methods("GET", "PUT", "DELETE")

	// 流水线模板
	router.HandleFunc("/api/v1/pipeline-templates", api.handlePipelineTemplates).Methods("GET")
	router.HandleFunc("/api/v1/pipeline-templates/{id}", api.handlePipelineTemplateDetail).Methods("GET")
	router.HandleFunc("/api/v1/pipeline-templates/{id}/use", api.handleUseTemplate).Methods("POST")

	// 流水线执行
	router.HandleFunc("/api/v1/pipelines/{id}/execute", api.handleExecutePipeline).Methods("POST")

	// 执行管理
	router.HandleFunc("/api/v1/executions", api.handleExecutions).Methods("GET")
	router.HandleFunc("/api/v1/executions/{id}", api.handleExecutionDetail).Methods("GET")
	router.HandleFunc("/api/v1/executions/{id}/pause", api.handlePauseExecution).Methods("POST")
	router.HandleFunc("/api/v1/executions/{id}/resume", api.handleResumeExecution).Methods("POST")
	router.HandleFunc("/api/v1/executions/{id}/cancel", api.handleCancelExecution).Methods("POST")
	router.HandleFunc("/api/v1/executions/{id}/logs", api.handleExecutionLogs).Methods("GET")
	router.HandleFunc("/api/v1/executions/{id}/progress", api.handleExecutionProgress).Methods("GET")

	// 监控统计
	router.HandleFunc("/api/v1/monitoring/stats", api.handleMonitoringStats).Methods("GET")

	// WebSocket
	router.HandleFunc("/ws/monitor", api.wsServer.HandleWebSocket)
	router.HandleFunc("/ws/execution/{id}", api.wsServer.HandleWebSocket)
}

// handlePipelines 处理流水线列表和创建
func (api *PipelineAPI) handlePipelines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		api.listPipelines(w, r)
	case "POST":
		api.createPipeline(w, r)
	}
}

// listPipelines 列出所有流水线
func (api *PipelineAPI) listPipelines(w http.ResponseWriter, r *http.Request) {
	pipelines, err := api.orchestrator.ListPipelines()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusOK, pipelines)
}

// createPipeline 创建流水线
func (api *PipelineAPI) createPipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                      `json:"name"`
		Description string                      `json:"description"`
		TemplateID  string                      `json:"template_id"`
		Config      pipeline.PipelineConfig     `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, fmt.Errorf("无效的请求体: %w", err))
		return
	}

	var p *pipeline.Pipeline

	// 如果提供了模板ID，使用模板创建
	if req.TemplateID != "" {
		template, err := pipeline.GetTemplate(req.TemplateID)
		if err != nil {
			sendError(w, http.StatusNotFound, err)
			return
		}

		p = template
		p.ID = "" // 重新生成ID
	} else {
		p = &pipeline.Pipeline{}
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Config.ParallelMode != false || req.Config.MaxParallel != 0 {
		p.Config = req.Config
	}

	if err := api.orchestrator.CreatePipeline(p); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusCreated, p)
}

// handlePipelineDetail 处理流水线详情
func (api *PipelineAPI) handlePipelineDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case "GET":
		api.getPipeline(w, r, id)
	case "PUT":
		api.updatePipeline(w, r, id)
	case "DELETE":
		api.deletePipeline(w, r, id)
	}
}

// getPipeline 获取流水线详情
func (api *PipelineAPI) getPipeline(w http.ResponseWriter, r *http.Request, id string) {
	p, err := api.orchestrator.GetPipeline(id)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	sendJSON(w, http.StatusOK, p)
}

// updatePipeline 更新流水线
func (api *PipelineAPI) updatePipeline(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Config      pipeline.PipelineConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	p, err := api.orchestrator.GetPipeline(id)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Config.ParallelMode != false || req.Config.MaxParallel != 0 {
		p.Config = req.Config
	}

	sendJSON(w, http.StatusOK, p)
}

// deletePipeline 删除流水线
func (api *PipelineAPI) deletePipeline(w http.ResponseWriter, r *http.Request, id string) {
	// TODO: 实现删除逻辑
	sendJSON(w, http.StatusOK, map[string]string{"message": "流水线已删除"})
}

// handlePipelineTemplates 处理流水线模板列表
func (api *PipelineAPI) handlePipelineTemplates(w http.ResponseWriter, r *http.Request) {
	templates := pipeline.ListTemplates()
	sendJSON(w, http.StatusOK, templates)
}

// handlePipelineTemplateDetail 处理流水线模板详情
func (api *PipelineAPI) handlePipelineTemplateDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	template, err := pipeline.GetTemplate(id)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	sendJSON(w, http.StatusOK, template)
}

// handleUseTemplate 使用模板创建流水线
func (api *PipelineAPI) handleUseTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	var req struct {
		Name   string                  `json:"name"`
		Config pipeline.PipelineConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	template, err := pipeline.GetTemplate(templateID)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	// 创建新流水线
	p := template
	p.ID = "" // 重新生成ID

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.Config.ParallelMode != false || req.Config.MaxParallel != 0 {
		p.Config = req.Config
	}

	if err := api.orchestrator.CreatePipeline(p); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusCreated, p)
}

// handleExecutePipeline 执行流水线
func (api *PipelineAPI) handleExecutePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipelineID := vars["id"]

	var req struct {
		Input map[string]interface{} `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	execution, err := api.orchestrator.ExecutePipeline(r.Context(), pipelineID, req.Input)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	// 通知 WebSocket
	api.wsServer.SendStatusChange(execution.ID, string(execution.Status))

	sendJSON(w, http.StatusCreated, execution)
}

// handleExecutions 处理执行列表
func (api *PipelineAPI) handleExecutions(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	_ = r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	_ = limit // 使用limit避免编译警告

	// TODO: 从存储中获取执行列表
	// 这里返回模拟数据
	executions := []*pipeline.PipelineExecution{}

	sendJSON(w, http.StatusOK, executions)
}

// handleExecutionDetail 处理执行详情
func (api *PipelineAPI) handleExecutionDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	execution, err := api.orchestrator.GetExecutionStatus(id)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	sendJSON(w, http.StatusOK, execution)
}

// handlePauseExecution 暂停执行
func (api *PipelineAPI) handlePauseExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := api.orchestrator.PausePipeline(id); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "执行已暂停"})
}

// handleResumeExecution 恢复执行
func (api *PipelineAPI) handleResumeExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := api.orchestrator.ResumePipeline(id); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "执行已恢复"})
}

// handleCancelExecution 取消执行
func (api *PipelineAPI) handleCancelExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := api.orchestrator.CancelPipeline(id); err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "执行已取消"})
}

// handleExecutionLogs 获取执行日志
func (api *PipelineAPI) handleExecutionLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	logs, err := api.orchestrator.GetExecutionLogs(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendJSON(w, http.StatusOK, logs)
}

// handleExecutionProgress 获取执行进度
func (api *PipelineAPI) handleExecutionProgress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	execution, err := api.orchestrator.GetExecutionStatus(id)
	if err != nil {
		sendError(w, http.StatusNotFound, err)
		return
	}

	// 计算进度
	completedSteps := 0
	for _, step := range execution.Steps {
		if step.Status == pipeline.StepStatusCompleted {
			completedSteps++
		}
	}

	progress := float64(completedSteps) / float64(len(execution.Steps)) * 100

	response := map[string]interface{}{
		"execution_id": execution.ID,
		"progress":     progress,
		"total_steps":  len(execution.Steps),
		"completed_steps": completedSteps,
		"status":       execution.Status,
		"steps":        execution.Steps,
	}

	sendJSON(w, http.StatusOK, response)
}

// handleMonitoringStats 获取监控统计
func (api *PipelineAPI) handleMonitoringStats(w http.ResponseWriter, r *http.Request) {
	// TODO: 从存储中获取真实统计数据
	stats := map[string]interface{}{
		"running":  3,
		"pending":  12,
		"today":    45,
		"week":     234,
		"success_rate": 96.5,
		"failed":   2,
	}

	sendJSON(w, http.StatusOK, stats)
}

// sendJSON 发送 JSON 响应
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError 发送错误响应
func sendError(w http.ResponseWriter, status int, err error) {
	logrus.Errorf("API Error: %v", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error": err.Error(),
	}

	if status >= 500 {
		response["message"] = "内部服务器错误"
	} else if status >= 400 {
		response["message"] = "请求错误"
	}

	json.NewEncoder(w).Encode(response)
}

// CORS 中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
