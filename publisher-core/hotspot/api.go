package hotspot

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type APIHandler struct {
	service *HotspotService
}

func NewAPIHandler(service *HotspotService) *APIHandler {
	return &APIHandler{service: service}
}

func (h *APIHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/hot-topics", h.ListTopics).Methods("GET")
	router.HandleFunc("/api/hot-topics/{id}", h.GetTopic).Methods("GET")
	router.HandleFunc("/api/hot-topics/newsnow/sources", h.GetNewsNowSources).Methods("GET")
	router.HandleFunc("/api/hot-topics/newsnow/fetch", h.FetchFromAllSources).Methods("POST")
	router.HandleFunc("/api/hot-topics/newsnow/fetch/{sourceId}", h.FetchFromSource).Methods("GET")
	router.HandleFunc("/api/hot-topics/update", h.RefreshTopics).Methods("POST")
	router.HandleFunc("/api/hot-topics/trends/new", h.GetNewTopics).Methods("GET")
	router.HandleFunc("/api/hot-topics/{id}", h.DeleteTopic).Methods("DELETE")
}

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

type ListResponse struct {
	Topics     []Topic    `json:"data"`
	Pagination Pagination `json:"pagination"`
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

func (h *APIHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := Filter{
		Category: Category(query.Get("category")),
		Source:   query.Get("source"),
		Limit:    limit,
		Offset:   (page - 1) * limit,
		SortBy:   query.Get("sortBy"),
		SortDesc: query.Get("sortOrder") != "asc",
	}

	if minHeat := query.Get("minHeat"); minHeat != "" {
		filter.MinHeat, _ = strconv.Atoi(minHeat)
	}
	if maxHeat := query.Get("maxHeat"); maxHeat != "" {
		filter.MaxHeat, _ = strconv.Atoi(maxHeat)
	}

	topics, total, err := h.service.List(filter)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pages := total / limit
	if total%limit > 0 {
		pages++
	}

	h.jsonSuccess(w, ListResponse{
		Topics: topics,
		Pagination: Pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			Pages:   pages,
			HasMore: page < pages,
		},
	})
}

func (h *APIHandler) GetTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	topic, err := h.service.Get(id)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if topic == nil {
		h.jsonError(w, "topic not found", http.StatusNotFound)
		return
	}

	h.jsonSuccess(w, topic)
}

func (h *APIHandler) GetNewsNowSources(w http.ResponseWriter, r *http.Request) {
	sources := h.service.GetSources()
	h.jsonSuccess(w, sources)
}

func (h *APIHandler) FetchFromAllSources(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Sources  []string `json:"sources"`
		MaxItems int      `json:"maxItems"`
	}

	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	if req.MaxItems == 0 {
		req.MaxItems = 20
	}

	ctx := r.Context()
	results, err := h.service.FetchFromAllSources(ctx, req.MaxItems)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var allTopics []Topic
	fetched := 0
	for _, topics := range results {
		allTopics = append(allTopics, topics...)
		fetched += len(topics)
	}

	h.jsonSuccess(w, map[string]interface{}{
		"fetched": fetched,
		"saved":   len(allTopics),
		"topics":  allTopics,
	})
}

func (h *APIHandler) FetchFromSource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sourceId := vars["sourceId"]

	maxItems, _ := strconv.Atoi(r.URL.Query().Get("maxItems"))
	if maxItems == 0 {
		maxItems = 20
	}

	ctx := r.Context()
	topics, err := h.service.FetchFromSource(ctx, sourceId, maxItems)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]interface{}{
		"source":     sourceId,
		"sourceName": sourceId,
		"count":      len(topics),
		"topics":     topics,
	})
}

func (h *APIHandler) RefreshTopics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	count, err := h.service.Refresh(ctx)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]interface{}{
		"message": "refresh completed",
		"count":   count,
	})
}

func (h *APIHandler) GetNewTopics(w http.ResponseWriter, r *http.Request) {
	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours <= 0 {
		hours = 24
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	ctx := r.Context()
	topics, err := h.service.GetNewTopics(ctx, since)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, topics)
}

func (h *APIHandler) DeleteTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.Delete(id); err != nil {
		h.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{"message": "deleted"})
}
