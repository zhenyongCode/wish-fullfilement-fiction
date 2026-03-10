// Package restclient 提供一个通用的 OpenAI 兼容 RESTful HTTP 大模型客户端，
// 实现 llm.Client 接口，可对接任意兼容 OpenAI Chat Completions API 的服务，
// 例如 OpenAI、Azure OpenAI、Moonshot、DeepSeek、GLM、Ollama 等。
//
// 支持特性：
//   - 多 Provider 配置（每个 provider 独立 base_url / api_key）
//   - 同一 provider 内多模型按权重随机选择
//   - 单 provider 内请求失败自动重试（指数退避）
//   - 全局失败回退链：主 provider 耗尽后依次尝试 fallback provider
//   - 向下兼容旧版单 provider 扁平配置
package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"wish-fullfilement-fiction/internal/llm"
)

// ─────────────────────────────────────────────
//  OpenAI-compatible request / response types
// ─────────────────────────────────────────────

type chatMessage struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"` // string or null
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []toolCall  `json:"tool_calls,omitempty"`
}

type toolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // "function"
	Function toolCallFunction `json:"function"`
}

type toolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string
}

type functionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type toolDef struct {
	Type     string      `json:"type"` // "function"
	Function functionDef `json:"function"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	Tools     []toolDef     `json:"tools,omitempty"`
	MaxTokens *int          `json:"max_tokens,omitempty"`
	Stream    bool          `json:"stream"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
	Error   *apiError    `json:"error,omitempty"`
}

type apiError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Code    interface{} `json:"code"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("api error [%v]: %s", e.Code, e.Message)
}

// ─────────────────────────────────────────────
//  Internal resolved provider (ready to call)
// ─────────────────────────────────────────────

type resolvedProvider struct {
	name       string
	baseURL    string
	apiKey     string
	models     []ModelEntry // weighted list
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
}

// pickModel 按权重随机选择一个模型；若 override 非空则直接使用。
// 返回 (chosen model name, remaining models for next attempt)
func (p *resolvedProvider) pickModel(override string, tried map[string]bool) (string, bool) {
	if override != "" {
		return override, true
	}
	// 过滤掉已尝试的模型
	candidates := make([]ModelEntry, 0, len(p.models))
	for _, m := range p.models {
		if !tried[m.Name] {
			candidates = append(candidates, m)
		}
	}
	if len(candidates) == 0 {
		return "", false
	}

	// 计算总权重
	total := 0
	for _, m := range candidates {
		w := m.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}

	// 加权随机选择
	r := rand.Intn(total) //nolint:gosec
	acc := 0
	for _, m := range candidates {
		w := m.Weight
		if w <= 0 {
			w = 1
		}
		acc += w
		if r < acc {
			return m.Name, true
		}
	}
	return candidates[len(candidates)-1].Name, true
}

// ─────────────────────────────────────────────
//  Client
// ─────────────────────────────────────────────

// Client 通用 OpenAI 兼容 REST 客户端，支持多模型路由与失败回退。
type Client struct {
	cfg        *Config
	primary    []*resolvedProvider          // 主 provider 列表（按配置顺序）
	provByName map[string]*resolvedProvider // provider name → resolvedProvider（用于 fallback 查找）
	fallback   []FallbackEntry              // 全局回退链
}

// New 从已有 Config 创建客户端
func New(cfg *Config) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	primary := make([]*resolvedProvider, 0, len(cfg.Providers))
	byName := make(map[string]*resolvedProvider, len(cfg.Providers))

	for i := range cfg.Providers {
		p := &cfg.Providers[i]
		apiKey, err := p.resolveAPIKey()
		if err != nil {
			return nil, fmt.Errorf("restclient: providers[%d] (%s): %w", i, p.Name, err)
		}
		timeout := p.resolveTimeout()
		rp := &resolvedProvider{
			name:       p.Name,
			baseURL:    strings.TrimRight(strings.TrimSpace(p.BaseURL), "/"),
			apiKey:     apiKey,
			models:     p.effectiveModels(),
			timeout:    timeout,
			maxRetries: p.MaxRetries,
			httpClient: &http.Client{Timeout: timeout + 5*time.Second},
		}
		primary = append(primary, rp)
		if p.Name != "" {
			byName[p.Name] = rp
		}
	}

	return &Client{
		cfg:        cfg,
		primary:    primary,
		provByName: byName,
		fallback:   cfg.Fallback,
	}, nil
}

// NewFromContext 从 gcfg 配置自动创建客户端
func NewFromContext(ctx context.Context) (*Client, error) {
	cfg, err := LoadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// ─────────────────────────────────────────────
//  llm.Client 接口实现
// ─────────────────────────────────────────────

// ChatCompletion 调用 OpenAI-compatible Chat Completions API。
//
// 路由策略：
//  1. 遍历主 provider 列表；对每个 provider 按权重随机选模型，失败则换模型重试，
//     直到该 provider 所有模型均已尝试 (+ max_retries 次)。
//  2. 所有主 provider 耗尽后，依次尝试 cfg.Fallback 回退链中的 provider/model。
//  3. 仍然失败则返回合并错误。
func (c *Client) ChatCompletion(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	// 若 req 指定了 model，将其锁定（跳过路由，直接用主 provider[0]）
	overrideModel := strings.TrimSpace(req.Model)

	var errs []string

	// ── Phase 1: 主 provider 列表 ──
	for _, prov := range c.primary {
		resp, err := c.tryProvider(ctx, prov, overrideModel, req)
		if err == nil {
			return resp, nil
		}
		errs = append(errs, fmt.Sprintf("[provider=%s] %s", prov.name, err.Error()))
	}

	// ── Phase 2: 全局回退链 ──
	for _, fb := range c.fallback {
		prov, ok := c.provByName[fb.Provider]
		if !ok {
			errs = append(errs, fmt.Sprintf("[fallback provider=%s not found]", fb.Provider))
			continue
		}
		model := strings.TrimSpace(fb.Model)
		resp, err := c.tryProvider(ctx, prov, model, req)
		if err == nil {
			return resp, nil
		}
		errs = append(errs, fmt.Sprintf("[fallback provider=%s model=%s] %s", fb.Provider, model, err.Error()))
	}

	return llm.ChatResponse{}, fmt.Errorf("restclient: all providers failed:\n  %s",
		strings.Join(errs, "\n  "))
}

// tryProvider 对单个 provider 尝试所有可用模型（+ 每模型 max_retries 次重试）
func (c *Client) tryProvider(ctx context.Context, prov *resolvedProvider, overrideModel string, req llm.ChatRequest) (llm.ChatResponse, error) {
	tried := make(map[string]bool)
	var lastErr error

	for {
		model, ok := prov.pickModel(overrideModel, tried)
		if !ok {
			break
		}
		tried[model] = true

		resp, err := c.tryWithRetry(ctx, prov, model, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		// override model 只有一个，失败后不再换模型
		if overrideModel != "" {
			break
		}
	}

	if lastErr != nil {
		return llm.ChatResponse{}, lastErr
	}
	return llm.ChatResponse{}, fmt.Errorf("no models available for provider %s", prov.name)
}

// tryWithRetry 对单个 provider + model 进行最多 maxRetries+1 次尝试（指数退避）
func (c *Client) tryWithRetry(ctx context.Context, prov *resolvedProvider, model string, req llm.ChatRequest) (llm.ChatResponse, error) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = prov.timeout
	}

	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	bodyBytes, err := buildRequestBody(model, req)
	if err != nil {
		return llm.ChatResponse{}, err
	}

	url := prov.baseURL + "/chat/completions"
	maxAttempts := prov.maxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))*500) * time.Millisecond
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
			select {
			case <-rctx.Done():
				return llm.ChatResponse{}, rctx.Err()
			case <-time.After(backoff):
			}
		}

		raw, err := doHTTPRequest(rctx, prov.httpClient, prov.apiKey, url, bodyBytes)
		if err != nil {
			lastErr = err
			// 上下文已取消，不再重试
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}
			continue
		}

		resp, err := parseResponse(raw)
		if err != nil {
			lastErr = err
			continue
		}
		return resp, nil
	}
	return llm.ChatResponse{}, fmt.Errorf("model %s: %w", model, lastErr)
}

// ─────────────────────────────────────────────
//  HTTP 层
// ─────────────────────────────────────────────

func buildRequestBody(model string, req llm.ChatRequest) ([]byte, error) {
	body := chatRequest{
		Model:    model,
		Messages: toLLMMessages(req.Messages),
		Tools:    toToolDefs(req.Tools),
		Stream:   false,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("restclient: marshal request failed: %w", err)
	}
	return b, nil
}

func doHTTPRequest(ctx context.Context, hc *http.Client, apiKey, url string, body []byte) (*chatResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("restclient: create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	httpResp, err := hc.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("restclient: http request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("restclient: read response body failed: %w", err)
	}

	var result chatResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("restclient: unmarshal response failed: %w (body: %s)", err, truncate(string(respBytes), 200))
	}
	if result.Error != nil {
		return nil, result.Error
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("restclient: http %d: %s", httpResp.StatusCode, truncate(string(respBytes), 300))
	}
	return &result, nil
}

// ─────────────────────────────────────────────
//  转换 / 解析辅助函数
// ─────────────────────────────────────────────

func toLLMMessages(msgs []llm.ChatMessage) []chatMessage {
	result := make([]chatMessage, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, chatMessage{Role: m.Role, Content: m.Content})
	}
	return result
}

func toToolDefs(tools []llm.ToolSpec) []toolDef {
	if len(tools) == 0 {
		return nil
	}
	result := make([]toolDef, 0, len(tools))
	for _, t := range tools {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		fd := functionDef{
			Name:        name,
			Description: strings.TrimSpace(t.Description),
		}
		if schema := strings.TrimSpace(t.JSONSchema); schema != "" {
			fd.Parameters = json.RawMessage(schema)
		}
		result = append(result, toolDef{Type: "function", Function: fd})
	}
	return result
}

func parseResponse(resp *chatResponse) (llm.ChatResponse, error) {
	if len(resp.Choices) == 0 {
		return llm.ChatResponse{}, fmt.Errorf("restclient: empty choices in response")
	}
	choice := resp.Choices[0]

	content := ""
	if s, ok := choice.Message.Content.(string); ok {
		content = strings.TrimSpace(s)
	}

	stopReason := strings.TrimSpace(choice.FinishReason)
	if stopReason == "" {
		stopReason = "stop"
	}
	if stopReason == "tool_calls" {
		stopReason = "tool_use"
	}

	toolCalls := parseToolCalls(choice.Message.ToolCalls)
	if len(toolCalls) > 0 {
		stopReason = "tool_use"
	}

	return llm.ChatResponse{
		StopReason: stopReason,
		Content:    content,
		ToolCalls:  toolCalls,
	}, nil
}

func parseToolCalls(calls []toolCall) []llm.ToolCall {
	if len(calls) == 0 {
		return nil
	}
	result := make([]llm.ToolCall, 0, len(calls))
	for _, c := range calls {
		name := strings.TrimSpace(c.Function.Name)
		if name == "" {
			continue
		}
		id := strings.TrimSpace(c.ID)
		if id == "" {
			id = name
		}
		result = append(result, llm.ToolCall{
			ID:    id,
			Name:  name,
			Input: parseArguments(c.Function.Arguments),
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func parseArguments(raw string) map[string]any {
	out := map[string]any{}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return out
	}
	if err := json.Unmarshal([]byte(trimmed), &out); err == nil {
		return out
	}
	return map[string]any{"raw": trimmed}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
