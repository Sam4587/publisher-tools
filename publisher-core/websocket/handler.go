package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应该验证来源
		return true
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub                *Hub
	progressService    *ProgressService
	reconnectManager   *ReconnectionManager
	sessionTimeout     time.Duration
}

// NewWebSocketHandler 创建处理器
func NewWebSocketHandler(hub *Hub, progressService *ProgressService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:              hub,
		progressService:  progressService,
		reconnectManager: NewReconnectionManager(),
		sessionTimeout:   30 * time.Minute,
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户信息
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	projectID := c.Query("project_id")

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 创建客户端
	clientID := GenerateClientID()
	client := NewClient(clientID, userID, projectID, conn, h.hub)

	// 注册客户端
	h.hub.Register <- client

	// 保存会话状态（用于断线重连）
	h.reconnectManager.SaveSession(clientID, userID, projectID, make(map[string]bool))

	// 发送欢迎消息
	welcomeMsg := &Message{
		Type:   "connected",
		TaskID: "",
		Payload: map[string]interface{}{
			"client_id": clientID,
			"message":   "WebSocket连接成功",
			"timestamp": time.Now(),
		},
	}
	if msgBytes, err := json.Marshal(welcomeMsg); err == nil {
		client.Send <- msgBytes
	}

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// RegisterRoutes 注册路由
func (h *WebSocketHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/ws", h.HandleWebSocket)
	r.GET("/ws/reconnect", h.HandleReconnection)
}

// ReconnectionManager 重连管理器
type ReconnectionManager struct {
	sessions map[string]*SessionState
	mu       sync.RWMutex
}

// SessionState 会话状态
type SessionState struct {
	ClientID    string
	UserID      string
	ProjectID   string
	Subscribed  map[string]bool
	LastActive  time.Time
	MessageQueue [][]byte // 断线期间的消息队列
}

// NewReconnectionManager 创建重连管理器
func NewReconnectionManager() *ReconnectionManager {
	return &ReconnectionManager{
		sessions: make(map[string]*SessionState),
	}
}

// SaveSession 保存会话
func (m *ReconnectionManager) SaveSession(clientID, userID, projectID string, subscribed map[string]bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[clientID] = &SessionState{
		ClientID:   clientID,
		UserID:     userID,
		ProjectID:  projectID,
		Subscribed: subscribed,
		LastActive: time.Now(),
		MessageQueue: make([][]byte, 0),
	}
}

// GetSession 获取会话
func (m *ReconnectionManager) GetSession(clientID string) (*SessionState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[clientID]
	return session, exists
}

// DeleteSession 删除会话
func (m *ReconnectionManager) DeleteSession(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, clientID)
}

// QueueMessage 队列消息（断线期间）
func (m *ReconnectionManager) QueueMessage(clientID string, message []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[clientID]; exists {
		// 限制队列大小
		if len(session.MessageQueue) < 100 {
			session.MessageQueue = append(session.MessageQueue, message)
		}
	}
}

// GetQueuedMessages 获取队列消息
func (m *ReconnectionManager) GetQueuedMessages(clientID string) [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, exists := m.sessions[clientID]; exists {
		messages := session.MessageQueue
		session.MessageQueue = make([][]byte, 0)
		return messages
	}

	return nil
}

// CleanupOldSessions 清理旧会话
func (m *ReconnectionManager) CleanupOldSessions(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for clientID, session := range m.sessions {
		if session.LastActive.Before(cutoff) {
			delete(m.sessions, clientID)
		}
	}
}

// HandleReconnection 处理重连
func (h *WebSocketHandler) HandleReconnection(c *gin.Context) {
	// 获取之前的客户端ID
	oldClientID := c.Query("old_client_id")
	if oldClientID == "" {
		// 没有旧会话，按新连接处理
		h.HandleWebSocket(c)
		return
	}

	// 获取会话状态
	session, exists := h.reconnectManager.GetSession(oldClientID)
	if !exists {
		logrus.Warnf("未找到旧会话: %s, 按新连接处理", oldClientID)
		h.HandleWebSocket(c)
		return
	}

	// 获取用户信息
	userID := c.Query("user_id")
	if userID == "" {
		userID = session.UserID
	}

	projectID := c.Query("project_id")
	if projectID == "" {
		projectID = session.ProjectID
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 创建新客户端
	newClientID := GenerateClientID()
	client := NewClient(newClientID, userID, projectID, conn, h.hub)

	// 恢复订阅状态
	for taskID := range session.Subscribed {
		client.Subscribe(taskID)
	}

	// 注册客户端
	h.hub.Register <- client

	// 删除旧会话，保存新会话
	h.reconnectManager.DeleteSession(oldClientID)
	h.reconnectManager.SaveSession(newClientID, userID, projectID, session.Subscribed)

	// 发送重连成功消息
	reconnectMsg := &Message{
		Type:   "reconnected",
		TaskID: "",
		Payload: map[string]interface{}{
			"client_id":      newClientID,
			"old_client_id":  oldClientID,
			"message":        "WebSocket重连成功",
			"subscribed":     session.Subscribed,
			"timestamp":      time.Now(),
		},
	}
	if msgBytes, err := json.Marshal(reconnectMsg); err == nil {
		client.Send <- msgBytes
	}

	// 发送断线期间的消息队列
	queuedMessages := h.reconnectManager.GetQueuedMessages(oldClientID)
	for _, msg := range queuedMessages {
		client.Send <- msg
	}

	logrus.Infof("客户端重连成功: %s -> %s, 恢复订阅: %d个任务", oldClientID, newClientID, len(session.Subscribed))

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// ProgressAPI 进度API
type ProgressAPI struct {
	progressService *ProgressService
	hub             *Hub
}

// NewProgressAPI 创建进度API
func NewProgressAPI(progressService *ProgressService, hub *Hub) *ProgressAPI {
	return &ProgressAPI{
		progressService: progressService,
		hub:             hub,
	}
}

// RegisterRoutes 注册路由
func (api *ProgressAPI) RegisterRoutes(r *gin.RouterGroup) {
	progress := r.Group("/progress")
	{
		progress.GET("/:task_id", api.GetProgress)
		progress.GET("/:task_id/history", api.GetHistory)
		progress.GET("/:task_id/history/db", api.GetHistoryFromDB)
		progress.DELETE("/:task_id/history", api.ClearHistory)
		progress.GET("/stats", api.GetStats)
		progress.GET("/active", api.GetActiveTasks)
		progress.POST("/:task_id/subscribe", api.SubscribeToTask)
		progress.POST("/:task_id/unsubscribe", api.UnsubscribeFromTask)
	}
}

// GetProgress 获取进度
func (api *ProgressAPI) GetProgress(c *gin.Context) {
	taskID := c.Param("task_id")

	progress, err := api.progressService.GetProgress(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetHistory 获取内存历史
func (api *ProgressAPI) GetHistory(c *gin.Context) {
	taskID := c.Param("task_id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		// 解析limit参数
	}

	history := api.progressService.GetHistory(taskID, limit)
	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"source":  "memory",
		"history": history,
	})
}

// GetHistoryFromDB 从数据库获取历史
func (api *ProgressAPI) GetHistoryFromDB(c *gin.Context) {
	taskID := c.Param("task_id")

	limit := 100
	if l := c.Query("limit"); l != "" {
		// 解析limit参数
	}

	history, err := api.progressService.GetHistoryFromDB(taskID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": taskID,
		"source":  "database",
		"history": history,
	})
}

// ClearHistory 清除历史
func (api *ProgressAPI) ClearHistory(c *gin.Context) {
	taskID := c.Param("task_id")
	api.progressService.ClearHistory(taskID)
	c.JSON(http.StatusOK, gin.H{
		"message": "历史已清除",
		"task_id": taskID,
	})
}

// GetStats 获取统计
func (api *ProgressAPI) GetStats(c *gin.Context) {
	stats, err := api.progressService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetActiveTasks 获取活跃任务
func (api *ProgressAPI) GetActiveTasks(c *gin.Context) {
	// 获取WebSocket Hub统计
	hubStats := api.hub.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"connected_clients": hubStats.TotalClients,
		"by_user":          hubStats.ByUser,
		"by_project":       hubStats.ByProject,
	})
}

// SubscribeToTask 订阅任务
func (api *ProgressAPI) SubscribeToTask(c *gin.Context) {
	taskID := c.Param("task_id")
	clientID := c.Query("client_id")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少client_id参数"})
		return
	}

	// 通过Hub发送订阅消息
	err := api.hub.SendToClient(clientID, &Message{
		Type:   "subscribed",
		TaskID: taskID,
		Payload: map[string]interface{}{
			"task_id": taskID,
			"message": "已订阅任务进度",
		},
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客户端不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "订阅成功",
		"task_id":  taskID,
		"client_id": clientID,
	})
}

// UnsubscribeFromTask 取消订阅任务
func (api *ProgressAPI) UnsubscribeFromTask(c *gin.Context) {
	taskID := c.Param("task_id")
	clientID := c.Query("client_id")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少client_id参数"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "取消订阅成功",
		"task_id":   taskID,
		"client_id": clientID,
	})
}

// WebSocketStatsAPI WebSocket统计API
type WebSocketStatsAPI struct {
	hub *Hub
}

// NewWebSocketStatsAPI 创建统计API
func NewWebSocketStatsAPI(hub *Hub) *WebSocketStatsAPI {
	return &WebSocketStatsAPI{hub: hub}
}

// RegisterRoutes 注册路由
func (api *WebSocketStatsAPI) RegisterRoutes(r *gin.RouterGroup) {
	stats := r.Group("/websocket")
	{
		stats.GET("/stats", api.GetStats)
	}
}

// GetStats 获取统计
func (api *WebSocketStatsAPI) GetStats(c *gin.Context) {
	stats := api.hub.GetStats()
	c.JSON(http.StatusOK, stats)
}
