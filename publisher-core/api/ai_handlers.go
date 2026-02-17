package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"publisher-core/ai/provider"
)

type AIServiceAPI interface {
	Generate(providerName string, opts *provider.GenerateOptions) (*provider.GenerateResult, error)
	GenerateStream(providerName string, opts *provider.GenerateOptions) (<-chan string, error)
	ListProviders() []string
	ListModels() map[string][]string
}

func (s *Server) WithAI(ai AIServiceAPI) *Server {
	s.ai = ai
	return s
}

func (s *Server) setupAIRoutes() {
	s.router.HandleFunc("/api/v1/ai/providers", s.listAIProviders).Methods("GET")
	s.router.HandleFunc("/api/v1/ai/models", s.listAIModels).Methods("GET")
	s.router.HandleFunc("/api/v1/ai/generate", s.aiGenerate).Methods("POST")
	s.router.HandleFunc("/api/v1/ai/generate/{provider}", s.aiGenerateWithProvider).Methods("POST")
	s.router.HandleFunc("/api/v1/ai/analyze/hotspot", s.aiAnalyzeHotspot).Methods("POST")
	s.router.HandleFunc("/api/v1/ai/content/generate", s.aiContentGenerate).Methods("POST")
	s.router.HandleFunc("/api/v1/ai/content/rewrite", s.aiContentRewrite).Methods("POST")
	s.router.HandleFunc("/api/v1/ai/content/audit", s.aiContentAudit).Methods("POST")
}

func (s *Server) listAIProviders(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	providers := s.ai.ListProviders()
	s.jsonSuccess(w, map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	})
}

func (s *Server) listAIModels(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	models := s.ai.ListModels()
	s.jsonSuccess(w, models)
}

func (s *Server) aiGenerate(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Messages    []provider.Message `json:"messages"`
		Model       string             `json:"model,omitempty"`
		MaxTokens   int                `json:"max_tokens,omitempty"`
		Temperature float64            `json:"temperature,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	opts := &provider.GenerateOptions{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	result, err := s.ai.Generate("", opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, result)
}

func (s *Server) aiGenerateWithProvider(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	providerName := vars["provider"]

	var req struct {
		Messages    []provider.Message `json:"messages"`
		Model       string             `json:"model,omitempty"`
		MaxTokens   int                `json:"max_tokens,omitempty"`
		Temperature float64            `json:"temperature,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	opts := &provider.GenerateOptions{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	result, err := s.ai.Generate(providerName, opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, result)
}

func (s *Server) aiAnalyzeHotspot(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位热点分析专家，擅长分析新闻热点、提取关键信息、判断趋势走向�?},
		{Role: provider.RoleUser, Content: buildHotspotPrompt(req.Title, req.Content)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 1000,
	}

	result, err := s.ai.Generate("", opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"analysis": result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentGenerate(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Topic    string `json:"topic"`
		Platform string `json:"platform"`
		Style    string `json:"style"`
		Length   int    `json:"length"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Length == 0 {
		req.Length = 500
	}
	if req.Style == "" {
		req.Style = "轻松幽默"
	}
	if req.Platform == "" {
		req.Platform = "通用"
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位专业的内容创作者，擅长撰写吸引人的文章和社交媒体内容�?},
		{Role: provider.RoleUser, Content: buildContentPrompt(req.Topic, req.Platform, req.Style, req.Length)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate("", opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"content":  result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentRewrite(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Content  string `json:"content"`
		Style    string `json:"style"`
		Platform string `json:"platform"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Style == "" {
		req.Style = "正式专业"
	}
	if req.Platform == "" {
		req.Platform = "通用"
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位专业的内容创作者，擅长改写内容以适应不同平台和风格�?},
		{Role: provider.RoleUser, Content: buildRewritePrompt(req.Content, req.Style, req.Platform)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate("", opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"content":  result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentAudit(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "AI服务未初始化", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "无效的请求格�? "+err.Error(), http.StatusBadRequest)
		return
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位内容审核专家，擅长识别内容中的敏感信息、违规内容和潜在风险�?},
		{Role: provider.RoleUser, Content: buildAuditPrompt(req.Content)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 500,
	}

	result, err := s.ai.Generate("", opts)
	if err != nil {
		s.jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"audit_result": result.Content,
		"provider":     result.Provider,
		"model":        result.Model,
	})
}

func buildHotspotPrompt(title, content string) string {
	return `请分析以下热点话题：

标题：` + title + `
内容：` + content + `

请从以下维度进行分析�?
1. 事件摘要�?0字以内）
2. 关键要点�?-5个要点）
3. 情感倾向（正�?负面/中性）
4. 相关性评分（1-10分）
5. 内容创作建议�?-3条）

请以JSON格式输出。`
}

func buildContentPrompt(topic, platform, style string, length int) string {
	return `请根据以下要求生成内容：

主题：` + topic + `
平台：` + platform + `
风格：` + style + `
字数要求：` + string(rune(length)) + `字左�?

请生成适合该平台发布的内容，包含标题和正文。`
}

func buildRewritePrompt(content, style, platform string) string {
	return `请将以下内容改写为` + style + `风格，适合` + platform + `平台发布�?

原文�?
` + content + `

要求�?
1. 保持原文核心意思不�?
2. 改变表达方式和语言风格
3. 符合平台内容规范

请直接输出改写后的内容。`
}

func buildAuditPrompt(content string) string {
	return `请审核以下内容是否存在问题：

` + content + `

请检查：
1. 是否包含敏感词汇或违规内�?
2. 是否存在事实错误
3. 是否有不当表�?
4. 是否适合公开平台发布

请以JSON格式输出审核结果。`
}
