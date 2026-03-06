package bifrost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	bifrostsdk "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"

	"wish-fullfilement-fiction/internal/llm"
)

var (
	ErrProviderRequired = errors.New("llm provider is required")
	ErrModelRequired    = errors.New("llm model is required")
)

// Client Bifrost LLM 客户端
type Client struct {
	provider string
	model    string
	timeout  time.Duration
	fallback []schemas.Fallback
	runtime  *bifrostsdk.Bifrost
}

// account 实现 bifrostsdk.Account 接口
type account struct {
	providers map[string]ProviderConfig
	timeout   int
	retries   int
}

// New 创建 Bifrost 客户端
func New(ctx context.Context, opts ...Option) (*Client, error) {
	// 加载配置
	cfg, err := LoadConfig(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("load bifrost config failed: %w", err)
	}

	// 应用选项
	opt := &options{
		provider: cfg.GetDefaultProvider(),
		model:    cfg.GetDefaultModel(),
		timeout:  cfg.GetTimeout(),
	}
	for _, o := range opts {
		o(opt)
	}

	// 验证必要参数
	provider := strings.TrimSpace(opt.provider)
	model := strings.TrimSpace(opt.model)
	if provider == "" {
		return nil, ErrProviderRequired
	}
	if model == "" {
		return nil, ErrModelRequired
	}

	// 创建 account
	acc := &account{
		providers: cfg.GetEnabledProviders(),
		timeout:   int(cfg.GetTimeout() / time.Second),
		retries:   cfg.Global.MaxRetries,
	}
	if acc.retries <= 0 {
		acc.retries = 2
	}

	// 如果没有启用的提供商，创建默认配置
	if len(acc.providers) == 0 {
		acc.providers = map[string]ProviderConfig{
			strings.ToLower(provider): {
				Enabled:   true,
				APIKeyEnv: strings.ToUpper(strings.ReplaceAll(provider, "-", "_")) + "_API_KEY",
				Models:    []string{model},
				Weight:    1,
			},
		}
	}

	// 初始化 Bifrost 运行时
	rt, err := bifrostsdk.Init(ctx, schemas.BifrostConfig{Account: acc})
	if err != nil {
		return nil, fmt.Errorf("init bifrost failed: %w", err)
	}

	// 构建回退链
	fallback := make([]schemas.Fallback, 0, len(cfg.Routing.Fallback))
	for _, f := range cfg.Routing.Fallback {
		p := strings.TrimSpace(strings.ToLower(f.Provider))
		m := strings.TrimSpace(f.Model)
		if p == "" || m == "" {
			continue
		}
		fallback = append(fallback, schemas.Fallback{
			Provider: schemas.ModelProvider(p),
			Model:    m,
		})
	}

	return &Client{
		provider: strings.ToLower(provider),
		model:    model,
		timeout:  opt.timeout,
		fallback: fallback,
		runtime:  rt,
	}, nil
}

// Option 客户端选项
type Option func(*options)

type options struct {
	provider string
	model    string
	timeout  time.Duration
}

// WithProvider 设置提供商
func WithProvider(provider string) Option {
	return func(o *options) {
		o.provider = provider
	}
}

// WithModel 设置模型
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.timeout = timeout
	}
}

// 实现 bifrostsdk.Account 接口
func (a *account) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	providers := make([]schemas.ModelProvider, 0, len(a.providers))
	for provider := range a.providers {
		providers = append(providers, schemas.ModelProvider(provider))
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i] < providers[j] })
	return providers, nil
}

func (a *account) GetKeysForProvider(_ context.Context, provider schemas.ModelProvider) ([]schemas.Key, error) {
	providerCfg, ok := a.providers[strings.ToLower(string(provider))]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	envName := strings.TrimSpace(providerCfg.APIKeyEnv)
	apiKey := ""
	if envName != "" {
		apiKey = strings.TrimSpace(os.Getenv(envName))
	}

	enabled := true
	weight := float64(providerCfg.Weight)
	if weight <= 0 {
		weight = 1.0
	}

	key := schemas.Key{
		ID:      strings.ToLower(string(provider)) + "-key",
		Name:    strings.ToLower(string(provider)) + "-key",
		Value:   *schemas.NewEnvVar(apiKey),
		Models:  providerCfg.Models,
		Weight:  weight,
		Enabled: &enabled,
	}
	return []schemas.Key{key}, nil
}

func (a *account) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	providerCfg, ok := a.providers[strings.ToLower(string(provider))]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	return &schemas.ProviderConfig{
		NetworkConfig: schemas.NetworkConfig{
			BaseURL:                        providerCfg.BaseURL,
			DefaultRequestTimeoutInSeconds: a.timeout,
			MaxRetries:                     a.retries,
		},
	}, nil
}

// ChatCompletion 执行聊天补全
func (c *Client) ChatCompletion(ctx context.Context, req llm.ChatRequest) (llm.ChatResponse, error) {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = c.model
	}
	if model == "" {
		return llm.ChatResponse{}, ErrModelRequired
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = c.timeout
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	bctx := schemas.NewBifrostContext(ctx, time.Now().Add(timeout))

	messages := make([]schemas.ChatMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		messages = append(messages, schemas.ChatMessage{
			Role: schemas.ChatMessageRole(msg.Role),
			Content: &schemas.ChatMessageContent{
				ContentStr: bifrostsdk.Ptr(content),
			},
		})
	}

	request := &schemas.BifrostRequest{
		RequestType: schemas.ChatCompletionRequest,
		ChatRequest: &schemas.BifrostChatRequest{
			Provider: schemas.ModelProvider(c.provider),
			Model:    model,
			Input:    messages,
			Params: &schemas.ChatParameters{
				Tools: toBifrostTools(req.Tools),
			},
			Fallbacks: c.fallback,
		},
	}

	result, err := c.runtime.ChatCompletionRequest(bctx, request.ChatRequest)
	if err != nil {
		msg := bifrostsdk.GetErrorMessage(err)
		if strings.TrimSpace(msg) == "" {
			msg = "unknown bifrost error"
		}
		return llm.ChatResponse{}, fmt.Errorf("bifrost chat completion failed: %s", msg)
	}

	if result == nil || len(result.Choices) == 0 {
		return llm.ChatResponse{}, errors.New("empty bifrost response")
	}

	choice := result.Choices[0]
	if choice.ChatNonStreamResponseChoice == nil || choice.ChatNonStreamResponseChoice.Message == nil {
		return llm.ChatResponse{}, errors.New("invalid bifrost response choice")
	}

	content := ""
	if choice.ChatNonStreamResponseChoice.Message.Content != nil && choice.ChatNonStreamResponseChoice.Message.Content.ContentStr != nil {
		content = strings.TrimSpace(*choice.ChatNonStreamResponseChoice.Message.Content.ContentStr)
	}

	stopReason := "end_turn"
	if choice.FinishReason != nil && *choice.FinishReason == string(schemas.BifrostFinishReasonToolCalls) {
		stopReason = "tool_use"
	}

	toolCalls := toLLMToolCalls(choice.ChatNonStreamResponseChoice.Message)
	if len(toolCalls) > 0 {
		stopReason = "tool_use"
	}

	return llm.ChatResponse{
		StopReason: stopReason,
		Content:    content,
		ToolCalls:  toolCalls,
	}, nil
}

func toBifrostTools(tools []llm.ToolSpec) []schemas.ChatTool {
	if len(tools) == 0 {
		return nil
	}
	result := make([]schemas.ChatTool, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}

		var params *schemas.ToolFunctionParameters
		schema := strings.TrimSpace(tool.JSONSchema)
		if schema != "" {
			parsed := &schemas.ToolFunctionParameters{}
			if err := schemas.Unmarshal([]byte(schema), parsed); err == nil {
				params = parsed
			}
		}

		desc := strings.TrimSpace(tool.Description)
		chatTool := schemas.ChatTool{
			Type: schemas.ChatToolTypeFunction,
			Function: &schemas.ChatToolFunction{
				Name:       name,
				Parameters: params,
			},
		}
		if desc != "" {
			chatTool.Function.Description = bifrostsdk.Ptr(desc)
		}

		result = append(result, chatTool)
	}
	return result
}

func toLLMToolCalls(message *schemas.ChatMessage) []llm.ToolCall {
	if message == nil || message.ChatAssistantMessage == nil || len(message.ChatAssistantMessage.ToolCalls) == 0 {
		return nil
	}

	toolCalls := make([]llm.ToolCall, 0, len(message.ChatAssistantMessage.ToolCalls))
	for _, call := range message.ChatAssistantMessage.ToolCalls {
		name := ""
		if call.Function.Name != nil {
			name = strings.TrimSpace(*call.Function.Name)
		}
		if name == "" {
			continue
		}

		id := ""
		if call.ID != nil {
			id = strings.TrimSpace(*call.ID)
		}
		if id == "" {
			id = name
		}

		toolCalls = append(toolCalls, llm.ToolCall{
			ID:    id,
			Name:  name,
			Input: parseToolArguments(call.Function.Arguments),
		})
	}

	if len(toolCalls) == 0 {
		return nil
	}
	return toolCalls
}

func parseToolArguments(raw string) map[string]any {
	input := map[string]any{}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return input
	}
	if err := json.Unmarshal([]byte(trimmed), &input); err == nil {
		return input
	}
	return map[string]any{"raw": trimmed}
}