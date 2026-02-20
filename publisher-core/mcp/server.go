package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

// Server MCP 服务器
type Server struct {
	name     string
	version  string
	tools    map[string]Tool
	resources map[string]Resource
	prompts  map[string]Prompt
	mu       sync.RWMutex
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewServer 创建 MCP 服务器
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = &ServerConfig{
			Name:    "publisher-tools",
			Version: "1.0.0",
		}
	}

	return &Server{
		name:     config.Name,
		version:  config.Version,
		tools:    make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:  make(map[string]Prompt),
	}
}

// =====================================================
// MCP 协议类型定义
// =====================================================

// JSONRPCRequest JSON-RPC 请求
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse JSON-RPC 响应
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError RPC 错误
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler 工具处理函数
type ToolHandler func(ctx context.Context, args map[string]interface{}) (*ToolResult, error)

// ToolResult 工具执行结果
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// Resource 资源定义
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Handler     ResourceHandler        `json:"-"`
}

// ResourceHandler 资源处理函数
type ResourceHandler func(ctx context.Context, uri string) (*ResourceContent, error)

// ResourceContent 资源内容
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// Prompt 提示词定义
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Handler     PromptHandler          `json:"-"`
}

// PromptArgument 提示词参数
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required,omitempty"`
}

// PromptHandler 提示词处理函数
type PromptHandler func(ctx context.Context, args map[string]string) (*PromptResult, error)

// PromptResult 提示词结果
type PromptResult struct {
	Messages []PromptMessage `json:"messages"`
}

// PromptMessage 提示词消息
type PromptMessage struct {
	Role    string       `json:"role"`
	Content ContentBlock `json:"content"`
}

// =====================================================
// 服务器能力
// =====================================================

// ServerInfo 服务器信息
type ServerInfo struct {
	Name    string       `json:"name"`
	Version string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities 服务器能力
type Capabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// ToolsCapability 工具能力
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability 资源能力
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability 提示词能力
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// =====================================================
// 工具注册
// =====================================================

// RegisterTool 注册工具
func (s *Server) RegisterTool(tool Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools[tool.Name] = tool
	logrus.Infof("MCP tool registered: %s", tool.Name)
	return nil
}

// RegisterResource 注册资源
func (s *Server) RegisterResource(resource Resource) error {
	if resource.URI == "" {
		return fmt.Errorf("resource URI is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.resources[resource.URI] = resource
	logrus.Infof("MCP resource registered: %s", resource.URI)
	return nil
}

// RegisterPrompt 注册提示词
func (s *Server) RegisterPrompt(prompt Prompt) error {
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.prompts[prompt.Name] = prompt
	logrus.Infof("MCP prompt registered: %s", prompt.Name)
	return nil
}

// =====================================================
// 请求处理
// =====================================================

// HandleRequest 处理 JSON-RPC 请求
func (s *Server) HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(ctx, req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(ctx, req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

// handleInitialize 处理初始化请求
func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": ServerInfo{
			Name:    s.name,
			Version: s.version,
			Capabilities: Capabilities{
				Tools:     &ToolsCapability{ListChanged: true},
				Resources: &ResourcesCapability{ListChanged: true},
				Prompts:   &PromptsCapability{ListChanged: true},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleToolsList 处理工具列表请求
func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleToolsCall 处理工具调用请求
func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	s.mu.RLock()
	tool, ok := s.tools[params.Name]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Tool not found",
				Data:    params.Name,
			},
		}
	}

	result, err := tool.Handler(ctx, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: &ToolResult{
				Content: []ContentBlock{
					{Type: "text", Text: err.Error()},
				},
				IsError: true,
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleResourcesList 处理资源列表请求
func (s *Server) handleResourcesList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]map[string]interface{}, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, map[string]interface{}{
			"uri":         resource.URI,
			"name":        resource.Name,
			"description": resource.Description,
			"mimeType":    resource.MimeType,
		})
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"resources": resources,
		},
	}
}

// handleResourcesRead 处理资源读取请求
func (s *Server) handleResourcesRead(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	s.mu.RLock()
	resource, ok := s.resources[params.URI]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Resource not found",
			},
		}
	}

	content, err := resource.Handler(ctx, params.URI)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"contents": []ResourceContent{*content},
		},
	}
}

// handlePromptsList 处理提示词列表请求
func (s *Server) handlePromptsList(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]map[string]interface{}, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   prompt.Arguments,
		})
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"prompts": prompts,
		},
	}
}

// handlePromptsGet 处理提示词获取请求
func (s *Server) handlePromptsGet(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	s.mu.RLock()
	prompt, ok := s.prompts[params.Name]
	s.mu.RUnlock()

	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Prompt not found",
			},
		}
	}

	result, err := prompt.Handler(ctx, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// =====================================================
// 辅助方法
// =====================================================

// TextContent 创建文本内容块
func TextContent(text string) ContentBlock {
	return ContentBlock{
		Type: "text",
		Text: text,
	}
}

// ImageContent 创建图片内容块
func ImageContent(data, mimeType string) ContentBlock {
	return ContentBlock{
		Type:     "image",
		Data:     data,
		MimeType: mimeType,
	}
}

// ResourceContent 创建资源内容块
func ResourceContentBlock(uri, mimeType string) ContentBlock {
	return ContentBlock{
		Type:     "resource",
		MimeType: mimeType,
	}
}

// SuccessResult 创建成功结果
func SuccessResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{TextContent(text)},
	}
}

// ErrorResult 创建错误结果
func ErrorResult(err error) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{TextContent(err.Error())},
		IsError: true,
	}
}
