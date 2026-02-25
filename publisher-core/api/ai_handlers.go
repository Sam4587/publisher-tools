package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"publisher-core/ai/provider"
)

type AIServiceAPI interface {
	Generate(ctx context.Context, providerName string, opts *provider.GenerateOptions) (*provider.GenerateResult, error)
	GenerateStream(ctx context.Context, providerName string, opts *provider.GenerateOptions) (<-chan string, error)
	ListProviders() []string
	ListModels() map[string][]string
	GenerateContent(ctx context.Context, prompt string, options map[string]interface{}) (interface{}, error)
	OptimizeTitle(ctx context.Context, title string, platform string) (string, error)
	AnalyzeContent(ctx context.Context, content string) (interface{}, error)
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
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	providers := s.ai.ListProviders()
	jsonSuccess(w, map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	})
}

func (s *Server) listAIModels(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	models := s.ai.ListModels()
	jsonSuccess(w, models)
}

func (s *Server) aiGenerate(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Messages    []provider.Message `json:"messages"`
		Model       string             `json:"model,omitempty"`
		MaxTokens   int                `json:"max_tokens,omitempty"`
		Temperature float64            `json:"temperature,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	opts := &provider.GenerateOptions{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	result, err := s.ai.Generate(r.Context(), "", opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) aiGenerateWithProvider(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
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
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	opts := &provider.GenerateOptions{
		Messages:    req.Messages,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	result, err := s.ai.Generate(r.Context(), providerName, opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) aiAnalyzeHotspot(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "You are a hotspot analysis expert skilled in analyzing news hotspots, extracting key information, and determining trend directions."},
		{Role: provider.RoleUser, Content: buildHotspotPrompt(req.Title, req.Content)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 1000,
	}

	result, err := s.ai.Generate(r.Context(), "", opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"analysis": result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentGenerate(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Topic    string `json:"topic"`
		Platform string `json:"platform"`
		Style    string `json:"style"`
		Length   int    `json:"length"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Length == 0 {
		req.Length = 500
	}
	if req.Style == "" {
		req.Style = "casual"
	}
	if req.Platform == "" {
		req.Platform = "general"
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位专业的中文内容创作者，擅长撰写吸引人的文章和社交媒体内容。请始终使用中文进行创作。"},
		{Role: provider.RoleUser, Content: buildContentPrompt(req.Topic, req.Platform, req.Style, req.Length)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate(r.Context(), "", opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"content":  result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentRewrite(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Content  string `json:"content"`
		Style    string `json:"style"`
		Platform string `json:"platform"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Style == "" {
		req.Style = "professional"
	}
	if req.Platform == "" {
		req.Platform = "general"
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "你是一位专业的中文内容创作者，擅长为不同平台和风格改写内容。请始终使用中文进行创作。"},
		{Role: provider.RoleUser, Content: buildRewritePrompt(req.Content, req.Style, req.Platform)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate(r.Context(), "", opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"content":  result.Content,
		"provider": result.Provider,
		"model":    result.Model,
	})
}

func (s *Server) aiContentAudit(w http.ResponseWriter, r *http.Request) {
	if s.ai == nil {
		jsonError(w, "SERVICE_UNAVAILABLE", "AI service not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", "Invalid request format: "+err.Error(), http.StatusBadRequest)
		return
	}

	messages := []provider.Message{
		{Role: provider.RoleSystem, Content: "You are a content review expert skilled in identifying sensitive information, violations, and potential risks in content."},
		{Role: provider.RoleUser, Content: buildAuditPrompt(req.Content)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 500,
	}

	result, err := s.ai.Generate(r.Context(), "", opts)
	if err != nil {
		jsonError(w, "AI_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"audit_result": result.Content,
		"provider":     result.Provider,
		"model":        result.Model,
	})
}

func buildHotspotPrompt(title, content string) string {
	return `Analyze the following hotspot topic:

Title: ` + title + `
Content: ` + content + `

Analyze from the following dimensions:
1. Event summary (within 50 words)
2. Key points (3-5 points)
3. Sentiment (positive/negative/neutral)
4. Relevance score (1-10)
5. Content creation suggestions (2-3 suggestions)

Output in JSON format.`
}

func buildContentPrompt(topic, platform, style string, length int) string {
	return `请根据以下要求生成内容（请用中文输出）：

主题：` + topic + `
平台：` + platform + `
风格：` + style + `
字数：约 ` + strconv.Itoa(length) + ` 字

请生成适合该平台的内容，包括标题和正文。输出语言必须为中文。`
}

func buildRewritePrompt(content, style, platform string) string {
	return `请将以下内容改写为 ` + style + ` 风格，适合 ` + platform + ` 平台（请用中文输出）：

原文：
` + content + `

要求：
1. 保持核心含义不变
2. 改变表达方式和语言风格
3. 符合平台内容规范

请直接输出改写后的内容，语言必须为中文。`
}

func buildAuditPrompt(content string) string {
	return `Review the following content for issues:

` + content + `

Check:
1. Whether it contains sensitive words or violations
2. Whether there are factual errors
3. Whether there are inappropriate expressions
4. Whether it is suitable for public platform publishing

Output the review result in JSON format.`
}
