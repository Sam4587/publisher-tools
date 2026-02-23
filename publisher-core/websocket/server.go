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

// Server WebSocket服务器 (使用Hub实现)
type Server struct {
	hub *Hub
	mu  sync.RWMutex
}

// NewServer 创建WebSocket服务器
func NewServer() *Server {
	hub := NewHub()
	go hub.Run()
	return &Server{
		hub: hub,
	}
}

// HandleWebSocket 处理WebSocket连接
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 生产环境需要更严格的检查
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 生成客户端ID
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	client := &Client{
		ID:         clientID,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		Hub:        s.hub,
		Subscribed: make(map[string]bool),
	}

	// 注册客户端
	s.hub.Register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// Broadcast 广播消息到所有客户端
func (s *Server) Broadcast(message *Message) {
	data, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("序列化消息失败: %v", err)
		return
	}

	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	for _, client := range s.hub.Clients {
		select {
		case client.Send <- data:
		default:
			// 发送通道已满
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

	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	for _, client := range s.hub.Clients {
		client.mu.RLock()
		subscribed := client.Subscribed[topic]
		client.mu.RUnlock()

		if subscribed {
			select {
			case client.Send <- data:
			default:
				// 发送通道已满
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

	s.hub.mu.RLock()
	client, exists := s.hub.Clients[clientID]
	s.hub.mu.RUnlock()

	if exists {
		select {
		case client.Send <- data:
		default:
			// 发送通道已满
		}
	}
}

// GetConnectedClients 获取已连接的客户端
func (s *Server) GetConnectedClients() []string {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	clients := make([]string, 0, len(s.hub.Clients))
	for id := range s.hub.Clients {
		clients = append(clients, id)
	}

	return clients
}

// GetClientCount 获取客户端数量
func (s *Server) GetClientCount() int {
	s.hub.mu.RLock()
	defer s.hub.mu.RUnlock()

	return len(s.hub.Clients)
}

// readPump 读取协程
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
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

		// 处理客户端消息 (使用hub.go中的handleMessage)
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

// SendProgress 发送进度更新
func (s *Server) SendProgress(executionID string, progress ProgressMessage) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type:    "progress",
		TaskID:  executionID,
		Payload: progress,
	})
}

// SendStatusChange 发送状态变更
func (s *Server) SendStatusChange(executionID string, status string) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type:   "status_change",
		TaskID: executionID,
		Payload: StatusMessage{
			TaskID:    executionID,
			Status:    status,
			Timestamp: time.Now(),
		},
	})
}

// SendError 发送错误
func (s *Server) SendError(executionID string, stepID string, err error) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type:   "error",
		TaskID: executionID,
		Payload: map[string]interface{}{
			"execution_id": executionID,
			"step_id":      stepID,
			"error":        err.Error(),
			"timestamp":    time.Now(),
		},
	})
}

// SendCompleted 发送完成通知
func (s *Server) SendCompleted(executionID string, status string, output map[string]interface{}) {
	s.BroadcastToTopic(fmt.Sprintf("execution:%s", executionID), &Message{
		Type:   "completed",
		TaskID: executionID,
		Payload: map[string]interface{}{
			"execution_id": executionID,
			"status":       status,
			"output":       output,
			"timestamp":    time.Now(),
		},
	})
}
