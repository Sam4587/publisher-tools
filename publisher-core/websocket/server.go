// Package websocket 提供WebSocket实时通信服务
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境需要更严格的检查
	},
}

// Server WebSocket服务器
type Server struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

// Client WebSocket客户端
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Topics []string
	Server *Server
}

// Message WebSocket消息
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewServer 创建WebSocket服务器
func NewServer() *Server {
	server := &Server{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}

	go server.run()

	return server
}

// run 运行服务器
func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.ID] = client
			s.mu.Unlock()
			logrus.Infof("WebSocket客户端已连接: %s", client.ID)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
			}
			s.mu.Unlock()
			logrus.Infof("WebSocket客户端已断开: %s", client.ID)

		case message := <-s.broadcast:
			s.broadcastMessage(message)
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 生成客户端ID
	clientID := generateClientID()

	client := &Client{
		ID:     clientID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Topics: []string{},
		Server: s,
	}

	// 注册客户端
	s.register <- client

	// 启动读写协程
	go client.readPump()
	go client.writePump()
}

// broadcastMessage 广播消息
func (s *Server) broadcastMessage(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		select {
		case client.Send <- data:
		default:
			// 发送通道已满，关闭客户端
			s.unregister <- client
		}
	}
}

// BroadcastToTopic 向特定主题广播消息
func (s *Server) BroadcastToTopic(topic string, message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		// 检查客户端是否订阅了该主题
		if client.isSubscribedTo(topic) {
			select {
			case client.Send <- data:
			default:
				s.unregister <- client
			}
		}
	}
}

// BroadcastToClient 向特定客户端发送消息
func (s *Server) BroadcastToClient(clientID string, message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	s.mu.RLock()
	client, exists := s.clients[clientID]
	s.mu.RUnlock()

	if exists {
		select {
		case client.Send <- data:
		default:
			s.unregister <- client
		}
	}
}

// GetConnectedClients 获取已连接的客户端
func (s *Server) GetConnectedClients() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]string, 0, len(s.clients))
	for id := range s.clients {
		clients = append(clients, id)
	}

	return clients
}

// GetClientCount 获取客户端数量
func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.clients)
}

// readPump 读取协程
func (c *Client) readPump() {
	defer func() {
		c.Server.unregister <- c
		c.Conn.Close()
	}()

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

// writePump 写入协程
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
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

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logrus.Errorf("WebSocket写入错误: %v", err)
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
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		logrus.Errorf("解析消息失败: %v", err)
		return
	}

	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg.Data)
	case "unsubscribe":
		c.handleUnsubscribe(msg.Data)
	case "ping":
		c.handlePing()
	default:
		logrus.Warnf("未知消息类型: %s", msg.Type)
	}
}

// handleSubscribe 处理订阅
func (c *Client) handleSubscribe(data interface{}) {
	var payload struct {
		Topics []string `json:"topics"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("序列化订阅数据失败: %v", err)
		return
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		logrus.Errorf("解析订阅数据失败: %v", err)
		return
	}

	c.mu.Lock()
	for _, topic := range payload.Topics {
		c.Topics = append(c.Topics, topic)
	}
	c.mu.Unlock()

	logrus.Infof("客户端 %s 订阅主题: %v", c.ID, payload.Topics)

	// 发送确认消息
	c.SendMessage(&Message{
		Type: "subscribed",
		Data: map[string]interface{}{
			"topics": payload.Topics,
		},
	})
}

// handleUnsubscribe 处理取消订阅
func (c *Client) handleUnsubscribe(data interface{}) {
	var payload struct {
		Topics []string `json:"topics"`
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("序列化取消订阅数据失败: %v", err)
		return
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		logrus.Errorf("解析取消订阅数据失败: %v", err)
		return
	}

	c.mu.Lock()
	for _, topic := range payload.Topics {
		for i, t := range c.Topics {
			if t == topic {
				c.Topics = append(c.Topics[:i], c.Topics[i+1:]...)
				break
			}
		}
	}
	c.mu.Unlock()

	logrus.Infof("客户端 %s 取消订阅主题: %v", c.ID, payload.Topics)

	// 发送确认消息
	c.SendMessage(&Message{
		Type: "unsubscribed",
		Data: map[string]interface{}{
			"topics": payload.Topics,
		},
	})
}

// handlePing 处理心跳
func (c *Client) handlePing() {
	c.SendMessage(&Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	})
}

// SendMessage 发送消息到客户端
func (c *Client) SendMessage(message *Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return fmt.Errorf("发送通道已满")
	}
}

// isSubscribedTo 检查是否订阅了主题
func (c *Client) isSubscribedTo(topic string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, t := range c.Topics {
		if t == topic {
			return true
		}
	}

	return false
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// ProgressMessage 进度消息
type ProgressMessage struct {
	ExecutionID string                 `json:"execution_id"`
	StepID      string                 `json:"step_id"`
	Progress    int                    `json:"progress"`
	CurrentStep string                 `json:"current_step"`
	TotalSteps  int                    `json:"total_steps"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SendProgress 发送进度更新
func (s *Server) SendProgress(executionID string, progress ProgressMessage) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type: "progress",
		Data: progress,
	})
}

// StatusChangeMessage 状态变更消息
type StatusChangeMessage struct {
	ExecutionID string    `json:"execution_id"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

// SendStatusChange 发送状态变更
func (s *Server) SendStatusChange(executionID string, status string) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type: "status_change",
		Data: StatusChangeMessage{
			ExecutionID: executionID,
			Status:      status,
			Timestamp:   time.Now(),
		},
	})
}

// ErrorMessage 错误消息
type ErrorMessage struct {
	ExecutionID string    `json:"execution_id"`
	StepID      string    `json:"step_id,omitempty"`
	Error       string    `json:"error"`
	Timestamp   time.Time `json:"timestamp"`
}

// SendError 发送错误
func (s *Server) SendError(executionID string, stepID string, err error) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type: "error",
		Data: ErrorMessage{
			ExecutionID: executionID,
			StepID:      stepID,
			Error:       err.Error(),
			Timestamp:   time.Now(),
		},
	})
}

// CompletedMessage 完成消息
type CompletedMessage struct {
	ExecutionID string                 `json:"execution_id"`
	Status      string                 `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SendCompleted 发送完成通知
func (s *Server) SendCompleted(executionID string, status string, output map[string]interface{}) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type: "completed",
		Data: CompletedMessage{
			ExecutionID: executionID,
			Status:      status,
			Output:      output,
			Timestamp:   time.Now(),
		},
	})
}
