package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/windlant/mcp-server/internal/config"
	"github.com/windlant/mcp-server/internal/tools/registry"
	"github.com/windlant/protocol/protocol/mcp_protocol"
	"github.com/windlant/protocol/types/tools_types"
)

// MCPServer 是 MCP 服务器的核心结构
type MCPServer struct {
	config   *config.Config
	registry *registry.Registry
}

// NewMCPServer 创建新的 MCP 服务器实例
func NewMCPServer(cfg *config.Config) *MCPServer {
	return &MCPServer{
		config:   cfg,
		registry: registry.NewRegistry(),
	}
}

// RegisterTool 注册一个工具到服务器
func (s *MCPServer) RegisterTool(def tools_types.ToolDefinition) {
	s.registry.Register(def)
}

// Start 启动 HTTP 服务器
func (s *MCPServer) Start(addr string) error {
	http.HandleFunc("/mcp", s.handleMCPRequest)
	return http.ListenAndServe(addr, nil)
}

// handleMCPRequest 处理 MCP 协议的 HTTP 请求
func (s *MCPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// 只允许 POST 方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 解码请求体
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 检查 method 字段是否存在
	method, ok := request["method"]
	if !ok {
		http.Error(w, "Missing 'method' field", http.StatusBadRequest)
		return
	}

	methodStr, ok := method.(string)
	if !ok {
		http.Error(w, "'method' must be a string", http.StatusBadRequest)
		return
	}

	// 根据 method 路由到相应的处理函数
	switch methodStr {
	case mcp_protocol.MCPMethodListTools:
		s.handleListTools(w, request)
	case mcp_protocol.MCPMethodCallTool:
		s.handleCallTool(w, request)
	default:
		http.Error(w, "Unknown method", http.StatusBadRequest)
	}
}

// handleListTools 处理 list_tools 请求
func (s *MCPServer) handleListTools(w http.ResponseWriter, req map[string]interface{}) {
	// 验证请求格式
	listReq := mcp_protocol.MCPListToolsRequest{}
	if err := mapToStruct(req, &listReq); err != nil {
		http.Error(w, "Invalid list_tools request format", http.StatusBadRequest)
		return
	}

	// 获取所有工具定义
	tools := s.registry.ListAll()

	// 准备响应（移除 Function 字段，因为不能 JSON 序列化）
	toolDefs := make([]tools_types.ToolDefinition, len(tools))
	for i, tool := range tools {
		toolDefs[i] = tools_types.ToolDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Function:    nil, // 不包含在响应中
		}
	}

	response := mcp_protocol.MCPListToolsResponse{
		Tools: toolDefs,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding list_tools response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleCallTool 处理 call_tool 请求
func (s *MCPServer) handleCallTool(w http.ResponseWriter, req map[string]interface{}) {
	callReq := mcp_protocol.MCPToolCallRequest{}
	if err := mapToStruct(req, &callReq); err != nil {
		http.Error(w, "Invalid call_tool request format", http.StatusBadRequest)
		return
	}

	// 查找工具
	toolDef, exists := s.registry.Get(callReq.Name)
	if !exists {
		response := mcp_protocol.MCPToolCallResponse{
			Error: "Tool not found: " + callReq.Name,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding call_tool error response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 执行工具
	result, err := toolDef.Function(callReq.Args)
	if err != nil {
		response := mcp_protocol.MCPToolCallResponse{
			Error: err.Error(),
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding call_tool error response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 返回成功响应
	response := mcp_protocol.MCPToolCallResponse{
		Result: result,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding call_tool success response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// mapToStruct 将 map[string]interface{} 转换为结构体
func mapToStruct(m map[string]interface{}, v interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
