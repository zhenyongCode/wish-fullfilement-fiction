package service

import (
	"context"
	"encoding/json"
	"fmt"
	"wish-fullfilement-fiction/internal/agent"

	"github.com/gogf/gf/v2/frame/g"
	"wish-fullfilement-fiction/internal/llm"
	"wish-fullfilement-fiction/internal/llm/bifrost"
)

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages []llm.ChatMessage
	Tools    []llm.ToolSpec
	Timeout  int64 // seconds
}

// ChatResponse 聊天响应
type ChatResponse struct {
	StopReason string
	Content    string
	ToolCalls  []llm.ToolCall
}

// ChatService 聊天服务
type ChatService struct {
	client *bifrost.Client
}

// NewChatService 创建聊天服务
func NewChatService(ctx context.Context) (*ChatService, error) {
	client, err := bifrost.New(ctx)
	if err != nil {
		return nil, err
	}
	return &ChatService{client: client}, nil
}

// Chat 执行聊天
func (s *ChatService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	testA := agent.NewAgent("test_agent", s.client) // TODO: agent name and maxRounds

	res, err := testA.Run(ctx, "天气怎么样") // TODO: task
	if err != nil {
		return nil, err
	}

	//chatReq := llm.ChatRequest{
	//	Messages: req.Messages,
	//	Tools:    req.Tools,
	//}
	//
	//resp, err := s.client.ChatCompletion(ctx, chatReq)
	//if err != nil {
	//	return nil, err
	//}

	return &ChatResponse{
		Content: res,
	}, nil
}

// Exec 执行聊天（servicefunc 兼容方法）
// 符合 servicefunc.ServiceFunc 签名: func(ctx context.Context, params g.Map) (g.Map, error)
func (s *ChatService) Exec(ctx context.Context, params g.Map) (g.Map, error) {
	// Parse messages
	messages, err := parseMessages(params)
	if err != nil {
		return nil, err
	}

	// Parse tools (optional)
	var tools []llm.ToolSpec
	if t, ok := params["tools"]; ok && t != nil {
		tools, err = parseTools(t)
		if err != nil {
			return nil, err
		}
	}

	// Get timeout (optional)
	var timeout int64
	if t, ok := params["timeout"]; ok {
		switch v := t.(type) {
		case int:
			timeout = int64(v)
		case int64:
			timeout = v
		case float64:
			timeout = int64(v)
		}
	}

	// Call Chat
	resp, err := s.Chat(ctx, &ChatRequest{
		Messages: messages,
		Tools:    tools,
		Timeout:  timeout,
	})
	if err != nil {
		return nil, err
	}

	// Build result
	result := g.Map{
		"stop_reason": resp.StopReason,
		"content":     resp.Content,
	}
	if len(resp.ToolCalls) > 0 {
		result["tool_calls"] = resp.ToolCalls
	}

	return result, nil
}

// parseMessages parses messages from params
func parseMessages(params g.Map) ([]llm.ChatMessage, error) {
	msgsRaw, ok := params["messages"]
	if !ok {
		return nil, fmt.Errorf("messages parameter required")
	}

	jsonBytes, err := json.Marshal(msgsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse messages: %v", err)
	}

	var messages []llm.ChatMessage
	if err := json.Unmarshal(jsonBytes, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse messages: %v", err)
	}

	return messages, nil
}

// parseTools parses tools from params
func parseTools(toolsRaw interface{}) ([]llm.ToolSpec, error) {
	jsonBytes, err := json.Marshal(toolsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tools: %v", err)
	}

	var tools []llm.ToolSpec
	if err := json.Unmarshal(jsonBytes, &tools); err != nil {
		return nil, fmt.Errorf("failed to parse tools: %v", err)
	}

	return tools, nil
}
