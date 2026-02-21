package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Client WebSocket客户端
type Client struct {
	ID         string
	UserID     string
	ProjectID  string
	Conn       *websocket.Conn
	Send       chan []byte
	Hub        *Hub
	Subscribed map[string]bool // 订阅的任务ID
	mu         sync.RWMutex
}

// Hub WebSocket中心
type Hub struct {
	Clients    map[string]*Client
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

// Message WebSocket消息
type Message struct {
	Type    string      `json:"type"`    // progress, status, notification
	TaskID  string      `json:"task_id"` // 任务ID
	Payload interface{} `json:"payload"` // 消息内容
}

// ProgressMessage 进度消息
type ProgressMessage struct {
	TaskID       string    `json:"task_id"`
	Progress     int       `json:"progress"`      // 0-100
	CurrentStep  string    `json:"current_step"`
	TotalSteps   int       `json:"total_steps"`
	CompletedSteps int     `json:"completed_steps"`
	Message      string    `json:"message"`
	Status       string    `json:"status"` // pending, running, completed, failed
	Timestamp    time.Time `json:"timestamp"`
}

// StatusMessage 状态消息
type StatusMessage struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// NewHub 创建Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.ID] = client
			h.mu.Unlock()
			logrus.Infof("WebSocket客户端已连接: %s, 总数: %d", client.ID, len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				delete(h.Clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			logrus.Infof("WebSocket客户端已断开: %s, 总数: %d", client.ID, len(h.Clients))

		case message := <-h.Broadcast:
			h.broadcastMessage(message)
		}
	}
}

// broadcastMessage 广播消息
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	for _, client := range h.Clients {
		client.mu.RLock()
		subscribed := client.Subscribed[message.TaskID] || message.TaskID == ""
		client.mu.RUnlock()

		if subscribed {
			select {
			case client.Send <- messageBytes:
			default:
				// 发送缓冲区已满，关闭连接
				close(client.Send)
				delete(h.Clients, client.ID)
			}
		}
	}
}

// BroadcastProgress 广播进度更新
func (h *Hub) BroadcastProgress(taskID string, progress *ProgressMessage) {
	h.Broadcast <- &Message{
		Type:   "progress",
		TaskID: taskID,
		Payload: progress,
	}
}

// BroadcastStatus 广播状态更新
func (h *Hub) BroadcastStatus(taskID string, status *StatusMessage) {
	h.Broadcast <- &Message{
		Type:   "status",
		TaskID: taskID,
		Payload: status,
	}
}

// BroadcastToUser 向特定用户广播
func (h *Hub) BroadcastToUser(userID string, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	for _, client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- messageBytes:
			default:
				logrus.Warnf("客户端发送缓冲区已满: %s", client.ID)
			}
		}
	}
}

// BroadcastToProject 向项目广播
func (h *Hub) BroadcastToProject(projectID string, message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	for _, client := range h.Clients {
		if client.ProjectID == projectID {
			select {
			case client.Send <- messageBytes:
			default:
				logrus.Warnf("客户端发送缓冲区已满: %s", client.ID)
			}
		}
	}
}

// NewClient 创建客户端
func NewClient(id, userID, projectID string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		ID:         id,
		UserID:     userID,
		ProjectID:  projectID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        hub,
		Subscribed: make(map[string]bool),
	}
}

// ReadPump 读取消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("WebSocket读取错误: %v", err)
			}
			break
		}

		// 处理客户端消息
		c.handleMessage(message)
	}
}

// WritePump 发送消息
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送消息
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Action string `json:"action"` // subscribe, unsubscribe
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		logrus.Errorf("解析客户端消息失败: %v", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	switch msg.Action {
	case "subscribe":
		c.Subscribed[msg.TaskID] = true
		logrus.Infof("客户端 %s 订阅任务: %s", c.ID, msg.TaskID)

	case "unsubscribe":
		delete(c.Subscribed, msg.TaskID)
		logrus.Infof("客户端 %s 取消订阅任务: %s", c.ID, msg.TaskID)
	}
}

// Subscribe 订阅任务
func (c *Client) Subscribe(taskID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Subscribed[taskID] = true
}

// Unsubscribe 取消订阅
func (c *Client) Unsubscribe(taskID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Subscribed, taskID)
}

// GetStats 获取统计信息
func (h *Hub) GetStats() *HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := &HubStats{
		TotalClients: len(h.Clients),
		ByUser:      make(map[string]int),
		ByProject:   make(map[string]int),
	}

	for _, client := range h.Clients {
		stats.ByUser[client.UserID]++
		stats.ByProject[client.ProjectID]++
	}

	return stats
}

// HubStats Hub统计
type HubStats struct {
	TotalClients int            `json:"total_clients"`
	ByUser       map[string]int `json:"by_user"`
	ByProject    map[string]int `json:"by_project"`
}

// GetClientCount 获取客户端数量
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

// GetClientsByUser 获取用户的客户端
func (h *Hub) GetClientsByUser(userID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	for _, client := range h.Clients {
		if client.UserID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientsByProject 获取项目的客户端
func (h *Hub) GetClientsByProject(projectID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	for _, client := range h.Clients {
		if client.ProjectID == projectID {
			clients = append(clients, client)
		}
	}
	return clients
}

// SendToClient 向特定客户端发送消息
func (h *Hub) SendToClient(clientID string, message *Message) error {
	h.mu.RLock()
	client, exists := h.Clients[clientID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("客户端不存在: %s", clientID)
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	select {
	case client.Send <- messageBytes:
		return nil
	default:
		return fmt.Errorf("客户端发送缓冲区已满")
	}
}
