package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"strings"
	"time"
	"wish-fullfilement-fiction/internal/consts"
	"wish-fullfilement-fiction/internal/llm"
	"wish-fullfilement-fiction/internal/llm/tools"
)

// Errors
var (
	ErrEmptyTask       = errors.New("task is required")
	ErrTooManyRounds   = errors.New("agent loop exceeds max rounds")
	ErrToolUnsupported = errors.New("unsupported tool")
)

const (
	toolResultPrefix = "tool_result:"
)

type Loop struct {
	// TODO：定义 Loop 结构体，包含必要的字段，例如 LLM 客户端、工具列表, skills, prompt等
	llm       llm.Client
	maxRounds int
	toolHub   *tools.Hub // 工具中心，管理可用的工具实例
}

func NewLoop(llm llm.Client, maxRounds int) *Loop {
	return &Loop{
		llm:       llm,
		maxRounds: maxRounds,
		toolHub:   tools.NewHub(consts.DefaultToolTimeout), // workspace，可以根据需要设置
	}
}

func (l *Loop) Execute(ctx context.Context, messages []llm.ChatMessage) (string, []llm.ChatMessage, error) {
	loopStart := time.Now()
	for round := 1; round <= l.maxRounds; round++ {
		roundStart := time.Now()
		resp, err := l.llm.ChatCompletion(ctx, llm.ChatRequest{
			Messages: messages,
			Tools:    l.toolHub.GetTools(),
		})
		if err != nil {
			return "", messages, err
		}
		assistantContent := strings.TrimSpace(resp.Content)
		if assistantContent == "" && len(resp.ToolCalls) > 0 {
			assistantContent = formatToolUseMessages(resp.ToolCalls)
		}
		if assistantContent != "" {
			messages = append(messages, llm.ChatMessage{Role: "assistant", Content: assistantContent})
		}
		// 根据停止原因判断是否继续循环
		if resp.StopReason != consts.StopReasonToolUse {
			g.Log().Debugf(ctx, "agent loop round %d stop reason: %s, content: %s", round, resp.StopReason, assistantContent)
			g.Log().Infof(ctx, "agent loop stopped at round %d, total time: %s , round time: %s", round, time.Since(loopStart), time.Since(roundStart))
			return resp.Content, messages, nil
		}
		// 处理工具调用
		toolResults, err := l.dispatchTools(ctx, resp.ToolCalls)
		if err != nil {
			return "", messages, err
		}
		resultPayload := strings.Join(toolResults, "\n")
		g.Log().Debugf(ctx, "agent loop round %d tool results: %s", round, resultPayload)
		// 将工具结果作为用户消息添加到对话中，继续下一轮交互
		messages = append(messages, llm.ChatMessage{Role: "assistant", Content: resultPayload})
	}

	return "I've completed processing but have no response to give.", messages, nil
}
func formatToolUseMessages(calls []llm.ToolCall) string {
	if len(calls) == 0 {
		return ""
	}
	lines := make([]string, 0, len(calls))
	for _, call := range calls {
		payload, _ := json.Marshal(map[string]any{
			"id":   call.ID,
			"name": call.Name,
		})
		lines = append(lines, "tool_use:"+string(payload))
	}
	return strings.Join(lines, "\n")
}
func (l *Loop) dispatchTools(ctx context.Context, calls []llm.ToolCall) ([]string, error) {
	results := make([]string, 0, len(calls))
	for _, call := range calls {
		g.Log().Debugf(ctx, "agent loop round %d tool calls: %s", call.ID, call.Name)
		result, err := l.executeTool(ctx, call)
		g.Log().Debugf(ctx, "agent loop round %d tool result: %s", call.ID, result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
func (l *Loop) executeTool(ctx context.Context, call llm.ToolCall) (string, error) {

	if l.toolHub == nil {
		return "", fmt.Errorf("%w: %s", ErrToolUnsupported, call.Name)
	}

	res, err := l.toolHub.Execute(ctx, call.Name, call.Input, consts.DefaultToolTimeout)
	if err != nil {
		if errors.Is(err, tools.ErrToolNotFound) {
			return "", fmt.Errorf("%w: %s", ErrToolUnsupported, call.Name)
		}
		return formatToolResult(call.ID, err.Error(), 1), nil
	}
	return formatToolResult(call.ID, res.Output, res.ExitCode), nil
}

func formatToolResult(toolCallID, output string, exitCode int) string {
	payload, _ := json.Marshal(map[string]any{
		"tool_use_id": toolCallID,
		"content":     output,
		"exit_code":   exitCode,
	})
	return toolResultPrefix + string(payload)
}
