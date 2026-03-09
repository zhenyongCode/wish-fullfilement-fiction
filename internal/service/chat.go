package service

import (
	"context"
	"fmt"
	"strings"
	"wish-fullfilement-fiction/internal/agent"
	"wish-fullfilement-fiction/internal/consts"

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
func (s *ChatService) Chat(ctx context.Context, task string) (*ChatResponse, error) {
	testA := agent.GetAgent(consts.BuildInAgentTranslation) // TODO: agent name and maxRounds

	res, err := testA.Run(ctx, task) // TODO: task
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Content: res,
	}, nil
}

// Exec 执行聊天（servicefunc 兼容方法）
// 符合 servicefunc.ServiceFunc 签名: func(ctx context.Context, params g.Map) (g.Map, error)
func (s *ChatService) Exec(ctx context.Context, params g.Map) (g.Map, error) {
	// Parse messages
	t := params["task"].(string)
	t = strings.TrimSpace(t)
	if len(t) == 0 {
		return nil, fmt.Errorf("empty task")
	}

	// Call Chat

	resp, err := s.Chat(ctx, t)
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
