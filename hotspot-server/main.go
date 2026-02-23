package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
)

// HotTopic 热点话题
type HotTopic struct {
	ID          string    `json:"_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Heat        int       `json:"heat"`
	Trend       string    `json:"trend"`
	Source      string    `json:"source"`
	Keywords    []string  `json:"keywords"`
	Suitability int       `json:"suitability"`
	PublishedAt time.Time `json:"publishedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// HotSource 热点数据源
type HotSource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// HotspotStorage 热点存储接口
type HotspotStorage interface {
	Save(topic *HotTopic) error
	Get(id string) (*HotTopic, error)
	List(limit int) ([]*HotTopic, error)
	Delete(id string) error
}

// MemoryStorage 内存存储
type MemoryStorage struct {
	mu     sync.RWMutex
	topics map[string]*HotTopic
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		topics: make(map[string]*HotTopic),
	}
}

func (s *MemoryStorage) Save(topic *HotTopic) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.topics[topic.ID] = topic
	return nil
}

func (s *MemoryStorage) Get(id string) (*HotTopic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	topic, ok := s.topics[id]
	if !ok {
		return nil, fmt.Errorf("topic not found")
	}
	return topic, nil
}

func (s *MemoryStorage) List(limit int) ([]*HotTopic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*HotTopic, 0, len(s.topics))
	for _, topic := range s.topics {
		result = append(result, topic)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (s *MemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.topics, id)
	return nil
}

// HotspotService 热点服务
type HotspotService struct {
	storage  HotspotStorage
	sources  []*HotSource
	fetchers map[string]HotspotFetcher
}

// HotspotFetcher 热点抓取器接口
type HotspotFetcher interface {
	Fetch() ([]*HotTopic, error)
	Name() string
}

// MockFetcher 模拟抓取器
type MockFetcher struct {
	name string
}

func NewMockFetcher(name string) *MockFetcher {
	return &MockFetcher{name: name}
}

func (f *MockFetcher) Fetch() ([]*HotTopic, error) {
	// 模拟抓取热点数据
	now := time.Now()
	topics := []*HotTopic{
		{
			ID:          fmt.Sprintf("%s-%d", f.name, now.Unix()),
			Title:       fmt.Sprintf("[%s] AI 技术突破：GPT-5 即将发布", f.name),
			Description: "OpenAI 宣布下一代语言模型即将发布",
			Category:    "科技",
			Heat:        9999,
			Trend:       "hot",
			Source:      f.name,
			Keywords:    []string{"AI", "GPT-5", "OpenAI"},
			Suitability: 85,
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          fmt.Sprintf("%s-%d-2", f.name, now.Unix()),
			Title:       fmt.Sprintf("[%s] 新能源汽车销量创新高", f.name),
			Description: "2024年新能源汽车销量突破千万辆",
			Category:    "财经",
			Heat:        8888,
			Trend:       "up",
			Source:      f.name,
			Keywords:    []string{"新能源", "汽车", "销量"},
			Suitability: 90,
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	return topics, nil
}

func (f *MockFetcher) Name() string {
	return f.name
}

func NewHotspotService(storage HotspotStorage) *HotspotService {
	sources := []*HotSource{
		{ID: "weibo", Name: "微博热搜", Enabled: true},
		{ID: "douyin", Name: "抖音热点", Enabled: true},
		{ID: "toutiao", Name: "今日头条", Enabled: true},
		{ID: "zhihu", Name: "知乎热榜", Enabled: true},
		{ID: "bilibili", Name: "B站热门", Enabled: true},
	}

	fetchers := make(map[string]HotspotFetcher)
	for _, source := range sources {
		fetchers[source.ID] = NewMockFetcher(source.ID)
	}

	return &HotspotService{
		storage:  storage,
		sources:  sources,
		fetchers: fetchers,
	}
}

func (s *HotspotService) FetchAll() (int, int, error) {
	var totalFetched, totalSaved int

	for _, source := range s.sources {
		if !source.Enabled {
			continue
		}

		fetcher, ok := s.fetchers[source.ID]
		if !ok {
			continue
		}

		topics, err := fetcher.Fetch()
		if err != nil {
			log.Printf("Failed to fetch from %s: %v", source.ID, err)
			continue
		}

		totalFetched += len(topics)

		for _, topic := range topics {
			if err := s.storage.Save(topic); err != nil {
				log.Printf("Failed to save topic: %v", err)
				continue
			}
			totalSaved++
		}
	}

	return totalFetched, totalSaved, nil
}

func (s *HotspotService) GetSources() []*HotSource {
	return s.sources
}

func (s *HotspotService) ListTopics(limit int) ([]*HotTopic, error) {
	return s.storage.List(limit)
}

// APIHandler API 处理器
type APIHandler struct {
	service *HotspotService
	cron    *cron.Cron
}

func NewAPIHandler(service *HotspotService) *APIHandler {
	return &APIHandler{
		service: service,
		cron:    cron.New(),
	}
}

func (h *APIHandler) StartScheduler() {
	// 每30分钟自动抓取一次热点
	h.cron.AddFunc("*/30 * * * *", func() {
		log.Println("Starting scheduled hotspot fetch...")
		fetched, saved, err := h.service.FetchAll()
		if err != nil {
			log.Printf("Scheduled fetch failed: %v", err)
		} else {
			log.Printf("Scheduled fetch completed: fetched=%d, saved=%d", fetched, saved)
		}
	})

	h.cron.Start()
	log.Println("Scheduler started: fetching hotspots every 30 minutes")
}

func (h *APIHandler) StopScheduler() {
	ctx := h.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

// HTTP Handlers

func (h *APIHandler) GetHotTopics(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	topics, err := h.service.ListTopics(limit)
	if err != nil {
		jsonError(w, "LIST_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"topics": topics,
		"total":  len(topics),
	})
}

func (h *APIHandler) GetSources(w http.ResponseWriter, r *http.Request) {
	sources := h.service.GetSources()
	jsonSuccess(w, sources)
}

func (h *APIHandler) FetchHotspots(w http.ResponseWriter, r *http.Request) {
	fetched, saved, err := h.service.FetchAll()
	if err != nil {
		jsonError(w, "FETCH_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取抓取到的热点数据
	topics, _ := h.service.ListTopics(50)

	jsonSuccess(w, map[string]interface{}{
		"fetched": fetched,
		"saved":   saved,
		"topics":  topics,
	})
}

// GetNewTrends 获取新增热点
func (h *APIHandler) GetNewTrends(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		fmt.Sscanf(h, "%d", &hours)
	}

	// 返回空列表（实际应该查询最近 hours 小时内的新增热点）
	jsonSuccess(w, []*HotTopic{})
}

// UpdateHotspots 更新热点
func (h *APIHandler) UpdateHotspots(w http.ResponseWriter, r *http.Request) {
	// 触发抓取
	_, saved, err := h.service.FetchAll()
	if err != nil {
		jsonError(w, "UPDATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Hotspots updated",
		"count":   saved,
	})
}

// AIAnalyze AI 分析热点
func (h *APIHandler) AIAnalyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topics  []HotTopic `json:"topics"`
		Options struct {
			Provider string `json:"provider"`
			Focus    string `json:"focus"`
		} `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	// 模拟 AI 分析结果
	result := map[string]interface{}{
		"summary": "根据分析，当前热点话题主要集中在科技创新和社会发展领域。AI技术的快速发展正在改变各个行业的格局，新能源汽车等新兴产业展现出强劲的增长势头。",
		"keyPoints": []string{
			"AI 技术持续突破，GPT-5 等大模型引领新一轮技术革命",
			"新能源汽车市场快速增长，反映出消费者对环保出行的认可",
			"娱乐产业复苏，春节档票房创新高显示消费信心恢复",
		},
		"sentiment":      "positive",
		"recommendations": []string{"关注 AI 技术发展动态", "布局新能源产业链", "把握消费复苏机遇"},
	}

	jsonSuccess(w, result)
}

// AIBriefing 生成热点简报
func (h *APIHandler) AIBriefing(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topics    []HotTopic `json:"topics"`
		MaxLength int        `json:"maxLength"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	// 模拟生成简报
	brief := "【热点简报】\n\n今日热点主要集中在科技和财经领域。AI技术持续突破，OpenAI宣布GPT-5即将发布，引发广泛关注。新能源汽车销量创新高，2024年突破千万辆大关，显示出强劲的市场需求。娱乐产业方面，春节档票房破纪录，反映消费市场复苏态势良好。\n\n建议关注：AI技术发展、新能源产业链、消费复苏主题。"

	jsonSuccess(w, map[string]interface{}{
		"brief": brief,
	})
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"service": "hotspot-server",
		"time":    time.Now().Unix(),
	})
}

// Helper functions

func jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func jsonError(w http.ResponseWriter, code, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   code,
		"message": message,
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 初始化存储和服务
	storage := NewMemoryStorage()
	service := NewHotspotService(storage)
	handler := NewAPIHandler(service)

	// 启动定时任务
	handler.StartScheduler()
	defer handler.StopScheduler()

	// 初始抓取一次
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Performing initial hotspot fetch...")
		fetched, saved, err := service.FetchAll()
		if err != nil {
			log.Printf("Initial fetch failed: %v", err)
		} else {
			log.Printf("Initial fetch completed: fetched=%d, saved=%d", fetched, saved)
		}
	}()

	// 设置路由
	router := mux.NewRouter()
	router.Use(corsMiddleware)

	// API 路由
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	api.HandleFunc("/hot-topics", handler.GetHotTopics).Methods("GET")
	api.HandleFunc("/hot-topics/sources", handler.GetSources).Methods("GET")
	api.HandleFunc("/hot-topics/fetch", handler.FetchHotspots).Methods("POST")
	api.HandleFunc("/hot-topics/update", handler.UpdateHotspots).Methods("POST")
	api.HandleFunc("/hot-topics/trends/new", handler.GetNewTrends).Methods("GET")
	api.HandleFunc("/hot-topics/ai/analyze", handler.AIAnalyze).Methods("POST")
	api.HandleFunc("/hot-topics/ai/briefing", handler.AIBriefing).Methods("POST")
	
	// 兼容前端路径
	api.HandleFunc("/hot-topics/newsnow/sources", handler.GetSources).Methods("GET")
	api.HandleFunc("/hot-topics/newsnow/fetch", handler.FetchHotspots).Methods("POST")

	// 启动服务器
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Hotspot server starting on %s", addr)
	log.Printf("Health check: http://localhost:%s/api/health", port)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

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
