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
	GenerateContent(prompt string, options map[string]interface{}) (interface{}, error)
	OptimizeTitle(title string, platform string) (string, error)
	AnalyzeContent(content string) (interface{}, error)
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

	result, err := s.ai.Generate("", opts)
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

	result, err := s.ai.Generate(providerName, opts)
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

	result, err := s.ai.Generate("", opts)
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
		{Role: provider.RoleSystem, Content: "You are a professional content creator skilled in writing engaging articles and social media content."},
		{Role: provider.RoleUser, Content: buildContentPrompt(req.Topic, req.Platform, req.Style, req.Length)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate("", opts)
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
		{Role: provider.RoleSystem, Content: "You are a professional content creator skilled in rewriting content for different platforms and styles."},
		{Role: provider.RoleUser, Content: buildRewritePrompt(req.Content, req.Style, req.Platform)},
	}

	opts := &provider.GenerateOptions{
		Messages:  messages,
		MaxTokens: 2000,
	}

	result, err := s.ai.Generate("", opts)
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

	result, err := s.ai.Generate("", opts)
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
	return `Generate content based on the following requirements:

Topic: ` + topic + `
Platform: ` + platform + `
Style: ` + style + `
Word count: Around ` + string(rune(length)) + ` words

Generate content suitable for the platform, including title and body.`
}

func buildRewritePrompt(content, style, platform string) string {
	return `Rewrite the following content in ` + style + ` style, suitable for ` + platform + ` platform:

Original:
` + content + `

Requirements:
1. Keep the core meaning unchanged
2. Change expression and language style
3. Comply with platform content guidelines

Output the rewritten content directly.`
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
