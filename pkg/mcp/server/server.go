package server

import (
	"encoding/json"
	"errors"

	"github.com/0xYeah/shmClaw/pkg/mcp"
)

var (
	ErrMethodNotFound = errors.New("method not found")
	ErrInvalidParams  = errors.New("invalid params")
)

// ToolHandler is a function type that handles a specific tool call.
type ToolHandler func(args map[string]interface{}) (string, error)

// Server represents an MCP server that handles standard JSON-RPC requests.
type Server struct {
	Name    string
	Version string
	Tools   map[string]ToolHandler
}

// NewServer creates a new MCP Server instance.
func NewServer(name, version string) *Server {
	return &Server{
		Name:    name,
		Version: version,
		Tools:   make(map[string]ToolHandler),
	}
}

// RegisterTool registers a tool handler for a given tool name.
func (s *Server) RegisterTool(name string, handler ToolHandler) {
	s.Tools[name] = handler
}

// HandleRequest processes an incoming JSON-RPC payload and returns the appropriate response payload.
func (s *Server) HandleRequest(payload []byte) ([]byte, error) {
	var req mcp.JSONRPCRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return s.errorResponse(nil, -32700, "Parse error"), nil
	}

	var result interface{}
	var rpcErr *mcp.RPCError

	switch req.Method {
	case "initialize":
		result, rpcErr = s.handleInitialize(req.Params)
	case "tools/call":
		result, rpcErr = s.handleCallTool(req.Params)
	default:
		rpcErr = &mcp.RPCError{
			Code:    -32601,
			Message: ErrMethodNotFound.Error(),
		}
	}

	if rpcErr != nil {
		return s.errorResponse(req.ID, rpcErr.Code, rpcErr.Message), nil
	}

	return s.successResponse(req.ID, result), nil
}

func (s *Server) handleInitialize(params json.RawMessage) (interface{}, *mcp.RPCError) {
	var initReq mcp.InitializeRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &initReq); err != nil {
			return nil, &mcp.RPCError{Code: -32602, Message: ErrInvalidParams.Error()}
		}
	}

	res := mcp.InitializeResult{
		ProtocolVersion: "1.0",
	}
	res.ServerInfo.Name = s.Name
	res.ServerInfo.Version = s.Version
	res.Capabilities.Tools = len(s.Tools) > 0

	return res, nil
}

func (s *Server) handleCallTool(params json.RawMessage) (interface{}, *mcp.RPCError) {
	var callReq mcp.CallToolRequest
	if err := json.Unmarshal(params, &callReq); err != nil {
		return nil, &mcp.RPCError{Code: -32602, Message: ErrInvalidParams.Error()}
	}

	handler, ok := s.Tools[callReq.Name]
	if !ok {
		return nil, &mcp.RPCError{Code: -32601, Message: "Tool not found"}
	}

	output, err := handler(callReq.Arguments)
	isError := false
	if err != nil {
		output = err.Error()
		isError = true
	}

	var res mcp.CallToolResult
	res.IsError = isError
	res.Content = append(res.Content, struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}{
		Type: "text",
		Text: output,
	})

	return res, nil
}

func (s *Server) successResponse(id interface{}, result interface{}) []byte {
	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	return data
}

func (s *Server) errorResponse(id interface{}, code int, message string) []byte {
	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcp.RPCError{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}
