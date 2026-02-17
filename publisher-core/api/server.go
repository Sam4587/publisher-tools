// Package api 提供 REST API 服务
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Server API服务�?
type Server struct {
	router      *mux.Router
	taskManager TaskManagerAPI
	publisher   PublisherAPI
	storage     StorageAPI
	ai          AIServiceAPI
	middleware  []Middleware
	server      *http.Server
}

// TaskManagerAPI 任务管理器接�?
type TaskManagerAPI interface {
	CreateTask(taskType string, platform string, payload map[string]interface{}) (interface{}, error)
	GetTask(taskID string) (interface{}, error)
	ListTasks(status string, platform string, limit int) (interface{}, error)
	CancelTask(taskID string) error
}

// PublisherAPI 发布器接�?
type PublisherAPI interface {
	GetPlatforms() []string
	GetPlatformInfo(platform string) (interface{}, error)
	Login(platform string) (interface{}, error)
	CheckLogin(platform string) (interface{}, error)
	Logout(platform string) (interface{}, error)
}

// StorageAPI 存储接口
type StorageAPI interface {
	Upload(file []byte, path string) (string, error)
	Download(path string) ([]byte, error)
	List(prefix string) ([]string, error)
	Delete(path string) error
}

// Middleware 中间�?
type Middleware func(http.Handler) http.Handler

// APIResponse 统一API响应
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewServer 创建API服务�?
func NewServer() *Server {
	s := &Server{
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

// WithTaskManager 设置任务管理�?
func (s *Server) WithTaskManager(tm TaskManagerAPI) *Server {
	s.taskManager = tm
	return s
}

// WithPublisher 设置发布�?
func (s *Server) WithPublisher(p PublisherAPI) *Server {
	s.publisher = p
	return s
}

// WithStorage 设置存储
func (s *Server) WithStorage(st StorageAPI) *Server {
	s.storage = st
	return s
}

// WithMiddleware 添加中间�?
func (s *Server) WithMiddleware(m Middleware) *Server {
	s.middleware = append(s.middleware, m)
	return s
}

func (s *Server) setupRoutes() {
	// 健康检�?
	s.router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// 平台相关
	s.router.HandleFunc("/api/v1/platforms", s.listPlatforms).Methods("GET")
	s.router.HandleFunc("/api/v1/platforms/{platform}", s.getPlatformInfo).Methods("GET")
	s.router.HandleFunc("/api/v1/platforms/{platform}/login", s.login).Methods("POST")
	s.router.HandleFunc("/api/v1/platforms/{platform}/check", s.checkLogin).Methods("GET")
	s.router.HandleFunc("/api/v1/platforms/{platform}/logout", s.logout).Methods("POST")

	// 任务相关
	s.router.HandleFunc("/api/v1/tasks", s.createTask).Methods("POST")
	s.router.HandleFunc("/api/v1/tasks", s.listTasks).Methods("GET")
	s.router.HandleFunc("/api/v1/tasks/{taskId}", s.getTask).Methods("GET")
	s.router.HandleFunc("/api/v1/tasks/{taskId}/cancel", s.cancelTask).Methods("POST")

	// 发布相关
	s.router.HandleFunc("/api/v1/publish", s.publish).Methods("POST")
	s.router.HandleFunc("/api/v1/publish/async", s.publishAsync).Methods("POST")

	// 存储相关
	s.router.HandleFunc("/api/v1/storage/upload", s.uploadFile).Methods("POST")
	s.router.HandleFunc("/api/v1/storage/download", s.downloadFile).Methods("GET")
	s.router.HandleFunc("/api/v1/storage/list", s.listFiles).Methods("GET")
	s.router.HandleFunc("/api/v1/storage/delete", s.deleteFile).Methods("DELETE")

	// AI 相关
	s.setupAIRoutes()
}

// Router 返回路由�?
func (s *Server) Router() *mux.Router {
	return s.router
}

// Start 启动服务�?
func (s *Server) Start(addr string) error {
	var handler http.Handler = s.router

	// 应用中间�?
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logrus.Infof("API服务器启�? %s", addr)
	return s.server.ListenAndServe()
}

// Shutdown 关闭服务�?
func (s *Server) Shutdown() error {
	if s.server != nil {
		return s.server.Shutdown(nil)
	}
	return nil
}

// 响应辅助方法

func (s *Server) jsonSuccess(w http.ResponseWriter, data interface{}) {
	resp := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	s.jsonResponse(w, http.StatusOK, resp)
}

func (s *Server) jsonError(w http.ResponseWriter, code string, message string, statusCode int) {
	resp := APIResponse{
		Success:   false,
		Error:     message,
		ErrorCode: code,
		Timestamp: time.Now().Unix(),
	}
	s.jsonResponse(w, statusCode, resp)
}

func (s *Server) jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// API 处理方法

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.jsonSuccess(w, map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

func (s *Server) listPlatforms(w http.ResponseWriter, r *http.Request) {
	if s.publisher == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "发布服务未初始化", http.StatusServiceUnavailable)
		return
	}

	platforms := s.publisher.GetPlatforms()
	s.jsonSuccess(w, map[string]interface{}{
		"platforms": platforms,
		"count":     len(platforms),
	})
}

func (s *Server) getPlatformInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	if s.publisher == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "发布服务未初始化", http.StatusServiceUnavailable)
		return
	}

	info, err := s.publisher.GetPlatformInfo(platform)
	if err != nil {
		s.jsonError(w, "PLATFORM_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	s.jsonSuccess(w, info)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	if s.publisher == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "发布服务未初始化", http.StatusServiceUnavailable)
		return
	}

	result, err := s.publisher.Login(platform)
	if err != nil {
		s.jsonError(w, "LOGIN_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, result)
}

func (s *Server) checkLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	if s.publisher == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "发布服务未初始化", http.StatusServiceUnavailable)
		return
	}

	result, err := s.publisher.CheckLogin(platform)
	if err != nil {
		s.jsonError(w, "CHECK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, result)
}


func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	platform := vars["platform"]

	if s.publisher == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "发布服务未初始化", http.StatusServiceUnavailable)
		return
	}

	result, err := s.publisher.Logout(platform)
	if err != nil {
		s.jsonError(w, "LOGOUT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, result)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type     string                 `json:"type"`
		Platform string                 `json:"platform"`
		Payload  map[string]interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "请求体格式错�?, http.StatusBadRequest)
		return
	}

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	task, err := s.taskManager.CreateTask(req.Type, req.Platform, req.Payload)
	if err != nil {
		s.jsonError(w, "CREATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, task)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	status := query.Get("status")
	platform := query.Get("platform")
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	tasks, err := s.taskManager.ListTasks(status, platform, limit)
	if err != nil {
		s.jsonError(w, "LIST_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, tasks)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	task, err := s.taskManager.GetTask(taskID)
	if err != nil {
		s.jsonError(w, "TASK_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	s.jsonSuccess(w, task)
}

func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["taskId"]

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	if err := s.taskManager.CancelTask(taskID); err != nil {
		s.jsonError(w, "CANCEL_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]string{"message": "任务已取�?})
}

func (s *Server) publish(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "请求体格式错�?, http.StatusBadRequest)
		return
	}

	if req.Platform == "" {
		s.jsonError(w, "INVALID_PLATFORM", "平台不能为空", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		s.jsonError(w, "INVALID_TITLE", "标题不能为空", http.StatusBadRequest)
		return
	}

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	payload := map[string]interface{}{
		"platform": req.Platform,
		"type":     req.Type,
		"title":    req.Title,
		"content":  req.Content,
		"images":   req.Images,
		"video":    req.Video,
		"tags":     req.Tags,
	}

	newTask, err := s.taskManager.CreateTask("publish", req.Platform, payload)
	if err != nil {
		s.jsonError(w, "CREATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	taskID := ""
	if t, ok := newTask.(map[string]interface{}); ok {
		if id, ok := t["id"].(string); ok {
			taskID = id
		}
	} else if t, ok := newTask.(*struct{ ID string }); ok {
		taskID = t.ID
	}

	s.jsonSuccess(w, map[string]interface{}{
		"task_id":  taskID,
		"status":   "created",
		"message":  "发布任务已创�?,
		"platform": req.Platform,
		"title":    req.Title,
	})
}

func (s *Server) publishAsync(w http.ResponseWriter, r *http.Request) {
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "INVALID_REQUEST", "请求体格式错�?, http.StatusBadRequest)
		return
	}

	if req.Platform == "" {
		s.jsonError(w, "INVALID_PLATFORM", "平台不能为空", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		s.jsonError(w, "INVALID_TITLE", "标题不能为空", http.StatusBadRequest)
		return
	}

	if s.taskManager == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "任务管理服务未初始化", http.StatusServiceUnavailable)
		return
	}

	payload := map[string]interface{}{
		"platform": req.Platform,
		"type":     req.Type,
		"title":    req.Title,
		"content":  req.Content,
		"images":   req.Images,
		"video":    req.Video,
		"tags":     req.Tags,
	}

	newTask, err := s.taskManager.CreateTask("publish", req.Platform, payload)
	if err != nil {
		s.jsonError(w, "CREATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	taskID := ""
	if t, ok := newTask.(map[string]interface{}); ok {
		if id, ok := t["id"].(string); ok {
			taskID = id
		}
	} else if t, ok := newTask.(*struct{ ID string }); ok {
		taskID = t.ID
	}

	s.jsonSuccess(w, map[string]interface{}{
		"task_id":  taskID,
		"status":   "pending",
		"message":  "异步发布任务已创�?,
		"platform": req.Platform,
	})
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "存储服务未初始化", http.StatusServiceUnavailable)
		return
	}

	maxSize := int64(100 * 1024 * 1024) // 100MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		s.jsonError(w, "FILE_TOO_LARGE", "文件大小超过限制(100MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, "INVALID_FILE", "无法读取上传文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		s.jsonError(w, "READ_FAILED", "读取文件内容失败", http.StatusInternalServerError)
		return
	}

	// 生成存储路径
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	storagePath := filepath.Join("uploads", time.Now().Format("2006/01/02"), filename)

	url, err := s.storage.Upload(data, storagePath)
	if err != nil {
		s.jsonError(w, "UPLOAD_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"filename":     header.Filename,
		"size":         header.Size,
		"storage_path": storagePath,
		"url":          url,
	})
}

func (s *Server) downloadFile(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "存储服务未初始化", http.StatusServiceUnavailable)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.jsonError(w, "INVALID_PATH", "文件路径不能为空", http.StatusBadRequest)
		return
	}

	data, err := s.storage.Download(path)
	if err != nil {
		s.jsonError(w, "DOWNLOAD_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应�?
	filename := filepath.Base(path)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func (s *Server) listFiles(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "存储服务未初始化", http.StatusServiceUnavailable)
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = ""
	}

	files, err := s.storage.List(prefix)
	if err != nil {
		s.jsonError(w, "LIST_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

func (s *Server) deleteFile(w http.ResponseWriter, r *http.Request) {
	if s.storage == nil {
		s.jsonError(w, "SERVICE_UNAVAILABLE", "存储服务未初始化", http.StatusServiceUnavailable)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.jsonError(w, "INVALID_PATH", "文件路径不能为空", http.StatusBadRequest)
		return
	}

	if err := s.storage.Delete(path); err != nil {
		s.jsonError(w, "DELETE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.jsonSuccess(w, map[string]string{
		"message": "文件已删�?,
		"path":    path,
	})
}

// PublishRequest 发布请求
type PublishRequest struct {
	Platform   string   `json:"platform"`    // 平台
	Type       string   `json:"type"`        // 类型: images/video
	Title      string   `json:"title"`       // 标题
	Content    string   `json:"content"`     // 正文
	Images     []string `json:"images"`      // 图片路径
	Video      string   `json:"video"`       // 视频路径
	Tags       []string `json:"tags"`        // 标签
	ScheduleAt *string  `json:"schedule_at"` // 定时发布时间
}

// LoggingMiddleware 日志中间�?
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logrus.Infof("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		logrus.Infof("[%s] %s 完成 (%v)", r.Method, r.URL.Path, time.Since(start))
	})
}

// CORSMiddleware CORS中间�?
func CORSMiddleware(next http.Handler) http.Handler {
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
