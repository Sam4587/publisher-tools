package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"publisher-core/config"
)

type Server struct {
	router      *mux.Router
	taskManager TaskManagerAPI
	publisher   PublisherAPI
	storage     StorageAPI
	ai          AIServiceAPI
	middleware  []Middleware
	server      *http.Server
	security    *config.SecurityConfig
}

type TaskManagerAPI interface {
	CreateTask(taskType string, platform string, payload map[string]interface{}) (interface{}, error)
	GetTask(taskID string) (interface{}, error)
	ListTasks(status string, platform string, limit int) (interface{}, error)
	CancelTask(taskID string) error
}

type PublisherAPI interface {
	GetPlatforms() []string
	GetPlatformInfo(platform string) (interface{}, error)
	Login(platform string) (interface{}, error)
	CheckLogin(platform string) (interface{}, error)
	Logout(platform string) (interface{}, error)
}

type StorageAPI interface {
	Upload(file []byte, path string) (string, error)
	Download(path string) ([]byte, error)
	List(prefix string) ([]string, error)
	Delete(path string) error
}

type Middleware func(http.Handler) http.Handler

// RouteRegistrar 接口用于注册自定义路由
type RouteRegistrar interface {
	RegisterRoutes(router *mux.Router)
}

func NewServer(taskManager TaskManagerAPI, publisher PublisherAPI, storage StorageAPI, ai AIServiceAPI) *Server {
	securityConfig := config.LoadSecurityConfig()
	s := &Server{
		router:      mux.NewRouter(),
		taskManager: taskManager,
		publisher:   publisher,
		storage:     storage,
		ai:          ai,
		security:    securityConfig,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(LoggingMiddleware)
	s.router.Use(CORSMiddleware(
		s.security.AllowedOrigins,
		s.security.AllowedMethods,
		s.security.AllowedHeaders,
		s.security.AllowCredentials,
		s.security.MaxAge,
	))

	apiRouter := s.router.PathPrefix("/api/v1").Subrouter()

	healthRouter := apiRouter.PathPrefix("/health").Subrouter()
	healthRouter.HandleFunc("", s.healthCheck).Methods("GET")
	healthRouter.HandleFunc("/detailed", s.detailedHealthCheck).Methods("GET")

	tasksRouter := apiRouter.PathPrefix("/tasks").Subrouter()
	tasksRouter.HandleFunc("", s.createTask).Methods("POST")
	tasksRouter.HandleFunc("", s.listTasks).Methods("GET")
	tasksRouter.HandleFunc("/{id}", s.getTask).Methods("GET")
	tasksRouter.HandleFunc("/{id}/cancel", s.cancelTask).Methods("POST")

	publisherRouter := apiRouter.PathPrefix("/publisher").Subrouter()
	publisherRouter.HandleFunc("/platforms", s.getPlatforms).Methods("GET")
	publisherRouter.HandleFunc("/platforms/{platform}", s.getPlatformInfo).Methods("GET")
	publisherRouter.HandleFunc("/platforms/{platform}/login", s.loginPlatform).Methods("POST")
	publisherRouter.HandleFunc("/platforms/{platform}/check", s.checkLogin).Methods("GET")
	publisherRouter.HandleFunc("/platforms/{platform}/logout", s.logoutPlatform).Methods("POST")

	aiRouter := apiRouter.PathPrefix("/ai").Subrouter()
	aiRouter.HandleFunc("/generate", s.generateContent).Methods("POST")
	aiRouter.HandleFunc("/optimize-title", s.optimizeTitle).Methods("POST")
	aiRouter.HandleFunc("/analyze", s.analyzeContent).Methods("POST")

	cacheRouter := apiRouter.PathPrefix("/cache").Subrouter()
	cacheRouter.HandleFunc("/stats", s.getCacheStats).Methods("GET")
	cacheRouter.HandleFunc("/clear", s.clearCache).Methods("POST")

	fs := http.FileServer(http.Dir("static"))
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
}

// RegisterRoutes 注册额外的路由
func (s *Server) RegisterRoutes(registrar RouteRegistrar) {
	registrar.RegisterRoutes(s.router)
}

// Router 返回路由器实例
func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logrus.Infof("Server starting on %s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	logrus.Info("Server shutting down...")
	return s.server.Shutdown(ctx)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

func (s *Server) detailedHealthCheck(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
		"uptime": "0s",
		"services": map[string]interface{}{
			"task_manager": "ok",
			"publisher":    "ok",
			"storage":      "ok",
			"ai":           "ok",
		},
	})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskType string                 `json:"task_type"`
		Platform string                 `json:"platform"`
		Payload  map[string]interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	task, err := s.taskManager.CreateTask(req.TaskType, req.Platform, req.Payload)
	if err != nil {
		jsonError(w, "CREATE_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, task)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		jsonError(w, "TASK_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, task)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	platform := r.URL.Query().Get("platform")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	tasks, err := s.taskManager.ListTasks(status, platform, limit)
	if err != nil {
		jsonError(w, "LIST_TASKS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, tasks)
}

func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]

	if err := s.taskManager.CancelTask(taskID); err != nil {
		jsonError(w, "CANCEL_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Task cancelled successfully",
		"task_id": taskID,
	})
}

func (s *Server) getPlatforms(w http.ResponseWriter, r *http.Request) {
	platforms := s.publisher.GetPlatforms()
	jsonSuccess(w, platforms)
}

func (s *Server) getPlatformInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	info, err := s.publisher.GetPlatformInfo(platform)
	if err != nil {
		jsonError(w, "PLATFORM_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, info)
}

func (s *Server) loginPlatform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	result, err := s.publisher.Login(platform)
	if err != nil {
		jsonError(w, "LOGIN_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) checkLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	result, err := s.publisher.CheckLogin(platform)
	if err != nil {
		jsonError(w, "CHECK_LOGIN_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) logoutPlatform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	result, err := s.publisher.Logout(platform)
	if err != nil {
		jsonError(w, "LOGOUT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) generateContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt  string                 `json:"prompt"`
		Options map[string]interface{} `json:"options"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.ai.GenerateContent(req.Prompt, req.Options)
	if err != nil {
		jsonError(w, "GENERATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) optimizeTitle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		Platform string `json:"platform"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.ai.OptimizeTitle(req.Title, req.Platform)
	if err != nil {
		jsonError(w, "OPTIMIZE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]string{
		"optimized_title": result,
	})
}

func (s *Server) analyzeContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.ai.AnalyzeContent(req.Content)
	if err != nil {
		jsonError(w, "ANALYZE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, result)
}

func (s *Server) getCacheStats(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, map[string]interface{}{
		"hits":   0,
		"misses": 0,
		"size":   0,
	})
}

func (s *Server) clearCache(w http.ResponseWriter, r *http.Request) {
	jsonSuccess(w, map[string]interface{}{
		"message": "Cache cleared",
	})
}

func jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logrus.Infof("Request started: %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		logrus.Infof("Request completed: %s %s (duration: %s)",
			r.Method, r.URL.Path, time.Since(start))
	})
}

func CORSMiddleware(allowedOrigins []string, allowedMethods []string, allowedHeaders []string, allowCredentials bool, maxAge int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查来源是否在允许列表中
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					break
				}
			}

			// 如果来源不被允许,不设置CORS头
			if !allowed && origin != "" {
				next.ServeHTTP(w, r)
				return
			}

			// 设置其他CORS头
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))
			}

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
