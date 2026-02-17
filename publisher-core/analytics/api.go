package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type APIHandler struct {
	service *AnalyticsService
}

func NewAPIHandler(service *AnalyticsService) *APIHandler {
	return &APIHandler{service: service}
}


func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/analytics/dashboard", h.GetDashboard).Methods("GET")
	router.HandleFunc("/api/analytics/posts", h.ListPostMetrics).Methods("GET")
	router.HandleFunc("/api/analytics/posts/{postId}", h.GetPostMetrics).Methods("GET")
	router.HandleFunc("/api/analytics/trends", h.GetTrends).Methods("GET")
	router.HandleFunc("/api/analytics/refresh", h.RefreshMetrics).Methods("POST")
	router.HandleFunc("/api/analytics/mock", h.GenerateMockData).Methods("POST")
	router.HandleFunc("/api/analytics/report/weekly", h.GetWeeklyReport).Methods("GET")
	router.HandleFunc("/api/analytics/report/monthly", h.GetMonthlyReport).Methods("GET")
	router.HandleFunc("/api/analytics/report/export", h.ExportReport).Methods("GET")
}


type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func (h *APIHandler) jsonSuccess(w http.ResponseWriter, data interface{}) {
	resp := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) jsonError(w http.ResponseWriter, message string, code int) {
	resp := APIResponse{
		Success:   false,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboardStats()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.jsonSuccess(w, stats)
}

func (h *APIHandler) ListPostMetrics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	platform := Platform(query.Get("platform"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	// For now, return from storage
	// TODO: Implement proper storage interface

	h.jsonSuccess(w, map[string]interface{}{
		"platform": platform,
		"limit":    limit,
		"posts":    []interface{}{},
	})
}

func (h *APIHandler) GetPostMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postId"]

	h.jsonSuccess(w, map[string]interface{}{
		"post_id": postID,
		"message": "Post metrics retrieval not implemented",
	})
}

func (h *APIHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	metricType := MetricType(query.Get("metric"))
	if metricType == "" {
		metricType = MetricTypeViews
	}

	platform := Platform(query.Get("platform"))
	days, _ := strconv.Atoi(query.Get("days"))
	if days <= 0 {
		days = 7
	}

	trends, err := h.service.GetTrendData(metricType, platform, days)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]interface{}{
		"metric_type": metricType,
		"platform":    platform,
		"days":        days,
		"trends":      trends,
	})
}

func (h *APIHandler) RefreshMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.service.RefreshMetrics(ctx)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{"message": "Metrics refresh initiated"})
}

func (h *APIHandler) GenerateMockData(w http.ResponseWriter, r *http.Request) {
	// Generate mock data for testing
	if storage, ok := h.service.storage.(*JSONStorage); ok {
		if err := storage.GenerateMockData(); err != nil {
			h.jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.jsonSuccess(w, map[string]string{"message": "Mock data generated"})
		return
	}

	h.jsonError(w, "Storage does not support mock data generation", http.StatusBadRequest)
}

func (h *APIHandler) GetWeeklyReport(w http.ResponseWriter, r *http.Request) {
	generator := NewReportGenerator(h.service.storage)
	report, err := generator.GenerateWeeklyReport()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.jsonSuccess(w, report)
}

func (h *APIHandler) GetMonthlyReport(w http.ResponseWriter, r *http.Request) {
	generator := NewReportGenerator(h.service.storage)
	report, err := generator.GenerateMonthlyReport()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.jsonSuccess(w, report)
}

func (h *APIHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	format := query.Get("format")
	if format == "" {
		format = "json"
	}

	generator := NewReportGenerator(h.service.storage)
	report, err := generator.GenerateWeeklyReport()
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if format == "markdown" || format == "md" {
		md := generator.ExportMarkdown(report)
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", "attachment; filename=report.md")
		w.Write([]byte(md))
		return
	}

	// 默认JSON格式
	json, err := generator.ExportJSON(report)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=report.json")
	w.Write([]byte(json))
}
