package llm

import (
	"context"
	"time"
)

type ChatMessage struct {
	Role    string
	Content string
}

type ToolSpec struct {
	Name        string
	Description string
	JSONSchema  string
}

type ChatRequest struct {
	Model    string
	Messages []ChatMessage
	Tools    []ToolSpec
	Timeout  time.Duration
}

type ToolCall struct {
	ID    string
	Name  string
	Input map[string]any
}

type ChatResponse struct {
	StopReason string
	Content    string
	ToolCalls  []ToolCall
}

type Client interface {
	ChatCompletion(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
